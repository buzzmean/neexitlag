//go:build windows

package windowsplatform

import (
	"errors"
	"golang.org/x/sys/windows"
	"net/netip"
	"unsafe"
)

var getUDPTable = windows.NewLazySystemDLL("iphlpapi.dll").NewProc("GetExtendedUdpTable")

// Resolve only unique owners. Reused/wildcard sockets with multiple PIDs bypass.
func resolveUDPOwner(source netip.AddrPort) (uint32, error) {
	var size uint32
	result, _, _ := getUDPTable.Call(0, uintptr(unsafe.Pointer(&size)), 1, 2, 1, 0)
	if uint32(result) != uint32(windows.ERROR_INSUFFICIENT_BUFFER) || size < 4 || size > 8*1024*1024 {
		return 0, errors.New("UDP owner table unavailable")
	}
	for tries := 0; tries < 2; tries++ {
		b := make([]byte, size)
		result, _, _ = getUDPTable.Call(uintptr(unsafe.Pointer(&b[0])), uintptr(unsafe.Pointer(&size)), 1, 2, 1, 0)
		if uint32(result) == uint32(windows.ERROR_INSUFFICIENT_BUFFER) && size <= 8*1024*1024 {
			continue
		}
		if result != 0 {
			return 0, errors.New("UDP owner table query failed")
		}
		return udpOwner(b, source)
	}
	return 0, errors.New("UDP owner table changed")
}
