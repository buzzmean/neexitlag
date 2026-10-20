// Root SSH enrollment helper. No listener, shared bootstrap key or admin token on PC.
package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/buzzmean/neexitlag/core/controlclient"
	"github.com/buzzmean/neexitlag/core/freshpc"
	"github.com/buzzmean/neexitlag/core/security"
	"github.com/buzzmean/neexitlag/core/transport"
	"github.com/buzzmean/neexitlag/internal/e2eprobe"
	cp "github.com/buzzmean/neexitlag/services/control-plane"
)

func readJSON(path string, v any) error {
	b, e := os.ReadFile(path)
	if e != nil {
		return errors.New("INPUT_MISSING")
	}
	return security.Parse(b, v)
}
func privateKey(path string) (ed25519.PrivateKey, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return nil, errors.New("PRIVATE_STATE_UNAVAILABLE")
	}
	p, _ := pem.Decode(b)
	if p == nil {
		return nil, errors.New("PRIVATE_STATE_INVALID")
	}
	v, e := x509.ParsePKCS8PrivateKey(p.Bytes)
	if e != nil {
		return nil, errors.New("PRIVATE_STATE_INVALID")
	}
	k, ok := v.(ed25519.PrivateKey)
	if !ok {
		return nil, errors.New("ED25519_REQUIRED")
	}
	return k, nil
}
func checkHost(ip string) error {
	a, e := netip.ParseAddr(ip)
	if e != nil || a.String() != "2.56.177.4" || !a.Is4() || !a.IsGlobalUnicast() || a.IsPrivate() || a.IsLoopback() {
		return errors.New("PUBLIC_VPS_IPV4_REQUIRED")
	}
	host, _ := os.Hostname()
	if strings.EqualFold(strings.Split(host, ".")[0], "veliq-swe-01") {
		return errors.New("EXCLUDED_HOST")
	}
	if runtime.GOOS != "linux" || os.Geteuid() != 0 {
		return errors.New("ROOT_LINUX_REQUIRED")
	}
	raw, e := os.ReadFile("/etc/neexitlag/public/relay.env")
	if e != nil {
		return errors.New("EXISTING_VPS_REQUIRED")
	}
	matched := false
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.TrimSpace(line) == "VPS_PUBLIC_IP="+ip {
			matched = true
		}
	}
	if !matched {
		return errors.New("VPS_IDENTITY_MISMATCH")
	}
	return nil
}
func rootState(ip string) (*x509.Certificate, ed25519.PrivateKey, ed25519.PrivateKey, map[string]string, *cp.Store, error) {
	fail := func() (*x509.Certificate, ed25519.PrivateKey, ed25519.PrivateKey, map[string]string, *cp.Store, error) {
		return nil, nil, nil, nil, nil, errors.New("SERVER_ENROLLMENT_PREFLIGHT_FAILED")
	}
	if checkHost(ip) != nil {
		return fail()
	}
	credential, e := os.ReadFile("/etc/systemd/system/neexitlag-control-plane.service.d/91-veliq-fresh-ca.conf")
	if e != nil || !strings.Contains(string(credential), "LoadCredential=veliq-device-ca-key:/etc/neexitlag/control/ca.key") || !rootPrivateCA("/etc/neexitlag/control/ca.key") {
		return nil, nil, nil, nil, nil, errors.New("CA_MIGRATION_REQUIRED: root-owned 0600 CA key and systemd credential required")
	}
	raw, e := os.ReadFile("/etc/neexitlag/control/ca.pem")
	if e != nil {
		return fail()
	}
	ca, e := freshpc.ParseCertificate(string(raw))
	if e != nil || !ca.IsCA || ca.CheckSignatureFrom(ca) != nil || time.Now().Before(ca.NotBefore) || !time.Now().Before(ca.NotAfter) {
		return fail()
	}
	key, e := privateKey("/etc/neexitlag/control/ca.key")
	caPub, caOK := ca.PublicKey.(ed25519.PublicKey)
	if e != nil || !caOK || !caPub.Equal(key.Public()) {
		return fail()
	}
	signer, e := privateKey("/etc/neexitlag/control/signing.key")
	if e != nil {
		return fail()
	}
	var trust map[string]string
	if readJSON("/etc/neexitlag/relay/trust.json", &trust) != nil || trust["vps-1"] != base64.RawURLEncoding.EncodeToString(signer.Public().(ed25519.PublicKey)) {
		return fail()
	}
	env, e := os.ReadFile("/etc/neexitlag/control/control-plane.env")
	if e != nil {
		return fail()
	}
	database := ""
	for _, line := range strings.Split(string(env), "\n") {
		if strings.HasPrefix(line, "DATABASE_URL=") {
			database = strings.TrimSpace(strings.TrimPrefix(line, "DATABASE_URL="))
		}
	}
	u, e := url.Parse(database)
	if e != nil || u.Scheme != "postgres" && u.Scheme != "postgresql" || u.User == nil {
		return fail()
	}
	pass, e := os.ReadFile("/etc/neexitlag/control/database.password")
	if e != nil {
		return fail()
	}
	u.User = url.UserPassword(u.User.Username(), strings.TrimSpace(string(pass)))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	store, e := cp.Open(ctx, u.String())
	if e != nil {
		return fail()
	}
	return ca, key, signer, trust, store, nil
}
func serve(ip, expectedCA string) error {
	var r freshpc.Request
	b, e := io.ReadAll(io.LimitReader(os.Stdin, 16385))
	if e != nil || len(b) > 16384 || security.Parse(b, &r) != nil {
		return errors.New("INVALID_REQUEST")
	}
	pub, e := freshpc.ValidateRequest(r)
	if e != nil {
		return e
	}
	ca, key, signer, trust, store, e := rootState(ip)
	if e != nil {
		return e
	}
	defer store.DB.Close()
	if expectedCA != freshpc.Fingerprint(ca) {
		return errors.New("CA_FINGERPRINT_MISMATCH: no enrollment performed")
	}
	cert, e := freshpc.Issue(ca, key, pub, "")
	if e != nil {
		return errors.New("CSR_SIGN_FAILED")
	}
	token, e := e2eprobe.Token("/etc/neexitlag/echo/echo.token")
	if e != nil {
		return errors.New("ECHO_CREDENTIAL_UNAVAILABLE")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if e = store.EnrollFreshDevice(ctx, r.MachineHash, cp.Device{ID: r.DeviceID, Name: "Fresh Windows PC", PublicKey: base64.RawURLEncoding.EncodeToString(pub)}, "vps-1", signer, trust); e != nil {
		return errors.New("ENROLLMENT_TRANSACTION_FAILED")
	}
	return json.NewEncoder(os.Stdout).Encode(freshpc.Response{DeviceID: r.DeviceID, MachineHash: r.MachineHash, CA: string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: ca.Raw})), CASHA256: freshpc.Fingerprint(ca), Certificate: cert, Trust: trust, EchoToken: string(token)})
}

// Probe uses signed device proof, a fresh one-shot ticket and real authenticated
// TCP+UDP echo. No anonymous QUIC handshake is mistaken for relay readiness.
func probe(state, ip string) error {
	return probeEndpoints(state, ip, "https://"+ip+":8443", ip+":4433", ip+":40000")
}

// Endpoint parameters are internal fixture seams, never a production test flag.
func probeEndpoints(state, ip, controlURL, relayAddress, target string) error {
	var r freshpc.Request
	var response freshpc.Response
	if readJSON(filepath.Join(state, "request.json"), &r) != nil || readJSON(filepath.Join(state, "response.json"), &response) != nil {
		return errors.New("PROBE_INPUT_MISSING")
	}
	if e := freshpc.ValidateResponse(r, response, response.CASHA256, time.Now()); e != nil {
		return e
	}
	id, key, e := security.LoadOrCreateIdentity(filepath.Join(state, "device", "identity.json"))
	if e != nil {
		return errors.New("PROBE_IDENTITY_INVALID")
	}
	roots := x509.NewCertPool()
	roots.AppendCertsFromPEM([]byte(response.CA))
	hc := &http.Client{Timeout: 5 * time.Second, Transport: &http.Transport{Proxy: nil, TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS13, RootCAs: roots}}, CheckRedirect: func(*http.Request, []*http.Request) error { return errors.New("redirect refused") }}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cc := &controlclient.Client{BaseURL: controlURL, HTTP: hc, DeviceID: id.ID, DeviceKey: key}
	resp, e := hc.Get(cc.BaseURL + "/healthz")
	if e != nil {
		return errors.New("CP_TLS_HEALTH_FAILED")
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New("CP_HEALTH_FAILED")
	}
	ticket, e := cc.Ticket(ctx, "VPS-01")
	if e != nil {
		return errors.New("SIGNED_DEVICE_PROOF_OR_TICKET_FAILED")
	}
	var signed security.Ticket
	keys := map[string]ed25519.PublicKey{}
	for kid, v := range response.Trust {
		k, _ := security.DecodePublicKey(v)
		keys[kid] = k
	}
	if security.Verify("ticket", keys, ticket, &signed) != nil || signed.DeviceID != id.ID || signed.DeviceKey != security.Fingerprint(key.Public().(ed25519.PublicKey)) || signed.RelayID != "VPS-01" || !time.Now().Before(signed.ExpiresAt) {
		return errors.New("TICKET_SIGNATURE_OR_BINDING_FAILED")
	}
	pair, e := tls.X509KeyPair([]byte(response.Certificate), freshpc.KeyPEM(key))
	if e != nil {
		return errors.New("DEVICE_PAIR_INVALID")
	}
	relay, e := transport.DialWithTicket(ctx, relayAddress, &tls.Config{MinVersion: tls.VersionTLS13, RootCAs: roots, ServerName: ip, Certificates: []tls.Certificate{pair}}, ticket)
	if e != nil {
		return errors.New("QUIC_4433_MTLS_AUTH_FAILED")
	}
	defer relay.Close()
	if !relay.Lease().Supported || relay.Lease().Duration != 30*time.Minute {
		return errors.New("PROTOCOL_3_OR_PRODUCTION_LEASE_FAILED")
	}
	dst := netip.MustParseAddrPort(target)
	nonce := e2eprobe.Packet{Sequence: 1}
	if _, e = rand.Read(nonce.Run[:]); e != nil {
		return e
	}
	wire := e2eprobe.Encode(nonce, []byte(response.EchoToken))
	validate := func(b []byte) bool {
		p, e := e2eprobe.Decode(b, []byte(response.EchoToken))
		return e == nil && p.Reply && p.Run == nonce.Run && p.Sequence == nonce.Sequence && p.Peer.Addr().Unmap().String() == ip
	}
	tcp, e := relay.OpenTCP(ctx, dst)
	if e != nil {
		return errors.New("RELAY_TCP_OPEN_FAILED")
	}
	defer tcp.Close()
	tcp.SetDeadline(time.Now().Add(4 * time.Second))
	if _, e = tcp.Write(wire); e != nil {
		return errors.New("RELAY_TCP_WRITE_FAILED")
	}
	reply := make([]byte, e2eprobe.Size)
	if _, e = io.ReadFull(tcp, reply); e != nil || !validate(reply) {
		return errors.New("RELAY_TCP_ECHO_FAILED")
	}
	udp, e := relay.OpenUDP(ctx, dst)
	if e != nil {
		return errors.New("RELAY_UDP_OPEN_FAILED")
	}
	defer udp.Close()
	if udp.Send(wire) != nil {
		return errors.New("RELAY_UDP_SEND_FAILED")
	}
	rx, e := udp.Receive(ctx)
	if e != nil || !validate(rx) {
		return errors.New("RELAY_UDP_ECHO_FAILED")
	}
	fmt.Println("PASS: pinned CP TLS, signed proof/ticket, mTLS QUIC /3, authenticated TCP+UDP echo")
	return nil
}
func run() error {
	if len(os.Args) < 2 {
		return errors.New("mode required")
	}
	f := flag.NewFlagSet("fresh-pc", flag.ContinueOnError)
	state := f.String("state", "", "private repository staging")
	machine := f.String("machine-hash", "", "hardware inventory SHA256")
	pin := f.String("ca-sha256", "", "expected DER CA SHA256 from authenticated SSH or owner")
	ip := f.String("vps-ip", "", "explicit VPS IPv4")
	out := f.String("out", "", "new CA staging directory")
	if e := f.Parse(os.Args[2:]); e != nil {
		return errors.New("ARGUMENTS_INVALID")
	}
	if f.NArg() != 0 {
		return errors.New("ARGUMENTS_INVALID")
	}
	if (os.Args[1] == "request" || os.Args[1] == "complete" || os.Args[1] == "probe") && !filepath.IsAbs(*state) {
		return errors.New("ABSOLUTE_PRIVATE_STATE_REQUIRED")
	}
	switch os.Args[1] {
	case "request":
		_, e := freshpc.RequestDevice(*state, *machine)
		return e
	case "complete":
		var r freshpc.Request
		var reply freshpc.Response
		if readJSON(filepath.Join(*state, "request.json"), &r) != nil || readJSON(filepath.Join(*state, "response.json"), &reply) != nil {
			return errors.New("RESPONSE_INVALID")
		}
		return freshpc.Complete(*state, r, reply, *pin)
	case "serve":
		return serve(*ip, *pin)
	case "probe":
		return probe(*state, *ip)
	case "generate-ca":
		if checkHost(*ip) != nil {
			return errors.New("ROOT_VPS_PREFLIGHT_FAILED")
		}
		return freshpc.GenerateCA(*out, *ip)
	case "migration-probe":
		return migrationProbe(*ip)
	}
	return errors.New("unknown mode")
}
func migrationProbe(ip string) (err error) {
	ca, cak, signer, trust, store, e := rootState(ip)
	if e != nil {
		return e
	}
	defer store.DB.Close()
	dir, e := os.MkdirTemp("/var/lib/neexitlag-fresh", "probe-")
	if e != nil {
		return errors.New("PROBE_STAGE_FAILED")
	}
	defer os.RemoveAll(dir)
	slot, _ := security.NewID()
	h := fmt.Sprintf("%x", sha256.Sum256([]byte(slot)))
	r, e := freshpc.RequestDevice(dir, h)
	if e != nil {
		return e
	}
	pub, _ := freshpc.ValidateRequest(r)
	cert, e := freshpc.Issue(ca, cak, pub, "")
	if e != nil {
		return e
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if e = store.Register(ctx, cp.Device{ID: r.DeviceID, Name: "CA rotation health probe", PublicKey: base64.RawURLEncoding.EncodeToString(pub)}); e != nil {
		return errors.New("PROBE_REGISTER_FAILED")
	}
	defer func() {
		cleanup, stop := context.WithTimeout(context.Background(), 10*time.Second)
		defer stop()
		if ce := store.Revoke(cleanup, r.DeviceID); ce != nil {
			err = errors.New("PROBE_REVOKE_FAILED")
			return
		}
		if _, ce := store.Publish(cleanup, "vps-1", signer, trust); ce != nil {
			err = errors.New("PROBE_REVOCATION_PUBLICATION_FAILED")
		}
	}()
	token, e := e2eprobe.Token("/etc/neexitlag/echo/echo.token")
	if e != nil {
		return errors.New("PROBE_TOKEN_UNAVAILABLE")
	}
	response := freshpc.Response{DeviceID: r.DeviceID, MachineHash: r.MachineHash, CA: string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: ca.Raw})), CASHA256: freshpc.Fingerprint(ca), Certificate: cert, Trust: trust, EchoToken: string(token)}
	raw, _ := json.Marshal(response)
	if e = os.WriteFile(filepath.Join(dir, "response.json"), raw, 0600); e != nil {
		return e
	}
	return probe(dir, ip)
}

func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, "FRESH_PC_FAILED:", e)
		os.Exit(1)
	}
}
