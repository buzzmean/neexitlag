package clientservice

import (
	"context"
	"errors"
	"github.com/buzzmean/neexitlag/core/transport"
	"io"
	"net"
	"net/netip"
	"testing"

	"github.com/buzzmean/neexitlag/core/browserproxy"
	"github.com/buzzmean/neexitlag/core/flowpolicy"
	wp "github.com/buzzmean/neexitlag/platform/windows"
)

func TestBrowserKnownGroupsRelayFirstWithoutColdDirectOrFallback(t *testing.T) {
	c, cur := timeoutFixture(t)
	for _, host := range []string{"web.telegram.org", "kws2.web.telegram.org", "venus.web.telegram.org", "pluto2.web.telegram.org", "zws1.web.telegram.org", "t.me", "telegram.org", "claude.ai", "api.anthropic.com", "assets-proxy.anthropic.com", "bridge.claudeusercontent.com", "claude.com"} {
		q := browserproxy.Request{Host: host, Port: 443, Secure: true}
		p := boundedDirectPlan() // Also rejects a stale DIRECT plan made before this rule.
		calls := 0
		_, got, err := c.openBrowserPlan(context.Background(), cur, q, p, func(_ context.Context, _ browserproxy.Request, p browserproxy.Plan) (net.Conn, error) {
			calls++
			if p.Policy.Action != flowpolicy.RelayRequired || p.Policy.Fallback != flowpolicy.Required || p.Policy.RelayID != "VPS-01" || p.DirectAttempts != 0 {
				t.Fatal(host, p)
			}
			return nil, errors.New("isolated relay unavailable")
		})
		if calls != 1 || err == nil || got.Probe != "KNOWN_RELAY_FIRST" {
			t.Fatal(host, calls, got, err)
		}
		if directTimeoutEligible(q, p) {
			t.Fatal("known group used ordinary DIRECT timeout learning", host)
		}
	}
	if s := c.browser.directTimeouts.Stats(); s.Stalls != 0 || s.Created != 0 {
		t.Fatal("cold groups paid DIRECT timeout", s)
	}
	// Do not spread YouTube cache keys or local strategy eligibility to other groups.
	if flowpolicy.DirectTimeoutHost("web.telegram.org") == "youtube.com" || youtubeRelayDomain("claude.ai") || youtubeDomain("claude.ai") {
		t.Fatal("family policy cross-contamination")
	}
}
func TestBrowserKnownGroupsExactScopeDNSAndOrdinaryRoutes(t *testing.T) {
	c, cur := timeoutFixture(t)
	for _, host := range []string{"claude.ai.attacker.example", "notclaude.ai", "anthropic.com.evil.example", "web.telegram.org.evil.example", "telegram.example", "ordinary.example", "www.google.com"} {
		result := c.browserFlow(cur.profile, wp.ProcessIdentity{}, wp.Tuple{Protocol: "tcp", Destination: netip.MustParseAddrPort("8.8.8.8:443")}, host, cur.network)
		if result.Action != flowpolicy.Direct || flowpolicy.KnownBrowserRelayGroup(host) != "" {
			t.Fatal("ordinary origin captured by known rule", host, result)
		}
	}
	for _, host := range []string{"web.telegram.org", "claude.ai", "chatgpt.com"} {
		result := c.browserFlow(cur.profile, wp.ProcessIdentity{}, wp.Tuple{Protocol: "udp", Destination: netip.MustParseAddrPort("192.168.1.1:53")}, host, cur.network)
		if result.Action != flowpolicy.Direct {
			t.Fatal("DNS to relay", host, result)
		}
	}
	blocked := cur.profile
	blocked.Rules = []flowpolicy.Rule{{ID: "block", Domains: []string{"claude.ai"}, Action: flowpolicy.Block}}
	if got := c.browserFlow(blocked, wp.ProcessIdentity{}, wp.Tuple{Protocol: "tcp", Destination: netip.MustParseAddrPort("8.8.8.8:443")}, "claude.ai", cur.network); got.Action != flowpolicy.Block {
		t.Fatal("explicit BLOCK weakened", got)
	}

	// Policy on explicitly disabled optimization stays outside automatic forcing.
	p := boundedDirectPlan()
	p.Enabled = false
	calls := 0
	_, got, _ := c.openBrowserPlan(context.Background(), cur, browserproxy.Request{Host: "claude.ai", Port: 443}, p, func(_ context.Context, _ browserproxy.Request, p browserproxy.Plan) (net.Conn, error) {
		calls++
		return nil, errors.New("test")
	})
	if calls != 1 || got.Policy.Action != flowpolicy.Direct {
		t.Fatal("disabled optimization changed", got)
	}
}

func TestBrowserKnownGroupsSuccessfulRelayAndDNSFailureScope(t *testing.T) {
	c, cur := timeoutFixture(t)
	for _, host := range []string{"venus.web.telegram.org", "api.anthropic.com"} {
		a, b := net.Pipe()
		done := make(chan struct{})
		go func() { defer close(done); defer b.Close(); _, _ = b.Write([]byte("origin reply")) }()
		conn, plan, err := c.openBrowserPlan(context.Background(), cur, browserproxy.Request{Host: host, Port: 443, Secure: true}, boundedDirectPlan(), func(_ context.Context, _ browserproxy.Request, p browserproxy.Plan) (net.Conn, error) {
			if p.Policy.Action != flowpolicy.RelayRequired || p.Policy.Fallback != flowpolicy.Required || p.DirectAttempts != 0 {
				t.Fatal(p)
			}
			return a, nil
		})
		if err != nil || plan.Probe != "KNOWN_RELAY_FIRST" {
			t.Fatal(plan, err)
		}
		data, err := io.ReadAll(conn)
		conn.Close()
		<-done
		if err != nil || string(data) != "origin reply" {
			t.Fatal(string(data), err)
		}
	}
	for _, code := range []string{"upstream_dial_timeout", "upstream_connection_refused", "upstream_dial_failed"} {
		if !recoverableKnownOriginOpen(&transport.OpenError{Reason: code}) {
			t.Fatal(code)
		}
	}
	for _, code := range []string{"ticket_rejected", "session_closed", "stream_capacity", "transport_open_failed"} {
		if recoverableKnownOriginOpen(&transport.OpenError{Reason: code}) {
			t.Fatal("auth/session failure triggered DNS recovery", code)
		}
	}
}
