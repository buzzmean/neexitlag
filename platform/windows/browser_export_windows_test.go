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

func (b *WinDivertBackend) TraceNativeFixture(f func(string, Tuple, byte, byte)) {
	if !b.testBrowserLoopback {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.tcp != nil {
		b.tcp.fixtureTrace.Store(&nativeTCPTraceHook{f: f})
	}
}
