$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot '..\..\tools\windows\simultaneous-checks.ps1')
. (Join-Path $PSScriptRoot '..\..\tools\windows\hot-attach-checks.ps1')
function Check($value,[string]$message){if(!$value){throw $message}}
function Reject([scriptblock]$code){$failed=$false;try{&$code|Out-Null}catch{$failed=$true};if(!$failed){throw 'Invalid HOT ATTACH evidence accepted'}}
$root=[pscustomobject]@{PID=5;CreationTime=10;Path='C:\Apps\chrome.exe';ParentPID=1}
$child=[pscustomobject]@{PID=6;CreationTime=11;Path=$root.Path;ParentPID=5}
$roots=@(Get-HotAttachRoots @($root,$child) @('chrome.exe'));Check ($roots.Count -eq 1 -and $roots[0].PID -eq 5) 'Chrome child mistaken for root'
Assert-HotAttachRoots $roots @($root)
Reject {Assert-HotAttachRoots $roots @($child)}
$reused=[pscustomobject]@{PID=5;CreationTime=99;Path=$root.Path};Reject {Assert-HotAttachRoots $roots @($reused)}
Reject {Get-HotAttachRoots @($root) @('Discord.exe')}
$relay=@{flow_id=1;pid=7;creation_time=20;exe='msedge.exe';hot_attach=@{epoch=2;state='HOT_ATTACH_ROUTE_OPENED'};selected_route='RELAY';policy=@{action='RELAY_REQUIRED'};counters=@{relayed_bytes=10;relay_reply_bytes=20;direct_bytes=0;direct_packets=0}}
$yt=@{flow_id=2;pid=5;creation_time=10;exe=$root.Path;protocol='tcp';domain='www.youtube.com';hot_attach=@{epoch=2;state='HOT_ATTACH_TRANSFORM_APPLIED'};selected_route='DIRECT_PLUS';strategy='p2-youtube-disorder';chain='tcp-multisplit + tcp-disorder';policy=@{source='configured_baseline'};counters=@{direct_plus_packets=1;relayed_bytes=0}}
$state=@{Flows=@{'1'=$relay;'2'=$yt};Lost=$false};$hot=@{epoch=2;state='HOT_ATTACH_NETWORK_READY'}
Check ((Get-HotAttachProof $state $hot 2 @('msedge.exe')).transport_evidence_pass) 'Fresh HOT ATTACH proof rejected'
$relay.hot_attach.state='HOT_ATTACH_RETIRED';Check (!(Get-HotAttachProof $state $hot 2 @('msedge.exe')).transport_evidence_pass) 'Old connection counted as fresh relay'
$relay.hot_attach.state='HOT_ATTACH_ROUTE_OPENED';$relay.counters.direct_bytes=1
Check (!(Get-HotAttachProof $state $hot 2 @('msedge.exe')).transport_evidence_pass) 'Transition DIRECT leak hidden'
$relay.counters.direct_bytes=0;$hot.state='HOT_ATTACH_INCOMPLETE';Reject {Get-HotAttachProof $state $hot 2 @('msedge.exe')}
$hot.state='HOT_ATTACH_NETWORK_READY';Reject {Get-HotAttachProof $state $hot 3 @('msedge.exe')}
$state.Lost=$true;Reject {Get-HotAttachProof $state $hot 2 @('msedge.exe')}
Write-Host 'PASS HOT ATTACH original root/PID-birth, fresh route/transform evidence, transition leak, deadline/generation and loss gates'
