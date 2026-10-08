package main

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"flag"
	"github.com/buzzmean/neexitlag/core/controlclient"
	"github.com/buzzmean/neexitlag/core/security"
	"github.com/buzzmean/neexitlag/core/transport"
	"github.com/buzzmean/neexitlag/services/relay"
	"log/slog"
	"net/http"
	"net/netip"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func main() {
	addr := flag.String("listen", ":4433", "QUIC UDP listener")
	cert := flag.String("cert", "", "server certificate PEM")
	key := flag.String("key", "", "server key PEM")
	ca := flag.String("ca", "", "device CA PEM")
	metrics := flag.String("metrics", "127.0.0.1:9090", "local health/metrics address")
	allow := flag.String("allow-cidrs", "", "explicit destination allowlist (comma-separated); default public only")
	relayID := flag.String("relay-id", "", "ticket audience (enables ticket-required ALPN)")
	ticketKeys := flag.String("ticket-keys", "", "out-of-band Ed25519 trust JSON")
	replayFile := flag.String("replay-cache", "relay-state/replay.json", "persistent replay cache")
	cpURL := flag.String("control-plane", "", "signed config/revocation endpoint")
	heartbeatFile := flag.String("heartbeat-token-file", "", "relay-specific heartbeat token")
	allowHTTP := flag.Bool("allow-http-development", false, "allow HTTP on isolated development network")
	flag.Parse()
	fail := func(e error) { slog.Error("relay startup failed", "error", e); os.Exit(1) }
	t, e := transport.TLS(*cert, *key, *ca, true)
	if e != nil {
		fail(e)
	}
	var prefixes []netip.Prefix
	if *allow != "" {
		for _, s := range strings.Split(*allow, ",") {
			p, e := netip.ParsePrefix(s)
			if e != nil {
				fail(e)
			}
			prefixes = append(prefixes, p)
		}
	}
	var verifier *security.TicketVerifier
	var cp *controlclient.Client
	var heartbeat string
	if *relayID != "" {
		raw, e := os.ReadFile(*ticketKeys)
		if e != nil {
			fail(e)
		}
		var encoded map[string]string
		if e = json.Unmarshal(raw, &encoded); e != nil {
			fail(e)
		}
		keys := map[string]ed25519.PublicKey{}
		for id, value := range encoded {
			k, e := security.DecodePublicKey(value)
			if e != nil {
				fail(e)
			}
			keys[id] = k
		}
		verifier, e = security.NewPersistentTicketVerifier(keys, *relayID, *replayFile)
		if e != nil {
			fail(e)
		}
		if *cpURL != "" {
			cache := security.NewConfigCache(keys, filepath.Join(filepath.Dir(*replayFile), "signed-config.json"))
			_ = cache.Load(time.Now())
			cp = &controlclient.Client{BaseURL: *cpURL, Cache: cache, AllowHTTPDevelopment: *allowHTTP}
		}
	}
	if *heartbeatFile != "" {
		raw, e := os.ReadFile(*heartbeatFile)
		if e != nil {
			fail(e)
		}
		heartbeat = strings.TrimSpace(string(raw))
	}
	r, e := relay.Listen(*addr, relay.Config{TLS: t, Allow: prefixes, TicketVerifier: verifier})
	if e != nil {
		fail(e)
	}
	background, cancelBackground := context.WithCancel(context.Background())
	defer cancelBackground()
	if cp != nil {
		go func() {
			tick := time.NewTicker(30 * time.Second)
			defer tick.Stop()
			for {
				request, stop := context.WithTimeout(background, 5*time.Second)
				cfg, e := cp.Config(request)
				if e == nil {
					keys := map[string]ed25519.PublicKey{}
					for id, value := range cfg.PublicKeys {
						k, e := security.DecodePublicKey(value)
						if e == nil {
							keys[id] = k
						}
					}
					verifier.Update(keys, cfg.RevokedDevices)
				}
				if heartbeat != "" {
					_ = cp.Heartbeat(request, *relayID, heartbeat, float64(r.Metrics.Sessions.Load())/128)
				}
				stop()
				select {
				case <-background.Done():
					return
				case <-tick.C:
				}
			}
		}()
	}
	h := &http.Server{Addr: *metrics, Handler: r.Handler(), ReadHeaderTimeout: 3 * time.Second, WriteTimeout: 3 * time.Second, IdleTimeout: 30 * time.Second}
	go func() {
		if e := h.ListenAndServe(); e != nil && e != http.ErrServerClosed {
			slog.Error("metrics listener failed", "error", e)
		}
	}()
	slog.Info("relay listening", "address", r.Addr().String())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	r.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	h.Shutdown(ctx)
}
