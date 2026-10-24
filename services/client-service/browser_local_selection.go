package clientservice

import (
	"context"
	"net"
	"net/netip"
	"sort"
	"strings"
	"time"

	"github.com/buzzmean/neexitlag/core/browserproxy"
	"github.com/buzzmean/neexitlag/core/flowpolicy"
	"github.com/buzzmean/neexitlag/core/localstrategy"
)

const youtubeTrialCooldown = 3 * time.Minute
const youtubeTrialSuccesses = 3

type youtubeSelectionSnapshot struct {
	BackgroundAttempts, BackgroundFailures                  uint64 `json:",omitempty"`
	Trials, TrialSuccesses, TrialFailures, FailedAfterReply uint64
	ActiveTrial                                             bool   `json:"active_trial"`
	BackgroundActive                                        bool   `json:"background_active"`
	Contract                                                string `json:"contract"`
}
type youtubeQualification struct {
	Network, Profile, Strategy string
	HintOnly                   bool
}

type youtubeScoreKey struct {
	Scope    youtubeKey
	Strategy string
}
type youtubeScore struct {
	Good, Bad uint32
	Updated   time.Time
}

type youtubePending struct {
	Current profilesResult
	Request browserproxy.Request
	Plan    browserproxy.Plan
	Seen    time.Time
}

func (r *youtubeRuntime) register(k youtubeKey, cur profilesResult, q browserproxy.Request, p browserproxy.Plan, now time.Time) {
	if !q.Source.IsValid() || q.Source.Addr().IsUnspecified() || k.Network == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for key, v := range r.pending {
		if now.Sub(v.Seen) > 30*time.Minute {
			delete(r.pending, key)
		}
	}
	if _, ok := r.pending[k]; !ok && len(r.pending) >= 64 {
		return
	}
	r.pending[k] = youtubePending{cur, q, p, now}
}
func (r *youtubeRuntime) stop() { r.mu.Lock(); r.stopped = true; r.mu.Unlock(); r.workers.Wait() }
func (c *Controller) watchLocalSelection() {
	timer := time.NewTicker(15 * time.Second)
	defer timer.Stop()
	for {
		select {
		case <-c.ctx.Done():
			return
		case now := <-timer.C:
			c.recheckPendingLocal(now)
		}
	}
}
func (c *Controller) recheckPendingLocal(now time.Time) {
	r := c.browser.youtube
	r.mu.Lock()
	work := make(map[youtubeKey]youtubePending)
	for k, v := range r.pending {
		if now.Sub(v.Seen) > 30*time.Minute {
			delete(r.pending, k)
			continue
		}
		if !now.Before(r.cache[k].Retry) {
			work[k] = v
		}
	}
	r.mu.Unlock()
	for k, v := range work {
		// Re-resolve the owner and policy. A background recheck must never retain
		// permission after disable, process replacement, VPN or a network change.
		cur, err := c.browserProfile(v.Request.Owner)
		if err != nil || !cur.enabled || cur.network != k.Network || c.policyRevision(cur.profile) != c.policyRevision(v.Current.profile) {
			continue
		}
		c.youtubeBackgroundRecheck(cur, v.Request, v.Plan, k)
	}
}

// Caller holds r.mu. Bounded and network/subgroup scoped; no device identifiers.
func (r *youtubeRuntime) score(k youtubeKey, strategy string, good bool, now time.Time) {
	key := youtubeScoreKey{k, strategy}
	v, ok := r.scores[key]
	if !ok && len(r.scores) >= 2048 {
		for key, v := range r.scores {
			if now.Sub(v.Updated) > 24*time.Hour {
				delete(r.scores, key)
			}
		}
		if len(r.scores) >= 2048 {
			return
		}
	}
	if now.Sub(v.Updated) > 24*time.Hour {
		v = youtubeScore{}
	}
	if v.Good+v.Bad >= 100 {
		v.Good /= 2
		v.Bad /= 2
	}
	if good {
		v.Good++
	} else {
		v.Bad++
	}
	v.Updated = now
	r.scores[key] = v
}

// LOCAL results are deliberately NOT the family-wide DIRECT-timeout key.
// A single account endpoint cannot poison the video CDN or www/thumbnail origin.
func youtubeSubgroup(host string) string {
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	if host == "googlevideo.com" || strings.HasSuffix(host, ".googlevideo.com") {
		for _, label := range strings.Split(host, ".") {
			if i := strings.Index(label, "---sn-"); i >= 0 {
				return "googlevideo:" + label[i+3:]
			}
		}
		return "googlevideo:unclassified"
	}
	return host
}
func youtubeCDN(k youtubeKey) bool {
	return strings.HasPrefix(k.Domain, "googlevideo:") || k.Domain == "googlevideo.com"
}
func youtubeBackoff(failures uint32) time.Duration {
	if failures > 5 {
		failures = 5
	}
	pause := 15 * time.Second * time.Duration(uint64(1)<<failures)
	if pause > 5*time.Minute {
		pause = 5 * time.Minute
	}
	return pause
}
func (r *youtubeRuntime) room(k youtubeKey, now time.Time) bool {
	for key, v := range r.cache {
		if now.Sub(v.Until) > 24*time.Hour {
			delete(r.cache, key)
		}
	}
	_, ok := r.cache[k]
	return ok || len(r.cache) < 256
}
func (r *youtubeRuntime) setPlans(plans []localstrategy.CatalogPlan) {
	r.mu.Lock()
	defer r.mu.Unlock()
	old := strings.Join(r.order, "|")
	r.plans = map[string]localstrategy.CatalogPlan{}
	r.order = nil
	for _, p := range plans {
		if len(r.order) >= 48 {
			break
		}
		if _, ok := r.plans[p.ID]; !ok {
			r.plans[p.ID] = p
			r.order = append(r.order, p.ID)
		}
	}
	r.order = append(r.order, youtubeLegacyOrder...)
	if old != "" && old != strings.Join(r.order, "|") {
		for k, v := range r.cache {
			v.Next = 0
			v.Tried = 0
			r.cache[k] = v
		}
	}
}

// Compilation order is a prior; measured outcomes of this subgroup rank plans.
// Every failed candidate is tried once per sweep before any candidate is repeated.
func (r *youtubeRuntime) nextCandidate(k youtubeKey, v youtubeLearn) string {
	if len(r.order) == 0 {
		return ""
	}
	best := ""
	bestScore := -1000000
	for n := 0; n < len(r.order); n++ {
		i := (v.Next + n) % len(r.order)
		if v.Tried&(1<<uint(i)) == 0 {
			measured := r.scores[youtubeScoreKey{k, r.order[i]}]
			score := 0
			if time.Since(measured.Updated) < 24*time.Hour {
				score = int(measured.Good)*8 - int(measured.Bad)*2
			}
			if best == "" || score > bestScore {
				best = r.order[i]
				bestScore = score
			}
		}
	}
	if best != "" {
		return best
	}
	return r.order[v.Next%len(r.order)]
}
func (r *youtubeRuntime) takeRecheck(k youtubeKey, now time.Time) (youtubeLearn, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if k.Network == "" || r.stopped || r.selecting || !r.room(k, now) {
		return youtubeLearn{}, false
	}
	v := r.cache[k]
	if now.Before(v.Retry) {
		return youtubeLearn{}, false
	}
	if v.Stage == "" {
		v.Stage = "RELAY"
		v.Until = now.Add(youtubeLearnTTL)
	}
	if v.Tried == (uint64(1)<<uint(len(r.order)))-1 {
		v.Tried = 0
	}
	v.Retry = now.Add(youtubeTrialCooldown)
	r.cache[k] = v
	r.selecting = true
	r.workers.Add(1)
	r.selection.BackgroundActive = true
	return v, true
}
func (r *youtubeRuntime) endRecheck() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.selecting = false
	r.selection.BackgroundActive = false
	r.workers.Done()
}
func (r *youtubeRuntime) qualified(k youtubeKey, strategy string, now time.Time) bool {
	return now.Before(r.qualifications[youtubeQualification{k.Network, k.Profile, strategy, k.HintOnly}])
}
func (r *youtubeRuntime) takeTrial(k youtubeKey, now time.Time, workingRelay bool) (string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	v := r.cache[k]
	if !youtubeCDN(k) || !workingRelay || r.trialActive || now.Before(v.Retry) || v.Qualified == "" || !r.qualified(k, v.Qualified, now) {
		return "", false
	}
	r.trialActive = true
	r.selection.ActiveTrial = true
	r.selection.Trials++
	v.Retry = now.Add(youtubeTrialCooldown)
	r.cache[k] = v
	return v.Qualified, true
}
func (r *youtubeRuntime) finishTrial(k youtubeKey, strategy string, success, afterReply bool, now time.Time, received ...uint64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.trialActive = false
	r.selection.ActiveTrial = false
	if !r.room(k, now) {
		r.selection.TrialFailures++
		if afterReply {
			r.selection.FailedAfterReply++
		}
		return
	}
	v := r.cache[k]
	if success && v.Qualified == strategy && r.qualified(k, strategy, now) {
		r.score(k, strategy, true, now)
		r.selection.TrialSuccesses++
		if len(received) > 0 && received[0] > 64*1024 {
			v.VerifiedBytes = received[0]
		}
		v.Streak++
		if v.Streak >= youtubeTrialSuccesses {
			v.Stage = "LOCAL"
			v.Strategy = strategy
			v.Confirmed = true
			v.Reason = "THREE_SUSTAINED_CDN_TRIALS"
		}
	} else {
		r.selection.TrialFailures++
		if afterReply {
			r.selection.FailedAfterReply++
		}
		if v.Qualified != "" {
			v.Failures++
			for i, id := range r.order {
				if id == strategy {
					v.Next = (i + 1) % len(r.order)
					v.Tried |= 1 << uint(i)
				}
			}
		}
		v.Qualified = ""
		v.Streak = 0
		v.Stage = "RELAY"
		v.Strategy = ""
		v.Confirmed = false
		v.Reason = "CDN_TRIAL_FAILED_NEXT_CONNECTION_RELAY"
	}
	pause := youtubeTrialCooldown
	if b := youtubeBackoff(v.Failures); b > pause {
		pause = b
	}
	v.Retry = now.Add(pause)
	v.Until = now.Add(youtubeLearnTTL)
	r.cache[k] = v
}

// Only data-receiving, still-open media tunnels qualify. A preconnect or an
// unrelated profile, www page or an old completed download does not protect a trial.
func (u *browserRelayUsage) workingVideo(profile string, now time.Time) bool {
	u.mu.Lock()
	defer u.mu.Unlock()
	for _, v := range u.videos {
		if v.Profile == profile && v.Closed.IsZero() && v.Received > 64*1024 && !v.LastRead.IsZero() && now.Sub(v.LastRead) < 5*time.Second {
			return true
		}
	}
	return false
}
func youtubeTransferOK(p browserproxy.ProbeResult) bool {
	return p.Failure == "" && p.TLSVersion >= 0x0303 && p.Status >= 200 && p.Status < 300 && p.BodyBytes > 64*1024
}

func (c *Controller) youtubeBackgroundRecheck(cur profilesResult, q browserproxy.Request, p browserproxy.Plan, key youtubeKey) {
	if !q.Source.IsValid() || q.Source.Addr().IsUnspecified() || !c.youtubeLocalAllowed(cur.network) {
		return
	}
	// Do not let a recheck postpone an already-qualified CDN trial's cooldown.
	r := c.browser.youtube
	allowed, sharedCandidate := true, ""
	if c.sharedScopeAllowed(key) {
		allowed, sharedCandidate = r.shared.decision(key, time.Now())
	}
	if !allowed {
		return
	}
	r.mu.Lock()
	v := r.cache[key]
	waitingTrial := v.Qualified != "" && r.qualified(key, v.Qualified, time.Now())
	r.mu.Unlock()
	if waitingTrial {
		return
	}
	learned, ok := r.takeRecheck(key, time.Now())
	if !ok {
		return
	}
	select {
	case c.browser.probes <- struct{}{}:
	default:
		r.endRecheck()
		return
	}
	go func() {
		defer func() { <-c.browser.probes; r.endRecheck() }()
		ctx, cancel := context.WithTimeout(c.ctx, 12*time.Second)
		defer cancel()
		probeQ := q
		probeQ.Ingress = "YOUTUBE_BACKGROUND_RECHECK"
		r.mu.Lock()
		strategy := r.nextCandidate(key, learned)
		if sharedCandidate != "" {
			strategy = sharedCandidate
		}
		r.selection.BackgroundAttempts++
		r.mu.Unlock()
		if strategy == "" {
			return
		}
		// For a CDN no signed media URL is visible without MITM. These are necessary
		// prerequisites only, never a claim that the actual CDN has been validated.
		var verifiedBytes uint64
		passed := youtubePrerequisites(key, q.Host, func(origin [2]string) bool {
			// Separate held source reservations prevent the second origin from borrowing
			// the first origin's committed-transform evidence on a shared CDN address.
			socket, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IP(q.Source.Addr().AsSlice())})
			if err != nil {
				return false
			}
			defer socket.Close()
			probeQ.Source = socket.LocalAddr().(*net.UDPAddr).AddrPort()
			probeQ.Host = origin[0]
			work, stop := context.WithTimeout(ctx, 5*time.Second)
			relay := p
			relay.Policy = browserDNSPolicy(cur)
			answer, e := c.knownOriginResolve(work, cur, probeQ, relay)
			if e != nil {
				stop()
				c.recordLocalProbe(probeQ, p, strategy, browserproxy.ProbeResult{Phase: "ORIGIN_DNS", Failure: "DNS_FAILED", Detail: "RESOLVE_FAILED"}, false)
				r.backgroundFailure(key, strategy, "BACKGROUND_DNS_FAILED")
				return false
			}
			addresses := append([]netip.Addr(nil), answer.Addresses...)
			sort.SliceStable(addresses, func(i, j int) bool { return addresses[i].Is4() && !addresses[j].Is4() })
			var dst netip.AddrPort
			for _, ip := range addresses {
				if ip.Is4() && browserproxy.PublicDNSAddress(ip) {
					dst = netip.AddrPortFrom(ip, 443)
					break
				}
			}
			if !dst.IsValid() {
				stop()
				c.recordLocalProbe(probeQ, p, strategy, browserproxy.ProbeResult{Phase: "ORIGIN_DNS", Failure: "DNS_FAILED", Detail: "NO_PUBLIC_IPV4"}, false)
				r.backgroundFailure(key, strategy, "BACKGROUND_NO_PUBLIC_IPV4")
				return false
			}
			candidate := p
			candidate.Destination = dst
			candidate.DirectCandidates = nil
			candidate.DirectDeadline = time.Time{}
			id, _ := c.youtubeEngine(cur, probeQ, candidate, strategy)
			if id == "" {
				stop()
				c.recordLocalProbe(probeQ, candidate, strategy, browserproxy.ProbeResult{Phase: "LOCAL_GUARD", Failure: "LOCAL_GUARD", Detail: "NETWORK_OR_PROFILE_GUARD"}, false)
				r.backgroundFailure(key, strategy, "BACKGROUND_LOCAL_GUARD")
				return false
			}
			candidate.Policy = flowpolicy.Result{Action: flowpolicy.DirectPlus, Strategy: id, Fallback: flowpolicy.FailBlock, Source: youtubeLadderSource}
			var wire *localProbeConn
			result := browserproxy.ProbeOriginTransfer(work, probeQ.Host, origin[1], c.BrowserProbeRoots, func(ctx context.Context) (net.Conn, error) {
				conn, err := c.browserOpenUnobserved(ctx, probeQ, candidate)
				if err != nil {
					return nil, err
				}
				wire = &localProbeConn{Conn: conn}
				return wire, nil
			})
			applied := c.browserTransform(probeQ, candidate)
			sent := uint64(0)
			if wire != nil {
				sent = wire.sent.Load()
			}
			c.recordLocalProbe(probeQ, candidate, strategy, result, applied, sent)
			stop()
			if !youtubeTransferOK(result) || !applied {
				r.backgroundFailure(key, strategy, "BACKGROUND_NOT_SUSTAINED_OR_NO_TRANSFORM")
				return false
			}
			verifiedBytes = result.BodyBytes
			return true
		})
		if !passed {
			return
		}

		now := time.Now()
		if youtubeCDN(key) {
			r.completeCDNQualification(key, strategy, now)
		} else {
			r.confirmed(key, "LOCAL", strategy, now)
			r.recordVerification(key, verifiedBytes)
		}
	}()
}
func (r *youtubeRuntime) backgroundFailure(k youtubeKey, strategy, reason string) {
	r.mu.Lock()
	r.selection.BackgroundFailures++
	if reason == "BACKGROUND_DNS_FAILED" || reason == "BACKGROUND_NO_PUBLIC_IPV4" || reason == "BACKGROUND_LOCAL_GUARD" {
		v := r.cache[k]
		v.Reason = reason
		v.Retry = time.Now().Add(youtubeTrialCooldown)
		r.cache[k] = v
		r.mu.Unlock()
		return
	}
	r.mu.Unlock()
	r.failed(k, "LOCAL", strategy, reason, time.Now())
}

// Mixed account/page/CDN hints may select the stricter RELAY route, but may not
// authorize LOCAL for whichever name sorts first on a shared address.
func youtubeSameHintSubgroup(e browserDomainEvidence) bool {
	return e.YoutubeOnly && e.YoutubeSubgroup != ""
}

func youtubePrerequisites(k youtubeKey, host string, probe func([2]string) bool) bool {
	origins := [][2]string{{host, "/"}}
	if youtubeCDN(k) {
		origins = [][2]string{{"www.youtube.com", "/"}, {"i.ytimg.com", "/vi/aqz-KE-bpKQ/maxresdefault.jpg"}}
	}
	for _, origin := range origins {
		if !probe(origin) {
			return false
		}
	}
	return true
}

func (r *youtubeRuntime) recordVerification(k youtubeKey, bytes uint64) {
	r.mu.Lock()
	v, ok := r.cache[k]
	if !ok {
		r.mu.Unlock()
		return
	}
	v.VerifiedBytes = bytes
	r.cache[k] = v
	r.mu.Unlock()
	r.shared.notify()
}

// Public prerequisites do not extend a real CDN confirmation deadline.
func (r *youtubeRuntime) completeCDNQualification(key youtubeKey, strategy string, now time.Time) {

	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.qualifications) >= 256 {
		for k, until := range r.qualifications {
			if !now.Before(until) {
				delete(r.qualifications, k)
			}
		}
	}
	if len(r.qualifications) >= 256 {
		return
	}
	r.qualifications[youtubeQualification{key.Network, key.Profile, strategy, key.HintOnly}] = now.Add(youtubeLearnTTL)
	v := r.cache[key]
	stillConfirmed := v.Confirmed && v.Stage == "LOCAL" && v.Strategy == strategy && now.Before(v.Until)
	if !stillConfirmed {
		v.Streak = 0
		v.Confirmed = false
		v.Stage = "RELAY"
		v.Strategy = ""
		v.VerifiedBytes = 0
		v.Until = now.Add(youtubeLearnTTL)
	}
	v.Qualified = strategy
	v.Retry = now
	v.Reason = "TWO_ORIGINS_VERIFIED_CDN_TRIAL_PENDING"
	if stillConfirmed {
		v.Reason = "PUBLIC_RECHECK_PASSED; EXISTING_MEDIA_CONFIRMATION_DEADLINE_PRESERVED"
	}
	r.cache[key] = v
}
