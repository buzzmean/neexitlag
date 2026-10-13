package windowsplatform

import (
	"encoding/binary"
	"errors"
	"github.com/buzzmean/neexitlag/core/routing"
	"net/netip"
	"testing"
	"time"
)

type discoveryFixture struct{ identity ProcessIdentity }

func (d *discoveryFixture) Lookup(uint32) (ProcessIdentity, error) {
	if d.identity.PID == 0 {
		return ProcessIdentity{}, errors.New("exited")
	}
	return d.identity, nil
}
func (d *discoveryFixture) List() ([]ProcessIdentity, error) {
	return []ProcessIdentity{d.identity}, nil
}
func TestCachePIDReuseAndCleanup(t *testing.T) {
	d := &discoveryFixture{ProcessIdentity{PID: 10, Path: `C:\A.exe`, CreationTime: 1}}
	c := NewProcessCache(d, func(i ProcessIdentity) bool { return i.CreationTime == d.identity.CreationTime })
	a, e := c.Lookup(10)
	if e != nil || a.CreationTime != 1 {
		t.Fatal(a, e)
	}
	d.identity = ProcessIdentity{PID: 10, Path: `C:\B.exe`, CreationTime: 2}
	b, e := c.Lookup(10)
	if e != nil || b.Path != d.identity.Path || b.CreationTime != 2 {
		t.Fatal("PID reuse", b, e)
	}
	c.Cleanup(time.Now().Add(3 * time.Minute))
	if len(c.entries) != 0 {
		t.Fatal("idle cache not cleaned")
	}
	d.identity = ProcessIdentity{}
	if _, e = c.Lookup(10); e == nil {
		t.Fatal("exited process reused")
	}
}
func TestUDPOwnerAmbiguityAndWildcard(t *testing.T) {
	row := func(ip [4]byte, pid uint32) []byte {
		b := make([]byte, 12)
		copy(b, ip[:])
		binary.BigEndian.PutUint16(b[4:6], 55000)
		binary.LittleEndian.PutUint32(b[8:], pid)
		return b
	}
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, 1)
	b = append(b, row([4]byte{}, 42)...)
	source := netip.MustParseAddrPort("192.0.2.9:55000")
	if pid, e := udpOwner(b, source); e != nil || pid != 42 {
		t.Fatal(pid, e)
	}
	binary.LittleEndian.PutUint32(b, 2)
	b = append(b, row([4]byte{192, 0, 2, 9}, 43)...)
	if _, e := udpOwner(b, source); e == nil {
		t.Fatal("ambiguous owner routed")
	}
	if _, e := udpOwner(b[:7], source); e == nil {
		t.Fatal("truncated owner table")
	}
}
func TestServiceExclusionAndEvidence(t *testing.T) {
	id := ProcessIdentity{PID: 20, Path: `C:\service.exe`, CreationTime: 99}
	tuple := Tuple{Protocol: "udp", Source: netip.MustParseAddrPort("192.0.2.1:55000"), Destination: netip.MustParseAddrPort("198.51.100.2:4433")}
	called := false
	d, bypass := ClassifyCapture(id, tuple, 20, func(ProcessIdentity, Tuple) Decision {
		called = true
		return Decision{Route: routing.Route{RelayID: "bad"}}
	})
	if called || !bypass || !d.Route.Direct() {
		t.Fatal("service entered relay classification")
	}
	l := NewEvidenceLedger()
	l.Record(id, tuple, "captured", "")
	l.Record(id, tuple, "self", "")
	l.Record(id, tuple, "direct", "")
	l.Record(ProcessIdentity{}, Tuple{}, "reinject", "")
	s := l.Snapshot(20)
	if s.Counters.Captured != 1 || s.Counters.LoopPrevented != 1 || s.Counters.Reinject != 1 || len(s.Flows) != 1 || s.Flows[0].Route != "DIRECT" {
		t.Fatal(s)
	}
	id.CreationTime = 100
	l.Record(id, tuple, "relay", "FRA-01")
	if len(l.Snapshot(20).Flows) != 2 {
		t.Fatal("PID reuse evidence merged")
	}
	l.Cleanup(time.Now().Add(6 * time.Minute))
	if len(l.Snapshot(0).Flows) != 0 || l.Snapshot(0).Counters.Relayed != 1 {
		t.Fatal("evidence cleanup changed totals")
	}
}
func TestIPv4OwnershipThroughDualStackAndCrossFamilyAmbiguity(t *testing.T) {
	source := netip.MustParseAddrPort("192.0.2.9:55000")
	b := make([]byte, 32)
	binary.LittleEndian.PutUint32(b, 1)
	binary.BigEndian.PutUint16(b[24:26], source.Port())
	binary.LittleEndian.PutUint32(b[28:32], 42)
	owners, e := udpOwners(b, source, true)
	if e != nil {
		t.Fatal(e)
	}
	if pid, e := uniqueUDPOwner(owners); e != nil || pid != 42 {
		t.Fatal("dual-stack wildcard owner", pid, e)
	}
	owners[43] = true
	if _, e = uniqueUDPOwner(owners); e == nil {
		t.Fatal("cross-family port ambiguity accepted")
	}
	ip := netip.MustParseAddr("::ffff:192.0.2.9").As16()
	copy(b[4:20], ip[:])
	owners, e = udpOwners(b, source, true)
	if e != nil || !owners[42] {
		t.Fatal("mapped IPv4 owner", owners, e)
	}
}
