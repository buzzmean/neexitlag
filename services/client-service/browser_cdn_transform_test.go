package clientservice

import (
	"context"
	"github.com/buzzmean/neexitlag/core/browserproxy"
	"net/netip"
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
	if _, ok := r.takeTrial(k, time.Now(), testSibling(false)); ok {
		t.Fatal("trial without live relay sibling")
	}
	if _, ok := r.takeTrial(k, time.Now(), testSibling(true)); !ok {
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

func TestCDNBackgroundSelectsExactSustainedPublicProofInsteadOfCatalogPrior(t *testing.T) {
	c, cur, q, p := youtubeFixture(t)
	q.Source = netip.MustParseAddrPort("192.0.2.2:51000")
	r := c.browser.youtube
	k := youtubeScope(cur, q)
	weak, strong := r.order[0], r.order[1]
	public := q
	public.Host = "www.youtube.com"
	p.Network = cur.network
	c.recordLocalProbe(public, p, weak, browserproxy.ProbeResult{Failure: "NETWORK_TIMEOUT"}, true)
	c.recordLocalProbe(public, p, strong, browserproxy.ProbeResult{Status: 200, TLSVersion: 0x304, BodyBytes: 131073}, true)
	r.completeCDNQualification(k, weak, time.Now())
	c.youtubeBackgroundRecheck(cur, q, p, k)
	id, ok := r.takeTrial(k, time.Now(), testSibling(true))
	if !ok || id != strong {
		t.Fatal("public proof was ignored", id, strong, r.cache[k])
	}
	r.finishTrial(k, id, false, false, time.Now())
}
