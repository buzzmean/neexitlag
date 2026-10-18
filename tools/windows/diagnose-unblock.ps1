#requires -Version 5.1
<#
Read-only diagnostics for "blocked services do not open with VPN off".
Run as the installation owner (normal, non-elevated PowerShell is enough):
  powershell -NoProfile -ExecutionPolicy Bypass -File .\tools\windows\diagnose-unblock.ps1
It changes nothing: no service restart, no files outside the report folder, no
settings. IPC token, identity, device certificate and keys are never read into the
report. The only network traffic is TCP connects, one TLS handshake that merely
RECORDS the control-plane certificate (the product's own verification is untouched),
DNS lookups and HEAD requests to public sites.
#>
param(
 [string]$VpsHost='2.56.177.4',
 [int]$ControlPlanePort=8443,
 [string]$InstallDirectory="$env:ProgramData\neexitlag",
 [string]$BuildDirectory=(Join-Path (Split-Path $PSScriptRoot -Parent | Split-Path -Parent) 'bin\windows-e2e')
)
. (Join-Path $PSScriptRoot 'common.ps1')
$ErrorActionPreference='Continue'
$stamp=(Get-Date).ToString('yyyyMMdd-HHmmss')
$out=Join-Path $env:TEMP "veliq-diagnose-$stamp"
New-Item -ItemType Directory -Path $out -Force | Out-Null
$report=Join-Path $out 'report.txt'
function Say([string]$Text){$Text|Tee-Object -FilePath $report -Append}
function Section([string]$Title){Say '';Say ('=== '+$Title+' ===')}
# Last line of defence: never let a secret-shaped value reach the report.
function Scrub([string]$Text){
 $Text=[regex]::Replace($Text,'-----BEGIN [A-Z ]+-----[\s\S]*?-----END [A-Z ]+-----','[PEM REMOVED]')
 $Text=[regex]::Replace($Text,'(?i)("?(token|secret|private[_ ]?key|seed|password|authorization)"?\s*[:=]\s*)"?[^",\s}]+"?','$1"[REDACTED]"')
 return $Text
}
function Json($Value){Scrub ($Value|ConvertTo-Json -Depth 12)}
function Try-Step([string]$Name,[scriptblock]$Body){
 try{& $Body}catch{Say ("[$Name] FAILED: "+(Scrub $_.Exception.Message))}
}
function Prop($Object,[string]$Name,$Default=$null){
 if($null -ne $Object -and $Object.PSObject.Properties.Name -contains $Name){return $Object.$Name};return $Default
}

Section 'Environment'
Try-Step 'os' {
 $os=Get-CimInstance Win32_OperatingSystem
 Say ("OS: $($os.Caption) build $($os.BuildNumber); user: $env:USERNAME; PowerShell $($PSVersionTable.PSVersion)")
 Say ("Elevated: "+([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator))
 $vpn=@(Get-NetAdapter|Where-Object{$_.Status -eq 'Up' -and ($_.InterfaceDescription+' '+$_.Name) -match 'VPN|Happ|TUN|TAP|WireGuard|Wintun'}|ForEach-Object{$_.Name})
 Say ("Active VPN-like adapters: "+$(if($vpn.Count){$vpn -join ', '}else{'none'}))
 Say ("System DNS: "+((Get-DnsClientServerAddress -AddressFamily IPv4|Where-Object{$_.ServerAddresses}|ForEach-Object{$_.ServerAddresses -join ','}|Select-Object -Unique) -join '; '))
 Say ("WinHTTP proxy: "+((netsh winhttp show proxy|Out-String).Trim() -replace '\s+',' '))
}

Section 'Service and binary'
Try-Step 'service' {
 $svc=Get-Service neexitlag -ErrorAction Stop
 Say ("Service neexitlag: $($svc.Status), start type $($svc.StartType)")
 $installed=Join-Path $InstallDirectory 'neexitlag-service.exe'
 $built=Join-Path $BuildDirectory 'neexitlag-service.exe'
 foreach($pair in @(@('installed',$installed),@('built',$built))){
  if(Test-Path $pair[1]){Say ("SHA256 $($pair[0]): "+(Get-FileHash $pair[1] -Algorithm SHA256).Hash+"  modified "+(Get-Item $pair[1]).LastWriteTime)}
  else{Say ("SHA256 $($pair[0]): file not found ($($pair[1]))")}
 }
 if((Test-Path $installed) -and (Test-Path $built)){
  Say ("Installed == built: "+((Get-FileHash $installed).Hash -eq (Get-FileHash $built).Hash))
 }
}

Section 'Service IPC (status)'
$status=$null
Try-Step 'status' {
 $script:status=Invoke-ServiceIPC 'status'
 $s=$script:status
 Say ("optimizing: "+(Prop $s 'optimizing'))
 Say ("settings: "+(Json (Prop $s 'settings')))
 Say ("relays: "+(Json (Prop $s 'relays')))
 Say ("routes: "+(Json (Prop $s 'routes')))
 Say ("active flows (first 20): "+(Json (@(Prop $s 'flows' @())|Select-Object -First 20)))
}
Section 'Profiles (what "Optimize" actually covers)'
Try-Step 'profiles' {
 $rows=@(Invoke-ServiceIPC 'profiles')
 Say ("Profile count: "+$rows.Count)
 foreach($p in $rows){
  Say ("- id=$(Prop $p 'id') name=$(Prop $p 'name') mode=$(Prop $p 'mode') fallback=$(Prop $p 'fallback') preset=$(Prop $p 'app_preset') exe=$((@(Prop $p 'executables' @()))-join ' | ') rules=$(@(Prop $p 'rules' @()).Count)")
 }
 $hasChrome=@($rows|Where-Object{(@(Prop $_ 'executables' @())-join ' ') -match 'chrome\.exe'}).Count
 $hasTelegram=@($rows|Where-Object{(@(Prop $_ 'executables' @())-join ' ') -match 'telegram\.exe'}).Count
 Say ("Profile for Chrome: "+($hasChrome -gt 0)+"; for Telegram: "+($hasTelegram -gt 0))
}
Section 'Browser transport (Chrome PAC/proxy)'
Try-Step 'browser' {Say (Json (Invoke-ServiceIPC 'browser.status'))}
Section 'Capture diagnostics'
Try-Step 'capture' {Say (Json (Invoke-ServiceIPC 'capture.diagnostics'))}
Section 'Quality diagnostics'
Try-Step 'quality' {Say (Json (Invoke-ServiceIPC 'quality.diagnostics'))}
Section 'Relay pool diagnostics (handshake / x509 errors appear here)'
Try-Step 'pool' {
 $ids=@()
 if($script:status){$ids=@(Prop $script:status 'relays' @()|ForEach-Object{Prop $_ 'id'}|Where-Object{$_})}
 if($ids.Count -eq 0){$ids=@('VPS-01')}
 foreach($id in $ids){Say ("relay $id : "+(Json (Invoke-ServiceIPC 'relay.pool.diagnostics' @{relay_id=$id})))}
}
Section 'Service diagnostics'
Try-Step 'diag' {Say (Json (Invoke-ServiceIPC 'diagnostics'))}

Section "Reachability of $VpsHost (this ISP, this PC)"
Try-Step 'tcp' {
 $client=New-Object Net.Sockets.TcpClient
 $watch=[Diagnostics.Stopwatch]::StartNew()
 $task=$client.ConnectAsync($VpsHost,$ControlPlanePort)
 if($task.Wait(5000) -and $client.Connected){Say ("TCP $ControlPlanePort : OPEN in $($watch.ElapsedMilliseconds) ms")}
 else{Say "TCP $ControlPlanePort : NO CONNECTION within 5 s"}
 if($client.Connected){
  # Record the presented certificate only; verification outcome is reported, not changed.
  $policy=@{Errors='unknown'}
  $callback=[Net.Security.RemoteCertificateValidationCallback]{param($s,$c,$ch,$e) $policy.Errors=[string]$e;$true}
  $ssl=New-Object Net.Security.SslStream($client.GetStream(),$false,$callback)
  $ssl.AuthenticateAsClient($VpsHost)
  $cert=New-Object Security.Cryptography.X509Certificates.X509Certificate2($ssl.RemoteCertificate)
  $san=($cert.Extensions|Where-Object{$_.Oid.Value -eq '2.5.29.17'}|ForEach-Object{$_.Format($false)}) -join ''
  Say ("Control-plane TLS: protocol $($ssl.SslProtocol); subject $($cert.Subject); issuer $($cert.Issuer)")
  Say ("  valid $($cert.NotBefore.ToString('yyyy-MM-dd')) .. $($cert.NotAfter.ToString('yyyy-MM-dd')); SAN: $san")
  Say ("  system-store validation result (private CA is expected to be untrusted here): $($policy.Errors)")
  $ssl.Dispose()
 }
 $client.Close()
}
Say 'UDP 4433 (QUIC relay): UDP has no connect test. The relay result above (pool diagnostics, quality.diagnostics, status relays) is the real evidence; if every relay attempt times out while TCP 8443 is open, the ISP is probably dropping UDP/QUIC to this host.'

Section 'DNS and plain HTTPS with the router NOT in the path of curl'
$targets=@('chatgpt.com','claude.ai','www.youtube.com','discord.com','web.telegram.org','t.me')
foreach($name in $targets){
 Try-Step "dns $name" {
  $ips=@(Resolve-DnsName $name -Type A -DnsOnly -ErrorAction Stop|Where-Object{$_.Type -eq 'A'}|ForEach-Object{$_.IPAddress})
  $http=(& curl.exe -sS -m 10 -o NUL -I -w '%{http_code} connect=%{time_connect}s tls=%{time_appconnect}s' "https://$name" 2>&1|Out-String).Trim()
  Say ("$name -> $($ips -join ',') | curl: $http")
 }
}

Section 'Installed applications and browser coverage'
Try-Step 'apps' {
 $paths=[ordered]@{
  'Chrome'=@("$env:ProgramFiles\Google\Chrome\Application\chrome.exe","${env:ProgramFiles(x86)}\Google\Chrome\Application\chrome.exe","$env:LOCALAPPDATA\Google\Chrome\Application\chrome.exe")
  'Edge'=@("${env:ProgramFiles(x86)}\Microsoft\Edge\Application\msedge.exe","$env:ProgramFiles\Microsoft\Edge\Application\msedge.exe")
  'Firefox'=@("$env:ProgramFiles\Mozilla Firefox\firefox.exe","${env:ProgramFiles(x86)}\Mozilla Firefox\firefox.exe")
  'Telegram Desktop'=@("$env:APPDATA\Telegram Desktop\Telegram.exe","$env:ProgramFiles\Telegram Desktop\Telegram.exe")
  'Discord'=@("$env:LOCALAPPDATA\Discord\Update.exe")
 }
 foreach($k in $paths.Keys){
  $found=@($paths[$k]|Where-Object{Test-Path $_})
  Say ("$k : "+$(if($found.Count){'installed at '+($found -join '; ')}else{'not found'}))
 }
 Say ("Running now: "+((Get-Process -ErrorAction SilentlyContinue|Where-Object{$_.ProcessName -match '^(chrome|msedge|firefox|telegram|discord)$'}|Select-Object -ExpandProperty ProcessName -Unique) -join ', '))
}

Section 'Done'
Say ("Report folder: $out")
Say 'Send report.txt back. It was scrubbed of tokens/keys, but read it once before sharing.'
