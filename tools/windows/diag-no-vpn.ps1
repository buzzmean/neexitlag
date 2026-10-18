#requires -Version 5.1
<#
Non-interactive, read-only diagnostics to run with the VPN OFF, as a normal user:
  powershell -NoProfile -ExecutionPolicy Bypass -File .\tools\windows\diag-no-vpn.ps1
Writes diag-result.txt next to this script (falls back to %TEMP% if that folder is
not writable). Changes nothing. Reads no token/identity/certificate into the report;
the whole report is scrubbed of secret-shaped values before the run ends.
Self-contained: does not need common.ps1.
#>
param(
 [string]$VpsHost='2.56.177.4',
 [int]$ControlPlanePort=8443,
 [int]$QuicPort=4433,
 [string]$InstallDirectory="$env:ProgramData\neexitlag",
 [string]$BuildDirectory=''
)
$ErrorActionPreference='Continue'
Set-StrictMode -Version 2
$scriptDir=if($PSScriptRoot){$PSScriptRoot}else{(Get-Location).Path}
if(-not $BuildDirectory){$BuildDirectory=Join-Path (Split-Path $scriptDir -Parent | Split-Path -Parent) 'bin\windows-e2e'}
$result=Join-Path $scriptDir 'diag-result.txt'
try{[IO.File]::WriteAllText($result,'',[Text.UTF8Encoding]::new($false))}catch{$result=Join-Path $env:TEMP 'diag-result.txt';[IO.File]::WriteAllText($result,'',[Text.UTF8Encoding]::new($false))}

function Say([string]$Text){Write-Host $Text;[IO.File]::AppendAllText($result,$Text+"`r`n",[Text.UTF8Encoding]::new($false))}
function Section([string]$Title){Say '';Say ('=== '+$Title+' ===')}
function Scrub([string]$Text){
 $Text=[regex]::Replace($Text,'-----BEGIN [A-Z ]+-----[\s\S]*?-----END [A-Z ]+-----','[PEM REMOVED]')
 [regex]::Replace($Text,'(?i)("?(token|secret|private[_ ]?key|seed|password|authorization)"?\s*[:=]\s*)"?[^",\s}]+"?','$1"[REDACTED]"')
}
function Json($Value){Scrub ($Value|ConvertTo-Json -Depth 12)}
function Step([string]$Name,[scriptblock]$Body){try{& $Body}catch{Say ("[$Name] ERROR: "+(Scrub $_.Exception.Message))}}

# Minimal IPC client (same wire format as the GUI: 4-byte big-endian length + JSON).
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
 $tokenFile=Join-Path $InstallDirectory 'ipc.token'
 $token=[IO.File]::ReadAllText($tokenFile).Trim()   # used only to authenticate; never printed
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
function Prop($Object,[string]$Name,$Default=$null){
 if($null -ne $Object -and $Object.PSObject.Properties.Name -contains $Name){return $Object.$Name};return $Default
}

Say "VELIQ diag-no-vpn, started $((Get-Date).ToString('yyyy-MM-dd HH:mm:ss zzz'))"
Say "Result file: $result"

Section '0. Environment (is the VPN really off?)'
Step 'env' {
 $os=Get-CimInstance Win32_OperatingSystem
 Say "OS: $($os.Caption) build $($os.BuildNumber); user: $env:USERNAME; PowerShell $($PSVersionTable.PSVersion)"
 $vpn=@(Get-NetAdapter|Where-Object{$_.Status -eq 'Up' -and ($_.InterfaceDescription+' '+$_.Name) -match 'VPN|Happ|TUN|TAP|WireGuard|Wintun'}|ForEach-Object{$_.Name})
 Say ("Active VPN-like adapters: "+$(if($vpn.Count){($vpn -join ', ')+'  <-- VPN looks ON, results below are NOT a no-VPN test'}else{'none (good)'}))
 Say ("Default route(s): "+((Get-NetRoute -DestinationPrefix '0.0.0.0/0' -ErrorAction SilentlyContinue|Sort-Object RouteMetric|ForEach-Object{"if$($_.InterfaceIndex) via $($_.NextHop) metric $($_.RouteMetric)"}) -join '; '))
 Say ("System DNS: "+((Get-DnsClientServerAddress -AddressFamily IPv4|Where-Object{$_.ServerAddresses}|ForEach-Object{$_.ServerAddresses -join ','}|Select-Object -Unique) -join '; '))
}

Section '1a. Service and binary'
Step 'service' {
 $svc=Get-Service neexitlag -ErrorAction Stop
 Say "Service neexitlag: $($svc.Status), start type $($svc.StartType)"
 $installed=Join-Path $InstallDirectory 'neexitlag-service.exe'
 $built=Join-Path $BuildDirectory 'neexitlag-service.exe'
 $hashes=@{}
 foreach($pair in @(@('installed',$installed),@('built',$built))){
  if(Test-Path $pair[1]){$h=(Get-FileHash $pair[1] -Algorithm SHA256).Hash;$hashes[$pair[0]]=$h;Say "SHA256 $($pair[0]): $h (modified $((Get-Item $pair[1]).LastWriteTime.ToString('yyyy-MM-dd HH:mm')))"}
  else{Say "SHA256 $($pair[0]): file not found ($($pair[1]))"}
 }
 if($hashes.ContainsKey('installed') -and $hashes.ContainsKey('built')){Say ("installed == built: "+($hashes['installed'] -eq $hashes['built']))}
}
Step 'desktop' {
 $exe=Get-ChildItem "$env:LOCALAPPDATA\Programs","$env:LOCALAPPDATA\NeExitLag","$env:ProgramFiles" -Filter 'VELIQ.exe' -Recurse -Depth 3 -ErrorAction SilentlyContinue|Select-Object -First 1
 Say ("VELIQ.exe: "+$(if($exe){"$($exe.FullName) v$($exe.VersionInfo.FileVersion)"}else{'not found in the usual places'}))
}

Section '1b. IPC status'
$status=$null
Step 'status' {
 $script:status=Invoke-IPC 'status'
 Say ("optimizing: "+(Prop $script:status 'optimizing'))
 Say (Json $script:status)
}
Section '1c. IPC profiles (what Optimize covers)'
Step 'profiles' {
 $rows=@(Invoke-IPC 'profiles');Say "Profile count: $($rows.Count)"
 foreach($p in $rows){Say ("- id=$(Prop $p 'id') name=$(Prop $p 'name') mode=$(Prop $p 'mode') fallback=$(Prop $p 'fallback') preset=$(Prop $p 'app_preset') exe=$((@(Prop $p 'executables' @())) -join ' | ') rules=$(@(Prop $p 'rules' @()).Count)")}
 $exes=($rows|ForEach-Object{@(Prop $_ 'executables' @())})-join ' '
 Say ("Has Chrome profile: "+($exes -match 'chrome\.exe')+"; Edge: "+($exes -match 'msedge\.exe')+"; Telegram: "+($exes -match '(?i)telegram\.exe')+"; Discord: "+($exes -match '(?i)discord'))
}
Section '1d. IPC browser.status'
Step 'browser' {Say (Json (Invoke-IPC 'browser.status'))}
Section '1e. IPC capture.diagnostics'
Step 'capture' {Say (Json (Invoke-IPC 'capture.diagnostics'))}
Section '1f. IPC quality.diagnostics'
Step 'quality' {Say (Json (Invoke-IPC 'quality.diagnostics'))}
Section '1g. IPC relay.pool.diagnostics (handshake/x509/dial errors show up here)'
Step 'pool' {
 $ids=@();if($script:status){$ids=@(Prop $script:status 'relays' @()|ForEach-Object{Prop $_ 'id'}|Where-Object{$_})}
 if($ids.Count -eq 0){$ids=@('VPS-01')}
 foreach($id in $ids){Say ("relay $id : "+(Json (Invoke-IPC 'relay.pool.diagnostics' @{relay_id=$id})))}
}

Section "2a. $VpsHost TCP $ControlPlanePort (control plane)"
Step 'tcp' {
 $client=New-Object Net.Sockets.TcpClient
 try{
  $watch=[Diagnostics.Stopwatch]::StartNew()
  $task=$client.ConnectAsync($VpsHost,$ControlPlanePort)
  $ok=$false;try{$ok=$task.Wait(6000) -and $client.Connected}catch{Say ("TCP connect error: "+$_.Exception.GetBaseException().Message)}
  if(-not $ok){Say "TCP ${ControlPlanePort}: NO CONNECTION within 6 s (blocked, dropped or nothing listening)";return}
  Say "TCP ${ControlPlanePort}: connected in $($watch.ElapsedMilliseconds) ms"
  # Records the presented certificate only. The product's own verification is not involved.
  $seen=@{Errors='unknown'}
  $cb=[Net.Security.RemoteCertificateValidationCallback]{param($s,$c,$ch,$e) $seen.Errors=[string]$e;$true}
  $ssl=New-Object Net.Security.SslStream($client.GetStream(),$false,$cb)
  $ssl.AuthenticateAsClient($VpsHost)
  $cert=New-Object Security.Cryptography.X509Certificates.X509Certificate2($ssl.RemoteCertificate)
  $san=(($cert.Extensions|Where-Object{$_.Oid.Value -eq '2.5.29.17'}|ForEach-Object{$_.Format($false)}) -join '')
  Say "TLS handshake OK: $($ssl.SslProtocol); subject $($cert.Subject); issuer $($cert.Issuer)"
  Say "  valid $($cert.NotBefore.ToString('yyyy-MM-dd')) .. $($cert.NotAfter.ToString('yyyy-MM-dd')); SAN: $san"
  Say "  system-store verdict (a private CA is expected to be untrusted here): $($seen.Errors)"
  $ssl.Dispose()
 }finally{$client.Close()}
}

Section "2b. $VpsHost UDP $QuicPort (QUIC relay)"
Step 'udp' {
 # A QUIC long-header packet with a reserved version makes a QUIC server answer with a
 # Version Negotiation packet. That is a real UDP round trip. Silence does NOT prove a
 # block (relay down, rate limit, or reply dropped) and is reported as such.
 $rng=[Security.Cryptography.RandomNumberGenerator]::Create()
 $dcid=New-Object byte[] 8;$scid=New-Object byte[] 8;$rng.GetBytes($dcid);$rng.GetBytes($scid)
 $packet=New-Object byte[] 1200
 $packet[0]=0xC0;$packet[1]=0x1A;$packet[2]=0x2A;$packet[3]=0x3A;$packet[4]=0x4A
 $packet[5]=8;[Array]::Copy($dcid,0,$packet,6,8);$packet[14]=8;[Array]::Copy($scid,0,$packet,15,8)
 $udp=New-Object Net.Sockets.UdpClient
 try{
  $udp.Client.ReceiveTimeout=2500
  $udp.Connect($VpsHost,$QuicPort)
  $answered=$false
  for($attempt=1;$attempt -le 3 -and -not $answered;$attempt++){
   $watch=[Diagnostics.Stopwatch]::StartNew()
   [void]$udp.Send($packet,$packet.Length)
   try{
    $remote=New-Object Net.IPEndPoint([Net.IPAddress]::Any,0)
    $reply=$udp.Receive([ref]$remote)
    $kind=if($reply.Length -ge 5 -and ($reply[0] -band 0x80) -and $reply[1] -eq 0 -and $reply[2] -eq 0 -and $reply[3] -eq 0 -and $reply[4] -eq 0){'QUIC Version Negotiation'}else{'unrecognised datagram'}
    Say "UDP ${QuicPort}: REPLY on attempt $attempt after $($watch.ElapsedMilliseconds) ms: $($reply.Length) bytes, $kind -> UDP/QUIC path to the relay works"
    $answered=$true
   }catch [Net.Sockets.SocketException]{
    Say "UDP ${QuicPort}: attempt $attempt no reply ($($_.Exception.SocketErrorCode))"
   }
  }
  if(-not $answered){Say "UDP ${QuicPort}: NO REPLY in 3 attempts. If TCP ${ControlPlanePort} is open, the ISP is probably dropping UDP/QUIC to this host (the relay transport is QUIC only); not proven."}
 }finally{$udp.Close()}
}

Section '3. HTTPS to the target sites (plain curl.exe, not through Chrome/PAC)'
$sites=@('chatgpt.com','claude.ai','www.youtube.com','web.telegram.org','discord.com')
$curl=(Get-Command curl.exe -ErrorAction SilentlyContinue)
foreach($name in $sites){
 Step "site $name" {
  $ips='dns failed'
  try{$ips=(@(Resolve-DnsName $name -Type A -DnsOnly -ErrorAction Stop|Where-Object{$_.Type -eq 'A'}|ForEach-Object{$_.IPAddress}) -join ',')}catch{$ips='DNS error: '+$_.Exception.Message}
  if($curl){
   $text=(& curl.exe -sS -m 15 -o NUL -w 'HTTP %{http_code} ip=%{remote_ip} connect=%{time_connect}s tls=%{time_appconnect}s total=%{time_total}s' "https://$name/" 2>&1|Out-String).Trim() -replace '\s+',' '
   $verdict=if($text -match 'HTTP (2|3)\d\d'){'OPENS'}elseif($text -match 'HTTP (4|5)\d\d'){'ANSWERS (HTTP error, host reachable)'}else{'FAILS'}
   Say "$name [$verdict] dns=$ips | $text"
  }else{
   try{$r=Invoke-WebRequest "https://$name/" -UseBasicParsing -TimeoutSec 15 -MaximumRedirection 0 -ErrorAction Stop;Say "$name [OPENS] dns=$ips | HTTP $([int]$r.StatusCode)"}
   catch{$resp=$_.Exception.Response;if($resp){Say "$name [ANSWERS] dns=$ips | HTTP $([int]$resp.StatusCode)"}else{Say "$name [FAILS] dns=$ips | $($_.Exception.Message)"}}
  }
 }
}

Section '4. Installed apps and browser coverage'
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
}

Say ''
Say "Finished $((Get-Date).ToString('yyyy-MM-dd HH:mm:ss'))."
# Final scrub of the whole file, then tell the user where it is.
try{$all=[IO.File]::ReadAllText($result);[IO.File]::WriteAllText($result,(Scrub $all),[Text.UTF8Encoding]::new($false))}catch{}
Write-Host "`nSaved: $result"
