package clientservice

import (
	"github.com/buzzmean/neexitlag/core/localstrategy"
	"io"
	"testing"
	"time"
)

func TestFailedCallbackEndsCDNTrialWithoutReadOrClose(t *testing.T) {
	m, r, _, _ := trialMeterFixture(t, localstrategy.TCPTrace{}, io.EOF)
	m.rx = 203
	m.failure("MIDSTREAM_STALL_NEW_CONNECTION_REQUIRED")
	if r.snapshot().Selection.ActiveTrial || r.selection.TrialFailures != 1 {
		t.Fatalf("hard terminal callback still holds max1: %+v", r.selection)
	}
}

func TestCDNTrialNeedsBulkSameHostOrClusterWithoutRRClassifier(t *testing.T) {
	u := &browserRelayUsage{}
	now := time.Now()
	add := func(host string, n uint64) {
		id := u.beginVideo(videoPerformance{Profile: "chrome", Network: "net", Port: 443, Domain: host, Started: now})
		u.mu.Lock()
		u.videoIO(id, 1000, n, now)
		u.mu.Unlock()
	}
	add("r1.googlevideo.com", 10*cdnStrongBytes)
	if u.mediaTrialSibling("chrome", "net", "r2.googlevideo.com", now).MediaBytes != 0 {
		t.Fatal("unclassified is not a proven cluster")
	}
	host := "rr5---sn-media.googlevideo.com"
	add("rr6---sn-other.googlevideo.com", 10*cdnStrongBytes)
	add(host, 7000)
	peer := u.mediaTrialSibling("chrome", "net", host, now)
	if !peer.Eligible || !peer.MediaRequired || peer.MediaBytes != 0 {
		t.Fatal("unrelated bulk/control guessed segment", peer)
	}
	_, r, k, id := trialMeterFixture(t, localstrategy.TCPTrace{}, io.EOF)
	r.finishTrialOutcome(k, id, "INCONCLUSIVE", false, now.Add(-youtubeTrialCooldown), 0)
	endpoint := cdnEndpoint{host, "74.125.8.99:443"}
	if _, _, ok := r.takeEndpointTrial(k, now, peer, endpoint); ok || r.cache[k].TrialAdmission != "SCOPED_BULK_RELAY_NOT_OBSERVED" {
		t.Fatal(r.cache[k])
	}
	add(host, cdnStrongBytes+1) // rr5 itself can qualify; no rrN blacklist.
	peer = u.mediaTrialSibling("chrome", "net", host, now)
	if peer.MediaMatch != "EXACT_HOST" || peer.MediaBytes <= cdnStrongBytes {
		t.Fatal(peer)
	}
	if _, _, ok := r.takeEndpointTrial(k, now, peer, endpoint); !ok {
		t.Fatal("bulk host not admitted", r.cache[k])
	}
	if u.mediaTrialSibling("edge", "net", host, now).MediaBytes != 0 || u.mediaTrialSibling("chrome", "net", host, now.Add(cdnSiblingTTL+time.Millisecond)).MediaBytes != 0 {
		t.Fatal("bulk proof crossed scope/TTL")
	}
	clusterPeer := u.mediaTrialSibling("chrome", "net", "rr1---sn-media.googlevideo.com", now)
	if clusterPeer.MediaMatch != "SAME_SUBGROUP" {
		t.Fatal(clusterPeer)
	}
}

func TestCDNActiveIdentityDistinguishesFastRetryOfSameSubgroup(t *testing.T) {
	m, r, k, id := trialMeterFixture(t, localstrategy.TCPTrace{RemoteFIN: 1}, io.EOF)
	old := r.trialID
	m.trialID = old
	m.host = "rr4---sn-media.googlevideo.com"
	m.destination = "74.125.8.99:443"
	m.failure("REMOTE_EOF_AFTER_REPLY")
	if r.snapshot().Selection.ActiveTrial {
		t.Fatal("neutral still active")
	}
	peer := testSibling(true)
	peer.MediaRequired = true
	peer.MediaBytes = cdnStrongBytes + 1
	endpoint := cdnEndpoint{"rr5---sn-media.googlevideo.com", "74.125.8.100:443"}
	_, token, ok := r.takeEndpointTrial(k, time.Now(), peer, endpoint)
	if !ok || token == old {
		t.Fatal("retry identity not unique")
	}
	r.trialRecord = 2
	r.records[2] = youtubeRecord{cdnTrialSlot: cdnTrialSlot{TrialID: token}, ID: 2, Domain: endpoint.Host, Destination: endpoint.Address, Active: true, TrialResult: "PROBE_WAITING_DATA"}
	r.finishTrialOutcomeFor(old, k, id, "REMOTE_SMALL_EOF", true, time.Now(), 6735)
	m.Close()
	s := r.snapshot().Selection
	if !s.ActiveTrial || s.ActiveTrialID != token || s.ActiveRecordID != 2 || s.ActiveHost != endpoint.Host || s.ActiveDestination != endpoint.Address || s.ActiveRecordResult != "PROBE_WAITING_DATA" {
		t.Fatal("old terminal confused with new exposure", s)
	}
}

func TestCDNStallQuarantinesOnlyHostIPStrategy(t *testing.T) {
	m, r, k, id := trialMeterFixture(t, localstrategy.TCPTrace{}, io.EOF)
	e := cdnEndpoint{"rr5---sn-media.googlevideo.com", "74.125.8.99:443"}
	r.trialEndpoint = e
	m.trialID = r.trialID
	m.host = e.Host
	m.destination = e.Address
	m.rx = 203
	m.failure("MIDSTREAM_STALL_NEW_CONNECTION_REQUIRED")
	v := r.cache[k]
	if r.trialActive || v.Qualified != id || v.Tried != 0 || r.selection.TrialFailures != 1 || r.selection.FailedAfterReply != 1 || v.TrialAfter.Before(time.Now().Add(2*time.Minute)) {
		t.Fatal("one endpoint poisoned subgroup", v, r.selection)
	}
	v.Stage = "LOCAL"
	v.Strategy = id
	v.Confirmed = true
	v.Until = time.Now().Add(youtubeLearnTTL)
	r.cache[k] = v
	if stage, _ := r.chooseEndpoint(k, e, time.Now()); stage != "RELAY" {
		t.Fatal("bad endpoint retried")
	}
	for _, other := range []cdnEndpoint{{e.Host, "74.125.8.100:443"}, {"rr6---sn-media.googlevideo.com", e.Address}} {
		if stage, _ := r.chooseEndpoint(k, other, time.Now()); stage != "LOCAL" {
			t.Fatal("healthy peer quarantined", other)
		}
	}
	if stage, _ := r.chooseEndpoint(k, e, time.Now().Add(youtubeLearnTTL+time.Second)); stage == "LOCAL" {
		t.Fatal("group TTL expired")
	}
	r.mu.Lock()
	bad := r.endpointBlockedLocked(k, id, e, time.Now().Add(youtubeLearnTTL+time.Second))
	r.mu.Unlock()
	if bad {
		t.Fatal("permanent endpoint pin")
	}
}
