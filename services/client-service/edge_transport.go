package clientservice

// Edge has its own listener, PAC, readiness and recovery state. Only the existing
// immutable per-origin planner/open path is shared with Chrome.
import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/buzzmean/neexitlag/core/browserproxy"
	"github.com/buzzmean/neexitlag/core/flowpolicy"
	"github.com/buzzmean/neexitlag/core/profiles"
	"github.com/buzzmean/neexitlag/core/security"
	wp "github.com/buzzmean/neexitlag/platform/windows"
)

type edgeSaved struct {
	Address     string   `json:"address"`
	ExpectedPAC string   `json:"expected_pac"`
	Recovery    string   `json:"recovery_address"`
	Ready       []string `json:"ready_profiles"`
	Owned       []string `json:"owned_policy"`
}
type edgeRuntime struct {
	lastError string
	saved     edgeSaved
	server    *browserproxy.Server
	// Test embedders may provide an isolated policy store. Production uses HKU.
	write  func(string, []string, string) error
	active func(string, string) bool
}

func edgeExecutable(exe string) bool {
	return strings.EqualFold(filepath.Base(strings.ReplaceAll(exe, `\`, "/")), "msedge.exe")
}
func edgeSitesProfile(p profiles.Profile) bool { return edgeProfile(p) && p.AppPreset == "edge-sites" }
func isEdgeMutation(method string) bool {
	return method == "edge.prepare" || method == "edge.confirm" || method == "edge.policy"
}
func edgeProfile(p profiles.Profile) bool {
	if len(p.Executables) == 0 {
		return false
	}
	for _, exe := range p.Executables {
		if !edgeExecutable(exe) {
			return false
		}
	}
	return true
}
func (c *Controller) edgeLocked() *edgeRuntime {
	if c.browser.edge == nil {
		c.browser.edge = &edgeRuntime{}
	}
	return c.browser.edge
}
func (e *edgeRuntime) policyActive(sid string) bool {
	if e.active != nil {
		return e.active(sid, e.saved.ExpectedPAC)
	}
	return wp.EdgeProxyPolicyActive(sid, e.saved.ExpectedPAC)
}
func (c *Controller) edgeOwner(ctx context.Context, source, local netip.AddrPort) (browserproxy.Owner, error) {
	id, err := wp.LookupTCPPeer(source, local)
	if err != nil {
		return browserproxy.Owner{}, browserproxy.ErrOwnerLookup
	}
	sid, err := wp.ProcessUserSID(id)
	if err != nil {
		return browserproxy.Owner{}, browserproxy.ErrOwnerToken
	}
	if c.BrowserOwnerSID == "" || sid != c.BrowserOwnerSID {
		return browserproxy.Owner{}, browserproxy.ErrOwnerSID
	}
	p, ok := c.matchApplication(id)
	if !ok || !edgeProfile(p) || p.Trial != nil {
		return browserproxy.Owner{}, browserproxy.ErrOwnerProfile
	}
	return browserproxy.Owner{PID: id.PID, Birth: id.CreationTime, Exe: id.Path, Profile: p.ID}, nil
}
func (c *Controller) edgeOwnerProfile(owner browserproxy.Owner) (profilesResult, error) {
	id := wp.ProcessIdentity{PID: owner.PID, CreationTime: owner.Birth, Path: owner.Exe}
	if !wp.IsCurrent(id) {
		return profilesResult{}, errors.New("Edge owner exited/reused")
	}
	p, ok := c.matchApplication(id)
	if !ok || !edgeProfile(p) || p.Trial != nil || p.ID != owner.Profile || (p.Mode != profiles.Auto && p.Mode != profiles.Direct) {
		return profilesResult{}, errors.New("Edge AUTO profile changed")
	}
	network := ""
	if n := c.network.Load(); n != nil && n.Error == "" && time.Since(n.ObservedAt) <= 45*time.Second {
		network = n.Fingerprint
	}
	c.mu.Lock()
	enabled := c.enabled && c.settings.Automatic && p.Mode == profiles.Auto && edgeSitesProfile(p)
	c.mu.Unlock()
	return profilesResult{profile: p, id: id, network: network, enabled: enabled}, nil
}
func (c *Controller) edgeRecoveryOwner(ctx context.Context, source, local, original, listener netip.AddrPort) (browserproxy.Owner, error) {
	helper, err := wp.LookupTCPPeer(source, local)
	if err != nil {
		return browserproxy.Owner{}, browserproxy.ErrOwnerLookup
	}
	program, err := os.Executable()
	if c.BrowserRecoveryExecutable != "" {
		program = c.BrowserRecoveryExecutable
	}
	if err != nil || !strings.EqualFold(helper.Path, program) {
		return browserproxy.Owner{}, browserproxy.ErrOwnerProfile
	}
	sid, err := wp.ProcessUserSID(helper)
	if err != nil || sid != c.BrowserOwnerSID {
		return browserproxy.Owner{}, browserproxy.ErrOwnerSID
	}
	return c.edgeOwner(ctx, original, listener)
}
func (c *Controller) listenEdge(e *edgeRuntime) error {
	if e.server != nil {
		return nil
	}
	server, err := browserproxy.Listen(c.ctx, browserproxy.Config{Address: e.saved.Address, Authorize: c.edgeOwner, AuthorizeForwarded: c.edgeRecoveryOwner, Plan: c.browserPlan, Open: c.browserOpen, TransformApplied: c.browserTransform, OnStall: c.browserStall, OnDirectAnswered: c.browserDirectAnswered})
	if err != nil {
		return err
	}
	e.server = server
	e.saved.Address = server.Address()
	return nil
}
func edgeGuard(p profiles.Profile) ([]string, bool) {
	domains := []string{}
	strict := false
	for _, r := range append(append([]flowpolicy.Rule{}, p.Rules...), p.BrowserRecoveryRules...) {
		if r.Action == flowpolicy.Direct {
			continue
		}
		if len(r.Domains) == 0 {
			strict = true
		} else {
			domains = append(domains, r.Domains...)
		}
	}
	return domains, strict
}
func edgeClosedPAC(p profiles.Profile) string {
	domains, strict := edgeGuard(p)
	domains = append(domains, "chatgpt.com", "openai.com", "oaistatic.com", "oaiusercontent.com", "chat.com", "youtube.com", "youtu.be", "googlevideo.com", "ytimg.com", "youtube-nocookie.com", "ggpht.com")
	raw, _ := json.Marshal(domains)
	script := "var guarded=" + string(raw) + ";var strict="
	if strict {
		script += "true"
	} else {
		script += "false"
	}
	script += `;function FindProxyForURL(url,host){host=host.toLowerCase().replace(/\.$/,'');for(var i=0;i<guarded.length;i++){var d=guarded[i].toLowerCase().replace(/\.$/,'');if(host===d||host.slice(-(d.length+1))==='.'+d)return 'PROXY fail-closed.neexitlag.invalid:9';}return strict?'PROXY fail-closed.neexitlag.invalid:9':'DIRECT';}`
	return "data:application/x-ns-proxy-autoconfig;base64," + base64.StdEncoding.EncodeToString([]byte(script))
}
func edgePolicyValue(pac string) string {
	raw, _ := json.Marshal(map[string]any{"ProxyMode": "pac_script", "ProxyPacUrl": pac, "ProxyPacMandatory": true})
	return string(raw)
}
func (c *Controller) saveEdge(e *edgeRuntime) error {
	if c.customPath == "" {
		return nil
	}
	if e.server != nil {
		e.saved.Ready = e.server.Snapshot().ReadyProfiles
	}
	raw, err := json.Marshal(e.saved)
	if err != nil {
		return err
	}
	dir := filepath.Dir(c.customPath)
	if err = os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".edge-transport-*")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if err = f.Chmod(0600); err == nil {
		_, err = f.Write(raw)
	}
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(name, filepath.Join(dir, "edge-transport.json"))
}
func (c *Controller) writeEdgePolicy(e *edgeRuntime, pac string) error {
	next := edgePolicyValue(pac)
	// Journal both authorized values before touching HKU: crash between the two
	// writes can safely resume, but a foreign value is never overwritten.
	old := append([]string(nil), e.saved.Owned...)
	if len(old) > 0 {
		old = old[len(old)-1:]
	}
	e.saved.Owned = append(old, next)
	if err := c.saveEdge(e); err != nil {
		return err
	}
	write := e.write
	if write == nil {
		write = wp.EdgeProxyPolicyWrite
	}
	return write(c.BrowserOwnerSID, e.saved.Owned, next)
}
func (c *Controller) handleEdge(method string, raw json.RawMessage) (any, error) {
	c.browser.mu.Lock()
	defer c.browser.mu.Unlock()
	e := c.edgeLocked()
	if method == "edge.status" {
		snapshot := browserproxy.Snapshot{ReadyProfiles: []string{}, Failures: map[string]uint64{}, Records: []browserproxy.Record{}}
		if e.server != nil {
			snapshot = e.server.Snapshot()
		}
		snapshot.RecoveryAddress = e.saved.Recovery
		snapshot.PolicyReady = e.policyActive(c.BrowserOwnerSID)
		for {
			b, _ := json.Marshal(snapshot)
			if len(b) < 58000 || len(snapshot.Records) == 0 {
				break
			}
			snapshot.Truncated = true
			snapshot.Records = snapshot.Records[:len(snapshot.Records)-1]
		}
		return struct {
			browserproxy.Snapshot
			LastError string `json:"last_error,omitempty"`
		}{snapshot, e.lastError}, nil
	}
	if c.BrowserOwnerSID == "" {
		return nil, errors.New("Edge requires installation owner SID")
	}
	var q struct {
		Profile   string `json:"profile_id"`
		Directory string `json:"directory"`
		Recovery  string `json:"recovery_address"`
		Health    string `json:"health_url"`
		Operation string `json:"operation"`
	}
	if err := security.Parse(raw, &q); err != nil {
		return nil, err
	}
	var p profiles.Profile
	found := false
	for _, candidate := range c.Profiles.List() {
		if candidate.ID == q.Profile && edgeProfile(candidate) && candidate.Trial == nil {
			p = candidate
			found = true
			break
		}
	}
	if !found {
		return nil, errors.New("Edge profile required")
	}
	switch method {
	case "edge.prepare":
		if (p.Mode != profiles.Auto && p.Mode != profiles.Direct) || q.Directory == "" || len(q.Directory) > 2048 {
			return nil, errors.New("Edge DIRECT/AUTO profile and persistent directory required")
		}
		if _, err := netip.ParseAddrPort(q.Recovery); err != nil {
			return nil, errors.New("owner-scoped Edge recovery endpoint required")
		}
		if err := c.listenEdge(e); err != nil {
			return nil, err
		}
		domains, strict := edgeGuard(p)
		args, err := browserproxy.ChromeArgumentsWithRecovery(e.server.Address(), q.Directory, domains, strict, q.Recovery)
		if err != nil {
			return nil, err
		}
		health, err := e.server.Challenge(p.ID)
		if err != nil {
			return nil, err
		}
		e.saved.ExpectedPAC = strings.TrimPrefix(args[1], "--proxy-pac-url=")
		e.saved.Recovery = q.Recovery
		if err = c.saveEdge(e); err != nil {
			return nil, err
		}
		program, _ := os.Executable()
		return map[string]any{"arguments": args, "recovery_arguments": args, "address": e.server.Address(), "health_url": health, "recovery_program": program, "recovery_settings": browserproxy.RecoverySettings{GuardFile: c.browserGuardFile(), Primary: e.server.Address(), Profile: p.ID, Domains: domains, Strict: strict}, "scope": "Edge only; Chrome and Windows proxy unchanged"}, nil
	case "edge.policy":
		pac := e.saved.ExpectedPAC
		switch q.Operation {
		case "guard":
			pac = edgeClosedPAC(p)
		case "install":
			if pac == "" || e.server == nil || e.saved.Recovery == "" {
				return nil, errors.New("prepare Edge before installing policy")
			}
		default:
			return nil, errors.New("unsupported Edge policy operation")
		}
		if err := c.writeEdgePolicy(e, pac); err != nil {
			return nil, err
		}
		if q.Operation == "guard" && e.server != nil {
			e.server.Forget(p.ID)
			e.server.CloseConnections()
			if err := c.saveEdge(e); err != nil {
				return nil, err
			}
		}
		return map[string]bool{"applied": true}, nil
	case "edge.confirm":
		if e.server == nil || e.saved.Recovery == "" || !e.policyActive(c.BrowserOwnerSID) || !e.server.HealthVerified(p.ID, q.Health) {
			return nil, errors.New("Edge PAC readiness not observed: policy, owner/SID or nonce")
		}
		if !e.server.AuthorizeProfile(p.ID, q.Health) {
			return nil, errors.New("Edge activation proof expired/changed")
		}
		if err := c.saveEdge(e); err != nil {
			return nil, err
		}
		return map[string]bool{"ready": true}, nil
	}
	return nil, errors.New("unsupported Edge IPC")
}
func (c *Controller) edgeReadyLocked(p profiles.Profile) bool {
	e := c.browser.edge
	return e != nil && e.server != nil && e.saved.Recovery != "" && e.server.Confirmed(p.ID) && e.policyActive(c.BrowserOwnerSID)
}
func (c *Controller) requireEdgeActivation() error {
	if c.BrowserOwnerSID == "" {
		return nil
	}
	c.browser.mu.Lock()
	defer c.browser.mu.Unlock()
	for _, p := range c.Profiles.List() {
		if edgeSitesProfile(p) && p.Mode == profiles.Auto && p.Trial == nil && !c.edgeReadyLocked(p) {
			return errors.New("Edge AUTO requires verified Edge CONNECT/PAC; capture unchanged")
		}
	}
	return nil
}
func (c *Controller) closeEdgeConnections() {
	c.browser.mu.Lock()
	var server *browserproxy.Server
	if e := c.browser.edge; e != nil {
		server = e.server
	}
	c.browser.mu.Unlock()
	if server != nil {
		server.CloseConnections()
	}
}
func (c *Controller) closeEdge() {
	c.browser.mu.Lock()
	var server *browserproxy.Server
	if e := c.browser.edge; e != nil {
		server = e.server
	}
	c.browser.mu.Unlock()
	if server != nil {
		server.Close()
	}
}
func (c *Controller) loadEdgeTransport() error {
	if c.customPath == "" {
		return nil
	}
	f, err := os.Open(filepath.Join(filepath.Dir(c.customPath), "edge-transport.json"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, 262145))
	if err != nil || len(raw) > 262144 {
		return errors.New("invalid saved Edge state")
	}
	var saved edgeSaved
	if err = security.Parse(raw, &saved); err != nil {
		return err
	}
	addr, err := netip.ParseAddrPort(saved.Address)
	if err != nil || addr.Addr() != netip.MustParseAddr("127.0.0.1") || addr.Port() == 0 || len(saved.Ready) > 128 || len(saved.Owned) > 2 || len(saved.ExpectedPAC) > 128*1024 {
		return errors.New("invalid saved Edge endpoint/state")
	}
	c.browser.mu.Lock()
	defer c.browser.mu.Unlock()
	e := c.edgeLocked()
	e.saved = saved
	if err = c.listenEdge(e); err != nil {
		return err
	}
	e.server.RestoreConfirmed(saved.Ready)
	return nil
}

func (c *Controller) forgetEdgeProfile(id string) {
	c.browser.mu.Lock()
	defer c.browser.mu.Unlock()
	if e := c.browser.edge; e != nil && e.server != nil {
		e.server.Forget(id)
		_ = c.saveEdge(e)
	}
}

// A broken Edge port/state must not prevent the existing Chrome listener from
// restoring. Edge stays unready/fail-closed and the GUI can repair only that lane.
func (c *Controller) restoreEdgeTransport() {
	if err := c.loadEdgeTransport(); err != nil {
		c.browser.mu.Lock()
		c.edgeLocked().lastError = err.Error()
		c.browser.mu.Unlock()
	}
}
