package clientservice

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/buzzmean/neexitlag/core/browserproxy"
	"github.com/buzzmean/neexitlag/core/flowpolicy"
	"github.com/buzzmean/neexitlag/core/profiles"
	wp "github.com/buzzmean/neexitlag/platform/windows"
)

func edgeTestProfile() profiles.Profile {
	return profiles.Profile{ID: "edge", Name: "Edge", Executables: []string{`C:\Edge\msedge.exe`}, Mode: profiles.Direct, Custom: true}
}
func edgeCall(c *Controller, method string, payload any) (any, error) {
	raw, _ := json.Marshal(payload)
	return c.Handle(context.Background(), method, raw)
}
func TestEdgePrepareIsolationAndNoSyntheticReadiness(t *testing.T) {
	c := New(context.Background())
	defer c.Close()
	c.BrowserOwnerSID = "owner"
	p := edgeTestProfile()
	c.Profiles.Upsert(p)
	chrome := profiles.Profile{ID: "chrome", Name: "Chrome", Mode: profiles.Direct, Executables: []string{`C:\Chrome\chrome.exe`}, Custom: true}
	c.Profiles.Upsert(chrome)
	old, err := edgeCall(c, "browser.prepare", map[string]string{"profile_id": chrome.ID, "directory": t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	chromePAC, chromeAddress, chromeRecovery := c.browser.expectedPAC, c.browser.address, c.browser.recoveryAddress
	result, err := edgeCall(c, "edge.prepare", map[string]string{"profile_id": p.ID, "directory": t.TempDir(), "recovery_address": "127.0.0.1:17777"})
	if err != nil {
		t.Fatal(err)
	}
	_ = old
	edge := result.(map[string]any)
	if edge["address"] == chromeAddress || edge["health_url"] == "" {
		t.Fatal("missing independent Edge endpoint/challenge")
	}
	if c.browser.expectedPAC != chromePAC || c.browser.address != chromeAddress || c.browser.recoveryAddress != chromeRecovery || c.enabled {
		t.Fatal("Edge changed Chrome/Optimize")
	}
	if _, err = edgeCall(c, "edge.confirm", map[string]string{"profile_id": p.ID, "health_url": edge["health_url"].(string)}); err == nil {
		t.Fatal("listener alone certified Edge")
	}
	p.Mode = profiles.Auto
	p.AppPreset = "edge-sites"
	if _, err = edgeCall(c, "profile.set", p); err == nil {
		t.Fatal("unproved Edge AUTO accepted")
	}
	if _, err = edgeCall(c, "browser.prepare", map[string]string{"profile_id": p.ID, "directory": t.TempDir()}); err == nil {
		t.Fatal("Edge admitted into Chrome lane")
	}
}
func TestEdgeBindFailureDoesNotBreakChrome(t *testing.T) {
	c := New(context.Background())
	defer c.Close()
	c.BrowserOwnerSID = "owner"
	p := edgeTestProfile()
	c.Profiles.Upsert(p)
	occupied, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer occupied.Close()
	c.browser.edge = &edgeRuntime{saved: edgeSaved{Address: occupied.Addr().String()}}
	if _, err = edgeCall(c, "edge.prepare", map[string]string{"profile_id": p.ID, "directory": t.TempDir(), "recovery_address": "127.0.0.1:17777"}); err == nil {
		t.Fatal("bind failure ignored")
	}
	if c.enabled || c.browser.server != nil || c.browser.expectedPAC != "" {
		t.Fatal("failed Edge prepare changed Chrome/capture")
	}
}
func TestEdgePolicyJournalBeforeRegistryWriteAndGuardNoDirectLeak(t *testing.T) {
	c := New(context.Background())
	defer c.Close()
	c.BrowserOwnerSID = "owner"
	if runtime.GOOS == "windows" {
		self, err := wp.NewDiscovery().Lookup(uint32(os.Getpid()))
		if err != nil {
			t.Fatal(err)
		}
		c.BrowserOwnerSID, err = wp.ProcessUserSID(self)
		if err != nil {
			t.Fatal(err)
		}
	}
	owner := c.BrowserOwnerSID
	dir := t.TempDir()
	if err := c.SetCustomPath(filepath.Join(dir, "profiles.json")); err != nil {
		t.Fatal(err)
	}
	p := edgeTestProfile()
	p.Rules = []flowpolicy.Rule{{ID: "private", Domains: []string{"private.example"}, Action: flowpolicy.RelayRequired, RelayID: "VPS-01"}}
	c.Profiles.Upsert(p)
	e := c.edgeLocked()
	writes := 0
	e.write = func(sid string, owned []string, next string) error {
		writes++
		if sid != owner {
			t.Fatal("wrong owner hive")
		}
		b, err := os.ReadFile(filepath.Join(dir, "edge-transport.json"))
		if err != nil || !strings.Contains(string(b), "owned_policy") {
			t.Fatal("registry written before durable ownership journal")
		}
		if len(owned) > 2 || owned[len(owned)-1] != next {
			t.Fatal("unbounded/incorrect owned values")
		}
		return nil
	}
	if _, err := edgeCall(c, "edge.policy", map[string]string{"profile_id": p.ID, "operation": "guard"}); err != nil {
		t.Fatal(err)
	}
	if writes != 1 || c.browser.server != nil {
		t.Fatal("guard changed Chrome")
	}
	pac := edgeClosedPAC(p)
	raw, _ := base64.StdEncoding.DecodeString(strings.TrimPrefix(pac, "data:application/x-ns-proxy-autoconfig;base64,"))
	for _, term := range []string{"chatgpt.com", "private.example", "youtube.com", "fail-closed.neexitlag.invalid", "return strict?"} {
		if !strings.Contains(string(raw), term) {
			t.Fatal("guard missing", term)
		}
	}
	if _, err := edgeCall(c, "edge.policy", map[string]string{"profile_id": p.ID, "operation": "remove"}); err == nil {
		t.Fatal("protected guard could be removed")
	}
}
func TestLegacyEdgeProfilesPreserveRelayRequiredUntilExplicitSitesOptIn(t *testing.T) {
	c := New(context.Background())
	defer c.Close()
	p := edgeTestProfile()
	p.Mode = profiles.Auto
	c.Profiles.Upsert(p)
	c.enabled = true
	id := wp.ProcessIdentity{PID: 11, CreationTime: 1, Path: p.Executables[0]}
	d := c.Classify(id, policyTuple("tcp", 443))
	if d.Route.RelayID != "VPS-01" || !d.BlockOnFailure || browserAuto(p) {
		t.Fatal("legacy Edge semantics changed", d)
	}
	p.AppPreset = "edge-sites"
	c.Profiles.Upsert(p)
	d = c.Classify(id, policyTuple("tcp", 53))
	if !d.Route.Direct() || d.Deferred || d.Policy.Source != "browser_dns" {
		t.Fatal("Edge DNS hijacked", d)
	}
	d = wp.ResolveDeferredDecision(c.Classify(id, policyTuple("tcp", 443)), "chatgpt.com", "SNI_FOUND")
	if d.Route.RelayID != "VPS-01" || !d.BlockOnFailure {
		t.Fatal("protected Edge leak", d)
	}
	d = wp.ResolveDeferredDecision(c.Classify(id, policyTuple("tcp", 443)), "ordinary.example", "SNI_FOUND")
	if !d.Route.Direct() || d.Block {
		t.Fatal("ordinary Edge not DIRECT", d)
	}
}
func TestEdgeGuardPublishedAlongsideUnchangedChrome(t *testing.T) {
	c := New(context.Background())
	defer c.Close()
	dir := t.TempDir()
	if err := c.SetCustomPath(filepath.Join(dir, "profiles.json")); err != nil {
		t.Fatal(err)
	}
	edge := edgeTestProfile()
	chrome := profiles.Profile{ID: "chrome", Name: "Chrome", Mode: profiles.Direct, Executables: []string{`C:\Chrome\chrome.exe`}}
	c.Profiles.Upsert(edge)
	c.Profiles.Upsert(chrome)
	if err := c.publishBrowserGuard(false); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(c.browserGuardFile())
	var g browserproxy.RecoveryGuard
	if json.Unmarshal(raw, &g) != nil || len(g.Profiles) != 2 || g.Profiles[edge.ID].Executables[0] != edge.Executables[0] || g.Profiles[chrome.ID].Executables[0] != chrome.Executables[0] {
		t.Fatal("guard identities crossed")
	}
}
func TestEdgeRestoreFailureCannotPreventChromeRestore(t *testing.T) {
	c := New(context.Background())
	defer c.Close()
	dir := t.TempDir()
	if err := c.SetCustomPath(filepath.Join(dir, "profiles.json")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "edge-transport.json"), []byte(`{"address":"8.8.8.8:443"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := c.RestoreBrowserTransport(); err != nil {
		t.Fatal("Edge blocked independent Chrome restore", err)
	}
	if c.browser.edge.lastError == "" || c.browser.edge.server != nil {
		t.Fatal("invalid Edge state reported healthy")
	}
}

func TestEdgeGuardOnlyJournalSurvivesRestartBeforePrepare(t *testing.T) {
	c := New(context.Background())
	defer c.Close()
	dir := t.TempDir()
	if err := c.SetCustomPath(filepath.Join(dir, "profiles.json")); err != nil {
		t.Fatal(err)
	}
	c.browser.edge = &edgeRuntime{saved: edgeSaved{Owned: []string{edgePolicyValue(edgeClosedPAC(edgeTestProfile()))}}}
	if err := c.saveEdge(c.browser.edge); err != nil {
		t.Fatal(err)
	}
	c.browser.edge = nil
	c.restoreEdgeTransport()
	if c.browser.edge == nil || c.browser.edge.lastError != "" || c.browser.edge.server != nil || len(c.browser.edge.saved.Owned) != 1 {
		t.Fatal("guard-only ownership lost", c.browser.edge)
	}
}
