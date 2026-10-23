package clientservice

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestVideoPerformanceSeparatesAdmissionOpenFirstByteAndRate(t *testing.T) {
	u := &browserRelayUsage{}
	start := time.Now().Add(-5 * time.Second)
	id := u.beginVideo(videoPerformance{Started: start, Profile: "chrome", Domain: "r.googlevideo.com", Ingress: "NATIVE_TCP"})
	u.videoOpened(id, start.Add(20*time.Millisecond), start.Add(100*time.Millisecond), nil)
	u.accountIO("chrome", "r.googlevideo.com", 100, 0, start.Add(150*time.Millisecond), id)
	u.accountIO("chrome", "r.googlevideo.com", 0, 1000000, start.Add(time.Second), id)
	u.accountIO("chrome", "r.googlevideo.com", 0, 1000000, start.Add(2*time.Second), id)
	u.mu.Lock()
	rows := u.videoSnapshot(start.Add(4 * time.Second))
	u.mu.Unlock()
	v := rows[0]
	if v.AdmissionMS != 20 || v.OpenMS != 80 || v.FirstByteMS != 1000 || v.WriteToFirstMS != 850 || v.AverageMbps != 4 || v.ReceiveSpanMbps != 8 || v.MaxReadGapMS != 1000 || v.Received != 2000000 || v.Sent != 100 {
		t.Fatalf("%+v", v)
	}
	u.videoClosed(id, false)
	a := u.snapshot().VideoConnections[0]
	time.Sleep(time.Millisecond)
	b := u.snapshot().VideoConnections[0]
	if a.AverageMbps != b.AverageMbps || a.State != "CLOSED" {
		t.Fatal("closed rate changed", a, b)
	}
}
func TestVideoPerformanceNoReplyBoundAndSnapshotAreHonest(t *testing.T) {
	u := &browserRelayUsage{}
	now := time.Now()
	for i := 0; i < videoPerformanceCapacity+10; i++ {
		id := u.beginVideo(videoPerformance{Started: now, Domain: fmt.Sprint(i)})
		u.videoClosed(id, false)
	}
	if len(u.videos) != videoPerformanceCapacity {
		t.Fatal(len(u.videos))
	}
	s := u.snapshot()
	if len(s.VideoConnections) != 24 || s.VideoOmitted != 114 {
		t.Fatal(s.VideoOmitted)
	}
	for _, r := range s.VideoConnections {
		if r.FirstByteMS != -1 || r.AverageMbps != -1 || r.ReceiveSpanMbps != -1 {
			t.Fatal("no reply fabricated", r)
		}
	}
	u = &browserRelayUsage{}
	for i := 0; i < videoPerformanceCapacity; i++ {
		if u.beginVideo(videoPerformance{Started: now}) == 0 {
			t.Fatal("early cap")
		}
	}
	if u.beginVideo(videoPerformance{Started: now}) != 0 || u.videoLost != 1 {
		t.Fatal("unbounded active ledger")
	}
}
func TestVideoPerformanceMeterActualIOAndClose(t *testing.T) {
	u := &browserRelayUsage{}
	start := time.Now()
	id := u.beginVideo(videoPerformance{Started: start})
	u.videoOpened(id, start, start, nil)
	a, b := net.Pipe()
	defer b.Close()
	c := &relayMeterConn{Conn: a, usage: u, videoID: id, release: func() {}}
	go func() { defer b.Close(); buf := make([]byte, 3); b.Read(buf); b.Write([]byte("reply")) }()
	if _, e := c.Write([]byte("ask")); e != nil {
		t.Fatal(e)
	}
	buf := make([]byte, 8)
	if n, e := c.Read(buf); n != 5 || e != nil {
		t.Fatal(n, e)
	}
	c.Close()
	c.Close()
	v := u.snapshot().VideoConnections[0]
	if v.Received != 5 || v.Sent != 3 || v.FirstByteMS < 0 || v.Closed.IsZero() {
		t.Fatal(v)
	}
}
