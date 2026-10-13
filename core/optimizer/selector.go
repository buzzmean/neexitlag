// Package optimizer scores fresh route observations and selects routes for new flows.
package optimizer

import (
	"errors"
	"math"
	"sort"
	"sync"
	"time"
)

const Direct = "DIRECT"

type RelayCandidate struct {
	ID      string  `json:"id"`
	Address string  `json:"address,omitempty"`
	Region  string  `json:"region,omitempty"`
	City    string  `json:"city,omitempty"`
	Load    float64 `json:"load,omitempty"`
}
type RouteMeasurement struct {
	RouteID     string        `json:"route_id"`
	RTT         time.Duration `json:"rtt"`
	Jitter      time.Duration `json:"jitter"`
	Loss        float64       `json:"loss"`
	FailureRate float64       `json:"failure_rate"`
	Load        float64       `json:"load"`
	Stability   float64       `json:"stability"`
	Available   bool          `json:"available"`
	ObservedAt  time.Time     `json:"observed_at"`
	Samples     int           `json:"samples"`
}
type RouteScore struct {
	ID          string           `json:"id"`
	Score       float64          `json:"score"`
	Measurement RouteMeasurement `json:"measurement"`
}
type RouteHistory struct {
	Measurements []RouteMeasurement `json:"measurements"`
	Changes      []Change           `json:"changes"`
}
type Change struct {
	From, To string
	At       time.Time
}
type Parameters struct {
	MinimumImprovement float64       `json:"minimum_improvement"`
	MinimumLifetime    time.Duration `json:"minimum_lifetime"`
	ConfirmationPeriod time.Duration `json:"confirmation_period"`
	MaxAge             time.Duration `json:"max_age"`
}

func DefaultParameters() Parameters {
	return Parameters{5, 30 * time.Second, 10 * time.Second, 90 * time.Second}
}
func validFraction(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) && x >= 0 && x <= 1 }

// Score is in equivalent milliseconds. Loss/failure penalties dominate small RTT gains.
func Score(m RouteMeasurement) float64 {
	if !m.Available || m.Samples < 1 || m.RTT < 0 || m.Jitter < 0 || !validFraction(m.Loss) || !validFraction(m.FailureRate) || !validFraction(m.Load) || !validFraction(m.Stability) {
		return math.Inf(1)
	}
	return float64(m.RTT)/float64(time.Millisecond) + 2*float64(m.Jitter)/float64(time.Millisecond) + 1000*m.Loss + 2000*m.FailureRate + 15*m.Load + 20*(1-m.Stability)
}

type Selection struct {
	Primary   string       `json:"primary"`
	Secondary string       `json:"secondary"`
	Tertiary  string       `json:"tertiary"`
	Ranked    []RouteScore `json:"ranked"`
	Changed   bool         `json:"changed"`
}
type RouteSelector struct {
	mu                  sync.Mutex
	p                   Parameters
	measurements        map[string]RouteMeasurement
	history             RouteHistory
	current, pending    string
	since, pendingSince time.Time
}

func New(p Parameters) *RouteSelector {
	if p.MinimumImprovement <= 0 {
		p.MinimumImprovement = 5
	}
	if p.MinimumLifetime < 0 {
		p.MinimumLifetime = 0
	}
	if p.ConfirmationPeriod < 0 {
		p.ConfirmationPeriod = 0
	}
	if p.MaxAge <= 0 {
		p.MaxAge = 90 * time.Second
	}
	return &RouteSelector{p: p, measurements: make(map[string]RouteMeasurement), current: Direct}
}
func (s *RouteSelector) Observe(m RouteMeasurement) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if m.RouteID == "" || m.ObservedAt.IsZero() {
		return
	}
	old, ok := s.measurements[m.RouteID]
	if ok && m.ObservedAt.Before(old.ObservedAt) {
		return
	}
	if !ok && len(s.measurements) >= 64 {
		return
	}
	if ok && validFraction(old.Stability) && validFraction(m.Stability) {
		drift := math.Abs(float64(m.RTT-old.RTT)) / float64(50*time.Millisecond)
		m.Stability = math.Max(0, .75*old.Stability+.25*m.Stability-.1*math.Min(1, drift))
	}
	s.measurements[m.RouteID] = m
	s.history.Measurements = append(s.history.Measurements, m)
	if len(s.history.Measurements) > 256 {
		s.history.Measurements = s.history.Measurements[len(s.history.Measurements)-256:]
	}
}
func (s *RouteSelector) Select(now time.Time) Selection {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.since.IsZero() {
		s.since = now
	}
	ranked := make([]RouteScore, 0, len(s.measurements))
	for id, m := range s.measurements {
		if now.Sub(m.ObservedAt) > s.p.MaxAge || m.ObservedAt.After(now) {
			continue
		}
		v := Score(m)
		if !math.IsInf(v, 1) {
			ranked = append(ranked, RouteScore{id, v, m})
		}
	}
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].Score == ranked[j].Score {
			return ranked[i].ID < ranked[j].ID
		}
		return ranked[i].Score < ranked[j].Score
	})
	out := Selection{Primary: s.current, Ranked: ranked}
	currentScore := math.Inf(1)
	for _, r := range ranked {
		if r.ID == s.current {
			currentScore = r.Score
		}
	}
	best := Direct
	if len(ranked) > 0 {
		best = ranked[0].ID
	}
	// Unknown DIRECT observations do not establish a relay improvement.
	directMeasurement, known := s.measurements[Direct]
	directFresh := known && !directMeasurement.ObservedAt.After(now) && now.Sub(directMeasurement.ObservedAt) <= s.p.MaxAge
	if !directFresh && s.current == Direct {
		best = Direct
	}
	unhealthy := math.IsInf(currentScore, 1) && s.current != Direct
	if best != s.current && (unhealthy || currentScore-lookup(ranked, best) >= s.p.MinimumImprovement) {
		if unhealthy {
			out.Changed = true
		} else {
			if s.pending != best {
				s.pending = best
				s.pendingSince = now
			}
			if now.Sub(s.since) >= s.p.MinimumLifetime && now.Sub(s.pendingSince) >= s.p.ConfirmationPeriod {
				out.Changed = true
			}
		}
		if out.Changed {
			from := s.current
			s.current = best
			s.since = now
			s.pending = ""
			s.history.Changes = append(s.history.Changes, Change{from, best, now})
			if len(s.history.Changes) > 128 {
				s.history.Changes = s.history.Changes[1:]
			}
		}
	} else {
		s.pending = ""
	}
	out.Primary = s.current
	backups := []string{}
	for _, r := range ranked {
		if r.ID != out.Primary {
			backups = append(backups, r.ID)
		}
	}
	if len(backups) > 0 {
		out.Secondary = backups[0]
	}
	if len(backups) > 1 {
		out.Tertiary = backups[1]
	}
	return out
}
func lookup(r []RouteScore, id string) float64 {
	for _, v := range r {
		if v.ID == id {
			return v.Score
		}
	}
	return math.Inf(1)
}
func (s *RouteSelector) History() RouteHistory {
	s.mu.Lock()
	defer s.mu.Unlock()
	return RouteHistory{append([]RouteMeasurement(nil), s.history.Measurements...), append([]Change(nil), s.history.Changes...)}
}

func ValidateParameters(p Parameters) error {
	if math.IsNaN(p.MinimumImprovement) || math.IsInf(p.MinimumImprovement, 0) || p.MinimumImprovement <= 0 || p.MinimumImprovement > 10000 || p.MinimumLifetime < 0 || p.MinimumLifetime > time.Hour || p.ConfirmationPeriod < 0 || p.ConfirmationPeriod > time.Hour || p.MaxAge < time.Second || p.MaxAge > time.Hour {
		return errors.New("invalid routing parameters")
	}
	return nil
}
func (s *RouteSelector) Configure(p Parameters) error {
	if e := ValidateParameters(p); e != nil {
		return e
	}
	s.mu.Lock()
	s.p = p
	s.pending = ""
	s.mu.Unlock()
	return nil
}
