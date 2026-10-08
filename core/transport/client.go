// Package transport multiplexes flow handles over a persistent secure QUIC connection.
package transport

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/buzzmean/neexitlag/core/protocol"
	"github.com/quic-go/quic-go"
	"io"
	"net"
	"net/netip"
	"sync"
	"sync/atomic"
	"time"
)

var ErrClosed = errors.New("flow or session closed")

type Client struct {
	conn  *quic.Conn
	next  atomic.Uint32
	mu    sync.Mutex
	flows map[uint32]*UDPFlow
	done  chan struct{}
	drops atomic.Uint64
}
type UDPFlow struct {
	client *Client
	id     uint32
	stream *quic.Stream
	queue  chan []byte
	done   chan struct{}
	once   sync.Once
}

func Dial(ctx context.Context, addr string, t *tls.Config) (*Client, error) {
	c, e := quic.DialAddr(ctx, addr, t, QUICConfig())
	if e != nil {
		return nil, e
	}
	if !c.ConnectionState().SupportsDatagrams {
		c.CloseWithError(1, "datagrams required")
		return nil, errors.New("relay lacks QUIC DATAGRAM")
	}
	x := &Client{conn: c, flows: make(map[uint32]*UDPFlow), done: make(chan struct{})}
	go x.receive()
	return x, nil
}
func (c *Client) Close() error          { return c.conn.CloseWithError(0, "shutdown") }
func (c *Client) Done() <-chan struct{} { return c.done }
func (c *Client) Dropped() uint64       { return c.drops.Load() }
func (c *Client) open(ctx context.Context, kind byte, dst netip.AddrPort) (*quic.Stream, uint32, error) {
	id := c.next.Add(1)
	if id == 0 {
		c.Close()
		return nil, 0, errors.New("flow IDs exhausted")
	}
	s, e := c.conn.OpenStreamSync(ctx)
	if e != nil {
		return nil, 0, e
	}
	deadline := time.Now().Add(5 * time.Second)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	s.SetDeadline(deadline)
	stop := context.AfterFunc(ctx, func() { s.CancelRead(1); s.CancelWrite(1) })
	defer stop()
	if e = (protocol.Open{Kind: kind, ID: id, Destination: dst}).Write(s); e == nil {
		var ack [1]byte
		_, e = io.ReadFull(s, ack[:])
		if e == nil && ack[0] != 0 {
			e = errors.New("relay rejected flow")
		}
	}
	if e != nil {
		s.CancelRead(1)
		s.CancelWrite(1)
		return nil, 0, e
	}
	s.SetDeadline(time.Time{})
	return s, id, nil
}
func (c *Client) OpenUDP(ctx context.Context, dst netip.AddrPort) (*UDPFlow, error) {
	s, id, e := c.open(ctx, protocol.UDP, dst)
	if e != nil {
		return nil, e
	}
	f := &UDPFlow{client: c, id: id, stream: s, queue: make(chan []byte, 32), done: make(chan struct{})}
	c.mu.Lock()
	select {
	case <-c.done:
		c.mu.Unlock()
		s.CancelRead(1)
		s.CancelWrite(1)
		return nil, ErrClosed
	default:
	}
	c.flows[id] = f
	c.mu.Unlock()
	go func() { var b [1]byte; _, _ = s.Read(b[:]); f.Close() }()
	return f, nil
}
func (c *Client) OpenTCP(ctx context.Context, dst netip.AddrPort) (net.Conn, error) {
	s, _, e := c.open(ctx, protocol.TCP, dst)
	if e != nil {
		return nil, e
	}
	return &streamConn{s, c.conn}, nil
}
func (c *Client) receive() {
	defer func() {
		close(c.done)
		c.mu.Lock()
		all := make([]*UDPFlow, 0, len(c.flows))
		for _, f := range c.flows {
			all = append(all, f)
		}
		c.mu.Unlock()
		for _, f := range all {
			f.Close()
		}
	}()
	for {
		b, e := c.conn.ReceiveDatagram(context.Background())
		if e != nil {
			return
		}
		id, p, e := protocol.Decode(b)
		if e != nil {
			c.drops.Add(1)
			continue
		}
		c.mu.Lock()
		f := c.flows[id]
		c.mu.Unlock()
		if f == nil {
			c.drops.Add(1)
			continue
		}
		select {
		case f.queue <- p:
		default:
			c.drops.Add(1)
		}
	}
}
func (f *UDPFlow) Send(p []byte) error {
	select {
	case <-f.done:
		return ErrClosed
	case <-f.client.done:
		return ErrClosed
	default:
	}
	b, e := protocol.Encode(f.id, p)
	if e != nil {
		return e
	}
	return f.client.conn.SendDatagram(b)
}
func (f *UDPFlow) Receive(ctx context.Context) ([]byte, error) {
	select {
	case <-f.done:
		return nil, ErrClosed
	case <-f.client.done:
		return nil, ErrClosed
	default:
	}
	select {
	case b := <-f.queue:
		return b, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-f.done:
		return nil, ErrClosed
	case <-f.client.done:
		return nil, ErrClosed
	}
}
func (f *UDPFlow) Close() error {
	f.once.Do(func() {
		close(f.done)
		f.client.mu.Lock()
		delete(f.client.flows, f.id)
		f.client.mu.Unlock()
		f.stream.CancelRead(0)
		f.stream.Close()
	})
	return nil
}

type streamConn struct {
	*quic.Stream
	c *quic.Conn
}

func (s *streamConn) LocalAddr() net.Addr  { return s.c.LocalAddr() }
func (s *streamConn) RemoteAddr() net.Addr { return s.c.RemoteAddr() }
func (s *streamConn) Close() error         { s.CancelRead(0); return s.Stream.Close() }
func (s *streamConn) CloseWrite() error    { return s.Stream.Close() }
