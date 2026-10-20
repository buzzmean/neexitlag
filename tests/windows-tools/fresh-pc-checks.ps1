$ErrorActionPreference='Stop'
Set-StrictMode -Version Latest
$tokens=$null;$errors=$null
. (Join-Path $PSScriptRoot '..\..\tools\windows\fresh-pc-execution.ps1')
$executionPath=Join-Path $PSScriptRoot '..\..\tools\windows\fresh-pc-execution.ps1'
$executionAst=[System.Management.Automation.Language.Parser]::ParseFile((Resolve-Path $executionPath),[ref]$tokens,[ref]$errors)
if($errors.Count){throw ($errors|Out-String)}
# Compile the exact native adapter even on Linux; only Windows CI calls Read().
$native=$executionAst.Find({param($a)$a -is [System.Management.Automation.Language.StringConstantExpressionAst] -and $a.Value.Contains('public sealed class FreshPCToken')},$true)
if($null -eq $native){throw 'Native token adapter missing'}
Add-Type -TypeDefinition $native.Value
$path=Join-Path $PSScriptRoot '..\..\tools\windows\bootstrap-fresh-pc.ps1'
$tokens=$null;$errors=$null
$ast=[System.Management.Automation.Language.Parser]::ParseFile((Resolve-Path $path),[ref]$tokens,[ref]$errors)
if($errors.Count){throw ($errors|Out-String)}
foreach($name in @('Require-FreshPin','Fresh-Property','Assert-FreshSummary','Wait-FreshServiceStartup','Fresh-Quote')){
 $fn=$ast.Find({param($a)$a -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $a.Name -eq $name},$true)
 if(!$fn){throw "Missing $name"};Invoke-Expression $fn.Extent.Text
}
function Check($ok,$message){if(!$ok){throw $message}}
function Reject([scriptblock]$Action){$failed=$false;try{& $Action}catch{$failed=$true};Check $failed 'Invalid evidence accepted'}
Check ((Require-FreshPin ('AB:'*31+'AB')) -ceq ('ab'*32)) 'Pin normalization failed'
Reject {Require-FreshPin 'invalid'}
Reject {Require-FreshPin ('a'*63)}
Reject {Assert-FreshSummary ([pscustomobject]@{demo=$false;connected=$false;backend=[pscustomobject]@{ready=$true}})}
Reject {Assert-FreshSummary ([pscustomobject]@{demo=$true;connected=$true;backend=[pscustomobject]@{ready=$true}})}
Reject {Assert-FreshSummary ([pscustomobject]@{demo=$false;connected=$true;backend=[pscustomobject]@{}})}
Reject {Assert-FreshSummary ([pscustomobject]@{demo=$false;connected=$true;backend=[pscustomobject]@{ready=$false}})}
Assert-FreshSummary ([pscustomobject]@{demo=$false;connected=$true;backend=[pscustomobject]@{ready=$true}})
Check ((Fresh-Quote 'C:\space path\') -eq '"C:\space path\\"') 'Native trailing slash quoting failed'
$text=[IO.File]::ReadAllText((Resolve-Path $path))
Check ($text -match 'dotnet --list-sdks' -and $text -match 'RestoreConfigFile') 'SDK/source detection missing'
Check ($text -match 'StrictHostKeyChecking=ask' -and $text -notmatch 'StrictHostKeyChecking=no') 'SSH host trust weakened'
Check ($text -match 'ca-sha256 \$pin' -and $text.IndexOf('complete --state') -lt $text.IndexOf('-Verb RunAs')) 'Trust pin not checked before installation'
Check ($text -match 'WaitForExit\(\);\$p.Refresh\(\)' -and $text -match '\$p.ExitCode -ne 0') 'Redirected SSH process exit not validated'
Check ($text -notmatch '&\s+.*(?:setup-e2e|vps-provision|update-vps)') 'Fresh bootstrap reprovisions server'
Write-Host 'PASS fresh-PC PS parsing, SDK/pin/StrictMode errors, quoting, SSH exit-check structure and no implicit provisioning'

function Invoke-ServiceIPC($method){
 if($method -eq 'status'){return [pscustomobject]@{connected=$true;demo=$false}}
 return [pscustomobject]@{fingerprint='network';observed_at=[DateTimeOffset]::UtcNow.ToString('o')}
}
Wait-FreshServiceStartup 0 # ready at deadline, with optional error ABSENT under StrictMode
function Invoke-ServiceIPC($method){return [pscustomobject]@{connected=$false;demo=$false}}
Reject {Wait-FreshServiceStartup 0}
Write-Host 'PASS bounded startup wait, optional network.error and final deadline sample'

# Pure policy/token matrix: never change UAC, obtain an elevated token, install,
# enroll or contact SSH. Consent=0 alone is NOT proof that UAC is disabled.
function Execution-Fixture {
 $e=New-FreshExecutionEvidence
 $e.EnableLUA=0;$e.EnableLUAKind='DWord';$e.ConsentPromptBehaviorAdmin=0;$e.ConsentKind='DWord'
 foreach($name in @('PolicyReadError','OwnerQueryError','NativeReadError','OpenTokenError','ElevationTypeError','ElevationError','RestrictedError','LinkedTokenError')){$e.$name=0}
 $e.IsAdministrator=$true;$e.ElevationType=1;$e.Elevated=$true;$e.Restricted=$false;$e.HasLinkedToken=$false;$e.LinkedTokenQuery='ABSENT'
 return $e
}
foreach($lua in @(0,1)){
 foreach($consent in @(0,5)){
  foreach($admin in @($false,$true)){
   foreach($elevated in @($false,$true)){
    foreach($type in @(1,2,3)){
     foreach($linked in @($false,$true)){
      foreach($restricted in @($false,$true)){
       $e=Execution-Fixture
       $e.EnableLUA=$lua;$e.ConsentPromptBehaviorAdmin=$consent;$e.IsAdministrator=$admin;$e.Elevated=$elevated;$e.ElevationType=$type;$e.HasLinkedToken=$linked;$e.Restricted=$restricted
       $e.LinkedTokenQuery=if($linked){'PRESENT'}else{'ABSENT'}
       $full=($lua -eq 0 -and $admin -and $elevated -and $type -eq 1 -and !$linked -and !$restricted)
       $owner=($lua -eq 1 -and !$admin -and !$elevated -and !$restricted -and (($type -eq 3 -and $linked) -or ($type -eq 1 -and !$linked)))
       if($full){Check ((Resolve-FreshExecutionMode $e) -ceq 'UAC_DISABLED_FULL_ADMIN') 'Full-admin UAC-off rejected'}
       elseif($owner){Check ((Resolve-FreshExecutionMode $e) -ceq 'UAC_ENABLED_OWNER') 'Ordinary owner rejected'}
       else{Reject {Resolve-FreshExecutionMode $e}}
      }
     }
    }
   }
  }
 }
}
Reject {Resolve-FreshExecutionMode ([pscustomobject]@{})}
Reject {Resolve-FreshExecutionMode ([pscustomobject]@{Restricted=$false;EnableLUA=2;ConsentPromptBehaviorAdmin=0;ElevationType=1})}
foreach($errorCode in @(5,87,1008,1312)){
 # Exact owner facts: UAC=0/consent=0/Default=1/elevated=1/restricted=0,
 # while querying TokenLinkedToken fails. No failure of a different query.
 $e=Execution-Fixture;$e.HasLinkedToken=$null;$e.LinkedTokenQuery='ERROR';$e.LinkedTokenError=$errorCode
 Check ((Resolve-FreshExecutionMode $e) -ceq 'UAC_DISABLED_FULL_ADMIN') 'Default linked-query error rejected owner facts'
 $diag=Format-FreshExecutionEvidence $e 'ACCEPTED'
 Check ($diag.Contains('EnableLUA=0') -and $diag.Contains('ElevationType=1') -and $diag.Contains('Elevated=True') -and $diag.Contains('Restricted=False') -and $diag.Contains('LinkedTokenQuery=ERROR') -and $diag.Contains('LinkedTokenError='+$errorCode)) 'Diagnostic lost owner token facts'
 $e.Restricted=$true;Reject {Resolve-FreshExecutionMode $e};$e.Restricted=$false
 foreach($type in @(2,3)){$e.ElevationType=$type;Reject {Resolve-FreshExecutionMode $e}}
 $e.ElevationType=1;$e.EnableLUA=1;Reject {Resolve-FreshExecutionMode $e}
 $e.EnableLUA=0;$e.Restricted=$null;$e.RestrictedError=5;Reject {Resolve-FreshExecutionMode $e}
}
$e=Execution-Fixture # successful linked query with null handle => ABSENT
Check ((Resolve-FreshExecutionMode $e) -ceq 'UAC_DISABLED_FULL_ADMIN') 'Null linked handle rejected Default token'
$e.LinkedTokenQuery='NOT_RUN';$e.HasLinkedToken=$null;Reject {Resolve-FreshExecutionMode $e}
$e=Execution-Fixture;$e.EnableLUA=1
try{Resolve-FreshExecutionMode $e;throw 'Elevated UAC-on accepted'}catch{Check ($_.Exception.Message -ceq 'UAC_ENABLED_ELEVATED_REFUSED') 'Refusal lost UAC-on condition'}
$e=Execution-Fixture;$e.Restricted=$true
try{Resolve-FreshExecutionMode $e;throw 'Restricted accepted'}catch{Check ($_.Exception.Message -ceq 'RESTRICTED_TOKEN_REFUSED') 'Refusal lost restricted condition'}
$e=Execution-Fixture;$e.ElevationType=3;$e.HasLinkedToken=$null;$e.LinkedTokenQuery='ERROR';$e.LinkedTokenError=87
try{Resolve-FreshExecutionMode $e;throw 'Limited linked-query error accepted'}catch{Check ($_.Exception.Message -ceq 'LINKED_TOKEN_QUERY_UNAVAILABLE') 'Limited linked-query failure was hidden'}
$e=Execution-Fixture;$e.HasLinkedToken=$null;$e.LinkedTokenQuery='ERROR';$e.LinkedTokenError=0
Reject {Resolve-FreshExecutionMode $e} # ERROR without an actual error code is not evidence
$e=Execution-Fixture;$e.EnableLUAKind='String';$e.EnableLUA='0'
try{Resolve-FreshExecutionMode $e;throw 'Wrong registry kind accepted'}catch{Check ($_.Exception.Message -ceq 'ENABLE_LUA_NOT_DWORD') 'Refusal did not identify registry condition'}
$e=New-FreshExecutionEvidence
$e | Add-Member -NotePropertyName PrivatePayload -NotePropertyValue 'DO_NOT_PRINT_FIXTURE'
$diag=Format-FreshExecutionEvidence $e 'DO_NOT_PRINT_FIXTURE / arbitrary error'
Check ($diag.Contains('EnableLUA=unknown') -and $diag.Contains('Restricted=unknown') -and !$diag.Contains('DO_NOT_PRINT_FIXTURE')) 'Partial/redacted diagnostic failed'
Check ($text.Contains('Format-FreshExecutionEvidence $evidence $_.Exception.Message')) 'Bootstrap does not emit structured rejection diagnostics'
Write-Host 'PASS exact UAC-off facts, Default linked-query errors/null handle, restricted/limited refusal and safe failure diagnostics'
foreach($mode in @('UAC_DISABLED_FULL_ADMIN','UAC_ENABLED_OWNER')){
 $calls=[Collections.Generic.List[string]]::new()
 Invoke-FreshServicePhase $mode 'S-1-5-21-test' {param($sid)$calls.Add('same-process:'+ $sid)} {param($sid)$calls.Add('elevated-child:'+ $sid)}
 $expected=if($mode -eq 'UAC_DISABLED_FULL_ADMIN'){'same-process:S-1-5-21-test'}else{'elevated-child:S-1-5-21-test'}
 Check ($calls.Count -eq 1 -and $calls[0] -ceq $expected) 'Service phase escalated incorrectly or changed owner SID'
}
$calls.Clear()
Reject {Invoke-FreshServicePhase 'unknown' 'S-1-5-21-test' {$calls.Add('invalid')} {$calls.Add('invalid')}}
Check ($calls.Count -eq 0) 'Unknown mode reached installation'
Check ($text -match 'Install-FreshService \$installationSid' -and $text -match 'InstallationOwnerSid -cne \$who.User.Value') 'Same-process install lost owner check'
Check ($text -match 'SetAccessRuleProtection\(\$true,\$false\)' -and $text -match 'INSTALLED_CREDENTIAL_MISMATCH') 'Staging ACL or credential checks removed'
Check ($text -notmatch [regex]::Escape('C:\neexitlag')) 'Bootstrap depends on root of drive'
Check ((Fresh-Quote 'C:\Users\Test Owner\neexitlag\') -ceq '"C:\Users\Test Owner\neexitlag\\"') 'User-profile path quoting failed'
# Evaluate only the path assignment from the real AST in an isolated profile-like
# directory. Do not execute the bootstrap entry point or any installation action.
$fixture=Join-Path ([IO.Path]::GetTempPath()) ('veliq path-'+[Guid]::NewGuid().ToString('N'))
try{
 $fixtureRepo=Join-Path $fixture 'Users\Test Owner\neexitlag'
 $fixtureScripts=Join-Path $fixtureRepo 'tools\windows'
 [void][IO.Directory]::CreateDirectory($fixtureScripts)
 $assignment=$ast.Find({param($a)$a -is [System.Management.Automation.Language.AssignmentStatementAst] -and $a.Left.Extent.Text -ceq '$root'},$true)
 $pathCode=$assignment.Extent.Text.Replace('$PSScriptRoot','$fixtureScripts')
 Invoke-Expression $pathCode
 Check ($root -ceq (Resolve-Path $fixtureRepo).Path) 'Bootstrap root depends on current directory or C drive root'
}finally{if(Test-Path -LiteralPath $fixture){Remove-Item -LiteralPath $fixture -Recurse -Force}}
if($env:OS -eq 'Windows_NT'){
 # Actual read-only native probe in PS5.1 and PS7 CI; this does not require admin.
 $actual=Get-FreshExecutionEvidence
 Check ($actual.EnableLUA -in @(0,1) -and $actual.ElevationType -in @(1,2,3)) 'Windows token/policy query failed'
 Check (($actual.HasLinkedToken -is [bool] -or ($actual.ElevationType -eq 1 -and $actual.LinkedTokenQuery -ceq 'ERROR' -and $actual.LinkedTokenError -ne 0)) -and $actual.Restricted -is [bool]) 'Linked/restricted token evidence missing'
 Check ($actual.PolicyReadError -eq 0 -and $actual.NativeReadError -eq 0 -and $actual.RestrictedError -eq 0 -and $actual.ElevationError -eq 0) (Format-FreshExecutionEvidence $actual 'CI_QUERY_FAILED')
}
Write-Host 'PASS UAC/token matrix, restricted-token refusal, same-SID dispatch and user-profile paths'
