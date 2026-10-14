// Ordinary sockets only. No relay SDK, explicit proxy or packet driver in client.
package main

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/buzzmean/neexitlag/internal/webprobe"
	windowsplatform "github.com/buzzmean/neexitlag/platform/windows"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Attempt struct {
	Started     time.Time `json:"started"`
	Finished    time.Time `json:"finished"`
	Destination string    `json:"destination"`
	Source      string    `json:"response_source,omitempty"`
	Protocol    string    `json:"protocol,omitempty"`
	TLS         bool      `json:"tls"`
	Success     bool      `json:"success"`
	Error       string    `json:"error,omitempty"`
	DNS         []string  `json:"dns,omitempty"`
}
type Report struct {
	PID          uint32    `json:"pid"`
	Exe          string    `json:"exe"`
	CreationTime uint64    `json:"creation_time"`
	Started      time.Time `json:"measurement_started"`
	Finished     time.Time `json:"measurement_finished"`
	Attempts     []Attempt `json:"attempts"`
	Error        string    `json:"error,omitempty"`
}

func probe(ctx context.Context, endpoint, token, caFile, mode string) (a Attempt) {
	a.Started = time.Now()
	defer func() { a.Finished = time.Now() }()
	// Named return is required for the deferred finished timestamp.
	u, e := url.Parse(endpoint)
	if e != nil {
		a.Error = e.Error()
		return
	}
	if mode == "websocket" {
		u.Path = "/ws"
	} else {
		u.Path = "/observe"
	}
	nonceBytes := make([]byte, 16)
	if _, e = rand.Read(nonceBytes); e != nil {
		a.Error = e.Error()
		return
	}
	nonce := hex.EncodeToString(nonceBytes)
	q := u.Query()
	q.Set("nonce", nonce)
	u.RawQuery = q.Encode()
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
	if caFile != "" {
		pem, err := os.ReadFile(caFile)
		if err != nil {
			a.Error = err.Error()
			return
		}
		roots, err := x509.SystemCertPool()
		if err != nil {
			roots = x509.NewCertPool()
		}
		if !roots.AppendCertsFromPEM(pem) {
			a.Error = "invalid endpoint CA"
			return
		}
		tlsConfig.RootCAs = roots
	}
	tr := &http.Transport{Proxy: nil, TLSClientConfig: tlsConfig, ForceAttemptHTTP2: mode == "http2", DisableKeepAlives: true, DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "tcp4", addr)
	}}
	defer tr.CloseIdleConnections()
	trace := &httptrace.ClientTrace{GotConn: func(i httptrace.GotConnInfo) { a.Destination = i.Conn.RemoteAddr().String() }, DNSDone: func(i httptrace.DNSDoneInfo) {
		for _, v := range i.Addrs {
			a.DNS = append(a.DNS, v.String())
		}
	}}
	request, e := http.NewRequestWithContext(httptrace.WithClientTrace(ctx, trace), "GET", u.String(), nil)
	if e != nil {
		a.Error = e.Error()
		return
	}
	request.Header.Set("X-Neexitlag-Proof", webprobe.SignRequest(token, "GET", u.Path, nonce))
	wsKey := ""
	if mode == "websocket" {
		wsKey = base64.StdEncoding.EncodeToString(nonceBytes)
		request.Header.Set("Upgrade", "websocket")
		request.Header.Set("Connection", "Upgrade")
		request.Header.Set("Sec-WebSocket-Key", wsKey)
		request.Header.Set("Sec-WebSocket-Version", "13")
	}
	response, e := (&http.Client{Transport: tr, Timeout: 8 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}).Do(request)
	if e != nil {
		a.Error = e.Error()
		return
	}
	defer response.Body.Close()
	var observation webprobe.Observation
	if mode == "websocket" {
		sum := sha1.Sum([]byte(wsKey + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
		if response.StatusCode != 101 || response.Header.Get("Sec-WebSocket-Accept") != base64.StdEncoding.EncodeToString(sum[:]) {
			a.Error = "invalid WebSocket upgrade"
			return
		}
		observation, e = readWebSocket(response.Body)
	} else {
		if response.StatusCode != 200 {
			a.Error = fmt.Sprintf("HTTP %d", response.StatusCode)
			return
		}
		observation, e = webprobe.ReadObservation(response.Body)
	}
	if e != nil {
		a.Error = e.Error()
		return
	}
	if observation.Nonce != nonce || net.ParseIP(observation.Source) == nil || !webprobe.VerifyObservation(token, observation) {
		a.Error = "invalid destination nonce/source"
		return
	}
	a.Source = observation.Source
	a.Protocol = observation.Protocol
	a.TLS = response.TLS != nil
	if u.Scheme == "https" && (!a.TLS || !observation.TLS) {
		a.Error = "TLS not established"
		return
	}
	if mode == "http2" && (response.ProtoMajor != 2 || observation.Protocol != "HTTP/2.0") {
		a.Error = "HTTP/2 not negotiated"
		return
	}
	a.Success = true
	return
}
func readWebSocket(r io.Reader) (webprobe.Observation, error) {
	var o webprobe.Observation
	var header [2]byte
	if _, e := io.ReadFull(r, header[:]); e != nil {
		return o, e
	}
	if header[0] != 0x81 || header[1]&128 != 0 {
		return o, errors.New("invalid WebSocket frame")
	}
	n := int(header[1] & 127)
	if n == 126 {
		var b [2]byte
		if _, e := io.ReadFull(r, b[:]); e != nil {
			return o, e
		}
		n = int(binary.BigEndian.Uint16(b[:]))
	}
	if n > 4096 || n == 127 {
		return o, errors.New("oversized frame")
	}
	b := make([]byte, n)
	_, e := io.ReadFull(r, b)
	if e == nil {
		e = json.Unmarshal(b, &o)
	}
	return o, e
}
func main() {
	endpoint := flag.String("url", "", "controlled HTTP/HTTPS endpoint URL")
	tokenFile := flag.String("token-file", "", "endpoint token file")
	ca := flag.String("ca", "", "destination CA PEM (normal certificate verification)")
	mode := flag.String("mode", "http", "http, http2, websocket")
	count := flag.Int("count", 30, "connections, 1..200")
	interval := flag.Duration("interval", 200*time.Millisecond, "between connections")
	report := flag.String("report", "", "report JSON file")
	flag.Parse()
	r := Report{PID: uint32(os.Getpid()), Started: time.Now(), Attempts: []Attempt{}}
	r.Exe, _ = os.Executable()
	r.Exe, _ = filepath.Abs(r.Exe)
	if id, e := windowsplatform.NewDiscovery().Lookup(r.PID); e == nil {
		r.CreationTime = id.CreationTime
		r.Exe = id.Path
	}
	u, e := url.Parse(*endpoint)
	raw, tokenErr := os.ReadFile(*tokenFile)
	if e != nil || u == nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || (*mode != "http" && *mode != "http2" && *mode != "websocket") || *count < 1 || *count > 200 || *interval < 0 || *interval > time.Minute || tokenErr != nil || len(strings.TrimSpace(string(raw))) < 24 {
		r.Error = "invalid URL/mode/count/interval/token-file"
	} else {
		for i := 0; i < *count; i++ {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			a := probe(ctx, *endpoint, strings.TrimSpace(string(raw)), *ca, *mode)
			cancel()
			r.Attempts = append(r.Attempts, a)
			if !a.Success {
				r.Error = "one or more connections failed"
			}
			if i+1 < *count {
				time.Sleep(*interval)
			}
		}
	}
	r.Finished = time.Now()
	b, _ := json.MarshalIndent(r, "", "  ")
	if *report != "" {
		if e := os.WriteFile(*report, b, 0600); e != nil {
			fmt.Fprintln(os.Stderr, e)
			os.Exit(1)
		}
	}
	fmt.Println(string(b))
	if r.Error != "" {
		os.Exit(1)
	}
}
