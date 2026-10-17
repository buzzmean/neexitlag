package clientservice

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"github.com/buzzmean/neexitlag/core/browserproxy"
	"github.com/buzzmean/neexitlag/core/flowpolicy"
	"github.com/buzzmean/neexitlag/core/profiles"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
	"time"
)

func TestOriginComparisonStrictlyRelayOnlyAndReadOnly(t *testing.T) {
	c := New(context.Background())
	defer c.Close()
	c.Profiles.Upsert(profiles.Profile{ID: "keep", Name: "existing", Executables: []string{`C:\Apps\chrome.exe`}, Mode: profiles.Auto})
	before, _ := json.Marshal(c.Profiles.List())
	for _, raw := range []string{`{"host":"a.local","candidates":["8.8.8.8"]}`, `{"host":"example.com","candidates":["127.0.0.1"]}`, `{"host":"example.com","candidates":["169.254.169.254"]}`, `{"host":"example.com","candidates":["8.8.8.8","8.8.8.8"]}`, `{"host":"example.com","candidates":["8.8.8.8","::ffff:8.8.8.8"]}`, `{"host":"example.com","candidates":[]}`, `{"host":"example.com","candidates":["8.8.8.8"],"insecure":true}`} {
		if _, err := c.Handle(context.Background(), "browser.origin.check", []byte(raw)); err == nil {
			t.Fatal("invalid diagnostic accepted", raw)
		}
	}
	// No relay configured: a public destination must fail BEFORE any DIRECT dial.
	response, err := c.Handle(context.Background(), "browser.origin.check", []byte(`{"host":"example.com","candidates":["8.8.8.8"]}`))
	if err != nil {
		t.Fatal(err)
	}
	wire, _ := json.Marshal(response)
	var report struct {
		Candidates []struct {
			Route    string                          `json:"route"`
			Verified bool                            `json:"origin_verified"`
			Origin   struct{ Failure, Phase string } `json:"origin"`
		} `json:"candidates"`
	}
	if json.Unmarshal(wire, &report) != nil || len(report.Candidates) != 1 {
		t.Fatal(string(wire))
	}
	result := report.Candidates[0]
	if result.Route != string(flowpolicy.RelayRequired) || result.Verified || result.Origin.Phase != "RELAY_OR_TCP_OPEN" || result.Origin.Failure == "" {
		t.Fatal(string(wire))
	}
	after, _ := json.Marshal(c.Profiles.List())
	if string(before) != string(after) || c.Snapshot().Optimizing {
		t.Fatal("read-only check changed profiles/optimization")
	}
}

func TestDNSRecoveredPinAvailableToQueuedChromeRequests(t *testing.T) {
	c := New(context.Background())
	defer c.Close()
	target := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("independent origin proof")) }))
	defer target.Close()
	roots := x509.NewCertPool()
	roots.AddCert(target.Certificate())
	destination := target.Listener.Addr().(*net.TCPAddr).AddrPort()
	current := profilesResult{profile: profiles.Profile{ID: "chrome", Name: "Chrome", Executables: []string{`C:\Apps\chrome.exe`}, Mode: profiles.Auto}, network: "network-A", enabled: true}
	q := browserproxy.Request{Host: "example.com", Port: destination.Port(), Secure: true}
	// This test supplies a local TLS origin to the cache; the separate core
	// integration proves actual authenticated relay DNS + origin I/O.
	result := c.browser.dns.Recover(context.Background(), c.browserDNSKey(q, current), q.Host, q.Port, roots, func(context.Context, string, browserproxy.DNSOpen) (browserproxy.DNSAnswer, error) {
		return browserproxy.DNSAnswer{Addresses: []netip.Addr{destination.Addr()}, TTL: time.Minute, Source: "ISOLATED_TLS", ObtainedAt: time.Now()}, nil
	}, func(ctx context.Context, _ netip.AddrPort) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "tcp", destination.String())
	})
	if !result.Verified {
		t.Fatal(result)
	}
	// Exactly as after the primary planner completed recovery: ordinary learning
	// is empty. Both initial requests and pending-channel waiters use this reader.
	for i := 0; i < 2; i++ {
		plan, ok := c.browserDNSCachedPlan(q, current)
		if !ok || plan.Destination != destination || plan.Policy.Action != flowpolicy.RelayRequired || plan.Policy.Fallback != flowpolicy.Required {
			t.Fatal("queued request reverted to bad system DNS", plan, ok)
		}
	}
	changed := current
	changed.network = "network-B"
	if _, ok := c.browserDNSCachedPlan(q, changed); ok {
		t.Fatal("cross-network pin")
	}
	changed = current
	changed.profile.RelayID = "VPS-02"
	if _, ok := c.browserDNSCachedPlan(q, changed); ok {
		t.Fatal("cross-relay pin")
	}
	changed = current
	changed.enabled = false
	if _, ok := c.browserDNSCachedPlan(q, changed); ok {
		t.Fatal("disabled optimization used recovered pin")
	}
	changed = current
	changed.profile.Rules = []flowpolicy.Rule{{ID: "block", Domains: []string{q.Host}, Action: flowpolicy.Block}}
	if _, ok := c.browserDNSCachedPlan(q, changed); ok {
		t.Fatal("new BLOCK policy ignored")
	}
}
