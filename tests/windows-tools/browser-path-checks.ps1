$ErrorActionPreference='Stop'
Set-StrictMode -Version Latest
$path=Join-Path $PSScriptRoot '..\..\tools\windows\diag-no-vpn.ps1'
$tokens=$null;$errors=$null
$ast=[System.Management.Automation.Language.Parser]::ParseFile((Resolve-Path $path),[ref]$tokens,[ref]$errors)
if($errors.Count){throw ($errors|Out-String)}
foreach($name in @('Prop','Write-BrowserPathEvidence','Write-NativeIngressEvidence','Write-NativeFirewallEvidence','Write-BrowserFirstByteEvidence')){
 $fn=$ast.Find({param($a)$a -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $a.Name -eq $name},$true)
 if(!$fn){throw "Missing function $name"};Invoke-Expression $fn.Extent.Text
}
function Check($Condition,[string]$Message){if(!$Condition){throw $Message}}
$script:lines=@()
function Say([string]$Line){$script:lines+=@($Line)}
Write-BrowserPathEvidence ([pscustomobject]@{listening=$true}) 'edge'
Check (($script:lines -join "`n") -match 'PAC_LEGACY_UNINSTRUMENTED') 'Old service claimed measured PAC'
$script:lines=@()
Write-BrowserPathEvidence ([pscustomobject]@{
 path='PAC_CONNECT';metrics_lost=2;failure_details_lost=3
 path_metrics=@([pscustomobject]@{profile_id='edge';accepted=10;opened=0;failed=10})
 failure_details=@([pscustomobject]@{profile_id='edge';domain='mtalk.google.com';phase='OPEN';reason='UPSTREAM_OPEN_TIMEOUT';count=10})
 secret='NEVER_LOG'
}) 'edge'
$text=$script:lines -join "`n"
Check ($text -match 'NO_NEW_PAC_TUNNEL_OBSERVED' -and $text -match 'first_open=NOT_OBSERVED') 'Failed OPEN fabricated first tunnel'
Check ($text -match 'domain=mtalk.google.com stage=OPEN reason=UPSTREAM_OPEN_TIMEOUT.*count=10') 'Domain/phase/reason missing'
Check ($text -match 'failure_details_lost=3' -and $text -notmatch 'NEVER_LOG') 'Truncation or safe fields broken'
$script:lines=@()
Write-BrowserPathEvidence ([pscustomobject]@{path='PAC_CONNECT';path_metrics=@([pscustomobject]@{profile_id='chrome';opened=3;first_tunnel_opened_at='2026-10-19T14:00:00Z';listen_to_first_open_ms=10})}) 'browser'
Check (($script:lines -join "`n") -match 'state=PAC_TUNNEL_OBSERVED') 'Actual PAC OPEN missing'
Write-Host 'PASS: separate PAC/native scope, no fabricated first OPEN, scoped failures, old service, safe fields, bounded evidence'

$script:lines=@()
Write-NativeIngressEvidence ([pscustomobject]@{records=@([pscustomobject]@{ingress='PAC_NATIVE_UPSTREAM';sent_bytes=100;received_bytes=100})})
Check (($script:lines -join "`n") -match 'NATIVE_ONLY_UNINSTRUMENTED') 'Legacy PAC/native inferred as native-only'
$script:lines=@()
Write-NativeIngressEvidence ([pscustomobject]@{native_only_state='NATIVE_ONLY_NOT_OBSERVED';native_only_totals=[pscustomobject]@{bidirectional_exchanges=0};pac_native_upstream_totals=[pscustomobject]@{bidirectional_exchanges=100}})
Check (($script:lines -join "`n") -match 'state=NATIVE_ONLY_NOT_OBSERVED') 'Hybrid forged native-only'
$script:lines=@()
Write-NativeIngressEvidence ([pscustomobject]@{native_only_state='NATIVE_ONLY_OBSERVED';native_only_totals=[pscustomobject]@{bidirectional_exchanges=1;sent_bytes=10;received_bytes=20}})
Check (($script:lines -join "`n") -match 'state=NATIVE_ONLY_OBSERVED') 'True native ingress hidden'
$script:lines=@()
Write-NativeIngressEvidence ([pscustomobject]@{native_only_state='NATIVE_ONLY_OBSERVED';native_only_totals=[pscustomobject]@{bidirectional_exchanges=0}})
Check (($script:lines -join "`n") -match 'state=NATIVE_ONLY_INVALID_EVIDENCE') 'Impossible evidence accepted'
$script:lines=@()
Write-NativeFirewallEvidence ([pscustomobject]@{state='APPLIED';known=$true;rule_present=$true;program_matches=$true;effective=$true;secret='NEVER_LOG'})
Check (($script:lines -join "`n") -match 'state=APPLIED.*effective=True' -and ($script:lines -join "`n") -notmatch 'NEVER_LOG') 'Firewall proof/safe output failed'
$script:lines=@()
Write-NativeFirewallEvidence ([pscustomobject]@{state='UNKNOWN_QUERY_TIMEOUT';known=$false})
Check (($script:lines -join "`n") -match 'UNKNOWN: UNKNOWN_QUERY_TIMEOUT') 'Unknown was reported as missing/applied'
Write-Host 'PASS native-only vs PAC upstream and actual firewall evidence under StrictMode'

$script:lines=@()
Write-BrowserFirstByteEvidence ([pscustomobject]@{first_byte_metrics=@(
 [pscustomobject]@{network='net';profile_id='chrome';group='youtube-media';path='PAC_CONNECT';mode='DIRECT_FIRST';replies=1;total_ms=2100;closed_without_reply=0},
 [pscustomobject]@{network='net';profile_id='chrome';group='youtube-media';path='PAC_CONNECT';mode='RELAY_FIRST';replies=2;total_ms=120;closed_without_reply=1}
);secret='NEVER_LOG'}) 'browser'
$text=$script:lines -join "`n"
Check ($text -match 'before_direct_first_avg_ms=2100 after_relay_first_avg_ms=60') 'Observed group comparison lost'
Check ($text -match 'n=2 no_reply=1 avg_ms=60' -and $text -notmatch 'NEVER_LOG') 'First-byte failures/safe fields lost'
$script:lines=@()
Write-BrowserFirstByteEvidence ([pscustomobject]@{first_byte_metrics=@([pscustomobject]@{mode='RELAY_FIRST';replies=0;total_ms=0;closed_without_reply=3})}) 'edge'
Check (($script:lines -join "`n") -match 'before_direct_first_avg_ms=NOT_OBSERVED after_relay_first_avg_ms=NOT_OBSERVED') 'Failures forged zero-ms improvement'
Write-Host 'PASS scoped first-byte comparison and missing samples under StrictMode'

$script:lines=@()
Write-BrowserFirstByteEvidence ([pscustomobject]@{first_byte_metrics=@()}) 'browser'
Check (($script:lines -join "`n") -match 'NO_SAMPLES' -and ($script:lines -join "`n") -notmatch 'UNINSTRUMENTED') 'Empty measured ledger misreported as old service'
