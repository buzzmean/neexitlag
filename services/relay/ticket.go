package relay

import (
	"context"
	"crypto/ed25519"
	"errors"
	"github.com/buzzmean/neexitlag/core/protocol"
	"time"
)

func (x *session) authenticate() error {
	ctx, cancel := context.WithTimeout(x.conn.Context(), 3*time.Second)
	defer cancel()
	s, e := x.conn.AcceptStream(ctx)
	if e != nil {
		return e
	}
	defer s.Close()
	defer s.CancelRead(0)
	s.SetDeadline(time.Now().Add(3 * time.Second))
	raw, e := protocol.ReadTicket(s)
	if e != nil {
		return e
	}
	certs := x.conn.ConnectionState().TLS.PeerCertificates
	if len(certs) < 1 {
		return errors.New("device certificate missing")
	}
	peer, ok := certs[0].PublicKey.(ed25519.PublicKey)
	if !ok {
		return errors.New("Ed25519 device certificate required")
	}
	ticket, e := x.server.config.TicketVerifier.Accept(raw, peer, time.Now())
	if e != nil {
		s.Write([]byte{1})
		return e
	}
	x.deviceID = ticket.DeviceID
	x.leaseExpires = time.Now().Add(30 * time.Minute)
	_, e = s.Write([]byte{0})
	return e
}
