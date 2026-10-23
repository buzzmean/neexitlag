package clientservice

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net"
	"net/netip"
	"sort"
	"sync"
	"time"

	"github.com/buzzmean/neexitlag/core/browserproxy"
	"github.com/buzzmean/neexitlag/core/flowpolicy"
	"github.com/buzzmean/neexitlag/core/localstrategy"
)

const youtubeLadderSource = "youtube_local_ladder"
const youtubeLearnTTL = 10 * time.Minute
const youtubeRetryAfter = time.Minute

var youtubeOrder = []string{"fake5x2-disorder", "fake8x2-disorder", "fake5", "fake8-disorder", "fake2", "disorder", "sni", "split1", "fake5-only"}

type youtubeKey struct {
	Network, Profile, Domain string
	Port                     uint16
}
type youtubeLearn struct {
	Stage, Strategy string
	Until, Retry    time.Time
	Next            int
	Tried           uint16
	Reason          string
	Confirmed       bool
}
type youtubeStage struct {
	Stage                                    string `json:"stage"`
	Strategy                                 string `json:"strategy,omitempty"`
	Connections, TX, RX, Failures, Confirmed uint64
}
type youtubeRecord struct {
	Active                 bool   `json:"active"`
	ID                     uint64 `json:"id"`
	Profile                string `json:"profile"`
	Domain                 string `json:"domain"`
	Destination            string `json:"destination"`
	Stage                  string `json:"stage"`
	Strategy               string `json:"strategy,omitempty"`
	Reason                 string `json:"reason"`
	TX, RX                 uint64
	ServerHello, Transform bool
	TransformKnown         bool
	Updated                time.Time `json:"updated_at"`
}
type youtubeCacheView struct {
	Network, Profile, Domain, Stage, Strategy, Reason string
	Port                                              uint16
	Until, Retry                                      time.Time
	Tried                                             uint16
}
type youtubeSnapshot struct {
	RecordsOmitted uint64             `json:"records_omitted"`
	CacheOmitted   int                `json:"cache_entries_omitted"`
	Cache          []youtubeCacheView `json:"cache"`
	CacheEntries   int                `json:"cache_entries"`
	Stages         []youtubeStage     `json:"stages"`
	Records        []youtubeRecord    `json:"records"`
	Replays        string             `json:"replay_contract"`
}
type youtubeEngine struct {
	Engine *localstrategy.LocalStrategyEngine
	Until  time.Time
}
type youtubeRuntime struct {
	mu      sync.Mutex
	cache   map[youtubeKey]youtubeLearn
	engines map[string]youtubeEngine
	stats   map[string]youtubeStage
	records map[uint64]youtubeRecord
	seq     uint64
}

func newYoutubeRuntime() *youtubeRuntime {
	return &youtubeRuntime{cache: map[youtubeKey]youtubeLearn{}, engines: map[string]youtubeEngine{}, stats: map[string]youtubeStage{}, records: map[uint64]youtubeRecord{}}
}
func youtubeLadderEligible(q browserproxy.Request, p browserproxy.Plan) bool {
	return !(p.Policy.Action == flowpolicy.RelayRequired && p.Policy.Source == "") && p.Enabled && !q.OriginAmbiguous && q.Secure && q.Port == 443 && youtubeRelayDomain(q.Host) && !protectedBrowserDomain(q.Host) && flowpolicy.KnownBrowserRelayGroup(q.Host) == "" && p.Policy.Source != "rule" && p.Policy.Action != flowpolicy.Block && p.Destination.Addr().Is4() && p.Destination.Addr().IsGlobalUnicast() && !p.Destination.Addr().IsPrivate()
}
func (c *Controller) youtubeLocalAllowed(network string) bool {
	n := c.network.Load()
	return network != "" && n != nil && n.Fingerprint == network && n.Error == "" && time.Since(n.ObservedAt) < 45*time.Second && len(n.VPN) == 0 && len(n.Conflicts) == 0 && !n.ConflictsObservedAt.IsZero() && time.Since(n.ConflictsObservedAt) < 3*time.Minute
}
func youtubeScope(cur profilesResult, q browserproxy.Request) youtubeKey {
	return youtubeKey{cur.network, cur.profile.ID, cacheHost(q.Host), q.Port}
}
func (r *youtubeRuntime) choose(k youtubeKey, now time.Time) (string, string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for key, v := range r.cache {
		if !now.Before(v.Until) {
			delete(r.cache, key)
		}
	}
	if k.Network != "" {
		if v, ok := r.cache[k]; ok {

			if v.Stage == "LOCAL" {
				return v.Stage, v.Strategy
			}
			return v.Stage, ""
		}
	}
	return "DIRECT", ""
}
func (r *youtubeRuntime) failed(k youtubeKey, stage, strategy, reason string, now time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if k.Network == "" {
		return
	}
	if len(r.cache) >= 256 {
		if _, ok := r.cache[k]; !ok {
			return
		}
	}
	v := r.cache[k]
	if stage == "LOCAL" {
		for i, id := range youtubeOrder {
			if id == strategy {
				v.Next = (i + 1) % len(youtubeOrder)
				v.Tried |= 1 << uint(i)
			}
		}
		v.Stage = "RELAY"
		v.Strategy = ""
		v.Retry = now.Add(youtubeRetryAfter)
	} else if stage == "DIRECT" {
		v.Stage = "LOCAL"
		v.Strategy = youtubeOrder[v.Next%len(youtubeOrder)]
	}
	v.Until = now.Add(youtubeLearnTTL)
	v.Reason = reason
	v.Confirmed = false
	if v.Tried == (1<<uint(len(youtubeOrder)))-1 {
		v.Retry = v.Until
	}
	r.cache[k] = v
}
func (r *youtubeRuntime) confirmed(k youtubeKey, stage, strategy string, now time.Time) {
	if k.Network == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.cache) >= 256 {
		if _, ok := r.cache[k]; !ok {
			return
		}
	}
	r.cache[k] = youtubeLearn{Stage: stage, Strategy: strategy, Until: now.Add(youtubeLearnTTL), Retry: now.Add(youtubeRetryAfter), Confirmed: true}
}
func (r *youtubeRuntime) snapshot() youtubeSnapshot {
	r.mu.Lock()
	defer r.mu.Unlock()
	s := youtubeSnapshot{RecordsOmitted: r.seq - uint64(len(r.records)), CacheEntries: len(r.cache), Cache: []youtubeCacheView{}, Stages: []youtubeStage{}, Records: []youtubeRecord{}, Replays: "complete ClientHello only before ANY origin byte; established TLS pinned; transport health is not HTTP/playback PASS"}
	for k, v := range r.cache {
		if len(s.Cache) < 24 {
			s.Cache = append(s.Cache, youtubeCacheView{k.Network, k.Profile, k.Domain, v.Stage, v.Strategy, v.Reason, k.Port, v.Until, v.Retry, v.Tried})
		}
	}
	for _, v := range r.stats {
		s.Stages = append(s.Stages, v)
	}
	sort.Slice(s.Stages, func(i, j int) bool {
		return s.Stages[i].Stage+s.Stages[i].Strategy < s.Stages[j].Stage+s.Stages[j].Strategy
	})
	for _, v := range r.records {
		s.Records = append(s.Records, v)
	}
	sort.Slice(s.Records, func(i, j int) bool {
		if s.Records[i].Active != s.Records[j].Active {
			return s.Records[i].Active
		}
		return s.Records[i].Updated.After(s.Records[j].Updated)
	})
	s.CacheOmitted = len(r.cache) - len(s.Cache)
	if len(s.Records) > 24 {
		s.RecordsOmitted += uint64(len(s.Records) - 24)
		s.Records = s.Records[:24]
	}
	return s
}
func (c *Controller) youtubeEngine(cur profilesResult, q browserproxy.Request, p browserproxy.Plan, candidate string) (string, *localstrategy.LocalStrategyEngine) {
	if !c.youtubeLocalAllowed(cur.network) {
		return "", nil
	}
	if candidate == "baseline" {
		return c.browserScopedLocal(cur.profile, q.Host, cur.network, q.Port)
	}
	var actions []localstrategy.Action
	for _, v := range localstrategy.BrowserYouTubeCandidates() {
		if v.ID == candidate {
			actions = append([]localstrategy.Action(nil), v.Actions...)
		}
	}

	if len(actions) == 0 {
		return "", nil
	}
	// A DNS association is not authenticated SNI. Bind exactly this destination,
	// never claim TrustedOrigin, and use absolute segmentation for hidden SNI.
	domains := []string{q.Host}
	var ips []string
	if q.HintOnly {
		domains = nil
		ips = []string{p.Destination.Addr().String() + "/32"}
		for i := range actions {
			if actions[i].Kind == "tcp-multisplit" {
				actions[i].Anchor = ""
				actions[i].Positions = []uint16{1, 5}
			}
		}
	}
	raw := cur.network + cur.profile.ID + q.Owner.Exe + q.Host + p.Destination.String() + candidate
	sum := sha256.Sum256([]byte(raw))
	id := "yt-" + hex.EncodeToString(sum[:12])
	r := c.browser.youtube
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	for key, v := range r.engines {
		if now.After(v.Until) {
			delete(r.engines, key)
		}
	}
	if v, ok := r.engines[id]; ok {
		return id, v.Engine
	}
	if len(r.engines) >= 256 {
		return "", nil
	}
	engine := localstrategy.New()
	profile := localstrategy.Profile{ID: id, Name: "YouTube local / " + candidate, Version: 1, Application: q.Owner.Exe, Protocol: "tcp", Ports: []uint16{443}, Domains: domains, IPs: ips, Strategy: "chain", Actions: actions, Preset: "youtube", PresetVersion: localstrategy.PresetVersion, Cutoff: 1, Enabled: true}
	if engine.Reload(localstrategy.Config{Version: 1, Enabled: true, User: []localstrategy.Profile{profile}}) != nil {
		return "", nil
	}
	engine.Guard = func(m localstrategy.Metadata) bool { return c.youtubeLocalAllowed(cur.network) }
	r.engines[id] = youtubeEngine{engine, now.Add(youtubeLearnTTL)}
	return id, engine
}
func (r *youtubeRuntime) engine(id string) *localstrategy.LocalStrategyEngine {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.engines[id].Engine
}

type youtubePlanRef struct {
	mu     sync.Mutex
	plan   browserproxy.Plan
	origin browserproxy.OriginReporter
}

func (r *youtubePlanRef) set(p browserproxy.Plan, conn net.Conn) {
	r.mu.Lock()
	r.plan = p
	r.origin, _ = conn.(browserproxy.OriginReporter)
	r.mu.Unlock()
}

type youtubePlanConn struct {
	net.Conn
	ref *youtubePlanRef
}

func (c *youtubePlanConn) CurrentPlan() browserproxy.Plan {
	c.ref.mu.Lock()
	p, origin := c.ref.plan, c.ref.origin
	c.ref.mu.Unlock()
	if origin != nil {
		report := origin.OriginAttempts()
		if addr, e := netip.ParseAddrPort(report.Selected); e == nil {
			p = browserproxy.RetargetOrigin(p, addr)
		}
	}
	return p
}
func (c *youtubePlanConn) OriginAttempts() browserproxy.OriginReport {
	c.ref.mu.Lock()
	p, origin := c.ref.plan, c.ref.origin
	c.ref.mu.Unlock()
	if origin != nil {
		return origin.OriginAttempts()
	}
	return browserproxy.OriginReport{Selected: p.Destination.String(), State: "YOUTUBE_LOCAL_LADDER_NO_RELAY_IP_RECOVERY"}
}

func (c *youtubePlanConn) CurrentTCPPolicy() (flowpolicy.Result, bool) {
	return c.CurrentPlan().Policy, true
}
func (c *youtubePlanConn) CloseWrite() error {
	if v, ok := c.Conn.(interface{ CloseWrite() error }); ok {
		return v.CloseWrite()
	}
	return nil
}

func (c *Controller) openYoutubeLadder(ctx context.Context, cur profilesResult, q browserproxy.Request, p browserproxy.Plan, open func(context.Context, browserproxy.Request, browserproxy.Plan) (net.Conn, error)) (net.Conn, browserproxy.Plan, error) {
	r := c.browser.youtube
	ref := &youtubePlanRef{}
	key := youtubeScope(cur, q)
	stage, strategy := r.choose(key, time.Now())
	if stage == "DIRECT" && p.Policy.Source == "configured_baseline" && p.Policy.Action == flowpolicy.DirectPlus {
		stage = "LOCAL"
		strategy = "baseline"
	}
	c.youtubeBackgroundRecheck(cur, q, p, key)
	// A prior DIRECT timeout is a hint to try LOCAL, never an instruction to skip it.
	if stage == "DIRECT" && cur.network != "" && c.directAddressFailed(cur, q, p) {
		stage = "LOCAL"
		strategy = youtubeOrder[0]
	}
	var step func(context.Context, string, string, int) (net.Conn, browserproxy.Plan, error)
	step = func(work context.Context, stage, strategy string, depth int) (net.Conn, browserproxy.Plan, error) {
		actual := p
		actual.FirstFlightManaged = true
		actual.DirectCandidates = nil
		actual.DirectDeadline = time.Time{}
		actual.Probe = "YOUTUBE_LOCAL_LADDER"
		actual.Policy = flowpolicy.Result{Action: flowpolicy.Direct, Fallback: flowpolicy.FailBlock, Source: youtubeLadderSource, Reason: "YouTube DIRECT → bounded LOCAL → authenticated RELAY; no replay after origin bytes"}
		if stage == "LOCAL" {
			id, _ := c.youtubeEngine(cur, q, actual, strategy)
			if id == "" {
				r.failed(key, "LOCAL", strategy, "LOCAL_GUARD_OR_CAPACITY", time.Now())
				r.mu.Lock()
				v, exists := r.cache[key]
				if exists {
					v.Retry = time.Now().Add(5 * time.Second)
					v.Tried = 0
					r.cache[key] = v
				}
				r.mu.Unlock()
				return step(work, "RELAY", "", depth+1)
			}
			actual.Policy.Action = flowpolicy.DirectPlus
			actual.Policy.Strategy = id
		}
		if stage == "RELAY" {
			actual = stallRetryPlan(cur, actual)
			actual.Policy.Source = youtubeLadderSource
			actual.FirstFlightManaged = true
			actual.Probe = "YOUTUBE_LOCAL_LADDER"

		}
		budget := 650 * time.Millisecond
		if stage == "RELAY" {
			budget = 5 * time.Second
		}
		attempt, cancel := context.WithTimeout(work, budget)
		var conn net.Conn
		var err error
		if stage == "RELAY" {
			conn, actual, err = c.openKnownBrowserGroup(attempt, cur, q, actual, open)
			actual.FirstFlightManaged = true
		} else {
			conn, err = open(attempt, q, actual)
		}
		cancel()
		r.mu.Lock()
		r.seq++
		record := r.seq
		if len(r.records) >= 96 {
			old := record
			for id := range r.records {
				if id < old {
					old = id
				}
			}
			delete(r.records, old)
		}
		r.records[record] = youtubeRecord{Active: err == nil, ID: record, Profile: cur.profile.ID, Domain: q.Host, Destination: actual.Destination.String(), Stage: stage, Strategy: strategy, Reason: "OPEN", Updated: time.Now()}
		label := stage + ":" + strategy
		stats := r.stats[label]
		stats.Stage = stage
		stats.Strategy = strategy
		stats.Connections++
		if err != nil {
			stats.Failures++
			v := r.records[record]
			v.Reason = "OPEN_FAILED"
			r.records[record] = v
		}
		r.stats[label] = stats
		r.mu.Unlock()
		next := func(nextCtx context.Context) (net.Conn, error) {
			if stage == "RELAY" || depth >= 2 {
				return nil, errors.New("YOUTUBE_LADDER_EXHAUSTED")
			}
			nextStage, nextStrategy := "RELAY", ""
			if stage == "DIRECT" {
				nextStage = "LOCAL"
				nextStrategy = youtubeOrder[0]
			}
			n, _, e := step(nextCtx, nextStage, nextStrategy, depth+1)
			return n, e
		}
		if err != nil {
			if work.Err() != nil {
				return nil, actual, work.Err()
			}
			r.failed(key, stage, strategy, "OPEN_FAILED", time.Now())
			if stage == "RELAY" {
				return nil, actual, err
			}
			conn, err = next(work)
			if err == nil {
				ref.mu.Lock()
				actual = ref.plan
				ref.mu.Unlock()
			}
			return conn, actual, err
		}
		ref.set(actual, conn)
		observer := &youtubeMeter{Conn: conn, r: r, id: record, label: label, key: key, stage: stage, strategy: strategy, started: time.Now(), transform: func() bool { return c.browserTransform(q, actual) }}
		if stage == "RELAY" {
			return observer, actual, nil
		}
		monitored := browserproxy.WatchMidstream(observer, 6*time.Second, func() { observer.failure("MIDSTREAM_STALL_NEW_CONNECTION_REQUIRED") })
		timeout := 400 * time.Millisecond
		if stage == "LOCAL" {
			timeout = 900 * time.Millisecond
		}
		conn = browserproxy.WatchFirstFlight(c.ctx, monitored, timeout, next, observer.failure)
		return conn, actual, nil
	}
	conn, actual, err := step(ctx, stage, strategy, 0)
	if err != nil {
		return nil, actual, err
	}
	return &youtubePlanConn{Conn: conn, ref: ref}, actual, nil
}

type youtubeMeter struct {
	net.Conn
	r                       *youtubeRuntime
	id                      uint64
	label                   string
	key                     youtubeKey
	stage, strategy         string
	started                 time.Time
	transform               func() bool
	mu                      sync.Mutex
	wire                    browserproxy.TLSWireObserver
	tx, rx                  uint64
	healthy                 bool
	applied, witnessChecked bool
	lastReply               time.Time
	gap                     sync.Once
	bad                     sync.Once
}

func (m *youtubeMeter) failure(reason string) {
	m.bad.Do(func() {
		m.mu.Lock()
		if m.stage == "LOCAL" && !m.witnessChecked && m.tx > 0 && m.transform != nil {
			m.witnessChecked = true
			m.applied = m.transform()
		}
		applied, known := m.applied, m.witnessChecked
		m.mu.Unlock()
		m.r.failed(m.key, m.stage, m.strategy, reason, time.Now())
		m.r.mu.Lock()
		defer m.r.mu.Unlock()
		s := m.r.stats[m.label]
		s.Failures++
		m.r.stats[m.label] = s
		if v, ok := m.r.records[m.id]; ok {
			v.Reason = reason
			v.Transform = applied
			v.TransformKnown = known
			v.Updated = time.Now()
			m.r.records[m.id] = v
		}
	})
}
func (m *youtubeMeter) Write(p []byte) (int, error) {
	n, e := m.Conn.Write(p)
	m.wire.ObserveClient(p[:n])
	m.account(uint64(n), 0)
	return n, e
}
func (m *youtubeMeter) Read(p []byte) (int, error) {
	n, e := m.Conn.Read(p)
	m.wire.ObserveOrigin(p[:n])
	m.account(0, uint64(n))
	if n == 0 && isYoutubeTimeout(e) {
		m.progressGap(time.Now())
	}
	v := m.wire.Snapshot().Origin
	if v.AlertLevel != 0 || v.State == "NON_TLS_OR_INVALID_HEADER" {
		m.failure("EARLY_TLS_ALERT_OR_INVALID_HEADER_NO_REPLAY")
	}
	if e != nil && !errors.Is(e, net.ErrClosed) {
		m.mu.Lock()
		spoken := m.tx > 0
		healthy := m.healthy
		m.mu.Unlock()
		if spoken && !healthy && (errors.Is(e, io.EOF) || !isYoutubeTimeout(e)) {
			m.failure("EARLY_EOF_OR_IO_NO_REPLAY")
		}
	}
	return n, e
}
func isYoutubeTimeout(e error) bool { var n net.Error; return errors.As(e, &n) && n.Timeout() }
func (m *youtubeMeter) account(tx, rx uint64) {
	m.mu.Lock()
	m.tx += tx
	m.rx += rx
	if rx > 0 {
		m.lastReply = time.Now()
	}
	hello := m.wire.Snapshot().Origin.ServerHelloComplete
	if m.stage == "LOCAL" && !m.witnessChecked && m.tx > 0 && rx > 0 {
		m.witnessChecked = true
		m.applied = m.transform()
	}
	transformed := m.stage != "LOCAL" || m.applied

	confirmed := !m.healthy && hello && m.rx >= 64*1024 && m.tx > 0 && time.Since(m.started) >= 250*time.Millisecond && transformed
	if confirmed {
		m.healthy = true
	}
	m.mu.Unlock()
	if confirmed && m.stage != "RELAY" {
		m.r.confirmed(m.key, m.stage, m.strategy, time.Now())
	}
	m.r.mu.Lock()
	defer m.r.mu.Unlock()
	s := m.r.stats[m.label]
	s.TX += tx
	s.RX += rx
	if confirmed {
		s.Confirmed++
	}
	m.r.stats[m.label] = s
	if v, ok := m.r.records[m.id]; ok {
		v.TX += tx
		v.RX += rx
		v.ServerHello = hello
		v.Transform = m.stage == "LOCAL" && transformed
		v.TransformKnown = m.stage != "LOCAL" || m.witnessChecked
		if confirmed {
			v.Reason = "SUSTAINED_TLS_TRANSPORT_NOT_PLAYBACK_PROOF"
		}
		v.Updated = time.Now()
		m.r.records[m.id] = v
	}
}
func (m *youtubeMeter) CloseWrite() error {
	if c, ok := m.Conn.(interface{ CloseWrite() error }); ok {
		return c.CloseWrite()
	}
	return nil
}

// A quiet encrypted response cannot be distinguished from a completed HTTP/2
// response without MITM. This is a short-lived hint for NEW connections only:
// do not close, move or replay the current TLS session on this observation.
func (m *youtubeMeter) progressGap(now time.Time) {
	if m.stage == "RELAY" {
		return
	}
	m.mu.Lock()
	gap := !m.healthy && m.tx > 0 && m.rx >= 8*1024 && m.rx < 64*1024 && !m.lastReply.IsZero() && now.Sub(m.lastReply) >= 4*time.Second
	m.mu.Unlock()
	if !gap {
		return
	}
	m.r.mu.Lock()
	positive := m.r.cache[m.key]
	m.r.mu.Unlock()
	if positive.Confirmed && positive.Stage == m.stage && positive.Strategy == m.strategy {
		return
	}
	m.gap.Do(func() {
		m.r.failed(m.key, m.stage, m.strategy, "EARLY_PROGRESS_GAP_HINT_NOT_HTTP_FAILURE", now)
		m.r.mu.Lock()
		if v, ok := m.r.records[m.id]; ok {
			v.Reason = "EARLY_PROGRESS_GAP_HINT_NEXT_CONNECTION_ONLY"
			m.r.records[m.id] = v
		}
		m.r.mu.Unlock()
	})
}

// takeRecheck is single-flight per scope/cooldown. Admission remains globally
// bounded by browser.probes. A busy foreground never waits for this probe.
func (r *youtubeRuntime) takeRecheck(k youtubeKey, now time.Time) (youtubeLearn, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	v, ok := r.cache[k]
	if !ok || now.Before(v.Retry) || !now.Before(v.Until) {
		return youtubeLearn{}, false
	}
	v.Retry = now.Add(youtubeRetryAfter)
	r.cache[k] = v
	return v, true
}
func (c *Controller) youtubeBackgroundRecheck(cur profilesResult, q browserproxy.Request, p browserproxy.Plan, key youtubeKey) {
	if !q.Source.IsValid() || q.Source.Addr().IsUnspecified() {
		return
	}
	learned, ok := c.browser.youtube.takeRecheck(key, time.Now())
	if !ok {
		return
	}
	select {
	case c.browser.probes <- struct{}{}:
	default:
		return
	}
	go func() {
		defer func() { <-c.browser.probes }()
		ctx, cancel := context.WithTimeout(c.ctx, 3*time.Second)
		defer cancel()
		if cur.network == "" || !c.youtubeLocalAllowed(cur.network) {
			return
		}
		// Reserve a separate source tuple: never collide with or borrow the live
		// application's Engine reservation/evidence while probing in the background.
		socket, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IP(q.Source.Addr().AsSlice())})
		if err != nil {
			return
		}
		defer socket.Close()
		probeQ := q
		probeQ.Source = socket.LocalAddr().(*net.UDPAddr).AddrPort()
		probeQ.Ingress = "YOUTUBE_BACKGROUND_RECHECK"
		for _, stage := range []string{"DIRECT", "LOCAL"} {
			if stage == "LOCAL" && learned.Stage == "LOCAL" {
				return
			}
			candidate := p
			candidate.Policy = flowpolicy.Result{Action: flowpolicy.Direct, Fallback: flowpolicy.FailBlock, Source: youtubeLadderSource}
			strategy := ""
			if stage == "LOCAL" {
				strategy = youtubeOrder[learned.Next%len(youtubeOrder)]
				id, _ := c.youtubeEngine(cur, probeQ, p, strategy)
				if id == "" {
					return
				}
				candidate.Policy.Action = flowpolicy.DirectPlus
				candidate.Policy.Strategy = id
			}
			budget := 800 * time.Millisecond
			if stage == "LOCAL" {
				budget = 1500 * time.Millisecond
			}
			trial, stop := context.WithTimeout(ctx, budget)
			result := browserproxy.ProbeOriginDetailed(trial, q.Host, q.Port, true, c.BrowserProbeRoots, func(work context.Context) (net.Conn, error) { return c.browserOpenUnobserved(work, probeQ, candidate) })
			stop()
			if result.Failure == "" && result.Status >= 200 && result.Status < 400 && result.Bytes > 0 && (stage != "LOCAL" || c.browserTransform(probeQ, candidate)) {
				c.browser.youtube.confirmed(key, stage, strategy, time.Now())
				return
			}
			if stage == "LOCAL" {
				c.browser.youtube.failed(key, stage, strategy, "BACKGROUND_TLS_HTTP_GET_FAILED", time.Now())
			}
		}
	}()
}

func (m *youtubeMeter) Close() error {
	err := m.Conn.Close()
	m.r.mu.Lock()
	if v, ok := m.r.records[m.id]; ok {
		v.Active = false
		v.Updated = time.Now()
		m.r.records[m.id] = v
	}
	m.r.mu.Unlock()
	return err
}
