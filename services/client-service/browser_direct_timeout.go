package clientservice

import (
	"context"
	"errors"
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
// and YouTube origins have their own policy, IP literals have no domain identity.
func directTimeoutEligible(q browserproxy.Request, plan browserproxy.Plan) bool {
	if plan.Policy.Action != flowpolicy.Direct || protectedBrowserDomain(q.Host) || youtubeDomain(q.Host) {
		return false
	}
	_, ipErr := netip.ParseAddr(q.Host)
	return ipErr != nil
}

func (c *Controller) directTimeoutKey(current profilesResult, q browserproxy.Request) flowpolicy.DirectTimeoutKey {
	return flowpolicy.DirectTimeoutKey{Network: current.network, Application: current.profile.ID, Host: strings.ToLower(q.Host), Port: q.Port}
}

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
		c.browser.directTimeouts.RecordSuccess(key)
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

// revalidateDirect runs the single background DIRECT probe of a TTL. The open itself
// reports success (origin returns to DIRECT) or a fresh timeout (entry renewed).
func (c *Controller) revalidateDirect(current profilesResult, q browserproxy.Request, plan browserproxy.Plan, key flowpolicy.DirectTimeoutKey) {
	probe := plan
	probe.Policy = flowpolicy.Result{Action: flowpolicy.Direct, Fallback: flowpolicy.FailOpen, Source: "direct_timeout_recheck", Reason: "periodic background DIRECT recheck of a timed-out origin"}
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		ctx, cancel := context.WithTimeout(c.ctx, 1500*time.Millisecond)
		defer cancel()
		conn, err := c.browserOpen(ctx, q, probe)
		if conn != nil {
			conn.Close()
		}
		conclusive := err == nil || errors.Is(err, context.DeadlineExceeded)
		c.browser.directTimeouts.ProbeDone(key, time.Now(), conclusive)
	}()
}
