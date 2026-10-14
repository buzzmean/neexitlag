package windowsplatform

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/buzzmean/neexitlag/core/localstrategy"
	"github.com/buzzmean/neexitlag/core/routing"
	"net/netip"
	"testing"
)

func localTestPacket() []byte {
	p := make([]byte, 80)
	p[0] = 0x45
	p[8] = 64
	p[9] = 6
	p[32] = 0x50
	p[33] = 0x18
	binary.BigEndian.PutUint16(p[2:], 80)
	copy(p[12:16], []byte{192, 0, 2, 1})
	copy(p[16:20], []byte{203, 0, 113, 9})
	binary.BigEndian.PutUint16(p[20:], 1234)
	binary.BigEndian.PutUint16(p[22:], 443)
	copy(p[40:], bytes.Repeat([]byte{7}, 40))
	return p
}
func TestLocalSimultaneousClassificationAndSend(t *testing.T) {
	e := localstrategy.New()
	e.Reload(localstrategy.Config{Version: 1, Enabled: true, User: []localstrategy.Profile{{ID: "split", Version: 1, Application: "chosen.exe", Protocol: "tcp", Ports: []uint16{443}, IPs: []string{"203.0.113.9"}, Strategy: "tcp-split-1", Cutoff: 1, Enabled: true}}})
	ledger := NewEvidenceLedger()
	tuple := Tuple{"tcp", netip.MustParseAddrPort("192.0.2.1:1234"), netip.MustParseAddrPort("203.0.113.9:443")}
	classify := func(id ProcessIdentity, _ Tuple) Decision {
		if id.PID == 1 {
			return Decision{Route: routing.Route{Kind: routing.DirectPlusRoute}, Local: e}
		}
		if id.PID == 3 {
			return Decision{Route: routing.Route{RelayID: "VPS-01"}}
		}
		return Decision{}
	}
	for _, id := range []ProcessIdentity{{PID: 1, CreationTime: 10, Path: "chosen.exe"}, {PID: 2, CreationTime: 20, Path: "control.exe"}, {PID: 3, CreationTime: 30, Path: "relay.exe"}, {PID: 4, CreationTime: 40, Path: "service.exe"}} {
		d, self := ClassifyCapture(id, tuple, 4, classify)
		if id.PID == 3 {
			if d.Route.RelayID != "VPS-01" {
				t.Fatal(d)
			}
			continue
		}
		sent := 0
		state := &localstrategy.State{}
		if err := SendLocal(d, state, id, tuple, localTestPacket(), func([]byte) error { sent++; return nil }, ledger); err != nil {
			t.Fatal(err)
		}
		if id.PID == 1 && sent != 2 || id.PID != 1 && sent != 1 || id.PID == 4 && !self {
			t.Fatalf("PID association/send %d %d", id.PID, sent)
		}
	}
	if ledger.Snapshot(1).Flows[0].Route != "DIRECT_PLUS" || ledger.Snapshot(2).Flows[0].Route != "DIRECT" {
		t.Fatal("evidence")
	}
}
func TestLocalReinjectionFailureAndUnknownIdentityFailOpen(t *testing.T) {
	e := localstrategy.New()
	e.Reload(localstrategy.Config{Version: 1, Enabled: true, User: []localstrategy.Profile{{ID: "split", Version: 1, Application: "chosen.exe", Protocol: "tcp", Ports: []uint16{443}, IPs: []string{"203.0.113.9"}, Strategy: "tcp-split-1", Cutoff: 1, Enabled: true}}})
	tuple := Tuple{"tcp", netip.MustParseAddrPort("192.0.2.1:1234"), netip.MustParseAddrPort("203.0.113.9:443")}
	id := ProcessIdentity{PID: 1, CreationTime: 10, Path: "chosen.exe"}
	raw := localTestPacket()
	d := Decision{Route: routing.Route{Kind: routing.DirectPlusRoute}, Local: e}
	n := 0
	ledger := NewEvidenceLedger()
	st := &localstrategy.State{}
	err := SendLocal(d, st, id, tuple, raw, func(p []byte) error {
		n++
		if n == 2 {
			return errors.New("send failure")
		}
		if n == 3 && !bytes.Equal(p, raw) {
			t.Fatal("not original failopen")
		}
		return nil
	}, ledger)
	if err != nil || n != 3 || st.Applied() {
		t.Fatal(err, n, st)
	}
	if ledger.Snapshot(1).Flows[0].Route != "DIRECT" {
		t.Fatal("false local success")
	}
	n = 0
	SendLocal(d, &localstrategy.State{}, ProcessIdentity{PID: 1}, tuple, raw, func([]byte) error { n++; return nil }, ledger)
	if n != 1 {
		t.Fatal("unknown identity modified")
	}
}

func TestFragmentAddressHasCompletePrecomputedUDPChecksum(t *testing.T) {
	var address [80]byte
	address[10] = 2
	p := make([]byte, 44)
	p[0] = 0x45
	p[9] = 17
	p[6] = 0x20
	got := LocalPacketAddress(p, address)
	if got[10] != 0xa2 {
		t.Fatal(got[10])
	}
	p[6] = 0x40
	if got = LocalPacketAddress(p, address); got != address {
		t.Fatal("unfragmented address changed")
	}
}
