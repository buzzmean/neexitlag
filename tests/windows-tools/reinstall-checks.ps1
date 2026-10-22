$ErrorActionPreference='Stop'
Set-StrictMode -Version Latest
. (Join-Path $PSScriptRoot '../../tools/windows/reinstall-state.ps1')
function Reject-Install([scriptblock]$Body,[string]$Pattern) {
 $message=$null;try{& $Body|Out-Null}catch{$message=$_.Exception.Message}
 if($null -eq $message -or $message -notlike $Pattern){throw "Expected reinstall rejection ($Pattern), got: $message"}
}
$temp=Join-Path ([IO.Path]::GetTempPath()) ('neexitlag reinstall '+[Guid]::NewGuid().ToString('N'))
try {
 New-Item $temp -ItemType Directory|Out-Null
 $installed=Join-Path $temp 'installed';$source=Join-Path $temp 'bundle'
 New-Item $installed,$source -ItemType Directory|Out-Null
 $owner='S-1-5-21-111111111-222222222-333333333-1234'
 $manifest=@{owner_sid=$owner;control_plane_url='https://fixture.invalid:8443';probe_target='203.0.113.1:40000';tcp_probe_target=''}
 [IO.File]::WriteAllText((Join-Path $installed 'service.json'),($manifest|ConvertTo-Json))
 $sources=@{}
 foreach($name in @('identity.json','device.pem','trust.json','ca.pem')) {
  $sources[$name]=Join-Path $source $name
  [IO.File]::WriteAllText($sources[$name],"fixture $name")
  Copy-Item $sources[$name] $installed
 }
 $arguments=@{Directory=$installed;OwnerSid=$owner;ControlPlaneUrl=$manifest.control_plane_url;ProbeTarget=$manifest.probe_target;Sources=$sources}
 Reject-Install {Assert-ServiceInstallState @arguments} '*explicitly use -Reinstall*'
 $before=(Get-FileHash (Join-Path $installed 'identity.json')).Hash
 [void](Assert-ServiceInstallState @arguments -Reinstall)
 Copy-ServiceCredential $sources['identity.json'] (Join-Path $installed 'identity.json')
 if((Get-FileHash (Join-Path $installed 'identity.json')).Hash -ne $before){throw 'Matching credential changed'}
 [IO.File]::WriteAllText($sources['identity.json'],'different identity')
 Reject-Install {Assert-ServiceInstallState @arguments -Reinstall} '*credential/trust mismatch*'
 if((Get-FileHash (Join-Path $installed 'identity.json')).Hash -ne $before){throw 'Credential replaced by mismatched bundle'}
 Copy-Item (Join-Path $installed 'identity.json') $sources['identity.json'] -Force
 $arguments.OwnerSid='different owner';Reject-Install {Assert-ServiceInstallState @arguments -Reinstall} '*owner_sid mismatch*';$arguments.OwnerSid=$owner
 $arguments.ProbeTarget='203.0.113.2:40000';Reject-Install {Assert-ServiceInstallState @arguments -Reinstall} '*probe_target mismatch*';$arguments.ProbeTarget=$manifest.probe_target
 Write-Output 'PASS explicit reinstall preflight, preserved identity and owner/target/credential mismatch rejection'

 if($env:OS -eq 'Windows_NT') {
  # Actual install/uninstall scripts and filesystem/ACL operations; only SCM
  # commands are mocked. No production service or driver is created/stopped.
  $global:neexitlagReinstallFixture=@{serviceCreated=$false;creates=0;deletes=0}
  $global:neexitlagReinstallFixture.mockService=[pscustomobject]@{Status='Running';Disposed=$false}
  $global:neexitlagReinstallFixture.mockService|Add-Member ScriptMethod WaitForStatus {param($status,$timeout) $this.Status=$status}
  $global:neexitlagReinstallFixture.mockService|Add-Member ScriptMethod Dispose {$this.Disposed=$true}
  function Get-Service {param([string]$Name) if($Name -ne 'neexitlag'){throw 'Unexpected SCM name'};if($global:neexitlagReinstallFixture.serviceCreated){return $global:neexitlagReinstallFixture.mockService}}
  function New-Service {param($Name,$BinaryPathName,$DisplayName,$StartupType) if($Name -ne 'neexitlag' -or $DisplayName -ne 'VELIQ Routing Service'){throw 'SCM compatibility/display-name mismatch'};$global:neexitlagReinstallFixture.creates++;$global:neexitlagReinstallFixture.serviceCreated=$true;$global:neexitlagReinstallFixture.mockService.Status='Running'}
  function Start-Service {param($Name) if($Name -ne 'neexitlag'){throw 'Unexpected SCM start'}}
  function Stop-Service {param($Name) if($Name -ne 'neexitlag'){throw 'Unexpected SCM stop'};$global:neexitlagReinstallFixture.mockService.Status='Stopped'}
  function sc.exe {if($args[0] -eq 'delete'){$global:neexitlagReinstallFixture.deletes++;$global:neexitlagReinstallFixture.serviceCreated=$false};$global:LASTEXITCODE=0}
  $target=Join-Path $temp 'actual install';$driver=Join-Path $temp 'driver'
  New-Item $driver -ItemType Directory|Out-Null
  foreach($name in @('WinDivert.dll','WinDivert64.sys')){[IO.File]::WriteAllText((Join-Path $driver $name),"fixture $name")}
  $exe=Join-Path $temp 'service.exe';[IO.File]::WriteAllText($exe,'first executable')
  $params=@{NativeTcpFirewall=$false;ServiceExecutable=$exe;OwnerSid=$owner;DeviceDirectory=$source;TrustDirectory=$source;DriverDirectory=$driver;DllSha256=(Get-FileHash (Join-Path $driver 'WinDivert.dll')).Hash;DriverSha256=(Get-FileHash (Join-Path $driver 'WinDivert64.sys')).Hash;ControlPlaneUrl=$manifest.control_plane_url;ProbeTarget=$manifest.probe_target;InstallDirectory=$target}
  $installScript=Join-Path $PSScriptRoot '../../tools/windows/install.ps1'
  & $installScript @params
  $state=Join-Path $target 'state/profile-fixture.json';[IO.File]::WriteAllText($state,'preserve profile/settings state')
  $protected=@('identity.json','device.pem','trust.json','ca.pem','ipc.token','state/profile-fixture.json')
  $hashes=@{};foreach($name in $protected){$hashes[$name]=(Get-FileHash (Join-Path $target $name)).Hash}
  & (Join-Path $PSScriptRoot '../../tools/windows/uninstall.ps1')
  if($global:neexitlagReinstallFixture.deletes -ne 1 -or !$global:neexitlagReinstallFixture.mockService.Disposed -or !(Test-Path $target)){throw 'Uninstall did not preserve files/release SCM handle'}
  Reject-Install {& $installScript @params} '*explicitly use -Reinstall*'
  [IO.File]::WriteAllText($exe,'upgraded executable')
  & $installScript @params -Reinstall
  if($global:neexitlagReinstallFixture.creates -ne 2 -or [IO.File]::ReadAllText((Join-Path $target 'neexitlag-service.exe')) -ne 'upgraded executable'){throw 'Explicit upgrade did not install new executable'}
  foreach($name in $protected){if((Get-FileHash (Join-Path $target $name)).Hash -ne $hashes[$name]){throw "Reinstall changed preserved $name"}}
  foreach($item in @((Get-Item $target)) + @(Get-ChildItem $target -Recurse -Force)) {
   $acl=Get-Acl $item.FullName
   if(!$acl.AreAccessRulesProtected){throw 'Service path retained inherited access'}
   foreach($rule in $acl.Access) {
    $sid=$rule.IdentityReference.Translate([Security.Principal.SecurityIdentifier]).Value
    if($sid -notin @('S-1-5-18','S-1-5-32-544',$owner)){throw "Unexpected access grant $sid"}
    if($sid -eq $owner) {
     if($item.FullName -ne $target -and $item.FullName -ne (Join-Path $target 'ipc.token') -and $item.FullName -ne (Join-Path $target 'neexitlag-service.exe') -and $item.FullName -ne (Join-Path (Join-Path $target 'state') 'browser-recovery-guard.json')){throw 'GUI owner gained credential/state access'}
     if($rule.InheritanceFlags -ne 'None' -or ($rule.FileSystemRights -band [Security.AccessControl.FileSystemRights]::Write)){throw 'GUI access became writable/inheritable'}
    }
   }
  }
  # Exercise only the updater's pure SCM-target validator and actual executable
  # ACL operation. No production SCM/IPC update is invoked by these filesystem tests.
  $parseErrors=$null;$tokens=$null
  $updateAst=[Management.Automation.Language.Parser]::ParseFile((Join-Path $PSScriptRoot '../../tools/windows/update-service.ps1'),[ref]$tokens,[ref]$parseErrors)
  if($parseErrors.Count){throw 'Updater did not parse'}
  foreach($name in @('Assert-ServiceUpdateScmTarget','Set-BinaryOwnerReadExecute')){
   $definition=$updateAst.FindAll({param($node)$node -is [Management.Automation.Language.FunctionDefinitionAst]},$true)|Where-Object{$_.Name -eq $name}|Select-Object -First 1
   if($null -eq $definition){throw "Missing updater validator: $name"}
   . ([scriptblock]::Create($definition.Extent.Text))
  }
  $binary=Join-Path $target 'neexitlag-service.exe'
  Assert-ServiceUpdateScmTarget ('"'+$binary+'" -config "'+(Join-Path $target 'service.json')+'"') $binary
  Reject-Install {Assert-ServiceUpdateScmTarget '"C:\Wrong\service.exe"' $binary} '*SCM executable does not match*'
  Reject-Install {Assert-ServiceUpdateScmTarget $binary $binary} '*SCM executable does not match*'
  Set-BinaryOwnerReadExecute $binary $owner
  foreach($rule in (Get-Acl -LiteralPath $binary).Access){
   $sid=$rule.IdentityReference.Translate([Security.Principal.SecurityIdentifier]).Value
   if($sid -eq $owner -and (($rule.FileSystemRights -band [Security.AccessControl.FileSystemRights]::Write) -or !($rule.FileSystemRights -band [Security.AccessControl.FileSystemRights]::ExecuteFile))){throw 'Updater owner binary rights must be RX only'}
   if($sid -notin @($owner,'S-1-5-18','S-1-5-32-544')){throw 'Updater broadened executable access'}
  }
  foreach($name in $protected){if((Get-FileHash (Join-Path $target $name)).Hash -ne $hashes[$name]){throw 'Executable ACL update modified preserved credentials/state'}}
  Write-Output 'PASS updater SCM target refusal and owner executable RX without credential/state changes'
  Write-Output 'PASS actual Windows install/uninstall/reinstall with mocked SCM: binary updated, credentials/token/state preserved, restrictive ACLs'
 }
}finally{Remove-Item $temp -Recurse -Force;Remove-Variable neexitlagReinstallFixture -Scope Global -ErrorAction SilentlyContinue}
