package clientservice

import (
	"fmt"
	"net"
	"sync"
	"testing"
	"time"
)

func TestVideoLimitIsPerPCAndReleasesExactlyOnceOnClose(t *testing.T) {
	u := &browserRelayUsage{minutes: map[relayMinuteKey]relayMinute{}}
	var release []func()
	for i := 0; i < 4; i++ {
		r, ok := u.acquireVideo()
		if !ok {
			t.Fatal("under limit rejected")
		}
		release = append(release, r)
	}
	if _, ok := u.acquireVideo(); ok {
		t.Fatal("fifth video admitted")
	}
	release[0]()
	release[0]()
	r, ok := u.acquireVideo()
	if !ok {
		t.Fatal("permit not released")
	}
	r()
	for _, r := range release[1:] {
		r()
	}
	if s := u.snapshot(); s.VideoActive != 0 || s.VideoLimitHits != 1 {
		t.Fatal(s)
	}
}
func TestRelayMeterCountsActualActiveIOInCorrectMinuteProfileDomain(t *testing.T) {
	u := &browserRelayUsage{minutes: map[relayMinuteKey]relayMinute{}}
	a, b := net.Pipe()
	defer b.Close()
	released := 0
	c := &relayMeterConn{Conn: a, usage: u, profile: "edge", domain: "video.googlevideo.com", release: func() { released++ }}
	go func() { buf := make([]byte, 16); b.Read(buf); b.Write([]byte("response")) }()
	if n, e := c.Write([]byte("request")); e != nil || n != 7 {
		t.Fatal(n, e)
	}
	buf := make([]byte, 16)
	if n, e := c.Read(buf); e != nil || n != 8 {
		t.Fatal(n, e)
	}
	s := u.snapshot()
	if len(s.Minutes) != 1 || s.Minutes[0].Sent != 7 || s.Minutes[0].Received != 8 || s.Minutes[0].Profile != "edge" || s.Minutes[0].Domain != "video.googlevideo.com" {
		t.Fatal(s)
	}
	c.Close()
	c.Close()
	if released != 1 {
		t.Fatal("permit release repeated")
	}
	now := time.Now().UTC().Truncate(time.Minute)
	u.account("chrome", "youtube.com", 20, 30, now.Add(-time.Minute))
	u.account("chrome", "youtube.com", 50, 60, now)
	if len(u.snapshot().Minutes) != 3 {
		t.Fatal("minute/profile attribution lost")
	}
}
func TestRelayUsageIsBoundedAndReportsOverflowInsteadOfSilentlyLosingEvidence(t *testing.T) {
	u := &browserRelayUsage{minutes: map[relayMinuteKey]relayMinute{}}
	now := time.Now()
	var wg sync.WaitGroup
	for g := 0; g < 4; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < 400; i++ {
				u.account("p", fmt.Sprintf("%d-%d.example", g, i), 1, 1, now)
			}
		}(g)
	}
	wg.Wait()
	if len(u.minutes) > relayUsageCapacity {
		t.Fatal("unbounded usage")
	}
	s := u.snapshot()
	if !s.Truncated || s.OverflowBytes == 0 || len(s.Minutes) > 64 {
		t.Fatal(s)
	}
	u.account("p", "new.example", 1, 1, now.Add(relayUsageRetention+time.Minute))
	if len(u.minutes) != 1 {
		t.Fatal("retention did not evict old minutes")
	}
}
