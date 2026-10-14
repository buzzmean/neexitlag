//go:build windows

package windowsplatform

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/buzzmean/neexitlag/core/localstrategy"
	"github.com/buzzmean/neexitlag/core/protocol"
	"github.com/buzzmean/neexitlag/core/routing"
	"github.com/buzzmean/neexitlag/core/telemetry"
	"golang.org/x/sys/windows"
	"net/netip"
	"os"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type captureFlow struct {
	localState localstrategy.State
	decision   Decision
	dropCause  telemetry.Category
	queue      chan []byte
	done       chan struct{}
	once       sync.Once
	datagram   routing.Datagram
	template   []byte
	address    [80]byte
	identity   ProcessIdentity
	endpoint   uint64
	lastSeen   time.Time
}
type WinDivertBackend struct {
	Files         DriverFiles
	Engine        *routing.Engine
	RX, TX, Drops atomic.Uint64
	LocalDrops    telemetry.Counters
	mu            sync.Mutex
	life          sync.Mutex
	status        CaptureStatus
	dll           *divertDLL
	native        *divertDLL
	network, flow uintptr
	socket        uintptr
	ipv6          uintptr
	v6guard       IPv6Guard
	flows         map[Tuple]*captureFlow
	associations  *FlowAssociations
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	workers       chan struct{}
	cache         *ProcessCache
	tcp           *tcpRedirect
	evidence      *EvidenceLedger
}

func NewWinDivert(files DriverFiles, engine *routing.Engine) *WinDivertBackend {
	return &WinDivertBackend{Files: files, Engine: engine, status: CaptureStatus{Backend: "WinDivert", Capabilities: Capabilities{UDP: true, TCP: true, IPv4: true, Transparent: true}}, flows: make(map[Tuple]*captureFlow), associations: NewAssociations(8192), workers: make(chan struct{}, 64), cache: NewProcessCache(NewDiscovery(), IsCurrent), evidence: NewEvidenceLedger()}
}
func (b *WinDivertBackend) Status() CaptureStatus { b.mu.Lock(); defer b.mu.Unlock(); return b.status }
func (d *divertDLL) openHandle(filter string, layer uintptr, flags uintptr) (uintptr, error) {
	s, e := windows.BytePtrFromString(filter)
	if e != nil {
		return 0, e
	}
	h, _, callErr := d.open.Call(uintptr(unsafe.Pointer(s)), layer, 0, flags)
	if h == ^uintptr(0) {
		return 0, fmt.Errorf("WinDivertOpen(layer=%d filter=%q flags=%d): %w; administrator access and signed driver required", layer, filter, flags, callErr)
	}
	return h, nil
}
func (b *WinDivertBackend) Start(parent context.Context, classify Classifier) (result error) {
	defer func() {
		if result != nil {
			b.mu.Lock()
			b.status.Ready = false
			b.status.Error = result.Error()
			b.mu.Unlock()
		}
	}()
	b.life.Lock()
	defer b.life.Unlock()
	b.mu.Lock()
	if b.dll != nil {
		b.mu.Unlock()
		return errors.New("capture already started")
	}
	b.mu.Unlock()
	dll, e := loadDivert(b.Files)
	if e != nil {
		b.mu.Lock()
		b.status.Error = e.Error()
		b.mu.Unlock()
		return e
	}
	flow, e := dll.openHandle("outbound and !loopback and (udp or tcp)", 2, 1|4)
	if e != nil {
		dll.dll.Release()
		return e
	}
	network, e := dll.openHandle("outbound and ip and udp and !loopback and !impostor", 0, 0)
	if e != nil {
		dll.close.Call(flow)
		dll.dll.Release()
		return e
	}
	socket, e := dll.openHandle("(tcp and event == CONNECT) or (ipv6 and (event == CONNECT or (udp and event == BIND)))", 3, 1|4)
	if e != nil {
		dll.close.Call(network)
		dll.close.Call(flow)
		dll.dll.Release()
		return e
	}
	ipv6, e := dll.openHandle("outbound and ipv6 and (tcp or udp) and !loopback and !impostor", 0, 0)
	if e != nil {
		dll.close.Call(socket)
		dll.close.Call(network)
		dll.close.Call(flow)
		dll.dll.Release()
		return e
	}
	var major uint64
	ok, _, _ := dll.getParam.Call(network, 3, uintptr(unsafe.Pointer(&major)))
	if ok == 0 || major != 2 {
		dll.close.Call(ipv6)
		dll.close.Call(socket)
		dll.close.Call(network)
		dll.close.Call(flow)
		dll.dll.Release()
		return errors.New("WinDivert v2 driver required")
	}
	dll.setParam.Call(network, 0, 256)
	dll.setParam.Call(network, 1, 100)
	dll.setParam.Call(network, 2, 512*1024)
	tcp, e := b.startTCP(dll)
	if e != nil {
		dll.close.Call(ipv6)
		dll.close.Call(socket)
		dll.close.Call(network)
		dll.close.Call(flow)
		dll.dll.Release()
		return e
	}
	b.mu.Lock()
	b.tcp = tcp
	b.socket = socket
	b.ipv6 = ipv6
	b.v6guard = IPv6Guard{}
	b.dll = dll
	b.native = dll
	b.flow = flow
	b.network = network
	b.ctx, b.cancel = context.WithCancel(parent)
	b.status.Ready = true
	b.status.Error = ""
	b.mu.Unlock()
	b.wg.Add(5)
	go b.track()
	go b.socketGuard(classify)
	go b.guardIPv6(classify)
	go b.cleanup()
	go b.packets(classify)
	tcp.start(classify)
	watch := b.ctx
	go func() {
		<-watch.Done()
		b.life.Lock()
		defer b.life.Unlock()
		if b.ctx == watch {
			b.stopLocked()
		}
	}()
	return nil
}
func (b *WinDivertBackend) Stop() error { b.life.Lock(); defer b.life.Unlock(); return b.stopLocked() }
func (b *WinDivertBackend) stopLocked() error {
	b.mu.Lock()
	if b.dll == nil {
		b.mu.Unlock()
		return nil
	}
	d := b.dll
	b.cancel()
	b.status.Ready = false
	b.dll = nil
	d.shutdown.Call(b.network, 1)
	d.shutdown.Call(b.flow, 1)
	d.shutdown.Call(b.socket, 1)
	d.shutdown.Call(b.ipv6, 1)
	for _, f := range b.flows {
		f.once.Do(func() { close(f.done) })
	}
	b.mu.Unlock()
	b.tcp.stop()
	b.wg.Wait()
	d.close.Call(b.network)
	d.close.Call(b.flow)
	d.close.Call(b.socket)
	d.close.Call(b.ipv6)
	d.dll.Release()
	b.mu.Lock()
	b.flows = make(map[Tuple]*captureFlow)
	b.associations = NewAssociations(8192)
	b.mu.Unlock()
	return nil
}
func (b *WinDivertBackend) track() {
	defer b.wg.Done()
	d := b.proc()
	discovery := b.cache
	var address [80]byte
	for {
		ok, _, e := d.recv.Call(b.flow, 0, 0, 0, uintptr(unsafe.Pointer(&address[0])))
		if ok == 0 {
			if b.ctx.Err() == nil {
				b.fail(fmt.Errorf("WinDivert flow receive: %w", e))
			}
			return
		}
		if b.ctx.Err() != nil {
			return
		}
		if address[10]&0x10 != 0 {
			continue
		}
		data := address[16:]
		pid := binary.LittleEndian.Uint32(data[16:20])
		proto := "udp"
		if data[56] == 6 {
			proto = "tcp"
		} else if data[56] != 17 {
			continue
		}
		src, dst := [4]byte{}, [4]byte{}
		binary.BigEndian.PutUint32(src[:], binary.LittleEndian.Uint32(data[32:36]))
		binary.BigEndian.PutUint32(dst[:], binary.LittleEndian.Uint32(data[48:52]))
		tuple := Tuple{proto, netip.AddrPortFrom(netip.AddrFrom4(src), binary.LittleEndian.Uint16(data[52:54])), netip.AddrPortFrom(netip.AddrFrom4(dst), binary.LittleEndian.Uint16(data[54:56]))}
		endpoint := binary.LittleEndian.Uint64(data[:8])
		event := address[9]
		if event == 2 {
			b.associations.Delete(tuple, endpoint)
			b.mu.Lock()
			if f := b.flows[tuple]; f != nil && f.endpoint == endpoint {
				f.once.Do(func() { close(f.done) })
				delete(b.flows, tuple)
			}
			b.mu.Unlock()
			continue
		}
		identity, e := discovery.Lookup(pid)
		if e == nil {
			b.associations.Put(Association{identity, tuple, endpoint})
		}
	}
}

// proc is retained until receive loops have stopped; Stop only clears publication.
func (b *WinDivertBackend) proc() *divertDLL { return b.native }
func (b *WinDivertBackend) fail(e error) {
	b.mu.Lock()
	b.status.Error = e.Error()
	b.mu.Unlock()
	b.cancel()
}
func (b *WinDivertBackend) inject(d *divertDLL, p []byte, a [80]byte) error {
	d.checksum.Call(uintptr(unsafe.Pointer(&p[0])), uintptr(len(p)), uintptr(unsafe.Pointer(&a[0])), 0)
	var sent uint32
	ok, _, e := d.send.Call(b.network, uintptr(unsafe.Pointer(&p[0])), uintptr(len(p)), uintptr(unsafe.Pointer(&sent)), uintptr(unsafe.Pointer(&a[0])))
	if ok == 0 {
		return fmt.Errorf("packet reinjection: %w", e)
	}
	b.evidence.Record(ProcessIdentity{}, Tuple{}, "reinject", "")
	return nil
}
func (b *WinDivertBackend) packets(classify Classifier) {
	defer b.wg.Done()
	d := b.proc()
	buffer := make([]byte, 65535)
	for {
		var n uint32
		var address [80]byte
		ok, _, e := d.recv.Call(b.network, uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)), uintptr(unsafe.Pointer(&n)), uintptr(unsafe.Pointer(&address[0])))
		if ok == 0 {
			if b.ctx.Err() == nil {
				b.LocalDrops.Drop(telemetry.CaptureReceiveErrors)
				b.fail(fmt.Errorf("WinDivert packet receive: %w", e))
			}
			return
		}
		if n == 0 || n > uint32(len(buffer)) {
			b.LocalDrops.Drop(telemetry.CaptureProcessing)
			b.fail(errors.New("invalid WinDivert packet size"))
			return
		}
		packet := buffer[:n]
		tuple, payload, e := ParseUDP(packet)
		if e != nil || b.ctx.Err() != nil {
			if e = b.inject(d, packet, address); e != nil && b.ctx.Err() == nil {
				b.fail(e)
				return
			}
			continue
		}
		b.associations.mu.RLock()
		association, known := b.associations.items[tuple]
		b.associations.mu.RUnlock()
		identity := association.Identity
		known = known && IsCurrent(identity)
		if !known {
			identity = ProcessIdentity{}
			if pid, err := resolveUDPOwner(tuple.Source); err == nil {
				if resolved, err := b.cache.Lookup(pid); err == nil {
					identity = resolved
					known = true
					association = Association{Identity: identity, Tuple: tuple}
					b.associations.Put(association)
				}
			}
		}
		b.associations.Touch(tuple)
		b.mu.Lock()
		f := b.flows[tuple]
		if f != nil && known && f.identity.CreationTime == 0 {
			f.identity = identity
			f.endpoint = association.Endpoint
		}
		if f != nil && known && f.endpoint == 0 && f.identity.CreationTime == identity.CreationTime {
			f.endpoint = association.Endpoint
		}
		if f != nil && known && f.identity.CreationTime != 0 && (f.identity.PID != identity.PID || f.identity.CreationTime != identity.CreationTime) {
			f.once.Do(func() { close(f.done) })
			delete(b.flows, tuple)
			f = nil
		}
		if f == nil {
			b.evidence.BeginFlow(identity, tuple)
		}
		b.evidence.RecordBytes(identity, tuple, "captured", "", uint64(len(packet)))
		if known && identity.PID == uint32(os.Getpid()) {
			b.evidence.Record(identity, tuple, "self", "")
		}
		if f == nil {
			decision := Decision{Route: routing.Route{}}
			if known {
				decision, _ = ClassifyCapture(identity, tuple, uint32(os.Getpid()), classify)
			}
			if len(b.flows) >= 8192 {
				b.mu.Unlock()
				if decision.Block || decision.BlockOnFailure {
					b.drop(identity, tuple)
				} else {
					if err := b.inject(d, packet, address); err != nil {
						b.fail(err)
						return
					}
					b.evidence.RecordBytes(identity, tuple, "direct", "", uint64(len(packet)))
				}
				continue
			}
			f = &captureFlow{decision: decision, dropCause: telemetry.Policy, identity: identity, endpoint: association.Endpoint, done: make(chan struct{}), address: address}
			if !decision.Route.Direct() && !decision.Block {
				f.queue = make(chan []byte, 32)
				f.template = append([]byte(nil), packet...)
			}
			b.flows[tuple] = f
			if !decision.Route.Direct() && !decision.Block {
				select {
				case b.workers <- struct{}{}:
					b.wg.Add(1)
					go b.forward(tuple, f, d)
				default:
					if f.decision.BlockOnFailure {
						f.dropCause = telemetry.WorkerSaturation
						f.decision.Block = true
					} else {
						f.decision.Route = routing.Route{}
					}
				}
			}
		}
		f.lastSeen = time.Now()
		decision := f.decision
		b.mu.Unlock()
		// An expired identity must not retain a previous process's BLOCK decision.
		if decision.Block && known {
			b.LocalDrops.Attempt(f.dropCause)
			b.LocalDrops.Drop(f.dropCause)
			b.drop(identity, tuple)
			continue
		}
		if decision.Route.Direct() || !known {
			if e = SendLocal(decision, &f.localState, identity, tuple, packet, func(q []byte) error { return b.inject(d, q, LocalPacketAddress(q, address)) }, b.evidence); e != nil {
				b.fail(e)
				return
			}
			continue
		}
		if len(payload) > protocol.MaxPayload {
			b.LocalDrops.Attempt(telemetry.CaptureProcessing)
			b.LocalDrops.Drop(telemetry.CaptureProcessing)
			b.drop(identity, tuple)
			continue
		}
		// Commit queueing under the same lock as fallback, so draining cannot miss a late enqueue.
		b.mu.Lock()
		latest := f.decision
		if latest.Block {
			b.mu.Unlock()
			b.LocalDrops.Attempt(telemetry.Policy)
			b.LocalDrops.Drop(telemetry.Policy)
			b.drop(identity, tuple)
			continue
		}
		if latest.Route.Direct() {
			b.mu.Unlock()
			b.LocalDrops.Attempt(telemetry.CaptureProcessing)
			if err := b.inject(d, packet, address); err != nil {
				b.LocalDrops.Drop(telemetry.CaptureProcessing)
				b.fail(err)
				return
			}
			b.evidence.RecordBytes(identity, tuple, "direct", "", uint64(len(packet)))
			continue
		}
		b.LocalDrops.Attempt(telemetry.CaptureQueue)
		select {
		case f.queue <- append([]byte(nil), payload...):
		default:
			b.LocalDrops.Drop(telemetry.CaptureQueue)
			b.drop(identity, tuple)
		}
		b.mu.Unlock()
	}
}
func (b *WinDivertBackend) forward(tuple Tuple, f *captureFlow, d *divertDLL) {
	defer b.wg.Done()
	defer func() { <-b.workers }()
	ctx, cancel := context.WithTimeout(b.ctx, 750*time.Millisecond)
	datagram, err := b.Engine.OpenUDP(ctx, f.decision.Flow)
	cancel()
	if err == nil {
		for _, flow := range b.Engine.Snapshot() {
			if flow.PID == f.decision.Flow.PID && flow.Source == tuple.Source && flow.Destination == tuple.Destination {
				if flow.Route.Direct() {
					err = errors.New("relay setup fell back to DIRECT")
				} else {
					b.mu.Lock()
					f.decision.Route = flow.Route
					b.mu.Unlock()
				}
			}
		}
	}
	if err != nil {
		if datagram != nil {
			datagram.Close()
		}
		b.fallback(tuple, f, d, nil)
		return
	}

	defer datagram.Close()
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			recvCtx, cancel := context.WithTimeout(b.ctx, time.Second)
			payload, e := datagram.Receive(recvCtx)
			cancel()
			if e != nil {
				if b.ctx.Err() != nil {
					return
				}
				select {
				case <-f.done:
					return
				default:
				}
				if errors.Is(e, context.DeadlineExceeded) {
					continue
				}
				return
			}
			packet, e := ReplyUDP(f.template, payload)
			if e != nil {
				b.LocalDrops.Attempt(telemetry.ReplyReinjection)
				b.LocalDrops.Drop(telemetry.ReplyReinjection)
				b.drop(f.identity, tuple)
				continue
			}
			address := f.address
			address[10] &= ^byte(2)
			b.LocalDrops.Attempt(telemetry.ReplyReinjection)
			if err := b.inject(d, packet, address); err != nil {
				b.LocalDrops.Drop(telemetry.ReplyReinjection)
				b.fail(err)
				return
			}
			b.mu.Lock()
			route := f.decision.Route.RelayID
			b.mu.Unlock()
			b.evidence.RecordBytes(f.identity, tuple, "reply", route, uint64(len(payload)))
			b.RX.Add(uint64(len(payload)))
		}
	}()
	for {
		select {
		case <-b.ctx.Done():
			datagram.Close()
			<-done
			return
		case <-f.done:
			datagram.Close()
			<-done
			return
		case <-done:
			b.fallback(tuple, f, d, nil)
			return
		case payload := <-f.queue:
			if datagram.Send(payload) != nil {
				b.fallback(tuple, f, d, payload)
				datagram.Close()
				<-done
				return
			}
			b.mu.Lock()
			route := f.decision.Route.RelayID
			b.mu.Unlock()
			b.evidence.RecordBytes(f.identity, tuple, "relay", route, uint64(len(payload)))
			b.TX.Add(uint64(len(payload)))
		}
	}
}

func (b *WinDivertBackend) Traffic() (uint64, uint64) { return b.RX.Load(), b.TX.Load() }

func (b *WinDivertBackend) Diagnostics(pid uint32) CaptureEvidence {
	return b.DiagnosticsTarget(pid, "")
}
func (b *WinDivertBackend) CaptureMarker() EvidenceMarker { return b.evidence.Marker() }
func (b *WinDivertBackend) DiagnosticsTarget(pid uint32, destination string) CaptureEvidence {
	out, _ := b.DiagnosticsScoped(EvidenceQuery{PID: pid, Destination: destination})
	return out
}
func (b *WinDivertBackend) DiagnosticsScoped(query EvidenceQuery) (CaptureEvidence, error) {
	b.mu.Lock()
	tcp := b.tcp
	b.mu.Unlock()
	active := map[evidenceKey]bool{}
	if tcp != nil {
		tcp.state.mu.Lock()
		for _, e := range tcp.state.entries {
			if e.mode == "relay" && e.local != nil && e.upstream != nil {
				active[evidenceKey{e.identity.PID, e.identity.CreationTime, e.tuple}] = true
			}
		}
		tcp.state.mu.Unlock()
	}
	out, err := b.evidence.SnapshotQuery(query, active)
	if err != nil {
		return out, err
	}
	for i := range out.Flows {
		f := &out.Flows[i]
		f.IdentityCurrent = IsCurrent(ProcessIdentity{PID: f.PID, CreationTime: f.CreationTime, Path: f.Executable})
	}
	out.Counters.Dropped = b.Drops.Load()
	out.LocalDrops = b.LocalDrops.Snapshot()
	out.LocalDrops["capture_kernel_queue_drops"] = telemetry.Metric{Available: false, Unit: "packets"}
	return out, nil
}
func (b *WinDivertBackend) fallback(tuple Tuple, f *captureFlow, d *divertDLL, failed []byte) {
	b.mu.Lock()
	block := f.decision.BlockOnFailure
	f.decision.Block = block
	if !block {
		f.decision.Route = routing.Route{}
	}
	b.mu.Unlock()
	inject := func(payload []byte) {
		if block {
			b.drop(f.identity, tuple)
			return
		}
		p, err := ReplyUDP(f.template, payload)
		if err != nil {
			b.drop(f.identity, tuple)
			return
		}
		ihl := int(p[0]&15) * 4
		copy(p[12:20], f.template[12:20])
		copy(p[ihl:ihl+4], f.template[ihl:ihl+4])
		if err = b.inject(d, p, f.address); err != nil {
			b.fail(err)
			return
		}
		b.evidence.RecordBytes(f.identity, tuple, "direct", "", uint64(len(p)))
	}
	if failed != nil {
		inject(failed)
	}
	for {
		select {
		case payload := <-f.queue:
			inject(payload)
		default:
			return
		}
	}
}
func (b *WinDivertBackend) cleanup() {
	defer b.wg.Done()
	tick := time.NewTicker(10 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-b.ctx.Done():
			return
		case now := <-tick.C:
			b.cache.Cleanup(now)
			b.evidence.Cleanup(now)
			b.associations.Cleanup(now)
			b.v6guard.Cleanup(now)
			b.mu.Lock()
			for t, f := range b.flows {
				if now.Sub(f.lastSeen) > 60*time.Second {
					f.once.Do(func() { close(f.done) })
					delete(b.flows, t)
				}
			}
			b.mu.Unlock()
		}
	}
}

func (b *WinDivertBackend) drop(id ProcessIdentity, t Tuple) {
	b.Drops.Add(1)
	b.evidence.Record(id, t, "drop", "")
}
