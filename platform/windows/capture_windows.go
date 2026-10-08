//go:build windows

package windowsplatform

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/buzzmean/neexitlag/core/routing"
	"golang.org/x/sys/windows"
	"net/netip"
	"os"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type captureFlow struct {
	decision Decision
	queue    chan []byte
	done     chan struct{}
	once     sync.Once
	datagram routing.Datagram
	template []byte
	address  [80]byte
	identity ProcessIdentity
	endpoint uint64
}
type WinDivertBackend struct {
	Files         DriverFiles
	Engine        *routing.Engine
	RX, TX, Drops atomic.Uint64
	mu            sync.Mutex
	life          sync.Mutex
	status        CaptureStatus
	dll           *divertDLL
	native        *divertDLL
	network, flow uintptr
	flows         map[Tuple]*captureFlow
	associations  *FlowAssociations
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	workers       chan struct{}
}

func NewWinDivert(files DriverFiles, engine *routing.Engine) *WinDivertBackend {
	return &WinDivertBackend{Files: files, Engine: engine, status: CaptureStatus{Backend: "WinDivert", Capabilities: Capabilities{UDP: true, Transparent: true}}, flows: make(map[Tuple]*captureFlow), associations: NewAssociations(8192), workers: make(chan struct{}, 64)}
}
func (b *WinDivertBackend) Status() CaptureStatus { b.mu.Lock(); defer b.mu.Unlock(); return b.status }
func (d *divertDLL) openHandle(filter string, layer uintptr, flags uintptr) (uintptr, error) {
	s, e := windows.BytePtrFromString(filter)
	if e != nil {
		return 0, e
	}
	h, _, callErr := d.open.Call(uintptr(unsafe.Pointer(s)), layer, 0, flags)
	if h == ^uintptr(0) {
		return 0, fmt.Errorf("WinDivertOpen: %w; administrator access and signed driver required", callErr)
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
	var major uint64
	ok, _, _ := dll.getParam.Call(network, 3, uintptr(unsafe.Pointer(&major)))
	if ok == 0 || major != 2 {
		dll.close.Call(network)
		dll.close.Call(flow)
		dll.dll.Release()
		return errors.New("WinDivert v2 driver required")
	}
	dll.setParam.Call(network, 0, 256)
	dll.setParam.Call(network, 1, 100)
	dll.setParam.Call(network, 2, 512*1024)
	b.mu.Lock()
	b.dll = dll
	b.native = dll
	b.flow = flow
	b.network = network
	b.ctx, b.cancel = context.WithCancel(parent)
	b.status.Ready = true
	b.status.Error = ""
	b.mu.Unlock()
	b.wg.Add(2)
	go b.track()
	go b.packets(classify)
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
	for _, f := range b.flows {
		f.once.Do(func() { close(f.done) })
	}
	b.mu.Unlock()
	b.wg.Wait()
	d.close.Call(b.network)
	d.close.Call(b.flow)
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
	discovery := NewDiscovery()
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
		if pid == uint32(os.Getpid()) {
			continue
		}
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
				b.fail(fmt.Errorf("WinDivert packet receive: %w", e))
			}
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
		b.mu.Lock()
		f := b.flows[tuple]
		if f != nil && known && f.identity.CreationTime == 0 {
			f.identity = identity
			f.endpoint = association.Endpoint
		}
		if f != nil && known && f.identity.CreationTime != 0 && f.identity.CreationTime != identity.CreationTime {
			f.once.Do(func() { close(f.done) })
			delete(b.flows, tuple)
			f = nil
		}
		if f == nil {
			decision := Decision{Route: routing.Route{}}
			if known && len(payload) <= 1100 {
				decision = classify(identity, tuple)
			}
			if len(b.flows) >= 8192 {
				b.mu.Unlock()
				if decision.Block || decision.BlockOnFailure {
					b.Drops.Add(1)
				} else {
					b.inject(d, packet, address)
				}
				continue
			}
			f = &captureFlow{decision: decision, identity: identity, endpoint: association.Endpoint, done: make(chan struct{}), address: address}
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
						f.decision.Block = true
					} else {
						f.decision.Route = routing.Route{}
					}
				}
			}
		}
		decision := f.decision
		b.mu.Unlock()
		// An expired identity must not retain a previous process's BLOCK decision.
		if decision.Block && known {
			b.Drops.Add(1)
			continue
		}
		if decision.Route.Direct() || !known {
			if e = b.inject(d, packet, address); e != nil {
				b.fail(e)
				return
			}
			continue
		}
		if len(payload) > 1100 {
			b.Drops.Add(1)
			continue
		}
		select {
		case f.queue <- append([]byte(nil), payload...):
		default:
			b.Drops.Add(1)
		}
	}
}
func (b *WinDivertBackend) forward(tuple Tuple, f *captureFlow, d *divertDLL) {
	defer b.wg.Done()
	defer func() { <-b.workers }()
	ctx, cancel := context.WithTimeout(b.ctx, 100*time.Millisecond)
	datagram, err := b.Engine.OpenUDP(ctx, f.decision.Flow)
	cancel()
	if err == nil {
		for _, flow := range b.Engine.Snapshot() {
			if flow.PID == f.decision.Flow.PID && flow.Source == tuple.Source && flow.Destination == tuple.Destination && flow.Route.Direct() {
				err = errors.New("relay setup fell back to DIRECT")
			}
		}
	}
	if err != nil {
		if datagram != nil {
			datagram.Close()
		}
		b.mu.Lock()
		if f.decision.BlockOnFailure {
			f.decision.Block = true
		} else {
			f.decision.Route = routing.Route{}
		}
		b.mu.Unlock()
		if f.decision.BlockOnFailure {
			return
		}
		for {
			select {
			case p := <-f.queue:
				packet, e := ReplyUDP(f.template, p)
				if e == nil {
					ihl := int(packet[0]&15) * 4
					copy(packet[12:16], f.template[12:16])
					copy(packet[16:20], f.template[16:20])
					copy(packet[ihl:ihl+4], f.template[ihl:ihl+4])
					b.inject(d, packet, f.address)
				}
			default:
				return
			}
		}
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
				b.Drops.Add(1)
				continue
			}
			address := f.address
			address[10] &= ^byte(2)
			if b.inject(d, packet, address) != nil {
				return
			}
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
			b.mu.Lock()
			if f.decision.BlockOnFailure {
				f.decision.Block = true
			} else {
				f.decision.Route = routing.Route{}
			}
			b.mu.Unlock()
			return
		case payload := <-f.queue:
			if datagram.Send(payload) != nil {
				b.mu.Lock()
				if f.decision.BlockOnFailure {
					f.decision.Block = true
				} else {
					f.decision.Route = routing.Route{}
				}
				b.mu.Unlock()
				datagram.Close()
				<-done
				return
			}
			b.TX.Add(uint64(len(payload)))
		}
	}
}

func (b *WinDivertBackend) Traffic() (uint64, uint64) { return b.RX.Load(), b.TX.Load() }
