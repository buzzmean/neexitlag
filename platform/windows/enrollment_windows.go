//go:build windows

package windowsplatform

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/buzzmean/neexitlag/core/enrollment"
	"golang.org/x/sys/windows"
	"os/exec"
	"syscall"
)

// Metadata only. No memory read/injection, MachineGuid, raw hardware logs or argv.
func EnrollmentHardware(ctx context.Context) (enrollment.Hardware, error) {
	script := `$ErrorActionPreference='Stop';$u=(Get-CimInstance Win32_ComputerSystemProduct).UUID;$b=(Get-CimInstance Win32_BaseBoard|Select-Object -First 1).SerialNumber;$a=@(Get-NetAdapter -Physical|Where-Object Status -eq 'Up');$r=Get-NetRoute -AddressFamily IPv4 -DestinationPrefix '0.0.0.0/0'|Where-Object {$a.ifIndex -contains $_.ifIndex}|Sort-Object RouteMetric,ifIndex|Select-Object -First 1;$n=$a|Where-Object ifIndex -eq $r.ifIndex|Select-Object -First 1;if(!$n){throw 'NO_PHYSICAL_NIC'};@{uuid=[string]$u;board_serial=[string]$b;mac=[string]$n.MacAddress}|ConvertTo-Json -Compress`
	cmd := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	raw, e := cmd.Output()
	if e != nil || len(raw) > 4096 {
		return enrollment.Hardware{}, errors.New("HARDWARE_ID_UNAVAILABLE")
	}
	var h enrollment.Hardware
	if json.Unmarshal(raw, &h) != nil {
		return h, errors.New("HARDWARE_ID_UNAVAILABLE")
	}
	return h, nil
}
func ProtectActivationFile(path string) error {
	d, e := windows.SecurityDescriptorFromString("D:P(A;;FA;;;SY)(A;;FA;;;BA)")
	if e != nil {
		return e
	}
	acl, _, e := d.DACL()
	if e != nil {
		return e
	}
	return windows.SetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, acl, nil)
}
