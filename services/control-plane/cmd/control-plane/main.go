package main

import (
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"github.com/buzzmean/neexitlag/core/optimizer"
	controlplane "github.com/buzzmean/neexitlag/services/control-plane"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func readKey(path string) (crypto.Signer, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return nil, e
	}
	p, _ := pem.Decode(b)
	if p == nil {
		return nil, fmt.Errorf("invalid PEM key")
	}
	k, e := x509.ParsePKCS8PrivateKey(p.Bytes)
	if e != nil {
		return nil, e
	}
	s, ok := k.(crypto.Signer)
	if !ok {
		return nil, fmt.Errorf("unsupported signing key")
	}
	return s, nil
}
func main() {
	addr := flag.String("listen", "127.0.0.1:8080", "HTTP listener; use TLS ingress outside localhost")
	sign := flag.String("signing-key", "", "Ed25519 PKCS8 PEM")
	kid := flag.String("key-id", "dev-1", "signing key ID")
	ca := flag.String("ca", "", "device CA PEM")
	caKey := flag.String("ca-key", "", "device CA PKCS8 PEM")
	adminTokenFile := flag.String("admin-token-file", "", "administrator token file")
	databasePasswordFile := flag.String("database-password-file", "", "database password file (overrides URL password)")
	seedRelayID := flag.String("seed-relay-id", "", "development relay registry seed ID")
	seedRelayAddress := flag.String("seed-relay-address", "", "development relay registry seed address")
	seedHeartbeatFile := flag.String("seed-heartbeat-token-file", "", "relay-specific heartbeat token seed file")
	flag.Parse()
	fail := func(e error) { slog.Error("control plane startup failed", "error", e); os.Exit(1) }
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	databaseURL := os.Getenv("DATABASE_URL")
	if *databasePasswordFile != "" {
		b, e := os.ReadFile(*databasePasswordFile)
		if e != nil {
			fail(e)
		}
		u, e := url.Parse(databaseURL)
		if e != nil || u.User == nil {
			fail(fmt.Errorf("database URL requires a username"))
		}
		u.User = url.UserPassword(u.User.Username(), strings.TrimSpace(string(b)))
		databaseURL = u.String()
	}
	adminToken := os.Getenv("NEEXITLAG_ADMIN_TOKEN")
	if *adminTokenFile != "" {
		b, e := os.ReadFile(*adminTokenFile)
		if e != nil {
			fail(e)
		}
		adminToken = strings.TrimSpace(string(b))
	}
	store, e := controlplane.Open(ctx, databaseURL)
	if e != nil {
		fail(e)
	}
	defer store.DB.Close()
	if e = store.Migrate(ctx); e != nil {
		fail(e)
	}
	signer, e := readKey(*sign)
	if e != nil {
		fail(e)
	}
	key, ok := signer.(ed25519.PrivateKey)
	if !ok {
		fail(fmt.Errorf("Ed25519 signing key required"))
	}
	b, e := os.ReadFile(*ca)
	if e != nil {
		fail(e)
	}
	p, _ := pem.Decode(b)
	if p == nil {
		fail(fmt.Errorf("invalid CA PEM"))
	}
	cert, e := x509.ParseCertificate(p.Bytes)
	if e != nil {
		fail(e)
	}
	cak, e := readKey(*caKey)
	if e != nil {
		fail(e)
	}
	api, e := controlplane.New(store, controlplane.Settings{AdminToken: adminToken, KeyID: *kid, SigningKey: key, CA: cert, CAKey: cak})
	if e != nil {
		fail(e)
	}
	if *seedRelayID != "" {
		if *seedRelayAddress == "" {
			fail(fmt.Errorf("seed relay address required"))
		}
		if e = store.PutRelay(ctx, optimizer.RelayCandidate{ID: *seedRelayID, Address: *seedRelayAddress}); e != nil {
			fail(e)
		}
		if *seedHeartbeatFile != "" {
			b, e := os.ReadFile(*seedHeartbeatFile)
			if e != nil {
				fail(e)
			}
			hash := sha256.Sum256([]byte(strings.TrimSpace(string(b))))
			if e = store.SetHeartbeatToken(ctx, *seedRelayID, hash[:]); e != nil {
				fail(e)
			}
		}
	}
	if e = api.Initialize(ctx); e != nil {
		fail(e)
	}
	server := &http.Server{Addr: *addr, Handler: api.Handler(), ReadHeaderTimeout: 3 * time.Second, ReadTimeout: 5 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 30 * time.Second, MaxHeaderBytes: 16 * 1024}
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-done
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()
	slog.Info("control plane listening", "address", *addr)
	if e = server.ListenAndServe(); e != nil && e != http.ErrServerClosed {
		fail(e)
	}
}
