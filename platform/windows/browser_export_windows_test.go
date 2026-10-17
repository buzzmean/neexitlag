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
