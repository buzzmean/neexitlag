//go:build windows

package windowsplatform

import (
	"net/netip"
	"testing"
	"time"
)

func TestBrowserAliasCleanupBeforeFirstCapturedSYN(t *testing.T) {
	ledger := NewEvidenceLedger()
	id := ProcessIdentity{PID: 42, CreationTime: 17, Path: `C:\Apps\chrome.exe`}
	redirect := &tcpRedirect{b: &WinDivertBackend{evidence: ledger}, state: newTCPState(16), delegated: map[Tuple]delegatedTCP{}}
	dest := netip.MustParseAddrPort("203.0.113.1:443")
	for i := uint16(0); i < 3; i++ {
		original := Tuple{Protocol: "tcp", Source: netip.AddrPortFrom(netip.MustParseAddr("192.0.2.1"), 1200+i), Destination: dest}
		upstream := original
		upstream.Source = netip.AddrPortFrom(original.Source.Addr(), 2200+i)
		ledger.BeginFlow(id, original)
		ledger.RecordBytes(id, original, "captured", "", 40)
		if !ledger.BindUpstream(id, upstream, original) {
			t.Fatal("bind")
		}
		if i < 2 {
			redirect.delegated[upstream] = delegatedTCP{}
		}
	}
	// Exactly the cleanup race: reservation exists, SYN has not arrived in state.
	redirect.state.mu.Lock()
	redirect.pruneEvidenceAliases()
	redirect.state.mu.Unlock()
	if len(ledger.aliases) != 2 {
		t.Fatal("lost live pre-SYN reservation or retained dead alias", len(ledger.aliases))
	}
	for i := uint16(0); i < 2; i++ {
		original := Tuple{Protocol: "tcp", Source: netip.AddrPortFrom(netip.MustParseAddr("192.0.2.1"), 1200+i), Destination: dest}
		upstream := original
		upstream.Source = netip.AddrPortFrom(original.Source.Addr(), 2200+i)
		ledger.BeginFlow(id, upstream)
		ledger.RecordLocal(id, upstream, "disorder", "www.youtube.com", 100+uint64(i))
		proof, err := ledger.SnapshotQuery(EvidenceQuery{Source: original.Source.String()}, nil)
		if err != nil || len(proof.Flows) != 1 || proof.Flows[0].Counters.DirectPlusBytes != 100+uint64(i) || proof.Flows[0].Counters.RelayedBytes != 0 {
			t.Fatal("transform not attributed to exact CONNECT", err, proof)
		}
		redirect.releaseDelegated(upstream, id)
	}
	if len(ledger.aliases) != 0 {
		t.Fatal("released aliases retained")
	}
}

func TestNativeDelegatedLoopbackReplyUsesExactReflection(t *testing.T) {
	b, _ := captureFixture(t)
	b.testBrowserLoopback = true
	redirect := &tcpRedirect{b: b, port: 19000, state: newTCPState(16), delegated: map[Tuple]delegatedTCP{}}
	tuple := Tuple{Protocol: "tcp", Source: netip.MustParseAddrPort("127.0.0.1:12000"), Destination: netip.MustParseAddrPort("127.0.0.1:443")}
	entry := redirect.state.add(tuple, ProcessIdentity{PID: 42}, Decision{Deferred: true}, "opening", time.Now())
	entry.mode = "relay"
	reply := Tuple{Protocol: "tcp", Source: netip.AddrPortFrom(tuple.Source.Addr(), redirect.port), Destination: netip.AddrPortFrom(tuple.Destination.Addr(), entry.virtual)}
	var written Tuple
	b.packetWriter = func(p []byte, _ [80]byte) error { written, _, _, _ = ParseTCP(p); return nil }
	packet := tcpPacket(t, reply)
	packet[33] = 0x12
	var addr [80]byte
	addr[10] = 6
	redirect.processPacket(packet, addr, func(ProcessIdentity, Tuple) Decision { t.Fatal("classified exact listener reply"); return Decision{} })
	if written.Source != tuple.Destination || written.Destination != tuple.Source {
		t.Fatal("listener reply bypassed native reflection", written)
	}
	// A different local port/address must still be passed unchanged.
	other := reply
	other.Destination = netip.MustParseAddrPort("127.0.0.1:15000")
	redirect.processPacket(tcpPacket(t, other), addr, func(ProcessIdentity, Tuple) Decision { t.Fatal("captured unrelated loopback"); return Decision{} })
	if written != other {
		t.Fatal("unrelated loopback intercepted", written)
	}
}

func TestNativeDelegationTestDestinationNeverChangesProduction(t *testing.T) {
	dest := netip.MustParseAddrPort("127.0.0.1:443")
	b := &WinDivertBackend{}
	if b.nativeDelegationDestination(dest) != dest {
		t.Fatal("production destination changed")
	}
	b.testBrowserLoopback = true
	if b.nativeDelegationDestination(dest).Addr() != netip.MustParseAddr("192.0.2.254") {
		t.Fatal("fixture still attempts loopback NAT")
	}
	public := netip.MustParseAddrPort("203.0.113.9:443")
	if b.nativeDelegationDestination(public) != public {
		t.Fatal("fixture changed a real origin")
	}
}
