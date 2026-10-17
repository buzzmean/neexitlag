//go:build windows

package windowsplatform_test

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"github.com/buzzmean/neexitlag/core/browserproxy"
	"github.com/buzzmean/neexitlag/core/flowpolicy"
	"github.com/buzzmean/neexitlag/core/localstrategy"
	"github.com/buzzmean/neexitlag/core/profiles"
	"github.com/buzzmean/neexitlag/core/security"
	"github.com/buzzmean/neexitlag/core/transport"
	"github.com/buzzmean/neexitlag/internal/browserfixture"
	wp "github.com/buzzmean/neexitlag/platform/windows"
	cs "github.com/buzzmean/neexitlag/services/client-service"
	"github.com/buzzmean/neexitlag/services/relay"
	"github.com/buzzmean/neexitlag/tools/labcert"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// This test uses real Chrome, real HTTP/2+TLS, real authenticated relay and real
// signed WinDivert packet reinjection on an isolated loopback target. It is not
// a measurement of public YouTube/OpenAI or the production VPS.
func TestChromeSitesNativeAcceptance(t *testing.T) {
	dir := os.Getenv("NEEXITLAG_TEST_DRIVER_DIR")
	if dir == "" {
		t.Skip("signed driver required; enabled explicitly on native Windows CI")
	}
	chrome := browserfixture.ChromePath()
	if chrome == "" {
		t.Fatal("native Chrome required, cannot report PASS without a browser")
	}
	raw, err := os.ReadFile(filepath.Join(dir, "pins.json"))
	if err != nil {
		t.Fatal(err)
	}
	var files wp.DriverFiles
	if err = json.Unmarshal(raw, &files); err != nil {
		t.Fatal(err)
	}
	files.Directory = dir
	f, err := browserfixture.New()
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 75*time.Second)
	defer cancel()
	c := cs.New(ctx)
	self, err := wp.NewDiscovery().Lookup(uint32(os.Getpid()))
	if err != nil {
		t.Fatal(err)
	}
	c.BrowserOwnerSID, err = wp.ProcessUserSID(self)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	backend := wp.NewBrowserLoopbackFixture(files, c.Engine)
	c.Backend = backend
	c.BrowserResolver = f.Resolver
	c.BrowserProbeRoots = f.Roots
	c.Start()
	deadline := time.Now().Add(12 * time.Second)
	for {
		if _, err = c.Handle(ctx, "local.network", nil); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("network identity not ready", err)
		}
		time.Sleep(100 * time.Millisecond)
	}
	st, ct, err := labcert.Pair()
	if err != nil {
		t.Fatal(err)
	}
	pub, key, _ := ed25519.GenerateKey(rand.Reader)
	server, err := relay.Listen("127.0.0.1:0", relay.Config{TLS: st, TicketVerifier: security.NewTicketVerifier(map[string]ed25519.PublicKey{"test": pub}, "VPS-01"), Allow: []netip.Prefix{netip.MustParsePrefix("127.0.0.0/8")}, Acceptance: &relay.AcceptanceConfig{LeaseDuration: 90 * time.Second, OutboundSource: netip.MustParseAddr("127.0.0.2")}})
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	peer := ct.Certificates[0].PrivateKey.(ed25519.PrivateKey).Public().(ed25519.PublicKey)
	pool, err := transport.NewPool(ctx, transport.DefaultPoolConfig(), func(ctx context.Context) (*transport.Client, error) {
		id, _ := security.NewID()
		now := time.Now()
		ticket, e := security.Sign("ticket", "test", key, security.Ticket{ID: id, DeviceID: "fixture", DeviceKey: security.Fingerprint(peer), RelayID: "VPS-01", IssuedAt: now, ExpiresAt: now.Add(time.Minute)})
		if e != nil {
			return nil, e
		}
		return transport.DialWithTicket(ctx, server.Addr().String(), ct, ticket)
	})
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	c.Engine.SetRelay("VPS-01", pool)
	entries, _, err := localstrategy.PresetProfiles("youtube", chrome, false)
	if err != nil {
		t.Fatal(err)
	}
	selected := ""
	for i := range entries {
		entries[i].Ports = []uint16{f.Port()}
		if strings.HasSuffix(entries[i].ID, "-disorder") && !strings.HasSuffix(entries[i].ID, "-fake8-disorder") {
			selected = entries[i].ID
		}
	}
	if err = c.Local.Reload(localstrategy.Config{Version: 1, Enabled: true, User: entries}); err != nil {
		t.Fatal(err)
	}
	p := profiles.Profile{ID: "native-chrome-sites", Name: "Chrome isolated acceptance", Executables: []string{chrome}, Mode: profiles.DirectPlus, LocalStrategyID: selected, Custom: true}
	put := func() {
		raw, _ := json.Marshal(p)
		if _, e := c.Handle(ctx, "profile.set", raw); e != nil {
			t.Fatal(e)
		}
	}
	put()
	p.Mode = profiles.Auto
	put()
	if _, err = c.Handle(ctx, "optimize", []byte(`{"enabled":true}`)); err != nil {
		t.Fatal(err)
	}
	userData := t.TempDir()
	raw, _ = json.Marshal(map[string]string{"profile_id": p.ID, "directory": userData})
	result, err := c.Handle(ctx, "browser.start", raw)
	if err != nil {
		t.Fatal(err)
	}
	args := result.(map[string]any)["arguments"].([]string)
	out, err := f.RunChrome(ctx, chrome, args, func() error {
		for _, enabled := range []string{`{"enabled":false}`, `{"enabled":true}`} {
			if _, e := c.Handle(ctx, "optimize", []byte(enabled)); e != nil {
				return e
			}
		}
		return nil
	})
	if err != nil {
		view, _ := c.Handle(ctx, "browser.status", nil)
		t.Fatal("controlled Chrome HTTP(S) failed", err, out, view)
	}
	// Wait for actual Chrome sockets to close so forwarded byte counters are final.
	var snap browserproxy.Snapshot
	until := time.Now().Add(3 * time.Second)
	for {
		view, e := c.Handle(ctx, "browser.status", nil)
		if e != nil {
			t.Fatal(e)
		}
		snap = view.(browserproxy.Snapshot)
		active := false
		for _, r := range snap.Records {
			active = active || r.Active
		}
		if !active || time.Now().After(until) {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if snap.Lost != 0 || snap.Truncated {
		t.Fatal("incomplete browser proof", snap.Lost, snap.Truncated)
	}
	found := map[string]bool{}
	for _, r := range snap.Records {
		switch r.Domain {
		case "chatgpt.com", "blocked.example":
			if r.Plan.Policy.Action != flowpolicy.RelayRequired || r.Plan.Policy.Fallback != flowpolicy.Required || r.Sent == 0 || r.Received == 0 {
				t.Fatal("protected route or response missing", r)
			}
			found[r.Domain] = true
		case "www.youtube.com":
			if r.Plan.Policy.Action != flowpolicy.DirectPlus || !r.DirectPlusApplied || r.Sent == 0 || r.Received == 0 {
				t.Fatal("YouTube local route missing", r)
			}
			found[r.Domain] = true
		case "ordinary.example":
			if r.Plan.Policy.Action != flowpolicy.Direct || r.Sent == 0 || r.Received == 0 {
				t.Fatal("ordinary origin borrowed another route", r)
			}
			found[r.Domain] = true
		}
	}
	if len(found) != 4 {
		t.Fatal("missing CONNECT origins; possible HTTP/2 pooling", found)
	}
	evidence, err := backend.DiagnosticsScoped(wp.EvidenceQuery{Destination: f.Destination().String()})
	if err != nil {
		t.Fatal(err)
	}
	if evidence.EvidenceLost || evidence.Truncated {
		t.Fatal("incomplete native transform proof")
	}
	applied := false
	for _, flow := range evidence.Flows {
		if flow.Domain == "www.youtube.com" && flow.Counters.DirectPlus > 0 && flow.Counters.DirectPlusBytes > 0 && flow.Counters.RelayedBytes == 0 {
			applied = true
		}
		if protected := flow.Domain == "chatgpt.com"; protected && flow.Counters.DirectBytes > 0 {
			t.Fatal("protected native DIRECT leak", flow)
		}
	}
	if !applied {
		t.Fatal("Chrome HTTP success without actual native DIRECT+ transform is NOT a PASS", evidence)
	}
	if server.Metrics.TCPConnections.Load() < 2 || server.Metrics.RXBytes.Load() == 0 {
		t.Fatal("real authenticated relay not used")
	}
	t.Log("PASS: same Chrome simultaneous controlled HTTPS+HTTP2: protected RELAY_REQUIRED, YouTube native disorder with modified packets, ordinary DIRECT, blocked unknown -> real relay; no production VPS touched")
}
