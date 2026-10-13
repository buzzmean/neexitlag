package main

import (
	"bytes"
	"context"
	"github.com/buzzmean/neexitlag/internal/e2eprobe"
	"net"
	"net/netip"
	"testing"
	"time"
)

func TestNormalSocketClientReportsRealTraffic(t *testing.T) {
	key := bytes.Repeat([]byte{42}, 32)
	server, e := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go e2eprobe.Serve(ctx, server, key, &e2eprobe.EchoStats{})
	r, e := run(netip.MustParseAddrPort(server.LocalAddr().String()), key, 5, 20*time.Millisecond, 100*time.Millisecond, true)
	if e != nil || r.Sent != 5 || r.Received != 5 || r.Loss != 0 || r.RTTMS <= 0 || r.Exe == "" || r.PID == 0 || len(r.ObservedSources) != 1 || r.ObservedSources[0] != "127.0.0.1" {
		t.Fatal(r, e)
	}
}
