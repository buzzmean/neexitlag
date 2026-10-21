package clientservice

import (
	"context"
	"github.com/buzzmean/neexitlag/core/browserproxy"
	"github.com/buzzmean/neexitlag/core/flowpolicy"
	"io"
	"net"
	"testing"
	"time"
)

func TestDisabledBrowserAdmissionUsesDirectSocketWithoutNativeOrRelay(t *testing.T) {
	l, e := net.Listen("tcp4", "127.0.0.1:0")
	if e != nil {
		t.Fatal(e)
	}
	defer l.Close()
	go func() {
		for {
			c, e := l.Accept()
			if e != nil {
				return
			}
			go func() { defer c.Close(); io.Copy(c, c) }()
		}
	}()
	dest := l.Addr().(*net.TCPAddr).AddrPort()
	c := New(context.Background())
	defer c.Close()
	for _, host := range []string{"vk.ru", "chatgpt.com", "www.youtube.com"} {
		plan := browserproxy.Plan{Enabled: true, Destination: dest, Policy: flowpolicy.Result{Action: flowpolicy.RelayRequired, RelayID: "VPS-01", Fallback: flowpolicy.Required}}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		conn, actual, e := c.openBrowserPlan(ctx, profilesResult{}, browserproxy.Request{Host: host, Port: dest.Port()}, plan, func(ctx context.Context, _ browserproxy.Request, p browserproxy.Plan) (net.Conn, error) {
			return openBrowserDirectCompatibility(ctx, p)
		})
		if e != nil {
			cancel()
			t.Fatal(e)
		}
		if actual.Enabled || actual.Policy.Action != flowpolicy.Direct || actual.Policy.RelayID != "" || actual.Probe != "BROWSER_DISABLED_DIRECT" {
			t.Fatal(actual)
		}
		conn.SetDeadline(time.Now().Add(time.Second))
		conn.Write([]byte("payload"))
		b := make([]byte, 7)
		_, e = io.ReadFull(conn, b)
		conn.Close()
		cancel()
		if e != nil || string(b) != "payload" {
			t.Fatal(e, string(b))
		}
	}
}
func TestDisabledListenerRetiresOnlyAfterIdleAndRebindsOnEnable(t *testing.T) {
	c, _, _ := silentFixture(t, false)
	old := c.browser.server
	c.retireDisabledBrowserListeners(time.Now(), 3*time.Minute)
	if c.browser.server != old {
		t.Fatal("retired too early")
	}
	c.retireDisabledBrowserListeners(time.Now().Add(4*time.Minute), 3*time.Minute)
	if c.browser.server != nil || old.Snapshot().Listening {
		t.Fatal("idle listener retained")
	}
	c.browser.mu.Lock()
	e := c.ensureBrowserListenerLocked(false)
	c.browser.mu.Unlock()
	if e != nil {
		t.Fatal(e)
	}
	if c.browser.server == nil || !c.browser.server.Snapshot().Listening {
		t.Fatal("cannot rebind after idle")
	}
}
