package clientservice

import (
	"context"
	"github.com/buzzmean/neexitlag/core/controlclient"
	"github.com/buzzmean/neexitlag/core/security"
	"testing"
)

func TestActivationCloseRemovesPolicyBeforeListener(t *testing.T) {
	c, value, _, _ := policyReadyFixture(t, false)
	if _, err := c.Handle(context.Background(), "browser.silent.apply", []byte(`{"profile_id":"browser"}`)); err != nil {
		t.Fatal(err)
	}
	removed := false
	c.browser.chromeRemove = func(_ string, _ []string) error {
		if *value != "" && (c.browser.server == nil || !c.browser.server.Snapshot().Listening) {
			t.Error("listener stopped before policy removal")
		}
		*value = ""
		removed = true
		return nil
	}
	c.Close()
	if !removed || *value != "" {
		t.Fatal("policy retained")
	}
}
func TestActivationRevocationRemovesPACBeforeRelayClose(t *testing.T) {
	c, value, _, _ := policyReadyFixture(t, false)
	defer c.Close()
	if _, err := c.Handle(context.Background(), "browser.silent.apply", []byte(`{"profile_id":"browser"}`)); err != nil {
		t.Fatal(err)
	}
	s := remoteState{controller: c, cp: &controlclient.Client{DeviceID: "revoked"}}
	if err := s.apply(context.Background(), security.Config{RevokedDevices: []string{"revoked"}}); err == nil {
		t.Fatal("revocation accepted")
	}
	c.RetrySilentBrowserPolicies()
	if c.activationNetworkAllowed() {
		t.Fatal("revoked device capture still permitted")
	}
	if *value != "" {
		t.Fatal("revoked identity retained PAC")
	}
}
