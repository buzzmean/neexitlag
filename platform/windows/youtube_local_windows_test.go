//go:build windows

package windowsplatform

import (
	"bytes"
	"encoding/binary"
	"net/netip"
	"testing"
	"unsafe"

	"github.com/buzzmean/neexitlag/core/localstrategy"
)

// Pinned native checksum helper only; no WinDivertOpen, driver, traffic or VPS.
func TestYouTubeRepeatedFakeChecksumsWithPinnedDLL(t *testing.T) {
	dll := loadPinnedFilterDLL(t)
	checksum, err := dll.FindProc("WinDivertHelperCalcChecksums")
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range localstrategy.BrowserYouTubeCandidates()[:2] {
		engine := localstrategy.New()
		if err = engine.Reload(localstrategy.Config{Version: 1, Enabled: true, User: []localstrategy.Profile{{ID: "browser-yt", Name: "fixture", Version: 1, Application: "chrome.exe", Protocol: "tcp", Ports: []uint16{443}, IPs: []string{"203.0.113.9"}, Strategy: "chain", Actions: candidate.Actions, Cutoff: 1, Enabled: true}}}); err != nil {
			t.Fatal(err)
		}
		metadata := localstrategy.Metadata{Application: "chrome.exe", Protocol: "tcp", Destination: netip.MustParseAddrPort("203.0.113.9:443")}
		state := &localstrategy.State{}
		syn := localTCPPacket(2, nil)
		binary.BigEndian.PutUint32(syn[24:], 99)
		engine.Process(syn, state, metadata)
		state.ConfirmHandshake(100, 199)
		raw := localTCPPacket(24, localTestHello(t))
		binary.BigEndian.PutUint32(raw[24:], 100)
		binary.BigEndian.PutUint32(raw[28:], 200)
		result := engine.Process(raw, state, metadata)
		if result.FakePackets != 2 || len(result.Packets) != 5 {
			t.Fatal("bounded repeated plan absent", candidate.ID, result)
		}
		for _, packet := range result.Packets {
			before := append([]byte(nil), packet...)
			var address [80]byte
			checksum.Call(uintptr(unsafe.Pointer(&packet[0])), uintptr(len(packet)), uintptr(unsafe.Pointer(&address[0])), 0)
			if !bytes.Equal(before, packet) {
				t.Fatal("independent native checksum disagreement", candidate.ID)
			}
		}
	}
}
