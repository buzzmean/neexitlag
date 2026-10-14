//go:build windows

package windowsplatform

import (
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiscoverCurrentWindowsProcessWithoutCapture(t *testing.T) {
	d := NewDiscovery()
	id, e := d.Lookup(uint32(os.Getpid()))
	if e != nil {
		t.Fatal(e)
	}
	exe, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	if id.CreationTime == 0 || !strings.EqualFold(filepath.Clean(id.Path), filepath.Clean(exe)) {
		t.Fatal("identity", id)
	}
	all, e := d.List()
	if e != nil {
		t.Fatal(e)
	}
	found := false
	for _, p := range all {
		if p.PID == id.PID && p.CreationTime == id.CreationTime {
			found = true
		}
	}
	if !found {
		t.Fatal("process not discovered")
	}
}
func TestRejectMissingAndUnpinnedDriverWithoutCapture(t *testing.T) {
	dir := t.TempDir()
	if _, e := loadDivert(DriverFiles{Directory: dir}); e == nil {
		t.Fatal("unconfigured pins accepted")
	}
	if e := os.WriteFile(filepath.Join(dir, "WinDivert.dll"), []byte("not a DLL"), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e := loadDivert(DriverFiles{Directory: dir, DLLSHA256: strings.Repeat("0", 64), DriverSHA256: strings.Repeat("0", 64)}); e == nil {
		t.Fatal("untrusted DLL accepted")
	}
}

func TestNativeUDPOwnerOfCurrentSocket(t *testing.T) {
	socket, e := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if e != nil {
		t.Fatal(e)
	}
	defer socket.Close()
	source := netip.MustParseAddrPort(socket.LocalAddr().String())
	pid, e := resolveUDPOwner(source)
	if e != nil || pid != uint32(os.Getpid()) {
		t.Fatal("native UDP owner", pid, e)
	}
	cache := NewProcessCache(NewDiscovery(), IsCurrent)
	identity, e := cache.Lookup(pid)
	if e != nil || identity.CreationTime == 0 || identity.Path == "" || !IsCurrent(identity) {
		t.Fatal("native identity", identity, e)
	}
}
func TestNativeDualStackUDPOwner(t *testing.T) {
	socket, e := net.ListenUDP("udp", nil)
	if e != nil {
		t.Fatal(e)
	}
	defer socket.Close()
	source := netip.AddrPortFrom(netip.MustParseAddr("127.0.0.1"), uint16(socket.LocalAddr().(*net.UDPAddr).Port))
	pid, e := resolveUDPOwner(source)
	if e != nil || pid != uint32(os.Getpid()) {
		t.Fatal("native dual-stack owner", pid, e)
	}
}

func TestNativeTCPOwnerOfFullTuple(t *testing.T) {
	l, e := net.Listen("tcp4", "127.0.0.1:0")
	if e != nil {
		t.Fatal(e)
	}
	defer l.Close()
	c, e := net.Dial("tcp4", l.Addr().String())
	if e != nil {
		t.Fatal(e)
	}
	defer c.Close()
	tuple := Tuple{"tcp", c.LocalAddr().(*net.TCPAddr).AddrPort(), c.RemoteAddr().(*net.TCPAddr).AddrPort()}
	pid, e := resolveTCPOwner(tuple)
	if e != nil || pid != uint32(os.Getpid()) {
		t.Fatalf("native TCP owner=%d err=%v", pid, e)
	}
	tuple.Destination = netip.AddrPortFrom(tuple.Destination.Addr(), 1)
	if _, e = resolveTCPOwner(tuple); e == nil {
		t.Fatal("unrelated tuple attributed")
	}
}
