package clientservice

import (
	"github.com/buzzmean/neexitlag/core/flowpolicy"
	"github.com/buzzmean/neexitlag/core/profiles"
	wp "github.com/buzzmean/neexitlag/platform/windows"
	"testing"
)

func TestFreshNativeDoesNotWaitForPACAndNeverAuthorizesOpaqueDirect(t *testing.T) {
	for _, edge := range []bool{false, true} {
		c, p, _ := nativeEdgeFixture(t)
		if !edge {
			p.Executables = []string{`C:\Apps\chrome.exe`}
			p.AppPreset = ""
		}
		p.Mode = profiles.Auto
		p.BrowserTransport = "native"
		c.Profiles.Upsert(p)
		c.enabled = true
		c.settings.Automatic = true
		c.browser.readiness.enforced.Store(true)
		id := wp.ProcessIdentity{PID: 77, CreationTime: 1, Path: p.Executables[0]}
		initial := c.Classify(id, policyTuple("tcp", 443))
		if !initial.Deferred || initial.OpenDeferred == nil || !initial.NewConnectionsOnly {
			t.Fatal("native still waits for PAC", initial)
		}
		for _, host := range []string{"chatgpt.com", "claude.ai", "web.telegram.org"} {
			d := wp.ResolveDeferredDecision(initial, host, "TLS_SNI")
			if d.Policy.Action != flowpolicy.RelayRequired || !d.BlockOnFailure || d.Route.RelayID != "VPS-01" {
				t.Fatal(host, d)
			}
		}
		opaque := wp.ResolveDeferredDecision(initial, "cloudflare-ech.com", "ECH_OUTER_SNI_ONLY")
		if opaque.Policy.Action != flowpolicy.RelayRequired || !opaque.BlockOnFailure || opaque.Route.Direct() {
			t.Fatal("ECH leaked", opaque)
		}
		ordinary := wp.ResolveDeferredDecision(initial, "ordinary.example", "TLS_SNI")
		if !ordinary.Route.Direct() || ordinary.Block {
			t.Fatal(ordinary)
		}
		quic := c.Classify(id, policyTuple("udp", 443))
		if quic.Route.Direct() || !quic.BlockOnFailure {
			t.Fatal("opaque QUIC leaked", quic)
		}
		dns := c.Classify(id, policyTuple("udp", 53))
		if !dns.Route.Direct() || dns.Deferred {
			t.Fatal(dns)
		}
		c.settings.Automatic = false
		off := c.Classify(id, policyTuple("tcp", 443))
		if !off.Route.Direct() || off.Block || off.Deferred || off.OpenDeferred != nil {
			t.Fatal("automatic off trapped browser", off)
		}
	}
}
