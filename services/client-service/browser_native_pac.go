package clientservice

import (
	"context"
	"errors"
	"net"
	"net/netip"

	"github.com/buzzmean/neexitlag/core/browserproxy"
	wp "github.com/buzzmean/neexitlag/platform/windows"
)

// A live online PAC owns the browser's loopback CONNECT, so no remote browser
// SYN exists to intercept. Delegate exactly its new, service-owned upstream to
// the existing native worker. PAC ingress and WinDivert upstream are distinct
// evidence, not two independent application connections or a native-only claim.
func (c *Controller) openNativePAC(ctx context.Context, q browserproxy.Request, plan browserproxy.Plan, d wp.Decision) (net.Conn, error) {
	backend, ok := c.Backend.(interface {
		OpenBrowserNativeTCP(context.Context, wp.ProcessIdentity, netip.AddrPort, netip.AddrPort, wp.Decision) (net.Conn, error)
	})
	if !ok {
		return nil, errors.New("NATIVE_PAC_DELEGATION_UNAVAILABLE")
	}
	if !c.nativeBackendReady() {
		return nil, errors.New("NATIVE_CAPTURE_NOT_READY")
	}
	d.OriginDomain = q.Host
	d.NewConnectionsOnly = true
	d.BlockOnFailure = true
	d.Deferred = false
	d.ResolveDomain = nil
	requestContext := ctx
	d.OpenDeferred = func(ctx context.Context, decision wp.Decision, _ string) (net.Conn, wp.Decision, error) {
		// The exact plan is known from CONNECT: open it before acknowledging
		// the reserved SYN, so PAC never reports OPEN before the actual route.
		work, stop := context.WithCancel(ctx)
		stopCaller := context.AfterFunc(requestContext, stop)
		defer stopCaller()
		defer stop()
		ctx = work
		// CONNECT authority is owner-validated origin metadata, including ECH. Do not
		// replace it with the TLS outer SNI or DNS association observed downstream.
		q.Ingress = "PAC_NATIVE_UPSTREAM"
		current, err := c.browserProfile(q.Owner)
		if err != nil || !current.enabled || !nativeBrowserProfile(current.profile) || current.network != plan.Network {
			return nil, decision, errors.New("NATIVE_PROFILE_CHANGED")
		}
		sid, err := wp.ProcessUserSID(current.id)
		if err != nil || sid != c.BrowserOwnerSID {
			return nil, decision, errors.New("NATIVE_OWNER_SID_MISMATCH")
		}
		n := &c.browser.native
		n.activate(q.Owner.Profile, false)
		record, err := n.begin(q, wp.Tuple{Protocol: "tcp", Source: q.Source, Destination: plan.Destination})
		if err != nil {
			return nil, decision, err
		}
		// Ingress marker prevents recursive delegation. This is the only real
		// authenticated OPEN; planning/probes are not replayed inside the worker.
		conn, err := c.browserOpen(ctx, q, plan)
		decision = applyFlowResult(current.profile, decision, plan.Policy, current.network, c.browserLocal(plan.Policy.Strategy))
		decision.BlockOnFailure = true
		n.opened(record, plan, err)
		if err != nil {
			return nil, decision, err
		}
		return &nativeBrowserConn{Conn: conn, runtime: n, record: record}, decision, nil
	}
	return backend.OpenBrowserNativeTCP(ctx, browserIdentity(q), q.Source, plan.Destination, d)
}
