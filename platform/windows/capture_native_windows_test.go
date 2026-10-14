//go:build windows

package windowsplatform

import (
	"context"
	"encoding/json"
	"github.com/buzzmean/neexitlag/core/routing"
	"os"
	"path/filepath"
	"testing"
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
