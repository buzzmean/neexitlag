package controlplane

import (
	"context"
	"crypto"
	"crypto/ed25519"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/buzzmean/neexitlag/core/optimizer"
	"github.com/buzzmean/neexitlag/core/profiles"
	"github.com/buzzmean/neexitlag/core/security"
	"io"
	"math/big"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

type Settings struct {
	AdminToken string
	KeyID      string
	SigningKey ed25519.PrivateKey
	CA         *x509.Certificate
	CAKey      crypto.Signer
}
type API struct {
	Store                                     *Store
	Settings                                  Settings
	Requests, Errors, AuthFailures, Telemetry atomic.Uint64
	LatencyNS                                 atomic.Uint64
	limit                                     *limiter
}

func New(s *Store, c Settings) (*API, error) {
	if len(c.AdminToken) < 24 || len(c.SigningKey) != ed25519.PrivateKeySize || c.KeyID == "" || c.CA == nil || c.CAKey == nil {
		return nil, errors.New("admin token, signing key and device CA required")
	}
	return &API{Store: s, Settings: c, limit: newLimiter(time.Now())}, nil
}
func (a *API) publish(ctx context.Context) ([]byte, error) {
	pub := a.Settings.SigningKey.Public().(ed25519.PublicKey)
	return a.Store.Publish(ctx, a.Settings.KeyID, a.Settings.SigningKey, map[string]string{a.Settings.KeyID: base64.RawURLEncoding.EncodeToString(pub)})
}
func (a *API) Initialize(ctx context.Context) error {
	raw, e := a.Store.LatestConfig(ctx)
	if e == nil {
		_, e = security.ValidateConfig(map[string]ed25519.PublicKey{a.Settings.KeyID: a.Settings.SigningKey.Public().(ed25519.PublicKey)}, raw, time.Now())
	}
	if e != nil {
		_, e = a.publish(ctx)
	}
	return e
}
func (a *API) admin(r *http.Request) bool {
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	return len(token) == len(a.Settings.AdminToken) && subtle.ConstantTimeCompare([]byte(token), []byte(a.Settings.AdminToken)) == 1
}
func body(r *http.Request, v any) error {
	b, e := io.ReadAll(io.LimitReader(r.Body, security.MaxDocument+1))
	if e != nil {
		return e
	}
	return security.Parse(b, v)
}
func respond(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
func (a *API) fail(w http.ResponseWriter, status int, msg string) {
	a.Errors.Add(1)
	respond(w, status, map[string]string{"error": msg})
}
func (a *API) Handler() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second)
		defer cancel()
		if a.Store.DB.Ping(ctx) != nil {
			a.fail(w, 503, "database unavailable")
			return
		}
		respond(w, 200, map[string]string{"status": "ok"})
	})
	m.HandleFunc("GET /metrics", a.metrics)
	m.HandleFunc("POST /api/v1/auth", a.auth)
	m.HandleFunc("GET /api/v1/config", a.config)
	m.HandleFunc("GET /api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		respond(w, 200, map[string]string{"version": "0.3.0-dev", "ipc_version": "1"})
	})
	m.HandleFunc("/api/v1/devices", a.devices)
	m.HandleFunc("/api/v1/relays", a.relays)
	m.HandleFunc("POST /api/v1/relays/heartbeat", a.heartbeat)
	m.HandleFunc("/api/v1/apps", a.apps)
	m.HandleFunc("POST /api/v1/telemetry", a.telemetry)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer func() { a.LatencyNS.Add(uint64(time.Since(start))) }()
		a.Requests.Add(1)
		// Liveness probes cost one ping and must work even when clients are throttled.
		allowed := r.URL.Path == "/healthz" || a.limit.allow(time.Now(), sourceKey(r))
		if !allowed {
			a.fail(w, 429, "rate limit")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		m.ServeHTTP(w, r.WithContext(ctx))
	})
}
func (a *API) auth(w http.ResponseWriter, r *http.Request) {
	var p security.Proof
	if body(r, &p) != nil {
		a.fail(w, 400, "invalid device proof")
		return
	}
	d, e := a.Store.Device(r.Context(), p.DeviceID)
	if e != nil || d.Revoked {
		a.AuthFailures.Add(1)
		a.fail(w, 401, "device unavailable")
		return
	}
	key, e := security.DecodePublicKey(d.PublicKey)
	if e != nil || p.Verify(key, time.Now()) != nil {
		a.AuthFailures.Add(1)
		a.fail(w, 401, "invalid device proof")
		return
	}
	relays, e := a.Store.Relays(r.Context())
	if e != nil {
		a.fail(w, 503, "registry unavailable")
		return
	}
	exists := false
	for _, relay := range relays {
		if relay.ID == p.RelayID {
			exists = true
		}
	}
	if !exists {
		a.fail(w, 400, "unknown relay")
		return
	}
	if e = a.Store.ConsumeProof(r.Context(), p); e != nil {
		a.AuthFailures.Add(1)
		a.fail(w, 401, "proof replay or revoked device")
		return
	}
	id, e := security.NewID()
	if e != nil {
		a.fail(w, 500, "identity generation failed")
		return
	}
	now := time.Now().UTC()
	ticket := security.Ticket{ID: id, DeviceID: d.ID, DeviceKey: security.Fingerprint(key), RelayID: p.RelayID, IssuedAt: now, ExpiresAt: now.Add(2 * time.Minute)}
	raw, e := security.Sign("ticket", a.Settings.KeyID, a.Settings.SigningKey, ticket)
	if e != nil {
		a.fail(w, 500, "signing failed")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}
func (a *API) config(w http.ResponseWriter, r *http.Request) {
	raw, e := a.Store.LatestConfig(r.Context())
	if e != nil {
		a.fail(w, 503, "config unavailable")
		return
	}
	var cfg security.Config
	keys := map[string]ed25519.PublicKey{a.Settings.KeyID: a.Settings.SigningKey.Public().(ed25519.PublicKey)}
	if security.Verify("config", keys, raw, &cfg) != nil {
		a.fail(w, 500, "stored config invalid")
		return
	}
	if time.Until(cfg.ExpiresAt) < time.Hour {
		raw, e = a.publish(r.Context())
		if e != nil {
			a.fail(w, 503, "config refresh failed")
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}
func (a *API) devices(w http.ResponseWriter, r *http.Request) {
	if !a.admin(r) {
		a.fail(w, 401, "administrator authorization required")
		return
	}
	switch r.Method {
	case "GET":
		all, e := a.Store.Devices(r.Context())
		if e != nil {
			a.fail(w, 503, "database unavailable")
			return
		}
		respond(w, 200, all)
	case "POST":
		var d Device
		if body(r, &d) != nil || len(d.ID) > 128 || d.ID == "" || len(d.Name) > 128 || d.Name == "" {
			a.fail(w, 400, "invalid device")
			return
		}
		pub, e := security.DecodePublicKey(d.PublicKey)
		if e != nil {
			a.fail(w, 400, "invalid device key")
			return
		}
		serialText, e := security.NewID()
		if e != nil {
			a.fail(w, 500, "identity generation failed")
			return
		}
		serial := new(big.Int)
		serial.SetString(serialText, 16)
		now := time.Now()
		cert := &x509.Certificate{SerialNumber: serial, NotBefore: now.Add(-time.Minute), NotAfter: now.Add(30 * 24 * time.Hour), KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}}
		der, e := x509.CreateCertificate(nilReader{}, cert, a.Settings.CA, pub, a.Settings.CAKey)
		if e != nil {
			a.fail(w, 500, "certificate issuance failed")
			return
		}
		if e = a.Store.Register(r.Context(), d); e != nil {
			a.fail(w, 409, "device already registered or database unavailable")
			return
		}
		respond(w, 201, map[string]string{"id": d.ID, "certificate": string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))})
	case "DELETE":
		if e := a.Store.Revoke(r.Context(), r.URL.Query().Get("id")); e != nil {
			a.fail(w, 404, "device not found")
			return
		}
		if _, e := a.publish(r.Context()); e != nil {
			a.fail(w, 503, "revoked but config publication failed")
			return
		}
		respond(w, 200, map[string]string{"status": "revoked"})
	default:
		a.fail(w, 405, "method not allowed")
	}
}
func (a *API) relays(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		all, e := a.Store.Relays(r.Context())
		if e != nil {
			a.fail(w, 503, "database unavailable")
			return
		}
		respond(w, 200, all)
	case "POST":
		if !a.admin(r) {
			a.fail(w, 401, "administrator authorization required")
			return
		}
		var node optimizer.RelayCandidate
		if body(r, &node) != nil || node.ID == "" || len(node.ID) > 128 {
			a.fail(w, 400, "invalid relay")
			return
		}
		host, port, e := net.SplitHostPort(node.Address)
		if e != nil || host == "" || port == "" {
			a.fail(w, 400, "invalid relay address")
			return
		}
		if e = a.Store.PutRelay(r.Context(), node); e != nil {
			a.fail(w, 503, "database unavailable")
			return
		}
		if _, e = a.publish(r.Context()); e != nil {
			a.fail(w, 503, "config publication failed")
			return
		}
		token, e := security.NewID()
		if e != nil {
			a.fail(w, 500, "identity generation failed")
			return
		}
		hash := sha256.Sum256([]byte(token))
		if e = a.Store.SetHeartbeatToken(r.Context(), node.ID, hash[:]); e != nil {
			a.fail(w, 503, "database unavailable")
			return
		}
		respond(w, 201, map[string]any{"relay": node, "heartbeat_token": token})
	default:
		a.fail(w, 405, "method not allowed")
	}
}
func (a *API) apps(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		all, e := a.Store.Apps(r.Context())
		if e != nil {
			a.fail(w, 503, "database unavailable")
			return
		}
		respond(w, 200, all)
	case "POST":
		if !a.admin(r) {
			a.fail(w, 401, "administrator authorization required")
			return
		}
		var p profiles.Profile
		if body(r, &p) != nil || p.Validate() != nil {
			a.fail(w, 400, "invalid profile")
			return
		}
		if e := a.Store.PutApp(r.Context(), p); e != nil {
			a.fail(w, 503, "database unavailable")
			return
		}
		if _, e := a.publish(r.Context()); e != nil {
			a.fail(w, 503, "config publication failed")
			return
		}
		respond(w, 201, p)
	default:
		a.fail(w, 405, "method not allowed")
	}
}

type Telemetry struct {
	DeviceID string  `json:"device_id"`
	RTTMS    float64 `json:"rtt_ms"`
	JitterMS float64 `json:"jitter_ms"`
	Loss     float64 `json:"loss"`
	Bytes    uint64  `json:"bytes"`
	Version  string  `json:"version"`
}

func (a *API) telemetry(w http.ResponseWriter, r *http.Request) {
	if !a.admin(r) {
		a.fail(w, 401, "administrator authorization required")
		return
	}
	var t Telemetry
	if body(r, &t) != nil || t.Loss < 0 || t.Loss > 1 || t.RTTMS < 0 || t.JitterMS < 0 || len(t.Version) > 64 {
		a.fail(w, 400, "invalid telemetry")
		return
	}
	a.Telemetry.Add(1)
	respond(w, 202, map[string]string{"status": "accepted"})
}
func (a *API) metrics(w http.ResponseWriter, r *http.Request) {
	var healthy, devices, connected int
	a.Store.DB.QueryRow(r.Context(), "SELECT count(*) FROM relay_nodes WHERE heartbeat_at > now()-interval '60 seconds' AND healthy").Scan(&healthy)
	a.Store.DB.QueryRow(r.Context(), "SELECT count(*) FROM devices WHERE NOT revoked").Scan(&devices)
	a.Store.DB.QueryRow(r.Context(), "SELECT count(*) FROM devices WHERE NOT revoked AND last_authenticated_at > now()-interval '30 minutes'").Scan(&connected)
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	fmt.Fprintf(w, "neexitlag_api_requests_total %d\nneexitlag_api_errors_total %d\nneexitlag_auth_failures_total %d\nneexitlag_api_latency_seconds_sum %g\nneexitlag_healthy_relays %d\nneexitlag_registered_devices %d\n# HELP neexitlag_connected_devices Devices authenticated in the last 30 minutes; lease estimate, not live relay inventory.\nneexitlag_connected_devices %d\nneexitlag_telemetry_accepted_total %d\n", a.Requests.Load(), a.Errors.Load(), a.AuthFailures.Load(), float64(a.LatencyNS.Load())/1e9, healthy, devices, connected, a.Telemetry.Load())
}

// Ed25519 does not use certificate-signing randomness; still use the standard CSPRNG.
type nilReader struct{}

func (nilReader) Read(b []byte) (int, error) { return cryptorand.Read(b) }

func (a *API) heartbeat(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ID      string  `json:"id"`
		Healthy bool    `json:"healthy"`
		Load    float64 `json:"load"`
	}
	if body(r, &input) != nil || input.Load < 0 || input.Load > 1 {
		a.fail(w, 400, "invalid heartbeat")
		return
	}
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if len(token) != 32 {
		a.fail(w, 401, "relay authorization required")
		return
	}
	hash := sha256.Sum256([]byte(token))
	if e := a.Store.Heartbeat(r.Context(), input.ID, hash[:], input.Healthy, input.Load); e != nil {
		a.fail(w, 401, "invalid relay identity")
		return
	}
	raw, e := a.Store.LatestConfig(r.Context())
	if e == nil {
		var cfg security.Config
		keys := map[string]ed25519.PublicKey{a.Settings.KeyID: a.Settings.SigningKey.Public().(ed25519.PublicKey)}
		if security.Verify("config", keys, raw, &cfg) == nil && time.Since(cfg.GeneratedAt) > 30*time.Second {
			if _, e = a.publish(r.Context()); e != nil {
				a.fail(w, 503, "config publication failed")
				return
			}
		}
	}
	respond(w, 200, map[string]string{"status": "accepted"})
}
