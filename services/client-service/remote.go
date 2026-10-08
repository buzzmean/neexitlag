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
	windowsplatform "github.com/buzzmean/neexitlag/platform/windows"
	"net/netip"
	"os"
	"path/filepath"
	"time"
)

type Manifest struct {
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
	c.ControlPlaneURL = m.ControlPlaneURL
	c.TCPDiagnosticTarget = m.TCPProbeTarget
	target, e := netip.ParseAddrPort(m.ProbeTarget)
	if e != nil {
		return errors.New("explicit UDP echo probe target required")
	}
	remote := &remoteState{controller: c, cp: cp, tls: cfg, target: target, clients: make(map[string]*transport.Client)}
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
	controller  *Controller
	cp          *controlclient.Client
	tls         *tls.Config
	target      netip.AddrPort
	clients     map[string]*transport.Client
	all         []*transport.Client
	version     uint64
	lastConnect time.Time
}

func (s *remoteState) apply(ctx context.Context, cfg security.Config) error {
	c := s.controller
	live := s.all[:0]
	for _, client := range s.all {
		select {
		case <-client.Done():
		default:
			live = append(live, client)
		}
	}
	s.all = live
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
	if len(s.clients) == 0 || time.Since(s.lastConnect) > 25*time.Minute {
		for _, node := range cfg.Relays {
			ticket, e := s.cp.Ticket(ctx, node.ID)
			if e != nil {
				continue
			}
			attempt, cancel := context.WithTimeout(ctx, 3*time.Second)
			client, e := transport.DialWithTicket(attempt, node.Address, s.tls, ticket)
			cancel()
			if e != nil {
				continue
			}
			s.clients[node.ID] = client
			s.all = append(s.all, client)
			c.Engine.SetRelay(node.ID, client)
		}
		s.lastConnect = time.Now()
	}
	candidates := []measurement.Candidate{{Metadata: optimizer.RelayCandidate{ID: optimizer.Direct}, Target: s.target, Open: measurement.OpenDirect}}
	for _, node := range cfg.Relays {
		client := s.clients[node.ID]
		if client == nil {
			continue
		}
		select {
		case <-client.Done():
			delete(s.clients, node.ID)
			c.Engine.SetRelay(node.ID, nil)
			s.lastConnect = time.Time{}
			continue
		default:
		}
		candidates = append(candidates, measurement.Candidate{Metadata: node, Target: s.target, Open: client.OpenDatagram, Load: node.Load})
	}
	c.Measurements.SetCandidates(candidates)
	return nil
}
func (s *remoteState) close() {
	for _, c := range s.all {
		c.Close()
	}
	s.clients = make(map[string]*transport.Client)
	s.all = nil
}
