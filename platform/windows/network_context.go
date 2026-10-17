package windowsplatform

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

type NetworkContext struct {
	ObservedAt  time.Time `json:"observed_at"`
	Fingerprint string    `json:"fingerprint"`
	Adapters    []string  `json:"adapters"`
	Routes      []string  `json:"routes"`
	DNS         []string  `json:"dns"`
	Proxy       string    `json:"proxy"`
	SecureDNS   string    `json:"secure_dns"`
	Conflicts   []string  `json:"conflicts"`
	VPN         []string  `json:"vpn"`
	Error       string    `json:"error,omitempty"`
	Egress      string    `json:"egress"`
}

func InspectNetwork(ctx context.Context) NetworkContext {
	n := inspectNetwork(ctx)
	if n.Error != "" {
		return n
	}
	b, _ := json.Marshal(struct {
		A, R, D, V []string
		P          string
	}{n.Adapters, n.Routes, n.DNS, n.VPN, n.Proxy})
	s := sha256.Sum256(b)
	n.Fingerprint = hex.EncodeToString(s[:])
	n.ObservedAt = time.Now()
	n.Egress = "DIRECT follows current Windows routing (may be VPN); physical/VPN/relay public source requires endpoint proof"
	return n
}
