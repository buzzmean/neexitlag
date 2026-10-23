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
$before=Invoke-ServiceIPC 'relay.primary.status'
Start-Sleep -Seconds 1
$after=Invoke-ServiceIPC 'relay.primary.status'
$after | ConvertTo-Json -Depth 8
foreach($row in @($after.connections)){
 $old=@($before.connections | Where-Object {$_.authenticated_relay_id -eq $row.authenticated_relay_id}) | Select-Object -First 1
 $txBefore=0;$rxBefore=0;if($null -ne $old){$txBefore=[long]$old.TX;$rxBefore=[long]$old.RX}
 [pscustomobject]@{authenticated_relay_id=$row.authenticated_relay_id;tcp_session_endpoint=$row.observed_tcp_session_endpoint;active=$row.Active;sent_payload_delta=([long]$row.TX-$txBefore);received_payload_delta=([long]$row.RX-$rxBefore);sample_seconds=1} | Format-List
}
Invoke-ServiceIPC 'relay.pool.diagnostics' @{relay_id='VPS-PRIMARY'} | ConvertTo-Json -Depth 8
Write-Host 'connections.observed_tcp_session_endpoint plus TX/RX proves actual new TCP session usage. Empty counters mean no observed traffic yet; UDP exposes selected identity, not peer socket endpoint. Existing streams stay pinned.'
