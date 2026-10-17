package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"github.com/buzzmean/neexitlag/core/browserproxy"
	wp "github.com/buzzmean/neexitlag/platform/windows"
	"io"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
)

// This owner-user child has no manifest, token, device key, driver or admin RPC.
// Its only configuration channel is its supervising GUI's inherited stdin pipe.
// EOF closes the listener. It cannot serve other users/processes or guarded
// domains, even when configured maliciously by its own unprivileged owner.
func runBrowserRecovery(parent context.Context, port uint16) error {
	return runBrowserRecoveryWithDial(parent, port, nil)
}
func runBrowserRecoveryWithDial(parent context.Context, port uint16, direct func(context.Context, browserproxy.Request) (net.Conn, error)) error {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	self, err := wp.NewDiscovery().Lookup(uint32(os.Getpid()))
	if err != nil {
		return err
	}
	sid, err := wp.ProcessUserSID(self)
	if err != nil {
		return err
	}
	address := netip.AddrPortFrom(netip.MustParseAddr("127.0.0.1"), port).String()
	server, err := browserproxy.ListenRecovery(ctx, address, func(ctx context.Context, source, local netip.AddrPort, cfg browserproxy.RecoverySettings) (browserproxy.Owner, error) {
		id, err := wp.LookupTCPPeer(source, local)
		if err != nil {
			return browserproxy.Owner{}, browserproxy.ErrOwnerLookup
		}
		user, err := wp.ProcessUserSID(id)
		if err != nil || user != sid {
			return browserproxy.Owner{}, browserproxy.ErrOwnerSID
		}
		if !strings.EqualFold(id.Path, cfg.Executable) {
			return browserproxy.Owner{}, browserproxy.ErrOwnerProfile
		}
		return browserproxy.Owner{PID: id.PID, Birth: id.CreationTime, Exe: id.Path, Profile: cfg.Profile}, nil
	}, direct)
	if err != nil {
		return err
	}
	defer server.Close()
	server.AuthorizeMetadata(func(ctx context.Context, source, local netip.AddrPort) error {
		id, err := wp.LookupTCPPeer(source, local)
		if err != nil {
			return err
		}
		user, err := wp.ProcessUserSID(id)
		if err != nil || user != sid {
			return browserproxy.ErrOwnerSID
		}
		return nil
	})
	encoder := json.NewEncoder(os.Stdout)
	if err = encoder.Encode(map[string]any{"address": server.Address(), "metadata_token": server.MetadataToken()}); err != nil {
		return err
	}
	configured := make(chan error, 1)
	go func() {
		defer cancel()
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Buffer(make([]byte, 4096), 262144)
		first := true
		for scanner.Scan() {
			var cfg browserproxy.RecoverySettings
			err := json.Unmarshal(scanner.Bytes(), &cfg)
			if err == nil && (!filepath.IsAbs(cfg.GuardFile) || filepath.Base(cfg.GuardFile) != "browser-recovery-guard.json" || !filepath.IsAbs(cfg.Executable)) {
				err = errors.New("absolute Chrome path required")
			}
			if err == nil {
				err = server.Configure(cfg)
			}
			if first {
				configured <- err
				first = false
			}
			if err != nil {
				return
			}
			if encoder.Encode(map[string]bool{"ready": true}) != nil {
				return
			}
		}
		if first {
			configured <- io.ErrUnexpectedEOF
		}
	}()
	select {
	case err = <-configured:
		if err != nil {
			return err
		}
	case <-ctx.Done():
		return ctx.Err()
	}
	<-ctx.Done()
	return nil
}
