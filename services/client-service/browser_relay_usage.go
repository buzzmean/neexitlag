package clientservice

import (
	"errors"
	"github.com/buzzmean/neexitlag/core/browserproxy"
	"github.com/buzzmean/neexitlag/core/flowpolicy"
	"net"
	"sort"
	"sync"
	"time"
)

var errVideoRelayCapacity = errors.New("GOOGLEVIDEO_RELAY_CAPACITY")

const videoRelayLimit = 4
const relayUsageCapacity = 1024
const relayUsageRetention = 10 * time.Minute

type relayMinuteKey struct {
	minute          time.Time
	profile, domain string
}
type relayMinute struct {
	Minute   time.Time `json:"minute_utc"`
	Profile  string    `json:"profile_id"`
	Domain   string    `json:"domain"`
	Sent     uint64    `json:"sent_bytes"`
	Received uint64    `json:"received_bytes"`
}
type relayUsageSnapshot struct {
	VideoActive    int           `json:"googlevideo_active_tunnels"`
	VideoLimit     int           `json:"googlevideo_tunnel_limit"`
	VideoLimitHits uint64        `json:"googlevideo_limit_events"`
	OverflowBytes  uint64        `json:"cardinality_overflow_bytes"`
	Truncated      bool          `json:"truncated"`
	Minutes        []relayMinute `json:"minutes"`
}
type browserRelayUsage struct {
	mu                  sync.Mutex
	minutes             map[relayMinuteKey]relayMinute
	video               int
	limitHits, overflow uint64
}

func (c *Controller) relayUsage() *browserRelayUsage {
	c.browser.mu.Lock()
	defer c.browser.mu.Unlock()
	if c.browser.usage == nil {
		c.browser.usage = &browserRelayUsage{minutes: map[relayMinuteKey]relayMinute{}}
	}
	return c.browser.usage
}
func (u *browserRelayUsage) acquireVideo() (func(), bool) {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.video >= videoRelayLimit {
		u.limitHits++
		return nil, false
	}
	u.video++
	var once sync.Once
	return func() { once.Do(func() { u.mu.Lock(); u.video--; u.mu.Unlock() }) }, true
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
	s := relayUsageSnapshot{VideoActive: u.video, VideoLimit: videoRelayLimit, VideoLimitHits: u.limitHits, OverflowBytes: u.overflow, Minutes: []relayMinute{}}
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
	s.Truncated = s.Truncated || u.overflow > 0
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
