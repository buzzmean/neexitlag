package clientservice

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/buzzmean/neexitlag/core/gameudp"
	"github.com/buzzmean/neexitlag/core/profiles"
	wp "github.com/buzzmean/neexitlag/platform/windows"
)

func gameController(t *testing.T) (*Controller, wp.ProcessIdentity) {
	path := `C:\Steam\steamapps\common\PUBG\TslGame\Binaries\Win64\TslGame.exe`
	p := profiles.Profile{ID: "pubg", Name: "PUBG", Executables: []string{path}, Mode: profiles.Direct}
	c := policyController(t, p)
	c.registerGame(gameudp.Presets()[0], path)
	g := c.games.items["pubg"]
	if g == nil {
		t.Fatal("PUBG not registered")
	}
	g.Engine.SetReason("")
	c.network.Store(&wp.NetworkContext{Fingerprint: strings.Repeat("a", 64), ObservedAt: time.Now()})
	return c, wp.ProcessIdentity{PID: 42, CreationTime: 100, Path: path}
}
func TestGameDefaultPUBGAndOtherOptInScope(t *testing.T) {
	c, id := gameController(t)
	preferGameRelay(c)
	d := c.Classify(id, policyTuple("udp", 7766))
	if d.GameUDP == nil || !d.Route.Direct() {
		t.Fatal(d)
	}
	for _, tc := range []struct {
		proto string
		port  uint16
	}{{"tcp", 443}, {"udp", 53}, {"udp", 443}} {
		d = c.Classify(id, policyTuple(tc.proto, tc.port))
		if d.GameUDP != nil || !d.Route.Direct() {
			t.Fatal("low port/TCP changed", d)
		}
	}
	d = c.Classify(wp.ProcessIdentity{PID: 43, CreationTime: 101, Path: `C:\Other\TslGame.exe`}, policyTuple("udp", 7766))
	if d.GameUDP != nil || !d.Route.Direct() {
		t.Fatal("basename inherited feature", d)
	}
	p := gameudp.Presets()[1]
	path := `C:\Steam\steamapps\common\CS2\game\bin\win64\cs2.exe`
	c.registerGame(p, path)
	if g := c.games.items["cs2"]; g == nil || g.Enabled {
		t.Fatal("CS2 automatically modified", g)
	}
	for _, p := range gameudp.Presets() {
		if p.ID != "pubg" && p.DefaultEnabled {
			t.Fatal("unverified anticheat enabled", p)
		}
	}
}
func TestGameOptOutPersistsAndRediscoveryDoesNotEnable(t *testing.T) {
	c, id := gameController(t)
	c.customPath = filepath.Join(t.TempDir(), "profiles.json")
	if _, err := c.handleGameUDP(context.Background(), "gameudp.set", json.RawMessage(`{"profile_id":"pubg","enabled":false}`)); err != nil {
		t.Fatal(err)
	}
	c.registerGame(gameudp.Presets()[0], id.Path)
	if c.games.items["pubg"].Enabled {
		t.Fatal("optout overwritten")
	}
	fresh := New(context.Background())
	defer fresh.Close()
	fresh.customPath = c.customPath
	fresh.loadGames()
	if fresh.games.error != "" || fresh.games.items["pubg"] == nil || fresh.games.items["pubg"].Enabled {
		t.Fatal("optout lost", fresh.games.error)
	}
}
func TestGameCannotOverrideProtectedOrAnticheatProfiles(t *testing.T) {
	for _, path := range []string{`C:\Apps\ChatGPT.exe`, `C:\Apps\chrome.exe`, `C:\Riot\VALORANT-Win64-Shipping.exe`, `C:\FACEIT\FACEIT.exe`} {
		p := profiles.Profile{ID: "custom-game", Name: "selected", Executables: []string{path}, Mode: profiles.Direct, Custom: true}
		c := policyController(t, p)
		if _, err := c.handleGameUDP(context.Background(), "gameudp.set", json.RawMessage(`{"profile_id":"custom-game","enabled":true}`)); err == nil {
			t.Fatal("unsafe game selection", path)
		}
	}
}
func TestGameArbitraryFullPathAndPerGameIsolation(t *testing.T) {
	c, id := gameController(t)
	path := `D:\Games\Other\game.exe`
	p := profiles.Profile{ID: "other", Name: "Other game", Executables: []string{path}, Mode: profiles.Direct, Custom: true}
	if err := c.Profiles.Upsert(p); err != nil {
		t.Fatal(err)
	}
	if _, err := c.handleGameUDP(context.Background(), "gameudp.set", json.RawMessage(`{"profile_id":"other","enabled":true}`)); err != nil {
		t.Fatal(err)
	}
	g := c.games.items["other"]
	if g == nil || g.Path != path || g.Engine == c.games.items["pubg"].Engine {
		t.Fatal("not isolated", g)
	}
	g.Engine.SetReason("")
	other := id
	other.Path = path
	other.PID++
	d := c.Classify(other, policyTuple("udp", 7766))
	if d.GameUDP != g.Engine {
		t.Fatal(d)
	}
}
func TestGamePathValidation(t *testing.T) {
	for _, p := range []string{"game.exe", `\\server\share\game.exe`, `C:\A\..\game.exe`, "C:/x\n/game.exe"} {
		if gamePath(p) {
			t.Fatal("invalid path", p)
		}
	}
	if !gamePath(`C:\Users\USER\Games\game.exe`) {
		t.Fatal("owner local profile path rejected")
	}
}
