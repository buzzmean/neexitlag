//go:build windows

package windowsplatform

import (
	"context"
	"encoding/json"
	"github.com/buzzmean/neexitlag/core/routing"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
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
	if e = backend.Stop(); e != nil {
		t.Fatal(e)
	}
	if backend.Status().Ready {
		t.Fatal("still ready")
	}
}
