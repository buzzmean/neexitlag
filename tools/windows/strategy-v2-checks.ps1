Set-StrictMode -Version Latest
function Assert-V2BrowserEvidence($Observation,$Capture,$Identities,[string]$Exe,[string]$Route,[string]$Strategy){
 if(!$Observation.success -or $Observation.status -lt 200 -or $Observation.status -ge 400 -or !$Observation.tls){throw 'No completed certificate-validated HTTPS response'}
 if($Capture.truncated){throw 'Target capture truncated: cannot prove a complete route'}
 if($Observation.protocol -notin @('h2','http/1.1','http/1.0')){throw 'TCP strategy requires a real TCP HTTPS response, not HTTP/3 inference'}
 $flows=@($Capture.flows|Where-Object{
  $f=$_
  $f.exe -ieq $Exe -and $f.destination -eq $Observation.destination -and @($Identities|Where-Object{$_.PID -eq $f.pid -and $_.CreationTime -eq $f.creation_time}).Count -gt 0
 })
 if(!$flows.Count){throw 'No exact PID/creation/executable/destination evidence'}
 $valid=@($flows|Where-Object{
  if($_.selected_route -ne $Route){return $false}
  if($Route -eq 'DIRECT_PLUS'){return $_.strategy -eq $Strategy -and $_.counters.direct_plus_packets -gt 0 -and $_.counters.direct_plus_bytes -gt 0 -and $_.counters.relayed_packets -eq 0 -and $_.counters.relayed_bytes -eq 0}
  if($Route -eq 'RELAY'){return $_.relay_id -eq 'VPS-01' -and $_.counters.relayed_bytes -gt 0 -and $_.counters.relay_reply_bytes -gt 0}
  return $_.counters.direct_packets -gt 0 -and $_.counters.relayed_packets -eq 0 -and $_.counters.direct_plus_packets -eq 0
 })
 if(!$valid.Count){throw 'Completed HTTP response has no matching actual route/transform counters'}
 if(@($flows|Where-Object{$_.selected_route -ne $Route}).Count){throw 'Target used mixed routes in a fresh trial'}
 return $valid[0]
}

# Diagnostics are independent of PASS validation. Missing evidence is UNKNOWN,
# not zero; positive counters survive truncated snapshots, negative claims do not.
function Get-V2Value($Object,[string]$Name,$Default=$null){
 if($null -eq $Object){return $Default}
 if($Object -is [Collections.IDictionary]){if($Object.Contains($Name)){return $Object[$Name]};return $Default}
 $p=$Object.PSObject.Properties[$Name];if($null -ne $p){return $p.Value};return $Default
}
function Get-V2ExceptionDetails([Exception]$ErrorObject){
 $out=@();$e=$ErrorObject
 for($i=0;$null -ne $e -and $i -lt 8;$i++){$out+=@{type=$e.GetType().FullName;message=$e.Message};$e=$e.InnerException}
 return $out
}
function Get-V2TrialDiagnostic($Capture,$Identities,[string]$Exe,[string]$Destination,[string]$Mode,[bool]$Success,[bool]$MeasurementValid,[string]$FailureKind,$Candidate){
 $flows=@(Get-V2Value $Capture 'flows' @()|Where-Object{
  $f=$_
  (Get-V2Value $f 'exe' '') -ieq $Exe -and (Get-V2Value $f 'destination' '') -eq $Destination -and @($Identities|Where-Object{(Get-V2Value $_ 'PID' -1) -eq (Get-V2Value $f 'pid' -2) -and (Get-V2Value $_ 'CreationTime' -1) -eq (Get-V2Value $f 'creation_time' -2)}).Count -gt 0
 })
 $complete=$null -ne $Capture -and !(Get-V2Value $Capture 'truncated' $true) -and $flows.Count -gt 0
 $totals=@{};$details=@();$tags=@();$applied=$false;$partial=$false;$quic=$false
 $names=@('captured_packets','direct_plus_packets','direct_plus_bytes','fake_tls_packets','quic_fallback_packets','relayed_bytes')
 foreach($n in $names){$totals[$n]=0L}
 foreach($f in $flows){
  $c=Get-V2Value $f 'counters';$row=@{}
  foreach($n in $names){$v=Get-V2Value $c $n;$row[$n]=$v;if($null -eq $v){$totals[$n]=$null;$complete=$false}elseif($null -ne $totals[$n]){$totals[$n]+=[long]$v}}
  $inspection=Get-V2Value $f 'local_inspection';$why=[string](Get-V2Value $inspection 'reason' '')
  if($why -eq 'TRANSFORM_SEND_FAILED'){$tags+=$why;if((Get-V2Value $inspection 'output_packets_sent' 0) -gt 0){$partial=$true}}
  if($why -in @('CLIENTHELLO_INCOMPLETE','SNI_NOT_FOUND','DOMAIN_NOT_MATCHED','REASSEMBLY_STARTED','REASSEMBLY_WAITING','REASSEMBLY_COMPLETE','REASSEMBLY_TIMEOUT','REASSEMBLY_LIMIT','REASSEMBLY_OUT_OF_ORDER')){$tags+=$why}
  $strategy=Get-V2Value $f 'strategy' ''; $candidateID=Get-V2Value $Candidate 'id' ''
  if((Get-V2Value $f 'protocol' '') -eq 'tcp' -and $strategy -eq $candidateID -and $candidateID -ne '' -and $null -ne $row.direct_plus_packets -and $null -ne $row.direct_plus_bytes -and $row.direct_plus_packets -gt 0 -and $row.direct_plus_bytes -gt 0){$applied=$true}
  if($null -ne $row.quic_fallback_packets -and $row.quic_fallback_packets -gt 0){$quic=$true}
  $details+=@{pid=(Get-V2Value $f 'pid');creation_time=(Get-V2Value $f 'creation_time');exe=(Get-V2Value $f 'exe');destination=(Get-V2Value $f 'destination');protocol=(Get-V2Value $f 'protocol');selected_route=(Get-V2Value $f 'selected_route');strategy=$strategy;observed_domain=(Get-V2Value $f 'domain' (Get-V2Value $inspection 'domain'));chain=(Get-V2Value $f 'chain');fake_ttl=(Get-V2Value $f 'fake_ttl');counters=$row;inspection=$inspection}
 }
 if(!$flows.Count){foreach($n in $names){$totals[$n]=$null};$tags+='NO_CAPTURE_EVIDENCE'}
 $connectionFailed=(!$Success -and $MeasurementValid -and $FailureKind -in @('NETWORK_TIMEOUT','BROWSER_NETWORK_FAILURE'))
 $state='UNKNOWN'
 if($applied){$state='APPLIED';if($connectionFailed){$tags+='TRANSFORM_APPLIED_CONNECTION_FAILED'}}
 elseif($partial){$state='PARTIAL_OUTPUT'}
 elseif($Mode -eq 'DIRECT_PLUS' -and $complete -and $totals.captured_packets -gt 0 -and $totals.direct_plus_packets -eq 0){$state='NOT_OBSERVED';$tags+='NO_TRANSFORM'}
 if($quic){$tags+='QUIC_FALLBACK_APPLIED'}
 if(!$Success){if($FailureKind){$tags+=$FailureKind};if(!$MeasurementValid){$tags+='CDP/INSTRUMENTATION_FAILURE'}}
 if($Success){$primary='PASS'}elseif(!$MeasurementValid){$primary='CDP/INSTRUMENTATION_FAILURE'}elseif($applied -and $connectionFailed){$primary='TRANSFORM_APPLIED_CONNECTION_FAILED'}elseif($tags.Count){$primary=$tags[0]}else{$primary='CONNECTION_FAILED'}
 $configuredChain=@(Get-V2Value $Candidate 'actions' @()|ForEach-Object{Get-V2Value $_ 'kind' ''}) -join ' + '
 $configuredTTL=@(Get-V2Value $Candidate 'actions' @()|Where-Object{(Get-V2Value $_ 'kind' '') -eq 'fake-tls'}|ForEach-Object{Get-V2Value $_ 'ttl'})
 return @{classification=$primary;classifications=@($tags|Select-Object -Unique);transform_status=$state;evidence_complete=[bool]$complete;counters=$totals;flows=$details;configured_chain=$configuredChain;configured_fake_ttl=$configuredTTL}
}
function Write-V2TrialDiagnostic($Diagnostic){
 Write-Host ('  classification={0}; tags={1}; TCP transform={2}; complete={3}' -f $Diagnostic.classification,($Diagnostic.classifications -join ','),$Diagnostic.transform_status,$Diagnostic.evidence_complete)
 $pairs=@();foreach($n in @('captured_packets','direct_plus_packets','direct_plus_bytes','fake_tls_packets','quic_fallback_packets','relayed_bytes')){$v=$Diagnostic.counters[$n];if($null -eq $v){$v='UNKNOWN'};$pairs+="$n=$v"};Write-Host ('  '+($pairs -join '; '))
 foreach($f in $Diagnostic.flows){$display=@{};foreach($field in @('observed_domain','chain','fake_ttl')){$display[$field]=$f[$field];if($null -eq $display[$field] -or [string]$display[$field] -eq ''){$display[$field]='NOT_OBSERVED'}};Write-Host ('  PID={0}; protocol={1}; route={2}; SNI/domain={3}; applied_chain={4}; applied_fake_TTL={5}; inspection={6}' -f $f.pid,$f.protocol,$f.selected_route,$display.observed_domain,$display.chain,$display.fake_ttl,(Get-V2Value $f.inspection 'reason' 'UNAVAILABLE'));Write-Host ('    hello_payload={0}; record={1}; handshake={2}; plan_sent={3}/{4}; detail={5}' -f (Get-V2Value $f.inspection 'payload_bytes' 'UNKNOWN'),(Get-V2Value $f.inspection 'record_bytes' 'UNKNOWN'),(Get-V2Value $f.inspection 'handshake_bytes' 'UNKNOWN'),(Get-V2Value $f.inspection 'output_packets_sent' 'UNKNOWN'),(Get-V2Value $f.inspection 'output_packets_planned' 'UNKNOWN'),(Get-V2Value $f.inspection 'detail' 'UNAVAILABLE'))}
 foreach($f in $Diagnostic.flows){$i=$f.inspection;if((Get-V2Value $i 'segments_seen' 0) -gt 0){Write-Host ('    segments_seen={0}; segments_buffered={1}; bytes_buffered={2}; expected_clienthello_bytes={3}; reassembly_ms={4}; reassembled_sni={5}; transform_applied={6}' -f (Get-V2Value $i 'segments_seen'),(Get-V2Value $i 'segments_buffered'),(Get-V2Value $i 'bytes_buffered'),(Get-V2Value $i 'expected_clienthello_bytes'),(Get-V2Value $i 'reassembly_ms'),(Get-V2Value $i 'reassembled_sni'),(Get-V2Value $i 'transform_applied')) }}
 Write-Host ('  configured_chain={0}; configured_fake_TTL={1} (configuration is not application evidence)' -f $Diagnostic.configured_chain,($Diagnostic.configured_fake_ttl -join ','))
}
function Get-V2ExceptionKind([Exception]$ErrorObject){
 $e=$ErrorObject
 for($i=0;$null -ne $e -and $i -lt 8;$i++){if($e.Data.Contains('v2_failure_kind')){return [string]$e.Data['v2_failure_kind']};$e=$e.InnerException}
 return ''
}

# observed_at is assigned only when InspectNetwork finishes a fresh OS scan.
# Re-reading local.network's cached snapshot never advances the stable count.
function Test-V2NetworkCycle($Snapshot,$State){
 $vpn=@(Get-V2Value $Snapshot 'vpn' @());if($vpn.Count){throw ('VPN present in network snapshot: '+($vpn -join ', ')+'. Turn Happ OFF and wait for a fresh service scan (up to 30 seconds).')}
 if((Get-V2Value $Snapshot 'error' '') -or !(Get-V2Value $Snapshot 'fingerprint' '') -or @(Get-V2Value $Snapshot 'conflicts' @()).Count){throw 'Valid network fingerprint without capture conflicts required'}
 $stamp=[string](Get-V2Value $Snapshot 'observed_at' '')
 $parsed=[DateTimeOffset]::MinValue
 if(![DateTimeOffset]::TryParse($stamp,[ref]$parsed) -or ([DateTimeOffset]::UtcNow-$parsed).TotalSeconds -gt 45 -or $parsed -gt [DateTimeOffset]::UtcNow.AddSeconds(5)){throw 'Fresh network refresh timestamp required'}
 if($stamp -ne $State.last_refresh){
  if($State.last_refresh -and $parsed -le [DateTimeOffset]::Parse($State.last_refresh)){throw 'Network refresh timestamp moved backwards'}
  if($State.fingerprint -eq $Snapshot.fingerprint){$State.cycles++}else{$State.fingerprint=$Snapshot.fingerprint;$State.cycles=1}
  $State.last_refresh=$stamp
 }
 return ($State.cycles -ge 2)
}
function Wait-V2StableNetwork([int]$TimeoutSeconds=100,[scriptblock]$GetSnapshot={Invoke-ServiceIPC 'local.network'}){
 $state=@{last_refresh='';fingerprint='';cycles=0};$watch=[Diagnostics.Stopwatch]::StartNew()
 while($true){$n=&$GetSnapshot;if(Test-V2NetworkCycle $n $state){return $n};if($watch.Elapsed.TotalSeconds -ge $TimeoutSeconds){throw 'Network preflight timeout: two equal fingerprints from different fresh OS refresh cycles required'};Start-Sleep -Milliseconds 250}
}
function Assert-V2FrozenNetwork($Frozen,$Current){
 $scratch=@{last_refresh='';fingerprint='';cycles=0};[void](Test-V2NetworkCycle $Current $scratch)
 if((Get-V2Value $Frozen 'fingerprint' '') -ne (Get-V2Value $Current 'fingerprint' '')){throw 'Network changed during test; discard trial measurements and restart preflight'}
}
function Get-V2MainFrame($Response){
 $id=[string](Get-V2Value (Get-V2Value (Get-V2Value $Response 'frameTree') 'frame') 'id' '')
 if(!$id){throw 'CDP Page.getFrameTree returned no main-frame identity; instrumentation unavailable'}
 return $id
}
function Stop-V2TrialBrowser($Browser,[string]$Data){
 # Cleanup is best-effort and never overwrites a trial result. taskkill can
 # race Chrome's normal exit; PS5.1 turns native stderr into terminating errors.
 $issues=@()
 if($script:socket){try{$script:socket.Dispose()}catch{$issues+=$_.Exception.Message};$script:socket=$null}
 if($Browser){
  try{if(!$Browser.HasExited){& taskkill.exe /PID $Browser.Id /T /F 2>&1|Out-Null}}catch{$issues+=$_.Exception.Message}
  try{if(!$Browser.HasExited){[void]$Browser.WaitForExit(3000)}}catch{$issues+=$_.Exception.Message}
  try{$Browser.Dispose()}catch{$issues+=$_.Exception.Message}
 }
 try{if(Test-Path $Data){Remove-Item $Data -Recurse -Force -ErrorAction Stop}}catch{$issues+='Temporary browser profile remains; delete privately, never share: '+$Data}
 return $issues
}
