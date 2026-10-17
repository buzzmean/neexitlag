package clientservice

import (
	"github.com/buzzmean/neexitlag/core/flowpolicy"
	"github.com/buzzmean/neexitlag/core/profiles"
	"github.com/buzzmean/neexitlag/core/routing"
	wp "github.com/buzzmean/neexitlag/platform/windows"
	"path"
	"strings"
	"time"
)

func presetFor(p profiles.Profile) string {
	if p.AppPreset != "" {
		return p.AppPreset
	}
	for _, exe := range p.Executables {
		switch path.Base(strings.ToLower(strings.ReplaceAll(exe, "\\", "/"))) {
		case "discord.exe":
			return "discord"
		case "chatgpt.exe":
			return "openai"
		case "msedge.exe":
			return "openai-web"
		}
	}
	return ""
}
func prepareAuto(p *profiles.Profile) {
	if p.Mode != profiles.Auto {
		return
	}
	if browserAuto(*p) && len(p.BrowserRecoveryRules) > 0 {
		if len(p.Rules) == 0 {
			p.Rules = append([]flowpolicy.Rule(nil), p.BrowserRecoveryRules...)
		}
		p.BrowserRecoveryRules = nil
	}
	if p.AppPreset == "" {
		p.AppPreset = presetFor(*p)
	}
	if p.Fallback == "" && (p.AppPreset == "openai" || p.AppPreset == "openai-web" || p.AppPreset == "discord") {
		p.Fallback = flowpolicy.Required
	}
	if p.Identity == nil && p.AppPreset != "" && len(p.Executables) > 0 && strings.ContainsAny(p.Executables[0], "/\\") {
		p.Identity = profiles.IdentityFor(p.Executables[0], p.AppPreset)
	}
}

type groupBinding struct {
	identity   profiles.Identity
	profile    string
	trialPID   uint32
	trialBirth uint64
}

func (c *Controller) matchApplication(id wp.ProcessIdentity) (profiles.Profile, bool) {
	direct, hasDirect := c.Profiles.Match(id)
	possible := false
	for _, p := range c.Profiles.List() {
		if p.Identity != nil {
			for _, h := range p.Identity.Helpers {
				if strings.EqualFold(path.Base(strings.ReplaceAll(id.Path, "\\", "/")), h) {
					possible = true
				}
			}
		}
	}
	if !possible {
		return direct, hasDirect
	}
	c.groupMu.Lock()
	defer c.groupMu.Unlock()
	if b, ok := c.groupBindings[id.PID]; ok && b.identity.CreationTime == id.CreationTime && strings.EqualFold(b.identity.Path, id.Path) {
		for _, p := range c.Profiles.List() {
			if p.ID == b.profile && validGroupBinding(b, p) {
				return p, true
			}
		}
	}
	_, nativeAncestry := c.discovery.(interface {
		Ancestry(wp.ProcessIdentity) ([]wp.ProcessIdentity, error)
	})
	if nativeAncestry || c.groupAt.IsZero() || time.Since(c.groupAt) > time.Second {
		var all []wp.ProcessIdentity
		var err error
		if source, ok := c.discovery.(interface {
			Ancestry(wp.ProcessIdentity) ([]wp.ProcessIdentity, error)
		}); ok {
			all, err = source.Ancestry(id)
		} else {
			all, err = c.discovery.List()
		}
		if err != nil || len(all) > 4096 {
			return direct, hasDirect
		}
		c.groupAll = all
		c.groupAt = time.Now()
		live := map[uint32]profiles.Identity{}
		for _, v := range all {
			live[v.PID] = v
		}
		for pid, b := range c.groupBindings {
			if nativeAncestry {
				if pid == id.PID && b.identity.CreationTime != id.CreationTime {
					delete(c.groupBindings, pid)
				}
				continue
			}
			v, ok := live[pid]
			if !ok || v.CreationTime != b.identity.CreationTime || !strings.EqualFold(v.Path, b.identity.Path) {
				delete(c.groupBindings, pid)
			}
		}
	}
	if b, ok := c.groupBindings[id.PID]; ok && b.identity.CreationTime == id.CreationTime && strings.EqualFold(b.identity.Path, id.Path) {
		for _, p := range c.Profiles.List() {
			if p.ID == b.profile && validGroupBinding(b, p) {
				return p, true
			}
		}
	}
	// Discovery records parent birth identities from the same bounded snapshot.
	for _, v := range c.groupAll {
		if v.PID == id.PID && v.CreationTime == id.CreationTime && strings.EqualFold(v.Path, id.Path) {
			id = v
			break
		}
	}
	for _, p := range c.Profiles.List() {
		if p.GroupMatch(id, c.groupAll) {
			if len(c.groupBindings) >= 4096 {
				for pid := range c.groupBindings {
					delete(c.groupBindings, pid)
					break
				}
			}
			if len(c.groupBindings) < 4096 {
				b := groupBinding{identity: id, profile: p.ID}
				if p.Trial != nil {
					b.trialPID = p.Trial.PID
					b.trialBirth = p.Trial.CreationTime
				}
				c.groupBindings[id.PID] = b
			}
			return p, true
		}
	}
	return direct, hasDirect
}
func (c *Controller) scopedDecision(p profiles.Profile, id wp.ProcessIdentity, t wp.Tuple, base wp.Decision) wp.Decision {
	network := ""
	if n := c.network.Load(); n != nil && n.Error == "" && time.Since(n.ObservedAt) <= 45*time.Second {
		network = n.Fingerprint
	}
	base.Flow.Network = network
	resolve := func(domain string) wp.Decision {
		if network != "" {
			n := c.network.Load()
			if n == nil || n.Fingerprint != network {
				d := base
				d.Block = true
				d.Deferred = false
				d.ResolveDomain = nil
				d.Policy = flowpolicy.Result{Action: flowpolicy.Block, Fallback: flowpolicy.FailBlock, Source: "network_changed", Reason: "network changed before upstream decision"}
				return d
			}
		}
		return c.evaluateFlow(p, id, t, base, domain, network)
	}
	out := resolve("")
	if t.Protocol == "tcp" && out.Policy.Source != "browser_dns" && (browserAuto(p) || flowpolicy.NeedsDomain(p.Rules, "tcp") || p.Mode == profiles.Auto && !c.hasBaseline(p, network) && c.AutoLearned.HasApplication(network, applicationKey(p), "tcp", time.Now())) {
		// The existing redirect accepts the app connection before any upstream is
		// opened. The closure captures a profile snapshot; edits affect new flows.
		out.Deferred = true
		out.ResolveDomain = resolve
	}
	if t.Destination.Addr().Is6() && (out.Deferred || out.ProbeDirect) {
		out.Block = true
		out.ProbeDirect = false
		out.Deferred = false
		out.ResolveDomain = nil
		out.Policy.Action = flowpolicy.Block
		out.Policy.Fallback = flowpolicy.FailBlock
		out.Policy.Source = "unsupported_ipv6_domain_policy"
		out.Policy.Reason = "IPv6 domain/adaptive UDP decision unsupported; no native bypass"
	}
	return out
}
func (c *Controller) evaluateFlow(p profiles.Profile, id wp.ProcessIdentity, t wp.Tuple, d wp.Decision, domain, network string) wp.Decision {
	if browserAuto(p) {
		return applyFlowResult(p, d, c.browserFlow(p, id, t, domain, network), network, c.Local)
	}
	result, matched := flowpolicy.Evaluate(p.Rules, flowpolicy.Input{Protocol: t.Protocol, Destination: t.Destination, Domain: domain, Network: network})
	if !matched && p.Mode == profiles.Auto && presetFor(p) != "openai" && presetFor(p) != "openai-web" && presetFor(p) != "discord" && domain != "" {
		target := domain
		if learned, ok := c.AutoLearned.Lookup(c.learnKey(p, network, t.Protocol, target, t.Destination.Port()), time.Now()); ok {
			result = learned
			matched = true
		}
	}
	if !matched && p.Mode == profiles.Auto && c.hasBaseline(p, network) && p.LocalStrategyID != "" && c.Local.HasApplication(id.Path) {
		for _, local := range c.Local.Config().User {
			if local.ID == p.LocalStrategyID && local.Enabled && strings.EqualFold(strings.ReplaceAll(local.Application, "\\", "/"), strings.ReplaceAll(id.Path, "\\", "/")) {
				result = flowpolicy.Result{Action: flowpolicy.DirectPlus, Strategy: local.ID, Fallback: flowpolicy.FailOpen, Reason: "adopted operator-confirmed configuration on this network; domain guards remain in native engine", Source: "configured_baseline"}
				matched = true
				break
			}
		}
	}
	if !matched && p.Mode == profiles.Auto {
		relay := p.RelayID
		if relay == "" {
			relay = "VPS-01"
		}
		switch presetFor(p) {
		case "openai", "openai-web":
			result = flowpolicy.Result{Action: flowpolicy.RelayRequired, RelayID: relay, Fallback: flowpolicy.Required, Reason: "OpenAI consistent egress", Source: "preset"}
			matched = true
		case "discord":
			result = flowpolicy.Result{Action: flowpolicy.RelayRequired, RelayID: relay, Fallback: flowpolicy.Required, Reason: "Discord control/update consistent egress", Source: "preset"}
			matched = true
			if t.Protocol == "udp" && t.Destination.Port() >= 1024 {
				result = flowpolicy.Result{Action: flowpolicy.Direct, RelayID: relay, Fallback: flowpolicy.FlowRelay, Reason: "Discord high-port UDP candidate; require bidirectional response", Source: "preset"}
			}
		}
	}
	if matched {
		d.Block = false
		d.Local = nil
		d.LocalStrategyID = ""
		d.Route = routing.Route{}
		d.Policy = result
		d.Network = network
		d.Group = p.ID
		switch result.Action {
		case flowpolicy.Block:
			d.Block = true
		case flowpolicy.Relay, flowpolicy.RelayRequired:
			d.Route = routing.Route{RelayID: result.RelayID}
		case flowpolicy.DirectPlus:
			d.Route = routing.Route{Kind: routing.DirectPlusRoute}
			d.Local = c.Local
			d.LocalStrategyID = result.Strategy
		}
		d.BlockOnFailure = result.Fallback == flowpolicy.Required || result.Fallback == flowpolicy.FailBlock
		d.ProbeDirect = result.Fallback == flowpolicy.FlowRelay
	} else {
		if p.Fallback != "" {
			d.BlockOnFailure = p.Fallback == flowpolicy.Required || p.Fallback == flowpolicy.FailBlock
		}
		d.Network = network
		d.Group = p.ID
		d.Policy = flowpolicy.Result{Action: flowpolicy.Action(d.Route.Type()), Strategy: d.LocalStrategyID, RelayID: d.Route.RelayID, Fallback: flowpolicy.FailOpen, Reason: "application default", Source: "profile"}
		if d.BlockOnFailure {
			d.Policy.Fallback = flowpolicy.Required
		}
	}
	// Only policies with explicit per-flow semantics bypass the legacy selector.
	if matched || p.Fallback != "" || len(p.Rules) > 0 {
		d.Flow.PinnedPolicy = &routing.Policy{Route: d.Route, Block: d.Block, BlockOnFailure: d.BlockOnFailure}
	}
	return d
}

func validGroupBinding(b groupBinding, p profiles.Profile) bool {
	if p.Trial == nil {
		return b.trialPID == 0
	}
	return p.Trial.PID == b.trialPID && p.Trial.CreationTime == b.trialBirth && time.Now().Before(p.Trial.Expires)
}

func (c *Controller) pruneTrials(now time.Time) {
	for _, p := range c.Profiles.List() {
		if p.Trial != nil && !now.Before(p.Trial.Expires) {
			c.Profiles.Delete(p.ID)
			c.mu.Lock()
			delete(c.localProfiles, p.ID)
			c.mu.Unlock()
			c.Engine.SetPolicy(p.ID, routing.Policy{})
		}
	}
}

func sameProfileExecutables(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for n := range a {
		if !strings.EqualFold(strings.ReplaceAll(a[n], "\\", "/"), strings.ReplaceAll(b[n], "\\", "/")) {
			return false
		}
	}
	return true
}
