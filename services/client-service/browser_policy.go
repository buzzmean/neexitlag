package clientservice

import (
	"path"
	"strings"
	"time"

	"github.com/buzzmean/neexitlag/core/flowpolicy"
	"github.com/buzzmean/neexitlag/core/localstrategy"
	"github.com/buzzmean/neexitlag/core/profiles"
	"github.com/buzzmean/neexitlag/core/routing"
	wp "github.com/buzzmean/neexitlag/platform/windows"
)

// Browser AUTO is an origin policy, never a process-wide optimizer winner. Trial
// profiles and explicit MANUAL/DIRECT/DIRECT+ retain their existing contracts.
func browserAuto(p profiles.Profile) bool {
	if nativeBrowserProfile(p) {
		return true
	}
	if edgeSitesProfile(p) {
		return p.Mode == profiles.Auto && p.Trial == nil
	}
	return p.Mode == profiles.Auto && p.Trial == nil && chromeProfile(p)
}
func chromeProfile(p profiles.Profile) bool {
	for _, exe := range p.Executables {
		if path.Base(strings.ToLower(strings.ReplaceAll(exe, `\`, "/"))) == "chrome.exe" {
			return true
		}
	}
	return false
}
func protectedBrowserDomain(domain string) bool {
	for _, suffix := range []string{"chatgpt.com", "openai.com", "oaistatic.com", "oaiusercontent.com", "chat.com"} {
		if flowpolicy.DomainMatches(domain, suffix) {
			return true
		}
	}
	return false
}
func youtubeDomain(domain string) bool {
	for _, preset := range localstrategy.Presets() {
		if preset.ID == "youtube" {
			for _, suffix := range preset.Domains {
				if flowpolicy.DomainMatches(domain, suffix) {
					return true
				}
			}
		}
	}
	return false
}
func (c *Controller) browserFlow(p profiles.Profile, id wp.ProcessIdentity, t wp.Tuple, domain, network string) flowpolicy.Result {
	relay := p.RelayID
	if relay == "" {
		relay = "VPS-01"
	}
	required := func(reason string) flowpolicy.Result {
		return flowpolicy.Result{Action: flowpolicy.RelayRequired, RelayID: relay, Fallback: flowpolicy.Required, Source: "browser_domain", Reason: reason}
	}
	// Browser resolver bootstrap is not an origin data flow. Sending Chrome's
	// DNS to a private LAN resolver through the VPS breaks even ordinary sites.
	// Keep existing system/browser DNS semantics; this grants no TCP/443 bypass.
	if t.Destination.Port() == 53 && (t.Protocol == "udp" || t.Protocol == "tcp") {
		return flowpolicy.Result{Action: flowpolicy.Direct, Fallback: flowpolicy.FailOpen, Source: "browser_dns", Reason: "resolver bootstrap: preserve browser/system DNS path"}
	}
	// Opaque is never classified from a DNS/IP guess or a request on another
	// connection. Safe transport pending richer origin evidence is authenticated
	// relay, including IPv6/QUIC which cannot use IPv4 ClientHello interception.
	if domain == "" {
		return required("domain unavailable: opaque browser transport; no DIRECT authorization")
	}
	if protectedBrowserDomain(domain) {
		return required("protected OpenAI origin: authenticated consistent egress")
	}
	if r, ok := flowpolicy.Evaluate(p.Rules, flowpolicy.Input{Protocol: t.Protocol, Domain: domain, Network: network, Destination: t.Destination}); ok {
		return r
	}
	if youtubeDomain(domain) && c.browserLocalAllowed(network) && c.hasBaseline(p, network) && p.LocalStrategyID != "" {
		for _, local := range c.Local.Config().User {
			if local.Enabled && local.ID == p.LocalStrategyID && strings.EqualFold(strings.ReplaceAll(local.Application, `\`, "/"), strings.ReplaceAll(id.Path, `\`, "/")) {
				return flowpolicy.Result{Action: flowpolicy.DirectPlus, Strategy: local.ID, Fallback: flowpolicy.FailOpen, Source: "configured_baseline", Reason: "YouTube origin only: verified local configuration on this network"}
			}
		}
	}
	if network != "" {
		if learned, ok := c.AutoLearned.Lookup(c.learnKey(p, network, t.Protocol, domain, t.Destination.Port()), time.Now()); ok {
			return learned
		}
	}
	return flowpolicy.Result{Action: flowpolicy.Direct, Fallback: flowpolicy.FailOpen, Source: "browser_domain", Reason: "ordinary origin: DIRECT first; TCP ACK is not application success"}
}
func applyFlowResult(p profiles.Profile, d wp.Decision, r flowpolicy.Result, network string, local *localstrategy.LocalStrategyEngine) wp.Decision {
	d.Block, d.ProbeDirect = false, false
	d.Local, d.LocalStrategyID = nil, ""
	d.Route = routing.Route{}
	d.Policy, d.Network, d.Group = r, network, p.ID
	switch r.Action {
	case flowpolicy.Block:
		d.Block = true
	case flowpolicy.Relay, flowpolicy.RelayRequired:
		d.Route = routing.Route{RelayID: r.RelayID}
	case flowpolicy.DirectPlus:
		d.Route = routing.Route{Kind: routing.DirectPlusRoute}
		d.Local, d.LocalStrategyID = local, r.Strategy
	}
	d.BlockOnFailure = r.Fallback == flowpolicy.Required || r.Fallback == flowpolicy.FailBlock
	d.ProbeDirect = r.Fallback == flowpolicy.FlowRelay
	d.Flow.PinnedPolicy = &routing.Policy{Route: d.Route, Block: d.Block, BlockOnFailure: d.BlockOnFailure}
	return d
}
