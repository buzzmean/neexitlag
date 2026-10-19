package browserproxy

import (
	"sort"
	"sync"
	"time"

	"github.com/buzzmean/neexitlag/core/flowpolicy"
)

// FirstByteMetric measures encrypted transport progress, NOT HTTP/page success.
// One sample per CONNECT, including DNS/OPEN, admission and initial stall retry.
// NoReply is separate: failures never become zero-latency successes.
type FirstByteMetric struct {
	Network string  `json:"network"`
	Profile string  `json:"profile_id"`
	Group   string  `json:"group"`
	Path    string  `json:"path"`
	Mode    string  `json:"mode"`
	Replies uint64  `json:"replies"`
	NoReply uint64  `json:"closed_without_reply"`
	TotalMS float64 `json:"total_ms"`
	MinMS   float64 `json:"min_ms"`
	MaxMS   float64 `json:"max_ms"`
	LastMS  float64 `json:"last_ms"`
}
type firstByteKey struct{ network, profile, group, path, mode string }
type firstByteLedger struct {
	mu   sync.Mutex
	rows map[firstByteKey]FirstByteMetric
	lost uint64
}

func firstByteMode(p Plan) string {
	if p.DirectAttempts > 0 || p.Policy.Action == flowpolicy.Direct {
		return "DIRECT_FIRST"
	}
	switch p.Policy.Action {
	case flowpolicy.Relay, flowpolicy.RelayRequired:
		return "RELAY_FIRST"
	case flowpolicy.DirectPlus:
		return "DIRECT_PLUS"
	}
	return "UNROUTED"
}
func (l *firstByteLedger) observe(r Record, elapsed time.Duration) {
	group := flowpolicy.BrowserRelayGroup(r.Domain)
	if group == "" {
		group = r.Domain
	}
	key := firstByteKey{r.Plan.Network, r.Owner.Profile, group, r.EvidenceSource, r.FirstByteMode}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.rows == nil {
		l.rows = map[firstByteKey]FirstByteMetric{}
	}
	m, ok := l.rows[key]
	if !ok {
		if len(l.rows) >= 128 {
			l.lost++
			return
		}
		m = FirstByteMetric{Network: key.network, Profile: key.profile, Group: key.group, Path: key.path, Mode: key.mode, MinMS: -1, MaxMS: -1, LastMS: -1}
	}
	if elapsed < 0 {
		m.NoReply++
	} else {
		ms := float64(elapsed) / float64(time.Millisecond)
		if m.Replies == 0 || ms < m.MinMS {
			m.MinMS = ms
		}
		if ms > m.MaxMS {
			m.MaxMS = ms
		}
		m.Replies++
		m.TotalMS += ms
		m.LastMS = ms
	}
	l.rows[key] = m
}
func (l *firstByteLedger) snapshot() ([]FirstByteMetric, uint64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]FirstByteMetric, 0, len(l.rows))
	for _, m := range l.rows {
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.Network != b.Network {
			return a.Network < b.Network
		}
		if a.Profile != b.Profile {
			return a.Profile < b.Profile
		}
		if a.Group != b.Group {
			return a.Group < b.Group
		}
		if a.Path != b.Path {
			return a.Path < b.Path
		}
		return a.Mode < b.Mode
	})
	return out, l.lost
}
