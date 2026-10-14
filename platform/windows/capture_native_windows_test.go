//go:build windows

package windowsplatform

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"github.com/buzzmean/neexitlag/core/localstrategy"
	"github.com/buzzmean/neexitlag/core/routing"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"testing"
	"time"
	"unsafe"
)

// Opt-in signed-driver lifecycle regression. This is NOT external TCP acceptance.
func TestNativeCaptureHandlesStartStop(t *testing.T) {
	dir := os.Getenv("NEEXITLAG_TEST_DRIVER_DIR")
	if dir == "" {
		t.Skip("signed driver directory not configured")
	}
	raw, e := os.ReadFile(filepath.Join(dir, "pins.json"))
	if e != nil {
		t.Fatal(e)
	}
	var pins DriverFiles
	if e = json.Unmarshal(raw, &pins); e != nil {
		t.Fatal(e)
	}
	pins.Directory = dir
	backend := NewWinDivert(pins, routing.New(16))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if e = backend.Start(ctx, func(ProcessIdentity, Tuple) Decision { return Decision{} }); e != nil {
		t.Fatal(e)
	}
	defer backend.Stop()
	for _, network := range []string{"tcp4", "tcp6"} {
		address := "127.0.0.1:0"
		if network == "tcp6" {
			address = "[::1]:0"
		}
		listener, err := net.Listen(network, address)
		if err != nil {
			if network == "tcp6" {
				t.Log("IPv6 loopback unavailable:", err)
				continue
			}
			t.Fatal(err)
		}
		conn, err := net.DialTimeout(network, listener.Addr().String(), 2*time.Second)
		listener.Close()
		if err != nil {
			t.Fatalf("ordinary %s connect blocked by capture: %v", network, err)
		}
		conn.Close()
	}
	if !backend.Status().Ready || !backend.Status().Capabilities.TCP {
		t.Fatal(backend.Status())
	}
	testNativeLocalChecksumHelper(t, backend.proc())
	if e = backend.Stop(); e != nil {
		t.Fatal(e)
	}
	if backend.Status().Ready {
		t.Fatal("still ready")
	}
}

// Real native DLL helper + real current PID/exe identity, no external-send claim.
func testNativeLocalChecksumHelper(t *testing.T, d *divertDLL) {
	id, err := NewDiscovery().Lookup(uint32(os.Getpid()))
	if err != nil || id.CreationTime == 0 || id.Path == "" {
		t.Fatal("native identity", err)
	}
	for _, kind := range []string{"tcp", "udp"} {
		raw := localTestPacket()
		strategy := "tcp-split-1"
		if kind == "udp" {
			raw = make([]byte, 128)
			raw[0] = 0x45
			raw[8] = 64
			raw[9] = 17
			raw[4] = 7
			raw[6] = 0x40
			binary.BigEndian.PutUint16(raw[2:], 128)
			binary.BigEndian.PutUint16(raw[24:], 108)
			binary.BigEndian.PutUint16(raw[20:], 1234)
			binary.BigEndian.PutUint16(raw[22:], 443)
			copy(raw[12:16], []byte{192, 0, 2, 1})
			copy(raw[16:20], []byte{203, 0, 113, 9})
			strategy = "udp-fragment-24"
		}
		e := localstrategy.New()
		if err := e.Reload(localstrategy.Config{Version: 1, Enabled: true, User: []localstrategy.Profile{{ID: "native", Version: 1, Application: id.Path, Protocol: kind, Ports: []uint16{443}, IPs: []string{"203.0.113.9"}, Strategy: strategy, Cutoff: 1, Enabled: true}}}); err != nil {
			t.Fatal(err)
		}
		r := e.Process(raw, &localstrategy.State{}, localstrategy.Metadata{Application: id.Path, Protocol: kind, Destination: netip.MustParseAddrPort("203.0.113.9:443")})
		if !r.Modified {
			t.Fatal("native transformation absent")
		}
		for _, q := range r.Packets {
			before := append([]byte(nil), q...)
			a := LocalPacketAddress(q, [80]byte{})
			d.checksum.Call(uintptr(unsafe.Pointer(&q[0])), uintptr(len(q)), uintptr(unsafe.Pointer(&a[0])), 0)
			if !bytes.Equal(before, q) {
				t.Fatalf("%s native helper disagrees with independent packet checksums", kind)
			}
			if kind == "udp" && a[10]&0xa0 != 0xa0 {
				t.Fatal("fragment checksum validity flags")
			}
		}
	}
}
