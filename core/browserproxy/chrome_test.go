package browserproxy

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"github.com/buzzmean/neexitlag/core/flowpolicy"
	"github.com/buzzmean/neexitlag/core/security"
	"github.com/buzzmean/neexitlag/core/transport"
	"github.com/buzzmean/neexitlag/internal/browserfixture"
	"github.com/buzzmean/neexitlag/services/relay"
	"github.com/buzzmean/neexitlag/tools/labcert"
	"net"
	"net/netip"
	"testing"
	"time"
)

func TestChromePACRealHTTPSOriginIsolation(t *testing.T) {
	chrome := browserfixture.ChromePath()
	if chrome == "" {
		t.Skip("Chrome unavailable; native Windows CI provides it")
	}
	f, e := browserfixture.New()
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	st, ct, e := labcert.Pair()
	if e != nil {
		t.Fatal(e)
	}
	pub, key, _ := ed25519.GenerateKey(rand.Reader)
	server, e := relay.Listen("127.0.0.1:0", relay.Config{TLS: st, TicketVerifier: security.NewTicketVerifier(map[string]ed25519.PublicKey{"test": pub}, "VPS-01"), Allow: []netip.Prefix{netip.MustParsePrefix("127.0.0.0/8")}, Acceptance: &relay.AcceptanceConfig{LeaseDuration: 90 * time.Second, OutboundSource: netip.MustParseAddr("127.0.0.2")}})
	if e != nil {
		t.Fatal(e)
	}
	defer server.Close()
	peer := ct.Certificates[0].PrivateKey.(ed25519.PrivateKey).Public().(ed25519.PublicKey)
	pool, e := transport.NewPool(context.Background(), transport.DefaultPoolConfig(), func(ctx context.Context) (*transport.Client, error) {
		id, _ := security.NewID()
		now := time.Now()
		ticket, e := security.Sign("ticket", "test", key, security.Ticket{ID: id, DeviceID: "fixture", DeviceKey: security.Fingerprint(peer), RelayID: "VPS-01", IssuedAt: now, ExpiresAt: now.Add(time.Minute)})
		if e != nil {
			return nil, e
		}
		return transport.DialWithTicket(ctx, server.Addr().String(), ct, ticket)
	})
	if e != nil {
		t.Fatal(e)
	}
	defer pool.Close()
	open := func(ctx context.Context, q Request, p Plan) (net.Conn, error) {
		if p.Policy.Action == flowpolicy.RelayRequired {
			return pool.OpenTCP(ctx, p.Destination)
		}
		return (&net.Dialer{}).DialContext(ctx, "tcp4", p.Destination.String())
	}
	learning := flowpolicy.NewBrowserLearning()
	proxy, e := Listen(context.Background(), Config{Authorize: func(context.Context, netip.AddrPort, netip.AddrPort) (Owner, error) {
		return Owner{PID: 1, Birth: 1, Exe: chrome, Profile: "fixture"}, nil
	}, Open: open, Plan: func(ctx context.Context, q Request) (Plan, error) {
		plan := Plan{Destination: f.Destination(), Policy: flowpolicy.Result{Action: flowpolicy.Direct, Fallback: flowpolicy.FailOpen}, Probe: "NOT_PROBED"}
		if q.Host == "chatgpt.com" {
			plan.Policy = flowpolicy.Result{Action: flowpolicy.RelayRequired, RelayID: "VPS-01", Fallback: flowpolicy.Required}
			return plan, nil
		}
		if q.Host == "blocked.example" {
			k := flowpolicy.LearnKey{Network: "isolated", Application: chrome, Protocol: "tcp", Target: q.Host, Port: q.Port}
			a, e := learning.Begin(k, "", "VPS-01", time.Now())
			if e != nil {
				return plan, e
			}
			if a.Confirmed {
				plan.Policy = a.Route
				plan.Probe = a.Outcome
				return plan, nil
			}
			for i := 0; i < 2; i++ {
				plan.Policy = a.Route
				status, rx, why := ProbeOrigin(ctx, q.Host, q.Port, true, f.Roots, func(ctx context.Context) (net.Conn, error) { return open(ctx, q, plan) })
				a, e = learning.Observe(k, flowpolicy.BrowserObservation{Token: a.Token, HTTPStatus: status, NetworkError: why, ResponseBytes: rx, Correlated: true}, "", "VPS-01", time.Now())
				if e != nil {
					return plan, e
				}
				if a.Confirmed {
					plan.Policy = a.Route
					plan.Probe = a.Outcome
					return plan, nil
				}
			}
		}
		return plan, nil
	}})
	if e != nil {
		t.Fatal(e)
	}
	defer proxy.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	args, e := ChromeArguments(proxy.Address(), t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	out, e := f.RunChrome(ctx, chrome, args)
	if e != nil {
		t.Fatal(e, out, proxy.Snapshot())
	}
	snap := proxy.Snapshot()
	seen := map[string]bool{}
	for _, r := range snap.Records {
		for _, host := range browserfixture.Domains {
			if r.Domain == host {
				seen[host] = true
				if (host == "chatgpt.com" || host == "blocked.example") && (r.Plan.Policy.Action != flowpolicy.RelayRequired || r.Plan.Policy.Fallback != flowpolicy.Required || r.Sent == 0 || r.Received == 0) {
					t.Fatal("protected transport not proved", r)
				}
			}
		}
	}
	if len(seen) != 4 || snap.Lost != 0 {
		t.Fatal("missing origin CONNECT / possible pooling", seen, snap.Lost)
	}
	if server.Metrics.TCPConnections.Load() < 2 || server.Metrics.RXBytes.Load() == 0 {
		t.Fatal("real authenticated relay not used")
	}
	t.Log("real Chrome + HTTPS/HTTP2 origin isolation, authenticated protected relay, ordinary DIRECT, blocked-origin recovery; DIRECT+ is verified separately by the signed-driver Windows fixture")
}
