// Package browserproxy provides a Chrome-only origin-bound transport. It never
// terminates TLS, reads HTTPS content, replays requests or changes system proxy.
package browserproxy

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/buzzmean/neexitlag/core/flowpolicy"
	"golang.org/x/net/idna"
	"io"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Owner struct {
	PID     uint32 `json:"pid"`
	Birth   uint64 `json:"creation_time"`
	Exe     string `json:"exe"`
	Profile string `json:"profile_id"`
}
type Request struct {
	Owner  Owner
	Source netip.AddrPort
	Host   string
	Port   uint16
	Secure bool
}
type Trial struct {
	Route            string `json:"route"`
	Outcome          string `json:"outcome"`
	HTTPStatus       int    `json:"http_status"`
	TransformApplied bool   `json:"transform_applied"`
	DurationMS       int64  `json:"duration_ms"`
}
type Plan struct {
	Trials      []Trial           `json:"trials,omitempty"`
	Network     string            `json:"network"`
	Enabled     bool              `json:"optimizing"`
	Policy      flowpolicy.Result `json:"policy"`
	Destination netip.AddrPort    `json:"destination"`
	Probe       string            `json:"probe"`
	HTTPStatus  int               `json:"http_status,omitempty"`
}
type Record struct {
	DirectPlusApplied bool      `json:"direct_plus_applied"`
	ID                uint64    `json:"id"`
	Owner             Owner     `json:"owner"`
	Domain            string    `json:"domain"`
	Port              uint16    `json:"port"`
	Source            string    `json:"source"`
	Plan              Plan      `json:"plan"`
	Started           time.Time `json:"started"`
	Active            bool      `json:"active"`
	Result            string    `json:"result"`
	Sent              uint64    `json:"sent_bytes"`
	Received          uint64    `json:"received_bytes"`
}
type Snapshot struct {
	Truncated bool     `json:"truncated"`
	Listening bool     `json:"listening"`
	Address   string   `json:"address"`
	Lost      uint64   `json:"lost"`
	Records   []Record `json:"records"`
}
type Config struct {
	Address          string // Empty allocates a port; otherwise fixed loopback only.
	TransformApplied func(Request, Plan) bool
	Authorize        func(context.Context, netip.AddrPort, netip.AddrPort) (Owner, error)
	Plan             func(context.Context, Request) (Plan, error)
	Open             func(context.Context, Request, Plan) (net.Conn, error)
}
type Server struct {
	cfg        Config
	l          net.Listener
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	slots      chan struct{}
	mu         sync.Mutex
	next, lost uint64
	records    map[uint64]Record
	conns      map[net.Conn]bool
	closed     bool
}

func Listen(parent context.Context, cfg Config) (*Server, error) {
	if cfg.Authorize == nil || cfg.Plan == nil || cfg.Open == nil {
		return nil, errors.New("browser transport callbacks required")
	}
	address := cfg.Address
	if address == "" {
		address = "127.0.0.1:0"
	}
	h, _, parseErr := net.SplitHostPort(address)
	if parseErr != nil || h != "127.0.0.1" {
		return nil, errors.New("browser transport must bind loopback")
	}
	l, e := net.Listen("tcp4", address)
	if e != nil {
		return nil, e
	}
	ctx, cancel := context.WithCancel(parent)
	s := &Server{cfg: cfg, l: l, ctx: ctx, cancel: cancel, slots: make(chan struct{}, 256), records: map[uint64]Record{}, conns: map[net.Conn]bool{}}
	s.wg.Add(1)
	go s.serve()
	return s, nil
}
func (s *Server) Address() string { return s.l.Addr().String() }
func (s *Server) CloseConnections() {
	s.mu.Lock()
	all := make([]net.Conn, 0, len(s.conns))
	for c := range s.conns {
		all = append(all, c)
	}
	s.mu.Unlock()
	for _, c := range all {
		c.Close()
	}
}
func (s *Server) Close() {
	s.cancel()
	s.l.Close()
	s.CloseConnections()
	s.wg.Wait()
	s.mu.Lock()
	s.closed = true
	s.mu.Unlock()
}
func (s *Server) Snapshot() Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := Snapshot{Listening: !s.closed, Address: s.Address(), Lost: s.lost, Records: make([]Record, 0, len(s.records))}
	for _, r := range s.records {
		if r.Plan.Policy.Action == flowpolicy.DirectPlus && !r.DirectPlusApplied && s.cfg.TransformApplied != nil {
			source, _ := netip.ParseAddrPort(r.Source)
			r.DirectPlusApplied = s.cfg.TransformApplied(Request{Owner: r.Owner, Source: source, Host: r.Domain, Port: r.Port}, r.Plan)
			if r.DirectPlusApplied {
				s.records[r.ID] = r
			}
		}
		out.Records = append(out.Records, r)
	}
	return out
}
func (s *Server) serve() {
	defer s.wg.Done()
	for {
		c, e := s.l.Accept()
		if e != nil {
			return
		}
		select {
		case s.slots <- struct{}{}:
			s.mu.Lock()
			s.conns[c] = true
			s.mu.Unlock()
			s.wg.Add(1)
			go func() {
				defer s.wg.Done()
				defer func() { <-s.slots; s.mu.Lock(); delete(s.conns, c); s.mu.Unlock(); c.Close() }()
				s.handle(c)
			}()
		default:
			c.Close()
		}
	}
}
func CanonicalHost(host string) (string, error) {
	host = strings.TrimSuffix(strings.ToLower(host), ".")
	if ip, e := netip.ParseAddr(strings.Trim(host, "[]")); e == nil {
		return ip.Unmap().String(), nil
	}
	ascii, e := idna.Lookup.ToASCII(host)
	if e != nil || len(ascii) == 0 || len(ascii) > 253 {
		return "", errors.New("invalid origin hostname")
	}
	for _, v := range strings.Split(ascii, ".") {
		if len(v) == 0 || len(v) > 63 || v[0] == '-' || v[len(v)-1] == '-' {
			return "", errors.New("invalid hostname label")
		}
		for _, c := range v {
			if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
				return "", errors.New("invalid hostname character")
			}
		}
	}
	return ascii, nil
}
func authority(value string, defaultPort uint16) (string, uint16, error) {
	h, p, e := net.SplitHostPort(value)
	if e != nil {
		h = value
		p = strconv.Itoa(int(defaultPort))
	}
	host, e := CanonicalHost(h)
	if e != nil {
		return "", 0, e
	}
	port, err := strconv.ParseUint(p, 10, 16)
	if err != nil || port == 0 {
		return "", 0, errors.New("invalid origin port")
	}
	return host, uint16(port), nil
}
func readHeaders(b *bufio.Reader) ([]byte, error) {
	var h []byte
	for len(h) < 16384 {
		line, e := b.ReadSlice('\n')
		if e != nil {
			return nil, e
		}
		h = append(h, line...)
		if bytes.HasSuffix(h, []byte("\r\n\r\n")) {
			return h, nil
		}
	}
	return nil, errors.New("proxy header bound")
}
func (s *Server) handle(c net.Conn) {
	ctx, cancel := context.WithTimeout(s.ctx, 8*time.Second)
	defer cancel()
	c.SetDeadline(time.Now().Add(8 * time.Second))
	b := bufio.NewReaderSize(c, 8192)
	head, e := readHeaders(b)
	if e != nil {
		return
	}
	parsed := bufio.NewReader(io.MultiReader(bytes.NewReader(head), b))
	request, e := http.ReadRequest(parsed)
	if e != nil {
		return
	}
	secure := request.Method == "CONNECT"
	target := request.Host
	defaultPort := uint16(80)
	if secure {
		target = request.RequestURI
		defaultPort = 443
	} else if request.URL.Scheme != "http" || request.URL.User != nil {
		writeError(c, 400)
		return
	}
	host, port, e := authority(target, defaultPort)
	if e != nil {
		writeError(c, 400)
		return
	}
	source, e := netip.ParseAddrPort(c.RemoteAddr().String())
	if e != nil || !source.Addr().IsLoopback() {
		writeError(c, 403)
		return
	}
	local, e := netip.ParseAddrPort(c.LocalAddr().String())
	if e != nil {
		return
	}
	owner, e := s.cfg.Authorize(ctx, source, local)
	if e != nil {
		writeError(c, 403)
		return
	}
	q := Request{Owner: owner, Source: source, Host: host, Port: port, Secure: secure}
	plan, e := s.cfg.Plan(ctx, q)
	if e != nil || plan.Policy.Action == flowpolicy.Block {
		writeError(c, 502)
		return
	}
	s.mu.Lock()
	s.next++
	id := s.next
	if len(s.records) >= 512 {
		old := uint64(0)
		for n, r := range s.records {
			if !r.Active && (old == 0 || n < old) {
				old = n
			}
		}
		if old != 0 {
			delete(s.records, old)
			s.lost++
		}
	}
	if len(s.records) >= 512 {
		s.mu.Unlock()
		writeError(c, 503)
		return
	}
	s.records[id] = Record{ID: id, Owner: owner, Domain: host, Port: port, Source: source.String(), Plan: plan, Started: time.Now(), Active: true, Result: "OPEN_PENDING"}
	s.mu.Unlock()
	finish := func(result string, tx, rx uint64) {
		s.mu.Lock()
		r := s.records[id]
		r.Active = false
		r.Result = result
		r.Sent = tx
		r.Received = rx
		s.records[id] = r
		s.mu.Unlock()
	}
	upstream, e := s.cfg.Open(ctx, q, plan)
	if e != nil {
		finish("UPSTREAM_OPEN_FAILED", 0, 0)
		writeError(c, 502)
		return
	}
	s.mu.Lock()
	s.conns[upstream] = true
	s.mu.Unlock()
	defer func() { upstream.Close(); s.mu.Lock(); delete(s.conns, upstream); s.mu.Unlock() }()
	// Once established, this exact upstream is immutable. No route fallback or
	// replay is permitted after any client bytes have been sent.
	c.SetDeadline(time.Time{})
	upstream.SetDeadline(time.Time{})
	if !secure {
		request.RequestURI = ""
		request.Close = true
		request.Header.Del("Proxy-Authorization")
		request.Header.Del("Proxy-Connection")
		if e = request.Write(upstream); e != nil {
			finish("HTTP_WRITE_FAILED", 0, 0)
			return
		}
		n, e := io.Copy(countWriter{Writer: c, add: func(n int) { s.account(id, 0, uint64(n)) }}, upstream)
		result := "HTTP_RESPONSE"
		if e != nil {
			result = "HTTP_READ_FAILED"
		}
		finish(result, 0, uint64(n))
		return
	}
	if _, e = io.WriteString(c, "HTTP/1.1 200 Connection Established\r\n\r\n"); e != nil {
		finish("CLIENT_CLOSED", 0, 0)
		return
	}
	s.mu.Lock()
	r := s.records[id]
	r.Result = "TCP_ESTABLISHED_HTTP_RESULT_UNKNOWN"
	s.records[id] = r
	s.mu.Unlock()
	type copied struct {
		n   int64
		err error
	}
	out := make(chan copied, 1)
	go func() {
		n, e := io.Copy(countWriter{Writer: upstream, add: func(n int) { s.account(id, uint64(n), 0) }}, parsed)
		if v, ok := upstream.(interface{ CloseWrite() error }); ok {
			v.CloseWrite()
		}
		out <- copied{n, e}
	}()
	rx, re := io.Copy(countWriter{Writer: c, add: func(n int) { s.account(id, 0, uint64(n)) }}, upstream)
	c.Close()
	upstream.Close()
	tx := <-out
	result := "TUNNEL_CLOSED"
	if re != nil || tx.err != nil {
		result = "TUNNEL_ENDED"
	}
	finish(result, uint64(tx.n), uint64(rx))
}
func writeError(c net.Conn, code int) {
	c.SetWriteDeadline(time.Now().Add(time.Second))
	fmt.Fprintf(c, "HTTP/1.1 %d %s\r\nConnection: close\r\nContent-Length: 0\r\n\r\n", code, http.StatusText(code))
}

// Account only bytes actually written to the corresponding socket. Active
// tunnels expose live counters; a completed copy replaces them with its totals.
func (s *Server) account(id, tx, rx uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := s.records[id]
	r.Sent += tx
	r.Received += rx
	s.records[id] = r
}

type countWriter struct {
	io.Writer
	add func(int)
}

func (w countWriter) Write(p []byte) (int, error) {
	n, e := w.Writer.Write(p)
	if n > 0 {
		w.add(n)
	}
	return n, e
}
