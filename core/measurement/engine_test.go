package measurement

import (
	"context"
	"errors"
	"github.com/buzzmean/neexitlag/core/optimizer"
	"github.com/buzzmean/neexitlag/core/routing"
	"github.com/buzzmean/neexitlag/tools/echo"
	"net/netip"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRealEchoAndCancellation(t *testing.T) {
	a, close, err := echo.UDP()
	if err != nil {
		t.Fatal(err)
	}
	defer close()
	e := New(Settings{Samples: 3, Timeout: 100 * time.Millisecond, Spacing: 20 * time.Millisecond, Concurrency: 1})
	c := Candidate{Metadata: optimizer.RelayCandidate{ID: optimizer.Direct}, Target: netip.MustParseAddrPort(a.String()), Open: OpenDirect}
	m := e.Probe(context.Background(), c)
	// Windows loopback may complete within a single clock tick; preserve measured zero.
	if !m.Available || m.Loss != 0 || m.RTT < 0 || m.Samples != 3 {
		t.Fatal(m)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	e.SetCandidates([]Candidate{c})
	if len(e.Round(ctx)) > 1 {
		t.Fatal("cancel")
	}
}

// Two callers (background probing and diagnostics) share one global budget.
func TestConcurrentRoundsShareProbeBudgetAndCancel(t *testing.T) {
	e := New(Settings{Samples: 1, Timeout: time.Second, Concurrency: 2})
	var active, peak atomic.Int32
	started := make(chan struct{}, 32)
	opener := func(ctx context.Context, _ netip.AddrPort) (routing.Datagram, error) {
		n := active.Add(1)
		defer active.Add(-1)
		for old := peak.Load(); n > old; old = peak.Load() {
			if peak.CompareAndSwap(old, n) {
				break
			}
		}
		started <- struct{}{}
		<-ctx.Done()
		return nil, errors.New("cancelled probe")
	}
	candidates := make([]Candidate, 8)
	for i := range candidates {
		candidates[i] = Candidate{Metadata: optimizer.RelayCandidate{ID: "relay"}, Target: netip.MustParseAddrPort("127.0.0.1:9000"), Open: opener}
	}
	e.SetCandidates(candidates)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); e.Round(ctx) }()
	}
	for i := 0; i < 2; i++ {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("probe did not start")
		}
	}
	select {
	case <-started:
		t.Error("global concurrency exceeded")
	case <-time.After(30 * time.Millisecond):
	}
	cancel()
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("cancel did not release rounds")
	}
	if peak.Load() > 2 || active.Load() != 0 {
		t.Fatal("probe budget leaked", peak.Load(), active.Load())
	}
}
