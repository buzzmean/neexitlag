package clientservice

import "time"

const cdnPeerProvisionalReason = "CDN_PEER_PROMOTED_PROVISIONAL"

// A real promoted CDN flow (strong running, completed burst or three sustained
// trials) on one sn-subgroup lets OTHER googlevideo subgroups of the same
// network/profile/hint mode open new TCP flows LOCAL with that exact strategy,
// without waiting for their own max1 trial. It is provisional: the flow is a
// normal LOCAL flow (first-flight fallback to RELAY before origin bytes, no
// replay), >64KiB confirms only its own subgroup, and any real failure revokes
// the peer proof for the whole network/profile/strategy.
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
	if len(r.peerProofs) >= 64 {
		return
	}
	r.peerProofs[youtubeQualification{k.Network, k.Profile, strategy, k.HintOnly}] = now.Add(youtubeLearnTTL)
}

// Caller holds mu. Returns a strategy only for a CDN subgroup with no own
// LOCAL decision, no pending backoff and that strategy not failed in this sweep.
func (r *youtubeRuntime) peerStrategyLocked(k youtubeKey, v youtubeLearn, now time.Time) string {
	if !youtubeCDN(k) || k.Network == "" || len(r.peerProofs) == 0 || now.Before(v.Retry) && v.Failures > 0 || !r.room(k, now) {
		return ""
	}
	for i, id := range r.order {
		if v.Tried&(uint64(1)<<uint(i)) != 0 {
			continue
		}
		if now.Before(r.peerProofs[youtubeQualification{k.Network, k.Profile, id, k.HintOnly}]) {
			return id
		}
	}
	return ""
}

func (r *youtubeRuntime) revokePeerProofLocked(k youtubeKey, strategy string) {
	delete(r.peerProofs, youtubeQualification{k.Network, k.Profile, strategy, k.HintOnly})
}
