package clientservice

import (
	"encoding/json"
	"github.com/buzzmean/neexitlag/core/browserproxy"
	"github.com/buzzmean/neexitlag/core/localstrategy"
	wp "github.com/buzzmean/neexitlag/platform/windows"
	"net"
	"net/netip"
	"sync/atomic"
	"time"
)

type localProbeEvidence struct {
	Sent       uint64                    `json:"sent_bytes"`
	At         time.Time                 `json:"at"`
	Strategy   string                    `json:"strategy"`
	Host       string                    `json:"host"`
	Service    string                    `json:"service"`
	Result     string                    `json:"result"`
	Phase      string                    `json:"phase"`
	Detail     string                    `json:"detail"`
	Received   uint64                    `json:"received_bytes"`
	Body       uint64                    `json:"body_bytes"`
	Committed  bool                      `json:"committed"`
	Inspection *localstrategy.Inspection `json:"inspection,omitempty"`
}

func localProbeOutcome(p browserproxy.ProbeResult, applied bool) string {
	if p.Phase == "ORIGIN_DNS" {
		return "ORIGIN_DNS_FAILED"
	}
	if p.Phase == "LOCAL_GUARD" {
		return "LOCAL_GUARD"
	}
	if p.Detail == "PEER_RESET" {
		return "RST"
	}
	if p.Bytes == 0 && p.Failure != "" {
		return "ORIGIN_SILENT"
	}
	if p.BodyBytes > 0 && p.Failure != "" {
		return "FREEZE_AFTER_N"
	}
	if p.BodyBytes > 0 && p.BodyBytes <= 64*1024 {
		return "SHORT_RESPONSE"
	}
	if !applied {
		return "NO_TRANSFORM"
	}
	if youtubeTransferOK(p) {
		return "OK"
	}
	return "ORIGIN_RESPONSE_NOT_SUSTAINED"
}
func (c *Controller) recordLocalProbe(q browserproxy.Request, p browserproxy.Plan, strategy string, result browserproxy.ProbeResult, applied bool, sent ...uint64) {
	var inspection *localstrategy.Inspection
	if b, ok := c.Backend.(interface {
		BrowserLocalInspection(wp.ProcessIdentity, netip.AddrPort, netip.AddrPort) *localstrategy.Inspection
	}); ok {
		inspection = b.BrowserLocalInspection(browserIdentity(q), q.Source, p.Destination)
	}
	row := localProbeEvidence{At: time.Now(), Strategy: strategy, Host: q.Host, Service: c.browserOriginPolicy(q.Host).Group, Result: localProbeOutcome(result, applied), Phase: result.Phase, Detail: result.Detail, Received: result.Bytes, Body: result.BodyBytes, Committed: applied, Inspection: inspection}
	if len(sent) > 0 {
		row.Sent = sent[0]
	}
	r := c.browser.youtube
	r.mu.Lock()
	defer r.mu.Unlock()
	r.probeEvidence = append(r.probeEvidence, row)
	if len(r.probeEvidence) > 128 {
		r.probeEvidence = append([]localProbeEvidence(nil), r.probeEvidence[len(r.probeEvidence)-128:]...)
	}
}

type localProbeConn struct {
	net.Conn
	sent atomic.Uint64
}

func (c *localProbeConn) Write(p []byte) (int, error) {
	n, e := c.Conn.Write(p)
	c.sent.Add(uint64(n))
	return n, e
}

func (c *Controller) localProbeSnapshot() map[string]any {
	r := c.browser.youtube
	r.mu.Lock()
	defer r.mu.Unlock()
	rows := append([]localProbeEvidence(nil), r.probeEvidence...)
	if len(rows) > 32 {
		rows = rows[len(rows)-32:]
	}
	for len(rows) > 0 {
		raw, _ := json.Marshal(rows)
		if len(raw) < 48000 {
			break
		}
		rows = rows[1:]
	}
	return map[string]any{"probes": rows, "retained": len(r.probeEvidence), "omitted_retained": len(r.probeEvidence) - len(rows), "retention_limit": 128, "scope": "bounded recent probes; packet headers only; no payload; missing evidence is not success"}
}
