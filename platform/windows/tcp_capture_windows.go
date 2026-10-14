//go:build windows

package windowsplatform

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/netip"
	"os"
	"sync"
	"time"
	"unsafe"
)

type tcpRedirect struct {
	b        *WinDivertBackend
	handle   uintptr
	listener *net.TCPListener
	port     uint16
	state    *tcpState
	slots    chan struct{}
	wg       sync.WaitGroup
}

func (b *WinDivertBackend) startTCP(d *divertDLL) (*tcpRedirect, error) {
	l, err := net.ListenTCP("tcp4", &net.TCPAddr{IP: net.IPv4zero})
	if err != nil {
		return nil, err
	}
	port := uint16(l.Addr().(*net.TCPAddr).Port)
	// Inbound physical connections to the internal listener are dropped. Reflected
	// packets injected below this handle are the only admitted listener traffic.
	filter := fmt.Sprintf("ip and tcp and !impostor and ((outbound and !loopback) or (inbound and tcp.DstPort == %d))", port)
	h, err := d.openHandle(filter, 0, 0)
	if err != nil {
		l.Close()
		return nil, err
	}
	d.setParam.Call(h, 0, 2048)
	d.setParam.Call(h, 1, 100)
	d.setParam.Call(h, 2, 4*1024*1024)
	return &tcpRedirect{b: b, handle: h, listener: l, port: port, state: newTCPState(8192), slots: make(chan struct{}, 128)}, nil
}
func (t *tcpRedirect) start(classify Classifier) {
	t.wg.Add(3)
	go t.packets(classify)
	go t.accept()
	go t.cleanup()
}
func (t *tcpRedirect) stop() {
	t.listener.Close()
	t.b.proc().shutdown.Call(t.handle, 1)
	t.state.mu.Lock()
	for _, e := range t.state.entries {
		if e.upstream != nil {
			e.upstream.Close()
		}
		if e.local != nil {
			e.local.Close()
		}
	}
	t.state.mu.Unlock()
	t.wg.Wait()
	t.b.proc().close.Call(t.handle)
}
func (t *tcpRedirect) send(p []byte, a [80]byte) error {
	d := t.b.proc()
	d.checksum.Call(uintptr(unsafe.Pointer(&p[0])), uintptr(len(p)), uintptr(unsafe.Pointer(&a[0])), 0)
	var n uint32
	ok, _, err := d.send.Call(t.handle, uintptr(unsafe.Pointer(&p[0])), uintptr(len(p)), uintptr(unsafe.Pointer(&n)), uintptr(unsafe.Pointer(&a[0])))
	if ok == 0 {
		return fmt.Errorf("TCP reinjection: %w", err)
	}
	return nil
}
func (t *tcpRedirect) record(e *tcpEntry, event string, n int) {
	t.b.evidence.RecordBytes(e.identity, e.tuple, event, e.decision.Route.RelayID, uint64(n))
}
func (t *tcpRedirect) reflect(e *tcpEntry, p []byte, a [80]byte, reply bool) error {
	source, destination := netip.AddrPortFrom(e.tuple.Destination.Addr(), e.virtual), netip.AddrPortFrom(e.tuple.Source.Addr(), t.port)
	if reply {
		source, destination = e.tuple.Destination, e.tuple.Source
	}
	out, err := ReflectTCP(p, source, destination)
	if err != nil {
		return err
	}
	a[10] &= ^byte(2) // inbound
	err = t.send(out, a)
	if err == nil {
		event := "relay"
		if reply {
			event = "reply"
		}
		t.record(e, event, 0)
	}
	return err
}
func (t *tcpRedirect) packets(classify Classifier) {
	defer t.wg.Done()
	d := t.b.proc()
	buf := make([]byte, 65535)
	for {
		var n uint32
		var a [80]byte
		ok, _, err := d.recv.Call(t.handle, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)), uintptr(unsafe.Pointer(&n)), uintptr(unsafe.Pointer(&a[0])))
		if ok == 0 {
			if t.b.ctx.Err() == nil {
				t.b.fail(fmt.Errorf("TCP capture receive: %w", err))
			}
			return
		}
		if n == 0 || n > uint32(len(buf)) {
			continue
		}
		p := buf[:n]
		tuple, flags, _, err := ParseTCP(p)
		if err != nil {
			if err = t.send(p, a); err != nil && t.b.ctx.Err() == nil {
				t.b.fail(err)
			}
			continue
		}
		// Never expose the local interception listener as an Internet proxy.
		if a[10]&2 == 0 {
			continue
		}
		t.state.mu.Lock()
		// Service listener responses have an exact mapping; do not trust a port alone.
		e := t.state.ports[tuple.Destination.Port()]
		if tuple.Source.Port() == t.port && e != nil && tuple.Source.Addr() == e.tuple.Source.Addr() && tuple.Destination.Addr() == e.tuple.Destination.Addr() {
			e.last = time.Now()
			err = t.reflect(e, p, a, true)
			t.state.mu.Unlock()
			if err != nil && t.b.ctx.Err() == nil {
				t.b.fail(err)
			}
			continue
		}
		e = t.state.entries[tuple]
		if e != nil {
			e.last = time.Now()
			t.record(e, "captured", len(p))
			mode := e.mode
			if mode == "relay" {
				err = t.reflect(e, p, a, false)
			} else if mode == "direct" {
				err = t.send(p, a)
				t.record(e, "direct", len(p))
			} else if mode == "block" || mode == "closed" {
				t.b.drop(e.identity, e.tuple)
			}
			t.state.mu.Unlock()
			if err != nil && t.b.ctx.Err() == nil {
				t.b.fail(err)
			}
			continue
		}
		t.state.mu.Unlock()
		// Capture only new SYNs. Midstream packets never acquire a new route.
		id := ProcessIdentity{}
		t.b.associations.mu.RLock()
		association, known := t.b.associations.items[tuple]
		t.b.associations.mu.RUnlock()
		if known && IsCurrent(association.Identity) {
			id = association.Identity
		} else {
			pid, ownerErr := resolveTCPOwner(tuple)
			if ownerErr == nil {
				id, _ = t.b.cache.Lookup(pid)
			}
		}
		decision, self := ClassifyCapture(id, tuple, uint32(os.Getpid()), classify)
		mode := "direct"
		if flags&0x12 == 0x02 && !self && id.PID != 0 {
			if decision.Block {
				mode = "block"
			} else if !decision.Route.Direct() {
				mode = "opening"
			}
		}
		t.state.mu.Lock()
		e = t.state.add(tuple, id, decision, mode, time.Now())
		if e == nil {
			t.state.mu.Unlock()
			if decision.Block || decision.BlockOnFailure {
				t.b.drop(id, tuple)
			} else {
				_ = t.send(p, a)
			}
			continue
		}
		t.record(e, "captured", len(p))
		if self {
			t.record(e, "self", len(p))
		}
		if mode == "opening" {
			select {
			case t.slots <- struct{}{}:
				t.wg.Add(1)
				go t.open(e, append([]byte(nil), p...), a)
			default:
				if decision.BlockOnFailure {
					e.mode = "block"
				} else {
					e.mode = "direct"
					err = t.send(p, a)
					t.record(e, "direct", len(p))
				}
			}
		} else if mode == "direct" {
			err = t.send(p, a)
			t.record(e, "direct", len(p))
		} else {
			t.b.drop(id, tuple)
		}
		t.state.mu.Unlock()
		if err != nil && t.b.ctx.Err() == nil {
			t.b.fail(err)
		}
	}
}
func (t *tcpRedirect) open(e *tcpEntry, syn []byte, a [80]byte) {
	defer t.wg.Done()
	defer func() { <-t.slots }()
	ctx, cancel := context.WithTimeout(t.b.ctx, 5*time.Second)
	up, err := t.b.Engine.OpenTCP(ctx, e.decision.Flow)
	cancel()
	route := e.decision.Route
	if err == nil {
		for _, f := range t.b.Engine.Snapshot() {
			if f.Protocol == "tcp" && f.PID == e.identity.PID && f.Source == e.tuple.Source && f.Destination == e.tuple.Destination {
				route = f.Route
				break
			}
		}
	}
	t.state.mu.Lock()
	defer t.state.mu.Unlock()
	if t.b.ctx.Err() != nil || t.state.entries[e.tuple] != e || e.mode != "opening" {
		if up != nil {
			up.Close()
		}
		return
	}
	if err != nil || route.Direct() {
		if up != nil {
			up.Close()
		}
		if e.decision.BlockOnFailure {
			e.mode = "block"
			t.b.drop(e.identity, e.tuple)
			return
		}
		e.mode = "direct"
		e.decision.Route = route
		err = t.send(syn, a)
		t.record(e, "direct", len(syn))
	} else {
		e.upstream = up
		e.mode = "relay"
		e.decision.Route = route
		err = t.reflect(e, syn, a, false)
	}
	if err != nil {
		t.b.fail(err)
	}
}
func (t *tcpRedirect) accept() {
	defer t.wg.Done()
	for {
		c, err := t.listener.AcceptTCP()
		if err != nil {
			return
		}
		local, remote := c.LocalAddr().(*net.TCPAddr).AddrPort(), c.RemoteAddr().(*net.TCPAddr).AddrPort()
		t.state.mu.Lock()
		e := t.state.ports[remote.Port()]
		if t.b.ctx.Err() != nil || e == nil || e.mode != "relay" || e.local != nil || e.upstream == nil || local.Addr() != e.tuple.Source.Addr() || remote.Addr() != e.tuple.Destination.Addr() {
			t.state.mu.Unlock()
			c.Close()
			continue
		}
		e.local = c
		up := e.upstream
		t.wg.Add(1)
		t.state.mu.Unlock()
		go t.pipe(e, c, up)
	}
}
func (t *tcpRedirect) pipe(e *tcpEntry, local, up net.Conn) {
	defer t.wg.Done()
	// Half-close permits HTTP/1.0, uploads and protocols which wait for EOF.
	done := make(chan struct{})
	go func() {
		defer close(done)
		t.copy(e, up, local, "relay")
		if h, ok := up.(interface{ CloseWrite() error }); ok {
			_ = h.CloseWrite()
		}
	}()
	t.copy(e, local, up, "reply")
	local.Close()
	up.Close()
	<-done
	t.state.mu.Lock()
	e.mode = "closed"
	e.last = time.Now()
	e.local = nil
	e.upstream = nil
	t.state.mu.Unlock()
}
func (t *tcpRedirect) copy(e *tcpEntry, dst, src net.Conn, event string) {
	buf := make([]byte, 32*1024)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			written, werr := writeTCP(dst, buf[:n])
			if written > 0 {
				t.state.mu.Lock()
				e.last = time.Now()
				t.b.evidence.AddBytes(e.identity, e.tuple, event, e.decision.Route.RelayID, uint64(written))
				t.state.mu.Unlock()
				if event == "relay" {
					t.b.TX.Add(uint64(written))
				} else {
					t.b.RX.Add(uint64(written))
				}
			}
			if werr != nil {
				return
			}
		}
		if err != nil {
			return
		}
	}
}
func writeTCP(w io.Writer, p []byte) (int, error) {
	total := 0
	for len(p) > 0 {
		n, err := w.Write(p)
		total += n
		p = p[n:]
		if err != nil {
			return total, err
		}
		if n == 0 {
			return total, io.ErrShortWrite
		}
	}
	return total, nil
}
func (t *tcpRedirect) cleanup() {
	defer t.wg.Done()
	tick := time.NewTicker(10 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-t.b.ctx.Done():
			return
		case now := <-tick.C:
			t.state.mu.Lock()
			t.state.expire(now)
			t.state.mu.Unlock()
		}
	}
}
