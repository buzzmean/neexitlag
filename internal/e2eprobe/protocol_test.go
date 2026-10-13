package e2eprobe

import (
	"bytes"
	"context"
	"github.com/buzzmean/neexitlag/core/measurement"
	"github.com/buzzmean/neexitlag/core/optimizer"
	"net"
	"net/netip"
	"testing"
	"time"
)

func TestRealEchoAuthenticationAndObservedPeer(t *testing.T) {
	key := bytes.Repeat([]byte{42}, 32)
	server, e := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stats := &EchoStats{}
	done := make(chan error, 1)
	go func() { done <- Serve(ctx, server, key, stats) }()
	client, e := net.DialUDP("udp4", nil, server.LocalAddr().(*net.UDPAddr))
	if e != nil {
		t.Fatal(e)
	}
	defer client.Close()
	request := Packet{Sequence: 5}
	request.Run[0] = 9
	client.Write(Encode(request, key))
	b := make([]byte, Size+1)
	client.SetReadDeadline(time.Now().Add(time.Second))
	n, e := client.Read(b)
	if e != nil {
		t.Fatal(e)
	}
	p, e := Decode(b[:n], key)
	want := netip.MustParseAddrPort(client.LocalAddr().String())
	if e != nil || !p.Reply || p.Sequence != 5 || p.Run != request.Run || p.Peer != want || n != Size {
		t.Fatal(p, e)
	}
	for _, invalid := range [][]byte{Encode(request, []byte("wrong")), append(Encode(request, key), 0), Encode(p, key)} {
		client.Write(invalid)
		client.SetReadDeadline(time.Now().Add(30 * time.Millisecond))
		if _, e = client.Read(b); e == nil {
			t.Fatal("invalid/reflected request accepted")
		}
	}
	cancel()
	if e = <-done; e != nil {
		t.Fatal(e)
	}
	if stats.Accepted.Load() != 1 || stats.Rejected.Load() != 3 {
		t.Fatal("echo counters")
	}
}
func FuzzEcho(f *testing.F) {
	key := bytes.Repeat([]byte{42}, 32)
	f.Add(Encode(Packet{Sequence: 1}, key))
	f.Fuzz(func(t *testing.T, b []byte) { Decode(b, key) })
}
func TestAuthenticatedMeasurementUsesRealEcho(t *testing.T) {
	key := bytes.Repeat([]byte{42}, 32)
	server, e := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stats := &EchoStats{}
	done := make(chan error, 1)
	go func() { done <- Serve(ctx, server, key, stats) }()
	engine := measurement.New(measurement.Settings{Samples: 3, Timeout: time.Second, Spacing: 20 * time.Millisecond, Concurrency: 1})
	result := engine.Probe(ctx, measurement.Candidate{Metadata: optimizer.RelayCandidate{ID: optimizer.Direct}, Target: netip.MustParseAddrPort(server.LocalAddr().String()), Open: ProbeOpener(measurement.OpenDirect, key)})
	// Sub-clock-resolution loopback replies can legitimately have zero RTT on Windows.
	if !result.Available || result.Samples != 3 || result.Loss != 0 || result.RTT < 0 {
		t.Fatal("authenticated measurement failed", result)
	}
	// Wait for the server to finish accounting for its last successful write.
	cancel()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if stats.Accepted.Load() != 3 || stats.Rejected.Load() != 0 {
		t.Fatal("authenticated traffic counters", stats.Accepted.Load(), stats.Rejected.Load())
	}
}
