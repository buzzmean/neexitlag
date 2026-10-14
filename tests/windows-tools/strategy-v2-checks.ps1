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

function Check($Condition,[string]$Message){if(!$Condition){throw $Message}}
$candidate=[pscustomobject]@{id='fake5';actions=@([pscustomobject]@{kind='fake-tls';ttl=5},[pscustomobject]@{kind='tcp-multisplit'})}
function New-DiagnosticFlow([int]$Port=443){return [pscustomobject]@{pid=42;creation_time=100;exe=$exe;destination="203.0.113.9:$Port";protocol='tcp';strategy='fake5';selected_route='DIRECT';local_inspection=$null;counters=[pscustomobject]@{captured_packets=4;direct_plus_packets=0;direct_plus_bytes=0;fake_tls_packets=0;quic_fallback_packets=0;relayed_bytes=0}}}
$f=New-DiagnosticFlow;$capture=[pscustomobject]@{truncated=$false;flows=@($f)}
function Diagnose([bool]$Valid=$true,[string]$Kind='NETWORK_TIMEOUT'){return Get-V2TrialDiagnostic $capture $ids $exe $obs.destination 'DIRECT_PLUS' $false $Valid $Kind $candidate}
$d=Diagnose;Check ($d.classifications -contains 'NO_TRANSFORM' -and $d.classifications -contains 'NETWORK_TIMEOUT') 'No-transform timeout not distinguished';Check ($d.counters.captured_packets -eq 4) 'Captured counters missing'
foreach($reason in @('CLIENTHELLO_INCOMPLETE','SNI_NOT_FOUND','DOMAIN_NOT_MATCHED')){
 $f.local_inspection=[pscustomobject]@{reason=$reason;domain='www.youtube.com'};$d=Diagnose
 Check ($d.classifications -contains $reason -and $d.classifications -contains 'NO_TRANSFORM') "Missing passive reason $reason"
 Check ($d.flows[0].observed_domain -eq 'www.youtube.com') 'Observed SNI lost'
}
$f.counters.direct_plus_packets=1;$f.counters.direct_plus_bytes=1200;$f.counters.fake_tls_packets=1;$f.selected_route='DIRECT_PLUS'
$d=Diagnose;Check ($d.classification -eq 'TRANSFORM_APPLIED_CONNECTION_FAILED' -and $d.transform_status -eq 'APPLIED' -and $d.classifications -contains 'NETWORK_TIMEOUT') 'Applied-but-failed timeout not distinguished'
Check ($d.counters.fake_tls_packets -eq 1 -and $d.counters.relayed_bytes -eq 0) 'Actual counters lost'
$d=Diagnose $false 'CDP/INSTRUMENTATION_FAILURE';Check ($d.classification -eq 'CDP/INSTRUMENTATION_FAILURE' -and $d.transform_status -eq 'APPLIED' -and $d.classifications -notcontains 'TRANSFORM_APPLIED_CONNECTION_FAILED' -and $d.classifications -notcontains 'NETWORK_TIMEOUT') 'CDP cancellation invented connection failure'
$f=New-DiagnosticFlow;$f.protocol='udp';$f.counters.quic_fallback_packets=3;$capture.flows=@($f)
$d=Diagnose;Check ($d.classifications -contains 'QUIC_FALLBACK_APPLIED' -and $d.transform_status -ne 'APPLIED') 'QUIC suppression incorrectly proves TCP transform'
$capture.truncated=$true;$f.counters.quic_fallback_packets=0;$d=Diagnose;Check ($d.classifications -notcontains 'NO_TRANSFORM' -and $d.transform_status -eq 'UNKNOWN') 'Truncation invented negative proof'
$f.protocol='tcp';$f.counters.direct_plus_packets=1;$f.counters.direct_plus_bytes=100;$d=Diagnose;Check ($d.transform_status -eq 'APPLIED' -and !$d.evidence_complete) 'Positive proof lost in truncated snapshot'
$capture.truncated=$false;$f.creation_time=101;$d=Diagnose;Check ($null -eq $d.counters.captured_packets -and $d.transform_status -eq 'UNKNOWN') 'PID reuse borrowed counters'
$f.creation_time=100;$f=New-DiagnosticFlow 8443;$capture.flows=@($f);$d=Diagnose;Check ($d.flows.Count -eq 0) 'Other target supplied proof'
$d=Get-V2TrialDiagnostic $null @() $exe $obs.destination 'DIRECT_PLUS' $false $false 'CDP/INSTRUMENTATION_FAILURE' $candidate
Check ($null -eq $d.counters.direct_plus_packets -and $d.classifications -notcontains 'NO_TRANSFORM') 'Missing backend invented zero counters'
Write-V2TrialDiagnostic $d

# Execute the production CDP wrapper with a real canceled .NET Task. Classification
# must use exception metadata/inner types, never localized GetResult text.
class V2CancelledSocket {
 [Threading.Tasks.Task] SendAsync([object]$Buffer,[object]$Kind,[bool]$End,[Threading.CancellationToken]$Token){return [Threading.Tasks.Task]::FromCanceled([Threading.CancellationToken]::new($true))}
}
$scriptFile=Join-Path $PSScriptRoot '..\..\tools\windows\strategy-tester-v2.ps1';$tokens=$null;$parseErrors=$null
$tree=[Management.Automation.Language.Parser]::ParseFile((Resolve-Path $scriptFile),[ref]$tokens,[ref]$parseErrors)
Check (!$parseErrors.Count) 'Tester script parse failed'
$cdpAst=$tree.Find({param($a)$a -is [Management.Automation.Language.FunctionDefinitionAst] -and $a.Name -eq 'CDP'},$true)
Invoke-Expression $cdpAst.Extent.Text
$script:socket=[V2CancelledSocket]::new();$script:cdpID=0;$caught=$null
try{CDP 'Browser.getVersion' @{}}catch{$caught=$_.Exception}
Check ($null -ne $caught -and (Get-V2ExceptionKind $caught) -eq 'CDP/INSTRUMENTATION_FAILURE') 'GetResult cancellation not identified by typed exception metadata'
$chain=@(Get-V2ExceptionDetails $caught);Check (@($chain|Where-Object{$_.type -match 'CanceledException'}).Count -gt 0) 'Inner cancellation type missing'
$script:socket=$null
Write-Host 'PASS v2 failure taxonomy: no transform, applied failure, incomplete CH, missing SNI, domain mismatch, QUIC, network timeout, typed CDP cancellation, unknown/truncated/foreign evidence'

# Exercise the actual Trial catch/finally with mocked OS/IPC startup: every
# failed trial must retain diagnostics and write proof even without navigation.
$trialAst=$tree.Find({param($a)$a -is [Management.Automation.Language.FunctionDefinitionAst] -and $a.Name -eq 'Trial'},$true)
Invoke-Expression $trialAst.Extent.Text
$AppPath=$exe;$URL='https://www.youtube.com/';$uri=[Uri]$URL;$TargetIPv4='203.0.113.9';$targetDestination=$obs.destination
$network=[pscustomobject]@{fingerprint='frozen';observed_at=[DateTimeOffset]::UtcNow.ToString('o');vpn=@();conflicts=@()}
$run=[Guid]::NewGuid().ToString('N');$profileID='!fixture';$candidates=@($candidate);$TimeoutSeconds=5;$script:socket=$null
$ReportDirectory=Join-Path ([IO.Path]::GetTempPath()) ('v2-diagnostic-fixture-'+$run);[void](New-Item -ItemType Directory $ReportDirectory)
$fixtureFlow=New-DiagnosticFlow;$fixtureFlow.local_inspection=[pscustomobject]@{reason='CLIENTHELLO_INCOMPLETE';payload_bytes=100;record_bytes=1600}
function Invoke-ServiceIPC($Method,$Payload){switch($Method){
 'local.network' {return $network}
 'profile.set' {return $null}
 'status' {return [pscustomobject]@{}}
 'processes' {return $ids}
 'capture.diagnostics' {return [pscustomobject]@{truncated=$false;flows=@($fixtureFlow)}}
 'local.record' {throw 'An instrumentation failure must not enter network cache'}
 default {throw "Unexpected method $Method"}
}}
function Get-AcceptanceProfile($Status,$Path){return [pscustomobject]@{id=$profileID;mode='DIRECT_PLUS'}}
function Start-Process($FilePath,$ArgumentList,[switch]$PassThru){$p=[pscustomobject]@{HasExited=$true};$p|Add-Member ScriptMethod Dispose {};return $p}
function Wait-AcceptanceReadiness($TimeoutMS,$PollMS,$GetSnapshot,$Evaluate){throw 'Unsupported launcher fixture (not an ISP failure)'}
try{
 $r=Trial 'DIRECT_PLUS' 'fake5' 0
 Check (!$r.success -and $r.instrumentation_failure -and !$r.measurement_recorded -and !$r.measurement_valid -and $null -eq $r.measurement) 'Startup failure was learned as network loss'
 Check ($r.classification -eq 'CDP/INSTRUMENTATION_FAILURE' -and $r.classifications -contains 'CLIENTHELLO_INCOMPLETE') 'Trial lost combined classifications'
 $proof=Get-Content (Join-Path $ReportDirectory "$run-0-proof.json") -Raw|ConvertFrom-Json
 Check ($proof.diagnostic.counters.captured_packets -eq 4 -and $proof.phase -eq 'browser_startup' -and $proof.exception_chain.Count -gt 0) 'Trial catch/finally did not save counters/phase/typed exception'
}finally{Remove-Item $ReportDirectory -Recurse -Force}
Write-Host 'PASS actual Trial catch/finally preserves counters and reason metadata and excludes instrumentation from learning'
$f=New-DiagnosticFlow;$f.local_inspection=[pscustomobject]@{reason='TRANSFORM_SEND_FAILED';output_packets_sent=1;send_failed=$true};$capture=[pscustomobject]@{truncated=$false;flows=@($f)}
$d=Diagnose;Check ($d.transform_status -eq 'PARTIAL_OUTPUT' -and $d.classifications -notcontains 'NO_TRANSFORM') 'Partial sends mislabeled as no transform'
$f.counters.direct_plus_packets=1;$f.counters.direct_plus_bytes=100;$d=Diagnose $true 'APPLICATION_RESPONSE_FAILURE'
Check ($d.transform_status -eq 'APPLIED' -and $d.classifications -notcontains 'TRANSFORM_APPLIED_CONNECTION_FAILED') 'Completed HTTP response mislabeled as connection failure'
Write-Host 'PASS partial output and HTTP/evidence failures do not invent transport failure or successful application'

# Distinct OS refresh cycles, rather than repeated IPC reads of one cache entry.
$state=@{last_refresh='';fingerprint='';cycles=0}
$n=[pscustomobject]@{fingerprint='isp';observed_at=[DateTimeOffset]::UtcNow.AddSeconds(-30).ToString('o');vpn=@();conflicts=@()}
Check (!(Test-V2NetworkCycle $n $state)) 'First snapshot accepted as stable'
Check (!(Test-V2NetworkCycle $n $state)) 'Cached IPC snapshot counted as another OS refresh'
$n.observed_at=[DateTimeOffset]::UtcNow.ToString('o');Check (Test-V2NetworkCycle $n $state) 'Second fresh equal fingerprint rejected'
$baseline=$n.PSObject.Copy();$n.fingerprint='changed';Reject {Assert-V2FrozenNetwork $baseline $n}
$n.fingerprint='isp';$n.vpn=@('HappTunnel');Reject {Assert-V2FrozenNetwork $baseline $n}
$state=@{last_refresh='';fingerprint='';cycles=0};Reject {Test-V2NetworkCycle $n $state}
$n.vpn=@();$n.observed_at=[DateTimeOffset]::UtcNow.AddMinutes(-2).ToString('o');Reject {Test-V2NetworkCycle $n $state}
foreach($missing in @($null,[pscustomobject]@{},[pscustomobject]@{frameTree=[pscustomobject]@{}},[pscustomobject]@{frameTree=[pscustomobject]@{frame=[pscustomobject]@{}}})){Reject {Get-V2MainFrame $missing}}
Check ((Get-V2MainFrame ([pscustomobject]@{frameTree=[pscustomobject]@{frame=[pscustomobject]@{id='main'}}})) -eq 'main') 'Frame identity lost'
foreach($field in @('result','instrumentation_failure','phase','success','mode','strategy')){Check ($null -eq (Get-V2Value ([pscustomobject]@{}) $field)) ('Unsafe optional property '+$field)}
# Exercise cleanup against native stderr/child-exit races under both PS5.1/7.
function taskkill.exe {throw 'The process no longer exists (simulated native stderr)'}
$p=[pscustomobject]@{HasExited=$false;Id=123}
$p|Add-Member ScriptMethod WaitForExit {param($ms) throw 'Already exited'}
$p|Add-Member ScriptMethod Dispose {throw 'Disposed after exit race'}
$script:socket=$null
$issues=@(Stop-V2TrialBrowser $p (Join-Path ([IO.Path]::GetTempPath()) ([Guid]::NewGuid().ToString('N'))))
Check ($issues.Count -eq 3) 'Cleanup race escaped trial or was silently discarded'
Write-Host 'PASS network refresh boundaries, frozen ISP guard, optional CDP/frame/result access and nonfatal cleanup races'
if($env:OS -eq 'Windows_NT'){
 $nativeDir=Join-Path ([IO.Path]::GetTempPath()) ('v2-native-'+[Guid]::NewGuid().ToString('N'));[void](New-Item -ItemType Directory $nativeDir)
 $stdout=Join-Path $nativeDir 'stdout.txt';$stderr=Join-Path $nativeDir 'stderr.txt'
 $native=Microsoft.PowerShell.Management\Start-Process -FilePath $env:ComSpec -ArgumentList '/d /c "echo v2-out & echo v2-err 1>&2"' -RedirectStandardOutput $stdout -RedirectStandardError $stderr -PassThru
 try{
  Check ($native.WaitForExit(10000)) 'Redirected child did not exit';$native.WaitForExit();$native.Refresh()
  Check ($native.ExitCode -eq 0 -and (Get-Content $stdout -Raw) -match 'v2-out' -and (Get-Content $stderr -Raw) -match 'v2-err') 'Real Windows redirected process evidence invalid'
  $issues=@(Stop-V2TrialBrowser $native $nativeDir);Check (!$issues.Count) 'Already exited redirected Windows child cleanup failed'
 }finally{if(Test-Path $nativeDir){Remove-Item $nativeDir -Recurse -Force}}
 Write-Host 'PASS real redirected Windows Start-Process and already exited process cleanup'
}
class V2OptionalReplySocket {
 [Threading.Tasks.Task] SendAsync([object]$Buffer,[object]$Kind,[bool]$End,[Threading.CancellationToken]$Token){return [Threading.Tasks.Task]::CompletedTask}
 [Threading.Tasks.Task[Net.WebSockets.WebSocketReceiveResult]] ReceiveAsync([object]$Buffer,[Threading.CancellationToken]$Token){
  $bytes=[Text.Encoding]::UTF8.GetBytes('{"id":1}')
  [Array]::Copy($bytes,0,$Buffer.Array,$Buffer.Offset,$bytes.Length)
  $done=[Threading.Tasks.TaskCompletionSource[Net.WebSockets.WebSocketReceiveResult]]::new()
  $done.SetResult([Net.WebSockets.WebSocketReceiveResult]::new($bytes.Length,[Net.WebSockets.WebSocketMessageType]::Text,$true))
  return $done.Task
 }
}
$script:socket=[V2OptionalReplySocket]::new();$script:cdpID=0;$script:navID=-2
try{$optional=CDP 'Browser.getVersion' @{};Check ($null -eq $optional) 'Optional CDP result invented data'}finally{$script:socket=$null}
Write-Host 'PASS actual CDP roundtrip with absent result under StrictMode Latest'
