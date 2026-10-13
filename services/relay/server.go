// Package relay implements an authenticated, resource-bounded UDP/TCP forwarder.
package relay

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/buzzmean/neexitlag/core/protocol"
	"github.com/buzzmean/neexitlag/core/security"
	"github.com/buzzmean/neexitlag/core/telemetry"
	"github.com/buzzmean/neexitlag/core/transport"
	"github.com/quic-go/quic-go"
	"io"
	"net"
	"net/http"
	"net/netip"
	"sync"
	"sync/atomic"
	"time"
)

type Config struct {
	TLS                   *tls.Config
	MaxSessions, MaxFlows int
	IdleTimeout           time.Duration
	Allow                 []netip.Prefix
	PacketsPerSecond      int
	TicketVerifier        *security.TicketVerifier
}
type Metrics struct {
	Sessions, Flows                                                                  atomic.Int64
	UDPPackets, TCPConnections, RXBytes, TXBytes, Drops, RejectedFlows, AuthFailures atomic.Uint64
	LocalDrops                                                                       telemetry.Counters
}
type Server struct {
	listener    *quic.Listener
	config      Config
	Metrics     Metrics
	sem         chan struct{}
	mu          sync.Mutex
	connections map[*quic.Conn]struct{}
	wg          sync.WaitGroup
	done        chan struct{}
	once        sync.Once
}

func Listen(addr string, c Config) (*Server, error) {
	if c.TLS == nil || c.TLS.ClientAuth != tls.RequireAndVerifyClientCert || c.TLS.ClientCAs == nil {
		return nil, errors.New("relay requires verified mutual TLS")
	}
	if c.MaxSessions <= 0 {
		c.MaxSessions = 128
	}
	if c.MaxFlows <= 0 {
		c.MaxFlows = 64
	}
	if c.MaxFlows > 64 {
		return nil, errors.New("max flows exceeds stream limit 64")
	}
	if c.IdleTimeout <= 0 {
		c.IdleTimeout = 60 * time.Second
	}
	if c.PacketsPerSecond <= 0 {
		c.PacketsPerSecond = 2000
	}
	q := transport.QUICConfig()
	q.MaxIncomingStreams = int64(c.MaxFlows)
	if c.TicketVerifier != nil {
		c.TLS = c.TLS.Clone()
		c.TLS.NextProtos = []string{protocol.TicketALPN}
	}
	l, e := quic.ListenAddr(addr, c.TLS, q)
	if e != nil {
		return nil, e
	}
	s := &Server{listener: l, config: c, sem: make(chan struct{}, c.MaxSessions), connections: make(map[*quic.Conn]struct{}), done: make(chan struct{})}
	s.wg.Add(1)
	go s.accept()
	return s, nil
}
func (s *Server) Addr() net.Addr { return s.listener.Addr() }
func (s *Server) Close() error {
	s.once.Do(func() {
		close(s.done)
		s.listener.Close()
		s.mu.Lock()
		for c := range s.connections {
			c.CloseWithError(0, "shutdown")
		}
		s.mu.Unlock()
	})
	s.wg.Wait()
	return nil
}
func (s *Server) accept() {
	defer s.wg.Done()
	for {
		c, e := s.listener.Accept(context.Background())
		if e != nil {
			return
		}
		select {
		case s.sem <- struct{}{}:
		default:
			c.CloseWithError(2, "capacity")
			continue
		}
		s.mu.Lock()
		select {
		case <-s.done:
			s.mu.Unlock()
			<-s.sem
			c.CloseWithError(0, "shutdown")
			return
		default:
		}
		s.connections[c] = struct{}{}
		s.wg.Add(1)
		s.mu.Unlock()
		s.Metrics.Sessions.Add(1)
		go func() {
			defer s.wg.Done()
			defer func() { s.mu.Lock(); delete(s.connections, c); s.mu.Unlock(); <-s.sem; s.Metrics.Sessions.Add(-1) }()
			x := &session{server: s, conn: c, flows: make(map[uint32]*flow), tokens: float64(s.config.PacketsPerSecond), refill: time.Now()}
			x.run()
		}()
	}
}

// Explicit CIDRs restrict deployment egress. Without them only globally routable unicast is allowed.
func (s *Server) allowed(a netip.Addr) bool {
	a = a.Unmap()
	if a.IsUnspecified() || a.IsMulticast() {
		return false
	}
	if len(s.config.Allow) > 0 {
		for _, p := range s.config.Allow {
			if p.Contains(a) {
				return true
			}
		}
		return false
	}
	if a.Is6() && !netip.MustParsePrefix("2000::/3").Contains(a) {
		return false
	}
	if !a.IsGlobalUnicast() || a.IsPrivate() || a.IsLoopback() || a.IsLinkLocalUnicast() {
		return false
	}
	for _, p := range reserved {
		if p.Contains(a) {
			return false
		}
	}
	return true
}

var reserved = []netip.Prefix{netip.MustParsePrefix("0.0.0.0/8"), netip.MustParsePrefix("100.64.0.0/10"), netip.MustParsePrefix("192.0.0.0/24"), netip.MustParsePrefix("192.0.2.0/24"), netip.MustParsePrefix("198.18.0.0/15"), netip.MustParsePrefix("198.51.100.0/24"), netip.MustParsePrefix("203.0.113.0/24"), netip.MustParsePrefix("240.0.0.0/4"), netip.MustParsePrefix("2001:db8::/32"), netip.MustParsePrefix("2001::/23"), netip.MustParsePrefix("3fff::/20"), netip.MustParsePrefix("192.88.99.0/24"), netip.MustParsePrefix("64:ff9b::/96"), netip.MustParsePrefix("2002::/16")}

type flow struct {
	udp    *net.UDPConn
	tcp    net.Conn
	stream *quic.Stream
	last   atomic.Int64
	once   sync.Once
}

func (f *flow) close() {
	f.once.Do(func() {
		if f.udp != nil {
			f.udp.Close()
		}
		if f.tcp != nil {
			f.tcp.Close()
		}
		f.stream.CancelRead(0)
		f.stream.Close()
	})
}

type session struct {
	server       *Server
	conn         *quic.Conn
	mu           sync.Mutex
	flows        map[uint32]*flow
	wg           sync.WaitGroup
	tokens       float64
	refill       time.Time
	deviceID     string
	leaseExpires time.Time
}

func (x *session) run() {
	if x.server.config.TicketVerifier != nil {
		if e := x.authenticate(); e != nil {
			x.server.Metrics.AuthFailures.Add(1)
			x.conn.CloseWithError(1, "ticket rejected")
			return
		}
	}
	x.wg.Add(2)
	go func() { defer x.wg.Done(); x.datagrams() }()
	go func() { defer x.wg.Done(); x.expire() }()
	for {
		s, e := x.conn.AcceptStream(context.Background())
		if e != nil {
			break
		}
		x.wg.Add(1)
		go func() { defer x.wg.Done(); x.open(s) }()
	}
	x.conn.CloseWithError(0, "session ended")
	x.mu.Lock()
	for _, f := range x.flows {
		f.close()
	}
	x.mu.Unlock()
	x.wg.Wait()
}
func (x *session) remove(id uint32, f *flow) {
	f.close()
	x.mu.Lock()
	if x.flows[id] == f {
		delete(x.flows, id)
		x.server.Metrics.Flows.Add(-1)
	}
	x.mu.Unlock()
}
func (x *session) open(s *quic.Stream) {
	s.SetDeadline(time.Now().Add(5 * time.Second))
	o, e := protocol.ReadOpen(s)
	if e != nil {
		s.CancelRead(1)
		s.CancelWrite(1)
		return
	}
	reject := func() { x.server.Metrics.RejectedFlows.Add(1); s.Write([]byte{1}); s.Close(); s.CancelRead(1) }
	if !x.server.allowed(o.Destination.Addr()) {
		reject()
		return
	}
	f := &flow{stream: s}
	f.last.Store(time.Now().UnixNano())
	x.mu.Lock()
	if len(x.flows) >= x.server.config.MaxFlows || x.flows[o.ID] != nil {
		x.mu.Unlock()
		reject()
		return
	}
	x.flows[o.ID] = f
	x.server.Metrics.Flows.Add(1)
	x.mu.Unlock()
	defer x.remove(o.ID, f)
	ctx, cancel := context.WithTimeout(x.conn.Context(), 3*time.Second)
	defer cancel()
	d := net.Dialer{}
	c, e := d.DialContext(ctx, map[byte]string{protocol.UDP: "udp", protocol.TCP: "tcp"}[o.Kind], o.Destination.String())
	if e != nil {
		reject()
		return
	}
	x.mu.Lock()
	if x.conn.Context().Err() != nil {
		x.mu.Unlock()
		c.Close()
		return
	}
	if o.Kind == protocol.UDP {
		f.udp = c.(*net.UDPConn)
	} else {
		f.tcp = c
	}
	x.mu.Unlock()
	if _, e = s.Write([]byte{0}); e != nil {
		return
	}
	s.SetDeadline(time.Time{})
	if o.Kind == protocol.UDP {
		done := make(chan struct{})
		go func() { defer close(done); x.readUDP(o.ID, f) }()
		var b [1]byte
		_, _ = s.Read(b[:])
		f.close()
		<-done
	} else {
		x.server.Metrics.TCPConnections.Add(1)
		done := make(chan struct{})
		go func() {
			defer close(done)
			n, _ := io.Copy(activityWriter{s, &f.last}, c)
			x.server.Metrics.TXBytes.Add(uint64(n))
			s.Close()
		}()
		n, _ := io.Copy(activityWriter{c, &f.last}, s)
		x.server.Metrics.RXBytes.Add(uint64(n))
		if t, ok := c.(*net.TCPConn); ok {
			t.CloseWrite()
		}
		<-done
	}
}

type activityWriter struct {
	io.Writer
	last *atomic.Int64
}

func (w activityWriter) Write(b []byte) (int, error) {
	n, e := w.Writer.Write(b)
	if n > 0 {
		w.last.Store(time.Now().UnixNano())
	}
	return n, e
}
func (x *session) readUDP(id uint32, f *flow) {
	defer f.close()
	b := make([]byte, protocol.MaxPayload+1)
	for {
		n, e := f.udp.Read(b)
		x.server.Metrics.LocalDrops.Attempt(telemetry.CaptureProcessing)
		if e != nil {
			return
		}
		if n > protocol.MaxPayload {
			x.server.Metrics.Drops.Add(1)
			x.server.Metrics.LocalDrops.Drop(telemetry.CaptureProcessing)
			continue
		}
		f.last.Store(time.Now().UnixNano())
		p, _ := protocol.Encode(id, b[:n])
		x.server.Metrics.LocalDrops.Attempt(telemetry.SocketSend)
		if e = x.conn.SendDatagram(p); e != nil {
			x.server.Metrics.Drops.Add(1)
			x.server.Metrics.LocalDrops.Drop(telemetry.SocketSend)
		} else {
			x.server.Metrics.TXBytes.Add(uint64(n))
		}
	}
}
func (x *session) datagrams() {
	for {
		b, e := x.conn.ReceiveDatagram(context.Background())
		if e != nil {
			return
		}
		x.server.Metrics.LocalDrops.Attempt(telemetry.Decode)
		x.server.Metrics.LocalDrops.Attempt(telemetry.UnknownFlow)
		id, p, e := protocol.Decode(b)
		if e != nil {
			x.server.Metrics.Drops.Add(1)
			x.server.Metrics.LocalDrops.Drop(telemetry.Decode)
			continue
		}
		now := time.Now()
		dt := now.Sub(x.refill).Seconds()
		x.refill = now
		x.tokens += dt * float64(x.server.config.PacketsPerSecond)
		if x.tokens > float64(x.server.config.PacketsPerSecond) {
			x.tokens = float64(x.server.config.PacketsPerSecond)
		}
		x.server.Metrics.LocalDrops.Attempt(telemetry.Policy)
		if x.tokens < 1 {
			x.server.Metrics.Drops.Add(1)
			x.server.Metrics.LocalDrops.Drop(telemetry.Policy)
			continue
		}
		x.tokens--
		x.mu.Lock()
		f := x.flows[id]
		var c *net.UDPConn
		if f != nil {
			c = f.udp
		}
		x.mu.Unlock()
		if c == nil {
			x.server.Metrics.Drops.Add(1)
			x.server.Metrics.LocalDrops.Drop(telemetry.UnknownFlow)
			continue
		}
		c.SetWriteDeadline(time.Now().Add(20 * time.Millisecond))
		x.server.Metrics.LocalDrops.Attempt(telemetry.SocketSend)
		if _, e = c.Write(p); e != nil {
			x.server.Metrics.Drops.Add(1)
			x.server.Metrics.LocalDrops.Drop(telemetry.SocketSend)
			continue
		}
		f.last.Store(now.UnixNano())
		x.server.Metrics.UDPPackets.Add(1)
		x.server.Metrics.RXBytes.Add(uint64(len(p)))
	}
}
func (x *session) expire() {
	interval := x.server.config.IdleTimeout / 4
	if interval < time.Millisecond {
		interval = time.Millisecond
	}
	if interval > time.Second {
		interval = time.Second
	}
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-x.conn.Context().Done():
			return
		case now := <-t.C:
			if x.server.config.TicketVerifier != nil && (now.After(x.leaseExpires) || x.server.config.TicketVerifier.Revoked(x.deviceID)) {
				x.conn.CloseWithError(1, "session lease expired or revoked")
				return
			}
			x.mu.Lock()
			for _, f := range x.flows {
				if (f.udp != nil || f.tcp != nil) && now.Sub(time.Unix(0, f.last.Load())) > x.server.config.IdleTimeout {
					f.close()
				}
			}
			x.mu.Unlock()
		}
	}
}
func (s *Server) Handler() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-s.done:
			http.Error(w, "stopped", 503)
		default:
			w.Write([]byte("ok\n"))
		}
	})
	m.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		fmt.Fprintf(w, "neexitlag_relay_health 1\nneexitlag_active_sessions %d\nneexitlag_active_flows %d\nneexitlag_udp_packets_total %d\nneexitlag_tcp_connections_total %d\nneexitlag_rx_bytes_total %d\nneexitlag_tx_bytes_total %d\nneexitlag_packet_drops_total %d\nneexitlag_rejected_flows_total %d\nneexitlag_auth_failures_total %d\n", s.Metrics.Sessions.Load(), s.Metrics.Flows.Load(), s.Metrics.UDPPackets.Load(), s.Metrics.TCPConnections.Load(), s.Metrics.RXBytes.Load(), s.Metrics.TXBytes.Load(), s.Metrics.Drops.Load(), s.Metrics.RejectedFlows.Load(), s.Metrics.AuthFailures.Load())
		for name, metric := range s.Metrics.LocalDrops.Snapshot() {
			fmt.Fprintf(w, "neexitlag_relay_%s_total %d\n", name, metric.Count)
		}
	})
	return m
}
