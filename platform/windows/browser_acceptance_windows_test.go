//go:build windows

package windowsplatform_test

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
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
	"golang.org/x/sys/windows/registry"
	"io"
	"net"
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
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
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
	defer func() { c.Close() }()
	backend := wp.NewBrowserLoopbackFixture(files, c.Engine)
	c.Backend = browserMTUFixture{backend}
	c.BrowserResolver = f.Resolver
	c.BrowserProbeRoots = f.Roots
	statePath := filepath.Join(t.TempDir(), "profiles.json")
	if err = c.SetCustomPath(statePath); err != nil {
		t.Fatal(err)
	}
	c.Start()
	// IPC success only means a snapshot exists. A failed background inspection
	// has no fingerprint/time and must never be adopted as a ready network.
	// Allow one real 30-second producer cycle plus its 8-second query budget.
	waitNetwork := func() error {
		until := time.Now().Add(45 * time.Second)
		var last any
		for {
			value, e := c.Handle(ctx, "local.network", nil)
			last = value
			if e == nil {
				if n, ok := value.(*wp.NetworkContext); ok && n.Error == "" && n.Fingerprint != "" && !n.ObservedAt.IsZero() && time.Since(n.ObservedAt) <= 45*time.Second && len(n.VPN) == 0 && len(n.Conflicts) == 0 {
					return nil
				}
			}
			if time.Now().After(until) {
				return fmt.Errorf("network readiness: snapshot=%+v error=%v", last, e)
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(100 * time.Millisecond):
			}
		}
	}
	if err = waitNetwork(); err != nil {
		t.Fatal(err)
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
	if _, err = c.Handle(ctx, "optimize", []byte(`{"enabled":true}`)); err != nil {
		t.Fatal(err)
	}
	userData := t.TempDir()
	raw, _ = json.Marshal(map[string]string{"profile_id": p.ID, "directory": userData})
	result, err := c.Handle(ctx, "browser.prepare", raw)
	if err != nil {
		t.Fatal(err)
	}
	args := result.(map[string]any)["arguments"].([]string)
	healthURL := result.(map[string]any)["health_url"].(string)

	// Exercise the actual current-user Chrome platform policy, not just launch
	// flags. This key is confined to the disposable Windows CI user; no WinINET,
	// WinHTTP or machine proxy value is read/written. Refuse any foreign policy.
	policy, _, e := registry.CreateKey(registry.CURRENT_USER, `Software\Policies\Google\Chrome`, registry.QUERY_VALUE|registry.SET_VALUE)
	if e != nil {
		t.Fatal(e)
	}
	defer policy.Close()
	if _, _, e = policy.GetStringValue("ProxySettings"); !errors.Is(e, registry.ErrNotExist) {
		t.Fatal("refusing foreign Chrome proxy policy", e)
	}
	pac := ""
	clean := []string{}
	for _, a := range args {
		if strings.HasPrefix(a, "--proxy-pac-url=") {
			pac = strings.TrimPrefix(a, "--proxy-pac-url=")
			continue
		}
		if strings.HasPrefix(a, "--proxy-bypass-list=") {
			continue
		}
		clean = append(clean, a)
	}
	if pac == "" {
		t.Fatal("missing real PAC")
	}
	// No policy must fail before publishing AUTO; a successful Chrome process
	// exit or a bound listener is not sufficient activation evidence.
	if err := browserfixture.RunPolicyHealthProbe(ctx, chrome, healthURL, t.TempDir()); err == nil {
		t.Fatal("Chrome activated without actual policy")
	}
	proof, _ := json.Marshal(map[string]string{"profile_id": p.ID, "health_url": healthURL})
	if _, err := c.Handle(ctx, "browser.confirm", proof); err == nil {
		t.Fatal("failed policy probe falsely confirmed readiness")
	}
	policyJSON, _ := json.Marshal(map[string]any{"ProxyMode": "pac_script", "ProxyPacUrl": pac, "ProxyPacMandatory": true})
	if e = policy.SetStringValue("ProxySettings", string(policyJSON)); e != nil {
		t.Fatal(e)
	}
	defer policy.DeleteValue("ProxySettings")
	args = clean

	rules := []string{}
	for _, host := range browserfixture.Domains {
		rules = append(rules, "MAP "+host+" 127.0.0.1")
	}
	args = append(args, "--host-resolver-rules="+strings.Join(rules, ", "))
	f.ExpectedFailures = map[int][]string{3: {"chatgpt.com", "www.youtube.com", "blocked.example"}}
	recoveryArgs := result.(map[string]any)["recovery_arguments"].([]string)
	recoveryPAC := strings.TrimPrefix(recoveryArgs[1], "--proxy-pac-url=")
	recoveryJSON, _ := json.Marshal(map[string]any{"ProxyMode": "pac_script", "ProxyPacUrl": recoveryPAC, "ProxyPacMandatory": true})
	// Actual policy, normal sandbox, a disposable health profile. Commit AUTO
	// before starting/restoring the real seven-origin Chrome process.
	if err := browserfixture.RunPolicyHealthProbe(ctx, chrome, healthURL, t.TempDir()); err != nil {
		t.Fatal(err)
	}
	if err = waitNetwork(); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Handle(ctx, "browser.confirm", proof); err != nil {
		t.Fatal(err)
	}
	put()
	earlier := []browserproxy.Record{}
	sample := func() error {
		view, e := c.Handle(ctx, "browser.status", nil)
		if e != nil {
			return e
		}
		snapshot := view.(browserproxy.Snapshot)
		if snapshot.Truncated || snapshot.Lost != 0 {
			return errors.New("incomplete browser evidence before restart")
		}
		earlier = append(earlier, snapshot.Records...)
		return nil
	}
	out, err := f.RunChrome(ctx, chrome, args, func() error {
		if e := sample(); e != nil {
			return e
		}
		for _, enabled := range []string{`{"enabled":false}`, `{"enabled":true}`} {
			if _, e := c.Handle(ctx, "optimize", []byte(enabled)); e != nil {
				return e
			}
		}
		return nil
	}, func() error {
		if e := sample(); e != nil {
			return e
		}
		if err := policy.SetStringValue("ProxySettings", string(recoveryJSON)); err != nil {
			return err
		}
		c.Close()
		return nil
	}, func() error {
		c = cs.New(ctx)
		c.BrowserOwnerSID, err = wp.ProcessUserSID(self)
		if err != nil {
			return err
		}
		backend = wp.NewBrowserLoopbackFixture(files, c.Engine)
		c.Backend = browserMTUFixture{backend}
		c.BrowserResolver = f.Resolver
		c.BrowserProbeRoots = f.Roots
		if err = c.SetCustomPath(statePath); err != nil {
			return err
		}
		if err = c.Local.Reload(localstrategy.Config{Version: 1, Enabled: true, User: entries}); err != nil {
			return err
		}
		c.Engine.SetRelay("VPS-01", pool)
		c.Start()
		if err = waitNetwork(); err != nil {
			return err
		}
		if err = c.RestoreBrowserTransport(); err != nil {
			return err
		}
		// A recovery PAC must NOT silently arm native Chrome AUTO on startup.
		if _, e := c.Handle(ctx, "optimize", []byte(`{"enabled":true}`)); e == nil {
			return errors.New("capture armed while actual Chrome policy was still recovery")
		}
		if err = policy.SetStringValue("ProxySettings", string(policyJSON)); err != nil {
			return err
		}
		fresh, e := c.Handle(ctx, "browser.prepare", raw)
		if e != nil {
			return e
		}
		f.HealthURL = fresh.(map[string]any)["health_url"].(string)
		_, err = c.Handle(ctx, "optimize", []byte(`{"enabled":true}`))
		return err
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
	if f.ECHClientHellos.Load() == 0 {
		t.Fatal("real sandboxed Chrome ECH/GREASE coverage missing")
	}
	if f.ProtectedDirectAttempts.Load() != 0 {
		t.Fatal("protected TLS attempted DIRECT during activation/outage/restart", f.ProtectedDirectAttempts.Load())
	}
	if snap.Lost != 0 || snap.Truncated {
		t.Fatal("incomplete browser proof", snap.Lost, snap.Truncated)
	}
	found := map[string]bool{}
	for _, r := range snap.Records {
		switch r.Domain {
		case "blocked.example":
			// Unknown AUTO must be allowed to try DIRECT first. Speculative
			// unconfirmed local attempts are NOT an OpenAI leak or HTTP PASS:
			// the target rejects their TLS before any HTTP, and every stage's
			// browser nonce already requires a relay-source HTTP response.
			if r.Plan.Policy.Action != flowpolicy.RelayRequired {
				if r.Plan.HTTPStatus >= 200 && r.Plan.HTTPStatus < 400 {
					t.Fatal("blocked origin falsely confirmed local", r)
				}
				continue
			}
			if r.Plan.Policy.Fallback != flowpolicy.Required || r.Sent == 0 || r.Received == 0 {
				t.Fatal("learned relay response missing", r)
			}
			found[r.Domain] = true
		case "chatgpt.com":
			if r.Plan.Policy.Action != flowpolicy.RelayRequired || r.Plan.Policy.Fallback != flowpolicy.Required || r.Sent == 0 || r.Received == 0 {
				t.Fatal("protected route or response missing", r)
			}
			found[r.Domain] = true
		case "www.youtube.com":
			if r.Plan.Policy.Action != flowpolicy.DirectPlus || !r.DirectPlusApplied || r.Sent == 0 || r.Received == 0 {
				proof, _ := backend.DiagnosticsScoped(wp.EvidenceQuery{Destination: f.Destination().String()})
				rawProof, _ := json.Marshal(proof)
				t.Fatal("YouTube local route missing", r, string(rawProof))
			}
			found[r.Domain] = true
		case "ordinary.example", "vk.com", "music.yandex.ru", "langame.ru":
			if r.Plan.Policy.Action != flowpolicy.Direct || r.Sent == 0 || r.Received == 0 {
				t.Fatal("ordinary origin borrowed another route", r)
			}
			found[r.Domain] = true
		}
	}
	if len(found) != len(browserfixture.Domains) {
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

// Loopback's jumbo MSS can emit a complete modern Chrome ClientHello in one
// >1500-byte packet, outside the production engine's explicit packet bound.
// This fixture models ordinary Internet MTU using actual bounded socket writes;
// the unchanged WinDivert engine still captures/reassembles/transforms every
// packet, and acceptance requires its real modified-packet counters. No TLS
// feature or cipher is disabled, and no production MTU/route is modified.
type browserMTUFixture struct{ *wp.WinDivertBackend }

func (b browserMTUFixture) OpenBrowserLocalTCP(ctx context.Context, id wp.ProcessIdentity, src, dst netip.AddrPort, d wp.Decision) (net.Conn, error) {
	c, e := b.WinDivertBackend.OpenBrowserLocalTCP(ctx, id, src, dst, d)
	if e != nil {
		return nil, e
	}
	return browserMTUConn{c}, nil
}

type browserMTUConn struct{ net.Conn }

func (c browserMTUConn) Write(p []byte) (int, error) {
	total := 0
	for len(p) > 0 {
		size := len(p)
		if size > 1000 {
			size = 1000
		}
		n, e := c.Conn.Write(p[:size])
		total += n
		if e != nil {
			return total, e
		}
		if n != size {
			return total, io.ErrShortWrite
		}
		p = p[size:]
	}
	return total, nil
}
func (c browserMTUConn) CloseWrite() error {
	if v, ok := c.Conn.(interface{ CloseWrite() error }); ok {
		return v.CloseWrite()
	}
	return nil
}
