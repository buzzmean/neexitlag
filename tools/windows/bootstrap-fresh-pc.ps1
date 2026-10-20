# Fresh owner PC only. Default mode NEVER rotates/provisions the VPS.
[CmdletBinding()]
param(
 [Parameter(Mandatory=$true)][string]$VpsIP,
 [Parameter(Mandatory=$true)][ValidatePattern('^[a-z_][a-z0-9_-]{0,31}$')][string]$SSHUser,
 [ValidateRange(1,65535)][int]$SSHPort=22,
 [string]$CaFingerprint='',
 [switch]$RotateCA,
 [string]$ConfirmCARotation='',
 [string]$ExpectedCurrentCAFingerprint='',
 [switch]$ServicePhase,
 [string]$OwnerSid=''
)
$ErrorActionPreference='Stop'
Set-StrictMode -Version Latest
. (Join-Path $PSScriptRoot 'common.ps1')
$root=(Resolve-Path (Join-Path $PSScriptRoot '..\..')).Path
$state=Join-Path $root '.e2e\fresh-pc'

function Require-FreshPin([string]$Pin){
 $pin=$Pin.Replace(':','').ToLowerInvariant()
 if($pin -notmatch '^[a-f0-9]{64}$'){throw 'CA_SHA256_INVALID'}
 return $pin
}
function Assert-FreshTools {
 foreach($tool in @('git','go','dotnet','ssh.exe')){[void](Get-Command $tool -ErrorAction Stop)}
 if(!@(& dotnet --list-sdks | Where-Object {$_ -match '^8\.'}).Count){throw '.NET SDK 8 required; runtime alone is insufficient'}
 if($LASTEXITCODE){throw 'DOTNET_SDK_QUERY_FAILED'}
}
function Fresh-Property($Object,[string]$Name){
 $p=$Object.PSObject.Properties[$Name];if($null -eq $p){throw "Missing evidence field: $Name"};return $p.Value
}
function Assert-FreshSummary($Status){
 if((Fresh-Property $Status 'demo') -or !(Fresh-Property $Status 'connected')){throw 'SERVICE_NOT_AUTHENTICATED'}
 $backend=Fresh-Property $Status 'backend'
 if(!(Fresh-Property $backend 'ready')){throw 'CAPTURE_NOT_READY'}
}
function Wait-FreshServiceStartup([int]$TimeoutMilliseconds=60000){
 $until=[DateTime]::UtcNow.AddMilliseconds($TimeoutMilliseconds)
 while($true){
  try{
   $status=Invoke-ServiceIPC 'status'
   $network=Invoke-ServiceIPC 'local.network'
   $errorProperty=$network.PSObject.Properties['error']
   $observed=[DateTimeOffset]::Parse((Fresh-Property $network 'observed_at'))
   if((Fresh-Property $status 'connected') -and !(Fresh-Property $status 'demo') -and (Fresh-Property $network 'fingerprint') -ne '' -and $observed -gt [DateTimeOffset]::UtcNow.AddSeconds(-45) -and ($null -eq $errorProperty -or [string]$errorProperty.Value -eq '')){return}
  }catch{# Startup can transiently lack IPC/network; never print arbitrary payloads.
  }
  # Check readiness BEFORE deadline, including the last permitted sample.
  if([DateTime]::UtcNow -ge $until){throw 'SERVICE_CONTROLPLANE_OR_NETWORK_STARTUP_TIMEOUT'}
  Start-Sleep -Milliseconds 500
 }
}
function Fresh-Quote([string]$Value){return '"'+[regex]::Replace([regex]::Replace($Value,'(\\*)"','$1$1\"'),'(\\+)$','$1$1')+'"'}
function Invoke-FreshSSH([string]$Payload,[string]$RemoteCode,[string]$KnownHosts){
 # Keep a real console for OpenSSH host-key/password prompts. Only protocol stdout
 # is captured; no password, CA private key or administrator token enters this PS.
 $code=[Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($RemoteCode))
 $remote="python3 -c `"exec(__import__('base64').b64decode('$code'))`""
 if($SSHUser -ne 'root'){$remote='sudo -n '+$remote}
 $sshArguments=@('-T','-F','none','-p',"$SSHPort",'-o','StrictHostKeyChecking=ask','-o',"UserKnownHostsFile=$KnownHosts",'-o','GlobalKnownHostsFile=NUL','-o','ConnectTimeout=15','-o','ServerAliveInterval=15','-o','ServerAliveCountMax=3',"${SSHUser}@${VpsIP}",$remote)
 $info=[Diagnostics.ProcessStartInfo]::new('ssh.exe',(($sshArguments|ForEach-Object {Fresh-Quote $_}) -join ' '))
 $info.UseShellExecute=$false;$info.RedirectStandardInput=$true;$info.RedirectStandardOutput=$true;$info.CreateNoWindow=$false
 $p=[Diagnostics.Process]::Start($info)
 try{
  $output=$p.StandardOutput.ReadToEndAsync()
  $p.StandardInput.Write($Payload);$p.StandardInput.Close()
  if(!$p.WaitForExit(600000)){try{$p.Kill()}catch{};throw 'SSH_TIMEOUT; migration watchdog remains on VPS if armed'}
  $p.WaitForExit();$p.Refresh()
  if($p.ExitCode -ne 0){throw 'SSH_ENROLLMENT_FAILED; no installation performed'}
  $result=$output.GetAwaiter().GetResult()
  if($result.Length -gt 32768){throw 'SSH_RESPONSE_TOO_LARGE'}
  return $result
 }finally{$p.Dispose()}
}
function Protect-FreshState([string]$Directory,[string]$Sid){
 foreach($item in @((Get-Item -LiteralPath $Directory -Force)) + @(Get-ChildItem -LiteralPath $Directory -Recurse -Force)){
  if($item.Attributes -band [IO.FileAttributes]::ReparsePoint){throw 'FRESH_REPARSE_POINT_REFUSED'}
  if($item.PSIsContainer){$acl=[Security.AccessControl.DirectorySecurity]::new();$inherit=[Security.AccessControl.InheritanceFlags]'ContainerInherit,ObjectInherit'}
  else{$acl=[Security.AccessControl.FileSecurity]::new();$inherit=[Security.AccessControl.InheritanceFlags]::None}
  $acl.SetAccessRuleProtection($true,$false)
  foreach($allowed in @('S-1-5-18','S-1-5-32-544',$Sid)){
   [void]$acl.AddAccessRule([Security.AccessControl.FileSystemAccessRule]::new([Security.Principal.SecurityIdentifier]::new($allowed),[Security.AccessControl.FileSystemRights]::FullControl,$inherit,[Security.AccessControl.PropagationFlags]::None,[Security.AccessControl.AccessControlType]::Allow))
  }
  Set-Acl -LiteralPath $item.FullName -AclObject $acl -ErrorAction Stop
 }
}
function Invoke-FreshPC {
 $script:FreshStage='preflight'
 if($env:OS -ne 'Windows_NT'){throw 'Windows 10/11 x64 required'}
 $script:VpsIP=Assert-VpsIPv4 $VpsIP
 if($VpsIP -ne '2.56.177.4'){throw 'This owner bootstrap is scoped to existing VPS 2.56.177.4'}
 $sid=[Security.Principal.WindowsIdentity]::GetCurrent().User.Value
 $principal=[Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()
 $admin=$principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
 $bin=Join-Path $root 'bin\windows-e2e';$helper=Join-Path $bin 'fresh-pc.exe';$driver=Join-Path $root '.e2e\windivert'
 if($ServicePhase){
  if(!$admin -or $OwnerSid -cne $sid){throw 'Elevate the SAME installation owner SID, not another administrator'}
  $pins=Get-Content (Join-Path $driver 'pins.json') -Raw|ConvertFrom-Json
  & (Join-Path $PSScriptRoot 'install.ps1') -ServiceExecutable (Join-Path $bin 'neexitlag-service.exe') -OwnerSid $OwnerSid -DeviceDirectory (Join-Path $state 'device') -TrustDirectory (Join-Path $state 'trust') -DriverDirectory $driver -DllSha256 $pins.dll_sha256 -DriverSha256 $pins.driver_sha256 -ControlPlaneUrl "https://${VpsIP}:8443" -ProbeTarget "${VpsIP}:40000" -TCPProbeTarget "${VpsIP}:40000" -ProbeTokenFile (Join-Path $state 'echo.token')
  . (Join-Path $PSScriptRoot 'native-tcp-firewall.ps1')
  [void](Add-NativeTcpFirewallAdmission (Join-Path $env:ProgramData 'neexitlag\neexitlag-service.exe'))
  $installed=Join-Path $env:ProgramData 'neexitlag'
  $manifest=Get-Content (Join-Path $installed 'service.json') -Raw|ConvertFrom-Json
  if($manifest.owner_sid -cne $OwnerSid){throw 'INSTALLED_OWNER_MISMATCH'}
  foreach($pair in @(@('device\identity.json','identity.json'),@('device\device.pem','device.pem'),@('trust\ca.pem','ca.pem'),@('trust\trust.json','trust.json'),@('echo.token','probe.token'))){
   if((Get-FileHash (Join-Path $state $pair[0])).Hash -ne (Get-FileHash (Join-Path $installed $pair[1])).Hash){throw 'INSTALLED_CREDENTIAL_MISMATCH'}
  }
  $proof=@{owner_sid=$OwnerSid;sha256=(Get-FileHash (Join-Path $installed 'neexitlag-service.exe')).Hash}
  [IO.File]::WriteAllText((Join-Path $state 'service-install-proof.json'),($proof|ConvertTo-Json),[Text.UTF8Encoding]::new($false))
  return
 }
 if($admin){throw 'Run bootstrap from ordinary owner PowerShell; only the service phase requests UAC. GUI must not run elevated.'}
 if(Get-Service neexitlag -ErrorAction SilentlyContinue){throw 'EXISTING_INSTALLATION_PRESERVED: use the normal updater; fresh bootstrap does not rotate an installed identity'}
 if(Test-Path (Join-Path $env:ProgramData 'neexitlag')){throw 'PRESERVED_INSTALLATION_FOUND: do not delete credentials; use explicit reinstall/update'}
 if($RotateCA -and $CaFingerprint -ne ''){throw 'Use ExpectedCurrentCAFingerprint for rotation; the new CA pin is generated on the VPS'}
 if($RotateCA -and $ConfirmCARotation -cne "ROTATE-CA-$VpsIP"){throw 'Explicit -ConfirmCARotation ROTATE-CA-2.56.177.4 required'}
 if(!$RotateCA -and $ConfirmCARotation -ne ''){throw 'Confirmation without -RotateCA is invalid'}
 if($CaFingerprint -ne ''){$script:CaFingerprint=Require-FreshPin $CaFingerprint}
 if($ExpectedCurrentCAFingerprint -ne ''){$script:ExpectedCurrentCAFingerprint=Require-FreshPin $ExpectedCurrentCAFingerprint}
 if(@(Get-Process -Name VELIQ,NeExitLag -ErrorAction SilentlyContinue).Count){throw 'Exit only VELIQ GUI before fresh bootstrap; Chrome/Edge are handled by existing activation'}
 Assert-FreshTools
 foreach($directory in @($root,(Join-Path $root '.e2e'),$state)){
  if((Test-Path -LiteralPath $directory) -and ((Get-Item -LiteralPath $directory -Force).Attributes -band [IO.FileAttributes]::ReparsePoint)){throw 'Reparse enrollment directories refused'}
 }
 New-Item -ItemType Directory -Force $state|Out-Null
 Protect-FreshState $state $sid
 $script:FreshStage='build'
 & (Join-Path $PSScriptRoot 'build-e2e.ps1')
 $nuget=Join-Path $state 'NuGet.Config'
 [IO.File]::WriteAllText($nuget,'<configuration><packageSources><clear/><add key="nuget.org" value="https://api.nuget.org/v3/index.json" /></packageSources></configuration>',[Text.UTF8Encoding]::new($false))
 $priorRestore=$env:RestoreConfigFile
 try{$env:RestoreConfigFile=$nuget;& (Join-Path $PSScriptRoot 'build-desktop.ps1')}
 finally{$env:RestoreConfigFile=$priorRestore}
 if(!(Test-Path (Join-Path $driver 'pins.json'))){& (Join-Path $PSScriptRoot 'fetch-windivert.ps1') -Destination $driver}
 Push-Location $root
 $oldOS=$env:GOOS;$oldArch=$env:GOARCH;$oldCGO=$env:CGO_ENABLED
 try{
  $env:GOOS='windows';$env:GOARCH='amd64';$env:CGO_ENABLED='0'
  & go build -trimpath -o $helper ./tools/fresh-pc;if($LASTEXITCODE){throw 'FRESH_HELPER_BUILD_FAILED'}
  $env:GOOS='linux'
  $linux=Join-Path $bin 'linux\fresh-pc'
  & go build -trimpath -o $linux ./tools/fresh-pc;if($LASTEXITCODE){throw 'LINUX_ENROLL_HELPER_BUILD_FAILED'}
 }finally{$env:GOOS=$oldOS;$env:GOARCH=$oldArch;$env:CGO_ENABLED=$oldCGO;Pop-Location}
 $uuid=([string](Get-CimInstance Win32_ComputerSystemProduct -ErrorAction Stop).UUID).Trim().ToLowerInvariant()
 if($uuid -notmatch '^[a-f0-9-]{36}$' -or $uuid -match '^(0|-)+$|^(f|-)+$'){throw 'Stable unique SMBIOS UUID required; cloned/empty UUIDs cannot identify a machine safely'}
 $hash=[Security.Cryptography.SHA256]::Create();try{$machine=([BitConverter]::ToString($hash.ComputeHash([Text.Encoding]::UTF8.GetBytes('veliq-machine-v1|'+$uuid)))).Replace('-','').ToLowerInvariant()}finally{$hash.Dispose()}
 $script:FreshStage='csr'
 & $helper request --state $state --machine-hash $machine;if($LASTEXITCODE){throw 'CSR_GENERATION_FAILED'}
 $request=Get-Content (Join-Path $state 'request.json') -Raw|ConvertFrom-Json
 $payload=@{request=$request;vps_ip=$VpsIP;helper=[Convert]::ToBase64String([IO.File]::ReadAllBytes($linux));helper_sha256=(Get-FileHash $linux).Hash.ToLowerInvariant();rotate=[bool]$RotateCA;confirm=$ConfirmCARotation;expected_ca=$ExpectedCurrentCAFingerprint;enrollment_ca=$CaFingerprint;migration=[Convert]::ToBase64String([IO.File]::ReadAllBytes((Join-Path $root 'deploy\vps\rotate-fresh-pc-ca.py')))}|ConvertTo-Json -Depth 5 -Compress
 $remote=[IO.File]::ReadAllText((Join-Path $root 'deploy\vps\fresh-pc-ssh.py'))
 $responsePath=Join-Path $state 'response.json'
 try{
 $script:FreshStage='ssh_enrollment'
  $wire=Invoke-FreshSSH $payload $remote (Join-Path $state 'known_hosts')
  [IO.File]::WriteAllText($responsePath,$wire,[Text.UTF8Encoding]::new($false))
  $reply=$wire|ConvertFrom-Json
  # This pin arrives over owner-verified SSH, NOT untrusted HTTPS. Owner can
  # supply -CaFingerprint as an independent stronger expected value.
  $pin=if($CaFingerprint -ne ''){$CaFingerprint}else{Require-FreshPin (Fresh-Property $reply 'ca_sha256')}
 $script:FreshStage='ca_and_device_verification'
  & $helper complete --state $state --ca-sha256 $pin;if($LASTEXITCODE){throw 'CA_PIN_OR_DEVICE_CERTIFICATE_REJECTED; service not installed'}
 $script:FreshStage='external_tcp_udp_probe'
  & $helper probe --state $state --vps-ip $VpsIP;if($LASTEXITCODE){throw 'REAL_CLIENT_RELAY_CHECK_FAILED; service not installed'}
 }finally{Remove-Item -LiteralPath $responsePath -Force -ErrorAction SilentlyContinue;$wire=$null;$payload=$null}
 # UAC only now: all credentials and the actual Internet relay have been verified.
 $scriptPath=Join-Path $PSScriptRoot 'bootstrap-fresh-pc.ps1'
 $q={param($v) "'"+$v.Replace("'","''")+"'"}
 $code='& '+(& $q $scriptPath)+' -VpsIP '+(& $q $VpsIP)+' -SSHUser '+(& $q $SSHUser)+' -SSHPort '+$SSHPort+' -ServicePhase -OwnerSid '+(& $q $sid)
 $encoded=[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes("try { $code; exit 0 } catch { Write-Error 'Service installation failed; preserved files were not deleted'; exit 1 }"))
 $script:FreshStage='service_install'
 Remove-Item (Join-Path $state 'service-install-proof.json') -Force -ErrorAction SilentlyContinue
 $child=Start-Process powershell.exe -Verb RunAs -ArgumentList @('-NoProfile','-EncodedCommand',$encoded) -Wait -PassThru
 $child.WaitForExit();$child.Refresh();if($child.ExitCode -ne 0){throw 'SERVICE_INSTALL_FAILED; inspect preserved state, do not delete credentials'}
 $proof=Get-Content (Join-Path $state 'service-install-proof.json') -Raw|ConvertFrom-Json
 if((Fresh-Property $proof 'owner_sid') -cne $sid -or (Fresh-Property $proof 'sha256') -ne (Get-FileHash (Join-Path $bin 'neexitlag-service.exe')).Hash){throw 'INSTALLED_SERVICE_PROOF_MISMATCH'}
 if((Get-FileHash (Join-Path $bin 'neexitlag-service.exe')).Hash -ne (Get-FileHash (Join-Path $env:ProgramData 'neexitlag\neexitlag-service.exe')).Hash){throw 'INSTALLED_SERVICE_HASH_MISMATCH'}
 # Installed private credentials are now SYSTEM/admin-only. Remove the owner's
 # enrollment staging copies; nothing from this boot is needed after freeze.
 Remove-Item -LiteralPath (Join-Path $state 'device\identity.json'),(Join-Path $state 'echo.token') -Force -ErrorAction Stop
 $script:FreshStage='service_startup'
 Wait-FreshServiceStartup
 $script:FreshStage='gui_install'
 & (Join-Path $PSScriptRoot 'install-desktop.ps1')
 $gui=Join-Path $env:LOCALAPPDATA 'Programs\VELIQ\VELIQ.exe'
 $report=Join-Path $state 'gui-readiness.json'
 Remove-Item $report -Force -ErrorAction SilentlyContinue
 # Reuses existing GUI PAC+nonce/owner readiness transaction. No fake IPC PASS.
 $script:FreshStage='pac_readiness'
 Start-Process $gui -ArgumentList @('--fresh-pc-setup',(Fresh-Quote $report))|Out-Null
 $until=[DateTime]::UtcNow.AddMinutes(5)
 while(!(Test-Path $report) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 500}
 if(!(Test-Path $report)){throw 'GUI_PAC_READINESS_TIMEOUT; installation retained, no PASS'}
 $ready=Get-Content $report -Raw|ConvertFrom-Json
 if((Fresh-Property $ready 'pass') -cne $true){throw ('GUI_SETUP_FAILED: '+(Fresh-Property $ready 'stage'))}
 $script:FreshStage='service_status'
 Assert-FreshSummary (Invoke-ServiceIPC 'status')
 Write-Host 'PASS: fresh identity, pinned trust, real TCP+UDP relay, installed service/GUI, owned PAC and Optimize readiness. Website/ISP performance not asserted.'
}
if($MyInvocation.InvocationName -eq '.'){return}
try{Invoke-FreshPC}catch{Write-Error ('FRESH_PC_BOOTSTRAP_FAILED stage='+$script:FreshStage+' class='+$_.Exception.GetType().Name+'; installation/state preserved; see last safe stage/error code above');exit 1}
