package clientservice

import (
	"context"
	"encoding/json"
	"github.com/buzzmean/neexitlag/core/browserproxy"
	"github.com/buzzmean/neexitlag/core/localstrategy"
	"strings"
	"testing"
)

func TestLocalProbeFailuresAreNotAllNoTransform(t *testing.T) {
	for _, tc := range []struct {
		p       browserproxy.ProbeResult
		applied bool
		want    string
	}{
		{browserproxy.ProbeResult{Failure: "NETWORK_TIMEOUT"}, false, "ORIGIN_SILENT"},
		{browserproxy.ProbeResult{Detail: "PEER_RESET"}, true, "RST"},
		{browserproxy.ProbeResult{BodyBytes: 16000, Bytes: 18000, Failure: "NETWORK_TIMEOUT"}, true, "FREEZE_AFTER_N"},
		{browserproxy.ProbeResult{Status: 200, TLSVersion: 0x304, BodyBytes: 70000}, false, "NO_TRANSFORM"},
		{browserproxy.ProbeResult{Bytes: 1000, BodyBytes: 200, Failure: "WRONG_WEBSOCKET_NONCE"}, true, "ORIGIN_PROTOCOL_OR_BODY_FAILED"},
		{browserproxy.ProbeResult{Bytes: 1000, BodyBytes: 200, Failure: "NETWORK_ERROR", Detail: "PEER_CLOSED"}, true, "EOF_AFTER_N"},
		{browserproxy.ProbeResult{Status: 200, TLSVersion: 0x304, BodyBytes: 70000}, true, "OK"},
	} {
		if got := localProbeOutcome(tc.p, tc.applied); got != tc.want {
			t.Fatal(got, tc.want)
		}
	}
}

func TestLocalProbeSnapshotBoundedBelowIPC(t *testing.T) {
	c := New(context.Background())
	defer c.Close()
	for i := 0; i < 128; i++ {
		c.browser.youtube.probeEvidence = append(c.browser.youtube.probeEvidence, localProbeEvidence{Host: strings.Repeat("x", 253), Strategy: strings.Repeat("s", 96), Inspection: &localstrategy.Inspection{Detail: strings.Repeat("d", 1024), SentPackets: make([]localstrategy.SentPacket, 20)}})
	}
	raw, e := json.Marshal(c.localProbeSnapshot())
	if e != nil || len(raw) >= 49000 {
		t.Fatal(len(raw), e)
	}
}
