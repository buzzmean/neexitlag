//go:build windows

package windowsplatform

import (
	"fmt"
	"golang.org/x/sys/windows"
	"path/filepath"
	"unsafe"
)

type ProcessDiscovery struct{}

var kernel = windows.NewLazySystemDLL("kernel32.dll")
var queryImage = kernel.NewProc("QueryFullProcessImageNameW")

func NewDiscovery() Discovery { return &ProcessDiscovery{} }
func (*ProcessDiscovery) Lookup(pid uint32) (ProcessIdentity, error) {
	handle, e := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if e != nil {
		return ProcessIdentity{}, e
	}
	defer windows.CloseHandle(handle)
	var creation, exit, kern, user windows.Filetime
	if e = windows.GetProcessTimes(handle, &creation, &exit, &kern, &user); e != nil {
		return ProcessIdentity{}, e
	}
	buf := make([]uint16, 32768)
	n := uint32(len(buf))
	ok, _, callErr := queryImage.Call(uintptr(handle), 0, uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&n)))
	if ok == 0 {
		return ProcessIdentity{}, fmt.Errorf("query process image: %w", callErr)
	}
	path := windows.UTF16ToString(buf[:n])
	return ProcessIdentity{PID: pid, Path: path, Name: filepath.Base(path), CreationTime: uint64(creation.HighDateTime)<<32 | uint64(creation.LowDateTime)}, nil
}
func (p *ProcessDiscovery) List() ([]ProcessIdentity, error) {
	snapshot, e := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if e != nil {
		return nil, e
	}
	defer windows.CloseHandle(snapshot)
	entry := windows.ProcessEntry32{Size: uint32(unsafe.Sizeof(windows.ProcessEntry32{}))}
	if e = windows.Process32First(snapshot, &entry); e != nil {
		return nil, e
	}
	out := []ProcessIdentity{}
	for {
		if id, e := p.Lookup(entry.ProcessID); e == nil {
			out = append(out, id)
		}
		if windows.Process32Next(snapshot, &entry) != nil {
			break
		}
	}
	return out, nil
}

func IsCurrent(id ProcessIdentity) bool {
	if id.PID == 0 || id.CreationTime == 0 {
		return false
	}
	h, e := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, id.PID)
	if e != nil {
		return false
	}
	defer windows.CloseHandle(h)
	var creation, exit, kernel, user windows.Filetime
	if e = windows.GetProcessTimes(h, &creation, &exit, &kernel, &user); e != nil {
		return false
	}
	return uint64(creation.HighDateTime)<<32|uint64(creation.LowDateTime) == id.CreationTime
}
