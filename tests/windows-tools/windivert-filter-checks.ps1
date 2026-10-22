$ErrorActionPreference='Stop';Set-StrictMode -Version Latest
$root=Split-Path (Split-Path $PSScriptRoot -Parent) -Parent
foreach($relative in @('tools/windows/test-windivert-filters.ps1','tools/windows/build-e2e.ps1','tools/windows/build-setup.ps1')){
 $tokens=$null;$errors=$null
 [Management.Automation.Language.Parser]::ParseFile((Join-Path $root $relative),[ref]$tokens,[ref]$errors)|Out-Null
 if($errors.Count){throw ($errors|Out-String)}
}
$gate=Get-Content -Raw (Join-Path $root 'tools/windows/test-windivert-filters.ps1')
foreach($name in @('TestServiceFiltersCompileWithActualWinDivert','TestWinDivertCompilerRejectsLegacyGroupedNot','TestWinDivertCompilerReportsComplexityLimits','TestPrivateISCSIExcludedByActualWinDivertFilterEvaluator','NEEXITLAG_REQUIRE_FILTER_COMPILER','WINDIVERT_FILTER_PROOF_MISSING','WINDIVERT_FILTER_WINDOWS_RUNTIME_REQUIRED')){
 if(!$gate.Contains($name)){throw ('Required DLL proof missing: '+$name)}
}
if($gate -match 'Start-Service|Stop-Service|Invoke-ServiceIPC|WinDivertOpen|\bssh\b|\bscp\b'){throw 'Compiler gate must not change service/network/VPS state'}
if(!$gate.Contains('finally{') -or !$gate.Contains('$env:NEEXITLAG_TEST_DRIVER_DIR=$oldDriver') -or !$gate.Contains('$env:NEEXITLAG_REQUIRE_FILTER_COMPILER=$oldRequired')){throw 'Fixture environment must be restored'}
$build=Get-Content -Raw (Join-Path $root 'tools/windows/build-e2e.ps1')
if($build.IndexOf("'test-windivert-filters.ps1'") -gt $build.IndexOf(' $targets=@{') -or !$build.Contains('TestServiceWinDivertFilterInventory')){throw 'DLL gate must run before building service payload'}
Write-Host 'PASS filter gate scripts parse; mandatory DLL proof precedes service build; no service/network/VPS calls. Actual DLL execution is a separate Windows test.'
