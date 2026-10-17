// Package clientservice owns routing, telemetry and privileged commands outside the UI.
package clientservice

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/buzzmean/neexitlag/core/flowpolicy"
	"github.com/buzzmean/neexitlag/core/localstrategy"
	"github.com/buzzmean/neexitlag/core/measurement"
	"github.com/buzzmean/neexitlag/core/optimizer"
	"github.com/buzzmean/neexitlag/core/profiles"
	"github.com/buzzmean/neexitlag/core/routing"
	"github.com/buzzmean/neexitlag/core/security"
	"github.com/buzzmean/neexitlag/core/telemetry"
	"github.com/buzzmean/neexitlag/core/transport"
	windowsplatform "github.com/buzzmean/neexitlag/platform/windows"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type RouteView struct {
	ID          string             `json:"id"`
	RTTMS       float64            `json:"rtt_ms"`
	JitterMS    float64            `json:"jitter_ms"`
	Loss        float64            `json:"loss"`
	FailureRate float64            `json:"failure_rate"`
	Score       *float64           `json:"score,omitempty"`
	Status      string             `json:"status"`
	ObservedAt  time.Time          `json:"observed_at"`
	Quality     *optimizer.Quality `json:"quality,omitempty"`
}
type Settings struct {
	Automatic     bool   `json:"automatic"`
	Failover      string `json:"failover"`
	UpdateChannel string `json:"update_channel"`
}
type Snapshot struct {
	HotAttach      *windowsplatform.HotAttachStatus `json:"hot_attach,omitempty"`
	RelayTransport *RelayTransportReadiness         `json:"relay_transport,omitempty"`
	Version        string                           `json:"version"`
	Connected      bool                             `json:"connected"`
	Optimizing     bool                             `json:"optimizing"`
	Demo           bool                             `json:"demo"`
	Backend        windowsplatform.CaptureStatus    `json:"backend"`
	Primary        string                           `json:"primary"`
	Secondary      string                           `json:"secondary"`
	Tertiary       string                           `json:"tertiary"`
	Routes         []RouteView                      `json:"routes"`
	Profiles       []profiles.Profile               `json:"profiles"`
	RXBytes        uint64                           `json:"rx_bytes"`
	TXBytes        uint64                           `json:"tx_bytes"`
	History        optimizer.RouteHistory           `json:"history"`
	Flows          []routing.Flow                   `json:"flows"`
	Settings       Settings                         `json:"settings"`
	ControlPlane   string                           `json:"control_plane"`
	ActiveApps     []string                         `json:"active_apps"`
}
type Controller struct {
	BrowserOwnerSID string // Existing installation owner; restrict the loopback endpoint to this user.
	// In-process fixture injection only; production uses the system resolver/roots.
	BrowserResolver             *net.Resolver
	BrowserProbeRoots           *x509.CertPool
	browser                     *browserRuntime
	activationMu                sync.Mutex
	relayPools                  map[string]*transport.Pool
	autoJobs                    map[string]autoLease
	groupBindings               map[uint32]groupBinding
	autoMu                      sync.Mutex
	AutoLearned                 *flowpolicy.Learning
	groupMu                     sync.Mutex
	groupAt                     time.Time
	groupAll                    []profiles.Identity
	discovery                   windowsplatform.Discovery
	Learned                     *localstrategy.Cache
	network                     atomic.Pointer[windowsplatform.NetworkContext]
	Local                       *localstrategy.LocalStrategyEngine
	localMu                     sync.Mutex
	Engine                      *routing.Engine
	Selector                    *optimizer.RouteSelector
	Measurements                *measurement.Engine
	Profiles                    *profiles.Registry
	Backend                     windowsplatform.WindowsCaptureBackend
	Demo                        bool
	ConfigPath, ControlPlaneURL string
	DiagnosticHTTPClient        *http.Client
	TCPDiagnosticTarget         string
	RX, TX                      atomic.Uint64
	mu                          sync.Mutex
	enabled                     bool
	settings                    Settings
	customPath                  string
	recentApps                  map[string]time.Time
	localProfiles               map[string]profiles.Profile
	acceptanceRelays            map[string]acceptanceRelaySource
	transportDrops              map[string]func() map[string]telemetry.Metric
	ctx                         context.Context
	cancel                      context.CancelFunc
	wg                          sync.WaitGroup
	FailRelay                   func(string) error
}

func New(ctx context.Context) *Controller {
	ctx, cancel := context.WithCancel(ctx)
	c := &Controller{browser: newBrowserRuntime(), relayPools: map[string]*transport.Pool{}, AutoLearned: flowpolicy.NewLearning(), groupBindings: map[uint32]groupBinding{}, autoJobs: map[string]autoLease{}, discovery: windowsplatform.NewDiscovery(), Local: localstrategy.New(), Learned: localstrategy.NewCache(1024), Engine: routing.New(4096), Selector: optimizer.New(optimizer.DefaultParameters()), Measurements: measurement.New(measurement.Defaults()), Profiles: profiles.New(profiles.Defaults()), recentApps: make(map[string]time.Time), localProfiles: make(map[string]profiles.Profile), acceptanceRelays: make(map[string]acceptanceRelaySource), transportDrops: make(map[string]func() map[string]telemetry.Metric), settings: Settings{Automatic: true, Failover: "DIRECT", UpdateChannel: "stable"}, ctx: ctx, cancel: cancel}
	c.Local.Guard = func(m localstrategy.Metadata) bool {
		if c.autoCandidateQuarantined(m) {
			return false
		}
		k, err := c.localKey(m.Application, m.Destination.String(), m.Protocol)
		return err != nil || !c.Learned.Quarantined(k, m.StrategyID, time.Now())
	}
	c.Measurements.OnFailure = c.Selector.ReportProbeFailure
	c.Engine.SetAutoSelector(func(f routing.Flow) routing.Policy {
		c.mu.Lock()
		automatic := c.settings.Automatic
		c.mu.Unlock()
		if !automatic {
			return routing.Policy{}
		}
		selection := c.Selector.Select(time.Now())
		route := func(id string) routing.Route {
			if id == optimizer.Direct {
				return routing.Route{}
			}
			return routing.Route{RelayID: id}
		}
		p := routing.Policy{Route: route(selection.Primary)}
		for _, id := range []string{selection.Secondary, selection.Tertiary} {
			if id != "" {
				p.Backups = append(p.Backups, route(id))
			}
		}
		// Same per-target policy at classification and actual transport open. The
		// engine invokes this under its own lock: no call back into Engine methods.
		d := windowsplatform.Decision{Route: p.Route}
		c.applyLocalAuto(windowsplatform.ProcessIdentity{Path: f.ProcessPath}, windowsplatform.Tuple{Protocol: f.Protocol, Source: f.Source, Destination: f.Destination}, &d)
		p.Route = d.Route
		return p
	})
	c.syncPolicies()
	return c
}
func (c *Controller) syncPolicies() {
	for _, p := range c.Profiles.List() {
		policy := routing.Policy{}
		switch p.Mode {
		case profiles.Auto:
			policy.Auto = true
		case profiles.DirectPlus:
			policy.Route = routing.Route{Kind: routing.DirectPlusRoute}
		case profiles.Manual:
			policy.Route = routing.Route{RelayID: p.RelayID}
		case profiles.Block:
			policy.Block = true
		}
		c.mu.Lock()
		policy.BlockOnFailure = c.settings.Failover == "BLOCK"
		c.mu.Unlock()
		if p.Fallback != "" {
			policy.BlockOnFailure = p.Fallback == flowpolicy.Required || p.Fallback == flowpolicy.FailBlock
		}
		if p.Mode == profiles.Auto && (presetFor(p) == "openai" || presetFor(p) == "openai-web") {
			policy.Auto = false
			policy.Route = routing.Route{RelayID: p.RelayID}
			if policy.Route.RelayID == "" {
				policy.Route.RelayID = "VPS-01"
			}
			policy.BlockOnFailure = true
		}
		if len(p.Rules) > 0 {
			policy.Auto = false
			policy.Block = true
		} // classified flows supply the exact pinned override
		c.Engine.SetPolicy(p.ID, policy)
	}
}
func (c *Controller) Start() {
	c.wg.Add(1)
	go func() { defer c.wg.Done(); c.watchNetwork() }()
	c.wg.Add(1)
	go func() { defer c.wg.Done(); c.Measurements.Run(c.ctx, c.Selector.Observe) }()
}
func (c *Controller) Close() {
	c.cancel()
	c.browser.mu.Lock()
	bs := c.browser.server
	c.browser.mu.Unlock()
	if bs != nil {
		bs.Close()
	}
	if c.Backend != nil {
		c.Backend.Stop()
	}
	c.wg.Wait()
}
func (c *Controller) SetCustomPath(path string) error {
	c.customPath = path
	if err := c.loadAutoLearning(); err != nil {
		return err
	}
	if err := c.loadLocal(); err != nil {
		return err
	}
	settingsRaw, e := os.ReadFile(filepath.Join(filepath.Dir(path), "settings.json"))
	if e == nil {
		var stored Settings
		if e = security.Parse(settingsRaw, &stored); e != nil {
			return e
		}
		if (stored.Failover != "DIRECT" && stored.Failover != "BLOCK") || (stored.UpdateChannel != "stable" && stored.UpdateChannel != "preview") {
			return errors.New("invalid persisted settings")
		}
		c.mu.Lock()
		c.settings = stored
		c.mu.Unlock()
	} else if !os.IsNotExist(e) {
		return e
	}
	b, e := os.ReadFile(path)
	if os.IsNotExist(e) {
		return nil
	}
	if e != nil {
		return e
	}
	var all []profiles.Profile
	if e = security.Parse(b, &all); e != nil {
		return e
	}
	for _, p := range all {
		prepareAuto(&p)
		c.mu.Lock()
		c.localProfiles[p.ID] = p
		c.mu.Unlock()
		if e = c.Profiles.Upsert(p); e != nil {
			return e
		}
	}
	c.syncPolicies()
	return nil
}
func (c *Controller) saveProfiles() error {
	if c.customPath == "" {
		return nil
	}
	all := []profiles.Profile{}
	c.mu.Lock()
	for _, p := range c.localProfiles {
		if p.Trial != nil {
			continue
		}
		all = append(all, p)
	}
	c.mu.Unlock()
	raw, e := json.Marshal(all)
	if e != nil {
		return e
	}
	if e = os.MkdirAll(filepath.Dir(c.customPath), 0700); e != nil {
		return e
	}
	f, e := os.CreateTemp(filepath.Dir(c.customPath), ".profiles-*")
	if e != nil {
		return e
	}
	name := f.Name()
	defer os.Remove(name)
	f.Chmod(0600)
	_, e = f.Write(raw)
	f.Close()
	if e != nil {
		return e
	}
	return os.Rename(name, c.customPath)
}
func (c *Controller) Classify(identity windowsplatform.ProcessIdentity, t windowsplatform.Tuple) windowsplatform.Decision {
	c.mu.Lock()
	enabled := c.enabled
	c.mu.Unlock()
	decision := windowsplatform.Decision{Route: routing.Route{}}
	if !enabled {
		return decision
	}
	profile, ok := c.matchApplication(identity)
	if !ok {
		return decision
	}
	f := routing.Flow{Protocol: t.Protocol, Source: t.Source, Destination: t.Destination, PID: identity.PID, ProcessPath: identity.Path, ApplicationID: profile.ID}
	decision.Flow = f
	decision.Revision = c.policyRevision(profile)
	c.mu.Lock()
	decision.BlockOnFailure = c.settings.Failover == "BLOCK"
	c.mu.Unlock()
	switch profile.Mode {
	case profiles.Block:
		decision.Block = true
	case profiles.DirectPlus:
		decision.Route = routing.Route{Kind: routing.DirectPlusRoute}
		decision.Local = c.Local
		decision.LocalStrategyID = profile.LocalStrategyID
		decision.BlockOnFailure = false
	case profiles.Manual:
		decision.Route = routing.Route{RelayID: profile.RelayID}
	case profiles.Auto:
		c.mu.Lock()
		automatic := c.settings.Automatic
		c.mu.Unlock()
		if !automatic {
			return decision
		}
		primary := c.Selector.Select(time.Now()).Primary
		if primary != optimizer.Direct {
			decision.Route = routing.Route{RelayID: primary}
		}
	}
	if profile.Mode == profiles.Auto && c.Local.HasApplication(identity.Path) {
		decision.Route = routing.Route{}
	}
	if profile.Mode == profiles.Auto {
		c.applyLocalAuto(identity, t, &decision)
	}
	decision = c.scopedDecision(profile, identity, t, decision)
	if n := c.network.Load(); n != nil && len(n.Conflicts) > 0 {
		if decision.BlockOnFailure {
			decision.Block = true
			decision.Deferred = false
			decision.ResolveDomain = nil
			return decision
		}
		return windowsplatform.Decision{Flow: f}
	}
	return decision
}
func (c *Controller) Snapshot() Snapshot {
	selection := c.Selector.Select(time.Now())
	history := c.Selector.History()
	if len(history.Measurements) > 48 {
		history.Measurements = history.Measurements[len(history.Measurements)-48:]
	}
	if len(history.Changes) > 32 {
		history.Changes = history.Changes[len(history.Changes)-32:]
	}
	latest := map[string]optimizer.RouteMeasurement{}
	for _, m := range history.Measurements {
		latest[m.RouteID] = m
	}
	qualities := c.Selector.Qualities(time.Now())
	// Route state must not disappear when display history is truncated, or while
	// a newly failed route has not yet completed its first probe round.
	for id, q := range qualities {
		if _, ok := latest[id]; !ok {
			latest[id] = optimizer.RouteMeasurement{RouteID: id, ObservedAt: q.ObservedAt}
		}
	}
	views := []RouteView{}
	for _, m := range latest {
		view := RouteView{ID: m.RouteID, RTTMS: float64(m.RTT) / float64(time.Millisecond), JitterMS: float64(m.Jitter) / float64(time.Millisecond), Loss: m.Loss, FailureRate: m.FailureRate, Status: "unavailable", ObservedAt: m.ObservedAt}
		if q, ok := qualities[m.RouteID]; ok {
			view.Quality = &q
			view.RTTMS = float64(q.LatestRTT) / float64(time.Millisecond)
			view.Loss = q.Loss
			view.FailureRate = q.FailureRate
			view.JitterMS = float64(q.Jitter) / float64(time.Millisecond)
			view.Status = q.Health
		}
		for _, r := range selection.Ranked {
			if r.ID == m.RouteID {
				score := r.Score
				view.Score = &score
				view.Status = "healthy"
			}
		}
		views = append(views, view)
	}
	sort.Slice(views, func(i, j int) bool { return views[i].ID < views[j].ID })
	backend := windowsplatform.CaptureStatus{Backend: "none", Error: "capture not configured; networking remains DIRECT"}
	if c.Backend != nil {
		backend = c.Backend.Status()
	}
	c.mu.Lock()
	enabled := c.enabled
	settings := c.settings
	recent := make(map[string]time.Time, len(c.recentApps))
	for id, t := range c.recentApps {
		recent[id] = t
	}
	c.mu.Unlock()
	active := map[string]bool{}
	for id, t := range recent {
		if time.Since(t) < 5*time.Second {
			active[id] = true
		}
	}
	flows := c.Engine.Snapshot()
	for _, f := range flows {
		if !f.Route.Direct() {
			active[f.ApplicationID] = true
		}
	}
	apps := []string{}
	for id := range active {
		apps = append(apps, id)
	}
	sort.Strings(apps)
	if len(history.Measurements) > 32 {
		history.Measurements = history.Measurements[len(history.Measurements)-32:]
	}
	if len(history.Changes) > 32 {
		history.Changes = history.Changes[len(history.Changes)-32:]
	}
	if len(flows) > 16 {
		flows = flows[:16]
	}
	rx, tx := c.RX.Load(), c.TX.Load()
	if b, ok := c.Backend.(interface{ Traffic() (uint64, uint64) }); ok {
		brx, btx := b.Traffic()
		rx += brx
		tx += btx
	}
	snapshot := Snapshot{HotAttach: c.hotAttachStatus(), Version: "0.3.0-dev", Connected: c.Demo || backend.Ready, Optimizing: enabled && (c.Demo || backend.Ready), Demo: c.Demo, Backend: backend, Primary: selection.Primary, Secondary: selection.Secondary, Tertiary: selection.Tertiary, Routes: views, Profiles: c.Profiles.List(), RXBytes: rx, TXBytes: tx, History: history, Flows: flows, Settings: settings, ControlPlane: c.ControlPlaneURL, ActiveApps: apps}
	// Keep mandatory profile/quality fields; shorten display history to the IPC frame.
	for {
		raw, _ := json.Marshal(snapshot)
		if len(raw) <= 60000 {
			break
		}
		if len(snapshot.History.Measurements) > 0 {
			snapshot.History.Measurements = snapshot.History.Measurements[1:]
			continue
		}
		if len(snapshot.History.Changes) > 0 {
			snapshot.History.Changes = snapshot.History.Changes[1:]
			continue
		}
		if len(snapshot.Flows) > 0 {
			snapshot.Flows = snapshot.Flows[:len(snapshot.Flows)-1]
			continue
		}
		break
	}
	return snapshot
}
func (c *Controller) Handle(ctx context.Context, method string, raw json.RawMessage) (any, error) {
	switch method {
	case "optimize", "profile.set", "profile.delete", "settings.set", "local.set", "local.preset.apply", "auto.commit":
		c.activationMu.Lock()
		defer c.activationMu.Unlock()
	}
	switch method {
	case "browser.start", "browser.status":
		return c.handleBrowser(method, raw)
	case "local.presets", "local.preset.apply", "local.get", "local.set", "local.metrics", "local.network", "local.record", "local.recommend", "local.compare":
		return c.handleLocal(method, raw)
	case "acceptance.readiness":
		var query struct {
			RelayID string `json:"relay_id"`
		}
		if err := security.Parse(raw, &query); err != nil {
			return nil, err
		}
		if query.RelayID == "" || len(query.RelayID) > 128 {
			return nil, errors.New("configured relay ID required")
		}
		return c.acceptanceReadiness(ctx, query.RelayID), nil
	case "relay.pool.diagnostics":
		var query struct {
			RelayID string `json:"relay_id"`
		}
		if err := security.Parse(raw, &query); err != nil {
			return nil, err
		}
		if query.RelayID == "" || len(query.RelayID) > 128 {
			return nil, errors.New("configured relay ID required")
		}
		c.mu.Lock()
		pool := c.relayPools[query.RelayID]
		c.mu.Unlock()
		if pool == nil {
			return nil, errors.New("relay pool unavailable")
		}
		return pool.Snapshot(), nil
	case "quality.diagnostics":
		return c.DropDiagnostics(), nil
	case "capture.changes.marker":
		if backend, ok := c.Backend.(interface {
			CaptureChangeMarker() windowsplatform.ChangeCursor
		}); ok {
			return backend.CaptureChangeMarker(), nil
		}
		return nil, errors.New("capture change cursor unavailable; update backend")
	case "capture.changes":
		var query windowsplatform.ChangeQuery
		if err := security.Parse(raw, &query); err != nil {
			return nil, err
		}
		if backend, ok := c.Backend.(interface {
			CaptureChanges(windowsplatform.ChangeQuery) (windowsplatform.ChangePage, error)
		}); ok {
			return backend.CaptureChanges(query)
		}
		return nil, errors.New("capture changes unavailable; update backend")
	case "capture.marker":
		if backend, ok := c.Backend.(interface {
			CaptureMarker() windowsplatform.EvidenceMarker
		}); ok {
			return backend.CaptureMarker(), nil
		}
		return nil, errors.New("capture markers unavailable; update service/backend")
	case "capture.diagnostics":
		var query windowsplatform.EvidenceQuery
		if len(raw) > 0 && string(raw) != "null" {
			if err := security.Parse(raw, &query); err != nil {
				return nil, err
			}
		}
		if query.CreationTime != 0 && query.PID == 0 {
			return nil, errors.New("creation_time requires exact PID")
		}
		if query.Source != "" {
			if _, err := netip.ParseAddrPort(query.Source); err != nil {
				return nil, errors.New("invalid diagnostic source")
			}
		}
		if len(query.Domain) > 253 || strings.ContainsAny(query.Domain, " /\\\t\r\n") {
			return nil, errors.New("invalid diagnostic domain")
		}
		if query.Since != nil || query.CreationTime != 0 || query.Source != "" || query.Domain != "" {
			if query.Destination != "" {
				if _, err := netip.ParseAddrPort(query.Destination); err != nil {
					return nil, errors.New("invalid diagnostic destination")
				}
			}
			if backend, ok := c.Backend.(interface {
				DiagnosticsScoped(windowsplatform.EvidenceQuery) (windowsplatform.CaptureEvidence, error)
			}); ok {
				return backend.DiagnosticsScoped(query)
			}
			return nil, errors.New("trial-scoped capture diagnostics unavailable; update service/backend")
		}
		if query.Destination != "" {
			if _, err := netip.ParseAddrPort(query.Destination); err != nil {
				return nil, errors.New("invalid diagnostic destination")
			}
			if backend, ok := c.Backend.(interface {
				DiagnosticsTarget(uint32, string) windowsplatform.CaptureEvidence
			}); ok {
				return backend.DiagnosticsTarget(query.PID, query.Destination), nil
			}
			return nil, errors.New("scoped capture diagnostics unavailable")
		}
		if backend, ok := c.Backend.(interface {
			Diagnostics(uint32) windowsplatform.CaptureEvidence
		}); ok {
			return backend.Diagnostics(query.PID), nil
		}
		return nil, errors.New("capture diagnostics unavailable on this backend")
	case "hot_attach.status":
		return c.hotAttachStatus(), nil
	case "status":
		return c.Snapshot(), nil
	case "profiles":
		return c.Profiles.List(), nil
	case "processes":
		return listProcesses(raw)
	case "profile.delete":
		var query struct {
			ID string `json:"id"`
		}
		if e := security.Parse(raw, &query); e != nil {
			return nil, e
		}
		c.mu.Lock()
		p, ok := c.localProfiles[query.ID]
		if !ok || !p.Custom {
			c.mu.Unlock()
			return nil, errors.New("only local custom profiles can be removed")
		}
		c.mu.Unlock()
		if e := c.reconfigure(query.ID, func() error {
			c.mu.Lock()
			delete(c.localProfiles, query.ID)
			c.mu.Unlock()
			c.Profiles.Delete(query.ID)
			c.Engine.SetPolicy(query.ID, routing.Policy{})
			return nil
		}); e != nil {
			return nil, e
		}
		if err := c.saveProfiles(); err != nil {
			return nil, err
		}
		return c.Profiles.List(), nil
	case "auto.plan", "auto.commit":
		c.pruneTrials(time.Now())
		return c.handleAuto(method, raw)
	case "profile.set":
		c.pruneTrials(time.Now())
		var p profiles.Profile
		if e := security.Parse(raw, &p); e != nil {
			return nil, e
		}
		// Old clients omit newly added policy fields. Preserve them for the
		// same executable identity; explicit null/empty values still clear them.
		var fields map[string]json.RawMessage
		if e := json.Unmarshal(raw, &fields); e != nil {
			return nil, e
		}
		for _, old := range c.Profiles.List() {
			if old.ID == p.ID && sameProfileExecutables(old.Executables, p.Executables) {
				if _, ok := fields["rules"]; !ok {
					p.Rules = old.Rules
				}
				if _, ok := fields["identity"]; !ok {
					p.Identity = old.Identity
				}
				if _, ok := fields["fallback"]; !ok {
					p.Fallback = old.Fallback
				}
				if _, ok := fields["app_preset"]; !ok {
					p.AppPreset = old.AppPreset
				}
				if _, ok := fields["auto_network"]; !ok {
					p.AutoNetwork = old.AutoNetwork
				}
				if _, ok := fields["auto_revision"]; !ok {
					p.AutoRevision = old.AutoRevision
				}
				break
			}
		}
		if p.Trial != nil {
			if !time.Now().Before(p.Trial.Expires) || p.Trial.Expires.After(time.Now().Add(15*time.Minute)) {
				return nil, errors.New("trial lifetime bound")
			}
			root, err := c.discovery.Lookup(p.Trial.PID)
			if err != nil || root.CreationTime != p.Trial.CreationTime || len(p.Executables) != 1 || !strings.EqualFold(root.Path, p.Executables[0]) {
				return nil, errors.New("trial root identity unavailable")
			}
			p.Identity = profiles.IdentityFor(root.Path, "")
			p.Identity.Helpers = []string{p.Identity.Executable}
		}
		prepareAuto(&p)
		if p.Mode == profiles.Auto && p.LocalStrategyID != "" && p.AutoNetwork == "" {
			for _, old := range c.Profiles.List() {
				if old.ID == p.ID && old.Mode == profiles.DirectPlus && old.LocalStrategyID == p.LocalStrategyID {
					n := c.network.Load()
					if n == nil || time.Since(n.ObservedAt) > 45*time.Second || len(n.VPN) > 0 || len(n.Conflicts) > 0 {
						return nil, errors.New("fresh VPN-free network required to adopt configured DIRECT+ into AUTO")
					}
					p.AutoNetwork = n.Fingerprint
					p.AutoRevision = c.baselineRevision(p)
				}
			}
		}
		if p.Mode == profiles.Block && !c.Demo && (c.Backend == nil || !c.Backend.Status().Capabilities.TCP) {
			return nil, errors.New("BLOCK profile requires UDP/TCP capture; current backend does not provide it")
		}
		if err := p.Validate(); err != nil {
			return nil, err
		}
		for _, old := range c.Profiles.List() {
			if old.ID == p.ID {
				a, _ := json.Marshal(old)
				b, _ := json.Marshal(p)
				if bytes.Equal(a, b) {
					return c.Profiles.List(), nil
				}
			}
		}
		if e := c.reconfigure(p.ID, func() error {
			if e := c.Profiles.Upsert(p); e != nil {
				return e
			}
			c.mu.Lock()
			c.localProfiles[p.ID] = p
			c.mu.Unlock()
			c.syncPolicies()
			return nil
		}); e != nil {
			return nil, e
		}
		if err := c.saveProfiles(); err != nil {
			return nil, err
		}
		return c.Profiles.List(), nil
	case "optimize":
		var v struct {
			Enabled bool `json:"enabled"`
		}
		if e := security.Parse(raw, &v); e != nil {
			return nil, e
		}
		if !c.Demo && c.Backend == nil {
			return nil, errors.New("Windows capture backend is not configured")
		}
		c.mu.Lock()
		wasEnabled := c.enabled
		if v.Enabled {
			c.enabled = true
		}
		c.mu.Unlock()
		if c.Backend != nil {
			if v.Enabled {
				if !c.Backend.Status().Ready {
					if e := c.Backend.Start(c.ctx, c.Classify); e != nil {
						c.mu.Lock()
						c.enabled = wasEnabled
						c.mu.Unlock()
						return nil, e
					}
				}
			} else {
				if e := c.Backend.Stop(); e != nil {
					return nil, e
				}
			}
		}
		c.mu.Lock()
		c.enabled = v.Enabled
		c.mu.Unlock()
		if v.Enabled != wasEnabled {
			c.browser.mu.Lock()
			bs := c.browser.server
			c.browser.mu.Unlock()
			if bs != nil {
				bs.CloseConnections()
			}
		}
		if v.Enabled && !wasEnabled {
			if b, ok := c.Backend.(interface{ RequestHotAttach(string) }); ok {
				b.RequestHotAttach("optimize_enabled")
			}
		}
		c.Measurements.Active.Store(v.Enabled)
		return c.Snapshot(), nil
	case "settings.set":
		var s Settings
		if e := security.Parse(raw, &s); e != nil {
			return nil, e
		}
		if (s.Failover != "DIRECT" && s.Failover != "BLOCK") || (s.UpdateChannel != "stable" && s.UpdateChannel != "preview") {
			return nil, errors.New("invalid settings")
		}
		c.mu.Lock()
		oldSettings := c.settings
		c.mu.Unlock()
		apply := func() error {
			c.mu.Lock()
			c.settings = s
			c.mu.Unlock()
			c.syncPolicies()
			return nil
		}
		var updateErr error
		if oldSettings.Automatic != s.Automatic || oldSettings.Failover != s.Failover {
			updateErr = c.reconfigure("", apply)
		} else {
			updateErr = apply()
		}
		if updateErr != nil {
			return nil, updateErr
		}
		if oldSettings.Automatic != s.Automatic {
			c.browser.mu.Lock()
			bs := c.browser.server
			c.browser.mu.Unlock()
			if bs != nil {
				bs.CloseConnections()
			}
		}
		if err := c.saveSettings(); err != nil {
			return nil, err
		}
		return s, nil
	case "diagnostics":
		return c.Diagnostics(ctx), nil
	case "demo.fail-relay":
		if !c.Demo || c.FailRelay == nil {
			return nil, errors.New("relay failure control is only available in local demo")
		}
		var v struct {
			ID string `json:"id"`
		}
		if e := security.Parse(raw, &v); e != nil {
			return nil, e
		}
		if e := c.FailRelay(v.ID); e != nil {
			return nil, e
		}
		return c.Snapshot(), nil
	default:
		return nil, fmt.Errorf("unsupported IPC method %q", method)
	}
}

type Check struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail"`
}

func (c *Controller) Diagnostics(ctx context.Context) []Check {
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	out := []Check{}
	if c.Backend != nil {
		b := c.Backend.Status()
		tcp, udp, ipv4 := "unavailable", "unavailable", "unavailable"
		if b.Ready && b.Capabilities.TCP {
			tcp = "healthy"
		}
		if b.Ready && b.Capabilities.UDP {
			udp = "healthy"
		}
		if b.Ready && b.Capabilities.IPv4 {
			ipv4 = "healthy"
		}
		out = append(out, Check{"TCP interception", tcp, "backend readiness, not an external TCP proof"}, Check{"UDP interception", udp, "backend readiness"}, Check{"IPv4", ipv4, "capture backend capability"})
	}
	out = append(out, Check{"IPv6", "unsupported", "no IPv6 relay transport: new attributable selected IPv6 packets blocked at NETWORK; existing/unattributed/unsupported traffic fails open"}, Check{"HTTPS through relay", "not configured", "not measured; encrypted TCP alone is not HTTPS evidence"}, Check{"QUIC through relay", "not configured", "not measured; UDP/443 alone is not QUIC evidence"}, Check{"Browser DNS", "not configured", "not attributed: system resolver service traffic stays DIRECT; selected browser DoH follows its socket profile"})
	dnsHost := "localhost"
	if c.ControlPlaneURL == "" {
		out = append(out, Check{"Control Plane", "not configured", "no endpoint configured"})
	} else {
		u, e := url.Parse(c.ControlPlaneURL)
		if e == nil {
			dnsHost = u.Hostname()
		}
		req, e := http.NewRequestWithContext(ctx, "GET", c.ControlPlaneURL+"/healthz", nil)
		status := "unavailable"
		if e == nil {
			client := c.DiagnosticHTTPClient
			if client == nil {
				client = &http.Client{Timeout: time.Second}
			}
			probe, stop := context.WithTimeout(ctx, time.Second)
			response, e := client.Do(req.WithContext(probe))
			stop()
			if e == nil {
				response.Body.Close()
				if response.StatusCode == 200 {
					status = "healthy"
				}
			}
		}
		out = append(out, Check{"Control Plane", status, "HTTP health request"})
	}
	_, e := net.DefaultResolver.LookupHost(ctx, dnsHost)
	dns := "healthy"
	if e != nil {
		dns = "unavailable"
	}
	out = append(out, Check{"DNS", dns, "resolved configured control host (localhost in standalone demo)"})
	measurements := c.Measurements.Round(ctx)
	udpStatus, relayStatus := "not configured", "not configured"
	for _, m := range measurements {
		if m.RouteID == "" {
			continue
		}
		c.Selector.Observe(m)
		status := "unavailable"
		if m.Available {
			status = "healthy"
		}
		if m.RouteID == optimizer.Direct {
			udpStatus = status
		} else if relayStatus != "healthy" {
			relayStatus = status
		}
	}
	out = append(out, Check{"UDP", udpStatus, "real application-level UDP echo probes"}, Check{"Relay", relayStatus, "authenticated relay UDP echo path"})
	tcpStatus := "not configured"
	if c.TCPDiagnosticTarget != "" {
		conn, e := (&net.Dialer{Timeout: 500 * time.Millisecond}).DialContext(ctx, "tcp", c.TCPDiagnosticTarget)
		tcpStatus = "unavailable"
		if e == nil {
			conn.Close()
			tcpStatus = "healthy"
		}
	}
	out = append(out, Check{"TCP", tcpStatus, "TCP connect to configured diagnostic target"})
	mtuStatus := "not configured"
	for _, candidate := range c.Measurements.Candidates() {
		if candidate.Metadata.ID != optimizer.Direct {
			continue
		}
		mtuStatus = "unavailable"
		probe, stop := context.WithTimeout(ctx, 500*time.Millisecond)
		flow, e := candidate.Open(probe, candidate.Target)
		if e == nil {
			payload := bytes.Repeat([]byte{0x6e}, 1100)
			if flow.Send(payload) == nil {
				got, e := flow.Receive(probe)
				if e == nil && bytes.Equal(got, payload) {
					mtuStatus = "healthy"
				}
			}
			flow.Close()
		}
		stop()
		break
	}
	out = append(out, Check{"MTU", mtuStatus, "1100-byte echo payload check; not end-to-end Internet PMTU discovery"})
	selection := c.Selector.Select(time.Now())
	for _, route := range selection.Ranked {
		if route.ID == selection.Primary {
			out = append(out, Check{"Latency / Loss", "measured", fmt.Sprintf("RTT %.1f ms, jitter %.1f ms, loss %.1f%%", float64(route.Measurement.RTT)/float64(time.Millisecond), float64(route.Measurement.Jitter)/float64(time.Millisecond), route.Measurement.Loss*100)})
			break
		}
	}
	return out
}
func Tuple(src, dst netip.AddrPort) windowsplatform.Tuple {
	return windowsplatform.Tuple{Protocol: "udp", Source: src, Destination: dst}
}

func (c *Controller) RecordTraffic(app string, rx, tx uint64) {
	c.RX.Add(rx)
	c.TX.Add(tx)
	c.mu.Lock()
	if len(c.recentApps) < 128 {
		c.recentApps[app] = time.Now()
	}
	c.mu.Unlock()
}

func (c *Controller) saveSettings() error {
	if c.customPath == "" {
		return nil
	}
	c.mu.Lock()
	raw, e := json.Marshal(c.settings)
	c.mu.Unlock()
	if e != nil {
		return e
	}
	dir := filepath.Dir(c.customPath)
	if e = os.MkdirAll(dir, 0700); e != nil {
		return e
	}
	f, e := os.CreateTemp(dir, ".settings-*")
	if e != nil {
		return e
	}
	name := f.Name()
	defer os.Remove(name)
	f.Chmod(0600)
	_, e = f.Write(raw)
	f.Close()
	if e != nil {
		return e
	}
	return os.Rename(name, filepath.Join(dir, "settings.json"))
}
