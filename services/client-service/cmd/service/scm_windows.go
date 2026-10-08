//go:build windows

package main

import (
	"context"
	"golang.org/x/sys/windows/svc"
	"time"
)

type handler struct {
	path   string
	parent context.Context
}

func (h handler) Execute(args []string, requests <-chan svc.ChangeRequest, status chan<- svc.Status) (bool, uint32) {
	ctx, cancel := context.WithCancel(h.parent)
	defer cancel()
	status <- svc.Status{State: svc.StartPending}
	done := make(chan error, 1)
	go func() { done <- run(ctx, h.path) }()
	status <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}
	for {
		select {
		case e := <-done:
			if e != nil {
				return false, 1
			}
			return false, 0
		case r := <-requests:
			switch r.Cmd {
			case svc.Interrogate:
				status <- r.CurrentStatus
			case svc.Stop, svc.Shutdown:
				status <- svc.Status{State: svc.StopPending}
				cancel()
				select {
				case <-done:
				case <-time.After(10 * time.Second):
				}
				return false, 0
			}
		}
	}
}
func platformRun(ctx context.Context, path string, console bool) error {
	if console {
		return run(ctx, path)
	}
	service, e := svc.IsWindowsService()
	if e != nil {
		return e
	}
	if !service {
		return run(ctx, path)
	}
	return svc.Run("neexitlag", handler{path, ctx})
}
