#requires -Version 5.1
<#
Non-interactive, read-only diagnostics to run with the VPN OFF, as a normal user:
  powershell -NoProfile -ExecutionPolicy Bypass -File .\tools\windows\diag-no-vpn.ps1
Writes diag-result.txt next to this script (falls back to %TEMP% if that folder is not
writable). Changes nothing on the machine. Reads no token/identity/certificate into the
report; the whole report is scrubbed of secret-shaped values before the run ends.
Self-contained: does not need common.ps1. -Only limits the run to some sections:
  env, service, ipc, tcp, udp, mtu, dns, sites, sni, browser, apps, youtube
  (youtube = SNI table for www.youtube.com + status flag + stall counters + YouTube records; run: & .\tools\windows\diag-no-vpn.ps1 -Only youtube)
#>
param(
 [string]$VpsHost='2.56.177.4',
 [int]$ControlPlanePort=8443,
 [int]$QuicPort=4433,
 [string]$InstallDirectory="$env:ProgramData\neexitlag",
 [string]$BuildDirectory='',
 [string[]]$Only=@(),
 [string]$ProxyAddress='',   # host:port of the local browser proxy; default: taken from browser.status
 [string]$OutputFile=''
)
$ErrorActionPreference='Continue'
Set-StrictMode -Version 2
$Only=@($Only|ForEach-Object{$_ -split ','}|Where-Object{$_})   # -File passes 'a,b' as one string
$scriptDir=if($PSScriptRoot){$PSScriptRoot}else{(Get-Location).Path}
if(-not $BuildDirectory){$BuildDirectory=Join-Path (Split-Path $scriptDir -Parent | Split-Path -Parent) 'bin\windows-e2e'}
$result=if($OutputFile){$OutputFile}else{Join-Path $scriptDir 'diag-result.txt'}
$utf8=New-Object Text.UTF8Encoding($false)      # appends: the BOM is written once, below
$utf8Bom=New-Object Text.UTF8Encoding($true)
$bom=[byte[]](0xEF,0xBB,0xBF)                    # UTF-8 BOM so Windows tools show Cyrillic correctly
try{[IO.File]::WriteAllBytes($result,$bom)}catch{$result=Join-Path ([IO.Path]::GetTempPath()) 'diag-result.txt';[IO.File]::WriteAllBytes($result,$bom)}
$isWin=($PSVersionTable.PSVersion.Major -lt 6) -or $IsWindows

function Say([string]$Text){Write-Host $Text;[IO.File]::AppendAllText($result,$Text+"`r`n",$utf8)}
function Want([string]$Section){return ($Only.Count -eq 0) -or ($Only -contains $Section) -or ($Section -eq 'sni' -and $Only -contains 'youtube')}
function Section([string]$Title){Say '';Say ('=== '+$Title+' ===')}
function Scrub([string]$Text){
 $Text=[regex]::Replace($Text,'-----BEGIN [A-Z ]+-----[\s\S]*?-----END [A-Z ]+-----','[PEM REMOVED]')
 [regex]::Replace($Text,'(?i)("?(token|secret|private[_ ]?key|seed|password|authorization)"?\s*[:=]\s*)"?[^",\s}]+"?','$1"[REDACTED]"')
}
function Json($Value){Scrub ($Value|ConvertTo-Json -Depth 12)}
# NB: parameter is not called $Name: script blocks run in the caller scope and
# loops below use their own variable names, never shadowing this one.
function Step([string]$StepName,[scriptblock]$Body){try{& $Body}catch{Say ("[$StepName] ERROR: "+(Scrub $_.Exception.Message))}}
function Prop($Object,[string]$PropName,$Default=$null){
 if($null -ne $Object -and $Object.PSObject.Properties.Name -contains $PropName){return $Object.$PropName};return $Default
}

# Reads a file another process may still hold open (a Chrome child that inherited the
# redirected handle): shared read access, short retries.
function Read-SharedText([string]$Path){
 if(-not (Test-Path -LiteralPath $Path)){return ''}
 for($try=0;$try -lt 20;$try++){
  try{
   $fs=[IO.File]::Open($Path,[IO.FileMode]::Open,[IO.FileAccess]::Read,([IO.FileShare]::ReadWrite -bor [IO.FileShare]::Delete))
   try{$sr=New-Object IO.StreamReader($fs,[Text.Encoding]::UTF8);return $sr.ReadToEnd()}finally{$fs.Dispose()}
  }catch [IO.IOException]{Start-Sleep -Milliseconds 250}
 }
 return ''
}
# Ends the browser we started: the tree first, then any leftover process whose command
# line carries our one-off profile directory (children detach when the parent exits).
function Stop-OwnBrowser($Process,[string]$Marker){
 try{if($Process -and -not $Process.HasExited){& taskkill.exe /T /F /PID $Process.Id 2>&1|Out-Null}}catch{}
 try{
  Get-CimInstance Win32_Process -Filter "Name='chrome.exe' OR Name='msedge.exe'" -ErrorAction SilentlyContinue|
   Where-Object{$_.CommandLine -and $_.CommandLine.Contains($Marker)}|
   ForEach-Object{Stop-Process -Id $_.ProcessId -Force -ErrorAction SilentlyContinue}
 }catch{}
}

# TCP connect + TLS handshake to one IP with a chosen SNI. Certificate validation stays
# ON: the callback returns the real verdict. It only records whether the server answered
# the ClientHello at all, which is what separates "SNI/data blocked" (timeout, reset) from
# "reachable" (server certificate received, even if it is for another name).
function Test-TlsSni([string]$Address,[string]$Sni,[int]$Port=443,[int]$TimeoutMs=5000){
 $result=[ordered]@{Ip=$Address;Sni=$Sni;Tcp='';TcpMs=-1;Tls='';Detail=''}
 $client=New-Object Net.Sockets.TcpClient
 try{
  $watch=[Diagnostics.Stopwatch]::StartNew()
  $task=$client.ConnectAsync($Address,$Port)
  $connected=$false;try{$connected=$task.Wait($TimeoutMs) -and $client.Connected}catch{}
  $result.TcpMs=[int]$watch.ElapsedMilliseconds
  if(-not $connected){$result.Tcp='FAILED';$result.Tls='not attempted';return [pscustomobject]$result}
  $result.Tcp='OK'
  $client.ReceiveTimeout=$TimeoutMs;$client.SendTimeout=$TimeoutMs
  $global:SniProbe=@{Got=$false;Errors=''}
  $cb=[Net.Security.RemoteCertificateValidationCallback]{param($sender,$certificate,$chain,$errors) $global:SniProbe.Got=$true;$global:SniProbe.Errors=[string]$errors;return ($errors -eq [Net.Security.SslPolicyErrors]::None)}
  $ssl=New-Object Net.Security.SslStream($client.GetStream(),$false,$cb)
  $hs=[Diagnostics.Stopwatch]::StartNew()
  try{
   $ssl.AuthenticateAsClient($Sni,$null,[Security.Authentication.SslProtocols]::Tls12,$false)
   $result.Tls="OK in $($hs.ElapsedMilliseconds) ms (certificate verified)"
  }catch{
   $inner=$_.Exception.GetBaseException()
   if($global:SniProbe.Got){$result.Tls="SERVER ANSWERED in $($hs.ElapsedMilliseconds) ms (certificate not accepted for this name: $($global:SniProbe.Errors))"}
   elseif($inner -is [Net.Sockets.SocketException] -and $inner.SocketErrorCode -eq 'ConnectionReset'){$result.Tls="RESET after ClientHello ($($hs.ElapsedMilliseconds) ms)"}
   elseif($inner -is [IO.IOException] -and $inner.Message -match 'timed out|timeout|expired'){$result.Tls="NO ANSWER (timeout $($hs.ElapsedMilliseconds) ms)"}
   elseif($inner -is [Net.Sockets.SocketException] -and $inner.SocketErrorCode -eq 'TimedOut'){$result.Tls="NO ANSWER (timeout $($hs.ElapsedMilliseconds) ms)"}
   else{$result.Tls="FAILED ($($inner.GetType().Name): $($inner.Message -replace '\s+',' '))"}
  }
  $ssl.Dispose()
 }catch{$result.Tls="ERROR $($_.Exception.Message)"}
 finally{$client.Close()}
 return [pscustomobject]$result
}

# ---------- IPC (same wire format as the GUI: 4-byte big-endian length + JSON) ----------
function Read-Exact($Stream,[int]$Count,[DateTime]$Deadline){
 $bytes=New-Object byte[] $Count;$offset=0
 while($offset -lt $Count){
  $left=[int][Math]::Max(1,($Deadline-[DateTime]::UtcNow).TotalMilliseconds)
  $pending=$Stream.BeginRead($bytes,$offset,$Count-$offset,$null,$null)
  if(-not $pending.AsyncWaitHandle.WaitOne($left)){throw 'IPC timeout'}
  $n=$Stream.EndRead($pending);if($n -eq 0){throw 'IPC stream closed'};$offset+=$n
 }
 return ,$bytes
}
function Invoke-IPCOnce([string]$Method,$Payload=$null){
 $token=[IO.File]::ReadAllText((Join-Path $InstallDirectory 'ipc.token')).Trim()   # authenticates only; never printed
 $id=[Guid]::NewGuid().ToString('N')
 $request=@{version=1;id=$id;sent_at=[DateTimeOffset]::UtcNow.ToString('o');token=$token;method=$Method;payload=$Payload}|ConvertTo-Json -Depth 8 -Compress
 $bytes=[Text.Encoding]::UTF8.GetBytes($request)
 $pipe=New-Object IO.Pipes.NamedPipeClientStream('.','neexitlag-v1',[IO.Pipes.PipeDirection]::InOut,[IO.Pipes.PipeOptions]::Asynchronous)
 try{
  $pipe.Connect(3000)
  $deadline=[DateTime]::UtcNow.AddSeconds(8)
  $header=[BitConverter]::GetBytes([uint32]$bytes.Length);[Array]::Reverse($header)
  foreach($part in @($header,$bytes)){$w=$pipe.BeginWrite($part,0,$part.Length,$null,$null);if(-not $w.AsyncWaitHandle.WaitOne(8000)){throw 'IPC write timeout'};$pipe.EndWrite($w)}
  $lenBytes=Read-Exact $pipe 4 $deadline;[Array]::Reverse($lenBytes);$length=[BitConverter]::ToUInt32($lenBytes,0)
  if($length -lt 1 -or $length -gt 1048576){throw 'invalid IPC response size'}
  $response=[Text.Encoding]::UTF8.GetString((Read-Exact $pipe ([int]$length) $deadline))|ConvertFrom-Json
  if(-not $response.ok){throw ([string]$response.error)}
  return $response.data
 }finally{$pipe.Dispose()}
}
function Invoke-IPC([string]$Method,$Payload=$null){
 $until=[DateTime]::UtcNow.AddSeconds(30)
 while($true){
  try{return Invoke-IPCOnce $Method $Payload}
  catch{if("$_" -notlike '*service is starting*' -or [DateTime]::UtcNow -ge $until){throw};Start-Sleep -Milliseconds 500}
 }
}

# ---------- network helpers ----------
# QUIC long header with a reserved version: a QUIC server answers with Version
# Negotiation, but only when the datagram is at least 1200 bytes (RFC 9000 14.1).
function New-QuicProbe([int]$Size){
 $rng=[Security.Cryptography.RandomNumberGenerator]::Create()
 $dcid=New-Object byte[] 8;$scid=New-Object byte[] 8;$rng.GetBytes($dcid);$rng.GetBytes($scid)
 $packet=New-Object byte[] $Size
 $packet[0]=0xC0;$packet[1]=0x1A;$packet[2]=0x2A;$packet[3]=0x3A;$packet[4]=0x4A
 $packet[5]=8;[Array]::Copy($dcid,0,$packet,6,8);$packet[14]=8;[Array]::Copy($scid,0,$packet,15,8)
 return ,$packet
}
function Test-UdpLoss([string]$Address,[int]$Port,[int]$Size,[int]$Count=50,[int]$GapMs=20){
 $packet=New-QuicProbe $Size
 $udp=New-Object Net.Sockets.UdpClient
 $replies=0;$rtts=New-Object Collections.Generic.List[double]
 try{
  $udp.Connect($Address,$Port)
  $watch=[Diagnostics.Stopwatch]::StartNew()
  $drain={
   while($udp.Available -gt 0){
    $remote=New-Object Net.IPEndPoint([Net.IPAddress]::Any,0)
    [void]$udp.Receive([ref]$remote);$script:replies++
   }
  }
  for($i=0;$i -lt $Count;$i++){
   [void]$udp.Send($packet,$packet.Length)
   Start-Sleep -Milliseconds $GapMs
   $replies+=0
   while($udp.Available -gt 0){$remote=New-Object Net.IPEndPoint([Net.IPAddress]::Any,0);[void]$udp.Receive([ref]$remote);$replies++}
  }
  $end=[DateTime]::UtcNow.AddMilliseconds(1500)
  while([DateTime]::UtcNow -lt $end){
   if($udp.Available -gt 0){$remote=New-Object Net.IPEndPoint([Net.IPAddress]::Any,0);[void]$udp.Receive([ref]$remote);$replies++}else{Start-Sleep -Milliseconds 20}
  }
 }finally{$udp.Close()}
 return [pscustomobject]@{Size=$Size;Sent=$Count;Replies=$replies;LossPercent=[Math]::Round(100.0*(1-[Math]::Min($replies,$Count)/$Count),1)}
}
function Test-IcmpLoss([string]$Address,[int]$Size,[int]$Count=50){
 $buffer=New-Object byte[] $Size
 $pings=@();$tasks=@()
 for($i=0;$i -lt $Count;$i++){
  $p=New-Object Net.NetworkInformation.Ping;$pings+=$p
  $tasks+=$p.SendPingAsync($Address,1500,$buffer)
  Start-Sleep -Milliseconds 20
 }
 try{[void][Threading.Tasks.Task]::WaitAll([Threading.Tasks.Task[]]$tasks,10000)}catch{}
 $ok=@($tasks|Where-Object{$_.Status -eq 'RanToCompletion' -and $_.Result.Status -eq 'Success'})
 $rtt=@($ok|ForEach-Object{$_.Result.RoundtripTime})
 foreach($p in $pings){$p.Dispose()}
 return [pscustomobject]@{Size=$Size;Sent=$Count;Replies=$ok.Count;LossPercent=[Math]::Round(100.0*(1-$ok.Count/$Count),1);AvgMs=$(if($rtt.Count){[Math]::Round(($rtt|Measure-Object -Average).Average,1)}else{$null})}
}
function Find-PathMtu([string]$Address){
 # Largest ICMP payload that passes with DF set; MTU = payload + 28.
 $ping=New-Object Net.NetworkInformation.Ping
 $options=New-Object Net.NetworkInformation.PingOptions(64,$true)
 $try={param($n) $r=$ping.Send($Address,1500,(New-Object byte[] $n),$options);return $r.Status}
 try{
  if((& $try 32) -ne 'Success'){return [pscustomobject]@{Reachable=$false;Detail='no ICMP reply even for a 32-byte DF probe (ICMP filtered or host down)'}}
  $lo=32;$hi=1472
  if((& $try $hi) -eq 'Success'){$lo=$hi}
  else{while($hi-$lo -gt 1){$mid=[int](($lo+$hi)/2);if((& $try $mid) -eq 'Success'){$lo=$mid}else{$hi=$mid}}}
  return [pscustomobject]@{Reachable=$true;MaxPayload=$lo;PathMtu=$lo+28;Detail=''}
 }finally{$ping.Dispose()}
}

$ScriptVersion='2026-10-19.1 (Edge fresh nonce + guarded rollback, googlevideo 8/2s, DNS recovery)'
Say "VELIQ diag-no-vpn, started $((Get-Date).ToString('yyyy-MM-dd HH:mm:ss zzz'))"
Say "Script version: $ScriptVersion"
Say "Script file: $(if($PSCommandPath){$PSCommandPath}else{'(unknown)'})"
Step 'repo' {
 $repo=Split-Path (Split-Path $scriptDir -Parent) -Parent
 if((Test-Path (Join-Path $repo '.git')) -and (Get-Command git -ErrorAction SilentlyContinue)){
  $head=(& git -C $repo rev-parse --short=12 HEAD 2>&1|Out-String).Trim()
  $branch=(& git -C $repo branch --show-current 2>&1|Out-String).Trim()
  $dirty=(& git -C $repo status --porcelain 2>&1|Out-String).Trim().Length -gt 0
  Say "Repository: $repo branch $branch commit $head$(if($dirty){' (+local changes)'})  <- compare with the commit you expect"
 }else{Say "Repository: not a git checkout here ($repo): this script file may be a stray copy"}
}
Say "Result file: $result"

if(Want 'env'){
Section '0. Environment (is the VPN really off?)'
Step 'env' {
 $os=Get-CimInstance Win32_OperatingSystem
 Say "OS: $($os.Caption) build $($os.BuildNumber); user: $env:USERNAME; PowerShell $($PSVersionTable.PSVersion)"
 $vpn=@(Get-NetAdapter|Where-Object{$_.Status -eq 'Up' -and ($_.InterfaceDescription+' '+$_.Name) -match 'VPN|Happ|TUN|TAP|WireGuard|Wintun'}|ForEach-Object{$_.Name})
 Say ("Active VPN-like adapters: "+$(if($vpn.Count){($vpn -join ', ')+'  <-- VPN looks ON, results below are NOT a no-VPN test'}else{'none (good)'}))
 Say ("Default route(s): "+((Get-NetRoute -DestinationPrefix '0.0.0.0/0' -ErrorAction SilentlyContinue|Sort-Object RouteMetric|ForEach-Object{"if$($_.InterfaceIndex) via $($_.NextHop) metric $($_.RouteMetric)"}) -join '; '))
 Say ("System DNS: "+((Get-DnsClientServerAddress -AddressFamily IPv4|Where-Object{$_.ServerAddresses}|ForEach-Object{$_.ServerAddresses -join ','}|Select-Object -Unique) -join '; '))
}}

if(Want 'service'){
Section '1a. Service and binary'
Step 'service' {
 $svc=Get-Service neexitlag -ErrorAction Stop
 Say "Service neexitlag: $($svc.Status), start type $($svc.StartType)"
 $hashes=@{}
 foreach($pair in @(@('installed',(Join-Path $InstallDirectory 'neexitlag-service.exe')),@('built',(Join-Path $BuildDirectory 'neexitlag-service.exe')))){
  if(Test-Path $pair[1]){$h=(Get-FileHash $pair[1] -Algorithm SHA256).Hash;$hashes[$pair[0]]=$h;Say "SHA256 $($pair[0]): $h (modified $((Get-Item $pair[1]).LastWriteTime.ToString('yyyy-MM-dd HH:mm')))"}
  else{Say "SHA256 $($pair[0]): file not found ($($pair[1]))"}
 }
 if($hashes.ContainsKey('installed') -and $hashes.ContainsKey('built')){Say ("installed == built: "+($hashes['installed'] -eq $hashes['built']))}
}}

$script:status=$null
$script:browserJson=''
if(Want 'ipc'){
Section '1b. IPC status'
Step 'status' {
 $script:status=Invoke-IPC 'status'
 Say ("optimizing: "+(Prop $script:status 'optimizing')+"; service build: "+(Prop $script:status 'build' '(none: the running service is OLDER than the build stamp)'))
 Say (Json $script:status)
}
Section '1c. IPC profiles (what Optimize covers)'
Step 'profiles' {
 $rows=@(Invoke-IPC 'profiles');Say "Profile count: $($rows.Count)"
 foreach($p in $rows){Say ("- id=$(Prop $p 'id') name=$(Prop $p 'name') mode=$(Prop $p 'mode') custom=$(Prop $p 'custom') fallback=$(Prop $p 'fallback') preset=$(Prop $p 'app_preset') exe=$((@(Prop $p 'executables' @())) -join ' | ') rules=$(@(Prop $p 'rules' @()).Count)")}
 $exes=($rows|ForEach-Object{@(Prop $_ 'executables' @())})-join ' '
 Say ("Has Chrome profile: "+($exes -match 'chrome\.exe')+"; Edge: "+($exes -match 'msedge\.exe')+"; Telegram: "+($exes -match '(?i)telegram\.exe')+"; Discord: "+($exes -match '(?i)discord'))
}
Section '1d. IPC browser.status'
Step 'browser' {$script:browserJson=Json (Invoke-IPC 'browser.status');Say $script:browserJson}
Section '1e. IPC capture.diagnostics'
Step 'capture' {Say (Json (Invoke-IPC 'capture.diagnostics'))}
Section '1f. IPC quality.diagnostics'
Step 'quality' {Say (Json (Invoke-IPC 'quality.diagnostics'))}
Section '1g. IPC relay.pool.diagnostics (handshake/x509/dial errors show up here)'
Step 'pool' {
 $ids=@();if($script:status){$ids=@(Prop $script:status 'relays' @()|ForEach-Object{Prop $_ 'id'}|Where-Object{$_})}
 if($ids.Count -eq 0){$ids=@('VPS-01')}
 foreach($relayId in $ids){Say ("relay $relayId : "+(Json (Invoke-IPC 'relay.pool.diagnostics' @{relay_id=$relayId})))}
}}

if(Want 'tcp'){
Section "2a. $VpsHost TCP $ControlPlanePort (control plane) - INFORMATIONAL"
Say 'Informational only. The control plane requires TLS 1.3. Windows PowerShell 5.1 (.NET Framework) cannot speak TLS 1.3, so this script does not handshake itself; curl.exe is used with the project CA and normal verification. A failure here is a limit of this probe, NOT a service fault: the service status (IPC, sections 1b/1g) is authoritative for the control plane and relay.'
Step 'tcp' {
 $watch=[Diagnostics.Stopwatch]::StartNew()
 $client=New-Object Net.Sockets.TcpClient
 try{
  $task=$client.ConnectAsync($VpsHost,$ControlPlanePort)
  $ok=$false;try{$ok=$task.Wait(6000) -and $client.Connected}catch{Say ("TCP connect error: "+$_.Exception.GetBaseException().Message)}
  if($ok){Say "TCP ${ControlPlanePort}: connected in $($watch.ElapsedMilliseconds) ms"}
  else{Say "TCP ${ControlPlanePort}: NO CONNECTION within 6 s (blocked, dropped or nothing listening)";return}
 }finally{$client.Close()}
 $caPath=Join-Path $InstallDirectory 'cp-ca.pem'
 if(-not (Test-Path $caPath)){
  $manifest=Join-Path $InstallDirectory 'service.json'
  if(Test-Path $manifest){try{$m=Get-Content $manifest -Raw|ConvertFrom-Json;$caPath=[string](Prop $m 'control_plane_ca_file' '')}catch{}}
 }
 $curlExe=(Get-Command curl.exe -ErrorAction SilentlyContinue)
 if(-not $curlExe){Say 'curl.exe not found: certificate check skipped (informational).';return}
 if(-not $caPath -or -not (Test-Path $caPath)){Say "Project CA file not readable ($caPath): certificate check skipped (informational). Verification is never disabled."; return}
 # Normal verification against the project CA. --ssl-no-revoke only skips the
 # revocation-list download that a private CA cannot answer; it does not weaken the chain,
 # validity or host checks.
 $raw=(& $curlExe.Source -sS -m 10 --ssl-no-revoke --cacert $caPath -o NUL -w 'HTTP %{http_code} tls_verify=%{ssl_verify_result} connect=%{time_connect}s tls=%{time_appconnect}s' "https://${VpsHost}:${ControlPlanePort}/healthz" 2>&1|Out-String)
 $curlExit=$LASTEXITCODE   # must be read right after the native call
 $text=($raw.Trim() -replace '\s+',' ')
 $tlsSeconds=0.0
 $tlsMatch=[regex]::Match($text,'tls=([0-9]+(?:[.,][0-9]+)?)s')
 if($tlsMatch.Success){[void][double]::TryParse(($tlsMatch.Groups[1].Value -replace ',','.'),[Globalization.NumberStyles]::Float,[Globalization.CultureInfo]::InvariantCulture,[ref]$tlsSeconds)}
 # OK only when curl itself succeeded, a TLS handshake really happened (time > 0) and the
 # certificate verified against the project CA. Anything else, including a curl/Schannel
 # error such as SEC_E_ILLEGAL_MESSAGE with tls=0, is "undetermined", never a pass.
 $verdict=if($curlExit -eq 0 -and $tlsSeconds -gt 0 -and $text -match 'tls_verify=0'){'handshake and certificate OK against the project CA'}else{"UNDETERMINED (curl exit code $curlExit, tls time ${tlsSeconds}s): the handshake did not complete here; this probe cannot tell a service fault from a client TLS limit"}
 Say "curl with project CA: $text -> $verdict"
}}

if(Want 'udp'){
Section "2b. $VpsHost UDP ${QuicPort}: loss by datagram size (QUIC Version Negotiation probes, 50 each)"
Step 'udp' {
 Say 'Only datagrams of >=1200 bytes make a QUIC server answer; smaller sizes cannot be measured on this port (server ignores them by design). Echo port 40000 needs a secret token and a fixed 128-byte format, so it is NOT probed here.'
 foreach($size in 100,600,1200,1400){
  if($size -lt 1200){Say "UDP ${QuicPort} size ${size}: not measurable (QUIC servers ignore datagrams <1200 bytes)";continue}
  $r=Test-UdpLoss $VpsHost $QuicPort $size 50 20
  Say "UDP ${QuicPort} size ${size}: sent $($r.Sent), replies $($r.Replies), loss $($r.LossPercent)%"
 }
}}

if(Want 'mtu'){
Section "2c. $VpsHost ICMP loss by size, path MTU, route"
Step 'icmp' {
 foreach($size in 100,600,1200,1400){
  $r=Test-IcmpLoss $VpsHost $size 50
  Say "ICMP payload ${size}: sent $($r.Sent), replies $($r.Replies), loss $($r.LossPercent)%, avg rtt $($r.AvgMs) ms"
 }
}
Step 'mtu' {
 $m=Find-PathMtu $VpsHost
 if($m.Reachable){Say "Path MTU (DF set): max payload $($m.MaxPayload) bytes -> MTU $($m.PathMtu)"}else{Say "Path MTU: $($m.Detail)"}
}
Step 'tracert' {
 if($isWin){
  Say 'tracert -d -h 15 -w 1000:'
  foreach($line in (& tracert.exe -d -h 15 -w 1000 $VpsHost 2>&1)){$t="$line".TrimEnd();if($t){Say "  $t"}}
 }else{Say 'tracert skipped (not Windows)'}
}}

if(Want 'dns'){
Section 'DNS: system resolver vs 8.8.8.8'
$dnsHosts=@('claude.ai','api.anthropic.com','chatgpt.com','www.youtube.com','web.telegram.org','discord.com','t.me')
foreach($dnsHost in $dnsHosts){
 Step "dns $dnsHost" {
  $sys='';$pub=''
  try{$sys=(@(Resolve-DnsName $dnsHost -Type A -DnsOnly -ErrorAction Stop|Where-Object{$_.Type -eq 'A'}|ForEach-Object{$_.IPAddress})|Sort-Object) -join ','}catch{$sys='ERROR '+($_.Exception.Message -replace '\s+',' ')}
  try{$pub=(@(Resolve-DnsName $dnsHost -Type A -DnsOnly -Server 8.8.8.8 -ErrorAction Stop|Where-Object{$_.Type -eq 'A'}|ForEach-Object{$_.IPAddress})|Sort-Object) -join ','}catch{$pub='ERROR '+($_.Exception.Message -replace '\s+',' ')}
  $sa=@($sys -split ',');$pa=@($pub -split ',')
  $verdict=if($sys -like 'ERROR*' -or $pub -like 'ERROR*'){'a lookup failed'}elseif(@($sa|Where-Object{$pa -contains $_}).Count -gt 0){'overlap (consistent)'}else{'NO OVERLAP (can be CDN rotation or DNS tampering)'}
  Say "$dnsHost | system: $sys | 8.8.8.8: $pub | $verdict"
 }
}}

if(Want 'sites'){
Section 'HTTPS to the target sites (plain curl.exe, not through Chrome/PAC)'
$curl=(Get-Command curl.exe -ErrorAction SilentlyContinue)
foreach($site in @('chatgpt.com','claude.ai','www.youtube.com','web.telegram.org','discord.com')){
 Step "site $site" {
  if($curl){
   $text=(& $curl.Source -sS -m 15 -o NUL -w 'HTTP %{http_code} ip=%{remote_ip} connect=%{time_connect}s tls=%{time_appconnect}s total=%{time_total}s' "https://$site/" 2>&1|Out-String).Trim() -replace '\s+',' '
   $verdict=if($text -match 'HTTP (2|3)\d\d'){'OPENS'}elseif($text -match 'HTTP (4|5)\d\d'){'ANSWERS (HTTP error, host reachable)'}else{'FAILS'}
   Say "$site [$verdict] $text"
  }else{
   try{$r=Invoke-WebRequest "https://$site/" -UseBasicParsing -TimeoutSec 15 -MaximumRedirection 0 -ErrorAction Stop;Say "$site [OPENS] HTTP $([int]$r.StatusCode)"}
   catch{$resp=$_.Exception.Response;if($resp){Say "$site [ANSWERS] HTTP $([int]$resp.StatusCode)"}else{Say "$site [FAILS] $($_.Exception.Message)"}}
  }
 }
}}

if(Want 'sni'){
Section 'YouTube: TCP + TLS per IP with different SNI (is it SNI blocking?)'
Step 'sni' {
 $addrs=@()
 foreach($srv in @($null,'8.8.8.8')){
  try{
   $q=@{Name='www.youtube.com';Type='A';DnsOnly=$true;ErrorAction='Stop'};if($srv){$q.Server=$srv}
   $addrs+=@(Resolve-DnsName @q|Where-Object{$_.Type -eq 'A'}|ForEach-Object{$_.IPAddress})
  }catch{Say ("DNS www.youtube.com via $(if($srv){$srv}else{'system'}): "+($_.Exception.Message -replace '\s+',' '))}
 }
 $addrs=@($addrs|Select-Object -Unique|Select-Object -First 8)
 if($addrs.Count -eq 0){Say 'No A records for www.youtube.com, SNI test skipped.';return}
 Say ("Addresses tested: "+($addrs -join ', '))
 $short={param($x)
  if($x.Tcp -ne 'OK'){return 'TCP-FAIL'}
  if($x.Tls -like 'OK*'){return 'OK'}
  if($x.Tls -like 'SERVER ANSWERED*'){return 'ANSWERED'}
  if($x.Tls -like 'RESET*'){return 'RESET'}
  if($x.Tls -like 'NO ANSWER*'){return 'NO-ANSWER'}
  return 'ERROR'}
 $answered={param($c) $c -eq 'OK' -or $c -eq 'ANSWERED'}
 $rows=@()
 foreach($ip in $addrs){
  $r=@{}
  foreach($sni in 'www.youtube.com','www.google.com','example.com'){
   $r[$sni]=Test-TlsSni $ip $sni 443 5000
   Say ("  $ip SNI=$sni : tcp $($r[$sni].Tcp) ($($r[$sni].TcpMs) ms); tls $($r[$sni].Tls)")
  }
  $yt=& $short $r['www.youtube.com'];$gg=& $short $r['www.google.com'];$ex=& $short $r['example.com']
  $verdict=if($yt -eq 'TCP-FAIL'){'TCP blocked at IP/port level (not an SNI verdict)'}
           elseif(& $answered $yt){'server answers for www.youtube.com: no block at this layer'}
           elseif((& $answered $gg) -or (& $answered $ex)){'SNI BLOCKING LIKELY: another SNI is answered, youtube is not'}
           else{'data after the handshake is dropped for this IP (IP-level), not SNI-specific'}
  $rows+=[pscustomobject]@{IP=$ip;TCP=$r['www.youtube.com'].Tcp;'SNI youtube'=$yt;'SNI google (control)'=$gg;'SNI example (foreign)'=$ex;Verdict=$verdict}
 }
 Say ''
 Say 'Verdict table (OK/ANSWERED = server replied to the ClientHello; NO-ANSWER = silence; RESET = connection reset):'
 foreach($line in ($rows|Format-Table -AutoSize -Wrap|Out-String -Width 220).Split("`n")){if($line.Trim()){Say ("  "+$line.TrimEnd())}}
}}

if($Only -contains 'youtube'){
Section 'YouTube: Chrome + Edge, address/domain cache and actual relay payload bytes per minute'
Step 'youtube' {
 $st=Invoke-IPC 'status'
 Say ("service build: "+(Prop $st 'build' '-')+"; optimizing: "+(Prop $st 'optimizing')+"; youtube_via_relay="+(Prop (Prop $st 'settings') 'youtube_via_relay' $false)+" (relay-first preference; fallback no longer needs this flag)")
 $allRecords=@();$shownUsage=$false
 foreach($lane in @('browser','edge')){
  try{$snapshot=Invoke-IPC ($lane+'.status')}catch{Say ($lane+': IPC unavailable: '+(Scrub $_.Exception.Message));continue}
  Say ($lane+": listening="+(Prop $snapshot 'listening')+" policy_ready="+(Prop $snapshot 'policy_ready')+" ready_profiles="+(@(Prop $snapshot 'ready_profiles' @()) -join ',')+" rollback_guard_active="+(Prop $snapshot 'rollback_guard_active' $false)+" readiness_state="+(Prop $snapshot 'readiness_state' '')+" pending_profile="+(Prop $snapshot 'pending_profile' '')+" last_error="+(Prop $snapshot 'last_error' '')+" truncated="+(Prop $snapshot 'truncated')+" lost="+(Prop $snapshot 'lost'))
  $cache=Prop $snapshot 'direct_timeout_cache'
  if($cache){
   Say ("cache (shared by both lanes): domains="+(Prop $cache 'active_origins')+" addresses="+(Prop $cache 'active_addresses' '-')+" stalls="+(Prop $cache 'direct_stalls_total')+" rechecks="+(Prop $cache 'background_direct_probes')+" recovered_verified="+(Prop $cache 'recovered_to_direct'))
   foreach($origin in @(Prop $cache 'top_origins' @())){
    Say (" cache: network="+(Prop $origin 'network' '-')+" profile="+(Prop $origin 'profile_id' '-')+" domain="+(Prop $origin 'host')+" port="+(Prop $origin 'port')+" address="+(Prop $origin 'address' '(domain)')+" reason="+(Prop $origin 'reason' '-')+" distinct="+(Prop $origin 'distinct_failed_addresses' '-')+" ttl_ms="+(Prop $origin 'ttl_remaining_ms' '-')+" next_probe="+(Prop $origin 'next_direct_probe'))
   }
  }
  $usage=Prop $snapshot 'relay_usage'
  if($usage -and -not $shownUsage){
   $shownUsage=$true
   Say ("relay usage (shared, TCP payload excluding QUIC overhead/probes): googlevideo="+(Prop $usage 'googlevideo_active_tunnels')+"/"+(Prop $usage 'googlevideo_tunnel_limit')+" waiters="+(Prop $usage 'googlevideo_waiters' 0)+" admission_failures="+(Prop $usage 'googlevideo_admission_failures' 0)+" cap_events="+(Prop $usage 'googlevideo_limit_events')+" truncated="+(Prop $usage 'truncated')+" overflow_bytes="+(Prop $usage 'cardinality_overflow_bytes'))
   foreach($minute in @(Prop $usage 'minutes' @())){Say (" minute="+(Prop $minute 'minute_utc')+" profile="+(Prop $minute 'profile_id')+" domain="+(Prop $minute 'domain')+" tx="+(Prop $minute 'sent_bytes')+" rx="+(Prop $minute 'received_bytes'))}
  }
  $failures=Prop $snapshot 'failures'
  if($failures){foreach($property in $failures.PSObject.Properties){if($property.Name -match 'DIRECT_|UPSTREAM_|MIDSTREAM'){Say (" $lane failure "+$property.Name+'='+$property.Value)}}}
  foreach($record in @(Prop $snapshot 'records' @())){
   $allRecords+= $record
   if((Prop $record 'domain' '') -notmatch '(^|\.)(youtube\.com|googlevideo\.com|ytimg\.com|ggpht\.com|youtu\.be|youtube-nocookie\.com)$'){continue}
   $plan=Prop $record 'plan';$policy=Prop $plan 'policy';$owner=Prop $record 'owner'
   Say (" $lane #"+(Prop $record 'id')+" process="+(Prop $owner 'exe')+" pid="+(Prop $owner 'pid')+" profile="+(Prop $owner 'profile_id')+" "+(Prop $record 'domain')+":"+(Prop $record 'port')+" -> "+(Prop $plan 'destination')+" action="+(Prop $policy 'action')+" source="+(Prop $policy 'source')+" probe="+(Prop $plan 'probe')+" attempts="+(Prop $plan 'direct_attempts' '-')+" result="+(Prop $record 'result')+" open_failure="+(Prop $record 'open_failure')+" tx="+(Prop $record 'sent_bytes')+" rx="+(Prop $record 'received_bytes'))
  }
 }
 $videoName=@($allRecords|ForEach-Object{Prop $_ 'domain' ''}|Where-Object{$_ -match '(^|\.)googlevideo\.com$'}|Select-Object -First 1)
 if($videoName.Count){$videoOrigin=[string]$videoName[0]}else{$videoOrigin='redirector.googlevideo.com';Say 'No real video CDN hostname in retained evidence; using redirector only as a DNS/TLS control, not as video PASS.'}
 Say ("googlevideo address checks: "+$videoOrigin)
 $videoAddresses=@()
 foreach($dnsServer in @($null,'8.8.8.8')){
  try{
   $query=@{Name=$videoOrigin;Type='A';DnsOnly=$true;ErrorAction='Stop'};if($dnsServer){$query.Server=$dnsServer}
   $found=@(Resolve-DnsName @query|Where-Object{$_.Type -eq 'A'}|ForEach-Object{$_.IPAddress}|Select-Object -Unique)
   Say (" DNS "+$(if($dnsServer){$dnsServer}else{'system'})+': '+($found -join ','));$videoAddresses+=$found
  }catch{Say (' DNS failed: '+(Scrub $_.Exception.Message))}
 }
 foreach($ipAddress in @($videoAddresses|Select-Object -Unique|Select-Object -First 8)){
  foreach($serverName in @($videoOrigin,'www.youtube.com','www.google.com','example.com')){
   $probe=Test-TlsSni $ipAddress $serverName 443 3000
   Say (" $ipAddress SNI=$serverName TCP="+$probe.Tcp+" tcp_ms="+$probe.TcpMs+" TLS="+$probe.Tls)
  }
 }
 Say 'These DNS/TCP/TLS controls do not establish video playback, HTTP/3, site usability or an Internet PASS.'
}}

if(Want 'browser'){
Section 'End-to-end through the browser path (the way users really browse)'
Step 'connect' {
 # Raw CONNECT from PowerShell. The proxy authorizes only Chrome-profile processes, so
 # a rejection (403) here is EXPECTED and only proves the proxy is alive. The real
 # end-to-end verdict is the headless-browser test below.
 $json=$script:browserJson
 if(-not $ProxyAddress){
  if(-not $json -and $isWin){try{$json=Json (Invoke-IPC 'browser.status')}catch{}}
  $m=[regex]::Match([string]$json,'"address"\s*:\s*"(127\.0\.0\.1:\d+)"')
  if($m.Success){$ProxyAddress=$m.Groups[1].Value}
 }
 if(-not $ProxyAddress){Say 'Local browser proxy address not found in browser.status (proxy not listening?). Raw CONNECT test skipped.';return}
 $parts=$ProxyAddress -split ':'
 Say "Local browser proxy: $ProxyAddress"
 foreach($target in 'claude.ai','chatgpt.com','www.youtube.com','web.telegram.org','discord.com'){
  $client=New-Object Net.Sockets.TcpClient
  try{
   $w=[Diagnostics.Stopwatch]::StartNew()
   $client.Connect($parts[0],[int]$parts[1]);$client.ReceiveTimeout=12000
   $stream=$client.GetStream()
   $req=[Text.Encoding]::ASCII.GetBytes("CONNECT ${target}:443 HTTP/1.1`r`nHost: ${target}:443`r`n`r`n")
   $stream.Write($req,0,$req.Length)
   $buf=New-Object byte[] 512;$n=$stream.Read($buf,0,512)
   $first=([Text.Encoding]::ASCII.GetString($buf,0,$n) -split "`r?`n")[0]
   $note=if($first -match ' 403'){' (owner check rejects non-browser processes: expected, proves only that the proxy is alive)'}else{''}
   Say "CONNECT ${target}:443 -> '$first' in $($w.ElapsedMilliseconds) ms$note"
  }catch{Say "CONNECT ${target}:443 -> ERROR $($_.Exception.GetBaseException().Message)"}
  finally{$client.Close()}
 }
}
Step 'headless' {
 if(-not $isWin){Say 'Headless browser test skipped (not Windows)';return}
 $browsers=[ordered]@{
  'Chrome'=@("$env:ProgramFiles\Google\Chrome\Application\chrome.exe","${env:ProgramFiles(x86)}\Google\Chrome\Application\chrome.exe","$env:LOCALAPPDATA\Google\Chrome\Application\chrome.exe")
  'Edge'=@("${env:ProgramFiles(x86)}\Microsoft\Edge\Application\msedge.exe","$env:ProgramFiles\Microsoft\Edge\Application\msedge.exe")
 }
 foreach($label in $browsers.Keys){
  $exe=@($browsers[$label]|Where-Object{Test-Path $_})|Select-Object -First 1
  if(-not $exe){Say "$label : not installed";continue}
  Say "$label headless (own empty profile, your windows and tabs are untouched): $exe"
  foreach($site in 'claude.ai','chatgpt.com','www.youtube.com','web.telegram.org','discord.com'){
   $marker='veliq-diag-'+[Guid]::NewGuid().ToString('N')
   $dir=Join-Path ([IO.Path]::GetTempPath()) $marker
   $o="$dir.out";$e="$dir.err";$p=$null
   try{
    New-Item -ItemType Directory -Path $dir -Force|Out-Null
    $w=[Diagnostics.Stopwatch]::StartNew()
    $browserArgs=@('--headless=new','--disable-gpu','--no-first-run','--no-default-browser-check','--timeout=15000','--dump-dom',"--user-data-dir=$dir","https://$site/")
    $p=Start-Process -FilePath $exe -ArgumentList $browserArgs -PassThru -WindowStyle Hidden -RedirectStandardOutput $o -RedirectStandardError $e
    # Wait for the main process, end the whole tree and any detached child that still
    # holds the redirected files, and only then read them (shared access).
    $finished=$p.WaitForExit(30000)
    Stop-OwnBrowser $p $marker
    $dom=Read-SharedText $o
    $errText=Read-SharedText $e
    $title=[regex]::Match($dom,'(?is)<title[^>]*>(.*?)</title>').Groups[1].Value.Trim()
    $err=[regex]::Match($dom+' '+$errText,'ERR_[A-Z_]+').Value
    $verdict=if(-not $finished){'TIMEOUT (30 s)'}elseif($err){"FAILED $err"}elseif($dom.Length -lt 200){'EMPTY PAGE'}else{'DOM RECEIVED (not full-site PASS)'}
    Say "  $site [$verdict] $($w.ElapsedMilliseconds) ms, dom $($dom.Length) bytes, title: '$title'"
   }catch{Say "  $site ERROR $($_.Exception.Message)"}
   finally{
    Stop-OwnBrowser $p $marker
    Remove-Item -LiteralPath $dir -Recurse -Force -ErrorAction SilentlyContinue;Remove-Item -LiteralPath $o,$e -Force -ErrorAction SilentlyContinue
   }
  }
 }
}}

if(Want 'apps'){
Section 'Installed apps and browser coverage'
Step 'apps' {
 $paths=[ordered]@{
  'Chrome'=@("$env:ProgramFiles\Google\Chrome\Application\chrome.exe","${env:ProgramFiles(x86)}\Google\Chrome\Application\chrome.exe","$env:LOCALAPPDATA\Google\Chrome\Application\chrome.exe")
  'Edge'=@("${env:ProgramFiles(x86)}\Microsoft\Edge\Application\msedge.exe","$env:ProgramFiles\Microsoft\Edge\Application\msedge.exe")
  'Firefox'=@("$env:ProgramFiles\Mozilla Firefox\firefox.exe","${env:ProgramFiles(x86)}\Mozilla Firefox\firefox.exe")
  'Telegram Desktop'=@("$env:APPDATA\Telegram Desktop\Telegram.exe","$env:ProgramFiles\Telegram Desktop\Telegram.exe","$env:LOCALAPPDATA\Telegram Desktop\Telegram.exe")
  'Discord'=@("$env:LOCALAPPDATA\Discord\Update.exe")
 }
 foreach($k in $paths.Keys){$f=@($paths[$k]|Where-Object{Test-Path $_});Say ("$k : "+$(if($f.Count){'installed: '+($f -join '; ')}else{'not found'}))}
 Say ("Running now: "+((Get-Process -ErrorAction SilentlyContinue|Where-Object{$_.ProcessName -match '^(chrome|msedge|firefox|telegram|discord)$'}|Select-Object -ExpandProperty ProcessName -Unique) -join ', '))
}}

Say ''
Say "Finished $((Get-Date).ToString('yyyy-MM-dd HH:mm:ss'))."
try{$all=[IO.File]::ReadAllText($result);[IO.File]::WriteAllText($result,(Scrub $all),$utf8Bom)}catch{}
Write-Host "`nSaved: $result"
