package clientservice

import (
	"context"
	"errors"
	"net"
	"strings"

	"github.com/buzzmean/neexitlag/core/browserproxy"
)

// Retain verified DNS recovery when replacing ordinary AUTO probes with a known
// relay-first rule. Only a destination OPEN rejection can suggest wrong DNS;
// authentication/session failures must not cause a DNS/probe storm.
func recoverableKnownOriginOpen(err error) bool {
	var typed interface{ FailureReason() string }
	if !errors.As(err, &typed) {
		return false
	}
	switch strings.ToLower(typed.FailureReason()) {
	case "upstream_dial_timeout", "upstream_connection_refused", "upstream_dial_failed":
		return true
	}
	return false
}
func (c *Controller) openKnownBrowserGroup(ctx context.Context, current profilesResult, q browserproxy.Request, p browserproxy.Plan, open func(context.Context, browserproxy.Request, browserproxy.Plan) (net.Conn, error)) (net.Conn, browserproxy.Plan, error) {
	conn, err := open(ctx, q, p)
	if err == nil || ctx.Err() != nil || !q.Secure || !recoverableKnownOriginOpen(err) {
		return conn, p, err
	}
	select {
	case c.browser.probes <- struct{}{}:
		defer func() { <-c.browser.probes }()
	default:
		return nil, p, err
	}
	recovered, ok := c.recoverBrowserDNS(ctx, q, current, p)
	if !ok {
		return nil, recovered, err
	}
	conn, err = open(ctx, q, recovered)
	return conn, recovered, err
}
