package clientservice

import (
	"context"
	"github.com/buzzmean/neexitlag/core/browserproxy"
	"github.com/buzzmean/neexitlag/core/flowpolicy"
	"net"
	"sort"
	"sync"
	"time"
)

var errVideoRelayCapacity = browserproxy.ErrVideoRelayAdmission

const videoRelayLimit = 8
const videoRelayWaitLimit = 32
const videoRelayWait = 2 * time.Second
const relayUsageCapacity = 1024
const relayUsageRetention = 10 * time.Minute

type relayMinuteKey struct {
	minute          time.Time
	profile, domain string
}
type relayMinute struct {
	Opened   uint64    `json:"tunnels_opened"`
	Minute   time.Time `json:"minute_utc"`
	Profile  string    `json:"profile_id"`
	Domain   string    `json:"domain"`
	Sent     uint64    `json:"sent_bytes"`
	Received uint64    `json:"received_bytes"`
}
type relayUsageSnapshot struct {
	OpaqueGoogleActive int `json:"opaque_google_active_tunnels"`
	SharedVideoActive  int `json:"googlevideo_shared_limit_active"`

	Active          int                 `json:"active_tcp_tunnels"`
	Opened          uint64              `json:"tcp_tunnels_opened"`
	ActiveOrigins   []relayActiveOrigin `json:"active_origins"`
	OverflowTunnels uint64              `json:"cardinality_overflow_tunnels"`
	Scope           string              `json:"scope"`
	VideoActive     int                 `json:"googlevideo_active_tunnels"`
	VideoLimit      int                 `json:"googlevideo_tunnel_limit"`
	VideoWaiting    int                 `json:"googlevideo_waiters"`
	VideoTimeouts   uint64              `json:"googlevideo_admission_failures"`
	VideoLimitHits  uint64              `json:"googlevideo_limit_events"`
	OverflowBytes   uint64              `json:"cardinality_overflow_bytes"`
	Truncated       bool                `json:"truncated"`
	Minutes         []relayMinute       `json:"minutes"`
}
type browserRelayUsage struct {
	activeOrigins           map[[2]string]int
	active                  int
	opened, overflowTunnels uint64
	mu                      sync.Mutex
	minutes                 map[relayMinuteKey]relayMinute
	opaqueGoogle            int
	video                   int
	waiting                 int
	admissionFailures       uint64
	changed                 chan struct{}
	limitHits, overflow     uint64
}

type relayActiveOrigin struct {
	Profile string `json:"profile_id"`
	Domain  string `json:"domain"`
	Active  int    `json:"active_tunnels"`
}

func (u *browserRelayUsage) openedTunnel(profile, domain string, at time.Time) func() {
	u.mu.Lock()
	if u.minutes == nil {
		u.minutes = map[relayMinuteKey]relayMinute{}
	}
	if u.activeOrigins == nil {
		u.activeOrigins = map[[2]string]int{}
	}
	key := [2]string{profile, domain}
	_, exists := u.activeOrigins[key]
	tracked := exists || len(u.activeOrigins) < relayUsageCapacity
	u.active++
	u.opened++
	if tracked {
		u.activeOrigins[key]++
	} else {
		u.overflowTunnels++
	}
	for k := range u.minutes {
		if at.Sub(k.minute) > relayUsageRetention {
			delete(u.minutes, k)
		}
	}
	mk := relayMinuteKey{at.UTC().Truncate(time.Minute), profile, domain}
	m, ok := u.minutes[mk]
	if ok || len(u.minutes) < relayUsageCapacity {
		m.Minute = mk.minute
		m.Profile = profile
		m.Domain = domain
		m.Opened++
		u.minutes[mk] = m
	} else {
		u.overflowTunnels++
	}
	u.mu.Unlock()
	var once sync.Once
	return func() {
		once.Do(func() {
			u.mu.Lock()
			defer u.mu.Unlock()
			u.active--
			if tracked {
				u.activeOrigins[key]--
				if u.activeOrigins[key] == 0 {
					delete(u.activeOrigins, key)
				}
			}
		})
	}
}

func (c *Controller) relayUsage() *browserRelayUsage {
	c.browser.mu.Lock()
	defer c.browser.mu.Unlock()
	if c.browser.usage == nil {
		c.browser.usage = &browserRelayUsage{minutes: map[relayMinuteKey]relayMinute{}}
	}
	return c.browser.usage
}
func (u *browserRelayUsage) acquireVideo() (func(), bool) { return u.tryAcquireVideo(true) }
func (u *browserRelayUsage) tryAcquireVideo(count bool, opaque ...bool) (func(), bool) {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.video >= videoRelayLimit {
		if count {
			u.limitHits++
		}
		return nil, false
	}
	u.video++
	isOpaque := len(opaque) > 0 && opaque[0]
	if isOpaque {
		u.opaqueGoogle++
	}
	var once sync.Once
	return func() {
		once.Do(func() {
			u.mu.Lock()
			u.video--
			if isOpaque {
				u.opaqueGoogle--
			}
			if u.changed != nil {
				close(u.changed)
				u.changed = nil
			}
			u.mu.Unlock()
		})
	}, true
}

// The caller's existing flow worker waits; no goroutine or unbounded queue is
// created. Eight active tunnels and at most 32 waiters are shared by both browsers.
func (u *browserRelayUsage) acquireVideoContext(ctx context.Context, opaque ...bool) (func(), bool) {
	if ctx.Err() != nil {
		return nil, false
	}
	if release, ok := u.tryAcquireVideo(true, opaque...); ok {
		return release, true
	}
	u.mu.Lock()
	if u.waiting >= videoRelayWaitLimit {
		u.admissionFailures++
		u.mu.Unlock()
		return nil, false
	}
	u.waiting++
	u.mu.Unlock()
	defer func() { u.mu.Lock(); u.waiting--; u.mu.Unlock() }()
	timer := time.NewTimer(videoRelayWait)
	defer timer.Stop()
	for {
		u.mu.Lock()
		if u.changed == nil {
			u.changed = make(chan struct{})
		}
		changed := u.changed
		u.mu.Unlock()
		if ctx.Err() == nil {
			if release, ok := u.tryAcquireVideo(false, opaque...); ok {
				return release, true
			}
		}
		select {
		case <-changed:
		case <-ctx.Done():
			u.mu.Lock()
			u.admissionFailures++
			u.mu.Unlock()
			return nil, false
		case <-timer.C:
			u.mu.Lock()
			u.admissionFailures++
			u.mu.Unlock()
			return nil, false
		}
	}
}

func (u *browserRelayUsage) account(profile, domain string, tx, rx uint64, at time.Time) {
	if tx+rx == 0 {
		return
	}
	u.mu.Lock()
	defer u.mu.Unlock()
	for k := range u.minutes {
		if at.Sub(k.minute) > relayUsageRetention {
			delete(u.minutes, k)
		}
	}
	key := relayMinuteKey{at.UTC().Truncate(time.Minute), profile, domain}
	v, ok := u.minutes[key]
	if !ok && len(u.minutes) >= relayUsageCapacity {
		u.overflow += tx + rx
		return
	}
	v.Minute, v.Profile, v.Domain = key.minute, profile, domain
	v.Sent += tx
	v.Received += rx
	u.minutes[key] = v
}
func (u *browserRelayUsage) snapshot() relayUsageSnapshot {
	u.mu.Lock()
	defer u.mu.Unlock()
	s := relayUsageSnapshot{VideoActive: u.video - u.opaqueGoogle, OpaqueGoogleActive: u.opaqueGoogle, SharedVideoActive: u.video, VideoLimit: videoRelayLimit, VideoLimitHits: u.limitHits, VideoWaiting: u.waiting, VideoTimeouts: u.admissionFailures, OverflowBytes: u.overflow, Minutes: []relayMinute{}}
	s.Active = u.active
	s.Opened = u.opened
	s.OverflowTunnels = u.overflowTunnels
	s.Scope = "browser TCP application payload; native + PAC; excludes QUIC overhead/origin probes"
	s.ActiveOrigins = []relayActiveOrigin{}
	for key, active := range u.activeOrigins {
		s.ActiveOrigins = append(s.ActiveOrigins, relayActiveOrigin{key[0], key[1], active})
	}
	sort.Slice(s.ActiveOrigins, func(i, j int) bool {
		a, b := s.ActiveOrigins[i], s.ActiveOrigins[j]
		if a.Active != b.Active {
			return a.Active > b.Active
		}
		if a.Profile != b.Profile {
			return a.Profile < b.Profile
		}
		return a.Domain < b.Domain
	})
	if len(s.ActiveOrigins) > 64 {
		s.ActiveOrigins = s.ActiveOrigins[:64]
		s.Truncated = true
	}
	for _, v := range u.minutes {
		if time.Since(v.Minute) <= relayUsageRetention {
			s.Minutes = append(s.Minutes, v)
		}
	}
	sort.Slice(s.Minutes, func(i, j int) bool {
		a, b := s.Minutes[i], s.Minutes[j]
		if !a.Minute.Equal(b.Minute) {
			return a.Minute.After(b.Minute)
		}
		if a.Profile != b.Profile {
			return a.Profile < b.Profile
		}
		return a.Domain < b.Domain
	})
	if len(s.Minutes) > 64 {
		s.Minutes = s.Minutes[:64]
		s.Truncated = true
	}
	s.Truncated = s.Truncated || u.overflow > 0 || u.overflowTunnels > 0
	return s
}

type relayMeterConn struct {
	net.Conn
	usage           *browserRelayUsage
	profile, domain string
	release         func()
	once            sync.Once
}

func (c *relayMeterConn) Read(p []byte) (int, error) {
	n, e := c.Conn.Read(p)
	if n > 0 {
		c.usage.account(c.profile, c.domain, 0, uint64(n), time.Now())
	}
	return n, e
}
func (c *relayMeterConn) Write(p []byte) (int, error) {
	n, e := c.Conn.Write(p)
	if n > 0 {
		c.usage.account(c.profile, c.domain, uint64(n), 0, time.Now())
	}
	return n, e
}
func (c *relayMeterConn) Close() error { e := c.Conn.Close(); c.once.Do(c.release); return e }
func (c *relayMeterConn) CloseWrite() error {
	if cw, ok := c.Conn.(interface{ CloseWrite() error }); ok {
		return cw.CloseWrite()
	}
	return nil
}
func isVideo(q browserproxy.Request) bool { return flowpolicy.DomainMatches(q.Host, "googlevideo.com") }

// Called while browser.mu is already held by status IPC.
func (c *Controller) relayUsageLocked() relayUsageSnapshot {
	if c.browser.usage == nil {
		return relayUsageSnapshot{VideoLimit: videoRelayLimit, Minutes: []relayMinute{}}
	}
	return c.browser.usage.snapshot()
}
