package clientservice

import (
	"context"
	"github.com/buzzmean/neexitlag/core/profiles"
	"os"
	"path/filepath"
	"testing"
)

func TestClubSeedsBrowsersBeforeActivationAndBlocksOwnerMutations(t *testing.T) {
	c := New(context.Background())
	defer c.Close()
	dir := t.TempDir()
	if err := c.SetCustomPath(filepath.Join(dir, "state", "profiles.json")); err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, name := range []string{"chrome.exe", "msedge.exe"} {
		p := filepath.Join(dir, name)
		os.WriteFile(p, []byte("fixture"), 0600)
		paths = append(paths, p)
	}
	if err := c.ConfigureClub(Manifest{ClubImage: true, ClubBrowserPaths: paths}); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"club-chrome", "club-edge"} {
		var p profiles.Profile
		ok := false
		for _, candidate := range c.Profiles.List() {
			if candidate.ID == id {
				p = candidate
				ok = true
			}
		}
		if !ok || p.Mode != profiles.Auto || p.BrowserTransport != "native" {
			t.Fatalf("browser not AUTO native: %s", id)
		}
	}
	if !c.Snapshot().Settings.Automatic {
		t.Fatal("automatic off")
	}
	for _, method := range []string{"optimize", "settings.set", "profile.set", "profile.delete", "activation.set", "gameudp.set", "browser.native.enable", "future.mutation"} {
		if _, err := c.HandleOwnerIPC(context.Background(), method, []byte(`{}`)); err == nil || err.Error() != "CLUB_ADMIN_MANAGED" {
			t.Fatalf("mutation accepted %s: %v", method, err)
		}
	}
	if _, err := c.HandleOwnerIPC(context.Background(), "status", nil); err != nil {
		t.Fatal(err)
	}
	// Internal activation automation must remain usable, not routed through owner gate.
	if _, err := c.Handle(context.Background(), "optimize", []byte(`{"enabled":false}`)); err != nil && err.Error() == "CLUB_ADMIN_MANAGED" {
		t.Fatal(err)
	}
	if err := c.ConfigureClub(Manifest{ClubImage: true, ClubBrowserPaths: paths}); err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, p := range c.Profiles.List() {
		if p.ID == "club-chrome" || p.ID == "club-edge" {
			n++
		}
	}
	if n != 2 {
		t.Fatal("duplicate browser profiles")
	}
}
func TestNonClubOwnerContractUnchanged(t *testing.T) {
	c := New(context.Background())
	defer c.Close()
	if err := c.ConfigureClub(Manifest{}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.HandleOwnerIPC(context.Background(), "optimize", []byte(`{"enabled":false}`)); err != nil && err.Error() == "CLUB_ADMIN_MANAGED" {
		t.Fatal(err)
	}
}
