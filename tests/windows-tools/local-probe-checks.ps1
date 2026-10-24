#requires -Version 5.1
$ErrorActionPreference='Stop';Set-StrictMode -Version Latest
$tokens=$null;$errors=$null
$ast=[System.Management.Automation.Language.Parser]::ParseFile((Join-Path $PSScriptRoot '..\..\tools\windows\diag-no-vpn.ps1'),[ref]$tokens,[ref]$errors)
if($errors.Count){throw ($errors|Out-String)}
foreach($name in @('Prop','Write-LocalProbeEvidence')){$f=$ast.Find({param($x)$x -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $x.Name -eq $name},$true);Invoke-Expression $f.Extent.Text}
$script:lines=@();function Say([string]$s){$script:lines+=@($s)}
Write-LocalProbeEvidence ([pscustomobject]@{probes=@([pscustomobject]@{host='discord.com';strategy='catalog-fixture';service='discord';result='NO_TRANSFORM';committed=$false;sent_bytes=1500;received_bytes=0;inspection=[pscustomobject]@{syn_seen=$true;handshake_confirmed=$false;reason='HANDSHAKE_MISSING';sent_packets=@([pscustomobject]@{length=1250;ttl=5;protocol=6;sequence_offset=-10000;tcp_flags=24;fake=$true});private_payload='SECRET'}})})
$s=$script:lines -join ' '
if($s -notmatch 'discord.com' -or $s -notmatch 'seq_delta=-10000' -or $s -notmatch 'handshake=False' -or $s -notmatch 'fake=True' -or $s -match 'SECRET'){throw 'Missing or unsafe LOCAL probe table'}
Write-Host 'PASS LOCAL strategy/host headers, handshake and secret allowlist; PS5.1-compatible syntax (runtime separately required)'
