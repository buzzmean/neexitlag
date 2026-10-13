# Pure checks shared by real acceptance and regression fixtures (no network simulation).
function Assert-AcceptanceRoute($Data,$Evidence,[string]$Exe,[string]$Target,[string]$Route,[string]$RelayEgress) {
 if($Evidence.truncated){throw 'Capture evidence truncated; cannot prove exclusive routing'}
 $flows=@($Evidence.flows|Where-Object {$_.pid -eq $Data.pid -and $_.creation_time -eq $Data.creation_time -and $_.exe -ieq $Exe -and $_.destination -eq $Target})
 if(!$flows.Count){throw "${Route}: no capture evidence for this PID, creation time, executable and destination"}
 $counts=@{}
 foreach($key in @('captured_packets','relayed_packets','direct_packets','relay_reply_packets')) {
  $counts[$key]=[double](($flows|ForEach-Object {$_.counters.$key}|Measure-Object -Sum).Sum)
 }
 if($counts.captured_packets -lt $Data.Sent){throw "$Route capture count is below measured sent count"}
 if(@($Data.observed_sources).Count -ne 1){throw "${Route}: expected exactly one authenticated observed source"}
 foreach($flow in $flows) {
  if($flow.selected_route -ne $Route){throw "${Route}: selected_route mismatch"}
  $relay=$flow.PSObject.Properties['relay_id']
  if($Route -eq 'RELAY' -and ($null -eq $relay -or $relay.Value -ne 'VPS-01')){throw 'RELAY: relay_id mismatch'}
  if($Route -eq 'DIRECT' -and $null -ne $relay -and $relay.Value){throw 'DIRECT: unexpected relay_id'}
 }
 if($Route -eq 'RELAY') {
  if($Data.observed_sources[0] -ne $RelayEgress -or $counts.relayed_packets -lt $Data.Sent -or $counts.direct_packets -ne 0 -or $counts.relay_reply_packets -lt $Data.Received){throw 'RELAY: source/counter exclusivity or reply reinjection proof failed'}
 } else {
  if($Data.observed_sources[0] -eq $RelayEgress -or $counts.direct_packets -lt $Data.Sent -or $counts.relayed_packets -ne 0 -or $counts.relay_reply_packets -ne 0){throw 'DIRECT: source/counter exclusivity proof failed'}
 }
 return [pscustomobject]$counts
}
function Assert-AcceptanceLoss($Data) {
 if($Data.Sent -le 0 -or $Data.Received -le 0 -or $Data.Received -gt $Data.Sent){throw 'Invalid sent/received counters or zero authenticated replies'}
 $loss=1.0-[double]$Data.Received/[double]$Data.Sent
 if([Math]::Abs($Data.Loss-$loss) -gt 0.000001 -or ($Data.Sent-$Data.Received)*10 -gt $Data.Sent){throw 'Echo loss exceeds the acceptance limit (10%) or report is inconsistent'}
}
function Get-AcceptanceProfile($Snapshot,[string]$Exe) {
 # Registry uses custom first, then ordinal ID; do not silently trust profile.set.
 $matches=@($Snapshot.profiles|Where-Object {
  $p=$_;@($p.executables|Where-Object { $_.Replace('\','/') -ieq $Exe.Replace('\','/') -or $_ -ieq [IO.Path]::GetFileName($Exe) }).Count -gt 0
 })
 $winner=$null
 foreach($p in $matches){if($null -eq $winner -or ($p.custom -and !$winner.custom) -or ($p.custom -eq $winner.custom -and [string]::CompareOrdinal($p.id,$winner.id) -lt 0)){$winner=$p}}
 return $winner
}
function Test-AcceptanceReady($Snapshot,[string]$Game,[string]$Control,[string]$SelectedId,[string]$ControlId,[string]$Mode) {
 if(!$Snapshot.optimizing -or !$Snapshot.backend.ready){return $false}
 $selected=Get-AcceptanceProfile $Snapshot $Game;$direct=Get-AcceptanceProfile $Snapshot $Control
 if($null -eq $selected -or $null -eq $direct -or $selected.id -ne $SelectedId -or $selected.mode -ne $Mode -or $direct.id -ne $ControlId -or $direct.mode -ne 'DIRECT'){return $false}
 if($Mode -eq 'MANUAL' -and $selected.relay_id -ne 'VPS-01'){return $false}
 return @($Snapshot.routes|Where-Object {$_.id -eq 'VPS-01' -and $_.status -eq 'healthy'}).Count -gt 0
}
function Assert-AcceptanceOverlap($Selected,$Direct) {
 $startA=[DateTimeOffset]::Parse($Selected.measurement_started.ToString());$endA=[DateTimeOffset]::Parse($Selected.measurement_finished.ToString())
 $startB=[DateTimeOffset]::Parse($Direct.measurement_started.ToString());$endB=[DateTimeOffset]::Parse($Direct.measurement_finished.ToString())
 if($startA -ge $endA -or $startB -ge $endB -or $startA -ge $endB -or $startB -ge $endA){throw 'Measured packet-send windows did not overlap or timestamps are invalid'}
}
