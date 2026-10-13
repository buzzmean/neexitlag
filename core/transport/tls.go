package transport

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"github.com/quic-go/quic-go"
	"os"
	"time"
)

const ALPN = "neexitlag/1"

func TLS(cert, key, ca string, server bool) (*tls.Config, error) {
	pair, e := tls.LoadX509KeyPair(cert, key)
	if e != nil {
		return nil, e
	}
	b, e := os.ReadFile(ca)
	if e != nil {
		return nil, e
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(b) {
		return nil, errors.New("invalid CA")
	}
	c := &tls.Config{MinVersion: tls.VersionTLS13, Certificates: []tls.Certificate{pair}, NextProtos: []string{ALPN}, RootCAs: pool}
	if server {
		c.ClientCAs = pool
		c.ClientAuth = tls.RequireAndVerifyClientCert
	}
	return c, nil
}
func QUICConfig() *quic.Config {
	return &quic.Config{EnableDatagrams: true, Allow0RTT: false, HandshakeIdleTimeout: 3 * time.Second, MaxIdleTimeout: 30 * time.Second, KeepAlivePeriod: 10 * time.Second, MaxIncomingStreams: 64, MaxIncomingUniStreams: -1, InitialStreamReceiveWindow: 64 * 1024, MaxStreamReceiveWindow: 256 * 1024, InitialConnectionReceiveWindow: 256 * 1024, MaxConnectionReceiveWindow: 1024 * 1024}
}
