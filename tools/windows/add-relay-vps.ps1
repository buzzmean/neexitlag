#requires -Version 5.1
# Owner-only preparation/deployment. Never invoked by GUI, installation or activation.
[CmdletBinding()]
param(
 [Parameter(Mandatory=$true)][Alias('Host')][string]$RelayHost,
 [ValidatePattern('^[A-Za-z0-9_-]{1,64}$')][string]$Name='hostkey-ams',
 [string]$SSHKey=(Join-Path $env:USERPROFILE '.ssh\id_ed25519'),
 [string]$Manifest=(Join-Path $env:ProgramData 'neexitlag\service.json'),
 [string]$ExpectedCommit='',
 [switch]$PromotePrimary,[switch]$Rollback,[switch]$Inspect
)
$ErrorActionPreference='Stop';Set-StrictMode -Version Latest
. (Join-Path $PSScriptRoot 'common.ps1')
if(!$PromotePrimary){throw 'P0_ONLY_REQUIRES_PromotePrimary; multi-relay A-E is deferred'}
$RelayHost=Assert-VpsIPv4 $RelayHost
if($RelayHost -eq '2.56.177.4'){throw 'NEW_RELAY_MUST_NOT_BE_CONTROL_PLANE'}
$repo=(Resolve-Path (Join-Path $PSScriptRoot '..\..')).Path
foreach($tool in @('git','go','python','ssh')){if(!(Get-Command $tool -ErrorAction SilentlyContinue)){throw ('REQUIRED_TOOL_'+$tool)}}
if(!(Test-Path -LiteralPath $SSHKey -PathType Leaf)){throw 'OWNER_SSH_PRIVATE_KEY_REQUIRED; contents are never printed'}
if(!$Rollback -and !$Inspect -and !(Test-Path -LiteralPath $Manifest -PathType Leaf)){throw 'ACTIVATED_OWNER_PC_MANIFEST_REQUIRED; no installation/enrollment is performed'}
$commit=(& git -C $repo rev-parse HEAD).Trim();if($LASTEXITCODE -or $commit -notmatch '^[a-f0-9]{40}$'){throw 'EXACT_COMMIT_REQUIRED'}
if($ExpectedCommit -and $ExpectedCommit -cne $commit){throw 'EXPECTED_COMMIT_MISMATCH'}
$dirty=@(& git -C $repo status --porcelain --untracked-files=no);if($LASTEXITCODE -or $dirty.Count){throw 'CLEAN_TRACKED_CHECKOUT_REQUIRED'}
$bin=Join-Path $repo 'bin\relay-fleet';[void][IO.Directory]::CreateDirectory($bin)
$oldOS=$env:GOOS;$oldArch=$env:GOARCH;$oldCGO=$env:CGO_ENABLED
Push-Location $repo
try{
 $env:GOOS='linux';$env:GOARCH='amd64';$env:CGO_ENABLED='0'
 foreach($tool in @('relay','relay-admin')){
  $package=if($tool -eq 'relay'){'./services/relay/cmd/relay'}else{'./tools/relay-admin'}
  $target=Join-Path $bin $tool;Remove-Item -LiteralPath $target -Force -ErrorAction SilentlyContinue
  & go build -trimpath -o $target $package;if($LASTEXITCODE){throw ('BUILD_FAILED_'+$tool)}
 }
 $env:GOOS='windows'
 $target=Join-Path $bin 'relay-client.exe';Remove-Item -LiteralPath $target -Force -ErrorAction SilentlyContinue
 & go build -trimpath -o $target './tools/relay-client';if($LASTEXITCODE){throw 'CLIENT_PROOF_BUILD_FAILED'}
}finally{$env:GOOS=$oldOS;$env:GOARCH=$oldArch;$env:CGO_ENABLED=$oldCGO;Pop-Location}
$arguments=@((Join-Path $repo 'deploy\vps\relay-fleet-owner.py'),'--host',$RelayHost,'--name',$Name,'--key',$SSHKey,'--bin',$bin,'--manifest',$Manifest)
if($PromotePrimary){$arguments+='--promote'}
if($Rollback){$arguments+='--rollback'}
if($Inspect){$arguments+='--inspect'}
Write-Host ('Owner relay-only operation; commit='+$commit+' host='+$RelayHost+'; no CA rotation, enrollment, activation deployment or revoke-tests')
& python @arguments
if($LASTEXITCODE){throw 'RELAY_FLEET_OPERATION_FAILED; inspect safe stage/rollback report above'}
