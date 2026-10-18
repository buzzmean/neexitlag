package clientservice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/buzzmean/neexitlag/core/browserproxy"
	"github.com/buzzmean/neexitlag/core/flowpolicy"
	"github.com/buzzmean/neexitlag/core/profiles"
)

func timeoutFixture(t *testing.T) (*Controller, profilesResult) {
	t.Helper()
	c := New(context.Background())
	t.Cleanup(c.Close)
	return c, profilesResult{profile: profiles.Profile{ID: "chrome-x"}, network: "net-1", enabled: true}
}

var directPlan = browserproxy.Plan{Policy: flowpolicy.Result{Action: flowpolicy.Direct, Fallback: flowpolicy.FailOpen}}

func TestTwoConnectTimeoutsSteerOriginToRelayFailClosed(t *testing.T) {
	c, cur := timeoutFixture(t)
	q := browserproxy.Request{Host: "Claude.AI", Port: 443, Secure: true}
	now := time.Now()
	if _, steered := c.steerTimedOutOrigin(directPlan, cur, q, now); steered {
		t.Fatal("steered before any timeout")
	}
	c.observeDirectOpen(cur, q, directPlan, context.DeadlineExceeded, now)
	if _, steered := c.steerTimedOutOrigin(directPlan, cur, q, now); steered {
		t.Fatal("a single timeout must not steer")
	}
	c.observeDirectOpen(cur, q, directPlan, fmt.Errorf("dial: %w", context.DeadlineExceeded), now.Add(30*time.Second))
	plan, steered := c.steerTimedOutOrigin(directPlan, cur, q, now.Add(31*time.Second))
	if !steered {
		t.Fatal("two timeouts within 60 s did not steer")
	}
	p := plan.Policy
	if p.Action != flowpolicy.RelayRequired || p.Fallback != flowpolicy.Required || p.RelayID != "VPS-01" || p.Source != "direct_timeout_cache" || plan.Probe != "DIRECT_TIMEOUT_CACHED" {
		t.Fatalf("not relay-first fail-closed: %+v probe=%s", p, plan.Probe)
	}
	// A different network, application or host never inherits the cache.
	other := cur
	other.network = "net-2"
	if _, steered := c.steerTimedOutOrigin(directPlan, other, q, now.Add(32*time.Second)); steered {
		t.Fatal("cache leaked across networks")
	}
	if _, steered := c.steerTimedOutOrigin(directPlan, cur, browserproxy.Request{Host: "example.org", Port: 443}, now.Add(32*time.Second)); steered {
		t.Fatal("cache leaked across hosts")
	}
}

func TestOnlyConnectDeadlinesOnOrdinaryDirectOriginsCount(t *testing.T) {
	c, cur := timeoutFixture(t)
	now := time.Now()
	count := func(q browserproxy.Request, plan browserproxy.Plan, err error) {
		for i := 0; i < 3; i++ {
			c.observeDirectOpen(cur, q, plan, err, now.Add(time.Duration(i)*time.Second))
		}
	}
	ordinary := browserproxy.Request{Host: "a.example", Port: 443}
	count(ordinary, directPlan, context.Canceled)
	count(ordinary, directPlan, errors.New("connection refused"))
	count(browserproxy.Request{Host: "203.0.113.7", Port: 443}, directPlan, context.DeadlineExceeded) // IP literal
	count(browserproxy.Request{Host: "chatgpt.com", Port: 443}, directPlan, context.DeadlineExceeded) // protected
	count(browserproxy.Request{Host: "www.youtube.com", Port: 443}, directPlan, context.DeadlineExceeded)
	relayPlan := browserproxy.Plan{Policy: flowpolicy.Result{Action: flowpolicy.RelayRequired, RelayID: "VPS-01", Fallback: flowpolicy.Required}}
	count(ordinary, relayPlan, context.DeadlineExceeded) // relay timeouts say nothing about DIRECT
	if s := c.browser.directTimeouts.Stats(); s.ActiveOrigins != 0 || s.Created != 0 {
		t.Fatalf("ineligible events were cached: %+v", s)
	}
	// A success in between resets the consecutive count.
	c.observeDirectOpen(cur, ordinary, directPlan, context.DeadlineExceeded, now)
	c.observeDirectOpen(cur, ordinary, directPlan, nil, now.Add(time.Second))
	c.observeDirectOpen(cur, ordinary, directPlan, context.DeadlineExceeded, now.Add(2*time.Second))
	if _, steered := c.steerTimedOutOrigin(directPlan, cur, ordinary, now.Add(3*time.Second)); steered {
		t.Fatal("timeouts separated by a success counted as consecutive")
	}
}

func TestCacheHitsAreVisibleInBrowserStatus(t *testing.T) {
	c, cur := timeoutFixture(t)
	q := browserproxy.Request{Host: "claude.ai", Port: 443, Secure: true}
	now := time.Now()
	c.observeDirectOpen(cur, q, directPlan, context.DeadlineExceeded, now)
	c.observeDirectOpen(cur, q, directPlan, context.DeadlineExceeded, now.Add(time.Second))
	for i := 0; i < 4; i++ {
		c.steerTimedOutOrigin(directPlan, cur, q, now.Add(2*time.Second))
	}
	status, err := c.Handle(context.Background(), "browser.status", nil)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(status)
	var out struct {
		Cache struct {
			Active int    `json:"active_origins"`
			Hits   uint64 `json:"relay_steered_flows"`
			Top    []struct {
				Host string `json:"host"`
				Hits uint64 `json:"relay_steered_flows"`
			} `json:"top_origins"`
		} `json:"direct_timeout_cache"`
		Records json.RawMessage `json:"records"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	if out.Cache.Active != 1 || out.Cache.Hits != 4 || len(out.Cache.Top) != 1 || out.Cache.Top[0].Host != "claude.ai" || !strings.Contains(string(out.Records), "[") {
		t.Fatalf("%s", raw)
	}
}

// The recheck runs once per TTL in the background. When it cannot conclude (here the
// browser owner is gone) traffic stays on the relay and the recheck is retried soon.
func TestBackgroundRecheckKeepsRelayWhenInconclusiveAndRetries(t *testing.T) {
	c, cur := timeoutFixture(t)
	q := browserproxy.Request{Host: "claude.ai", Port: 443, Secure: true}
	now := time.Now()
	c.observeDirectOpen(cur, q, directPlan, context.DeadlineExceeded, now)
	c.observeDirectOpen(cur, q, directPlan, context.DeadlineExceeded, now.Add(time.Second))
	later := now.Add(flowpolicy.DirectTimeoutTTL + 2*time.Second)
	plan, steered := c.steerTimedOutOrigin(directPlan, cur, q, later)
	if !steered || plan.Policy.Action != flowpolicy.RelayRequired {
		t.Fatal("traffic left the relay while the recheck was due")
	}
	key := c.directTimeoutKey(cur, q)
	deadline := time.Now().Add(5 * time.Second)
	for {
		// After the inconclusive probe finished, the short retry backoff applies.
		if _, due := c.browser.directTimeouts.Lookup(key, time.Now().Add(2*time.Minute)); due {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("background recheck never finished or was never rescheduled")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if s := c.browser.directTimeouts.Stats(); s.ActiveOrigins != 1 || s.Revalidations < 2 {
		t.Fatalf("%+v", s)
	}
}
