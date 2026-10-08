package transport

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/buzzmean/neexitlag/core/protocol"
	"github.com/quic-go/quic-go"
	"io"
	"time"
)

func DialWithTicket(ctx context.Context, addr string, t *tls.Config, ticket []byte) (*Client, error) {
	cfg := t.Clone()
	cfg.NextProtos = []string{protocol.TicketALPN}
	conn, e := quic.DialAddr(ctx, addr, cfg, QUICConfig())
	if e != nil {
		return nil, e
	}
	fail := func(e error) (*Client, error) { conn.CloseWithError(1, "authentication failed"); return nil, e }
	if !conn.ConnectionState().SupportsDatagrams {
		return fail(errors.New("relay lacks datagrams"))
	}
	s, e := conn.OpenStreamSync(ctx)
	if e != nil {
		return fail(e)
	}
	defer s.Close()
	deadline := time.Now().Add(3 * time.Second)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	s.SetDeadline(deadline)
	stop := context.AfterFunc(ctx, func() { s.CancelRead(1); s.CancelWrite(1) })
	defer stop()
	if e = protocol.WriteTicket(s, ticket); e != nil {
		return fail(e)
	}
	var ack [1]byte
	if _, e = io.ReadFull(s, ack[:]); e != nil || ack[0] != 0 {
		if e == nil {
			e = errors.New("relay rejected ticket")
		}
		return fail(e)
	}
	c := &Client{conn: conn, flows: make(map[uint32]*UDPFlow), done: make(chan struct{})}
	go c.receive()
	return c, nil
}
