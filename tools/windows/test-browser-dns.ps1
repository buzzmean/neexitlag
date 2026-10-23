<#
.SYNOPSIS
Readonly bounded observation of NEW native browser flows. No browser restart,
policy change, web request or installation. Actual page/playback is checked by owner.
.EXAMPLE
& .\tools\windows\test-browser-dns.ps1 -DurationSeconds 90
#>
[CmdletBinding()]
param([ValidateRange(10,120)][int]$DurationSeconds=90)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'common.ps1')
. (Join-Path $PSScriptRoot 'browser-dns-evidence.ps1')
$repo=(Resolve-Path (Join-Path $PSScriptRoot '..\..')).Path
$baseline=Join-Path $repo 'bin\diagnostics\browser-before-veliq.json'
if(!(Test-Path -LiteralPath $baseline)){throw 'BROWSER_DNS_BASELINE_MISSING: сначала & .\tools\windows\test-browser-startup.ps1 -Phase baseline при уже открытых Chrome/Edge; не закрывайте браузеры.'}
$before=Get-Content -Raw -LiteralPath $baseline|ConvertFrom-Json
function Assert-DNSBrowserIdentity {
 foreach($browser in @($before.browsers)){
  $p=Get-CimInstance Win32_Process -Filter ('ProcessId='+[uint32]$browser.pid) -ErrorAction SilentlyContinue
  if(!$p -or !$p.CreationDate -or $p.CreationDate.ToUniversalTime().Ticks.ToString([Globalization.CultureInfo]::InvariantCulture) -ne $browser.birth -or [string]$p.ExecutablePath -ine [string]$browser.path){throw 'BROWSER_PID_OR_BIRTH_CHANGED; результат без перезапуска не подтверждён'}
 }
}
Assert-DNSBrowserIdentity
$initial=Invoke-ServiceIPC 'browser.native.status'
$after=0L
foreach($r in @(Get-BrowserDNSField $initial 'records' @())){$id=[long](Get-BrowserDNSField $r 'id' 0);if($id -gt $after){$after=$id}}
$lost=[long](Get-BrowserDNSField $initial 'lost' 0);$maxLost=$lost
$seen=@{};$until=[DateTime]::UtcNow.AddSeconds($DurationSeconds)
Write-Output ('READ_ONLY_OBSERVATION '+$DurationSeconds+'s: откройте YouTube в НОВОЙ вкладке уже открытого Chrome/Edge, проиграйте и перемотайте видео; откройте vk.ru. Браузеры не закрывать. Старые HTTP/2 соединения могут остаться прежними.')
do {
 Assert-DNSBrowserIdentity
 $status=Invoke-ServiceIPC 'status'
 if((Get-BrowserDNSField $status 'connected' $false) -ne $true -or (Get-BrowserDNSField $status 'optimizing' $false) -ne $true){throw 'BROWSER_DNS_SERVICE_NOT_READY: connected/optimizing должны быть True; настройки скрипт не меняет'}
 $snapshot=Invoke-ServiceIPC 'browser.native.status'
 $maxLost=[Math]::Max($maxLost,[long](Get-BrowserDNSField $snapshot 'lost' 0))
 foreach($r in @(Get-BrowserDNSField $snapshot 'records' @())){
  $id=[long](Get-BrowserDNSField $r 'id' 0)
  if($id -le $after){continue}
  if(!$seen.ContainsKey($id) -and $seen.Count -ge 512){throw 'BROWSER_DNS_OBSERVATION_CAPACITY; ограничение512, сократите нагрузку'}
  $seen[$id]=Get-BrowserDNSTransportObservation $r
 }
 Start-Sleep -Milliseconds 500
}while([DateTime]::UtcNow -lt $until)
Assert-DNSBrowserIdentity
$rows=@($seen.Values)
$rows|Where-Object {$_.group -ne 'OTHER'}|Sort-Object id|Select-Object -Last 48|Format-Table id,profile_id,domain,action,source,destination,tx,rx,server_hello_complete,failure -AutoSize
$verdict=Get-BrowserDNSTransportVerdict $rows
Write-Output ('BROWSER_DNS_TRANSPORT verdict='+$verdict+' observed_new_flows='+$rows.Count+' retained_records_lost_delta='+($maxLost-$lost)+' browser_pid_birth=UNCHANGED')
Write-Output 'Bounded sampled transport evidence only; no claim of complete no-leak proof or page/video PASS. Opaque Google cannot prove googlevideo. Existing HTTP/2 may be reused; lack of a new hostname flow is NOT a failed page test.'
if($verdict -ne 'TRANSPORT_ROUTES_OBSERVED_PAGE_NOT_VERIFIED'){throw ('BROWSER_DNS_TRANSPORT_NOT_CONFIRMED: '+$verdict+'; сохраните diag-no-vpn.ps1 -Only browser,youtube; значения секретов не присылайте')}
