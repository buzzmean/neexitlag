package browserproxy

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"
)

// ProbeOrigin checks TLS certificate validation and an actual origin HTTP
// response, without cookies, credentials, URL path/query or browser page content.
// Success is ONLY origin reachability, never a claim about logged-in application.
func ProbeOrigin(ctx context.Context, host string, port uint16, secure bool, roots *x509.CertPool, open func(context.Context) (net.Conn, error)) (int, uint64, string) {
	conn, err := open(ctx)
	if err != nil {
		return 0, 0, "CONNECTION_FAILED"
	}
	defer conn.Close()
	deadline := time.Now().Add(1500 * time.Millisecond)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	conn.SetDeadline(deadline)
	if secure {
		tc := tls.Client(conn, &tls.Config{ServerName: host, RootCAs: roots, MinVersion: tls.VersionTLS12, NextProtos: []string{"http/1.1"}})
		if err = tc.HandshakeContext(ctx); err != nil {
			var certificate *tls.CertificateVerificationError
			if errors.As(err, &certificate) {
				return 0, 0, "TLS_CERTIFICATE_ERROR"
			}
			if errors.Is(err, context.DeadlineExceeded) {
				return 0, 0, "NETWORK_TIMEOUT"
			}
			return 0, 0, "TLS_FAILED"
		}
		conn = tc
	}
	authority := host
	if port != 443 && secure || port != 80 && !secure {
		authority = net.JoinHostPort(host, strconv.Itoa(int(port)))
	}
	request := &http.Request{Method: "HEAD", URL: nil, Host: authority, Header: http.Header{"User-Agent": []string{"NeExitLag-Origin-Probe/1"}, "Connection": []string{"close"}}}
	// Construct the fixed request; no user supplied path, headers or body.
	if _, err = io.WriteString(conn, "HEAD / HTTP/1.1\r\nHost: "+authority+"\r\nUser-Agent: NeExitLag-Origin-Probe/1\r\nConnection: close\r\n\r\n"); err != nil {
		return 0, 0, "CONNECTION_FAILED"
	}
	limited := &io.LimitedReader{R: conn, N: 32768}
	response, err := http.ReadResponse(bufio.NewReaderSize(limited, 8192), request)
	if err != nil {
		if e, ok := err.(net.Error); ok && e.Timeout() {
			return 0, 0, "NETWORK_TIMEOUT"
		}
		return 0, 0, "CONNECTION_FAILED"
	}
	defer response.Body.Close()
	return response.StatusCode, uint64(32768 - limited.N), ""
}
