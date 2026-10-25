package main

import (
	"context"
	"encoding/json"
	"errors"
	"sync/atomic"

	"github.com/buzzmean/neexitlag/core/ipc"
)

// errStarting is returned for every IPC method while the slow initialization
// stages (control plane, browser transport restore, capture) are still running.
// The service already answers on its pipe, so SCM reports Running promptly and the
// GUI sees an explicit "starting" state instead of a missing pipe. Nothing is
// captured or rerouted in this window: networking stays DIRECT.
var errStarting = errors.New("service is starting; retry shortly")

type startupGate struct{ ready atomic.Bool }

func (g *startupGate) wrap(h ipc.Handler) ipc.Handler {
	return func(ctx context.Context, method string, raw json.RawMessage) (any, error) {
		if !g.ready.Load() && method != "activation.status" && method != "activation.set" {
			return nil, errStarting
		}
		return h(ctx, method, raw)
	}
}
