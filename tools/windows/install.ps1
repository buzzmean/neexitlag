# Run from an elevated PowerShell after building the service and provisioning device credentials.
[CmdletBinding()]
param(
 [Parameter(Mandatory=$true)][string]$ServiceExecutable,
 [Parameter(Mandatory=$true)][string]$OwnerSid,
 [Parameter(Mandatory=$true)][string]$DeviceDirectory,
 [Parameter(Mandatory=$true)][string]$TrustDirectory,
 [Parameter(Mandatory=$true)][string]$DriverDirectory,
 [Parameter(Mandatory=$true)][string]$DllSha256,
 [Parameter(Mandatory=$true)][string]$DriverSha256,
 [Parameter(Mandatory=$true)][string]$ControlPlaneUrl,
 [Parameter(Mandatory=$true)][string]$ProbeTarget,
 [string]$TCPProbeTarget="",
 [string]$ProbeTokenFile="",
 [string]$InstallDirectory="$env:ProgramData\neexitlag"
)
$ErrorActionPreference="Stop"
$current=[Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()
if(!$current.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw "Administrator access is required to install the service."}
if(Get-Service neexitlag -ErrorAction SilentlyContinue){throw "Service already installed. Stop it and use an explicit upgrade/rollback procedure."}
[void][Security.Principal.SecurityIdentifier]::new($OwnerSid)
$ServiceExecutable=(Resolve-Path $ServiceExecutable).Path
$DriverDirectory=(Resolve-Path $DriverDirectory).Path
if($DllSha256 -notmatch '^[0-9a-fA-F]{64}$' -or $DriverSha256 -notmatch '^[0-9a-fA-F]{64}$'){throw "SHA256 pins must be provided from a trusted WinDivert distribution."}
foreach($pair in @(@("WinDivert.dll",$DllSha256),@("WinDivert64.sys",$DriverSha256))){if((Get-FileHash (Join-Path $DriverDirectory $pair[0]) -Algorithm SHA256).Hash -ne $pair[1]){throw "WinDivert hash mismatch: $($pair[0])"}}
New-Item -ItemType Directory -Path $InstallDirectory -ErrorAction Stop | Out-Null
& icacls $InstallDirectory /inheritance:r /grant:r '*S-1-5-18:(OI)(CI)F' '*S-1-5-32-544:(OI)(CI)F' | Out-Null
if($LASTEXITCODE){throw "Failed to protect service directory"}
Copy-Item $ServiceExecutable (Join-Path $InstallDirectory "neexitlag-service.exe")
New-Item -ItemType Directory -Path (Join-Path $InstallDirectory "driver") | Out-Null
Copy-Item (Join-Path $DriverDirectory "WinDivert.dll") (Join-Path $InstallDirectory "driver")
Copy-Item (Join-Path $DriverDirectory "WinDivert64.sys") (Join-Path $InstallDirectory "driver")
Copy-Item (Join-Path $DeviceDirectory "identity.json") (Join-Path $InstallDirectory "identity.json")
Copy-Item (Join-Path $DeviceDirectory "device.pem") (Join-Path $InstallDirectory "device.pem")
Copy-Item (Join-Path $TrustDirectory "trust.json") (Join-Path $InstallDirectory "trust.json")
Copy-Item (Join-Path $TrustDirectory "ca.pem") (Join-Path $InstallDirectory "ca.pem")
$bytes=New-Object byte[] 32
$rng=[Security.Cryptography.RandomNumberGenerator]::Create()
$rng.GetBytes($bytes)
$rng.Dispose()
$token=[Convert]::ToBase64String($bytes)
$tokenFile=Join-Path $InstallDirectory "ipc.token"
[IO.File]::WriteAllText($tokenFile,$token,[Text.UTF8Encoding]::new($false))
& icacls $InstallDirectory /grant "*$($OwnerSid):(RX)" | Out-Null
& icacls $tokenFile /inheritance:r /grant:r '*S-1-5-18:F' '*S-1-5-32-544:F' "*$($OwnerSid):R" | Out-Null
if($LASTEXITCODE){throw "Failed to protect IPC token"}
# The GUI owner can traverse the top directory and read only the IPC token; no inherited secret access.
New-Item -ItemType Directory -Path (Join-Path $InstallDirectory "state") | Out-Null
$manifest=@{owner_sid=$OwnerSid;token_file=$tokenFile;state_directory=(Join-Path $InstallDirectory "state");control_plane_url=$ControlPlaneUrl;trust_file=(Join-Path $InstallDirectory "trust.json");device_identity_file=(Join-Path $InstallDirectory "identity.json");device_certificate=(Join-Path $InstallDirectory "device.pem");ca_file=(Join-Path $InstallDirectory "ca.pem");probe_target=$ProbeTarget;tcp_probe_target=$TCPProbeTarget;driver=@{directory=(Join-Path $InstallDirectory "driver");dll_sha256=$DllSha256;driver_sha256=$DriverSha256}}
if($ProbeTokenFile -ne ""){Copy-Item $ProbeTokenFile (Join-Path $InstallDirectory "probe.token");$manifest.probe_token_file=(Join-Path $InstallDirectory "probe.token")}
$cpCA=Join-Path $TrustDirectory "cp-ca.pem"
if(Test-Path $cpCA){Copy-Item $cpCA (Join-Path $InstallDirectory "cp-ca.pem");$manifest.control_plane_ca_file=(Join-Path $InstallDirectory "cp-ca.pem")}
$manifestPath=Join-Path $InstallDirectory "service.json"
[IO.File]::WriteAllText($manifestPath,($manifest|ConvertTo-Json -Depth 4),[Text.UTF8Encoding]::new($false))
& icacls $InstallDirectory /setowner "*S-1-5-32-544" /T | Out-Null
if($LASTEXITCODE){throw "Failed to set privileged ownership on service files"}
$binary='"'+(Join-Path $InstallDirectory "neexitlag-service.exe")+'" -config "'+$manifestPath+'"'
New-Service -Name neexitlag -BinaryPathName $binary -DisplayName "neexitlag Routing Service" -StartupType Automatic | Out-Null
& sc.exe failure neexitlag reset= 86400 actions= restart/5000/restart/5000/restart/5000 | Out-Null
if($LASTEXITCODE){throw "Failed to configure service recovery"}
Start-Service neexitlag
(Get-Service neexitlag).WaitForStatus('Running',[TimeSpan]::FromSeconds(30))
Write-Host "neexitlag service installed. Run the desktop GUI without elevation."
