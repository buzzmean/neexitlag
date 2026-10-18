package clientservice

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/buzzmean/neexitlag/core/browserproxy"
	"github.com/buzzmean/neexitlag/core/flowpolicy"
)

const browserDirectBudget = 3 * time.Second
const browserDirectAttempt = 1300 * time.Millisecond

// Only the planner's DNS answers are tried, at most twice. The deadline survives
// a successful TCP connect so a silent TLS handshake cannot restart the budget.
func directAttemptContext(ctx context.Context, p browserproxy.Plan) (context.Context, context.CancelFunc) {
	deadline := time.Now().Add(browserDirectAttempt)
	if !p.DirectDeadline.IsZero() && p.DirectDeadline.Before(deadline) {
		deadline = p.DirectDeadline
	}
	return context.WithDeadline(ctx, deadline)
}
func nextDirectPlan(p browserproxy.Plan) (browserproxy.Plan, bool) {
	if p.DirectAttempts >= 2 || (!p.DirectDeadline.IsZero() && !time.Now().Before(p.DirectDeadline)) {
		return p, false
	}
	for _, addr := range p.DirectCandidates {
		if addr != p.Destination && addr.IsValid() {
			p.Destination = addr
			p.DirectAttempts = 2
			p.Probe = "DIRECT_SECOND_ADDRESS"
			return p, true
		}
	}
	return p, false
}
func retryableDirectOpen(err error) bool {
	if errors.Is(err, context.Canceled) {
		return false
	}
	var ne net.Error
	return errors.Is(err, context.DeadlineExceeded) || errors.As(err, &ne)
}
func (c *Controller) browserOpenPlan(ctx context.Context, q browserproxy.Request, p browserproxy.Plan) (net.Conn, browserproxy.Plan, error) {
	cur, err := c.browserProfile(q.Owner)
	if err != nil {
		return nil, p, err
	}
	return c.openBrowserPlan(ctx, cur, q, p, c.browserOpen)
}
func (c *Controller) openBrowserPlan(ctx context.Context, current profilesResult, q browserproxy.Request, p browserproxy.Plan, open func(context.Context, browserproxy.Request, browserproxy.Plan) (net.Conn, error)) (net.Conn, browserproxy.Plan, error) {
	if !p.Enabled || !directTimeoutEligible(q, p) {
		conn, err := open(ctx, q, p)
		if errors.Is(err, errVideoRelayCapacity) && ordinaryRelayFallback(q, p) {
			return c.openVideoOverflow(ctx, q, p, open)
		}
		return conn, p, err
	}
	if p.DirectDeadline.IsZero() {
		p.DirectDeadline = time.Now().Add(browserDirectBudget)
	}
	if p.DirectAttempts == 0 {
		p.DirectAttempts = 1
	}
	if c.directAddressFailed(current, q, p) {
		next, ok := nextDirectPlan(p)
		if !ok || c.directAddressFailed(current, q, next) {
			return c.relayAfterDirect(ctx, current, q, p, open)
		}
		p = next
	}

	attempt, stop := directAttemptContext(ctx, p)
	conn, err := open(attempt, q, p)
	stop()
	if err == nil || ctx.Err() != nil || !retryableDirectOpen(err) {
		return conn, p, err
	}
	if next, ok := nextDirectPlan(p); ok {
		p = next
		attempt, stop = directAttemptContext(ctx, p)
		conn, err = open(attempt, q, p)
		stop()
		if err == nil || ctx.Err() != nil || !retryableDirectOpen(err) {
			return conn, p, err
		}
	}
	return c.relayAfterDirect(ctx, current, q, p, open)
}
func (c *Controller) relayAfterDirect(ctx context.Context, current profilesResult, q browserproxy.Request, p browserproxy.Plan, open func(context.Context, browserproxy.Request, browserproxy.Plan) (net.Conn, error)) (net.Conn, browserproxy.Plan, error) {
	p = stallRetryPlan(current, p)
	p.Policy.Source = "direct_open_retry"
	p.Probe = "DIRECT_OPEN_FAILED_RELAY"
	conn, err := open(ctx, q, p)
	if errors.Is(err, errVideoRelayCapacity) {
		return c.openVideoOverflow(ctx, q, p, open)
	}
	return conn, p, err
}
func (c *Controller) directAddressFailed(cur profilesResult, q browserproxy.Request, p browserproxy.Plan) bool {
	return cur.network != "" && c.browser.directTimeouts.AddressFailed(c.directTimeoutKey(cur, q, p), time.Now())
}

func ordinaryRelayFallback(q browserproxy.Request, p browserproxy.Plan) bool {
	if protectedBrowserDomain(q.Host) || !flowpolicy.DomainMatches(q.Host, "googlevideo.com") {
		return false
	}
	switch p.Policy.Source {
	case "direct_timeout_cache", "direct_open_retry", "direct_stall_retry", "youtube_relay_first":
		return true
	}
	return false
}
func (c *Controller) openVideoOverflow(ctx context.Context, q browserproxy.Request, p browserproxy.Plan, open func(context.Context, browserproxy.Request, browserproxy.Plan) (net.Conn, error)) (net.Conn, browserproxy.Plan, error) {
	p.Policy = flowpolicy.Result{Action: flowpolicy.Direct, Fallback: flowpolicy.FailOpen, Source: "googlevideo_relay_limit", Reason: "Four googlevideo relay tunnels already active on this PC; this ordinary origin remains DIRECT"}
	p.Probe = "VIDEO_RELAY_CAPACITY_DIRECT"
	p.DirectCandidates = nil
	p.DirectAttempts = 2
	p.DirectDeadline = time.Now().Add(browserDirectAttempt)
	attempt, stop := directAttemptContext(ctx, p)
	defer stop()
	conn, err := open(attempt, q, p)
	return conn, p, err
}
