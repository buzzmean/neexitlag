[CmdletBinding()]
param(
 [string]$ChromePath="$env:ProgramFiles\Google\Chrome\Application\chrome.exe",
 [string]$TokenFile="$env:ProgramData\neexitlag\ipc.token",
 [ValidateRange(15,60)][int]$ObserveSeconds=30,
 [switch]$RestartService,
 [string]$ReportDirectory=(Join-Path $PSScriptRoot '..\..\bin\chrome-sites')
)
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'common.ps1')
. (Join-Path $PSScriptRoot 'simultaneous-checks.ps1')
if($RestartService){Assert-Administrator}
$ChromePath=(Resolve-Path -LiteralPath $ChromePath).Path
$run=Join-Path $ReportDirectory ([DateTimeOffset]::UtcNow.ToString('yyyyMMddTHHmmssfffZ'))
New-Item -ItemType Directory -Force -Path $run|Out-Null
$run=(Resolve-Path $run).Path
$urls=@('https://vk.com/','https://chatgpt.com/','https://www.twitch.tv/','https://music.yandex.ru/','https://langame.ru/','https://web.telegram.org/','https://www.youtube.com/')
$script:targets=@{};$script:requestId=0;$script:eventLoss=$false;$script:pages=0
$client=$null;$socket=$null;$proof=@{pass=$false;scope='Public HTTP(S), Chrome policy and route evidence; not logged-in functionality/video playback';stages=@()}
function Service([string]$Method,$Payload=$null){Invoke-ServiceIPC $Method $Payload $TokenFile}
function Save-Proof($Value,[string]$Name){$Value|ConvertTo-Json -Depth 28|Set-Content -LiteralPath (Join-Path $run $Name) -Encoding UTF8}
function Origin([string]$URL){try{return ([Uri]$URL).GetLeftPart([UriPartial]::Authority)}catch{return 'invalid'}}
function Observe-CDP($Message){
 $session=[string](Get-SimValue $Message 'sessionId' '')
 if(!$script:targets.ContainsKey($session)){return}
 $row=$script:targets[$session];$method=[string](Get-SimValue $Message 'method' '')
 if($method -eq 'Network.requestWillBeSent'){
  if($row.requests.Count -ge 2048){$script:eventLoss=$true}else{$row.requests[[string]$Message.params.requestId]=$true}
 }elseif($method -eq 'Network.responseReceived'){
  $p=$Message.params
  if($p.type -eq 'Document' -and $p.frameId -eq $row.frame){$row.document=@{id=$p.requestId;origin=(Origin $p.response.url);status=$p.response.status;protocol=$p.response.protocol;finished=$false}}
 }elseif($method -eq 'Network.loadingFinished'){
  if($null -ne $row.document -and $row.document.id -eq $Message.params.requestId){$row.document.finished=$true}
 }elseif($method -eq 'Network.loadingFailed'){
  $p=$Message.params
  if($row.requests.ContainsKey([string]$p.requestId) -and $p.type -in @('Document','Fetch','XHR','Script','Stylesheet')){
   if($row.failures.Count -ge 64){$script:eventLoss=$true}else{$row.failures+=@{type=$p.type;error=$p.errorText;request_id=$p.requestId}}
  }
 }
}
function CDP([string]$Method,$Params,[string]$Session=''){
 $script:requestId++;$id=$script:requestId
 $request=@{id=$id;method=$Method;params=$Params};if($Session){$request.sessionId=$Session}
 $bytes=[Text.Encoding]::UTF8.GetBytes(($request|ConvertTo-Json -Depth 12 -Compress))
 $cts=[Threading.CancellationTokenSource]::new(7000)
 try{
  $socket.SendAsync([ArraySegment[byte]]::new($bytes),[Net.WebSockets.WebSocketMessageType]::Text,$true,$cts.Token).GetAwaiter().GetResult()
  while($true){
   $stream=[IO.MemoryStream]::new()
   try{do{$buffer=New-Object byte[] 16384;$message=$socket.ReceiveAsync([ArraySegment[byte]]::new($buffer),$cts.Token).GetAwaiter().GetResult();if($message.MessageType -eq [Net.WebSockets.WebSocketMessageType]::Close){throw 'CDP closed'};$stream.Write($buffer,0,$message.Count);if($stream.Length -gt 1048576){throw 'CDP bounded frame exceeded'}}while(!$message.EndOfMessage);$response=[Text.Encoding]::UTF8.GetString($stream.ToArray())|ConvertFrom-Json}finally{$stream.Dispose()}
   Observe-CDP $response
   if((Get-SimValue $response 'id' 0) -eq $id){if(Test-SimHas $response 'error'){throw "CDP/INSTRUMENTATION_FAILURE $Method"};return $response.result}
  }
 }finally{$cts.Dispose()}
}
function Drain-Capture($State,[int]$Stage){
 for($batch=0;$batch -lt 64;$batch++){
  if($script:pages -ge 1024){throw 'Capture disk/page bound reached: incomplete evidence'}
  $page=Service 'capture.changes' @{cursor=$State.Cursor};Save-Proof $page ('stage-{0}-page-{1:D4}.json' -f $Stage,$script:pages);$script:pages++
  if(!(Merge-SimPage $State $page)){return}
 }
 throw 'Capture did not drain: no complete/no-leak claim allowed'
}
try{
 $policy=(Get-ItemProperty -LiteralPath 'HKCU:\Software\Policies\Google\Chrome' -Name ProxySettings).ProxySettings|ConvertFrom-Json
 if($policy.ProxyMode -ne 'pac_script' -or !(Get-SimValue $policy 'ProxyPacMandatory' $false)){throw 'Actual mandatory Chrome PAC policy required; activate Chrome AUTO in the GUI first'}
 $status=Service 'status';$browser=Service 'browser.status'
 $profile=@($status.profiles|Where-Object{$_.mode -eq 'AUTO' -and @($_.executables|Where-Object{$_ -ieq $ChromePath}).Count -gt 0})
 if($profile.Count -ne 1 -or !$browser.listening -or @((Get-SimValue $browser 'ready_profiles' @())|Where-Object{$_ -eq $profile[0].id}).Count -ne 1){throw 'Verified Chrome AUTO transport/profile missing; activate through GUI, not Optimize alone'}
 Save-Proof @{mode=$policy.ProxyMode;mandatory=$policy.ProxyPacMandatory} 'policy.json'
 $data=Join-Path $run 'isolated-chrome'
 if($data.Contains('"')){throw 'Invalid report directory'}
 # No proxy flags, no --no-sandbox, no ECH/QUIC disable, no certificate exception.
 # The real current-user Chrome policy must route this isolated profile.
 $client=Start-Process -FilePath $ChromePath -ArgumentList "--headless=new --no-first-run --no-default-browser-check --user-data-dir=`"$data`" --remote-debugging-address=127.0.0.1 --remote-debugging-port=0 about:blank" -PassThru
 $portFile=Join-Path $data 'DevToolsActivePort';$deadline=[DateTime]::UtcNow.AddSeconds(25)
 while(!(Test-Path $portFile)){if($client.HasExited -or [DateTime]::UtcNow -ge $deadline){throw 'CDP/INSTRUMENTATION_FAILURE: Chrome startup'};Start-Sleep -Milliseconds 200}
 $portInfo=Get-Content -LiteralPath $portFile;$port=[int]$portInfo[0]
 $socket=[Net.WebSockets.ClientWebSocket]::new();$cts=[Threading.CancellationTokenSource]::new(7000)
 try{$socket.ConnectAsync([Uri]("ws://127.0.0.1:"+$port+$portInfo[1]),$cts.Token).GetAwaiter().GetResult()}finally{$cts.Dispose()}
 $sessions=@()
 foreach($url in $urls){$target=CDP 'Target.createTarget' @{url='about:blank'};$attached=CDP 'Target.attachToTarget' @{targetId=$target.targetId;flatten=$true};$session=[string]$attached.sessionId;$sessions+=$session;$frame=(CDP 'Page.getFrameTree' @{} $session).frameTree.frame.id;$script:targets[$session]=@{site=$url;frame=[string]$frame;document=$null;failures=@();requests=@{}};[void](CDP 'Network.enable' @{} $session);[void](CDP 'Network.setCacheDisabled' @{cacheDisabled=$true} $session)}
 $stages=if($RestartService){2}else{1}
 for($stage=1;$stage -le $stages;$stage++){
  if($stage -gt 1){
   Restart-Service neexitlag
   $deadline=[DateTime]::UtcNow.AddSeconds(25)
   do{Start-Sleep -Milliseconds 300;$browser=Service 'browser.status'}while(!$browser.listening -and [DateTime]::UtcNow -lt $deadline)
   if(!$browser.listening){throw 'Chrome transport did not return after service restart'}
  }
  $state=@{Cursor=(Service 'capture.changes.marker');Flows=@{};Lost=$false}
  $before=Service 'quality.diagnostics';$records=@{}
  foreach($session in $sessions){$row=$script:targets[$session];$row.document=$null;$row.failures=@();$row.requests=@{};$nav=CDP 'Page.navigate' @{url=$row.site} $session;$row.frame=[string]$nav.frameId}
  $watch=[Diagnostics.Stopwatch]::StartNew()
  while($watch.Elapsed.TotalSeconds -lt $ObserveSeconds){
   foreach($session in $sessions){[void](CDP 'Runtime.evaluate' @{expression='0';returnByValue=$true} $session)}
   Drain-Capture $state $stage
   $browser=Service 'browser.status'
   if($browser.truncated){throw 'Browser IPC snapshot truncated; no complete route claim'}
   foreach($record in $browser.records){$records[[string]$record.id]=$record}
   Start-Sleep -Milliseconds 300
  }
  Drain-Capture $state $stage
  # Existing process discovery, including parent birth identity; no GUI scanner.
  $birth=[uint64]$client.StartTime.ToUniversalTime().ToFileTimeUtc()
  $processes=@(Service 'processes' @{root_pid=$client.Id;root_creation_time=$birth;executables=@($ChromePath)})
  $identities=@{};$owned=@{}
  foreach($p in $processes){$identities[[string]$p.PID]=$p;$owned[[string]$p.PID]=$true}
  if(!$owned.ContainsKey([string]$client.Id)){throw 'Exact diagnostic Chrome process family missing'}
  $ours=@($records.Values|Where-Object{$owned.ContainsKey([string]$_.owner.pid) -and $identities.ContainsKey([string]$_.owner.pid) -and $_.owner.creation_time -eq $identities[[string]$_.owner.pid].CreationTime})
  $observations=@();$h2=$false
  foreach($session in $sessions){
   $row=$script:targets[$session];$uri=[Uri]$row.site;$domain=$uri.DnsSafeHost
   $routes=@($ours|Where-Object{$_.domain -eq $domain -and $_.sent_bytes -gt 0 -and $_.received_bytes -gt 0 -and $_.result -notin @('ORIGIN_PLAN_FAILED','UPSTREAM_OPEN_FAILED')})
   $dom=CDP 'Runtime.evaluate' @{expression='({ready:document.readyState,characters:document.body?document.body.innerText.length:0})';returnByValue=$true} $session
   $ok=$null -ne $row.document -and $row.document.finished -and $row.document.status -ge 200 -and $row.document.status -lt 400 -and $row.failures.Count -eq 0 -and $dom.result.value.characters -ge 20 -and $routes.Count -gt 0
   if($null -ne $row.document -and $row.document.protocol -eq 'h2'){$h2=$true}
   if($domain -eq 'chatgpt.com'){$ok=$ok -and @($routes|Where-Object{$_.plan.policy.action -eq 'RELAY_REQUIRED' -and $_.plan.policy.relay_id -eq 'VPS-01'}).Count -gt 0}
   if($domain -eq 'www.youtube.com'){$ok=$ok -and @($routes|Where-Object{$_.plan.policy.action -eq 'DIRECT_PLUS' -and $_.direct_plus_applied}).Count -gt 0}
   $observations+=@{site=$row.site;pass=$ok;document=$row.document;dom_metadata=$dom.result.value;failures=$row.failures;routes=@($routes|Select-Object domain,owner,source,plan,result,sent_bytes,received_bytes,direct_plus_applied)}
  }
  $leaks=@($state.Flows.Values|Where-Object{$owned.ContainsKey([string]$_.pid) -and $_.policy.action -eq 'RELAY_REQUIRED' -and $_.counters.direct_bytes -gt 0})
  $result=@{stage=$stage;chrome_root_pid=$client.Id;processes=$processes;sites=$observations;evidence_lost=$state.Lost;http2_observed=$h2;protected_direct_leaks=$leaks.Count;pass=(!$state.Lost -and !$script:eventLoss -and $leaks.Count -eq 0 -and $h2 -and @($observations|Where-Object{!$_.pass}).Count -eq 0)}
  $drops=[pscustomobject]@{transport_evidence_pass=$true}
  Add-SimDropGates $drops $before (Service 'quality.diagnostics')
  $result.capture_queue_drops_delta=$drops.capture_queue_drops_delta;$result.worker_saturation_drops_delta=$drops.worker_saturation_drops_delta
  $result.pass=$result.pass -and $result.capture_queue_drops_delta -eq 0 -and $result.worker_saturation_drops_delta -eq 0
  $proof.stages+=,$result;Save-Proof $result "stage-$stage.json"
  if(!$result.pass){throw 'Public Chrome acceptance FAIL: inspect per-site HTTP/route/failure evidence; no page PASS inferred from CONNECT/HEAD'}
 }
 $proof.pass=$true
 Write-Host "Public Chrome HTTP/routes PASS. Evidence: $run"
}catch{$proof.error=$_.Exception.Message;throw}finally{
 try{Save-Proof (Service 'browser.status') 'browser-final.json';Save-Proof (Service 'status') 'service-final.json'}catch{}
 Save-Proof $proof 'proof.json'
 if($null -ne $socket){try{[void](CDP 'Browser.close' @{})}catch{};$socket.Dispose()}
 # Only the temporary diagnostic browser, never an existing user browser.
 if($null -ne $client){if(!$client.WaitForExit(5000)){& taskkill /PID $client.Id /T /F 2>$null|Out-Null}}
}
