package gameudp

import (
	"net/netip"
	"strings"
	"testing"
	"time"
)

func ladderFlow(t *testing.T, e *Engine, at time.Time, port uint16) *Flow {
	t.Helper()
	f, err := e.Open(Key{Network: strings.Repeat("b", 64), Application: app, PID: 77, Birth: 100, Source: netip.AddrPortFrom(netip.MustParseAddr("192.0.2.1"), port), Destination: netip.MustParseAddrPort("18.157.46.1:7766")}, at)
	if err != nil {
		t.Fatal(err)
	}
	return f
}
func failTrial(e *Engine, f *Flow, at time.Time) {
	p, k := e.Plan(f, at)
	for i := 0; i < k; i++ {
		e.FakeSent(f, len(p)+28, nil)
	}
	for i := 0; i <= 7; i++ {
		now := at.Add(time.Duration(i) * time.Second)
		e.Outgoing(f, 50, now)
		e.Tick(now)
	}
}
func TestLadderDirectLocalVariantsRelayThenCooldown(t *testing.T) {
	e := NewLadder(&Budget{})
	e.Configure(app, true)
	e.SetReason("")
	at := time.Unix(1000, 0)
	direct := ladderFlow(t, e, at, 50000)
	if direct.e.Stage != Direct {
		t.Fatal(direct.e)
	}
	if _, k := e.Plan(direct, at); k != 0 {
		t.Fatal("DIRECT primed")
	}
	failTrial(e, direct, at)
	at = at.Add(8 * time.Second)
	for i, v := range Variants() {
		f := ladderFlow(t, e, at, uint16(50001+i))
		if f.e.Stage != Local || f.e.Variant != v {
			t.Fatal(f.e)
		}
		failTrial(e, f, at)
		at = at.Add(8 * time.Second)
	}
	relay := ladderFlow(t, e, at, 50020)
	if relay.e.Stage != Relay {
		t.Fatal(relay.e)
	}
	if _, k := e.Plan(relay, at); k != 0 {
		t.Fatal("relay primed")
	}
	if direct.e.Stage != Direct {
		t.Fatal("existing tuple migrated")
	}
	failTrial(e, relay, at)
	if _, err := e.Open(relay.key, at.Add(8*time.Second)); err == nil {
		t.Fatal("unbounded ladder")
	}
	f := ladderFlow(t, e, at.Add(2*time.Hour), 50021)
	if f.e.Stage != Direct {
		t.Fatal("no cooldown recovery")
	}
}
func TestLadderHealthyDirectNeverUsesLocalOrRelay(t *testing.T) {
	e := NewLadder(nil)
	e.Configure(app, true)
	e.SetReason("")
	at := time.Unix(1000, 0)
	f := ladderFlow(t, e, at, 50000)
	sustain(e, f, at)
	next := ladderFlow(t, e, at.Add(5*time.Second), 50001)
	if next.e.Stage != Direct || len(e.Snapshot().Learning) != 1 || !e.Snapshot().Learning[0].Confirmed {
		t.Fatal(e.Snapshot())
	}
}
func TestGlobalBudgetAcrossGames(t *testing.T) {
	b := &Budget{}
	total := 0
	for i := 0; i < 30; i++ {
		e, _, at := fixture(t)
		e.shared = b
		f := ladderFlow(t, e, at, 50000)
		_, k := e.Plan(f, at)
		total += k
	}
	if total > MaxFakePerSecond {
		t.Fatal(total)
	}
}
func TestLadderRestoredLearningIsApplicationAndNetworkScoped(t *testing.T) {
	e := NewLadder(nil)
	e.Configure(app, true)
	e.SetReason("")
	at := time.Unix(1000, 0)
	f := ladderFlow(t, e, at, 50000)
	failTrial(e, f, at)
	v := e.Snapshot().Learning
	copy := NewLadder(nil)
	copy.Configure(app, true)
	copy.SetReason("")
	if err := copy.Restore(v, at.Add(8*time.Second)); err != nil {
		t.Fatal(err)
	}
	if ladderFlow(t, copy, at.Add(9*time.Second), 50001).e.Stage != Local {
		t.Fatal("lost failed DIRECT")
	}
	k := f.key
	k.Network = strings.Repeat("c", 64)
	g, err := copy.Open(k, at.Add(10*time.Second))
	if err != nil || g.e.Stage != Direct {
		t.Fatal("cross-network learning", err)
	}
}
