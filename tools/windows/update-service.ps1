[CmdletBinding()]
param(
 [string]$ServiceExecutable=(Join-Path $PSScriptRoot '..\..\bin\windows-e2e\neexitlag-service.exe'),
 [string]$InstallDirectory="$env:ProgramData\neexitlag"
)
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'common.ps1')
Assert-Administrator
function Assert-ServiceUpdateScmTarget([string]$CommandLine,[string]$ExpectedPath){
 $match=[regex]::Match($CommandLine,'^\s*"([^"]+)"(?:\s|$)')
 if(!$match.Success -or [IO.Path]::GetFullPath($match.Groups[1].Value) -ine [IO.Path]::GetFullPath($ExpectedPath)){throw 'SCM executable does not match the requested installation; update refused'}
}
$source=(Get-Item -LiteralPath $ServiceExecutable -Force).FullName
$root=(Get-Item -LiteralPath $InstallDirectory -Force).FullName
$target=Join-Path $root 'neexitlag-service.exe';$manifestFile=Join-Path $root 'service.json'
foreach($path in @($source,$root,$target,$manifestFile)){
 if((Get-Item -LiteralPath $path -Force).Attributes -band [IO.FileAttributes]::ReparsePoint){throw 'Service update refuses reparse points'}
}
if($source -ieq $target){throw 'Build source must differ from installed executable'}
$manifest=Get-Content -Raw -LiteralPath $manifestFile|ConvertFrom-Json
$owner=[Security.Principal.SecurityIdentifier]::new([string]$manifest.owner_sid)
$service=Get-Service -Name neexitlag -ErrorAction Stop
$scm=Get-CimInstance -ClassName Win32_Service -Filter "Name='neexitlag'" -OperationTimeoutSec 5
if($null -eq $scm){throw 'Installed SCM service not found'}
Assert-ServiceUpdateScmTarget ([string]$scm.PathName) $target
$serviceProcess=[int]$scm.ProcessId
foreach($process in @(Get-Process -Name neexitlag-service -ErrorAction SilentlyContinue)){
 if($process.Path -ieq $target -and $process.Id -ne $serviceProcess){throw 'Close NeExitLag GUI first: its owner-scoped recovery process still uses the installed binary'}
}
$backupRoot=Join-Path $root 'backups'
if(Test-Path -LiteralPath $backupRoot){if((Get-Item -LiteralPath $backupRoot -Force).Attributes -band [IO.FileAttributes]::ReparsePoint){throw 'Backup directory is a reparse point'}}else{New-Item -ItemType Directory -Path $backupRoot|Out-Null}
$backup=Join-Path $backupRoot ('service-'+[DateTime]::UtcNow.ToString('yyyyMMddTHHmmssfffZ')+'-'+[Guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $backup|Out-Null
$oldAcl=Get-Acl -LiteralPath $target
$oldAcl.GetSecurityDescriptorSddlForm([Security.AccessControl.AccessControlSections]::All)|Set-Content -LiteralPath (Join-Path $backup 'binary.sddl') -Encoding ASCII
Copy-Item -LiteralPath $target -Destination (Join-Path $backup 'neexitlag-service.exe')
function Set-BinaryOwnerReadExecute([string]$Path,[string]$OwnerSid){
 $binaryOwner=[Security.Principal.SecurityIdentifier]::new($OwnerSid)
 $acl=[Security.AccessControl.FileSecurity]::new();$acl.SetAccessRuleProtection($true,$false)
 $acl.SetOwner([Security.Principal.SecurityIdentifier]::new('S-1-5-32-544'))
 foreach($sid in @('S-1-5-18','S-1-5-32-544')){[void]$acl.AddAccessRule([Security.AccessControl.FileSystemAccessRule]::new([Security.Principal.SecurityIdentifier]::new($sid),[Security.AccessControl.FileSystemRights]::FullControl,[Security.AccessControl.AccessControlType]::Allow))}
 [void]$acl.AddAccessRule([Security.AccessControl.FileSystemAccessRule]::new($binaryOwner,[Security.AccessControl.FileSystemRights]::ReadAndExecute,[Security.AccessControl.AccessControlType]::Allow))
 Set-Acl -LiteralPath $Path -AclObject $acl
}
$pending=Join-Path $root ('.service-update-'+[Guid]::NewGuid().ToString('N')+'.exe')
try{
 Copy-Item -LiteralPath $source -Destination $pending
 Set-BinaryOwnerReadExecute $pending $owner.Value
 Stop-Service neexitlag
 $service.WaitForStatus('Stopped',[TimeSpan]::FromSeconds(30))
 if($serviceProcess -gt 0){try{$process=[Diagnostics.Process]::GetProcessById($serviceProcess);if(!$process.WaitForExit(30000)){throw 'Stopped service process did not exit; executable not replaced'}}catch [ArgumentException]{} }
 [IO.File]::Replace($pending,$target,$null)
 Set-BinaryOwnerReadExecute $target $owner.Value
 if((Get-FileHash -LiteralPath $target).Hash -ne (Get-FileHash -LiteralPath $source).Hash){throw 'Installed binary hash mismatch'}
 Start-Service neexitlag
 $service.WaitForStatus('Running',[TimeSpan]::FromSeconds(30))
 $deadline=[DateTime]::UtcNow.AddSeconds(30);$ready=$false
 do{
  try{$status=Invoke-ServiceIPC 'status' $null (Join-Path $root 'ipc.token');$ready=$null -ne $status.PSObject.Properties['backend']}catch{$ready=$false}
  if(!$ready -and [DateTime]::UtcNow -lt $deadline){Start-Sleep -Milliseconds 300}
 }while(!$ready -and [DateTime]::UtcNow -lt $deadline)
 if(!$ready){throw 'Updated service IPC readiness failed'}
 Write-Host "Windows service updated. Backup: $backup"
 Write-Host 'Only service executable ACL changed: owner RX, SYSTEM/admins F. Identity, credentials, driver and state ACLs preserved.'
}catch{
 $failure=$_.Exception.Message
 try{
  Stop-Service neexitlag -ErrorAction SilentlyContinue
  $service.WaitForStatus('Stopped',[TimeSpan]::FromSeconds(30))
  Copy-Item -LiteralPath (Join-Path $backup 'neexitlag-service.exe') -Destination $target -Force
  Set-Acl -LiteralPath $target -AclObject $oldAcl
  Start-Service neexitlag
  $service.WaitForStatus('Running',[TimeSpan]::FromSeconds(30))
 }catch{throw "Windows update failed ($failure); automatic binary rollback failed. Backup retained: $backup"}
 throw "Windows update failed; previous binary/ACL restored: $failure. Backup: $backup"
}finally{if(Test-Path -LiteralPath $pending){Remove-Item -LiteralPath $pending -Force}}
