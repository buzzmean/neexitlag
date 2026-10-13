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
	defer controller.Close()
	controller.Backend = windowsplatform.NewWinDivert(manifest.Driver, controller.Engine)
	if e = controller.SetCustomPath(manifest.StateDirectory + string(os.PathSeparator) + "profiles.json"); e != nil {
		return e
	}
	if manifest.ControlPlaneURL != "" {
		if e = controller.ConnectRemote(ctx, manifest); e != nil {
			slog.Warn("control plane initialization unavailable; capture remains DIRECT", "error", e)
		}
	}
	controller.Start()
	if controller.Snapshot().Settings.Automatic {
		if _, e := controller.Handle(ctx, "optimize", []byte(`{"enabled":true}`)); e != nil {
			slog.Warn("automatic capture unavailable; networking remains DIRECT", "error", e)
		}
	}
	server, e := ipc.New(strings.TrimSpace(string(token)), controller.Handle)
	if e != nil {
		return e
	}
	listener, e := clientservice.ListenIPC(manifest.OwnerSID)
	if e != nil {
		return e
	}
	defer listener.Close()
	if ready != nil {
		ready()
	}
	return clientservice.ServeIPC(ctx, listener, server)
}
func main() {
	path := flag.String("config", "", "absolute service manifest path")
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
	if e := platformRun(ctx, *path, *console); e != nil {
		slog.Error("client service stopped", "error", e)
		os.Exit(1)
	}
}
