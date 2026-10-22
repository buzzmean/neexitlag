package clientservice

import (
	"github.com/buzzmean/neexitlag/core/profiles"
	"testing"
)

func TestNativeReadinessIndependentOfPACAndNotPageSuccess(t *testing.T) {
	c, p, b := nativeEdgeFixture(t)
	p.Mode = profiles.Auto
	p.BrowserTransport = "native"
	c.Profiles.Upsert(p)
	c.enabled = true
	c.settings.Automatic = true
	c.browser.readiness.enforced.Store(true)
	if c.nativeSnapshot().State != "PREPARING" {
		t.Fatal(c.nativeSnapshot())
	}
	b.ready = true
	c.browser.readiness.relay = func() bool { return true }
	s := c.nativeSnapshot()
	if s.State != "READY_NEW_CONNECTIONS" || s.NativeOnlyState != "NATIVE_ONLY_NOT_OBSERVED" || s.ConfiguredProfiles != 1 {
		t.Fatal(s)
	}
	c.enabled = false
	if c.nativeSnapshot().State != "DISABLED" {
		t.Fatal(c.nativeSnapshot())
	}
}
