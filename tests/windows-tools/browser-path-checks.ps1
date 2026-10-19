$ErrorActionPreference='Stop'
Set-StrictMode -Version Latest
$path=Join-Path $PSScriptRoot '..\..\tools\windows\diag-no-vpn.ps1'
$tokens=$null;$errors=$null
$ast=[System.Management.Automation.Language.Parser]::ParseFile((Resolve-Path $path),[ref]$tokens,[ref]$errors)
if($errors.Count){throw ($errors|Out-String)}
foreach($name in @('Prop','Write-BrowserPathEvidence')){
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
