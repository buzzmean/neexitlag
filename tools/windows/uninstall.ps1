[CmdletBinding()]param()
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'common.ps1')
Assert-Administrator
$service=Get-Service neexitlag -ErrorAction SilentlyContinue
if($service){if($service.Status -ne 'Stopped'){Stop-Service neexitlag;$service.WaitForStatus('Stopped',[TimeSpan]::FromSeconds(20))};& sc.exe delete neexitlag|Out-Null;if($LASTEXITCODE){throw 'SCM deletion failed'}}
Write-Host 'SCM service removed. Credentials, reports and driver files were preserved for explicit rollback/cleanup.'
