package clientservice

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/buzzmean/neexitlag/core/browserproxy"
	"github.com/buzzmean/neexitlag/core/flowpolicy"
	"github.com/buzzmean/neexitlag/core/localstrategy"
	"github.com/buzzmean/neexitlag/core/profiles"
	"github.com/buzzmean/neexitlag/core/routing"
	"github.com/buzzmean/neexitlag/core/security"
	wp "github.com/buzzmean/neexitlag/platform/windows"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"
)

type browserRuntime struct {
	address string
	mu      sync.Mutex
	server  *browserproxy.Server
	learned *flowpolicy.BrowserLearning
	probes  chan struct{}
	local   map[string]*localstrategy.LocalStrategyEngine
	plans   map[flowpolicy.LearnKey]chan struct{}
}

func newBrowserRuntime() *browserRuntime {
	return &browserRuntime{learned: flowpolicy.NewBrowserLearning(), probes: make(chan struct{}, 4), local: map[string]*localstrategy.LocalStrategyEngine{}, plans: map[flowpolicy.LearnKey]chan struct{}{}}
}
func (c *Controller) browserOwner(ctx context.Context, source, local netip.AddrPort) (browserproxy.Owner, error) {
	id, err := wp.LookupTCPPeer(source, local)
	if err != nil {
		return browserproxy.Owner{}, err
	}
	sid, sidErr := wp.ProcessUserSID(id)
	if sidErr != nil || c.BrowserOwnerSID == "" || sid != c.BrowserOwnerSID {
		return browserproxy.Owner{}, errors.New("browser connection belongs to another user")
	}
	p, ok := c.matchApplication(id)
	if !ok || !browserAuto(p) {
		return browserproxy.Owner{}, errors.New("Chrome AUTO owner required")
	}
	return browserproxy.Owner{PID: id.PID, Birth: id.CreationTime, Exe: id.Path, Profile: p.ID}, nil
}
func (c *Controller) browserProfile(owner browserproxy.Owner) (profilesResult, error) {
	id := wp.ProcessIdentity{PID: owner.PID, CreationTime: owner.Birth, Path: owner.Exe}
	if !wp.IsCurrent(id) {
		return profilesResult{}, errors.New("Chrome owner exited/reused")
	}
	p, ok := c.matchApplication(id)
	if !ok || p.ID != owner.Profile || !browserAuto(p) {
		return profilesResult{}, errors.New("Chrome AUTO profile changed")
	}
	network := ""
	if n := c.network.Load(); n != nil && n.Error == "" && time.Since(n.ObservedAt) <= 45*time.Second {
		network = n.Fingerprint
	}
	c.mu.Lock()
	enabled := c.enabled && c.settings.Automatic
	c.mu.Unlock()
	return profilesResult{profile: p, id: id, network: network, enabled: enabled}, nil
}

// Keep mutable ownership resolution separate from immutable request plans.
type profilesResult struct {
	profile profiles.Profile
	id      wp.ProcessIdentity
	network string
	enabled bool
}

func (c *Controller) browserLocalAllowed(network string) bool {
	n := c.network.Load()
	return c.Local.Config().Enabled && n != nil && n.Fingerprint == network && n.Error == "" && time.Since(n.ObservedAt) <= 45*time.Second && len(n.VPN) == 0 && len(n.Conflicts) == 0
}
func (c *Controller) browserScopedLocal(p profiles.Profile, host, network string, port uint16) (string, *localstrategy.LocalStrategyEngine) {
	if !c.hasBaseline(p, network) || !c.browserLocalAllowed(network) {
		return "", nil
	}
	var selected *localstrategy.Profile
	for _, v := range c.Local.Config().User {
		if v.ID == p.LocalStrategyID && v.Enabled {
			copy := v
			selected = &copy
			break
		}
	}
	if selected == nil {
		return "", nil
	}
	if youtubeDomain(host) {
		return selected.ID, c.Local
	}
	raw, _ := json.Marshal([]any{host, port, p.AutoRevision})
	sum := sha256.Sum256(raw)
	key := "br-" + hex.EncodeToString(sum[:12])
	c.browser.mu.Lock()
	defer c.browser.mu.Unlock()
	if engine := c.browser.local[key]; engine != nil {
		return key, engine
	}
	// Bound ephemeral candidates separately; never edit the user's working preset
	// or its revision, and never activate a candidate for another domain.
	if len(c.browser.local) >= 512 {
		return "", nil
	}
	selected.ID = key
	selected.Domains = []string{host}
	selected.IPs = nil
	selected.QuicFallback = false
	selected.Ports = []uint16{port}
	engine := localstrategy.New()
	if engine.Reload(localstrategy.Config{Version: 1, Enabled: true, User: []localstrategy.Profile{*selected}}) != nil {
		return "", nil
	}
	c.browser.local[key] = engine
	return key, engine
}
func (c *Controller) browserLocal(strategy string) *localstrategy.LocalStrategyEngine {
	for _, v := range c.Local.Config().User {
		if v.ID == strategy && v.Enabled {
			return c.Local
		}
	}
	c.browser.mu.Lock()
	defer c.browser.mu.Unlock()
	return c.browser.local[strategy]
}
func browserIdentity(q browserproxy.Request) wp.ProcessIdentity {
	return wp.ProcessIdentity{PID: q.Owner.PID, CreationTime: q.Owner.Birth, Path: q.Owner.Exe}
}
func (c *Controller) browserOpen(ctx context.Context, q browserproxy.Request, plan browserproxy.Plan) (net.Conn, error) {
	current, err := c.browserProfile(q.Owner)
	if err != nil {
		return nil, err
	}
	if current.network != plan.Network || current.enabled != plan.Enabled {
		return nil, errors.New("browser policy changed before upstream open")
	}
	if current.enabled && protectedBrowserDomain(q.Host) && plan.Policy.Action != flowpolicy.RelayRequired {
		return nil, errors.New("protected origin cannot downgrade")
	}
	d := applyFlowResult(current.profile, wp.Decision{}, plan.Policy, current.network, c.browserLocal(plan.Policy.Strategy))
	d.OriginDomain = q.Host
	d.Flow = routing.Flow{Protocol: "tcp", Source: q.Source, Destination: plan.Destination, PID: q.Owner.PID, ProcessPath: q.Owner.Exe, ApplicationID: q.Owner.Profile, Network: current.network, PinnedPolicy: d.Flow.PinnedPolicy}
	if d.Block {
		return nil, errors.New("browser route blocked")
	}
	if d.Route.Local() {
		if !c.browserLocalAllowed(current.network) {
			return nil, errors.New("native local strategy unavailable on current network")
		}
		backend, ok := c.Backend.(interface {
			OpenBrowserLocalTCP(context.Context, wp.ProcessIdentity, netip.AddrPort, netip.AddrPort, wp.Decision) (net.Conn, error)
		})
		if !ok {
			return nil, errors.New("native DIRECT+ unavailable")
		}
		return backend.OpenBrowserLocalTCP(ctx, browserIdentity(q), q.Source, plan.Destination, d)
	}
	return c.Engine.OpenTCP(ctx, d.Flow)
}
func (c *Controller) browserTransform(q browserproxy.Request, plan browserproxy.Plan) bool {
	backend, ok := c.Backend.(interface {
		BrowserTransformApplied(wp.ProcessIdentity, netip.AddrPort, netip.AddrPort) bool
	})
	return ok && backend.BrowserTransformApplied(browserIdentity(q), q.Source, plan.Destination)
}
func (c *Controller) browserPlan(ctx context.Context, q browserproxy.Request) (browserproxy.Plan, error) {
	current, err := c.browserProfile(q.Owner)
	if err != nil {
		return browserproxy.Plan{}, err
	}
	lookup, stop := context.WithTimeout(ctx, time.Second)
	resolver := c.BrowserResolver
	if resolver == nil {
		resolver = net.DefaultResolver
	}
	addresses, err := resolver.LookupNetIP(lookup, "ip", q.Host)
	stop()
	if err != nil || len(addresses) == 0 {
		return browserproxy.Plan{}, errors.New("origin DNS unavailable; system DNS unchanged")
	}
	// Prefer IPv4 for the existing native DIRECT+ backend. IPv6 relay/direct can
	// use the existing authenticated transport, not an unsupported native bypass.
	sort.SliceStable(addresses, func(i, j int) bool { return addresses[i].Is4() && !addresses[j].Is4() })
	dest := netip.AddrPortFrom(addresses[0].Unmap(), q.Port)
	tuple := wp.Tuple{Protocol: "tcp", Source: q.Source, Destination: dest}
	result := c.browserFlow(current.profile, current.id, tuple, q.Host, current.network)
	if !current.enabled {
		result = flowpolicy.Result{Action: flowpolicy.Direct, Fallback: flowpolicy.FailOpen, Source: "optimization_disabled", Reason: "user disabled optimization"}
	}
	plan := browserproxy.Plan{Network: current.network, Enabled: current.enabled, Policy: result, Destination: dest, Probe: "NOT_PROBED"}
	if !current.enabled || protectedBrowserDomain(q.Host) || youtubeDomain(q.Host) || result.Source == "rule" {
		return plan, nil
	}
	if dest.Addr().IsPrivate() || dest.Addr().IsLoopback() || dest.Addr().IsLinkLocalUnicast() {
		if c.BrowserResolver == nil {
			plan.Probe = "PRIVATE_ORIGIN_NO_AUTOPROBE"
			return plan, nil
		}
	}
	if current.network == "" {
		plan.Probe = "NETWORK_IDENTITY_UNAVAILABLE"
		return plan, nil
	}
	if _, ipErr := netip.ParseAddr(q.Host); ipErr == nil {
		plan.Probe = "IP_LITERAL_NO_DOMAIN_LEARNING"
		return plan, nil
	}
	key := c.learnKey(current.profile, current.network, "tcp", q.Host, q.Port)
	if q.Secure {
		key.Revision += ":https"
	} else {
		key.Revision += ":http"
	}
	if cached, ok := c.browser.learned.Lookup(key, time.Now()); ok && (cached.Confirmed || !cached.RetryAfter.IsZero()) {
		plan.Policy = cached.Route
		plan.Probe = cached.Outcome
		plan.HTTPStatus = cached.HTTPStatus
		return plan, nil
	}
	// One bounded probe sequence per origin/network/revision; waiters never create
	// independent duplicate probes or borrow another domain's result.
	c.browser.mu.Lock()
	pending, exists := c.browser.plans[key]
	if !exists {
		if len(c.browser.plans) >= 32 {
			c.browser.mu.Unlock()
			return plan, errors.New("browser probe queue full")
		}
		pending = make(chan struct{})
		c.browser.plans[key] = pending
	}
	c.browser.mu.Unlock()
	if exists {
		select {
		case <-pending:
			if v, ok := c.browser.learned.Lookup(key, time.Now()); ok {
				plan.Policy = v.Route
				plan.Probe = v.Outcome
				plan.HTTPStatus = v.HTTPStatus
				return plan, nil
			}
			return plan, errors.New("origin probe unavailable")
		case <-ctx.Done():
			return plan, ctx.Err()
		}
	}
	defer func() { c.browser.mu.Lock(); delete(c.browser.plans, key); close(pending); c.browser.mu.Unlock() }()
	select {
	case c.browser.probes <- struct{}{}:
		defer func() { <-c.browser.probes }()
	case <-ctx.Done():
		return plan, ctx.Err()
	}
	strategy, _ := c.browserScopedLocal(current.profile, q.Host, current.network, q.Port)
	relay := current.profile.RelayID
	if relay == "" {
		relay = "VPS-01"
	}
	attempt, err := c.browser.learned.Begin(key, strategy, relay, time.Now())
	if err != nil {
		return plan, err
	}
	for tries := 0; tries < 3; tries++ {
		plan.Policy = attempt.Route
		started := time.Now()
		probe, stop := context.WithTimeout(ctx, 1500*time.Millisecond)
		status, rx, networkError := browserproxy.ProbeOrigin(probe, q.Host, q.Port, q.Secure, c.BrowserProbeRoots, func(ctx context.Context) (net.Conn, error) { return c.browserOpen(ctx, q, plan) })
		stop()
		applied := plan.Policy.Action == flowpolicy.DirectPlus && c.browserTransform(q, plan)
		outcome := networkError
		if outcome == "" {
			outcome = "HTTP_RESPONSE"
		}
		if plan.Policy.Action == flowpolicy.DirectPlus {
			if networkError == "" && !applied {
				networkError = "NO_TRANSFORM"
				outcome = networkError
				rx = 0
				status = 0
			} else if networkError != "" && applied && networkError != "TLS_CERTIFICATE_ERROR" {
				outcome = "TRANSFORM_APPLIED_CONNECTION_FAILED"
			}
		}
		plan.Trials = append(plan.Trials, browserproxy.Trial{Route: string(plan.Policy.Action), Outcome: outcome, HTTPStatus: status, TransformApplied: applied, DurationMS: time.Since(started).Milliseconds()})
		after, ownerErr := c.browserProfile(q.Owner)
		if ownerErr != nil || after.network != current.network || after.enabled != current.enabled {
			return plan, errors.New("network/activation changed during origin probe")
		}
		attempt, err = c.browser.learned.Observe(key, flowpolicy.BrowserObservation{Token: attempt.Token, HTTPStatus: status, NetworkError: networkError, ResponseBytes: rx, Correlated: true}, strategy, relay, time.Now())
		if err != nil {
			return plan, err
		}
		if attempt.Confirmed || !attempt.RetryAfter.IsZero() {
			plan.Policy = attempt.Route
			plan.Probe = attempt.Outcome
			plan.HTTPStatus = status
			return plan, nil
		}
	}
	return plan, errors.New("bounded origin probes exhausted")
}
func (c *Controller) handleBrowser(method string, raw json.RawMessage) (any, error) {
	c.browser.mu.Lock()
	defer c.browser.mu.Unlock()
	if method == "browser.status" {
		if c.browser.server == nil {
			return browserproxy.Snapshot{Records: []browserproxy.Record{}}, nil
		}
		view := c.browser.server.Snapshot()
		sort.Slice(view.Records, func(i, j int) bool { return view.Records[i].ID > view.Records[j].ID })
		for {
			raw, _ := json.Marshal(view)
			if len(raw) < 58000 {
				break
			}
			view.Truncated = true
			view.Records = view.Records[:len(view.Records)-1]
		}
		return view, nil
	}
	var q struct {
		Profile   string `json:"profile_id"`
		Directory string `json:"directory"`
	}
	if err := security.Parse(raw, &q); err != nil {
		return nil, err
	}
	found := false
	for _, p := range c.Profiles.List() {
		if p.ID == q.Profile && browserAuto(p) {
			found = true
		}
	}
	if c.BrowserOwnerSID == "" {
		return nil, errors.New("installation owner SID required for Chrome transport")
	}
	c.mu.Lock()
	enabled := c.enabled && c.settings.Automatic
	c.mu.Unlock()
	if !enabled {
		return nil, errors.New("Chrome site routing requires enabled optimization and automatic routing")
	}
	if !found || q.Directory == "" || len(q.Directory) > 2048 {
		return nil, errors.New("selected Chrome AUTO profile and persistent user directory required")
	}
	if c.browser.server == nil {
		server, err := browserproxy.Listen(c.ctx, browserproxy.Config{Address: c.browser.address, Authorize: c.browserOwner, Plan: c.browserPlan, Open: c.browserOpen, TransformApplied: c.browserTransform})
		if err != nil {
			return nil, err
		}
		c.browser.server = server
		c.browser.address = server.Address()
		if c.customPath != "" {
			if err := c.saveBrowserAddress(); err != nil {
				c.browser.server = nil
				c.browser.mu.Unlock()
				server.Close()
				c.browser.mu.Lock()
				return nil, err
			}
		}
	}
	args, err := browserproxy.ChromeArguments(c.browser.server.Address(), q.Directory)
	if err != nil {
		return nil, err
	}
	return map[string]any{"arguments": args, "address": c.browser.server.Address(), "scope": "Chrome launch only; system proxy unchanged", "http3": "HTTP CONNECT uses TCP; QUIC relay transport unchanged", "verification": "origin probe only; authenticated browser page result is not observed"}, nil
}

func (c *Controller) saveBrowserAddress() error {
	path := filepath.Join(filepath.Dir(c.customPath), "browser-transport.json")
	raw, _ := json.Marshal(map[string]string{"address": c.browser.address})
	f, err := os.CreateTemp(filepath.Dir(path), ".browser-transport-*")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	f.Chmod(0600)
	if _, err = f.Write(raw); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
}

// Restore the existing non-secret local port only after normal service startup
// has enabled capture. A service restart must not require changing Chrome's PAC.
func (c *Controller) RestoreBrowserTransport() error {
	if c.customPath == "" {
		return nil
	}
	raw, err := os.ReadFile(filepath.Join(filepath.Dir(c.customPath), "browser-transport.json"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	var saved struct {
		Address string `json:"address"`
	}
	if err = security.Parse(raw, &saved); err != nil {
		return err
	}
	host, p, err := net.SplitHostPort(saved.Address)
	port, pe := strconv.ParseUint(p, 10, 16)
	if err != nil || pe != nil || host != "127.0.0.1" || port == 0 {
		return errors.New("invalid saved Chrome loopback port")
	}
	c.mu.Lock()
	ready := !c.settings.Automatic || c.enabled
	c.mu.Unlock()
	if !ready {
		return errors.New("Chrome transport restore waits for enabled optimization")
	}
	c.browser.mu.Lock()
	defer c.browser.mu.Unlock()
	c.browser.address = saved.Address
	server, err := browserproxy.Listen(c.ctx, browserproxy.Config{Address: saved.Address, Authorize: c.browserOwner, Plan: c.browserPlan, Open: c.browserOpen, TransformApplied: c.browserTransform})
	if err != nil {
		return err
	}
	c.browser.server = server
	return nil
}
