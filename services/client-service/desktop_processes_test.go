package clientservice

import (
	windowsplatform "github.com/buzzmean/neexitlag/platform/windows"
	"testing"
)

func TestDesktopProcessFilterPreservesWindowsIdentity(t *testing.T) {
	all := []windowsplatform.ProcessIdentity{{PID: 1, Path: `C:\games\game.exe`, CreationTime: 100}, {PID: 2, Path: `C:\other\game.exe`, CreationTime: 200}, {PID: 3, Path: `C:\system\other.exe`, CreationTime: 300}}
	got := filterProcesses(all, []string{`c:/games/GAME.exe`})
	if len(got) != 1 || got[0].PID != 1 || got[0].CreationTime != 100 {
		t.Fatal(got)
	}
	got = filterProcesses(all, []string{"game.exe"})
	if len(got) != 2 {
		t.Fatal(got)
	}
	if len(filterProcesses(all, nil)) != 3 {
		t.Fatal("legacy query changed")
	}
}
