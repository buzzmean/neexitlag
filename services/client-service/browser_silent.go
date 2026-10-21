package clientservice

// Silent PAC configuration is not proof that an already-running browser read
// policy. Owner authorization on every accepted socket remains mandatory.
import (
	"encoding/json"
	"errors"
	"slices"
	"time"

	"github.com/buzzmean/neexitlag/core/browserproxy"
	"github.com/buzzmean/neexitlag/core/security"
	wp "github.com/buzzmean/neexitlag/platform/windows"
)

type browserSilentState struct {
	Profiles  []string  `json:"profiles"`
	Owned     []string  `json:"owned,omitempty"`
	AppliedAt time.Time `json:"applied_at"`
}
type browserSilentView struct {
	ConfiguredProfiles []string  `json:"configured_profiles"`
	State              string    `json:"state"`
	AppliedAt          time.Time `json:"applied_at"`
	FirstRequestAt     time.Time `json:"first_request_at"`
	ApplyToRequestMS   int64     `json:"apply_to_request_ms"`
	ObservedRequests   int       `json:"observed_requests"`
	ExistingTCP        string    `json:"existing_tcp"`
}

func (c *Controller) chromePolicyActive() bool {
	if c.browser.chromeActive != nil {
		return c.browser.chromeActive(c.BrowserOwnerSID, c.browser.expectedPAC)
	}
	return wp.ChromeProxyPolicyActive(c.BrowserOwnerSID, c.browser.expectedPAC)
}
func (c *Controller) silentConfiguredLocked(id string, edge bool) bool {
	s := c.browser.silent
	if edge {
		if c.browser.edge == nil {
			return false
		}
		s = c.browser.edge.saved.Silent
	}
	return slices.Contains(s.Profiles, id)
}
func (c *Controller) silentViewLocked(edge bool, snapshot browserproxy.Snapshot) browserSilentView {
	s := c.browser.silent
	if edge && c.browser.edge != nil {
		s = c.browser.edge.saved.Silent
	}
	v := browserSilentView{ConfiguredProfiles: append([]string{}, s.Profiles...), State: "NOT_CONFIGURED", AppliedAt: s.AppliedAt, ApplyToRequestMS: -1, ExistingTCP: "PINNED_UNTIL_CLOSED"}
	if len(s.Profiles) == 0 {
		return v
	}
	v.State = "POLICY_PENDING_BROWSER_REFRESH"
	active := c.chromePolicyActive()
	if edge {
		active = c.edgeLocked().policyActive(c.BrowserOwnerSID)
	}
	if !active || !snapshot.Listening {
		v.State = "POLICY_OR_LISTENER_UNAVAILABLE"
		return v
	}
	for _, r := range snapshot.Records {
		if !slices.Contains(s.Profiles, r.Owner.Profile) || r.Started.Before(s.AppliedAt) {
			continue
		}
		// Authorization succeeded on this exact real browser socket. This is policy
		// uptake/ingress evidence, NOT a page/TLS success or a fresh nonce proof.
		v.ObservedRequests++
		if v.FirstRequestAt.IsZero() || r.Started.Before(v.FirstRequestAt) {
			v.FirstRequestAt = r.Started
		}
	}
	if !v.FirstRequestAt.IsZero() {
		v.State = "PAC_REQUEST_OBSERVED"
		v.ApplyToRequestMS = v.FirstRequestAt.Sub(s.AppliedAt).Milliseconds()
	}
	return v
}
func (c *Controller) handleBrowserSilent(method string, raw json.RawMessage) (any, error) {
	var q struct {
		Profile string `json:"profile_id"`
	}
	if err := security.Parse(raw, &q); err != nil {
		return nil, err
	}
	edge := method == "edge.silent.apply"
	c.mu.Lock()
	automatic := c.settings.Automatic
	c.mu.Unlock()
	if method != "browser.silent.guard" && !automatic {
		return nil, errors.New("BROWSER_AUTOMATIC_DISABLED")
	}
	c.browser.mu.Lock()
	defer c.browser.mu.Unlock()
	if c.BrowserOwnerSID == "" {
		return nil, errors.New("BROWSER_OWNER_SID_REQUIRED")
	}
	found := false
	for _, p := range c.Profiles.List() {
		if p.ID == q.Profile && p.Trial == nil && ((edge && edgeProfile(p)) || (!edge && chromeProfile(p))) {
			found = true
		}
	}
	if !found {
		return nil, errors.New("BROWSER_PROFILE_REQUIRED")
	}
	if method == "browser.silent.guard" {
		previous := c.browser.expectedPAC
		for _, p := range c.Profiles.List() {
			if p.ID == q.Profile {
				c.browser.expectedPAC = edgeClosedPAC(p)
			}
		}
		if err := c.reapplySilentLocked(false); err != nil {
			c.browser.expectedPAC = previous
			return nil, err
		}
		c.browser.silent.Profiles = nil
		if err := c.saveBrowserAddress(); err != nil {
			return nil, err
		}
		return map[string]bool{"applied": true, "guard": true}, nil
	}
	s := &c.browser.silent
	server := c.browser.server
	pac := c.browser.expectedPAC
	recovery := c.browser.recoveryAddress
	if edge {
		e := c.edgeLocked()
		s = &e.saved.Silent
		server = e.server
		pac = e.saved.ExpectedPAC
		recovery = e.saved.Recovery
	}
	if server == nil || !server.Snapshot().Listening || pac == "" || recovery == "" {
		return nil, errors.New("BROWSER_PREPARE_REQUIRED")
	}
	if !slices.Contains(s.Profiles, q.Profile) {
		if len(s.Profiles) >= 128 {
			return nil, errors.New("BROWSER_PROFILE_CAPACITY")
		}
		s.Profiles = append(s.Profiles, q.Profile)
	}
	if err := c.reapplySilentLocked(edge); err != nil {
		// Keep the previous mandatory PAC; never remove it or authorize DIRECT on a
		// failed write. Registry writer refuses foreign values before mutation.
		if edge {
			c.edgeLocked().lastError = "edge.silent.apply: EDGE_POLICY_FAILED"
		} else {
			c.browser.lastError = "browser.silent.apply: CHROME_POLICY_FAILED"
		}
		return nil, err
	}
	return c.silentViewLocked(edge, server.Snapshot()), nil
}

// Also called on every service boot, after restoring the exact loopback port.
// A missing owner hive is retried after logon; machine-wide policy is never set.
func (c *Controller) reapplySilentLocked(edge bool) error {
	if edge {
		e := c.edgeLocked()
		if e.server == nil || !e.server.Snapshot().Listening {
			return errors.New("EDGE_LISTENER_UNAVAILABLE")
		}
		if err := c.writeEdgePolicy(e, e.saved.ExpectedPAC); err != nil {
			return err
		}
		e.lastError = ""
		e.readinessFailure = ""
		e.saved.Silent.AppliedAt = time.Now()
		return c.saveEdge(e)
	}
	if c.browser.server == nil || !c.browser.server.Snapshot().Listening {
		return errors.New("CHROME_LISTENER_UNAVAILABLE")
	}
	s := &c.browser.silent
	next := edgePolicyValue(c.browser.expectedPAC)
	old := append([]string{}, s.Owned...)
	if len(old) > 1 {
		old = old[len(old)-1:]
	}
	s.Owned = append(old, next)
	if err := c.saveBrowserAddress(); err != nil {
		s.Owned = old
		return err
	}
	write := c.browser.chromeWrite
	if write == nil {
		write = wp.ChromeProxyPolicyWrite
	}
	if err := write(c.BrowserOwnerSID, s.Owned, next); err != nil {
		s.Owned = old
		_ = c.saveBrowserAddress()
		return err
	}
	s.AppliedAt = time.Now()
	c.browser.lastError = ""
	return c.saveBrowserAddress()
}

// Recheck after owner logon (HKU may not be mounted at SCM boot). No HKLM,
// browser process manipulation, or foreign-policy overwrite. Serialized with IPC.
func (c *Controller) RetrySilentBrowserPolicies() {
	if !c.activationNetworkAllowed() {
		_ = c.CleanupPreactivationBrowserPolicies()
		return
	}

	if c.ctx.Err() != nil {
		return
	}
	c.activationMu.Lock()
	defer c.activationMu.Unlock()
	c.browser.mu.Lock()
	defer c.browser.mu.Unlock()
	if len(c.browser.silent.Profiles) > 0 && !c.chromePolicyActive() {
		if err := c.reapplySilentLocked(false); err != nil {
			c.browser.lastError = "browser.silent.apply: CHROME_POLICY_FAILED"
		}
	}
	if e := c.browser.edge; e != nil && len(e.saved.Silent.Profiles) > 0 && !e.policyActive(c.BrowserOwnerSID) {
		if err := c.reapplySilentLocked(true); err != nil {
			e.lastError = "edge.silent.apply: EDGE_POLICY_FAILED"
		}
	}
}

func (c *Controller) maintainSilentBrowserPolicies() {
	tick := time.NewTicker(5 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-tick.C:
			c.RetrySilentBrowserPolicies()
		}
	}
}
