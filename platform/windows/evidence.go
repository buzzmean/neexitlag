package windowsplatform

import (
	"encoding/json"
	"github.com/buzzmean/neexitlag/core/localstrategy"
	"github.com/buzzmean/neexitlag/core/telemetry"
	"sort"
	"sync"
	"time"
)

type CaptureCounters struct {
	FakeTLS         uint64 `json:"fake_tls_packets"`
	QuicFallback    uint64 `json:"quic_fallback_packets"`
	DirectPlus      uint64 `json:"direct_plus_packets"`
	DirectPlusBytes uint64 `json:"direct_plus_bytes"`
	IPv6Blocked     uint64 `json:"ipv6_blocked_packets"`
	CapturedBytes   uint64 `json:"captured_bytes"`
	RelayedBytes    uint64 `json:"relayed_bytes"`
	DirectBytes     uint64 `json:"direct_bytes"`
	ReplyBytes      uint64 `json:"relay_reply_bytes"`
	Captured        uint64 `json:"captured_packets"`
	Relayed         uint64 `json:"relayed_packets"`
	Direct          uint64 `json:"direct_packets"`
	Dropped         uint64 `json:"dropped_packets"`
	Reinject        uint64 `json:"reinject_packets"`
	LoopPrevented   uint64 `json:"loop_prevention_packets"`
	Replies         uint64 `json:"relay_reply_packets"`
}
type FlowEvidence struct {
	LocalInspection *localstrategy.Inspection `json:"local_inspection,omitempty"`

	StrategyRevision string          `json:"strategy_revision,omitempty"`
	Preset           string          `json:"preset,omitempty"`
	Chain            string          `json:"chain,omitempty"`
	FakeTTL          uint8           `json:"fake_ttl,omitempty"`
	Active           bool            `json:"active,omitempty"`
	IdentityCurrent  bool            `json:"identity_current,omitempty"`
	Strategy         string          `json:"strategy,omitempty"`
	Domain           string          `json:"domain,omitempty"`
	Protocol         string          `json:"protocol"`
	IPVersion        string          `json:"ip_version"`
	PID              uint32          `json:"pid"`
	Executable       string          `json:"exe"`
	CreationTime     uint64          `json:"creation_time"`
	Source           string          `json:"source"`
	Destination      string          `json:"destination"`
	Route            string          `json:"selected_route"`
	RelayID          string          `json:"relay_id,omitempty"`
	Counters         CaptureCounters `json:"counters"`
	LastSeen         time.Time       `json:"last_seen"`
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
	case "quic_fallback":
		c.QuicFallback++
	case "local":
		c.DirectPlus++
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

// BeginFlow retires evidence for a reused tuple, without resetting lifetime totals.
// Call once when the capture backend creates a new flow, not for retransmitted SYNs.
func (l *EvidenceLedger) BeginFlow(id ProcessIdentity, t Tuple) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.flows, evidenceKey{id.PID, id.CreationTime, t})
}
func (l *EvidenceLedger) RecordBytes(id ProcessIdentity, t Tuple, event, relay string, bytes uint64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.recordBytesLocked(id, t, event, relay, bytes)
}
func (l *EvidenceLedger) recordBytesLocked(id ProcessIdentity, t Tuple, event, relay string, bytes uint64) {
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
	} else if event == "direct" && v.Route != "DIRECT_PLUS" {
		// Native packets describe packet handling, not a new flow decision.
		// Only RecordLocal can establish successful DIRECT+ use.
		v.Route = "DIRECT"
		v.RelayID = ""
	}
	l.flows[key] = v
}
func (l *EvidenceLedger) Snapshot(pid uint32) CaptureEvidence { return l.snapshot(pid, nil) }
func (l *EvidenceLedger) snapshot(pid uint32, active map[evidenceKey]bool) CaptureEvidence {
	return l.snapshotTarget(pid, "", active)
}
func (l *EvidenceLedger) snapshotTarget(pid uint32, destination string, active map[evidenceKey]bool) CaptureEvidence {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := CaptureEvidence{Counters: l.totals, Flows: []FlowEvidence{}}
	for k, v := range l.flows {
		v.Active = active[k]
		if (pid == 0 || v.PID == pid) && (destination == "" || v.Destination == destination) {
			if v.LocalInspection != nil {
				copy := *v.LocalInspection
				v.LocalInspection = &copy
			}
			out.Flows = append(out.Flows, v)
		}
	}
	sort.Slice(out.Flows, func(i, j int) bool {
		if out.Flows[i].Active != out.Flows[j].Active {
			return out.Flows[i].Active
		}
		return out.Flows[i].LastSeen.After(out.Flows[j].LastSeen)
	})
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
	case "local":
		c.DirectPlusBytes += n
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

// RecordLocal counts only transformed original packets/bytes after ALL output
// packets have been sent successfully. Metadata and counters publish atomically.
// These are not a guessed VPS saving or a proof of application-level success.
func (l *EvidenceLedger) RecordLocal(id ProcessIdentity, t Tuple, strategy, domain string, n uint64) {
	l.RecordLocalResult(id, t, localstrategy.Result{Strategy: strategy, Domain: domain}, n)
}
func (l *EvidenceLedger) RecordLocalResult(id ProcessIdentity, t Tuple, r localstrategy.Result, n uint64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.recordBytesLocked(id, t, "local", "", n)
	key := evidenceKey{id.PID, id.CreationTime, t}
	v, ok := l.flows[key]
	if ok {
		v.Route = "DIRECT_PLUS"
		v.RelayID = ""
		v.Strategy = r.Strategy
		v.Domain = r.Domain
		v.Preset = r.Preset
		v.StrategyRevision = r.Revision
		v.Chain = r.Chain
		if r.FakePackets > 0 {
			v.FakeTTL = r.FakeTTL
			v.Counters.FakeTLS += r.FakePackets
			l.totals.FakeTLS += r.FakePackets
		}
		l.flows[key] = v
	}
}

func (l *EvidenceLedger) RecordQuicFallback(id ProcessIdentity, t Tuple, strategy, preset string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.recordBytesLocked(id, t, "quic_fallback", "", 0)
	k := evidenceKey{id.PID, id.CreationTime, t}
	if v, ok := l.flows[k]; ok {
		v.Route = "DIRECT_PLUS"
		v.RelayID = ""
		v.Strategy = strategy
		v.Preset = preset
		v.Chain = "HTTP/3/QUIC → TCP fallback"
		l.flows[k] = v
	}
}

// Passive domain attribution is metadata only and never changes the route.
func (l *EvidenceLedger) RecordDomain(id ProcessIdentity, t Tuple, domain string) {
	if domain == "" {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	k := evidenceKey{id.PID, id.CreationTime, t}
	if v, ok := l.flows[k]; ok && v.Domain == "" {
		v.Domain = domain
		l.flows[k] = v
	}
}

// Passive first-payload diagnostics never change route, chain or wire counters.
func (l *EvidenceLedger) RecordLocalInspection(id ProcessIdentity, t Tuple, d *localstrategy.Inspection) {
	if d == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	k := evidenceKey{id.PID, id.CreationTime, t}
	if v, ok := l.flows[k]; ok && (v.LocalInspection == nil || d.SegmentsSeen > 0 && (v.LocalInspection.Reason == "REASSEMBLY_STARTED" || v.LocalInspection.Reason == "REASSEMBLY_WAITING" || v.LocalInspection.Reason == "REASSEMBLY_OUT_OF_ORDER")) {
		copy := *d
		v.LocalInspection = &copy
		l.flows[k] = v
	}
}
