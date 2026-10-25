package clientservice

import (
	"strings"
	"time"
)

const cdnPeerProvisionalReason = "CDN_PEER_PROMOTED_PROVISIONAL"

// A real promoted CDN flow (strong running, completed burst or three sustained
// trials) on one sn-subgroup lets OTHER googlevideo subgroups of the same
// network/profile/hint mode open new TCP flows LOCAL with that exact strategy,
// without waiting for their own max1 trial. It is provisional: the flow is a
// normal LOCAL flow (first-flight fallback to RELAY before origin bytes, no
// replay), >64KiB confirms only its own subgroup, and a real CDN flow failure
// revokes the peer proof for the whole network/profile/strategy.
func (r *youtubeRuntime) notePeerProofLocked(k youtubeKey, strategy string, now time.Time) {
	if !youtubeCDN(k) || k.Network == "" || strategy == "" {
		return
	}
	if r.peerProofs == nil {
		r.peerProofs = map[youtubeQualification]time.Time{}
	}
	for q, until := range r.peerProofs {
		if !now.Before(until) {
			delete(r.peerProofs, q)
		}
	}
	q := youtubeQualification{k.Network, k.Profile, strategy, k.HintOnly}
	if _, exists := r.peerProofs[q]; !exists && len(r.peerProofs) >= 64 {
		return
	}
	r.peerProofs[q] = now.Add(youtubeLearnTTL)
}

// Caller holds mu. Each confirmed (>64KiB, post-hello write, transform) real
// googlevideo LOCAL flow with a live peer strategy keeps that proof alive, so a
// long session does not fall back to RELAY 10 minutes after the first promotion.
// An expired proof is never revived here; it needs a new promotion.
func (r *youtubeRuntime) renewPeerProofLocked(k youtubeKey, strategy string, now time.Time) {
	if !youtubeCDN(k) || k.Network == "" {
		return
	}
	q := youtubeQualification{k.Network, k.Profile, strategy, k.HintOnly}
	if until, ok := r.peerProofs[q]; ok && now.Before(until) {
		r.peerProofs[q] = now.Add(youtubeLearnTTL)
		r.selection.PeerRenewals++
	}
}

// Caller holds mu. Returns a strategy only for a CDN subgroup without its own
// LOCAL decision whose real flows have not failed with that strategy recently.
// Background public-origin checks (www/i.ytimg GET) are not CDN flow evidence
// and neither block nor revoke a peer proof.
func (r *youtubeRuntime) peerStrategyLocked(k youtubeKey, now time.Time) string {
	if !youtubeCDN(k) || k.Network == "" || len(r.peerProofs) == 0 || !r.room(k, now) {
		return ""
	}
	for _, id := range r.order {
		if now.Before(r.peerProofs[youtubeQualification{k.Network, k.Profile, id, k.HintOnly}]) && !now.Before(r.peerDenied[youtubeScoreKey{k, id}]) {
			return id
		}
	}
	return ""
}

// Caller holds mu. Only a real CDN flow failure counts; BACKGROUND_* reasons and
// non-CDN hosts (www/i.ytimg/account) never revoke CDN peer evidence.
func (r *youtubeRuntime) peerFlowFailedLocked(k youtubeKey, strategy, reason string, now time.Time) {
	if !youtubeCDN(k) || strategy == "" || strings.HasPrefix(reason, "BACKGROUND_") {
		return
	}
	q := youtubeQualification{k.Network, k.Profile, strategy, k.HintOnly}
	if now.Before(r.peerProofs[q]) {
		r.selection.PeerRevocations++
	}
	delete(r.peerProofs, q)
	if r.peerDenied == nil {
		r.peerDenied = map[youtubeScoreKey]time.Time{}
	}
	for key, until := range r.peerDenied {
		if !now.Before(until) {
			delete(r.peerDenied, key)
		}
	}
	if len(r.peerDenied) < 256 {
		r.peerDenied[youtubeScoreKey{k, strategy}] = now.Add(youtubeLearnTTL)
	}
}
