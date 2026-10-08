package clientservice

import (
	"context"
	"encoding/json"
	"github.com/buzzmean/neexitlag/core/optimizer"
	"github.com/buzzmean/neexitlag/core/profiles"
	windowsplatform "github.com/buzzmean/neexitlag/platform/windows"
	"net/netip"
	"path/filepath"
	"testing"
	"time"
)

func TestServiceUnavailableProfilesAndAuthenticatedCommandsBoundary(t *testing.T) {
	c := New(context.Background())
	defer c.Close()
	if c.Snapshot().Connected {
		t.Fatal("false connected state")
	}
	if _, e := c.Handle(context.Background(), "optimize", json.RawMessage(`{"enabled":true}`)); e == nil {
		t.Fatal("missing backend accepted")
	}
	if e := c.SetCustomPath(filepath.Join(t.TempDir(), "profiles.json")); e != nil {
		t.Fatal(e)
	}
	p := profiles.Profile{ID: "custom", Name: "Game", Executables: []string{`C:\Games\Game.exe`}, Mode: profiles.Auto, Custom: true}
	b, _ := json.Marshal(p)
	if _, e := c.Handle(context.Background(), "profile.set", b); e != nil {
		t.Fatal(e)
	}
	c.mu.Lock()
	c.enabled = true
	c.mu.Unlock()
	now := time.Now()
	c.Selector = optimizer.New(optimizer.Parameters{MinimumImprovement: 5, MaxAge: time.Minute})
	c.Selector.Observe(optimizer.RouteMeasurement{RouteID: optimizer.Direct, RTT: 50 * time.Millisecond, Stability: 1, Available: true, Samples: 3, ObservedAt: now})
	c.Selector.Observe(optimizer.RouteMeasurement{RouteID: "FRA", RTT: 20 * time.Millisecond, Stability: 1, Available: true, Samples: 3, ObservedAt: now})
	c.Selector.Select(now)
	tuple := windowsplatform.Tuple{Protocol: "udp", Source: netip.MustParseAddrPort("10.0.0.1:123"), Destination: netip.MustParseAddrPort("1.1.1.1:456")}
	identity := windowsplatform.ProcessIdentity{PID: 42, Path: `C:\Games\Game.exe`, CreationTime: 123}
	decision := c.Classify(identity, tuple)
	if decision.Route.RelayID != "FRA" || decision.Flow.PID != 42 {
		t.Fatal("real identity classification", decision)
	}
	identity.Path = `C:\Other\Game.exe`
	if !c.Classify(identity, tuple).Route.Direct() {
		t.Fatal("non-selected executable routed")
	}
	c2 := New(context.Background())
	defer c2.Close()
	if e := c2.SetCustomPath(c.customPath); e != nil {
		t.Fatal(e)
	}
	if len(c2.Profiles.List()) != 4 {
		t.Fatal("local profile persistence")
	}
}
