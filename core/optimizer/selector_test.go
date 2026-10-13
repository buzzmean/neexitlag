package optimizer

import (
	"math"
	"testing"
	"time"
)

func sample(id string, rtt float64, loss float64, now time.Time) RouteMeasurement {
	return RouteMeasurement{RouteID: id, RTT: time.Duration(rtt * float64(time.Millisecond)), Loss: loss, Stability: 1, Available: true, ObservedAt: now, Samples: 10}
}
func TestScoringHysteresis(t *testing.T) {
	now := time.Unix(1000, 0)
	for _, tc := range []struct {
		direct, relay, loss, relayLoss float64
		want                           string
	}{{24, 28, 0, 0, Direct}, {55, 37, .025, .001, "FRA-01"}, {30, 29, 0, 0, Direct}} {
		s := New(Parameters{5, 20 * time.Second, 10 * time.Second, time.Minute})
		s.Observe(sample(Direct, tc.direct, tc.loss, now))
		s.Observe(sample("FRA-01", tc.relay, tc.relayLoss, now))
		if got := s.Select(now).Primary; got != Direct {
			t.Fatal("early switch", got)
		}
		if got := s.Select(now.Add(21 * time.Second)).Primary; got != tc.want {
			t.Fatalf("want %s got %s", tc.want, got)
		}
	}
}
func TestConfirmationResetAndFailover(t *testing.T) {
	n := time.Unix(1000, 0)
	s := New(Parameters{5, 0, 10 * time.Second, time.Minute})
	s.Observe(sample(Direct, 80, 0, n))
	s.Observe(sample("FRA", 20, 0, n))
	s.Observe(sample("WAW", 35, 0, n))
	s.Select(n)
	if got := s.Select(n.Add(11 * time.Second)); got.Primary != "FRA" || got.Secondary != "WAW" || got.Tertiary != Direct {
		t.Fatal(got)
	}
	m := sample("FRA", 20, 0, n.Add(12*time.Second))
	m.Available = false
	s.Observe(m)
	if got := s.Select(n.Add(12 * time.Second)); got.Primary != "WAW" {
		t.Fatal("no immediate failover", got)
	}
	if got := s.Select(n.Add(2 * time.Minute)); got.Primary != Direct {
		t.Fatal("stale routes used", got)
	}
}
func TestInvalidAndUnknown(t *testing.T) {
	n := time.Now()
	s := New(DefaultParameters())
	s.Observe(sample("FRA", 1, 0, n))
	if s.Select(n.Add(time.Second)).Primary != Direct {
		t.Fatal("unknown direct optimized")
	}
	m := sample(Direct, 1, math.NaN(), n)
	if !math.IsInf(Score(m), 1) {
		t.Fatal("NaN accepted")
	}
}
func TestCandidateMustStayBetter(t *testing.T) {
	n := time.Unix(10, 0)
	s := New(Parameters{5, 0, 10 * time.Second, time.Minute})
	s.Observe(sample(Direct, 50, 0, n))
	s.Observe(sample("FRA", 20, 0, n))
	s.Select(n)
	s.Observe(sample("FRA", 60, 0, n.Add(5*time.Second)))
	s.Select(n.Add(5 * time.Second))
	s.Observe(sample("FRA", 20, 0, n.Add(7*time.Second)))
	if s.Select(n.Add(12*time.Second)).Primary != Direct {
		t.Fatal("noncontinuous confirmation")
	}
	if s.Select(n.Add(23*time.Second)).Primary != "FRA" {
		t.Fatal("confirmation not honored")
	}
}
