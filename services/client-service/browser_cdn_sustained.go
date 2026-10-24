package clientservice

import (
	"strings"
	"time"
)

const cdnStrongBytes = 1024 * 1024
const cdnStrongRecords = 16
const cdnStrongSpan = 3 * time.Second

type localFailureEvidence struct {
	At                                                    time.Time
	Domain, Strategy, Reason                              string
	TX, RX                                                uint64
	SinceClientHelloMS, SinceServerHelloMS, SinceLastRXMS int64
	ReceivedAfterHello, ApplicationRecords                uint64
	Promoted, PostHelloWrite                              bool
}

func (m *youtubeMeter) failureEvidenceLocked(reason string, now time.Time) localFailureEvidence {
	v := localFailureEvidence{At: now, Domain: m.key.Domain, Strategy: m.strategy, Reason: reason, TX: m.tx, RX: m.rx, SinceClientHelloMS: -1, SinceServerHelloMS: -1, SinceLastRXMS: -1, Promoted: m.promoted, PostHelloWrite: m.postHelloWrite}
	if !m.firstClientWrite.IsZero() {
		v.SinceClientHelloMS = now.Sub(m.firstClientWrite).Milliseconds()
	}
	if !m.serverHelloAt.IsZero() {
		v.SinceServerHelloMS = now.Sub(m.serverHelloAt).Milliseconds()
	}
	if !m.lastReply.IsZero() {
		v.SinceLastRXMS = now.Sub(m.lastReply).Milliseconds()
	}
	wire := m.wire.Snapshot().Origin
	v.ApplicationRecords = wire.ApplicationRecords
	if wire.ServerHelloEndBytes > 0 && m.rx >= wire.ServerHelloEndBytes {
		v.ReceivedAfterHello = m.rx - wire.ServerHelloEndBytes
	}
	return v
}
func (m *youtubeMeter) strongEvidenceLocked(now time.Time) bool {
	wire := m.wire.Snapshot().Origin
	return m.trial && !m.failed && !m.closing && m.healthy && m.applied && m.postHelloWrite &&
		wire.ServerHelloComplete && wire.AlertLevel == 0 && wire.State != "NON_TLS_OR_INVALID_HEADER" && wire.ServerHelloEndBytes > 0 && m.rx >= wire.ServerHelloEndBytes+cdnStrongBytes &&
		wire.ApplicationRecords >= cdnStrongRecords && wire.ApplicationBytes >= cdnStrongBytes &&
		!m.firstApplicationRX.IsZero() && m.lastReply.Sub(m.firstApplicationRX) >= cdnStrongSpan && now.Sub(m.lastReply) <= 3*time.Second
}
func (m *youtubeMeter) maybePromoteTrial(now time.Time) {
	m.mu.Lock()
	eligible := m.strongEvidenceLocked(now)
	m.mu.Unlock()
	if !eligible {
		return
	}
	// Same lock order as terminal finalization: once -> meter -> runtime.
	m.trialDone.Do(func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		if m.strongEvidenceLocked(now) && m.r.promoteRunning(m.key, m.strategy, m.rx, now) {
			m.promoted = true
			m.r.mu.Lock()
			if v, ok := m.r.records[m.id]; ok {
				v.TrialResult = "SUCCESS_RUNNING"
				v.Reason = "SUSTAINED_CDN_RUNNING_PROMOTED"
				m.r.records[m.id] = v
			}
			m.r.mu.Unlock()
			return
		}
		verdict := "INCONCLUSIVE"
		if m.failed {
			verdict = "FAILED"
		}
		m.r.finishTrialOutcome(m.key, m.strategy, verdict, m.rx > 0, now, m.rx)
	})
}
func (r *youtubeRuntime) promoteRunning(k youtubeKey, id string, received uint64, now time.Time) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	v := r.cache[k]
	if !r.trialActive || r.trialKey != k || r.trialStrategy != id || v.Qualified != id || !r.room(k, now) {
		return false
	}
	// This flow's current sustained proof supersedes its old public prerequisite;
	// it never grants another device/session authorization or another subgroup LOCAL.
	r.trialActive = false
	r.trialKey = youtubeKey{}
	r.trialStrategy = ""
	r.trialRecord = 0
	r.selection.ActiveTrial = false
	r.selection.TrialSuccesses++
	r.selection.StrongPromotions++
	r.score(k, id, true, now)
	v.Stage = "LOCAL"
	v.Strategy = id
	v.Confirmed = true
	v.Streak++
	v.VerifiedBytes = received
	v.Until = now.Add(youtubeLearnTTL)
	v.TrialAfter = now.Add(youtubeTrialCooldown)
	v.Retry = v.TrialAfter
	v.Reason = "SUSTAINED_CDN_RUNNING_PROMOTED"
	r.cache[k] = v
	return true
}
func (m *youtubeMeter) noteSmallEOF() bool {
	if m.stage != "LOCAL" {
		return false
	}
	inspection := m.currentInspection()
	if inspection != nil {
		t := inspection.TCPPackets
		if t.RemoteRST > 0 || t.ClientRST > 0 || t.ClientRetransmits > 0 || t.RemoteRetransmits > 0 {
			return false
		}
	}
	m.mu.Lock()
	wire := m.wire.Snapshot().Origin
	if m.failed || m.rx == 0 || m.rx >= 64*1024 || wire.AlertLevel != 0 || strings.Contains(wire.State, "INVALID") || wire.State == "NOT_OBSERVED" {
		m.mu.Unlock()
		return false
	}
	m.neutral = true
	m.smallEOF = true
	count := !m.neutralCounted
	m.neutralCounted = true
	m.mu.Unlock()
	m.r.mu.Lock()
	defer m.r.mu.Unlock()
	if count {
		s := m.r.stats[m.label]
		s.Inconclusive++
		m.r.stats[m.label] = s
	}
	if v, ok := m.r.records[m.id]; ok {
		v.Reason = "REMOTE_SMALL_EOF_INCONCLUSIVE"
		v.IOFailure = "REMOTE_SMALL_EOF"
		m.r.records[m.id] = v
	}
	return true
}
