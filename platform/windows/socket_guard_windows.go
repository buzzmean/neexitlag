//go:build windows

package windowsplatform

import (
	"fmt"
	"os"
	"unsafe"
)

func (b *WinDivertBackend) socketGuard(classify Classifier) {
	defer b.wg.Done()
	d := b.proc()
	for {
		var a [80]byte
		ok, _, err := d.recv.Call(b.socket, 0, 0, 0, uintptr(unsafe.Pointer(&a[0])))
		if ok == 0 {
			if b.ctx.Err() == nil {
				b.fail(fmt.Errorf("IPv6 socket guard: %w", err))
			}
			return
		}
		pid, t, decodeErr := DecodeSocketEvent(a)
		id, lookupErr := b.cache.Lookup(pid)
		// Unattributable sockets stay DIRECT. This is explicit fail-open, never a
		// claim of leak-proof IPv6. The service and unrelated apps are never blocked.
		if decodeErr == nil && lookupErr == nil && t.Protocol == "tcp" && t.Source.Addr().Is4() && t.Source.Port() != 0 && !t.Source.Addr().IsUnspecified() {
			b.associations.Put(Association{Identity: id, Tuple: t})
		}
		if decodeErr == nil && lookupErr == nil && t.Source.Addr().Is6() && BlockUnsupportedIPv6(id, t, a[9], uint32(os.Getpid()), classify) {
			b.evidence.Record(id, t, "ipv6_blocked", "")
			continue
		}
		ok, _, err = d.send.Call(b.socket, 0, 0, 0, uintptr(unsafe.Pointer(&a[0])))
		if ok == 0 && b.ctx.Err() == nil {
			b.fail(fmt.Errorf("IPv6 socket admission: %w", err))
			return
		}
	}
}
