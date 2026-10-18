package clientservice

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"strings"
	"time"

	"github.com/buzzmean/neexitlag/core/browserproxy"
	"github.com/buzzmean/neexitlag/core/flowpolicy"
)

// browserStatusView is browser.status plus the DIRECT-timeout cache counters, so the
// load this cache puts on the relay is visible in diagnostics.
type browserStatusView struct {
	browserproxy.Snapshot
	DirectTimeouts flowpolicy.DirectTimeoutStats `json:"direct_timeout_cache"`
}

// directTimeoutEligible limits the cache to ordinary DIRECT-first origins. Protected
// origins are relay-required from the start and IP literals have no domain identity.
// YouTube is eligible only while its policy is plain DIRECT; once DIRECT+ is configured
// the plan's action is DirectPlus and this cache stays out of its way.
func directTimeoutEligible(q browserproxy.Request, plan browserproxy.Plan) bool {
	if plan.Policy.Action != flowpolicy.Direct || protectedBrowserDomain(q.Host) {
		return false
	}
	_, ipErr := netip.ParseAddr(q.Host)
	return ipErr != nil
}

// cacheHost folds CDN families that use a different hostname per video or edge node
// (rr3---sn-xyz.googlevideo.com) into one entry; blocking hits the whole family.
func cacheHost(host string) string {
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	if flowpolicy.DomainMatches(host, "googlevideo.com") {
		return "googlevideo.com"
	}
	return host
}

func (c *Controller) directTimeoutKey(current profilesResult, q browserproxy.Request) flowpolicy.DirectTimeoutKey {
	return flowpolicy.DirectTimeoutKey{Network: current.network, Application: current.profile.ID, Host: cacheHost(q.Host), Port: q.Port}
}

// stallTracked: for HTTPS tunnels on 443 the proxy watches the origin's first byte, so a
// bare TCP connect must not count as DIRECT working (a blackholing path accepts it).
func stallTracked(q browserproxy.Request) bool { return q.Secure && q.Port == 443 }

// observeDirectOpen feeds the cache from every real DIRECT open and probe. Only a
// connect deadline counts as a timeout; cancellation by the browser and every other
// error say nothing about reachability.
func (c *Controller) observeDirectOpen(current profilesResult, q browserproxy.Request, plan browserproxy.Plan, err error, now time.Time) {
	if current.network == "" || !directTimeoutEligible(q, plan) {
		return
	}
	key := c.directTimeoutKey(current, q)
	switch {
	case err == nil:
		if !stallTracked(q) {
			c.browser.directTimeouts.RecordSuccess(key)
		}
	case errors.Is(err, context.DeadlineExceeded):
		c.browser.directTimeouts.RecordTimeout(key, now)
	}
}

// steerTimedOutOrigin moves a cached origin straight to the authenticated relay,
// fail-closed. It never returns a DIRECT decision.
func (c *Controller) steerTimedOutOrigin(plan browserproxy.Plan, current profilesResult, q browserproxy.Request, now time.Time) (browserproxy.Plan, bool) {
	if current.network == "" || !directTimeoutEligible(q, plan) {
		return plan, false
	}
	key := c.directTimeoutKey(current, q)
	active, due := c.browser.directTimeouts.Lookup(key, now)
	if !active {
		return plan, false
	}
	if due {
		c.revalidateDirect(current, q, plan, key)
	}
	relay := current.profile.RelayID
	if relay == "" {
		relay = "VPS-01"
	}
	plan.Policy = flowpolicy.Result{Action: flowpolicy.RelayRequired, RelayID: relay, Fallback: flowpolicy.Required, Source: "direct_timeout_cache", Reason: "DIRECT connect timed out twice on this network; relay-first until the next DIRECT recheck"}
	plan.Probe = "DIRECT_TIMEOUT_CACHED"
	return plan, true
}

// browserDirectAnswered: a DIRECT tunnel received its first byte from the origin, the
// application-level proof that DIRECT works for it.
func (c *Controller) browserDirectAnswered(q browserproxy.Request, plan browserproxy.Plan) {
	current, err := c.browserProfile(q.Owner)
	if err != nil {
		return
	}
	c.recordAnswered(current, q, plan)
}
func (c *Controller) recordAnswered(current profilesResult, q browserproxy.Request, plan browserproxy.Plan) {
	if current.network == "" || !directTimeoutEligible(q, plan) {
		return
	}
	c.browser.directTimeouts.RecordSuccess(c.directTimeoutKey(current, q))
}

// stallRetryPlan is the relay-first, fail-closed plan used to retry a stalled tunnel.
func stallRetryPlan(current profilesResult, plan browserproxy.Plan) browserproxy.Plan {
	relay := current.profile.RelayID
	if relay == "" {
		relay = "VPS-01"
	}
	plan.Policy = flowpolicy.Result{Action: flowpolicy.RelayRequired, RelayID: relay, Fallback: flowpolicy.Required, Source: "direct_stall_retry", Reason: "DIRECT origin silent after the ClientHello; retried through the authenticated relay"}
	plan.Probe = "DIRECT_STALL_RETRY"
	return plan
}

// browserStall is the proxy's OnStall hook. It counts the stall like a connect timeout
// (two within 60 s make the origin relay-first for the TTL) and opens a relay path for
// the current tunnel. It can only move traffic towards the relay: protected origins and
// anything that is not an ordinary DIRECT origin are refused, and the open itself enforces
// that a protected domain never downgrades.
func (c *Controller) browserStall(ctx context.Context, q browserproxy.Request, plan browserproxy.Plan) (net.Conn, browserproxy.Plan, error) {
	current, err := c.browserProfile(q.Owner)
	if err != nil {
		return nil, plan, err
	}
	return c.stallRetry(ctx, current, q, plan, c.browserOpen)
}

// stallRetry does the work of browserStall for an already resolved browser profile; open is
// the upstream opener (browserOpen in production).
func (c *Controller) stallRetry(ctx context.Context, current profilesResult, q browserproxy.Request, plan browserproxy.Plan, open func(context.Context, browserproxy.Request, browserproxy.Plan) (net.Conn, error)) (net.Conn, browserproxy.Plan, error) {
	if !directTimeoutEligible(q, plan) {
		return nil, plan, errors.New("stall retry not allowed for this origin")
	}
	if current.network != "" {
		c.browser.directTimeouts.RecordStall(c.directTimeoutKey(current, q), time.Now())
	}
	next := stallRetryPlan(current, plan)
	conn, err := open(ctx, q, next)
	c.browser.directTimeouts.NoteRetry(err == nil)
	return conn, next, err
}

// revalidateDirect runs the single background DIRECT probe of a TTL. The open itself
// reports success (origin returns to DIRECT) or a fresh timeout (entry renewed).
func (c *Controller) revalidateDirect(current profilesResult, q browserproxy.Request, plan browserproxy.Plan, key flowpolicy.DirectTimeoutKey) {
	probePlan := plan
	probePlan.Policy = flowpolicy.Result{Action: flowpolicy.Direct, Fallback: flowpolicy.FailOpen, Source: "direct_timeout_recheck", Reason: "periodic background DIRECT recheck of a timed-out origin"}
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		ctx, cancel := context.WithTimeout(c.ctx, 3*time.Second)
		defer cancel()
		conclusive := false
		defer func() { c.browser.directTimeouts.ProbeDone(key, time.Now(), conclusive) }()
		if _, err := c.browserProfile(q.Owner); err != nil {
			return // the browser is gone: nothing to conclude
		}
		// A real HTTPS exchange with the origin, not a bare connect: a blackholing path
		// accepts the connect and then swallows the ClientHello.
		open := func(ctx context.Context) (net.Conn, error) { return c.browserOpenUnobserved(ctx, q, probePlan) }
		detail := browserproxy.ProbeOriginDetailed(ctx, q.Host, q.Port, q.Secure, c.BrowserProbeRoots, open)
		switch detail.Failure {
		case "":
			c.browser.directTimeouts.RecordSuccess(key) // DIRECT works end to end again
			conclusive = true
		case "NETWORK_TIMEOUT", "CONNECTION_RESET", "TLS_FAILED", "CONNECTION_FAILED":
			c.browser.directTimeouts.RecordTimeout(key, time.Now()) // still broken: renew
			conclusive = true
		}
	}()
}
