$ErrorActionPreference='Stop'
Set-StrictMode -Version Latest
. (Join-Path $PSScriptRoot '../../tools/windows/acceptance-checks.ps1')
. (Join-Path $PSScriptRoot '../../tools/windows/common.ps1')
function Reject([scriptblock]$Body){$failed=$false;try{& $Body|Out-Null}catch{$failed=$true};if(!$failed){throw 'Invalid fixture was accepted'}}
$data=[pscustomobject]@{pid=10;creation_time=20;Sent=30;Received=27;Loss=0.1;observed_sources=@('203.0.113.1')}
$flow=[pscustomobject]@{pid=10;creation_time=20;exe='C:\test-game.exe';destination='203.0.113.1:40000';selected_route='RELAY';relay_id='VPS-01';counters=[pscustomobject]@{captured_packets=30;relayed_packets=30;direct_packets=0;relay_reply_packets=27}}
$evidence=[pscustomobject]@{truncated=$false;flows=@($flow)}
Assert-AcceptanceLoss $data
Assert-AcceptanceRoute $data $evidence $flow.exe $flow.destination 'RELAY' '203.0.113.1'|Out-Null
foreach($counter in @('captured_packets','relayed_packets','relay_reply_packets')) {
 $old=$flow.counters.$counter;$flow.counters.$counter=0
 Reject {Assert-AcceptanceRoute $data $evidence $flow.exe $flow.destination 'RELAY' '203.0.113.1'}
 $flow.counters.$counter=$old
}
$flow.counters.direct_packets=1
Reject {Assert-AcceptanceRoute $data $evidence $flow.exe $flow.destination 'RELAY' '203.0.113.1'}
$flow.counters.direct_packets=0
$flow.relay_id='wrong';Reject {Assert-AcceptanceRoute $data $evidence $flow.exe $flow.destination 'RELAY' '203.0.113.1'};$flow.relay_id='VPS-01'
$flow.selected_route='DIRECT';Reject {Assert-AcceptanceRoute $data $evidence $flow.exe $flow.destination 'RELAY' '203.0.113.1'}
$flow.relay_id='';$flow.counters.relayed_packets=0;$flow.counters.direct_packets=30;$flow.counters.relay_reply_packets=0;$data.observed_sources=@('198.51.100.1')
Assert-AcceptanceRoute $data $evidence $flow.exe $flow.destination 'DIRECT' '203.0.113.1'|Out-Null
$flow.counters.relay_reply_packets=1;Reject {Assert-AcceptanceRoute $data $evidence $flow.exe $flow.destination 'DIRECT' '203.0.113.1'};$flow.counters.relay_reply_packets=0
$evidence.truncated=$true;Reject {Assert-AcceptanceRoute $data $evidence $flow.exe $flow.destination 'DIRECT' '203.0.113.1'};$evidence.truncated=$false
$flow.creation_time=21;Reject {Assert-AcceptanceRoute $data $evidence $flow.exe $flow.destination 'DIRECT' '203.0.113.1'};$flow.creation_time=20
$data.Received=26;$data.Loss=1-26/30;Reject {Assert-AcceptanceLoss $data}
$data.Loss=0;Reject {Assert-AcceptanceLoss $data}
$selected=[pscustomobject]@{id='!selected';mode='MANUAL';relay_id='VPS-01';custom=$true;executables=@('C:\test-game.exe')}
$control=[pscustomobject]@{id='!control';mode='DIRECT';custom=$true;executables=@('C:\control-client.exe')}
$snapshot=[pscustomobject]@{optimizing=$true;backend=@{ready=$true};profiles=@($selected,$control);routes=@(@{id='VPS-01';status='healthy'})}
if(!(Test-AcceptanceReady $snapshot 'C:\test-game.exe' 'C:\control-client.exe' '!selected' '!control' 'MANUAL')){throw 'Valid readiness rejected'}
$snapshot.backend.ready=$false
if(Test-AcceptanceReady $snapshot 'C:\test-game.exe' 'C:\control-client.exe' '!selected' '!control' 'MANUAL'){throw 'Capture startup was ignored'}
$snapshot.backend.ready=$true;$snapshot.profiles+=@([pscustomobject]@{id='!!competing';mode='DIRECT';custom=$true;executables=@('test-game.exe')})
# Windows basename semantics; portable fixture uses absolute executable below.
$snapshot.profiles[-1].executables=@('C:\test-game.exe')
if(Test-AcceptanceReady $snapshot 'C:\test-game.exe' 'C:\control-client.exe' '!selected' '!control' 'MANUAL'){throw 'Competing profile was ignored'}
$temp=Join-Path ([IO.Path]::GetTempPath()) ('neexitlag-lf-'+[Guid]::NewGuid().ToString('N'))
try {
 New-Item $temp -ItemType Directory|Out-Null
 foreach($name in @('install.sh','database-sql.py','test.service')){[IO.File]::WriteAllText((Join-Path $temp $name),"line1`r`nline2`r`n")}
 [IO.File]::WriteAllBytes((Join-Path $temp 'binary'),[byte[]]@(1,13,10,255))
 Convert-DeploymentToLF $temp
 foreach($name in @('install.sh','database-sql.py','test.service')){if([IO.File]::ReadAllText((Join-Path $temp $name)).Contains("`r")){throw 'Staged CRLF not normalized'}}
 if([Convert]::ToHexString([IO.File]::ReadAllBytes((Join-Path $temp 'binary'))) -ne '010D0AFF'){throw 'Binary modified by normalization'}
}finally{Remove-Item $temp -Recurse -Force}
Write-Output 'PASS acceptance route/loss/readiness rejection fixtures and staged CRLF regression (not Windows E2E)'

# Exercise the actual orchestration functions without executing Windows/VPS actions.
$tokens=$null;$parseErrors=$null
$ast=[Management.Automation.Language.Parser]::ParseFile((Join-Path $PSScriptRoot '../../tools/windows/acceptance.ps1'),[ref]$tokens,[ref]$parseErrors)
foreach($name in @('Stage','Ensure','Save-UTF8','Complete','Evidence','Save-Diagnostics','Wait-Relay')) {
 $definition=$ast.Find({param($node) $node -is [Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq $name},$true)
 Invoke-Expression $definition.Extent.Text
}
$run='fixture';$game='C:\test-game.exe';$control='C:\control-client.exe';$selectedId='!selected';$controlId='!control';$profileMode='MANUAL'
$snapshot.profiles=@($selected,$snapshot.profiles[1]);$snapshot.backend.ready=$true
$script:polls=0
function Invoke-ServiceIPC([string]$Method,$Payload) {
 if($Method -eq 'status') {
  $script:polls++
  if($script:polls -eq 1){throw 'fixture transient IPC startup'}
  $script:snapshot.backend.ready=($script:polls -ge 3)
  return $script:snapshot
 }
 if($Method -eq 'capture.diagnostics'){return $evidence}
 throw 'Unexpected fixture method'
}
$ready=Wait-Relay
if($polls -ne 5 -or !$ready.backend.ready){throw 'Readiness did not retry startup or require stable snapshots'}
$temp=Join-Path ([IO.Path]::GetTempPath()) ('neexitlag-artifacts-'+[Guid]::NewGuid().ToString('N'))
try {
 New-Item $temp -ItemType Directory|Out-Null
 $ReportDirectory=$temp;$results=[Collections.Generic.List[object]]::new();$children=[Collections.Generic.List[object]]::new()
 $path=Join-Path $temp 'failed.json'
 Save-UTF8 $path '{"error":"no authenticated UDP responses","Sent":30,"Received":0}'
 Save-UTF8 ($path+'.stdout') 'fixture client stdout'
 Save-UTF8 ($path+'.stderr') 'fixture client stderr'
 $process=[pscustomobject]@{Id=10;ExitCode=1}
 $process|Add-Member -MemberType ScriptMethod -Name WaitForExit -Value {param($timeout) return $true}
 $process|Add-Member -MemberType ScriptMethod -Name Refresh -Value {}
 $client=[pscustomobject]@{process=$process;report=$path;exe=$game;timeout_ms=1000};$children.Add($client)
 if(Stage 'RELAY zero replies fixture' {Complete $client}){throw 'Zero-reply client accepted'}
 $failure=$results[0]
 if($failure.error -notlike '*no authenticated UDP responses*'){throw 'Wrong client failure reason'}
 $diagnostic=Get-Content $failure.diagnostics -Raw|ConvertFrom-Json
 if($null -eq $diagnostic.status -or $null -eq $diagnostic.clients[0].capture -or !(Test-Path $diagnostic.clients[0].stdout) -or !(Test-Path $diagnostic.clients[0].stderr)){throw 'Failure artifacts were not retained'}
 # An unavailable service must still retain client paths and diagnostic errors.
 function Invoke-ServiceIPC {throw 'fixture service unavailable'}
 $diagnostic=Get-Content (Save-Diagnostics 'unavailable') -Raw|ConvertFrom-Json
 if($diagnostic.clients.Count -ne 1 -or $diagnostic.diagnostic_errors.Count -ne 2 -or !(Test-Path $diagnostic.clients[0].report)){throw 'IPC failure discarded client artifacts'}
}finally{Remove-Item $temp -Recurse -Force}
Write-Output 'PASS actual acceptance startup polling and zero-reply/unavailable-service artifact retention fixtures'

$first=@{measurement_started='2026-10-13T01:00:00Z';measurement_finished='2026-10-13T01:00:20Z'}
$second=@{measurement_started='2026-10-13T01:00:01Z';measurement_finished='2026-10-13T01:00:21Z'}
Assert-AcceptanceOverlap $first $second
$second.measurement_started='2026-10-13T01:00:20Z';Reject {Assert-AcceptanceOverlap $first $second}
$second.measurement_finished='2026-10-13T01:00:19Z';Reject {Assert-AcceptanceOverlap $first $second}
Write-Output 'PASS measured send-window overlap checks (process lifetime alone is insufficient)'
