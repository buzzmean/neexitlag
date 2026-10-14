$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot '..\..\tools\windows\strategy-v2-checks.ps1')
function Reject([scriptblock]$Code){$failed=$false;try{&$Code|Out-Null}catch{$failed=$true};if(!$failed){throw 'Invalid v2 proof accepted'}}
$exe='C:\Apps\chrome.exe';$obs=[pscustomobject]@{success=$true;status=200;tls=$true;protocol='h2';destination='203.0.113.9:443'}
$ids=@([pscustomobject]@{PID=42;CreationTime=100})
$flow=[pscustomobject]@{pid=42;creation_time=100;exe=$exe;destination=$obs.destination;strategy='fake5';selected_route='DIRECT_PLUS';counters=[pscustomobject]@{direct_plus_packets=1;direct_plus_bytes=1200;relayed_packets=0;relayed_bytes=0}}
$ev=[pscustomobject]@{truncated=$false;flows=@($flow)}
[void](Assert-V2BrowserEvidence $obs $ev $ids $exe 'DIRECT_PLUS' 'fake5')
$obs.success=$false;Reject {Assert-V2BrowserEvidence $obs $ev $ids $exe 'DIRECT_PLUS' 'fake5'};$obs.success=$true
$obs.status=403;Reject {Assert-V2BrowserEvidence $obs $ev $ids $exe 'DIRECT_PLUS' 'fake5'};$obs.status=200
$obs.tls=$false;Reject {Assert-V2BrowserEvidence $obs $ev $ids $exe 'DIRECT_PLUS' 'fake5'};$obs.tls=$true
$obs.protocol='h3';Reject {Assert-V2BrowserEvidence $obs $ev $ids $exe 'DIRECT_PLUS' 'fake5'};$obs.protocol='h2'
$flow.counters.direct_plus_packets=0;Reject {Assert-V2BrowserEvidence $obs $ev $ids $exe 'DIRECT_PLUS' 'fake5'};$flow.counters.direct_plus_packets=1
$flow.counters.relayed_bytes=1;Reject {Assert-V2BrowserEvidence $obs $ev $ids $exe 'DIRECT_PLUS' 'fake5'};$flow.counters.relayed_bytes=0
$flow.creation_time=101;Reject {Assert-V2BrowserEvidence $obs $ev $ids $exe 'DIRECT_PLUS' 'fake5'};$flow.creation_time=100
$ev.truncated=$true;Reject {Assert-V2BrowserEvidence $obs $ev $ids $exe 'DIRECT_PLUS' 'fake5'}
Write-Host 'PASS v2: complete real HTTPS + TCP + exact identity + actual transform + zero relay; no sent-packet PASS'
