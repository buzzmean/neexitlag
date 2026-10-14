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
