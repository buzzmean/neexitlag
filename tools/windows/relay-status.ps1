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
Invoke-ServiceIPC 'relay.primary.status' | ConvertTo-Json -Depth 8
Invoke-ServiceIPC 'relay.pool.diagnostics' @{relay_id='VPS-PRIMARY'} | ConvertTo-Json -Depth 8
Write-Host 'connections.observed_tcp_session_endpoint plus TX/RX proves actual new TCP session usage. Empty counters mean no observed traffic yet; UDP exposes selected identity, not peer socket endpoint. Existing streams stay pinned.'
