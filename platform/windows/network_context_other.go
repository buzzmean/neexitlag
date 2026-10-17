//go:build !windows

package windowsplatform

import "context"

func inspectNetwork(context.Context) NetworkContext {
	return NetworkContext{Error: "Windows network context unavailable; AUTO local learning disabled"}
}
