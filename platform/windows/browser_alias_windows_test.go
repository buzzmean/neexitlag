//go:build windows

package windowsplatform

import (
	"net/netip"
	"testing"
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
