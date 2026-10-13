[CmdletBinding()]param()
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'common.ps1')
Assert-Administrator
Stop-Service neexitlag
(Get-Service neexitlag).WaitForStatus('Stopped',[TimeSpan]::FromSeconds(20))
Write-Host 'Service stopped. Verify ordinary networking remains DIRECT.'
