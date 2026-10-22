//go:build windows

package windowsplatform

import (
	"crypto/sha256"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

func TestPrivateISCSIExcludedByActualWinDivertFilterEvaluator(t *testing.T) {
	dir := os.Getenv("NEEXITLAG_TEST_DRIVER_DIR")
	if dir == "" {
		t.Skip("pinned WinDivert DLL required; no filter match proof")
	}
	path := filepath.Join(dir, "WinDivert.dll")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "c1e060ee19444a259b2162f8af0f3fe8c4428a1c6f694dce20de194ac8d7d9a2" {
		t.Fatal("unapproved fixture DLL")
	}
	dll, err := windows.LoadDLL(path)
	if err != nil {
		t.Fatal(err)
	}
	defer dll.Release()
	eval, err := dll.FindProc("WinDivertHelperEvalFilter")
	if err != nil {
		t.Fatal(err)
	}
	filter, _ := windows.BytePtrFromString(nativeTCPNetworkFilter(50165, false))
	cases := []struct {
		src, dst   string
		out, match bool
	}{
		{"192.168.88.10:50000", "192.168.88.252:3260", true, false},
		{"192.168.88.252:3260", "192.168.88.10:50000", false, false},
		{"192.168.88.10:50000", "8.8.8.8:443", true, true},
		{"192.168.88.10:50000", "8.8.8.8:3260", true, true},
	}
	for _, tt := range cases {
		packet := tcpPacket(t, Tuple{Protocol: "tcp", Source: netip.MustParseAddrPort(tt.src), Destination: netip.MustParseAddrPort(tt.dst)})
		var a [80]byte
		if tt.out {
			a[10] = 2
		}
		result, _, _ := eval.Call(uintptr(unsafe.Pointer(filter)), uintptr(unsafe.Pointer(&packet[0])), uintptr(len(packet)), uintptr(unsafe.Pointer(&a[0])))
		if (result != 0) != tt.match {
			t.Fatalf("actual filter src=%s dst=%s matched=%v expected=%v", tt.src, tt.dst, result != 0, tt.match)
		}
	}
}
