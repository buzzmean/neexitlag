//go:build windows

package windowsplatform

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/buzzmean/neexitlag/core/flowpolicy"
	"github.com/buzzmean/neexitlag/core/localstrategy"
	"github.com/buzzmean/neexitlag/core/routing"
	"golang.org/x/sys/windows"
	"io"
	"net"
	"net/netip"
	"os"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

type delegatedTCP struct {
	owner    *tcpEntry
	decision Decision
}
type tcpRedirect struct {
	delegated map[Tuple]delegatedTCP
	b         *WinDivertBackend
	handle    uintptr
	listener  *net.TCPListener
	port      uint16
	state     *tcpState
	slots     chan struct{}
	wg        sync.WaitGroup
}

func (b *WinDivertBackend) startTCP(d *divertDLL) (*tcpRedirect, error) {
	l, err := net.ListenTCP("tcp4", &net.TCPAddr{IP: net.IPv4zero})
	if err != nil {
		return nil, err
	}
	port := uint16(l.Addr().(*net.TCPAddr).Port)
	// Inbound physical connections to the internal listener are dropped. Reflected
	// packets injected below this handle are the only admitted listener traffic.
	filter := fmt.Sprintf("ip and tcp and !impostor and ((outbound and !loopback) or (inbound and (tcp.DstPort == %d or (tcp.Syn and tcp.Ack))))", port)
	h, err := d.openHandle(filter, 0, 0)
	if err != nil {
		l.Close()
		return nil, err
	}
	d.setParam.Call(h, 0, 2048)
	d.setParam.Call(h, 1, 100)
	d.setParam.Call(h, 2, 4*1024*1024)
	return &tcpRedirect{b: b, handle: h, listener: l, port: port, state: newTCPState(8192), delegated: map[Tuple]delegatedTCP{}, slots: make(chan struct{}, 128)}, nil
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
		if err := e.localBuffer.Flush(e.decision, &e.localState, e.identity, e.tuple, time.Now(), "NO_TRANSFORM", t.b.evidence); err != nil {
			t.b.fail(err)
		}
		if e.upstream != nil {
			e.upstream.Close()
		}
		if e.local != nil {
			e.local.Close()
		}
	}
	t.state.mu.Unlock()
	t.wg.Wait()
	// Shutdown(RECV) may drain queued captures after the first flush. With all
	// producers stopped, release any late ClientHello before closing SEND.
	t.state.mu.Lock()
	for _, e := range t.state.entries {
		if err := e.localBuffer.Flush(e.decision, &e.localState, e.identity, e.tuple, time.Now(), "NO_TRANSFORM", t.b.evidence); err != nil {
			t.b.fail(err)
		}
	}
	t.state.mu.Unlock()
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
	if n != uint32(len(p)) {
		return fmt.Errorf("TCP reinjection short packet: %d/%d", n, len(p))
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
		event := "captured_proxy"
		if !e.decision.Deferred {
			event = "relay"
			if reply {
				event = "reply"
			}
			if e.decision.Route.Direct() {
				event = "direct_proxy"
			}
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
		// External traffic to the interception listener is never admitted.
		// Native SYN/ACK observation shares this handle; it is reinjected
		// unchanged and only confirms an exact existing DIRECT+ handshake.
		if a[10]&2 == 0 {
			if tuple.Destination.Port() != t.port {
				reverse := Tuple{Protocol: "tcp", Source: tuple.Destination, Destination: tuple.Source}
				t.state.mu.Lock()
				if e := t.state.entries[reverse]; e != nil && e.mode == "direct" && e.decision.Route.Local() && p[0] == 0x45 && flags&0x3f == 0x12 && 20+int(p[32]>>4)*4 == len(p) {
					e.localState.ConfirmHandshake(binary.BigEndian.Uint32(p[28:]), binary.BigEndian.Uint32(p[24:]))
				}
				t.state.mu.Unlock()
				if err = t.send(p, a); err != nil && t.b.ctx.Err() == nil {
					t.b.fail(err)
				}
			}
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
		sequence := binary.BigEndian.Uint32(p[int(p[0]&15)*4+4:])
		if old := t.state.entries[tuple]; old != nil && flags&0x12 == 2 && sequence != old.initialSequence {
			if flushErr := old.localBuffer.Flush(old.decision, &old.localState, old.identity, old.tuple, time.Now(), "NO_TRANSFORM", t.b.evidence); flushErr != nil {
				t.b.fail(flushErr)
			}
		}
		e = t.state.existing(tuple, flags, sequence, time.Now())
		if e != nil {
			e.last = time.Now()
			t.record(e, "captured", len(p))
			if e.domainInspections < 8 && p[0] == 0x45 {
				e.domainInspections++
				h := 20 + int(p[32]>>4)*4
				if h <= len(p) {
					t.b.evidence.RecordDomain(e.identity, e.tuple, localstrategy.PacketDomain(p[h:]))
				}
			}
			mode := e.mode
			if mode == "relay" {
				err = t.reflect(e, p, a, false)
			} else if mode == "direct" {
				err = e.localBuffer.Push(e.decision, &e.localState, e.identity, e.tuple, p, func(q []byte) error { return t.send(q, a) }, t.b.evidence, time.Now())
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
		t.state.mu.Lock()
		delegated, owned := t.delegated[tuple]
		t.state.mu.Unlock()
		if owned {
			id = delegated.owner.identity
			decision = delegated.decision
			self = false
		}
		mode := "direct"
		if flags&0x12 != 0x02 || self {
			decision.Local = nil
			decision.Route = routing.Route{}
			decision.Block = false
			decision.BlockOnFailure = false
			decision.Deferred = false
			decision.ResolveDomain = nil
			if !self {
				decision.Policy = flowpolicy.Result{Action: flowpolicy.Direct, Fallback: flowpolicy.FailOpen, Source: "existing_tcp", Reason: "pre-existing TCP connection is not reclassified"}
			}
		}
		if flags&0x12 == 0x02 && !self && id.PID != 0 {
			if decision.Block {
				mode = "block"
			} else if !decision.Route.Direct() || decision.Deferred {
				mode = "opening"
			}
		}
		t.state.mu.Lock()
		t.b.evidence.BeginFlow(id, tuple)
		t.b.evidence.RecordPolicy(id, tuple, decision, "")
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
		e.initialSequence = sequence
		t.record(e, "captured", len(p))
		if self {
			t.record(e, "self", len(p))
		}
		if mode == "opening" {
			select {
			case t.slots <- struct{}{}:
				t.wg.Add(1)
				if decision.Deferred {
					t.wg.Done()
					e.mode = "relay"
					err = t.reflect(e, p, a, false)
					<-t.slots
				} else {
					go t.open(e, append([]byte(nil), p...), a)
				}
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
			err = e.localBuffer.Push(e.decision, &e.localState, e.identity, e.tuple, p, func(q []byte) error { return t.send(q, a) }, t.b.evidence, time.Now())
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
			t.b.evidence.RecordPolicy(e.identity, e.tuple, e.decision, "RELAY_OPEN_FAILED")
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
		if t.b.ctx.Err() != nil || e == nil || e.mode != "relay" || e.local != nil || (e.upstream == nil && !e.decision.Deferred) || local.Addr() != e.tuple.Source.Addr() || remote.Addr() != e.tuple.Destination.Addr() {
			t.state.mu.Unlock()
			c.Close()
			continue
		}
		e.local = c
		up := e.upstream
		t.wg.Add(1)
		t.state.mu.Unlock()
		if e.decision.Deferred {
			select {
			case t.slots <- struct{}{}:
				go t.deferredOpen(e, c)
			default:
				c.Close()
				t.state.mu.Lock()
				t.state.retire(e, time.Now())
				t.state.mu.Unlock()
				t.wg.Done()
			}
		} else {
			go t.pipe(e, c, up)
		}
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
	if e.mode != "closed" {
		e.mode = "closed"
		e.last = time.Now()
	}
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
				actualEvent := event
				if e.decision.Route.Direct() {
					actualEvent = "direct_proxy"
				}
				t.b.evidence.AddBytes(e.identity, e.tuple, actualEvent, e.decision.Route.RelayID, uint64(written))
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
	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-t.b.ctx.Done():
			return
		case now := <-tick.C:
			t.state.mu.Lock()
			for _, e := range t.state.entries {
				if e.localBuffer.Expired(now) {
					if err := e.localBuffer.Flush(e.decision, &e.localState, e.identity, e.tuple, now, "REASSEMBLY_TIMEOUT", t.b.evidence); err != nil {
						t.b.fail(err)
					}
				}
			}
			t.state.expire(now)
			t.b.evidence.pruneAliases(func(tuple Tuple) bool { return t.state.entries[tuple] != nil })
			t.state.mu.Unlock()
		}
	}
}

// The app-side handshake has completed locally; upstream is opened exactly once
// after outer SNI. A failure closes this connection, never retries its TLS bytes
// on a different egress. Only a future connection can choose a different route.
func (t *tcpRedirect) deferredOpen(e *tcpEntry, c net.Conn) {
	held := true
	defer func() {
		if held {
			<-t.slots
		}
	}()
	data, domain, reason := ReadDeferredHello(c)
	if reason == "CLIENT_CLOSED_BEFORE_DATA" {
		t.state.mu.Lock()
		t.state.retire(e, time.Now())
		t.state.mu.Unlock()
		t.wg.Done()
		return
	}
	d := e.decision.ResolveDomain(domain)
	d.Deferred = false
	d.ResolveDomain = nil
	if domain == "" && d.BlockOnFailure {
		d.Block = true
	}
	if d.Block {
		d.Policy.Action = flowpolicy.Block
		if d.Policy.Fallback == flowpolicy.FailOpen {
			d.Policy.Fallback = flowpolicy.FailBlock
		}
	}
	t.state.mu.Lock()
	if t.b.ctx.Err() != nil || e.mode == "closed" {
		t.state.mu.Unlock()
		c.Close()
		t.wg.Done()
		return
	}
	e.decision = d
	t.b.evidence.RecordPolicy(e.identity, e.tuple, d, reason)
	t.b.evidence.RecordDomain(e.identity, e.tuple, domain)
	t.state.mu.Unlock()
	if d.Block {
		t.state.mu.Lock()
		t.b.drop(e.identity, e.tuple)
		t.state.retire(e, time.Now())
		t.state.mu.Unlock()
		t.wg.Done()
		return
	}
	requested := d.Route
	ctx, cancel := context.WithTimeout(t.b.ctx, 5*time.Second)
	var up net.Conn
	var err error
	if d.Route.Local() {
		up, err = t.dialLocalStrategy(ctx, e, d)
	} else {
		up, err = t.b.Engine.OpenTCP(ctx, d.Flow)
	}
	cancel()
	if err == nil && !d.Route.Local() {
		for _, f := range t.b.Engine.Snapshot() {
			if f.Protocol == "tcp" && f.PID == e.identity.PID && f.Source == e.tuple.Source && f.Destination == e.tuple.Destination {
				d.Route = f.Route
				break
			}
		}
	}
	t.state.mu.Lock()
	e.decision = d
	if err != nil || t.b.ctx.Err() != nil || e.mode == "closed" {
		if up != nil {
			up.Close()
		}
		t.b.evidence.RecordPolicy(e.identity, e.tuple, d, "UPSTREAM_OPEN_FAILED")
		t.b.drop(e.identity, e.tuple)
		t.state.retire(e, time.Now())
		t.state.mu.Unlock()
		t.wg.Done()
		return
	}
	e.upstream = up
	if !requested.Direct() && d.Route.Direct() {
		t.b.evidence.RecordPolicy(e.identity, e.tuple, d, "RELAY_FELL_BACK_DIRECT")
	}
	t.b.evidence.RecordUpstreamOpened(e.identity, e.tuple, d.Route)
	t.state.mu.Unlock()
	if len(data) > 0 {
		n, werr := writeTCP(up, data)
		if n > 0 {
			event := "direct_proxy"
			if !d.Route.Direct() {
				event = "relay"
			}
			t.b.evidence.AddBytes(e.identity, e.tuple, event, d.Route.RelayID, uint64(n))
		}
		if werr != nil {
			c.Close()
			up.Close()
			t.state.mu.Lock()
			t.state.retire(e, time.Now())
			t.state.mu.Unlock()
			t.wg.Done()
			return
		}
	}
	data = nil
	<-t.slots
	held = false
	t.pipe(e, c, up)
}
func (t *tcpRedirect) dialLocalStrategy(ctx context.Context, e *tcpEntry, d Decision) (net.Conn, error) {
	var bound Tuple
	dialer := net.Dialer{Control: func(network, address string, raw syscall.RawConn) error {
		var bindErr error
		err := raw.Control(func(fd uintptr) {
			source := e.tuple.Source.Addr().As4()
			a := &windows.SockaddrInet4{Addr: source}
			if bindErr = windows.Bind(windows.Handle(fd), a); bindErr != nil {
				return
			}
			sa, er := windows.Getsockname(windows.Handle(fd))
			if er != nil {
				bindErr = er
				return
			}
			v, ok := sa.(*windows.SockaddrInet4)
			if !ok {
				bindErr = fmt.Errorf("unsupported upstream socket")
				return
			}
			bound = Tuple{Protocol: "tcp", Source: netip.AddrPortFrom(e.tuple.Source.Addr(), uint16(v.Port)), Destination: e.tuple.Destination}
			t.state.mu.Lock()
			defer t.state.mu.Unlock()
			if len(t.delegated) >= 128 || !t.b.evidence.BindUpstream(e.identity, bound, e.tuple) {
				bindErr = fmt.Errorf("delegated TCP limit")
				return
			}
			d.Flow = routing.Flow{}
			d.Deferred = false
			d.ResolveDomain = nil
			t.delegated[bound] = delegatedTCP{owner: e, decision: d}
		})
		if err != nil {
			return err
		}
		return bindErr
	}}
	c, err := dialer.DialContext(ctx, "tcp4", e.tuple.Destination.String())
	if err != nil {
		t.state.mu.Lock()
		delete(t.delegated, bound)
		t.state.mu.Unlock()
		t.b.evidence.UnbindUpstream(e.identity, bound)
		return nil, err
	}
	return &delegatedConn{Conn: c, release: func() { t.state.mu.Lock(); delete(t.delegated, bound); t.state.mu.Unlock() }}, nil
}

type delegatedConn struct {
	net.Conn
	once    sync.Once
	release func()
}

func (c *delegatedConn) Close() error { err := c.Conn.Close(); c.once.Do(c.release); return err }
func (c *delegatedConn) CloseWrite() error {
	if v, ok := c.Conn.(interface{ CloseWrite() error }); ok {
		return v.CloseWrite()
	}
	return nil
}
