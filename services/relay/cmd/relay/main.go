package main

import (
	"context"
	"flag"
	"github.com/buzzmean/neexitlag/core/transport"
	"github.com/buzzmean/neexitlag/services/relay"
	"log/slog"
	"net/http"
	"net/netip"
	"os"
	"os/signal"
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
	r, e := relay.Listen(*addr, relay.Config{TLS: t, Allow: prefixes})
	if e != nil {
		fail(e)
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
