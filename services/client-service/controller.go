// Package clientservice owns routing, telemetry and privileged commands outside the UI.
package clientservice

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/buzzmean/neexitlag/core/measurement"
	"github.com/buzzmean/neexitlag/core/optimizer"
	"github.com/buzzmean/neexitlag/core/profiles"
	"github.com/buzzmean/neexitlag/core/routing"
	"github.com/buzzmean/neexitlag/core/security"
	"github.com/buzzmean/neexitlag/core/telemetry"
	windowsplatform "github.com/buzzmean/neexitlag/platform/windows"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"sort"
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
	RelayTransport *RelayTransportReadiness      `json:"relay_transport,omitempty"`
	Version        string                        `json:"version"`
	Connected      bool                          `json:"connected"`
	Optimizing     bool                          `json:"optimizing"`
	Demo           bool                          `json:"demo"`
	Backend        windowsplatform.CaptureStatus `json:"backend"`
	Primary        string                        `json:"primary"`
	Secondary      string                        `json:"secondary"`
	Tertiary       string                        `json:"tertiary"`
	Routes         []RouteView                   `json:"routes"`
	Profiles       []profiles.Profile            `json:"profiles"`
	RXBytes        uint64                        `json:"rx_bytes"`
	TXBytes        uint64                        `json:"tx_bytes"`
	History        optimizer.RouteHistory        `json:"history"`
	Flows          []routing.Flow                `json:"flows"`
	Settings       Settings                      `json:"settings"`
	ControlPlane   string                        `json:"control_plane"`
	ActiveApps     []string                      `json:"active_apps"`
}
type Controller struct {
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
	c := &Controller{Engine: routing.New(4096), Selector: optimizer.New(optimizer.DefaultParameters()), Measurements: measurement.New(measurement.Defaults()), Profiles: profiles.New(profiles.Defaults()), recentApps: make(map[string]time.Time), localProfiles: make(map[string]profiles.Profile), acceptanceRelays: make(map[string]acceptanceRelaySource), transportDrops: make(map[string]func() map[string]telemetry.Metric), settings: Settings{Automatic: true, Failover: "DIRECT", UpdateChannel: "stable"}, ctx: ctx, cancel: cancel}
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
		case profiles.Manual:
			policy.Route = routing.Route{RelayID: p.RelayID}
		case profiles.Block:
			policy.Block = true
		}
		c.mu.Lock()
		policy.BlockOnFailure = c.settings.Failover == "BLOCK"
		c.mu.Unlock()
		c.Engine.SetPolicy(p.ID, policy)
	}
}
func (c *Controller) Start() {
	c.wg.Add(1)
	go func() { defer c.wg.Done(); c.Measurements.Run(c.ctx, c.Selector.Observe) }()
}
func (c *Controller) Close() {
	c.cancel()
	if c.Backend != nil {
		c.Backend.Stop()
	}
	c.wg.Wait()
}
func (c *Controller) SetCustomPath(path string) error {
	c.customPath = path
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
	profile, ok := c.Profiles.Match(identity)
	if !ok {
		return decision
	}
	f := routing.Flow{Protocol: t.Protocol, Source: t.Source, Destination: t.Destination, PID: identity.PID, ProcessPath: identity.Path, ApplicationID: profile.ID}
	decision.Flow = f
	c.mu.Lock()
	decision.BlockOnFailure = c.settings.Failover == "BLOCK"
	c.mu.Unlock()
	switch profile.Mode {
	case profiles.Block:
		decision.Block = true
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
	snapshot := Snapshot{Version: "0.3.0-dev", Connected: c.Demo || backend.Ready, Optimizing: enabled && (c.Demo || backend.Ready), Demo: c.Demo, Backend: backend, Primary: selection.Primary, Secondary: selection.Secondary, Tertiary: selection.Tertiary, Routes: views, Profiles: c.Profiles.List(), RXBytes: rx, TXBytes: tx, History: history, Flows: flows, Settings: settings, ControlPlane: c.ControlPlaneURL, ActiveApps: apps}
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
	case "quality.diagnostics":
		return c.DropDiagnostics(), nil
	case "capture.diagnostics":
		var query struct {
			PID uint32 `json:"pid"`
		}
		if len(raw) > 0 && string(raw) != "null" {
			if err := security.Parse(raw, &query); err != nil {
				return nil, err
			}
		}
		if backend, ok := c.Backend.(interface {
			Diagnostics(uint32) windowsplatform.CaptureEvidence
		}); ok {
			return backend.Diagnostics(query.PID), nil
		}
		return nil, errors.New("capture diagnostics unavailable on this backend")
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
		delete(c.localProfiles, query.ID)
		c.mu.Unlock()
		c.Profiles.Delete(query.ID)
		c.Engine.SetPolicy(query.ID, routing.Policy{})
		if e := c.saveProfiles(); e != nil {
			return nil, e
		}
		return c.Profiles.List(), nil
	case "profile.set":
		var p profiles.Profile
		if e := security.Parse(raw, &p); e != nil {
			return nil, e
		}
		if p.Mode == profiles.Block && !c.Demo && (c.Backend == nil || !c.Backend.Status().Capabilities.TCP) {
			return nil, errors.New("BLOCK profile requires full UDP/TCP capture; WinDivert MVP supports IPv4 UDP only")
		}
		if e := c.Profiles.Upsert(p); e != nil {
			return nil, e
		}
		c.mu.Lock()
		c.localProfiles[p.ID] = p
		c.mu.Unlock()
		if e := c.saveProfiles(); e != nil {
			return nil, e
		}
		c.syncPolicies()
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
		if c.Backend != nil {
			if v.Enabled {
				if !c.Backend.Status().Ready {
					if e := c.Backend.Start(c.ctx, c.Classify); e != nil {
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
		c.settings = s
		c.mu.Unlock()
		if e := c.saveSettings(); e != nil {
			return nil, e
		}
		c.syncPolicies()
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
