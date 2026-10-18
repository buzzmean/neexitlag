#requires -Version 5.1
<#
Non-interactive, read-only diagnostics to run with the VPN OFF, as a normal user:
  powershell -NoProfile -ExecutionPolicy Bypass -File .\tools\windows\diag-no-vpn.ps1
Writes diag-result.txt next to this script (falls back to %TEMP% if that folder is not
writable). Changes nothing on the machine. Reads no token/identity/certificate into the
report; the whole report is scrubbed of secret-shaped values before the run ends.
Self-contained: does not need common.ps1. -Only limits the run to some sections:
  env, service, ipc, tcp, udp, mtu, dns, sites, browser, apps
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
$utf8=New-Object Text.UTF8Encoding($false)
try{[IO.File]::WriteAllText($result,'',$utf8)}catch{$result=Join-Path ([IO.Path]::GetTempPath()) 'diag-result.txt';[IO.File]::WriteAllText($result,'',$utf8)}
$isWin=($PSVersionTable.PSVersion.Major -lt 6) -or $IsWindows

function Say([string]$Text){Write-Host $Text;[IO.File]::AppendAllText($result,$Text+"`r`n",$utf8)}
function Want([string]$Section){return ($Only.Count -eq 0) -or ($Only -contains $Section)}
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

Say "VELIQ diag-no-vpn, started $((Get-Date).ToString('yyyy-MM-dd HH:mm:ss zzz'))"
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
 Say ("optimizing: "+(Prop $script:status 'optimizing'))
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
Section "2a. $VpsHost TCP $ControlPlanePort (control plane, verified against the project CA)"
Step 'tcp' {
 $caPath=Join-Path $InstallDirectory 'cp-ca.pem'
 if(-not (Test-Path $caPath)){
  $manifest=Join-Path $InstallDirectory 'service.json'
  if(Test-Path $manifest){$m=Get-Content $manifest -Raw|ConvertFrom-Json;$caPath=[string](Prop $m 'control_plane_ca_file' '')}
 }
 $ca=$null
 if($caPath -and (Test-Path $caPath)){$ca=New-Object Security.Cryptography.X509Certificates.X509Certificate2($caPath);Say "Project control-plane CA: $($ca.Subject), valid until $($ca.NotAfter.ToString('yyyy-MM-dd'))"}
 else{Say "Project control-plane CA file not readable ($caPath): chain check skipped, only reachability and certificate facts are reported"}
 $client=New-Object Net.Sockets.TcpClient
 try{
  $watch=[Diagnostics.Stopwatch]::StartNew()
  $task=$client.ConnectAsync($VpsHost,$ControlPlanePort)
  $ok=$false;try{$ok=$task.Wait(6000) -and $client.Connected}catch{Say ("TCP connect error: "+$_.Exception.GetBaseException().Message)}
  if(-not $ok){Say "TCP ${ControlPlanePort}: NO CONNECTION within 6 s (blocked, dropped or nothing listening)";return}
  Say "TCP ${ControlPlanePort}: connected in $($watch.ElapsedMilliseconds) ms"
  # The callback performs the REAL check (chain must end in the project CA, no other
  # chain error, host name must match). Its verdict is returned, nothing is bypassed.
  $global:DiagTls=@{Verdict='not evaluated';Subject='';Issuer='';Window='';San=''}
  $cb=[Net.Security.RemoteCertificateValidationCallback]{
   param($sender,$certificate,$chain,$errors)
   $leaf=New-Object Security.Cryptography.X509Certificates.X509Certificate2($certificate)
   $global:DiagTls.Subject=$leaf.Subject;$global:DiagTls.Issuer=$leaf.Issuer
   $global:DiagTls.Window="$($leaf.NotBefore.ToString('yyyy-MM-dd')) .. $($leaf.NotAfter.ToString('yyyy-MM-dd'))"
   $global:DiagTls.San=(($leaf.Extensions|Where-Object{$_.Oid.Value -eq '2.5.29.17'}|ForEach-Object{$_.Format($false)}) -join '')
   if($null -eq $ca){$global:DiagTls.Verdict="not verified (no CA file); system-store flags: $errors";return $false}
   $own=New-Object Security.Cryptography.X509Certificates.X509Chain
   $own.ChainPolicy.RevocationMode='NoCheck'
   $own.ChainPolicy.VerificationFlags='AllowUnknownCertificateAuthority'
   [void]$own.ChainPolicy.ExtraStore.Add($ca)
   [void]$own.Build($leaf)
   $top=$own.ChainElements[$own.ChainElements.Count-1].Certificate.Thumbprint
   $bad=@($own.ChainStatus|Where-Object{$_.Status -ne 'UntrustedRoot'}|ForEach-Object{$_.Status.ToString()})
   $nameBad=([int]$errors -band ([int][Net.Security.SslPolicyErrors]::RemoteCertificateNameMismatch)) -ne 0
   $reasons=@()
   if($top -ne $ca.Thumbprint){$reasons+='chain does not end in the project CA'}
   if($bad.Count){$reasons+=('chain errors: '+($bad -join ','))}
   if($nameBad){$reasons+='host name/IP not in certificate SAN'}
   if($reasons.Count){$global:DiagTls.Verdict='REJECTED: '+($reasons -join '; ');return $false}
   $global:DiagTls.Verdict='ACCEPTED by project CA (chain, validity and SAN ok)';return $true
  }
  $ssl=New-Object Net.Security.SslStream($client.GetStream(),$false,$cb)
  try{$ssl.AuthenticateAsClient($VpsHost);Say "TLS handshake OK ($($ssl.SslProtocol))"}
  catch{Say ("TLS handshake FAILED: "+$_.Exception.GetBaseException().Message)}
  Say "  verdict: $($global:DiagTls.Verdict)"
  Say "  subject $($global:DiagTls.Subject); issuer $($global:DiagTls.Issuer); valid $($global:DiagTls.Window)"
  Say "  SAN: $($global:DiagTls.San)"
  $ssl.Dispose()
 }finally{$client.Close()}
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
   $dir=Join-Path ([IO.Path]::GetTempPath()) ('veliq-diag-'+[Guid]::NewGuid().ToString('N'))
   $o="$dir.out";$e="$dir.err";$p=$null
   try{
    New-Item -ItemType Directory -Path $dir -Force|Out-Null
    $w=[Diagnostics.Stopwatch]::StartNew()
    $args=@('--headless=new','--disable-gpu','--no-first-run','--no-default-browser-check','--timeout=15000','--dump-dom',"--user-data-dir=$dir","https://$site/")
    $p=Start-Process -FilePath $exe -ArgumentList $args -PassThru -WindowStyle Hidden -RedirectStandardOutput $o -RedirectStandardError $e
    $finished=$p.WaitForExit(30000)
    if(-not $finished){& taskkill.exe /T /F /PID $p.Id 2>&1|Out-Null}
    $dom=if(Test-Path $o){[IO.File]::ReadAllText($o)}else{''}
    $title=[regex]::Match($dom,'(?is)<title[^>]*>(.*?)</title>').Groups[1].Value.Trim()
    $err=[regex]::Match($dom+(Get-Content $e -Raw -ErrorAction SilentlyContinue),'ERR_[A-Z_]+').Value
    $verdict=if(-not $finished){'TIMEOUT (30 s)'}elseif($err){"FAILED $err"}elseif($dom.Length -lt 200){'EMPTY PAGE'}else{'PAGE LOADED'}
    Say "  $site [$verdict] $($w.ElapsedMilliseconds) ms, dom $($dom.Length) bytes, title: '$title'"
   }catch{Say "  $site ERROR $($_.Exception.Message)"}
   finally{
    if($p -and -not $p.HasExited){try{$p.Kill()}catch{}}
    Remove-Item $dir -Recurse -Force -ErrorAction SilentlyContinue;Remove-Item $o,$e -Force -ErrorAction SilentlyContinue
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
try{$all=[IO.File]::ReadAllText($result);[IO.File]::WriteAllText($result,(Scrub $all),$utf8)}catch{}
Write-Host "`nSaved: $result"
