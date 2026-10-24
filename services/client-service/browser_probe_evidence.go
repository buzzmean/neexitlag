package clientservice

import (
	"encoding/json"
	"github.com/buzzmean/neexitlag/core/browserproxy"
	"github.com/buzzmean/neexitlag/core/localstrategy"
	wp "github.com/buzzmean/neexitlag/platform/windows"
	"net"
	"net/netip"
	"sort"
	"sync/atomic"
	"time"
)

type localProbeEvidence struct {
	Failure        string                    `json:"failure,omitempty"`
	HTTPStatus     int                       `json:"http_status"`
	TLSVersion     uint16                    `json:"tls_version"`
	PlannedActions []localstrategy.Action    `json:"planned_actions,omitempty"`
	Sent           uint64                    `json:"sent_bytes"`
	At             time.Time                 `json:"at"`
	Strategy       string                    `json:"strategy"`
	Host           string                    `json:"host"`
	Service        string                    `json:"service"`
	Result         string                    `json:"result"`
	Phase          string                    `json:"phase"`
	Detail         string                    `json:"detail"`
	Received       uint64                    `json:"received_bytes"`
	Body           uint64                    `json:"body_bytes"`
	Committed      bool                      `json:"committed"`
	Inspection     *localstrategy.Inspection `json:"inspection,omitempty"`
}

func localProbeOutcome(p browserproxy.ProbeResult, applied bool) string {
	if p.Phase == "ORIGIN_DNS" {
		return "ORIGIN_DNS_FAILED"
	}
	if p.Phase == "LOCAL_GUARD" {
		return "LOCAL_GUARD"
	}
	if p.Failure == "" && (p.Phase == "PUBLIC_WS_CONTROL" || p.Phase == "PUBLIC_REDIRECT_CONTROL") {
		if applied {
			return "OK_CONTROL_NOT_BULK"
		}
		return "NO_TRANSFORM"
	}
	if p.Detail == "PEER_RESET" {
		return "RST"
	}
	if p.Bytes == 0 && p.Failure != "" {
		return "ORIGIN_SILENT"
	}
	if p.BodyBytes > 0 && p.Failure != "" {
		if p.Failure == "NETWORK_TIMEOUT" {
			return "FREEZE_AFTER_N"
		}
		if p.Detail == "PEER_CLOSED" {
			return "EOF_AFTER_N"
		}
		return "ORIGIN_PROTOCOL_OR_BODY_FAILED"
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
func (c *Controller) localPacketInspection(q browserproxy.Request, p browserproxy.Plan) *localstrategy.Inspection {
	if b, ok := c.Backend.(interface {
		BrowserLocalInspection(wp.ProcessIdentity, netip.AddrPort, netip.AddrPort) *localstrategy.Inspection
	}); ok {
		return b.BrowserLocalInspection(browserIdentity(q), q.Source, p.Destination)
	}
	return nil
}
func (c *Controller) recordLocalProbe(q browserproxy.Request, p browserproxy.Plan, strategy string, result browserproxy.ProbeResult, applied bool, sent ...uint64) {
	inspection := c.localPacketInspection(q, p)
	row := localProbeEvidence{Failure: result.Failure, HTTPStatus: result.Status, TLSVersion: result.TLSVersion, At: time.Now(), Strategy: strategy, Host: q.Host, Service: c.browserOriginPolicy(q.Host).Group, Result: localProbeOutcome(result, applied), Phase: result.Phase, Detail: result.Detail, Received: result.Bytes, Body: result.BodyBytes, Committed: applied, Inspection: inspection}
	if len(sent) > 0 {
		row.Sent = sent[0]
	}
	r := c.browser.youtube
	r.mu.Lock()
	defer r.mu.Unlock()
	if plan, ok := r.plans[strategy]; ok {
		row.PlannedActions = append([]localstrategy.Action(nil), plan.Actions...)
	} else {
		for _, candidate := range localstrategy.BrowserYouTubeCandidates() {
			if candidate.ID == strategy {
				row.PlannedActions = append([]localstrategy.Action(nil), candidate.Actions...)
			}
		}
	}
	if applied && youtubeDomain(q.Host) && p.Network != "" {
		for k, until := range r.qualifications {
			if !time.Now().Before(until) {
				delete(r.qualifications, k)
			}
		}
		key := youtubeQualification{p.Network, q.Owner.Profile, strategy, q.HintOnly}
		_, exists := r.qualifications[key]
		if len(r.qualifications) < 256 || exists {
			r.qualifications[key] = time.Now().Add(youtubeLearnTTL)
			if youtubeTransferOK(result) {
				r.sustained[key] = time.Now().Add(youtubeLearnTTL)
			}
			for k, until := range r.sustained {
				if !time.Now().Before(until) {
					delete(r.sustained, k)
				}
			}
		}
	}
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

type localCloseEvidence struct {
	Domain, Strategy, Reason, IOFailure string
	Timing                              youtubeTimings
	OpenMS, TransformSendMS             float64
	TX, RX                              uint64
	Updated                             time.Time
	Packets                             localstrategy.TCPTrace
}

func (c *Controller) localProbeSnapshot() map[string]any {
	r := c.browser.youtube
	r.mu.Lock()
	defer r.mu.Unlock()
	closed := []localCloseEvidence{}
	for _, v := range r.records {
		if v.Stage == "LOCAL" && v.Inspection != nil && (!v.Active || v.IOFailure != "") {
			closed = append(closed, localCloseEvidence{Domain: v.Domain, Strategy: v.Strategy, Reason: v.Reason, IOFailure: v.IOFailure, TX: v.TX, RX: v.RX, Updated: v.Updated, Packets: v.Inspection.TCPPackets, Timing: v.Timing, OpenMS: v.OpenMS, TransformSendMS: v.Inspection.TransformSendMS})
		}
	}
	sort.Slice(closed, func(i, j int) bool { return closed[i].Updated.After(closed[j].Updated) })
	if len(closed) > 8 {
		closed = closed[:8]
	}
	rows := append([]localProbeEvidence(nil), r.probeEvidence...)
	if len(rows) > 32 {
		rows = rows[len(rows)-32:]
	}
	recentFailures := r.failures
	if len(recentFailures) > 8 {
		recentFailures = recentFailures[len(recentFailures)-8:]
	}
	result := map[string]any{"failures": recentFailures, "tcp_closures": closed, "probes": rows, "retained": len(r.probeEvidence), "retention_limit": 128, "scope": "bounded recent probes and closes; headers/timing only; TLS1.3 type23 is NOT HTTP content proof"}
	for {
		result["probes"] = rows
		result["tcp_closures"] = closed
		result["omitted_retained"] = len(r.probeEvidence) - len(rows)
		raw, _ := json.Marshal(result)
		if len(raw) < 48000 {
			return result
		}
		if len(rows) > 0 {
			rows = rows[1:]
		} else if len(closed) > 0 {
			closed = closed[:len(closed)-1]
		} else {
			return result
		}
	}
}
