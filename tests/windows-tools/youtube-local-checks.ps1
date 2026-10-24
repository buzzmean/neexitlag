$ErrorActionPreference='Stop';Set-StrictMode -Version Latest
$tokens=$null;$errors=$null
$ast=[System.Management.Automation.Language.Parser]::ParseFile((Join-Path $PSScriptRoot '..\..\tools\windows\diag-no-vpn.ps1'),[ref]$tokens,[ref]$errors)
if($errors.Count){throw ($errors|Out-String)}
foreach($name in @('Prop','Write-YouTubeLocalEvidence')){$fn=$ast.Find({param($a)$a -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $a.Name -eq $name},$true);if(!$fn){throw ('Missing '+$name)};Invoke-Expression $fn.Extent.Text}
$script:lines=@();function Say([string]$s){$script:lines+=@($s)}
Write-YouTubeLocalEvidence $null
if(($script:lines -join "`n") -notmatch 'UNINSTRUMENTED'){throw 'Legacy service fabricated ladder'}
$script:lines=@()
Write-YouTubeLocalEvidence ([pscustomobject]@{cache_entries=1;replay_contract='before origin only';stages=@([pscustomobject]@{stage='LOCAL';strategy='fake5x2-disorder';Connections=3;TX=123;RX=65536;Failures=0;Confirmed=1});records=@([pscustomobject]@{id=9;profile='chrome';domain='r1.googlevideo.com';stage='LOCAL';TX=123;RX=65536;Transform=$true;TransformKnown=$true;reason='SUSTAINED_TLS_TRANSPORT_NOT_PLAYBACK_PROOF'});secret='DO_NOT_PRINT'})
$text=$script:lines -join "`n"
if($text -notmatch 'stage=LOCAL strategy=fake5x2-disorder connections=3 tx=123 rx=65536' -or $text -match 'DO_NOT_PRINT'){throw 'Scoped safe metrics broken'}
if($text -notmatch 'not playback PASS'){throw 'Transport promoted to playback proof'}
Write-Host 'PASS YouTube ladder diagnostics, StrictMode, legacy response and secret allowlist (PS7; native PS5.1 not executed)'

$script:lines=@()
Write-YouTubeLocalEvidence ([pscustomobject]@{selection=[pscustomobject]@{Trials=3;TrialSuccesses=2;TrialFailures=1;FailedAfterReply=1;active_trial=$false;contract='one trial';secret='HIDDEN'};cache=@([pscustomobject]@{Profile='chrome';Domain='googlevideo:sn-abc';Streak=2;Stage='RELAY'})})
$text=$script:lines -join "`n"
if($text -notmatch 'trials=3 trial_success=2 trial_fail=1 failed_after_reply=1' -or $text -notmatch 'streak=2' -or $text -match 'HIDDEN'){throw 'Trial cost or subgroup streak missing/unsafe'}
Write-Host 'PASS CDN trial counters including failures after origin reply and consecutive streak'
