package routing

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"
)

// AdaptiveUDP uses real DIRECT replies. Only the initial no-reply deadline or a
// real socket error can trigger one transition. Previously sent media is never
// replayed; an atomic UDP Send error can retry that unsent datagram on relay.
type AdaptiveUDP struct {
	mu                            sync.Mutex
	readMu                        sync.Mutex
	current                       Datagram
	e                             *Engine
	flow                          Flow
	relay                         string
	started                       time.Time
	deadline                      time.Duration
	replied, transitioned, closed bool
	reason                        string
}
type udpHealthKey struct{ Application, Network, Destination string }
type udpHealth struct {
	relay   bool
	expires time.Time
}

func (e *Engine) udpCached(f Flow, now time.Time) bool {
	if f.Network == "" {
		return false
	}
	e.healthMu.Lock()
	defer e.healthMu.Unlock()
	k := udpHealthKey{f.ApplicationID, f.Network, f.Destination.String()}
	v := e.health[k]
	return v.relay && now.Before(v.expires)
}
func (e *Engine) udpObserved(f Flow, relay bool) {
	if f.Network == "" {
		return
	}
	e.healthMu.Lock()
	defer e.healthMu.Unlock()
	now := time.Now()
	for k, v := range e.health {
		if !now.Before(v.expires) {
			delete(e.health, k)
		}
	}
	k := udpHealthKey{f.ApplicationID, f.Network, f.Destination.String()}
	if len(e.health) < 1024 || !e.health[k].expires.IsZero() {
		e.health[k] = udpHealth{relay, now.Add(2 * time.Minute)}
	}
}
func (e *Engine) OpenAdaptiveUDP(ctx context.Context, f Flow, relay string, deadline time.Duration) (*AdaptiveUDP, error) {
	if relay == "" || deadline <= 0 || deadline > 10*time.Second {
		return nil, errors.New("invalid adaptive UDP bound")
	}
	a := &AdaptiveUDP{e: e, flow: f, relay: relay, started: time.Now(), deadline: deadline}
	cached := e.udpCached(f, time.Now())
	f.PinnedPolicy = &Policy{}
	if cached {
		f.PinnedPolicy = &Policy{Route: Route{RelayID: relay}, BlockOnFailure: true}
		a.transitioned = true
		a.reason = "LEARNED_RECENT_DIRECT_FAILURE"
	}
	d, err := e.OpenUDP(ctx, f)
	if err != nil && !cached {
		f.PinnedPolicy = &Policy{Route: Route{RelayID: relay}, BlockOnFailure: true}
		d, err = e.OpenUDP(ctx, f)
		a.transitioned = true
		a.reason = "DIRECT_OPEN_FAILED"
		e.udpObserved(f, true)
	}
	if err != nil {
		return nil, err
	}
	a.current = d
	return a, nil
}
func (d *AdaptiveUDP) Route() Route {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.transitioned {
		return Route{RelayID: d.relay}
	}
	return Route{}
}
func (d *AdaptiveUDP) Reason() string { d.mu.Lock(); defer d.mu.Unlock(); return d.reason }
func (d *AdaptiveUDP) toRelayLocked(ctx context.Context, reason string) error {
	d.current.Close()
	d.reason = reason
	d.e.udpObserved(d.flow, true)
	f := d.flow
	f.PinnedPolicy = &Policy{Route: Route{RelayID: d.relay}, BlockOnFailure: true}
	openCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
	defer cancel()
	next, err := d.e.OpenUDP(openCtx, f)
	if err != nil {
		d.closed = true
		return err
	}
	d.current = next
	d.transitioned = true
	return nil
}
func (d *AdaptiveUDP) Send(b []byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return errors.New("UDP flow closed")
	}
	err := d.current.Send(b)
	if err != nil && !d.transitioned {
		if err = d.toRelayLocked(context.Background(), "DIRECT_SEND_FAILED"); err == nil {
			err = d.current.Send(b)
		}
	}
	return err
}
func (d *AdaptiveUDP) Receive(ctx context.Context) ([]byte, error) {
	d.readMu.Lock()
	defer d.readMu.Unlock()
	d.mu.Lock()
	if d.closed {
		d.mu.Unlock()
		return nil, errors.New("UDP flow closed")
	}
	current := d.current
	probe := !d.replied && !d.transitioned
	until := d.started.Add(d.deadline)
	d.mu.Unlock()
	readCtx := ctx
	cancel := func() {}
	if probe {
		readCtx, cancel = context.WithDeadline(ctx, until)
	}
	b, err := current.Receive(readCtx)
	cancel()
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return nil, errors.New("UDP flow closed")
	}
	if current != d.current {
		return nil, context.DeadlineExceeded
	}
	if err == nil {
		d.replied = true
		if !d.transitioned {
			d.e.udpObserved(d.flow, false)
		}
		return b, nil
	}
	if ctx.Err() == context.Canceled {
		return nil, err
	}
	var networkError net.Error
	timeout := errors.Is(err, context.DeadlineExceeded) || errors.As(err, &networkError) && networkError.Timeout()
	if d.transitioned {
		if timeout {
			return nil, context.DeadlineExceeded
		}
		return nil, err
	}
	if timeout && (!probe || time.Now().Before(until)) {
		return nil, context.DeadlineExceeded
	}
	reason := "DIRECT_RECEIVE_FAILED"
	if timeout {
		reason = "DIRECT_NO_REPLY_FLOW_RELAY"
	}
	if err = d.toRelayLocked(ctx, reason); err != nil {
		return nil, err
	}
	return nil, context.DeadlineExceeded
}
func (d *AdaptiveUDP) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.closed = true
	return d.current.Close()
}
