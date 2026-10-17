// Package measurement performs bounded application-level UDP echo probing.
package measurement

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"github.com/buzzmean/neexitlag/core/optimizer"
	"github.com/buzzmean/neexitlag/core/routing"
	"github.com/buzzmean/neexitlag/core/telemetry"
	"net"
	"net/netip"
	"sync"
	"sync/atomic"
	"time"
)

type Opener func(context.Context, netip.AddrPort) (routing.Datagram, error)
type Candidate struct {
	Metadata optimizer.RelayCandidate
	Target   netip.AddrPort
	Open     Opener
	Load     float64
}
type Settings struct {
	ActiveInterval, IdleInterval, Timeout, Spacing time.Duration
	Samples, Concurrency                           int
}

func Defaults() Settings {
	return Settings{5 * time.Second, 30 * time.Second, 700 * time.Millisecond, 100 * time.Millisecond, 5, 3}
}

type Engine struct {
	Settings   Settings
	Active     atomic.Bool
	mu         sync.Mutex
	candidates []Candidate
	wg         sync.WaitGroup
	slots      chan struct{}
	OnFailure  func(string, string, time.Time)
}

func New(s Settings) *Engine {
	if s.ActiveInterval < time.Second {
		s.ActiveInterval = time.Second
	}
	if s.IdleInterval < s.ActiveInterval {
		s.IdleInterval = s.ActiveInterval
	}
	if s.Timeout <= 0 {
		s.Timeout = time.Second
	}
	if s.Spacing < 20*time.Millisecond {
		s.Spacing = 20 * time.Millisecond
	}
	if s.Samples < 1 {
		s.Samples = 1
	}
	if s.Samples > 20 {
		s.Samples = 20
	}
	if s.Concurrency < 1 {
		s.Concurrency = 1
	}
	if s.Concurrency > 8 {
		s.Concurrency = 8
	}
	return &Engine{Settings: s, slots: make(chan struct{}, s.Concurrency)}
}
func (e *Engine) SetCandidates(c []Candidate) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if len(c) > 32 {
		c = c[:32]
	}
	e.candidates = append([]Candidate(nil), c...)
}
func (e *Engine) Round(ctx context.Context) []optimizer.RouteMeasurement { return e.round(ctx, nil) }
func (e *Engine) round(ctx context.Context, observe func(optimizer.RouteMeasurement)) []optimizer.RouteMeasurement {
	e.mu.Lock()
	all := append([]Candidate(nil), e.candidates...)
	e.mu.Unlock()
	out := make([]optimizer.RouteMeasurement, len(all))
	sem := make(chan struct{}, e.Settings.Concurrency)
	var wg sync.WaitGroup
launch:
	for i, c := range all {
		select {
		case <-ctx.Done():
			break launch
		case sem <- struct{}{}:
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			out[i] = e.Probe(ctx, c)
			if observe != nil && out[i].RouteID != "" {
				observe(out[i])
			}
		}()
	}
	wg.Wait()
	return out
}
func (e *Engine) Run(ctx context.Context, observe func(optimizer.RouteMeasurement)) {
	for {
		e.round(ctx, observe)
		interval := e.Settings.IdleInterval
		if e.Active.Load() {
			interval = e.Settings.ActiveInterval
		}
		t := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			t.Stop()
			return
		case <-t.C:
		}
	}
}
func (e *Engine) Probe(ctx context.Context, c Candidate) optimizer.RouteMeasurement {
	select {
	case e.slots <- struct{}{}:
		defer func() { <-e.slots }()
	case <-ctx.Done():
		return optimizer.RouteMeasurement{}
	}
	m := optimizer.RouteMeasurement{RouteID: c.Metadata.ID, Load: c.Load, ObservedAt: time.Now(), Stability: 1}
	if c.Open == nil || !c.Target.IsValid() {
		m.FailureRate = 1
		return m
	}
	openCtx, cancel := context.WithTimeout(ctx, e.Settings.Timeout)
	flow, err := c.Open(openCtx, c.Target)
	cancel()
	if err != nil {
		if ctx.Err() != nil || errors.Is(err, telemetry.ErrCapacity) || errors.Is(err, telemetry.ErrIsolatedSession) {
			// Local pool admission is not a remote quality sample.
			return optimizer.RouteMeasurement{}
		}
		m.FailureRate = 1
		m.TransportError = "open_failed"
		if e.OnFailure != nil && !telemetry.IsLocal(err) && ctx.Err() == nil {
			e.OnFailure(c.Metadata.ID, "transport_open_failed", time.Now())
		}
		return m
	}
	defer flow.Close()
	received, failed := 0, 0
	var total, jitter, previous time.Duration
	for i := 0; i < e.Settings.Samples; i++ {
		if ctx.Err() != nil {
			break
		}
		m.Samples++
		payload := make([]byte, 32)
		if _, err = rand.Read(payload); err != nil {
			failed++
			m.Probes = append(m.Probes, optimizer.ProbeSample{At: time.Now(), Outcome: optimizer.LocalFailure})
			continue
		}
		start := time.Now()
		sample := optimizer.ProbeSample{At: start}
		probe, stop := context.WithTimeout(ctx, e.Settings.Timeout)
		err = flow.Send(payload)
		sent := err == nil
		if err == nil {
			for {
				var b []byte
				b, err = flow.Receive(probe)
				if err != nil || bytes.Equal(b, payload) {
					break
				}
			}
		}
		stop()
		if err != nil {
			if errors.Is(err, telemetry.ErrIsolatedSession) {
				return optimizer.RouteMeasurement{}
			}
			if ctx.Err() != nil {
				m.Samples--
				break
			}
			switch {
			case telemetry.IsLocal(err):
				sample.Outcome = optimizer.LocalFailure
			case sent && errors.Is(err, context.DeadlineExceeded):
				sample.Outcome = optimizer.Timeout
				m.ConsecutiveLoss++
				m.MaxLossBurst = max(m.MaxLossBurst, m.ConsecutiveLoss)
			default:
				sample.Outcome = optimizer.TransportFailure
				m.TransportError = "transport_io_failed"
			}
			if e.OnFailure != nil && ctx.Err() == nil && (m.ConsecutiveLoss >= 3 || sample.Outcome == optimizer.TransportFailure) {
				e.OnFailure(c.Metadata.ID, "probe_unresponsive", time.Now())
			}
			if sample.Outcome != optimizer.Timeout {
				m.ConsecutiveLoss = 0
			}
			if !errors.Is(err, context.DeadlineExceeded) {
				failed++
			}
		} else {
			rtt := time.Since(start)
			sample.Outcome = optimizer.Reply
			sample.RTT = rtt
			m.ConsecutiveLoss = 0
			total += rtt
			if received > 0 {
				delta := rtt - previous
				if delta < 0 {
					delta = -delta
				}
				jitter += delta
			}
			previous = rtt
			received++
		}
		m.Probes = append(m.Probes, sample)
		if i+1 < e.Settings.Samples {
			timer := time.NewTimer(e.Settings.Spacing)
			select {
			case <-ctx.Done():
				timer.Stop()
			case <-timer.C:
			}
		}
	}
	if m.Samples > 0 {
		m.Loss = 1 - float64(received)/float64(m.Samples)
		m.FailureRate = float64(failed) / float64(m.Samples)
	}
	m.Available = received > 0
	if received > 0 {
		m.RTT = total / time.Duration(received)
	}
	if received > 1 {
		m.Jitter = jitter / time.Duration(received-1)
	}
	m.Stability = 1 - m.Loss
	m.ObservedAt = time.Now()
	return m
}

type direct struct {
	net.Conn
	mu sync.Mutex
}

func OpenDirect(ctx context.Context, dst netip.AddrPort) (routing.Datagram, error) {
	c, e := (&net.Dialer{}).DialContext(ctx, "udp", dst.String())
	if e != nil {
		return nil, e
	}
	return &direct{Conn: c}, nil
}
func (d *direct) Send(p []byte) error {
	d.SetWriteDeadline(time.Now().Add(time.Second))
	_, e := d.Write(p)
	return e
}
func (d *direct) Receive(ctx context.Context) ([]byte, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(time.Second)
	}
	d.SetReadDeadline(deadline)
	done := make(chan struct{})
	stop := context.AfterFunc(ctx, func() { d.SetReadDeadline(time.Now()); close(done) })
	defer func() {
		if !stop() {
			<-done
		}
	}()
	p := make([]byte, 1200)
	n, e := d.Read(p)
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if ne, ok := e.(net.Error); ok && ne.Timeout() {
		return nil, context.DeadlineExceeded
	}
	return p[:n], e
}

func (e *Engine) Candidates() []Candidate {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]Candidate(nil), e.candidates...)
}
