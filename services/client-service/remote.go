package clientservice

import (
	"context"
	"crypto/ed25519"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"github.com/buzzmean/neexitlag/core/controlclient"
	"github.com/buzzmean/neexitlag/core/measurement"
	"github.com/buzzmean/neexitlag/core/optimizer"
	"github.com/buzzmean/neexitlag/core/security"
	"github.com/buzzmean/neexitlag/core/transport"
	"github.com/buzzmean/neexitlag/internal/e2eprobe"
	windowsplatform "github.com/buzzmean/neexitlag/platform/windows"
	"net/netip"
	"os"
	"path/filepath"
	"time"
)

type Manifest struct {
	ProbeTokenFile     string                      `json:"probe_token_file,omitempty"`
	ControlPlaneCAFile string                      `json:"control_plane_ca_file,omitempty"`
	OwnerSID           string                      `json:"owner_sid"`
	TokenFile          string                      `json:"token_file"`
	StateDirectory     string                      `json:"state_directory"`
	ControlPlaneURL    string                      `json:"control_plane_url"`
	TrustFile          string                      `json:"trust_file"`
	DeviceIdentityFile string                      `json:"device_identity_file"`
	DeviceCertificate  string                      `json:"device_certificate"`
	CAFile             string                      `json:"ca_file"`
	ProbeTarget        string                      `json:"probe_target"`
	TCPProbeTarget     string                      `json:"tcp_probe_target"`
	Driver             windowsplatform.DriverFiles `json:"driver"`
}

func LoadManifest(path string) (Manifest, error) {
	var m Manifest
	raw, e := os.ReadFile(path)
	if e != nil {
		return m, e
	}
	if e = security.Parse(raw, &m); e != nil {
		return m, e
	}
	if !filepath.IsAbs(m.TokenFile) || !filepath.IsAbs(m.StateDirectory) {
		return m, errors.New("absolute token and state paths required")
	}
	return m, nil
}
func (c *Controller) ConnectRemote(ctx context.Context, m Manifest) error {
	raw, e := os.ReadFile(m.TrustFile)
	if e != nil {
		return e
	}
	var encoded map[string]string
	if e = json.Unmarshal(raw, &encoded); e != nil {
		return e
	}
	keys := map[string]ed25519.PublicKey{}
	for id, value := range encoded {
		key, e := security.DecodePublicKey(value)
		if e != nil {
			return e
		}
		keys[id] = key
	}
	identity, key, e := security.LoadOrCreateIdentity(m.DeviceIdentityFile)
	if e != nil {
		return e
	}
	certRaw, e := os.ReadFile(m.DeviceCertificate)
	if e != nil {
		return e
	}
	certPool := x509.NewCertPool()
	caRaw, e := os.ReadFile(m.CAFile)
	if e != nil {
		return e
	}
	if !certPool.AppendCertsFromPEM(caRaw) {
		return errors.New("invalid relay CA")
	}
	block, _ := pem.Decode(certRaw)
	if block == nil {
		return errors.New("invalid device certificate")
	}
	leaf, e := x509.ParseCertificate(block.Bytes)
	if e != nil {
		return errors.New("invalid device certificate")
	}
	public, ok := leaf.PublicKey.(ed25519.PublicKey)
	if !ok || !public.Equal(key.Public()) {
		return errors.New("device certificate does not match persisted identity")
	}
	cert := tls.Certificate{Certificate: [][]byte{block.Bytes}, PrivateKey: key, Leaf: leaf}
	cfg := &tls.Config{MinVersion: tls.VersionTLS13, RootCAs: certPool, Certificates: []tls.Certificate{cert}}
	cache := security.NewConfigCache(keys, filepath.Join(m.StateDirectory, "signed-config.json"))
	_ = cache.Load(time.Now())
	cp := &controlclient.Client{BaseURL: m.ControlPlaneURL, Cache: cache, DeviceID: identity.ID, DeviceKey: key}
	if m.ControlPlaneCAFile != "" {
		cp.HTTP, e = controlclient.HTTPSClient(m.ControlPlaneCAFile)
		if e != nil {
			return e
		}
	}
	c.DiagnosticHTTPClient = cp.HTTP
	c.ControlPlaneURL = m.ControlPlaneURL
	c.TCPDiagnosticTarget = m.TCPProbeTarget
	target, e := netip.ParseAddrPort(m.ProbeTarget)
	if e != nil {
		return errors.New("explicit UDP echo probe target required")
	}
	remote := &remoteState{controller: c, cp: cp, tls: cfg, target: target, addresses: make(map[string]string), clients: make(map[string]*transport.Pool)}
	if m.ProbeTokenFile != "" {
		remote.echoKey, e = e2eprobe.Token(m.ProbeTokenFile)
		if e != nil {
			return e
		}
	}
	signed, initialError := cp.Config(ctx)
	if initialError == nil {
		initialError = remote.apply(ctx, signed)
	}
	// Keep retrying after an initial outage; no valid config means DIRECT.
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer remote.close()
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-c.ctx.Done():
				return
			case <-ticker.C:
				cfg, e := cp.Config(c.ctx)
				if e == nil {
					remote.apply(c.ctx, cfg)
				}
			}
		}
	}()
	return initialError
}

type remoteState struct {
	addresses  map[string]string
	echoKey    []byte
	controller *Controller
	cp         *controlclient.Client
	tls        *tls.Config
	target     netip.AddrPort
	clients    map[string]*transport.Pool
	version    uint64
}

func (s *remoteState) apply(ctx context.Context, cfg security.Config) error {
	c := s.controller
	for _, id := range cfg.RevokedDevices {
		if id == s.cp.DeviceID {
			s.close()
			return errors.New("device revoked")
		}
	}
	if cfg.Version != s.version {
		if e := c.Selector.Configure(cfg.RoutingParameters); e != nil {
			return e
		}
		for _, p := range cfg.ApplicationProfiles {
			c.mu.Lock()
			_, local := c.localProfiles[p.ID]
			c.mu.Unlock()
			if !local {
				if e := c.Profiles.Upsert(p); e != nil {
					return e
				}
			}
		}
		c.syncPolicies()
		s.version = cfg.Version
	}
	// Signed removals retire only that relay; other pools/flows remain independent.
	configured := map[string]bool{}
	for _, node := range cfg.Relays {
		configured[node.ID] = true
	}
	for id, pool := range s.clients {
		if !configured[id] {
			pool.Close()
			delete(s.clients, id)
			delete(s.addresses, id)
			c.Engine.SetRelay(id, nil)
			c.mu.Lock()
			delete(c.acceptanceRelays, id)
			delete(c.relayPools, id)
			delete(c.transportDrops, id)
			c.mu.Unlock()
		}
	}
	// Stable provider per relay; session creation/rotation stays inside its pool.
	for _, node := range cfg.Relays {
		node := node
		factory := func(ctx context.Context) (*transport.Client, error) {
			// Tickets are one-shot. Each factory invocation obtains a fresh signed
			// proof/ticket, including retries after an uncertain failed handshake.
			ticket, err := s.cp.Ticket(ctx, node.ID)
			if err != nil {
				return nil, &transport.OpenError{Reason: "ticket_request_failed", Cause: err}
			}
			client, err := transport.DialWithTicket(ctx, node.Address, s.tls, ticket)
			if err != nil {
				return nil, &transport.OpenError{Reason: "session_auth_or_dial_failed", Cause: err}
			}
			return client, nil
		}
		if pool := s.clients[node.ID]; pool != nil {
			pool.UpdateFactory(factory, s.addresses[node.ID] != node.Address)
			s.addresses[node.ID] = node.Address
			continue
		}
		pool, err := transport.NewPool(c.ctx, transport.DefaultPoolConfig(), factory)
		if err != nil {
			return err
		}
		s.clients[node.ID] = pool
		s.addresses[node.ID] = node.Address
		c.Engine.SetRelay(node.ID, pool)
		generation := c.Selector.RegisterTransport(node.ID)
		c.mu.Lock()
		c.transportDrops[node.ID] = pool.DropSnapshot
		c.relayPools[node.ID] = pool
		c.acceptanceRelays[node.ID] = acceptanceRelaySource{session: pool, target: s.target, generation: generation}
		c.mu.Unlock()
		pool.SetFailureHandler(func(reason string) { c.Selector.ReportTransportFailure(node.ID, generation, reason, time.Now()) })
		attempt, cancel := context.WithTimeout(ctx, 3*time.Second)
		_ = pool.Warm(attempt)
		cancel()
	}
	candidates := []measurement.Candidate{{Metadata: optimizer.RelayCandidate{ID: optimizer.Direct}, Target: s.target, Open: measurement.OpenDirect}}
	for _, node := range cfg.Relays {
		client := s.clients[node.ID]
		if client == nil {
			continue
		}
		candidates = append(candidates, measurement.Candidate{Metadata: node, Target: s.target, Open: client.OpenProbeDatagram, Load: node.Load})
	}
	if len(s.echoKey) > 0 {
		for i := range candidates {
			candidates[i].Open = e2eprobe.ProbeOpener(candidates[i].Open, s.echoKey)
		}
	}
	c.Measurements.SetCandidates(candidates)
	return nil
}
func (s *remoteState) close() {
	for _, pool := range s.clients {
		pool.Close()
	}
	s.clients = make(map[string]*transport.Pool)
}
