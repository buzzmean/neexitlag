package windowsplatform

import (
	"encoding/json"
	"github.com/buzzmean/neexitlag/core/telemetry"
	"sort"
	"sync"
	"time"
)

type CaptureCounters struct {
	IPv6Blocked   uint64 `json:"ipv6_blocked_socket_events"`
	CapturedBytes uint64 `json:"captured_bytes"`
	RelayedBytes  uint64 `json:"relayed_bytes"`
	DirectBytes   uint64 `json:"direct_bytes"`
	ReplyBytes    uint64 `json:"relay_reply_bytes"`
	Captured      uint64 `json:"captured_packets"`
	Relayed       uint64 `json:"relayed_packets"`
	Direct        uint64 `json:"direct_packets"`
	Dropped       uint64 `json:"dropped_packets"`
	Reinject      uint64 `json:"reinject_packets"`
	LoopPrevented uint64 `json:"loop_prevention_packets"`
	Replies       uint64 `json:"relay_reply_packets"`
}
type FlowEvidence struct {
	Protocol     string          `json:"protocol"`
	IPVersion    string          `json:"ip_version"`
	PID          uint32          `json:"pid"`
	Executable   string          `json:"exe"`
	CreationTime uint64          `json:"creation_time"`
	Source       string          `json:"source"`
	Destination  string          `json:"destination"`
	Route        string          `json:"selected_route"`
	RelayID      string          `json:"relay_id,omitempty"`
	Counters     CaptureCounters `json:"counters"`
	LastSeen     time.Time       `json:"last_seen"`
}
type CaptureEvidence struct {
	Counters   CaptureCounters             `json:"counters"`
	Flows      []FlowEvidence              `json:"flows"`
	Truncated  bool                        `json:"truncated"`
	LocalDrops map[string]telemetry.Metric `json:"local_drops,omitempty"`
}
type evidenceKey struct {
	PID     uint32
	Started uint64
	Tuple   Tuple
}
type EvidenceLedger struct {
	mu     sync.Mutex
	totals CaptureCounters
	flows  map[evidenceKey]FlowEvidence
}

func NewEvidenceLedger() *EvidenceLedger {
	return &EvidenceLedger{flows: map[evidenceKey]FlowEvidence{}}
}
func increment(c *CaptureCounters, event string) {
	switch event {
	case "ipv6_blocked":
		c.IPv6Blocked++
	case "captured":
		c.Captured++
	case "relay":
		c.Relayed++
	case "direct":
		c.Direct++
	case "drop":
		c.Dropped++
	case "reinject":
		c.Reinject++
	case "self":
		c.LoopPrevented++
	case "reply":
		c.Replies++
	}
}

// Records metadata and counters only; packet payload never enters the ledger.
func (l *EvidenceLedger) Record(id ProcessIdentity, t Tuple, event, relay string) {
	l.RecordBytes(id, t, event, relay, 0)
}
func (l *EvidenceLedger) RecordBytes(id ProcessIdentity, t Tuple, event, relay string, bytes uint64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	increment(&l.totals, event)
	incrementBytes(&l.totals, event, bytes)
	if id.PID == 0 || id.CreationTime == 0 || id.Path == "" {
		return
	}
	key := evidenceKey{id.PID, id.CreationTime, t}
	v, exists := l.flows[key]
	if !exists && len(l.flows) >= 1024 {
		return
	}
	v.Protocol = t.Protocol
	v.IPVersion = "IPv4"
	if t.Source.Addr().Is6() {
		v.IPVersion = "IPv6"
	}
	v.PID = id.PID
	v.Executable = id.Path
	v.CreationTime = id.CreationTime
	v.Source = t.Source.String()
	v.Destination = t.Destination.String()
	v.LastSeen = time.Now()
	increment(&v.Counters, event)
	incrementBytes(&v.Counters, event, bytes)
	if event == "ipv6_blocked" {
		v.Route = "BLOCK_UNSUPPORTED_IPV6"
		v.RelayID = ""
	} else if event == "relay" || event == "reply" {
		v.Route = "RELAY"
		v.RelayID = relay
	} else if event == "direct" {
		v.Route = "DIRECT"
		v.RelayID = ""
	}
	l.flows[key] = v
}
func (l *EvidenceLedger) Snapshot(pid uint32) CaptureEvidence {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := CaptureEvidence{Counters: l.totals, Flows: []FlowEvidence{}}
	for _, v := range l.flows {
		if pid == 0 || v.PID == pid {
			out.Flows = append(out.Flows, v)
		}
	}
	sort.Slice(out.Flows, func(i, j int) bool { return out.Flows[i].LastSeen.After(out.Flows[j].LastSeen) })
	if len(out.Flows) > 32 {
		out.Truncated = true
		out.Flows = out.Flows[:32]
	}
	for len(out.Flows) > 0 {
		raw, _ := json.Marshal(out)
		if len(raw) < 60000 {
			break
		}
		out.Truncated = true
		out.Flows = out.Flows[:len(out.Flows)-1]
	}
	return out
}
func (l *EvidenceLedger) Cleanup(now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for k, v := range l.flows {
		if now.Sub(v.LastSeen) > 5*time.Minute {
			delete(l.flows, k)
		}
	}
}

// Explicitly excludes service transport even if a profile selects the service exe.
func ClassifyCapture(id ProcessIdentity, t Tuple, self uint32, classify Classifier) (Decision, bool) {
	if id.PID == self {
		return Decision{}, true
	}
	if id.PID == 0 || id.CreationTime == 0 || id.Path == "" {
		return Decision{}, false
	}
	return classify(id, t), false
}

func incrementBytes(c *CaptureCounters, event string, n uint64) {
	switch event {
	case "captured":
		c.CapturedBytes += n
	case "relay":
		c.RelayedBytes += n
	case "direct":
		c.DirectBytes += n
	case "reply":
		c.ReplyBytes += n
	}
}

// AddBytes records stream payload only. TCP read boundaries are not packets.
func (l *EvidenceLedger) AddBytes(id ProcessIdentity, t Tuple, event, relay string, n uint64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	incrementBytes(&l.totals, event, n)
	key := evidenceKey{id.PID, id.CreationTime, t}
	if v, ok := l.flows[key]; ok {
		incrementBytes(&v.Counters, event, n)
		v.LastSeen = time.Now()
		l.flows[key] = v
	}
}
