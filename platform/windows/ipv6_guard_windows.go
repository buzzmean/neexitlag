//go:build windows

package windowsplatform

import (
	"fmt"
	"os"
	"time"
	"unsafe"
)

// IPv6 is never redirected into the IPv4 relay. Attributable selected new flows
// are suppressed at NETWORK; other traffic is reinjected unchanged. Unsupported
// packets/unknown identity follow fail-open and are explicitly not leak-proof.
func (b *WinDivertBackend) guardIPv6(classify Classifier) {
	defer b.wg.Done()
	d := b.proc()
	buf := make([]byte, 65575)
	for {
		var n uint32
		var a [80]byte
		ok, _, err := d.recv.Call(b.ipv6, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)), uintptr(unsafe.Pointer(&n)), uintptr(unsafe.Pointer(&a[0])))
		if ok == 0 {
			if b.ctx.Err() == nil {
				b.fail(fmt.Errorf("IPv6 guard receive: %w", err))
			}
			return
		}
		if n == 0 || n > uint32(len(buf)) {
			continue
		}
		packet := buf[:n]
		t, flags, parseErr := ParseIPv6Transport(packet)
		id := ProcessIdentity{}
		if parseErr == nil {
			b.associations.mu.RLock()
			association, known := b.associations.items[t]
			b.associations.mu.RUnlock()
			if known && IsCurrent(association.Identity) {
				id = association.Identity
			} else {
				var pid uint32
				var ownerErr error
				if t.Protocol == "tcp" {
					pid, ownerErr = resolveTCPOwner(t)
				} else {
					var raw []byte
					raw, ownerErr = queryUDPTable(23)
					if ownerErr == nil {
						var owners map[uint32]bool
						owners, ownerErr = udpOwners(raw, t.Source, true)
						if ownerErr == nil {
							pid, ownerErr = uniqueUDPOwner(owners)
						}
					}
				}
				if ownerErr == nil {
					id, _ = b.cache.Lookup(pid)
				}
			}
			decision, _ := ClassifyCapture(id, t, uint32(os.Getpid()), classify)
			b.evidence.RecordBytes(id, t, "captured", "", uint64(n))
			b.evidence.RecordInitialPolicy(id, t, decision)
			if b.v6guard.Block(id, t, flags, decision, uint32(os.Getpid()), time.Now()) {
				b.evidence.Record(id, t, "ipv6_blocked", "")
				b.drop(id, t)
				continue
			}
			b.evidence.RecordBytes(id, t, "direct", "", uint64(n))
		}
		// No header or routing modification. NETWORK injection is supported.
		var sent uint32
		ok, _, err = d.send.Call(b.ipv6, uintptr(unsafe.Pointer(&packet[0])), uintptr(n), uintptr(unsafe.Pointer(&sent)), uintptr(unsafe.Pointer(&a[0])))
		if ok == 0 && b.ctx.Err() == nil {
			b.fail(fmt.Errorf("IPv6 passthrough: %w", err))
			return
		}
	}
}
