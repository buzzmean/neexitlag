$ErrorActionPreference='Stop'
Set-StrictMode -Version Latest
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
