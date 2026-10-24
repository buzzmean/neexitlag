package clientservice

import (
	"context"
	"github.com/buzzmean/neexitlag/core/browserproxy"
	"testing"
	"time"
)

func TestCDNCanTrialWithCommittedTransformDespiteSilentPublicTLS(t *testing.T) {
	c := New(context.Background())
	defer c.Close()
	r := c.browser.youtube
	k := youtubeKey{Network: "fixture-net", Profile: "chrome", Domain: "googlevideo:sn-fixture", Port: 443}
	q := browserproxy.Request{Host: "www.youtube.com", Owner: browserproxy.Owner{Profile: k.Profile}}
	p := browserproxy.Plan{Network: k.Network}
	failed := browserproxy.ProbeResult{Failure: "NETWORK_TIMEOUT", Phase: "ORIGIN_TLS"}
	c.recordLocalProbe(q, p, "candidate", failed, false)
	if r.qualified(k, "candidate", time.Now()) {
		t.Fatal("no transform qualified")
	}
	q.Host = "vk.ru"
	c.recordLocalProbe(q, p, "candidate", failed, true)
	if r.qualified(k, "candidate", time.Now()) {
		t.Fatal("ordinary host qualified YouTube")
	}
	q.Host = "www.youtube.com"
	c.recordLocalProbe(q, p, "candidate", failed, true)
	if !r.qualified(k, "candidate", time.Now()) {
		t.Fatal("committed transform should permit guarded trial despite failed public prerequisite")
	}
	r.completeCDNQualification(k, "candidate", time.Now())
	if _, ok := r.takeTrial(k, time.Now(), false); ok {
		t.Fatal("trial without live relay sibling")
	}
	if _, ok := r.takeTrial(k, time.Now(), true); !ok {
		t.Fatal("qualified CDN trial stayed stuck")
	}
	r.finishTrial(k, "candidate", false, true, time.Now())
	if r.cache[k].Stage != "RELAY" || r.cache[k].Confirmed {
		t.Fatal("trial failure declared LOCAL")
	}
	other := k
	other.Network = "different"
	if r.qualified(other, "candidate", time.Now()) {
		t.Fatal("proof crossed network")
	}
	other = k
	other.Profile = "edge"
	if r.qualified(other, "candidate", time.Now()) {
		t.Fatal("proof crossed browser profile")
	}
	if r.qualified(k, "different-plan", time.Now()) {
		t.Fatal("proof crossed strategy")
	}
}
