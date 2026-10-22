$ErrorActionPreference='Stop'
Set-StrictMode -Version Latest
$repo=(Resolve-Path (Join-Path $PSScriptRoot '..\..')).Path
. (Join-Path $repo 'tools\windows\browser-startup-evidence.ps1')
function Assert($value,$why){if(!$value){throw $why}}
$profiles=@([pscustomobject]@{id='chrome';mode='AUTO';browser_transport='native';executables=@('C:\Apps\chrome.exe')},[pscustomobject]@{id='edge';mode='AUTO';browser_transport='native';executables=@('C:\Apps\msedge.exe')})
$good=[pscustomobject]@{bidirectional_exchanges=1;sent_bytes=100;received_bytes=200;enable_to_first_exchange_ms=42}
$native=[pscustomobject]@{state='READY_NEW_CONNECTIONS';profiles=@([pscustomobject]@{profile_id='chrome';native_only=$good},[pscustomobject]@{profile_id='edge';native_only=$good})}
$exe=@('C:\Apps\chrome.exe','C:\Apps\msedge.exe')
$r=Test-BrowserNativeStartupEvidence $profiles $native $exe
Assert $r.ready 'valid both-browser native evidence rejected';Assert ($r.scope -eq 'NATIVE_TRANSPORT_ONLY_NOT_PAGE_PASS') 'bytes claimed page success'
$native.profiles[1].native_only=[pscustomobject]@{bidirectional_exchanges=0;sent_bytes=100;received_bytes=0}
$r=Test-BrowserNativeStartupEvidence $profiles $native $exe;Assert (!$r.ready) 'one browser or one-way traffic manufactured native proof'
$native.profiles[1].native_only=$null
$r=Test-BrowserNativeStartupEvidence $profiles $native $exe;Assert (!$r.ready) 'missing optional metrics accepted under StrictMode'
$native.profiles[1].native_only=$good;$profiles[0].mode='DIRECT'
$r=Test-BrowserNativeStartupEvidence $profiles $native $exe;Assert (!$r.ready) 'DIRECT default manufactured AUTO'
$profiles[0].mode='AUTO';$native.state='PREPARING'
$r=Test-BrowserNativeStartupEvidence $profiles $native $exe;Assert (!$r.ready) 'not-ready native reported success'
$r=Test-BrowserNativeStartupEvidence @() $null @();Assert (!$r.ready) 'empty snapshot manufactured proof'
foreach($name in @('test-browser-startup.ps1','browser-startup-evidence.ps1','install.ps1','update-service.ps1')){
 $tokens=$null;$errors=$null;[void][System.Management.Automation.Language.Parser]::ParseFile((Join-Path $repo ('tools\windows\'+$name)),[ref]$tokens,[ref]$errors)
 Assert ($errors.Count -eq 0) ('PS parser failure '+$name)
}
Write-Output 'PASS browser startup evidence, both lanes, no PAC inference, one-way/missing/DIRECT rejection and parser; no installation/VPS/browser access'
