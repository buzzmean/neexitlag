package clientservice

import (
	"testing"
	"time"
)

func TestCDNPeerProofSpreadsToOtherSubgroupUntilFailure(t *testing.T) {
	r := newYoutubeRuntime()
	now := time.Now()
	id := r.order[0]
	a := youtubeKey{Network: "net", Profile: "chrome", Domain: "googlevideo:sn-a", Port: 443}
	b := youtubeKey{Network: "net", Profile: "chrome", Domain: "googlevideo:sn-b", Port: 443}
	c := youtubeKey{Network: "net", Profile: "chrome", Domain: "googlevideo:sn-c", Port: 443}
	otherNet := youtubeKey{Network: "net2", Profile: "chrome", Domain: "googlevideo:sn-b", Port: 443}
	page := youtubeKey{Network: "net", Profile: "chrome", Domain: "www.youtube.com", Port: 443}
	e := cdnEndpoint{"rr1---sn-b.googlevideo.com", "1.2.3.4:443"}
	if stage, _ := r.chooseEndpoint(b, e, now); stage != "RELAY" {
		t.Fatal("no peer proof must stay RELAY")
	}
	r.mu.Lock()
	r.notePeerProofLocked(a, id, now)
	r.mu.Unlock()
	if stage, strategy := r.chooseEndpoint(b, e, now); stage != "LOCAL" || strategy != id {
		t.Fatalf("peer subgroup = %s %s", stage, strategy)
	}
	for _, k := range []youtubeKey{otherNet, page} {
		if stage, _ := r.chooseEndpoint(k, e, now); stage != "RELAY" {
			t.Fatalf("proof leaked to %+v", k)
		}
	}
	r.failed(b, "LOCAL", id, "MIDSTREAM_STALL_NEW_CONNECTION_REQUIRED", now)
	if r.selection.PeerProvisional != 1 || r.selection.PeerProvisionalFailures != 1 {
		t.Fatalf("counters %+v", r.selection)
	}
	for _, k := range []youtubeKey{b, c} {
		if stage, _ := r.chooseEndpoint(k, e, now); stage != "RELAY" {
			t.Fatalf("failure must revoke peer proof for %s", k.Domain)
		}
	}
}

func TestCDNPeerProofIgnoresBackgroundAndNonCDNFailures(t *testing.T) {
	r := newYoutubeRuntime()
	now := time.Now()
	id := r.order[0]
	a := youtubeKey{Network: "net", Profile: "chrome", Domain: "googlevideo:sn-a", Port: 443}
	b := youtubeKey{Network: "net", Profile: "chrome", Domain: "googlevideo:sn-b", Port: 443}
	page := youtubeKey{Network: "net", Profile: "chrome", Domain: "www.youtube.com", Port: 443}
	r.mu.Lock()
	r.notePeerProofLocked(a, id, now)
	r.mu.Unlock()
	r.backgroundFailure(b, id, "BACKGROUND_NOT_SUSTAINED_OR_NO_TRANSFORM")
	r.failed(page, "LOCAL", id, "MIDSTREAM_STALL_NEW_CONNECTION_REQUIRED", now)
	if stage, strategy := r.chooseEndpoint(b, cdnEndpoint{}, now); stage != "LOCAL" || strategy != id {
		t.Fatalf("background/non-CDN failure revoked peer proof: %s %s", stage, strategy)
	}
	if s := r.snapshot(); s.Selection.PeerProofs != 1 || s.Selection.PeerRevocations != 0 {
		t.Fatalf("snapshot %+v", s.Selection)
	}
}

func TestCDNPeerProofExpires(t *testing.T) {
	r := newYoutubeRuntime()
	now := time.Now()
	a := youtubeKey{Network: "net", Profile: "chrome", Domain: "googlevideo:sn-a", Port: 443}
	b := youtubeKey{Network: "net", Profile: "chrome", Domain: "googlevideo:sn-b", Port: 443}
	r.mu.Lock()
	r.notePeerProofLocked(a, r.order[0], now)
	r.mu.Unlock()
	if stage, _ := r.chooseEndpoint(b, cdnEndpoint{}, now.Add(youtubeLearnTTL+time.Second)); stage != "RELAY" {
		t.Fatal("expired peer proof must not grant LOCAL")
	}
}
