package windowsplatform

import (
	"net"
	"sync"
	"time"
)

type tcpEntry struct {
	tuple    Tuple
	identity ProcessIdentity
	decision Decision
	mode     string // opening, relay, direct, block, closed
	virtual  uint16
	last     time.Time
	upstream net.Conn
	local    net.Conn
}

// All methods/entry access require mu. Closed mappings are quarantined so late
// retransmissions cannot attach to a new connection. Active entries never evict.
type tcpState struct {
	mu      sync.Mutex
	entries map[Tuple]*tcpEntry
	ports   map[uint16]*tcpEntry
	limit   int
	next    uint16
}

func newTCPState(limit int) *tcpState {
	return &tcpState{entries: map[Tuple]*tcpEntry{}, ports: map[uint16]*tcpEntry{}, limit: limit, next: 20000}
}
func (s *tcpState) add(t Tuple, id ProcessIdentity, d Decision, mode string, now time.Time) *tcpEntry {
	if len(s.entries) >= s.limit {
		return nil
	}
	e := &tcpEntry{tuple: t, identity: id, decision: d, mode: mode, last: now}
	if mode == "opening" {
		for i := 0; i < 45536; i++ {
			p := s.next
			s.next++
			if s.next < 20000 {
				s.next = 20000
			}
			if s.ports[p] == nil {
				e.virtual = p
				s.ports[p] = e
				break
			}
		}
		if e.virtual == 0 {
			return nil
		}
	}
	s.entries[t] = e
	return e
}
func (s *tcpState) expire(now time.Time) {
	for t, e := range s.entries {
		age := now.Sub(e.last)
		timeout := 2 * time.Minute
		if e.mode == "closed" {
			timeout = 4 * time.Minute
		}
		if e.mode == "relay" {
			timeout = 2 * time.Hour
			if e.local == nil {
				timeout = 15 * time.Second
			}
		}
		if e.mode == "opening" {
			timeout = 20 * time.Second
		}
		if age <= timeout {
			continue
		}
		if e.upstream != nil {
			e.upstream.Close()
		}
		if e.local != nil {
			e.local.Close()
		}
		delete(s.entries, t)
		if e.virtual != 0 {
			delete(s.ports, e.virtual)
		}
	}
}
