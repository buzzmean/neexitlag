#requires -Version 5.1
[CmdletBinding()]
param([string]$Manifest=(Join-Path $env:ProgramData 'neexitlag\service.json'))
$ErrorActionPreference='Stop';Set-StrictMode -Version Latest
. (Join-Path $PSScriptRoot 'common.ps1')
$repo=(Resolve-Path (Join-Path $PSScriptRoot '..\..')).Path
$bin=Join-Path $repo 'bin\relay-fleet\relay-client.exe'
if(!(Test-Path -LiteralPath $bin)){throw 'Run add-relay-vps.ps1 -Inspect -PromotePrimary first to build the signed-cache reader'}
& $bin '--mode' 'status' '--manifest' $Manifest
if($LASTEXITCODE){throw 'SIGNED_CACHE_STATUS_FAILED'}
Invoke-ServiceIPC 'relay.pool.diagnostics' @{relay_id='VPS-01'} | ConvertTo-Json -Depth 8
Write-Host 'Legacy client: active sessions may include old pinned TCP flows. A published address does not prove every established flow moved.'
