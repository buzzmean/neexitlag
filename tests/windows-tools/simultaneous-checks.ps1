$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot '..\..\tools\windows\simultaneous-checks.ps1')
function Check($Value,[string]$Message){if(!$Value){throw $Message}}
function Reject([scriptblock]$Code){$rejected=$false;try{&$Code|Out-Null}catch{$rejected=$true};if(!$rejected){throw 'Invalid simultaneous evidence accepted'}}
$state=@{Cursor=[pscustomobject]@{ledger_id='test';after_update=0};Flows=@{};Lost=$false}
$f=[pscustomobject]@{flow_id=1;pid=7;creation_time=10;exe='C:\Apps\msedge.exe';selected_route='RELAY';policy=@{action='RELAY_REQUIRED'};counters=@{relayed_bytes=10;relay_reply_bytes=20;direct_bytes=0;direct_packets=0}}
$p=[pscustomobject]@{cursor=[pscustomobject]@{ledger_id='test';after_update=1};more=$false;evidence_lost=$false;flows=@($f)}
[void](Merge-SimPage $state $p)
$p.cursor.after_update=2;[void](Merge-SimPage $state $p)
Check ($state.Flows.Count -eq 1) 'Repeated cumulative flow duplicated'
$proof=Get-SimProof $state @('msedge.exe');Check (!$proof.transport_evidence_pass) 'No YouTube evidence passed'
$f.counters.direct_bytes=1;$p.cursor.after_update=3;[void](Merge-SimPage $state $p)
Check ((Get-SimProof $state @('msedge.exe')).direct_leaks -eq 1) 'Late existing-flow leak hidden'
$f.counters.direct_bytes=0
$yt=[pscustomobject]@{flow_id=2;pid=8;creation_time=11;exe='C:\Chrome\chrome.exe';domain='www.youtube.com';selected_route='DIRECT_PLUS';strategy='disorder';policy=@{source='configured_baseline'};counters=@{direct_plus_packets=1;relayed_bytes=0}}
$p.flows=@($yt);$p.cursor.after_update=4;[void](Merge-SimPage $state $p)
Check ((Get-SimProof $state @('msedge.exe')).transport_evidence_pass) 'Valid transport proof failed'
Check ((Get-SimProof $state @('msedge.exe')).application_acceptance -eq 'OPERATOR_OBSERVATION_REQUIRED') 'Application success invented'
$f.selected_route='BLOCK';Check (!(Get-SimProof $state @('msedge.exe')).transport_evidence_pass) 'Protected failure hidden';$f.selected_route='RELAY'
$f|Add-Member -NotePropertyName relay_open -NotePropertyValue @{failure_reason='stream_capacity'}
Check (!(Get-SimProof $state @('msedge.exe')).transport_evidence_pass) 'Open failure hidden';$f.relay_open.failure_reason=''
$p.evidence_lost=$true;Reject {Merge-SimPage $state $p};Reject {Get-SimProof $state @('msedge.exe')}
$p.evidence_lost=$false;$state.Lost=$false;$p.cursor.ledger_id='other';Reject {Merge-SimPage $state $p}
Write-Host 'PASS simultaneous cursor, cumulative replacement, late leaks, loss/restart, protected failures, no invented app success'
