package clientservice

import (
	"errors"
	"strings"

	"github.com/buzzmean/neexitlag/core/browserproxy"
	"github.com/buzzmean/neexitlag/core/profiles"
	wp "github.com/buzzmean/neexitlag/platform/windows"
)

type nativeBrowserCompatibility struct {
	OwnedPolicy   bool     `json:"owned_policy"`
	PolicyChanged bool     `json:"policy_changed"`
	Arguments     []string `json:"arguments"`
	Address       string   `json:"address"`
	Verification  string   `json:"verification"`
}

// Keep legacy PAC users on a real listener. A fresh native installation does
// not install PAC. This is compatibility, never a synthetic PAC readiness proof.
func (c *Controller) nativeBrowserCompatibility(p profiles.Profile, requested bool) (*nativeBrowserCompatibility, error) {
	c.browser.mu.Lock()
	defer c.browser.mu.Unlock()
	if edgeProfile(p) {
		e := c.edgeLocked()
		if len(e.saved.Owned) == 0 {
			return nil, nil
		}
		if !e.policyActive(c.BrowserOwnerSID) && !e.guardActive(c.BrowserOwnerSID) {
			return nil, errors.New("EDGE_POLICY_OWNERSHIP_CHANGED")
		}
		if err := c.listenEdge(e); err != nil {
			return nil, errors.New("EDGE_LISTENER_BIND_FAILED")
		}
		args, err := nativeEdgeCompatibilityArguments(p, e)
		if err != nil {
			return nil, err
		}
		pac := strings.TrimPrefix(args[1], "--proxy-pac-url=")
		if e.saved.ExpectedPAC != "" && e.policyActive(c.BrowserOwnerSID) {
			pac = e.saved.ExpectedPAC
		}
		changed := !e.activePolicyValue(c.BrowserOwnerSID, pac)
		if changed {
			if err := c.writeEdgePolicy(e, pac); err != nil {
				return nil, err
			}
		}
		e.saved.ExpectedPAC = pac
		e.saved.GuardPAC = pac
		if err := c.saveEdge(e); err != nil {
			return nil, err
		}
		return &nativeBrowserCompatibility{OwnedPolicy: true, PolicyChanged: changed, Address: e.server.Address(), Verification: "LIVE_LISTENER; running browser policy refresh NOT_OBSERVED"}, nil
	}
	// Chrome's owner-user journal/recovery remains unchanged. Replacing its PAC
	// here would invalidate that journal and could disable the independent guard.
	if !wp.ChromeProxyPolicyActive(c.BrowserOwnerSID, c.browser.expectedPAC) {
		return nil, errors.New("NATIVE_OWNED_PAC_GUARD_REQUIRED")
	}
	if c.browser.server == nil {
		server, err := browserproxy.Listen(c.ctx, browserproxy.Config{TransportCheck: c.transportCheckSettings, RequireClientHelloReplay: true, Address: c.browser.address, Authorize: c.browserOwner, AuthorizeForwarded: c.browserRecoveryOwner, Plan: c.browserPlan, Open: c.browserOpen, OpenPlan: c.browserOpenPlan, TransformApplied: c.browserTransform, OnStall: c.browserStall, OnDirectAnswered: c.browserDirectAnswered})
		if err != nil {
			return nil, errors.New("CHROME_LISTENER_BIND_FAILED")
		}
		if server.Address() != c.browser.address {
			server.Close()
			return nil, errors.New("NATIVE_GUARD_ADDRESS_CHANGED")
		}
		c.browser.server = server
	}
	return &nativeBrowserCompatibility{OwnedPolicy: true, Address: c.browser.server.Address(), Verification: "OWNED_PAC_PRESERVED; live browser policy uptake NOT_OBSERVED"}, nil
}
func nativeCompatibilityArguments(p profiles.Profile, address string) ([]string, error) {
	domains, strict := edgeGuard(p)
	return browserproxy.ChromeArgumentsGuarded(address, "native-compatibility-only", domains, strict)
}
func (e *edgeRuntime) activePolicyValue(sid, pac string) bool {
	if e.active != nil {
		return e.active(sid, pac)
	}
	copy := *e
	copy.saved.ExpectedPAC = pac
	return copy.policyActive(sid)
}

// Called under browser.mu. Only an explicitly armed native profile may recover
// protected requests through this listener without legacy PAC nonce admission.
func (c *Controller) edgeRollbackPAC(e *edgeRuntime, p profiles.Profile) (string, error) {
	if !nativeBrowserProfile(p) {
		return edgeClosedPAC(p), nil
	}
	if !c.nativeBackendReady() {
		return edgeClosedPAC(p), nil
	}
	if err := c.listenEdge(e); err != nil {
		return edgeClosedPAC(p), nil
	}
	args, err := nativeEdgeCompatibilityArguments(p, e)
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(args[1], "--proxy-pac-url="), nil
}

// This gate cannot be bypassed by profile.set or an authenticated but stale GUI.
// Only an existing service-owned policy for the installation SID is accepted.
// Registry presence does NOT certify uptake by an already-running browser.
func (c *Controller) nativeGuardReady(p profiles.Profile) error {
	c.browser.mu.Lock()
	defer c.browser.mu.Unlock()
	return c.nativeGuardReadyLocked(p)
}
func (c *Controller) nativeGuardReadyLocked(p profiles.Profile) error {
	if c.BrowserOwnerSID == "" {
		return errors.New("NATIVE_OWNED_PAC_GUARD_REQUIRED")
	}
	if edgeProfile(p) {
		e := c.edgeLocked()
		if len(e.saved.Owned) > 0 && (e.policyActive(c.BrowserOwnerSID) || e.guardActive(c.BrowserOwnerSID)) {
			return nil
		}
	} else if c.browser.expectedPAC != "" && wp.ChromeProxyPolicyActive(c.BrowserOwnerSID, c.browser.expectedPAC) {
		return nil
	}
	return errors.New("NATIVE_OWNED_PAC_GUARD_REQUIRED")
}

func nativeEdgeCompatibilityArguments(p profiles.Profile, e *edgeRuntime) ([]string, error) {
	domains, strict := edgeGuard(p)
	if e.saved.Recovery != "" {
		return browserproxy.ChromeArgumentsWithRecovery(e.server.Address(), "native-compatibility-only", domains, strict, e.saved.Recovery)
	}
	return nativeCompatibilityArguments(p, e.server.Address())
}
