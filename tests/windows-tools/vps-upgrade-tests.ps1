$ErrorActionPreference='Stop'
Set-StrictMode -Version Latest
$root=Split-Path (Split-Path $PSScriptRoot -Parent) -Parent
. (Join-Path $root 'tools/windows/update-vps.ps1') -ExpectedCommit ('a'*40) -RunId '1'
$now=[DateTimeOffset]::UtcNow
$p=[pscustomobject]@{pass=$true;same_connection=$true;ticket_replay_rejected=$true;wrong_ticket_rejected=$true;anonymous_mtls_rejected=$true;legacy_2=$true;receipt_retry=$true;renewal_replay_rejected=$true;alpn='neexitlag/3';renewals=2;default_lease_seconds=1800;tcp_source='127.0.0.1';udp_source='127.0.0.1';deadlines=@($now.ToString('o'),$now.AddMilliseconds(1).ToString('o'),$now.AddMilliseconds(2).ToString('o'))}
Assert-UpgradeProbe $p '127.0.0.1'
$p.same_connection=$false
$rejected=$false
try { Assert-UpgradeProbe $p '127.0.0.1' } catch { $rejected=$true }
if (-not $rejected) { throw 'Replacement flows accepted as continuity' }
$rejected=$false
try { Read-UpgradeProperty ([pscustomobject]@{}) 'error' } catch { $rejected=$true }
if (-not $rejected) { throw 'Missing required evidence not rejected under StrictMode' }
if ((Quote-NativeArgument 'C:\path with space\') -cne '"C:\path with space\\"') { throw 'Trailing slash quoting failed' }
$script:LogDirectory=Join-Path ([System.IO.Path]::GetTempPath()) ('neexitlag-upgrade-test-'+[Guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $script:LogDirectory | Out-Null
try {
    # Real Windows redirected Start-Process, no fake process adapter.
    $value=Invoke-UpgradeNative $env:ComSpec @('/d','/c','echo success') 'success'
    if ($value.Trim() -cne 'success') { throw 'Redirected stdout missing' }
    $rejected=$false
    try { Invoke-UpgradeNative $env:ComSpec @('/d','/c','exit 7') 'failure' | Out-Null } catch { $rejected=$true }
    if (-not $rejected) { throw 'Nonzero exit ignored' }
    $text=Get-Content -LiteralPath (Join-Path $root 'tools/windows/update-vps.ps1') -Raw
    $tokens=$null;$errors=$null
    [void][Management.Automation.Language.Parser]::ParseInput($text,[ref]$tokens,[ref]$errors)
    if ($errors.Count) { throw 'Upgrade orchestrator does not parse' }
    # The Prepare branch must end before every SSH/SCP invocation.
    if ($text.IndexOf("if (`$Mode -eq 'Prepare')") -gt $text.IndexOf("Invoke-UpgradeSSH 'test")) { throw 'Prepare can reach SSH' }
    Write-Output 'PASS: StrictMode, same-flow evidence, quoting, real redirected exit=0/nonzero, default Prepare boundary.'
} finally { Remove-Item -LiteralPath $script:LogDirectory -Recurse -Force }
