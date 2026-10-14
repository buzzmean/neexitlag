package clientservice

import (
	"context"
	"encoding/json"
	"github.com/buzzmean/neexitlag/core/localstrategy"
	"github.com/buzzmean/neexitlag/core/optimizer"
	"github.com/buzzmean/neexitlag/core/profiles"
	windowsplatform "github.com/buzzmean/neexitlag/platform/windows"
	"net/netip"
	"path/filepath"
	"testing"
	"time"
)

func TestLocalIPCAndAutoTargetIsolation(t *testing.T) {
	c := New(context.Background())
	defer c.Close()
	c.SetCustomPath(filepath.Join(t.TempDir(), "profiles.json"))
	cfg := localstrategy.Config{Version: 1, Enabled: true, User: []localstrategy.Profile{{ID: "split", Version: 1, Application: "chosen.exe", Protocol: "tcp", Ports: []uint16{443}, IPs: []string{"203.0.113.9"}, Strategy: "tcp-split-1", Cutoff: 1, Enabled: true}}}
	b, _ := json.Marshal(cfg)
	if _, err := c.Handle(context.Background(), "local.set", b); err != nil {
		t.Fatal(err)
	}
	other := New(context.Background())
	defer other.Close()
	if err := other.SetCustomPath(c.customPath); err != nil || !other.Local.Config().Enabled {
		t.Fatal("persist", err)
	}
	c.network.Store(&windowsplatform.NetworkContext{Fingerprint: "net", ObservedAt: time.Now()})
	c.enabled = true
	c.Profiles.Upsert(profiles.Profile{ID: "app", Name: "app", Executables: []string{"chosen.exe"}, Mode: profiles.Auto})
	id := windowsplatform.ProcessIdentity{PID: 42, CreationTime: 100, Path: "chosen.exe"}
	tuple := windowsplatform.Tuple{Protocol: "tcp", Source: netip.MustParseAddrPort("192.0.2.1:1234"), Destination: netip.MustParseAddrPort("203.0.113.9:443")}
	k, err := c.localKey(id.Path, tuple.Destination.String(), "tcp")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	good := optimizer.RouteMeasurement{Available: true, Samples: 10, ObservedAt: now, RTT: 20 * time.Millisecond, Stability: 1}
	bad := good
	bad.Available = false
	c.Learned.Record(k, "DIRECT", bad, false, now)
	c.Learned.Record(k, "split", good, true, now)
	if d := c.Classify(id, tuple); !d.Route.Local() || d.LocalStrategyID != "split" {
		t.Fatal(d)
	}
	c.network.Store(&windowsplatform.NetworkContext{Fingerprint: "changed", ObservedAt: now})
	if c.Classify(id, tuple).Route.Local() {
		t.Fatal("cache crossed network")
	}
	c.network.Store(&windowsplatform.NetworkContext{Fingerprint: "net", ObservedAt: now, Conflicts: []string{"winws"}})
	if !c.Classify(id, tuple).Route.Direct() || c.Classify(id, tuple).Route.Local() {
		t.Fatal("conflict not failopen")
	}
	c.network.Store(&windowsplatform.NetworkContext{Fingerprint: "net", ObservedAt: now})
	cfg.User[0].Cutoff = 2
	b, _ = json.Marshal(cfg)
	c.Handle(context.Background(), "local.set", b)
	if c.Classify(id, tuple).Route.Local() {
		t.Fatal("old evidence survived config change")
	}
}

func TestTargetFailureSelectsActualRelayInsteadOfGlobalDirect(t *testing.T) {
	c := New(context.Background())
	defer c.Close()
	c.enabled = true
	c.network.Store(&windowsplatform.NetworkContext{Fingerprint: "net", ObservedAt: time.Now()})
	id := windowsplatform.ProcessIdentity{PID: 42, CreationTime: 100, Path: "chosen.exe"}
	tuple := windowsplatform.Tuple{Protocol: "tcp", Source: netip.MustParseAddrPort("192.0.2.1:1234"), Destination: netip.MustParseAddrPort("203.0.113.9:443")}
	c.Profiles.Upsert(profiles.Profile{ID: "app", Name: "app", Executables: []string{"chosen.exe"}, Mode: profiles.Auto})
	now := time.Now()
	k, _ := c.localKey(id.Path, tuple.Destination.String(), "tcp")
	bad := optimizer.RouteMeasurement{Available: false, Samples: 10, ObservedAt: now}
	relay := optimizer.RouteMeasurement{RouteID: "VPS-01", Available: true, Samples: 10, ObservedAt: now, RTT: 20 * time.Millisecond, Stability: 1}
	c.Learned.Record(k, "DIRECT", bad, false, now)
	c.Learned.Record(k, "RELAY", relay, true, now)
	if got := c.Classify(id, tuple).Route.RelayID; got != "VPS-01" {
		t.Fatal("global direct overrode target evidence", got)
	}
}
