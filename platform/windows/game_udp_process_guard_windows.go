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
func gameProcessGuard() (foreign, anti, errorCode string) {
	h, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return "", "", "GAME_PROCESS_LIST_UNAVAILABLE"
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
			foreign = "FOREIGN_UDP_INTERCEPTOR"
		case "vgc.exe", "faceit.exe", "faceitclient.exe", "faceitservice.exe":
			anti = "VANGUARD_OR_FACEIT"
		}
		err = windows.Process32Next(h, &e)
	}
	if !errors.Is(err, windows.ERROR_NO_MORE_FILES) {
		return foreign, anti, "GAME_PROCESS_LIST_INCOMPLETE"
	}
	m, err := mgr.Connect()
	if err != nil {
		return foreign, anti, "GAME_DRIVER_STATUS_UNAVAILABLE"
	}
	defer m.Disconnect()
	for _, name := range []string{"vgk", "vgc", "FACEIT", "FACEITService"} {
		s, err := m.OpenService(name)
		if errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
			continue
		}
		if err != nil {
			return foreign, anti, "GAME_DRIVER_STATUS_UNAVAILABLE"
		}
		st, err := s.Query()
		s.Close()
		if err != nil {
			return foreign, anti, "GAME_DRIVER_STATUS_UNAVAILABLE"
		}
		if st.State != svc.Stopped {
			anti = "VANGUARD_OR_FACEIT"
		}
	}
	return
}
func (b *WinDivertBackend) gameProcessStatus() GameUDPGuardStatus {
	g := &b.gameGuard
	g.mu.Lock()
	defer g.mu.Unlock()
	if time.Since(g.processAt) >= time.Second {
		g.processForeign, g.processAnti, g.processError = gameProcessGuard()
		g.processAt = time.Now()
	}
	return GameUDPGuardStatus{Conflict: g.processForeign != "", DriverAntiCheat: g.processAnti, Error: g.processError}
}
