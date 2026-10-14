package windowsplatform

import (
	"encoding/binary"
	"github.com/buzzmean/neexitlag/core/routing"
	"net/netip"
	"testing"
	"time"
)

func TestIPv6NetworkGuardParsingAndIdentityIsolation(t *testing.T) {
	p := make([]byte, 60)
	p[0] = 0x60
	p[6] = 6
	binary.BigEndian.PutUint16(p[4:6], 20)
	s, d := netip.MustParseAddr("2001:db8::1").As16(), netip.MustParseAddr("2001:db8::2").As16()
	copy(p[8:24], s[:])
	copy(p[24:40], d[:])
	binary.BigEndian.PutUint16(p[40:42], 50001)
	binary.BigEndian.PutUint16(p[42:44], 443)
	p[52] = 0x50
	p[53] = 2
	tuple, flags, e := ParseIPv6Transport(p)
	if e != nil || tuple.Protocol != "tcp" || flags != 2 {
		t.Fatal(tuple, flags, e)
	}
	id := ProcessIdentity{PID: 42, CreationTime: 1, Path: "chrome.exe"}
	relay := Decision{Route: routing.Route{RelayID: "VPS-01"}}
	now := time.Now()
	g := &IPv6Guard{}
	if !g.Block(id, tuple, flags, relay, 1, now) {
		t.Fatal("selected new IPv6 passed")
	}
	id.PID = 43
	if g.Block(id, tuple, flags, Decision{}, 1, now) {
		t.Fatal("stale PID blocked unselected app")
	}
	fresh := &IPv6Guard{}
	id.PID = 42
	if fresh.Block(id, tuple, 16, relay, 1, now) {
		t.Fatal("existing TCP retargeted")
	}
	if (&IPv6Guard{}).Block(id, tuple, 2, relay, 42, now) {
		t.Fatal("service blocked")
	}
	p[6] = 44
	if _, _, e := ParseIPv6Transport(p); e == nil {
		t.Fatal("fragment mistaken for TCP")
	}
	g.Cleanup(now.Add(3 * time.Minute))
	if len(g.pins) != 0 {
		t.Fatal("IPv6 pins leaked")
	}
}
