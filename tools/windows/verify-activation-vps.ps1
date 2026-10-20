# Read-only existing-VPS snapshot. Does not build, enroll or stage remote files.
[CmdletBinding()]
param(
 [ValidateSet('2.56.177.4')][string]$VpsIP='2.56.177.4',
 [ValidatePattern('^[a-z_][a-z0-9_-]{0,31}$')][string]$SSHUser='root',
 [ValidateRange(1,65535)][int]$SSHPort=22,
 [ValidatePattern('^[a-f0-9]{40}$')][string]$BaselineCommit='2c06089f61c7b7a2cf575f5cea4b834992426d'
)
$ErrorActionPreference='Stop';Set-StrictMode -Version Latest
. (Join-Path $PSScriptRoot 'activation-ssh.ps1')
$root=(Resolve-Path (Join-Path $PSScriptRoot '..\..')).Path
function Source64([string]$Path){return [Convert]::ToBase64String([IO.File]::ReadAllBytes((Join-Path $root $Path)))}
$body=@{engine=(Source64 'deploy/vps/production-upgrade.py');fresh=(Source64 'deploy/vps/fresh-vps-upgrade.py');audit=(Source64 'deploy/vps/inspect-activation.py');baseline=$BaselineCommit}
$remote=@'
import base64,json,sys,types
try:
 p=json.loads(sys.stdin.buffer.read(1048576))
 a=types.ModuleType('readonly_activation');a.__file__='<readonly-activation>'
 exec(compile(base64.b64decode(p['audit'],validate=True),a.__file__,'exec'),a.__dict__)
 print(json.dumps(a.run(p),sort_keys=True))
except BaseException as e:
 print('ACTIVATION_READ_ONLY_FAILED '+json.dumps({'stage':'READ_ONLY_SNAPSHOT','exception':type(e).__name__,'exit_code':getattr(e,'exit_code',None),'vps_modified':'NO_PRODUCTION_MUTATION_REQUESTED'}),file=sys.stderr)
 sys.exit(1)
'@
$state=Join-Path $root '.e2e\fresh-pc';New-Item -ItemType Directory -Force -Path $state|Out-Null
Invoke-ActivationSSH -Payload ($body|ConvertTo-Json -Compress) -RemoteCode $remote -KnownHosts (Join-Path $state 'known_hosts') -VpsIP $VpsIP -SSHUser $SSHUser -SSHPort $SSHPort|Write-Host
