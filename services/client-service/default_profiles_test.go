package clientservice

import (
	"context"
	"net/netip"
	"testing"

	"github.com/buzzmean/neexitlag/core/profiles"
	wp "github.com/buzzmean/neexitlag/platform/windows"
)

func destTuple(proto, addr string, port uint16) wp.Tuple {
	return wp.Tuple{Protocol: proto, Source: netip.MustParseAddrPort("192.0.2.1:55000"), Destination: netip.AddrPortFrom(netip.MustParseAddr(addr), port)}
}

// A clean install must cover the blocked desktop apps without any manual profile.
// Packaged apps live in versioned WindowsApps folders, so matching is by exe name.
func TestBuiltInProfilesRouteBlockedDesktopAppsRelayFirstFailClosed(t *testing.T) {
	c := New(context.Background())
	t.Cleanup(c.Close)
	c.enabled = true
	cases := []struct {
		name, path, addr string
		port             uint16
	}{
		{"claude v1", `C:\Program Files\WindowsApps\Claude_1.0.100.0_x64__pzs8sxrjxfjjc\app\claude.exe`, "160.79.104.10", 443},
		{"claude v2 other folder", `C:\Program Files\WindowsApps\Claude_1.4.0.0_x64__pzs8sxrjxfjjc\app\claude.exe`, "160.79.104.10", 443},
		{"chatgpt", `C:\Program Files\WindowsApps\OpenAI.ChatGPT-Desktop_1.2025.1.0_x64__2p2nqsd0c76g0\app\ChatGPT.exe`, "104.18.39.21", 443},
		{"chatgpt next version", `C:\Program Files\WindowsApps\OpenAI.ChatGPT-Desktop_1.2026.9.0_x64__2p2nqsd0c76g0\app\ChatGPT.exe`, "104.18.39.21", 443},
		{"telegram tcp 443", `C:\Users\u\AppData\Roaming\Telegram Desktop\Telegram.exe`, "149.154.167.51", 443},
		{"telegram tcp 80", `C:\Users\u\AppData\Roaming\Telegram Desktop\Telegram.exe`, "149.154.175.50", 80},
		{"telegram tcp 5222", `D:\Apps\Telegram Desktop\Telegram.exe`, "91.108.56.130", 5222},
	}
	for i, tc := range cases {
		id := wp.ProcessIdentity{PID: uint32(100 + i), CreationTime: uint64(10 + i), Path: tc.path}
		d := c.Classify(id, destTuple("tcp", tc.addr, tc.port))
		if d.Route.RelayID != "VPS-01" || !d.BlockOnFailure || d.Block || d.Route.Direct() {
			t.Fatalf("%s: not relay-first fail-closed: %+v", tc.name, d)
		}
		if _, err := c.Engine.OpenTCP(context.Background(), d.Flow); err == nil {
			t.Fatalf("%s: leaked to DIRECT while the relay is unavailable", tc.name)
		}
	}
	other := wp.ProcessIdentity{PID: 900, CreationTime: 90, Path: `C:\Tools\notes.exe`}
	if d := c.Classify(other, destTuple("tcp", "160.79.104.10", 443)); !d.Route.Direct() || d.BlockOnFailure {
		t.Fatalf("unrelated application was captured: %+v", d)
	}
}

// Existing installs may already have a manual profile for the same program; the
// built-in one must not shadow it (two profiles for one exe would be ambiguous).
func TestPersistedProfileSuppressesShadowedBuiltIn(t *testing.T) {
	c := New(context.Background())
	t.Cleanup(c.Close)
	manual := profiles.Profile{ID: "my-telegram", Name: "Telegram", Executables: []string{`C:\Apps\Telegram\Telegram.exe`}, Mode: profiles.Auto, Custom: true}
	dropShadowedBuiltIns(c.Profiles, []profiles.Profile{manual})
	for _, p := range c.Profiles.List() {
		if p.ID == "telegram" {
			t.Fatal("built-in Telegram profile still registered next to the user's own")
		}
	}
	found := map[string]bool{}
	for _, p := range c.Profiles.List() {
		found[p.ID] = true
	}
	for _, id := range []string{"cs2", "discord", "steam", "claude-desktop", "chatgpt-desktop"} {
		if !found[id] {
			t.Fatalf("unrelated built-in %q was removed", id)
		}
	}
}
