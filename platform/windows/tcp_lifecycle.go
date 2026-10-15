package windowsplatform

import (
	"errors"
	"io"
)

func tcpPacketDropCause(mode string) string {
	if mode == "closed" {
		return "TCP_TEARDOWN_TAIL"
	}
	return "POLICY_BLOCK"
}

// App-side reset/close and a graceful upstream EOF are normal TCP lifecycle.
// A broken upstream read or write is a transport failure even after prior replies.
func tcpCopyFailure(event string, readErr, writeErr error, stopping bool) string {
	if stopping {
		return ""
	}
	if event == "relay" && writeErr != nil {
		return "TCP_RELAY_WRITE_FAILED"
	}
	if event == "reply" && readErr != nil && !errors.Is(readErr, io.EOF) {
		return "TCP_RELAY_READ_FAILED"
	}
	return ""
}
func tcpNormalEnd(event string, readErr, writeErr error) bool {
	return (event == "reply" && errors.Is(readErr, io.EOF)) || (event == "relay" && readErr != nil && !errors.Is(readErr, io.EOF)) || (event == "reply" && writeErr != nil)
}
