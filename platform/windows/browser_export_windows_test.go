//go:build windows

package windowsplatform

import "github.com/buzzmean/neexitlag/core/routing"

// Only compiled into this package's test binary. No production API enables the
// loopback driver seam, changes destination policy or certificate validation.
func NewBrowserLoopbackFixture(files DriverFiles, engine *routing.Engine) *WinDivertBackend {
	b := NewWinDivert(files, engine)
	b.testBrowserLoopback = true
	return b
}

// Exact fixture port only; never exported by a production build.
func (b *WinDivertBackend) NativeFixtureListenerPort() uint16 {
	if !b.testBrowserLoopback {
		return 0
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.tcp == nil || !b.status.Ready {
		return 0
	}
	return b.tcp.port
}
