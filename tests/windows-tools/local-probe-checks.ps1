#requires -Version 5.1
$ErrorActionPreference='Stop';Set-StrictMode -Version Latest
$tokens=$null;$errors=$null
$ast=[System.Management.Automation.Language.Parser]::ParseFile((Join-Path $PSScriptRoot '..\..\tools\windows\diag-no-vpn.ps1'),[ref]$tokens,[ref]$errors)
if($errors.Count){throw ($errors|Out-String)}
foreach($name in @('Prop','Write-CdnTrialEvidence','Write-LocalProbeEvidence')){$f=$ast.Find({param($x)$x -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $x.Name -eq $name},$true);Invoke-Expression $f.Extent.Text}
$script:lines=@();function Say([string]$s){$script:lines+=@($s)}
Write-LocalProbeEvidence ([pscustomobject]@{probes=@([pscustomobject]@{host='discord.com';strategy='catalog-fixture';service='discord';result='NO_TRANSFORM';committed=$false;sent_bytes=1500;received_bytes=0;inspection=[pscustomobject]@{syn_seen=$true;handshake_confirmed=$false;reason='HANDSHAKE_MISSING';sent_packets=@([pscustomobject]@{length=1250;ttl=5;protocol=6;sequence_offset=-10000;tcp_flags=24;fake=$true;timestamp_present=$true;timestamp_delta=-600000;ack_offset=0;md5_present=$false});private_payload='SECRET'}})})
Write-LocalProbeEvidence ([pscustomobject]@{tcp_closures=@([pscustomobject]@{Domain='rr.googlevideo.com';Strategy='candidate';Reason='EARLY_EOF_OR_IO_NO_REPLAY';IOFailure='REMOTE_RESET';Timing=[pscustomobject]@{hello_to_close_ms=1200;last_rx_to_close_ms=200;hello_to_server_hello_ms=150;rx_after_server_hello_observed=6000;ClientApplicationRecord=$true};Packets=[pscustomobject]@{first_close='REMOTE_PACKET';RemoteRST=1;SYNACKTTL=52;RemoteTerminalTTL=117};private_payload='SECRET'});probes=@()})
$s=$script:lines -join ' '
if($s -notmatch 'remote_rst=1' -or $s -notmatch 'remote_terminal_ttl=117' -or $s -notmatch 'ts_delta=-600000' -or $s -notmatch 'hello_to_server_hello_ms=150' -or $s -notmatch 'rx_after_hello_observed=6000'){throw 'Missing TCP/fooling diagnostics'}
if($s -notmatch 'discord.com' -or $s -notmatch 'seq_delta=-10000' -or $s -notmatch 'handshake=False' -or $s -notmatch 'fake=True' -or $s -match 'SECRET'){throw 'Missing or unsafe LOCAL probe table'}
Write-Host 'PASS LOCAL strategy/host headers, handshake and secret allowlist; PS5.1-compatible syntax (runtime separately required)'

$script:lines=@()
Write-LocalProbeEvidence ([pscustomobject]@{failures=@([pscustomobject]@{Domain='googlevideo:sn-test';Reason='MIDSTREAM_STALL_NEW_CONNECTION_REQUIRED';SinceClientHelloMS=8000;SinceLastRXMS=6001;RX=20000;ReceivedAfterHello=19900;Promoted=$true;private_payload='SECRET'});tcp_closures=@([pscustomobject]@{Timing=[pscustomobject]@{ApplicationRecords=65;ApplicationBytes=1064960;ApplicationSpanMS=3400}})})
$s=$script:lines -join ' '
if($s -notmatch 'last_rx_to_failure_ms=6001' -or $s -notmatch 'rx_after_hello=19900' -or $s -notmatch 'promoted=True' -or $s -notmatch 'application_records=65 application_bytes=1064960 application_span_ms=3400' -or $s -match 'SECRET'){throw 'Missing/unsafe sustained or midstream failure evidence'}
Write-Host 'PASS running sustained proof counters and exact failure timing/volume allowlist'

$script:lines=@()
Write-CdnTrialEvidence ([pscustomobject]@{trials=@([pscustomobject]@{Host='rr5---sn-test.googlevideo.com';TX=5791;RX=7109;TrialResult='INCONCLUSIVE_SMALL_EOF';Packets=[pscustomobject]@{first_close='REMOTE_PACKET';RemoteFIN=1;ClientRetransmits=2};Timing=[pscustomobject]@{hello_to_close_ms=290};Sibling=[pscustomobject]@{AgeMS=15000;Reason='RECENT_SAME_SUBGROUP'};private_payload='SECRET'})})
$s=$script:lines -join ' '
if($s -notmatch 'host=rr5---sn-test.googlevideo.com' -or $s -notmatch 'trial_result=INCONCLUSIVE_SMALL_EOF' -or $s -notmatch 'tx=5791 rx=7109' -or $s -notmatch 'hello_to_close_ms=290' -or $s -notmatch 'sibling_age_at_admission_ms=15000' -or $s -match 'SECRET'){throw 'CDN trial evidence missing/unsafe'}
Write-Host 'PASS per-trial exact host/neutral result/close timing/recent sibling allowlist'

$script:lines=@();Write-CdnTrialEvidence ([pscustomobject]@{trials=$null});if(($script:lines -join ' ') -match 'CDN trial: host='){throw 'Null history fabricated a trial'}
Write-Host 'PASS empty history does not fabricate a trial'
