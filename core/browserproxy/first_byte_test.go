package browserproxy

import (
	"fmt"
	"github.com/buzzmean/neexitlag/core/flowpolicy"
	"testing"
	"time"
)

func TestFirstByteSeparatesGroupNetworkProfileModeAndFailures(t *testing.T) {
	var ledger firstByteLedger
	r := Record{Domain: "www.youtube.com", Owner: Owner{Profile: "chrome"}, Plan: Plan{Network: "net-a"}, EvidenceSource: "PAC_CONNECT", FirstByteMode: "DIRECT_FIRST"}
	ledger.observe(r, 2100*time.Millisecond)
	r.Domain = "rr1---sn-any.googlevideo.com"
	r.FirstByteMode = "RELAY_FIRST"
	ledger.observe(r, 55*time.Millisecond)
	ledger.observe(r, 65*time.Millisecond)
	ledger.observe(r, -1) // Failure is not a zero-ms success.
	r.Plan.Network = "net-b"
	ledger.observe(r, 99*time.Millisecond)
	r.Owner.Profile = "edge"
	ledger.observe(r, 100*time.Millisecond)
	rows, lost := ledger.snapshot()
	if len(rows) != 4 || lost != 0 {
		t.Fatal(rows, lost)
	}
	for _, m := range rows {
		if m.Group != "youtube-media" {
			t.Fatal(m)
		}
		if m.Network == "net-a" && m.Mode == "RELAY_FIRST" && (m.Replies != 2 || m.NoReply != 1 || m.TotalMS != 120 || m.MinMS != 55 || m.MaxMS != 65) {
			t.Fatal(m)
		}
	}
	r.FirstByteMode = "RELAY_FIRST"
	r.Plan.DirectAttempts = 1
	r.Plan.Policy.Action = flowpolicy.RelayRequired
	if firstByteMode(r.Plan) != "DIRECT_FIRST" {
		t.Fatal("retry incorrectly called relay-first")
	}
}
func TestFirstByteEvidenceIsBoundedAndUnknownSamplesAreNotInvented(t *testing.T) {
	var ledger firstByteLedger
	rows, lost := ledger.snapshot()
	if len(rows) != 0 || lost != 0 {
		t.Fatal(rows, lost)
	}
	for i := 0; i < 140; i++ {
		ledger.observe(Record{Domain: fmt.Sprintf("site-%d.example", i)}, -1)
	}
	rows, lost = ledger.snapshot()
	if len(rows) != 128 || lost != 12 {
		t.Fatal(len(rows), lost)
	}
	for _, m := range rows {
		if m.Replies != 0 || m.NoReply != 1 || m.MinMS != -1 || m.LastMS != -1 {
			t.Fatal("failure fabricated latency", m)
		}
	}
}
