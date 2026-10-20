$ErrorActionPreference='Stop';Set-StrictMode -Version Latest
$root=(Resolve-Path (Join-Path $PSScriptRoot '..\..')).Path
. (Join-Path $root 'tools/windows/activation-ssh.ps1')
function Check($Value,[string]$Message){if(!$Value){throw $Message}}
foreach($name in @('activation-ssh','verify-activation-vps','update-activation-vps')){
 $tokens=$null;$errors=$null
 [void][Management.Automation.Language.Parser]::ParseFile((Join-Path $root ('tools/windows/'+$name+'.ps1')),[ref]$tokens,[ref]$errors)
 Check ($errors.Count -eq 0) ('Parse failed: '+$name)
}
$emptyRejected=$false
try{$null=Invoke-ActivationSSH -Payload '' -RemoteCode '' -KnownHosts '' -VpsIP '' -SSHUser ''}catch{$emptyRejected=$_.Exception.Message -like 'ACTIVATION_SSH_ARGUMENTS_INVALID*'}
Check $emptyRejected 'Empty transport arguments were not rejected before process start'
Check ((Quote-ActivationSSH 'C:\Users\Test Owner\') -ceq '"C:\Users\Test Owner\\"') 'SSH argument quoting lost trailing slash'
$source=[IO.File]::ReadAllText((Join-Path $root 'tools/windows/update-activation-vps.ps1'))
Check (!$source.Contains("'bootstrap-fresh-pc.ps1'")) 'Deployment must not source enrollment bootstrap'
if([Environment]::OSVersion.Platform -ne [PlatformID]::Win32NT){Write-Host 'PASS parsing/quoting; redirected fake ssh.exe native test requires Windows';return}
# A genuine child process, no Pester/new dependency, network or root credential.
$fixture=Join-Path ([IO.Path]::GetTempPath()) ('veliq-ssh-fixture-'+[Guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $fixture|Out-Null
$oldPath=$env:PATH
try{
 $program=@'
using System;
using System.IO;
using System.Text;
public class ActivationSSHFixture {
 public static void Main(string[] args) {
  string input;
  using(var reader=new StreamReader(Console.OpenStandardInput(),new UTF8Encoding(false),true)) { input=reader.ReadToEnd(); }
  if(input=="fail") {
   Console.Error.WriteLine("ACTIVATION_DEPLOYMENT_FAILED stage=CANDIDATE_READ_ONLY_PREFLIGHT exit_code=13 vps_modified=NO");
   Environment.Exit(23);
  }
  if(input!="success") { Console.Error.WriteLine("BAD_FIXTURE_INPUT_LENGTH="+input.Length); Environment.Exit(97); }
  Console.WriteLine("SAFE_FIXTURE_OK");
 }
}
'@
 Add-Type -TypeDefinition $program -OutputAssembly (Join-Path $fixture 'ssh.exe') -OutputType ConsoleApplication
 $env:PATH=$fixture+';'+$env:PATH
 $sshFixtureArguments=@{Payload='success';RemoteCode='print("unused fixture")';KnownHosts=(Join-Path $fixture 'known_hosts');VpsIP='127.0.0.1';SSHUser='fixture';SSHPort=1}
 $ok=Invoke-ActivationSSH @sshFixtureArguments
 Check ($ok.Trim() -eq 'SAFE_FIXTURE_OK') 'Redirected success output/exit was lost'
 $sshFixtureArguments.Payload='fail';$failed=$false
 try{$null=Invoke-ActivationSSH @sshFixtureArguments}catch{$failed=$true;Check ($_.Exception.Message -like 'ACTIVATION_REMOTE_OPERATION_FAILED*') 'Wrong failure code';Check (!$_.Exception.Message.Contains('no installation performed')) 'False unchanged-installation claim'}
 Check $failed 'Nonzero redirected SSH exit was accepted'
 Write-Host 'PASS native redirected ssh.exe success/failure, wait/exit validation and no false VPS-unchanged claim'
}finally{$env:PATH=$oldPath;Remove-Item -LiteralPath $fixture -Recurse -Force}
