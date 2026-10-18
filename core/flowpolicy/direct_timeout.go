package flowpolicy

import (
	"sort"
	"sync"
	"time"
)

// DirectTimeoutCache remembers origins whose DIRECT connection attempts hit a
// connect timeout (SYN without answer) so later flows go straight to the relay
// instead of paying the timeout again. It only ever moves traffic TOWARDS the relay:
// it never authorizes DIRECT, and a cached origin is re-checked by one background
// DIRECT probe per TTL so traffic returns to DIRECT when the path recovers.
//
// An origin is cached after two timeouts in a row within TimeoutWindow (a success in
// between resets the count). Only connect timeouts count: resets, TLS failures and
// HTTP errors are evidence about the site, not about reachability.
type DirectTimeoutCache struct {
	mu      sync.Mutex
	entries map[DirectTimeoutKey]*directTimeoutEntry
	// Cumulative counters, visible in diagnostics to judge the load on the relay.
	hits, created, revalidations, recovered  uint64
	stalls, stallRetries, stallRetryFailures uint64
}

type DirectTimeoutKey struct {
	Network     string
	Application string
	Host        string
	Port        uint16
}

const (
	TimeoutWindow     = 60 * time.Second
	DirectTimeoutTTL  = 10 * time.Minute
	directTimeoutCap  = 256
	probeRetryOnError = time.Minute
)

type directTimeoutEntry struct {
	firstFailure time.Time
	failures     int
	active       bool
	since        time.Time
	nextProbe    time.Time
	probing      bool
	hits         uint64
	lastHit      time.Time
}

func NewDirectTimeoutCache() *DirectTimeoutCache {
	return &DirectTimeoutCache{entries: make(map[DirectTimeoutKey]*directTimeoutEntry)}
}

// RecordTimeout notes one DIRECT connect timeout. It returns true when this call
// made (or renewed) the origin active.
func (c *DirectTimeoutCache) RecordTimeout(k DirectTimeoutKey, now time.Time) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	e := c.entries[k]
	if e == nil {
		c.makeRoom(now)
		e = &directTimeoutEntry{}
		c.entries[k] = e
	}
	if e.active {
		// A fresh timeout while cached confirms the path is still dead: renew.
		e.nextProbe = now.Add(DirectTimeoutTTL)
		return true
	}
	if e.failures == 0 || now.Sub(e.firstFailure) > TimeoutWindow {
		e.failures, e.firstFailure = 1, now
		return false
	}
	e.failures++
	e.active, e.since, e.nextProbe = true, now, now.Add(DirectTimeoutTTL)
	c.created++
	return true
}

// RecordStall notes a DIRECT tunnel that got no answer after the ClientHello. It is the
// same class of failure as a connect timeout and follows the same rules.
func (c *DirectTimeoutCache) RecordStall(k DirectTimeoutKey, now time.Time) bool {
	c.mu.Lock()
	c.stalls++
	c.mu.Unlock()
	return c.RecordTimeout(k, now)
}

// NoteRetry counts the outcome of a transparent retry through the relay.
func (c *DirectTimeoutCache) NoteRetry(ok bool) {
	c.mu.Lock()
	if ok {
		c.stallRetries++
	} else {
		c.stallRetryFailures++
	}
	c.mu.Unlock()
}

// RecordSuccess clears the count: DIRECT connected, so DIRECT works for this origin.
func (c *DirectTimeoutCache) RecordSuccess(k DirectTimeoutKey) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if e := c.entries[k]; e != nil {
		if e.active {
			c.recovered++
		}
		delete(c.entries, k)
	}
}

// Lookup reports whether flows to k must go straight to the relay. due is true
// when the TTL elapsed and the caller should start one background DIRECT probe
// (claimed atomically: only one caller gets due=true until ProbeDone).
func (c *DirectTimeoutCache) Lookup(k DirectTimeoutKey, now time.Time) (active, due bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e := c.entries[k]
	if e == nil || !e.active {
		return false, false
	}
	e.hits++
	e.lastHit = now
	c.hits++
	if !e.probing && !now.Before(e.nextProbe) {
		e.probing = true
		c.revalidations++
		return true, true
	}
	return true, false
}

// ProbeDone ends a background probe. Success/failure of the probe itself is fed
// through RecordSuccess/RecordTimeout; an inconclusive probe (the owner exited, a
// reset...) is retried soon instead of waiting a whole TTL.
func (c *DirectTimeoutCache) ProbeDone(k DirectTimeoutKey, now time.Time, conclusive bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e := c.entries[k]
	if e == nil {
		return
	}
	e.probing = false
	if !conclusive {
		e.nextProbe = now.Add(probeRetryOnError)
	}
}

func (c *DirectTimeoutCache) makeRoom(now time.Time) {
	for k, e := range c.entries {
		if !e.active && now.Sub(e.firstFailure) > TimeoutWindow {
			delete(c.entries, k)
		}
	}
	if len(c.entries) < directTimeoutCap {
		return
	}
	var oldest DirectTimeoutKey
	var at time.Time
	first := true
	for k, e := range c.entries {
		t := e.firstFailure
		if e.active {
			t = e.since
		}
		if first || t.Before(at) {
			oldest, at, first = k, t, false
		}
	}
	delete(c.entries, oldest)
}

type DirectTimeoutOrigin struct {
	Host      string    `json:"host"`
	Port      uint16    `json:"port"`
	Hits      uint64    `json:"relay_steered_flows"`
	Since     time.Time `json:"since"`
	NextProbe time.Time `json:"next_direct_probe"`
}

type DirectTimeoutStats struct {
	ActiveOrigins   int                   `json:"active_origins"`
	Hits            uint64                `json:"relay_steered_flows"`
	Created         uint64                `json:"origins_cached_total"`
	Revalidations   uint64                `json:"background_direct_probes"`
	RecoveredDirect uint64                `json:"recovered_to_direct"`
	Stalls          uint64                `json:"direct_stalls_total"`
	StallRetries    uint64                `json:"stall_retries_via_relay"`
	StallRetryFails uint64                `json:"stall_retries_failed"`
	Top             []DirectTimeoutOrigin `json:"top_origins"`
}

// Stats returns counters and the busiest origins (at most 20) for diagnostics.
func (c *DirectTimeoutCache) Stats() DirectTimeoutStats {
	c.mu.Lock()
	defer c.mu.Unlock()
	s := DirectTimeoutStats{Hits: c.hits, Created: c.created, Revalidations: c.revalidations, RecoveredDirect: c.recovered,
		Stalls: c.stalls, StallRetries: c.stallRetries, StallRetryFails: c.stallRetryFailures}
	for k, e := range c.entries {
		if !e.active {
			continue
		}
		s.ActiveOrigins++
		s.Top = append(s.Top, DirectTimeoutOrigin{Host: k.Host, Port: k.Port, Hits: e.hits, Since: e.since, NextProbe: e.nextProbe})
	}
	sort.Slice(s.Top, func(i, j int) bool {
		if s.Top[i].Hits != s.Top[j].Hits {
			return s.Top[i].Hits > s.Top[j].Hits
		}
		return s.Top[i].Host < s.Top[j].Host
	})
	if len(s.Top) > 20 {
		s.Top = s.Top[:20]
	}
	return s
}
