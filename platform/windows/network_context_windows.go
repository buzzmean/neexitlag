//go:build windows

package windowsplatform

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

type limitedOutput struct{ bytes.Buffer }

func (b *limitedOutput) Write(p []byte) (int, error) {
	if b.Len()+len(p) > 32768 {
		return 0, fmt.Errorf("network diagnostics output limit")
	}
	return b.Buffer.Write(p)
}

// Background-only read-only queries. Never called by the packet path. Failure is
// reported as unknown, not as proof that no VPN/driver/user conflict exists.
func inspectNetwork(ctx context.Context) NetworkContext {
	script := `$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue';$me=` + strconv.Itoa(os.Getpid()) + `;
$a=@(Get-NetAdapter|Sort-Object InterfaceIndex|ForEach-Object{"$($_.InterfaceIndex):$($_.InterfaceGuid):$($_.Status):$($_.InterfaceDescription)"});
$a+=@(Get-NetIPAddress|Where-Object{$_.AddressState -eq 'Preferred'}|Sort-Object InterfaceIndex,IPAddress|ForEach-Object{"$($_.InterfaceIndex):$($_.IPAddress):$($_.PrefixLength)"});
$a+=@(Get-NetConnectionProfile|Sort-Object InterfaceIndex|ForEach-Object{"$($_.InterfaceIndex):$($_.Name):$($_.NetworkCategory)"});
$r=@(Get-NetRoute -PolicyStore ActiveStore|Where-Object{$_.DestinationPrefix -in @('0.0.0.0/0','0.0.0.0/1','128.0.0.0/1','::/0') -or $_.DestinationPrefix -eq '2.56.177.4/32'}|Sort-Object InterfaceIndex,DestinationPrefix,NextHop|ForEach-Object{"$($_.InterfaceIndex):$($_.DestinationPrefix):$($_.NextHop):$($_.RouteMetric)"});
$d=@(Get-DnsClientServerAddress|Sort-Object InterfaceIndex,AddressFamily|ForEach-Object{"$($_.InterfaceIndex):$($_.AddressFamily):$($_.ServerAddresses -join ',')"});
$v=@(Get-NetAdapter|Where-Object{$_.Status -eq 'Up' -and ($_.InterfaceDescription+' '+$_.Name) -match 'VPN|Happ|TUN|TAP|WireGuard|Wintun'}|Sort-Object InterfaceIndex|ForEach-Object{$_.Name});
$c=@(Get-Process|Where-Object{$_.Id -ne $me -and $_.ProcessName -match '^(winws|winws2|goodbyedpi|easyanticheat.*|beservice.*|vgc)$'}|ForEach-Object{$_.ProcessName});
$unknown=$false;foreach($p in @(Get-Process|Where-Object{$_.Id -ne $me -and $_.Id -gt 4})){try{if(@($p.Modules|Where-Object{$_.ModuleName -ieq 'WinDivert.dll'}).Count -gt 0){$c+=($p.ProcessName+':WinDivert.dll')}}catch{$unknown=$true}};
$proxy='HKCU is service account, not interactive user; WinHTTP: '+((netsh winhttp show proxy|Out-String).Trim());
$doh='browser Secure DNS unknown; system DoH cmdlet unavailable';if(Get-Command Get-DnsClientDohServerAddress -ErrorAction SilentlyContinue){$doh='configured system DoH templates: '+@(Get-DnsClientDohServerAddress).Count+'; availability not probed'};
@{adapters=$a;routes=$r;dns=$d;vpn=$v;conflicts=$c;proxy=$proxy;secure_dns=$doh;egress='unmeasured';error=$(if($unknown){'Some process modules inaccessible; conflict scan incomplete'}else{''})}|ConvertTo-Json -Depth 4 -Compress`
	cmd := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script)
	var out limitedOutput
	cmd.Stdout = &out
	var stderr limitedOutput
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return NetworkContext{Error: "read-only Windows network inspection unavailable: " + err.Error()}
	}
	var n NetworkContext
	if err := json.Unmarshal(bytes.TrimPrefix(out.Bytes(), []byte{239, 187, 191}), &n); err != nil {
		return NetworkContext{Error: "invalid Windows diagnostic snapshot"}
	}
	// Incomplete module visibility must not invalidate observed route identity, but
	// is kept visible separately; no clean-conflict guarantee is made.
	incomplete := n.Error
	n.Error = ""
	if incomplete != "" {
		n.SecureDNS += "; " + incomplete
	}
	return n
}
