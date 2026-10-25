# Installs VELIQ from an unpacked kit; restore-kit.ps1 decrypts the kit and
# calls this. Run elevated as the club user who owns the installation. Never
# uninstalls, never touches the VPS; a partly restored PC stops with an error.
[CmdletBinding()]
param(
 [string]$KitDirectory='',
 [string]$InstallDirectory="$env:ProgramData\neexitlag",
 [switch]$NoLaunch
)
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'common.ps1')
. (Join-Path $PSScriptRoot 'reinstall-state.ps1')
Assert-Administrator
# Windows PowerShell 5.1 can leave $PSScriptRoot empty inside param().
if(!$KitDirectory){$KitDirectory=Join-Path $PSScriptRoot '..\..'}
$kit=(Resolve-Path -LiteralPath $KitDirectory).Path.TrimEnd('\')
$owner=[Security.Principal.WindowsIdentity]::GetCurrent().User.Value
# Elevation through another admin account would install for the wrong SID and
# put the GUI in that account's profile.
$shell=@(Get-CimInstance Win32_Process -Filter "Name='explorer.exe'"|ForEach-Object{(Invoke-CimMethod -InputObject $_ -MethodName GetOwnerSid).Sid}|Select-Object -Unique)
if($shell.Count -and $shell -notcontains $owner){throw 'Elevated account differs from the signed-in desktop user; elevate the same account'}
'===== KIT ====='
$info=Get-Content -Raw -LiteralPath (Join-Path $kit 'kit.json')|ConvertFrom-Json
if($info.format -ne 1){throw 'Unsupported kit format'}
foreach($entry in $info.files.PSObject.Properties){
 $path=Join-Path $kit $entry.Name
 if(!(Test-Path -LiteralPath $path -PathType Leaf)){throw "Kit file missing: $($entry.Name)"}
 if((Get-FileHash -LiteralPath $path -Algorithm SHA256).Hash -ne $entry.Value){throw "Kit file hash mismatch: $($entry.Name)"}
}
if($info.owner_sid -ne $owner){Write-Warning 'Current user SID differs from the kit owner; installing for the current user'}
# Pins from fetch-windivert.ps1, not from the kit: the kit cannot vouch for itself.
$dllHash='C1E060EE19444A259B2162F8AF0F3FE8C4428A1C6F694DCE20DE194AC8D7D9A2'
$sysHash='8DA085332782708D8767BCACE5327A6EC7283C17CFB85E40B03CD2323A90DDC2'
$driver=Join-Path $kit 'driver'
if($info.dll_sha256 -ne $dllHash -or $info.driver_sha256 -ne $sysHash){throw 'Kit WinDivert pins differ from WinDivert 2.2.2-A'}
if((Get-AuthenticodeSignature (Join-Path $driver 'WinDivert64.sys')).Status -ne 'Valid'){throw 'Kernel driver signature is not trusted by this Windows installation'}
$serviceExe=Join-Path $kit 'bin\neexitlag-service.exe'
$target=Join-Path $InstallDirectory 'neexitlag-service.exe'
'===== SERVICE ====='
$existing=Get-Service neexitlag -ErrorAction SilentlyContinue
if($existing){
 # Only an identical running install counts as already restored.
 if($existing.Status -ne 'Running' -or !(Test-Path -LiteralPath $target) -or (Get-FileHash -LiteralPath $target).Hash -ne (Get-FileHash -LiteralPath $serviceExe).Hash){throw 'A different or stopped neexitlag service exists; not modified. Ask the owner before uninstalling.'}
 'Service already restored; skipping install.'
}else{
 if(Test-Path -LiteralPath $InstallDirectory){throw "$InstallDirectory exists without a service; not modified. Ask the owner before cleaning it."}
 $probe=Join-Path $kit 'secrets\probe.token'
 & (Join-Path $PSScriptRoot 'install.ps1') -ServiceExecutable $serviceExe -OwnerSid $owner -DeviceDirectory (Join-Path $kit 'secrets\device') -TrustDirectory (Join-Path $kit 'secrets\trust') -DriverDirectory $driver -DllSha256 $dllHash -DriverSha256 $sysHash -ControlPlaneUrl $info.control_plane_url -ProbeTarget $info.probe_target -TCPProbeTarget $info.tcp_probe_target -ProbeTokenFile $(if(Test-Path -LiteralPath $probe){$probe}else{''}) -InstallDirectory $InstallDirectory
 $state=Join-Path $kit 'state'
 if(Test-Path -LiteralPath $state){
  '===== STATE ====='
  $service=Get-Service neexitlag
  Stop-Service neexitlag
  $service.WaitForStatus('Stopped',[TimeSpan]::FromSeconds(30))
  Copy-Item -Path (Join-Path $state '*.json') -Destination (Join-Path $InstallDirectory 'state') -Force
  Protect-ServiceDirectory $InstallDirectory $owner
  Start-Service neexitlag
  $service.WaitForStatus('Running',[TimeSpan]::FromSeconds(30))
 }
}
$deadline=[DateTime]::UtcNow.AddSeconds(30);$ready=$false
do{
 try{$status=Invoke-ServiceIPC 'status' $null (Join-Path $InstallDirectory 'ipc.token');$ready=$null -ne $status.PSObject.Properties['backend']}catch{$ready=$false}
 if(!$ready -and [DateTime]::UtcNow -lt $deadline){Start-Sleep -Milliseconds 300}
}while(!$ready -and [DateTime]::UtcNow -lt $deadline)
if(!$ready){throw 'Service is running but IPC status is not ready'}
'===== DESKTOP ====='
$prefs=Join-Path $kit 'desktop\desktop.json'
$prefsTarget=Join-Path $env:LOCALAPPDATA 'NeExitLag\desktop.json'
if((Test-Path -LiteralPath $prefs) -and !(Test-Path -LiteralPath $prefsTarget)){[void][IO.Directory]::CreateDirectory((Split-Path $prefsTarget));Copy-Item -LiteralPath $prefs -Destination $prefsTarget}
$gui=Join-Path $env:LOCALAPPDATA 'Programs\VELIQ\VELIQ.exe'
if(!(Test-Path -LiteralPath $gui) -or (Get-FileHash -LiteralPath $gui).Hash -ne (Get-FileHash -LiteralPath (Join-Path $kit 'bin\VELIQ.exe')).Hash){& (Join-Path $PSScriptRoot 'install-desktop.ps1') -SourceDirectory (Join-Path $kit 'bin')}
$ssh=Join-Path $kit 'secrets\ssh'
if(Test-Path -LiteralPath $ssh){
 '===== SSH ====='
 # A key already on this PC is never replaced; OpenSSH refuses a private key
 # that other users can read, so it gets owner/SYSTEM/admins only.
 $sshTarget=Join-Path $env:USERPROFILE '.ssh'
 [void][IO.Directory]::CreateDirectory($sshTarget)
 foreach($file in Get-ChildItem -LiteralPath $ssh -File){
  $destination=Join-Path $sshTarget $file.Name
  if(Test-Path -LiteralPath $destination){
   if((Get-FileHash -LiteralPath $destination).Hash -ne (Get-FileHash -LiteralPath $file.FullName).Hash){Write-Warning "Existing $($file.Name) differs from the kit; kept"}
   continue
  }
  Copy-Item -LiteralPath $file.FullName -Destination $destination
  if($file.Name -eq 'id_ed25519'){
   & icacls $destination /inheritance:r /grant:r "*$($owner):F" '*S-1-5-18:F' '*S-1-5-32-544:F'|Out-Null
   if($LASTEXITCODE){throw 'Failed to restrict SSH private key permissions'}
  }
 }
}
'===== VERIFY ====='
$same=(Get-FileHash -LiteralPath $target).Hash -eq (Get-FileHash -LiteralPath $serviceExe).Hash
$running=(Get-Service neexitlag).Status
"service.exe SHA256 matches kit: $same; service: $running; build: $(if($status.PSObject.Properties['build']){$status.build}else{'n/a'})"
if(!$same -or $running -ne 'Running'){throw 'Verification failed'}
# explorer.exe starts the GUI unelevated, as it must run.
if(!$NoLaunch -and !@(Get-Process -Name VELIQ -ErrorAction SilentlyContinue).Count){Start-Process explorer.exe -ArgumentList ('"'+$gui+'"')}
Write-Host 'VELIQ restored. Check routes in the GUI; restoration is not a network acceptance result.'
