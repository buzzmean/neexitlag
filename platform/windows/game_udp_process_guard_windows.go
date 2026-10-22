//go:build windows

package windowsplatform

import (
	"errors"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
	"strings"
	"time"
	"unsafe"
)

// Names and SCM state only. No process handles, module lists, memory, injection
// or anti-cheat probing. REFLECT additionally identifies overlapping v2 filters.
func gameProcessGuard() (out GameUDPGuardStatus) {
	h, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		out.Error = "GAME_PROCESS_LIST_UNAVAILABLE"
		return
	}
	defer windows.CloseHandle(h)
	e := windows.ProcessEntry32{Size: uint32(unsafe.Sizeof(windows.ProcessEntry32{}))}
	err = windows.Process32First(h, &e)
	count := 0
	for err == nil && count < 4096 {
		count++
		name := strings.ToLower(windows.UTF16ToString(e.ExeFile[:]))
		switch name {
		case "winws.exe", "winws2.exe", "goodbyedpi.exe":
			out.Conflict = true
		}
		observeGameAntiCheatProcess(&out, name)
		err = windows.Process32Next(h, &e)
	}
	if !errors.Is(err, windows.ERROR_NO_MORE_FILES) {
		out.Error = "GAME_PROCESS_LIST_INCOMPLETE"
		return
	}
	m, err := mgr.Connect()
	if err != nil {
		out.AntiCheatQueryError = "GAME_DRIVER_STATUS_UNAVAILABLE"
		return
	}
	defer m.Disconnect()
	for _, name := range []string{"vgk", "vgc", "FACEIT", "FACEITService"} {
		s, err := m.OpenService(name)
		if errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
			continue
		}
		if err != nil {
			out.AntiCheatQueryError = "GAME_DRIVER_STATUS_UNAVAILABLE"
			continue
		}
		st, err := s.Query()
		s.Close()
		if err != nil {
			out.AntiCheatQueryError = "GAME_DRIVER_STATUS_UNAVAILABLE"
			continue
		}
		observeGameAntiCheatService(&out, name, st.State != svc.Stopped)
	}
	return
}
func (b *WinDivertBackend) gameProcessStatus() GameUDPGuardStatus {
	return b.gameGuard.processSnapshot(time.Now(), gameProcessGuard)
}
