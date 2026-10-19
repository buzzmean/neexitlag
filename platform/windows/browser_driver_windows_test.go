//go:build windows

package windowsplatform

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/buzzmean/neexitlag/core/routing"
)

// Isolates the actual signed-driver TCP handoff from Chrome policy and slow
// network inventory. A mocked packetWriter cannot satisfy this test.
func TestNativeBrowserDriverHandshake(t *testing.T) {
	dir := os.Getenv("NEEXITLAG_TEST_DRIVER_DIR")
	if dir == "" {
		t.Skip("signed driver directory not configured")
	}
	raw, err := os.ReadFile(filepath.Join(dir, "pins.json"))
	if err != nil {
		t.Fatal(err)
	}
	var files DriverFiles
	if err = json.Unmarshal(raw, &files); err != nil {
		t.Fatal(err)
	}
	files.Directory = dir
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	b := NewBrowserLoopbackFixture(files, routing.New(16))
	if err = b.Start(ctx, func(ProcessIdentity, Tuple) Decision { return Decision{} }); err != nil {
		t.Fatal(err)
	}
	defer b.Stop()
	b.AdmitNativeFixtureListener(t)
	id, err := NewDiscovery().Lookup(uint32(os.Getpid()))
	if err != nil {
		t.Fatal(err)
	}
	var mu sync.Mutex
	trace := []string{}
	b.TraceNativeFixture(func(phase string, tuple Tuple, flags, addr byte) {
		mu.Lock()
		defer mu.Unlock()
		if len(trace) < 100 {
			trace = append(trace, phase+" "+tuple.Source.String()+" -> "+tuple.Destination.String()+" flags="+fmt.Sprint(flags)+" address="+fmt.Sprint(addr))
		}
	})
	d := Decision{NewConnectionsOnly: true, BlockOnFailure: true, OriginDomain: "controlled.example"}
	d.OpenDeferred = func(_ context.Context, d Decision, _ string) (net.Conn, Decision, error) {
		up, peer := net.Pipe()
		go func() { defer peer.Close(); _, _ = io.Copy(peer, peer) }()
		return up, d, nil
	}
	conn, err := b.OpenBrowserNativeTCP(ctx, id, netip.MustParseAddrPort("127.0.0.1:12345"), netip.MustParseAddrPort("127.0.0.1:443"), d)
	dump := func() {
		mu.Lock()
		defer mu.Unlock()
		for _, line := range trace {
			t.Log(line)
		}
	}
	if err != nil {
		dump()
		t.Fatal("real native TCP handshake", err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(2 * time.Second))
	const payload = "VELIQ-native-TCP-nonce"
	if _, err = io.WriteString(conn, payload); err != nil {
		dump()
		t.Fatal(err)
	}
	got := make([]byte, len(payload))
	if _, err = io.ReadFull(conn, got); err != nil || string(got) != payload {
		dump()
		t.Fatal("actual native bidirectional payload", string(got), err)
	}
	dump()
}
