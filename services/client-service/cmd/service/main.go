package main

import (
	"context"
	"flag"
	"github.com/buzzmean/neexitlag/core/ipc"
	windowsplatform "github.com/buzzmean/neexitlag/platform/windows"
	clientservice "github.com/buzzmean/neexitlag/services/client-service"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func run(ctx context.Context, path string) error { return runReady(ctx, path, nil) }
func runReady(ctx context.Context, path string, ready func()) error {
	manifest, e := clientservice.LoadManifest(path)
	if e != nil {
		return e
	}
	token, e := os.ReadFile(manifest.TokenFile)
	if e != nil {
		return e
	}
	controller := clientservice.New(ctx)
	defer func() {
		if err := controller.CloseForHost(ctx); err != nil {
			slog.Warn("owned browser policy cleanup incomplete", "class", "BROWSER_POLICY_CLEANUP_FAILED")
		}
	}()

	controller.BrowserOwnerSID = manifest.OwnerSID
	controller.Backend = windowsplatform.NewWinDivert(manifest.Driver, controller.Engine)
	if e = controller.SetCustomPath(manifest.StateDirectory + string(os.PathSeparator) + "profiles.json"); e != nil {
		return e
	}
	managedActivation := controller.ConfigureActivation(manifest)
	if managedActivation {
		if err := controller.CleanupPreactivationBrowserPolicies(); err != nil {
			slog.Warn("preactivation owned browser policy cleanup incomplete", "class", "BROWSER_POLICY_CLEANUP_FAILED")
		}
	}
	// The pipe opens before the slow stages so SCM does not wait on the network and
	// the GUI can tell "starting" from "not installed". Methods are rejected until
	// initialization finishes (see startupGate).
	var gate startupGate
	server, e := ipc.New(strings.TrimSpace(string(token)), gate.wrap(controller.Handle))
	if e != nil {
		return e
	}
	listener, e := clientservice.ListenIPC(manifest.OwnerSID)
	if e != nil {
		return e
	}
	defer listener.Close()
	initDone := make(chan struct{})
	go func() {
		defer close(initDone)
		// Restore listener and owned PAC before remote CP/network initialization.
		// A code-managed boot restores policy only after binding/authentication;
		// its activation worker owns that transition, never this parallel stage.
		if !managedActivation {
			if e := controller.RestoreBrowserTransport(); e != nil {
				slog.Warn("Chrome transport not restored; proxy remains fail-closed", "error", e)
			}
		}
		if manifest.ControlPlaneURL != "" && !managedActivation {
			if e := controller.ConnectRemote(ctx, manifest); e != nil {
				slog.Warn("control plane initialization unavailable; capture remains DIRECT", "error", e)
			}
		}
		controller.Start()
		if !managedActivation && controller.Snapshot().Settings.Automatic {
			if _, e := controller.Handle(ctx, "optimize", []byte(`{"enabled":true}`)); e != nil {
				slog.Warn("automatic capture unavailable; networking remains DIRECT", "error", e)
			}
		}
		gate.ready.Store(true)
		slog.Info("service initialization complete")
	}()
	defer func() { <-initDone }() // cancelled ctx unblocks the stages; Close must not race them
	if ready != nil {
		ready()
	}
	return clientservice.ServeIPC(ctx, listener, server)
}
func main() {
	recovery := flag.Bool("chrome-recovery", false, "owner-scoped Chrome recovery transport; no service privileges or credentials")
	recoveryPort := flag.Uint("recovery-port", 0, "persisted IPv4 loopback recovery port")
	path := flag.String("config", "", "absolute service manifest path")
	cleanup := flag.Bool("cleanup-browser-policy", false, "remove only journal-owned browser policies; no capture/enrollment/network")
	console := flag.Bool("console", false, "run as an interactive service host for diagnostics")
	flag.Parse()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-signals:
			cancel()
		case <-ctx.Done():
		}
	}()
	if *cleanup {
		manifest, err := clientservice.LoadManifest(*path)
		if err == nil {
			c := clientservice.New(ctx)
			c.BrowserOwnerSID = manifest.OwnerSID
			err = c.SetCustomPath(manifest.StateDirectory + string(os.PathSeparator) + "profiles.json")
			if err == nil {
				err = c.CleanupOwnedBrowserPolicies()
			}
			c.Close()
		}
		if err != nil {
			slog.Error("browser policy cleanup failed", "class", "BROWSER_POLICY_CLEANUP_FAILED")
			os.Exit(1)
		}
		slog.Info("owned browser policies removed or absent; foreign settings preserved")
		return
	}
	if *recovery {
		if *recoveryPort > 65535 {
			os.Exit(1)
		}
		if err := runBrowserRecovery(ctx, uint16(*recoveryPort)); err != nil {
			slog.Error("Chrome recovery stopped", "class", "RECOVERY_START_FAILED")
			os.Exit(1)
		}
		return
	}
	if e := platformRun(ctx, *path, *console); e != nil {
		slog.Error("client service stopped", "error", e)
		os.Exit(1)
	}
}
