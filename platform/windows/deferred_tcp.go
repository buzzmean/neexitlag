package windowsplatform

import (
	"github.com/buzzmean/neexitlag/core/flowpolicy"
	"github.com/buzzmean/neexitlag/core/localstrategy"
	"io"
	"net"
	"time"
)

// ReadDeferredHello retains original bytes (at most 32 KiB for two seconds),
// returning them even on incomplete/unsupported input. It never logs payload.
func ReadDeferredHello(c net.Conn) ([]byte, string, string) {
	c.SetReadDeadline(time.Now().Add(localstrategy.HelloDeadline))
	defer c.SetReadDeadline(time.Time{})
	data := make([]byte, 0, 4096)
	reads := 0
	for len(data) < localstrategy.MaxHelloBytes && reads < localstrategy.MaxHelloSegments {
		size := 4096
		if size > localstrategy.MaxHelloBytes-len(data) {
			size = localstrategy.MaxHelloBytes - len(data)
		}
		b := make([]byte, size)
		n, err := c.Read(b)
		reads++
		data = append(data, b[:n]...)
		if n > 0 {
			domain, reason, done := localstrategy.InspectHelloStream(data)
			if done {
				return data, domain, reason
			}
		}
		if err != nil {
			if err == io.EOF && len(data) == 0 {
				return nil, "", "CLIENT_CLOSED_BEFORE_DATA"
			}
			return data, "", "CLIENTHELLO_INCOMPLETE"
		}
	}
	return data, "", "REASSEMBLY_LIMIT"
}

func ResolveDeferredDecision(initial Decision, domain, reason string) Decision {
	if reason == "ECH_OUTER_SNI_ONLY" {
		domain = ""
	}
	if initial.ResolveDomain == nil {
		initial.Block = true
		initial.Deferred = false
		initial.Policy.Action = flowpolicy.Block
		initial.Policy.Fallback = flowpolicy.FailBlock
		initial.Policy.Reason = "missing deferred resolver"
		return initial
	}
	d := initial.ResolveDomain(domain)
	d.Deferred = false
	d.ResolveDomain = nil
	// Missing SNI is not a failed relay open. Chrome sends an ECH extension
	// even for GREASE; keep an explicit authenticated opaque-name route alive.
	// Other policies without such a route still fail closed.
	if domain == "" && d.ConnectOrigin == "" && d.BlockOnFailure && (d.Policy.Action != flowpolicy.RelayRequired || d.Route.Direct() || d.Route.Local()) {
		d.Block = true
	}
	if d.Block {
		d.Policy.Action = flowpolicy.Block
		if d.Policy.Fallback == flowpolicy.FailOpen {
			d.Policy.Fallback = flowpolicy.FailBlock
		}
	}
	return d
}
