# A real local Chromium instrumentation regression, NOT external relay/ISP E2E.
# Verify CDP connectionId -> TCP_CONNECT -> exact local source port using an
# independently observed server peer endpoint. No service, driver or TLS bypass.
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot '..\..\tools\windows\strategy-v2-checks.ps1')
function Check($Value,[string]$Message){if(!$Value){throw $Message}}
$choices=@((Join-Path $env:ProgramFiles 'Google\Chrome\Application\chrome.exe'),(Join-Path ${env:ProgramFiles(x86)} 'Microsoft\Edge\Application\msedge.exe'),(Join-Path $env:ProgramFiles 'Microsoft\Edge\Application\msedge.exe'))
$exe=@($choices|Where-Object{Test-Path -LiteralPath $_}|Select-Object -First 1)
if(!$exe.Count){throw 'Chromium/Edge required for exact socket metadata regression'}
$root=Join-Path ([IO.Path]::GetTempPath()) ('v2-socket-binding-'+[Guid]::NewGuid().ToString('N'));[void](New-Item -ItemType Directory $root)
$data=Join-Path $root 'browser';[void](New-Item -ItemType Directory $data);$netlog=Join-Path $data '.private-sockets.json';$ready=Join-Path $root 'listen';$peer=Join-Path $root 'peer'
$browser=$null;$job=$null;$idle=$null;$script:socket=$null
try{
 $job=Start-Job -ArgumentList $ready,$peer -ScriptBlock {
  param($Ready,$Peer)
  $listener=[Net.Sockets.TcpListener]::new([Net.IPAddress]::Loopback,0);$listener.Start()
  $clients=@();$complete=$false;$accepted=0
  try{
   [IO.File]::WriteAllText($Ready,[string]$listener.LocalEndpoint.Port)
   $deadline=[DateTime]::UtcNow.AddSeconds(40)
   # Chromium may open an idle speculative socket before its request socket.
   # Poll a bounded set instead of blocking on the first accepted connection.
   while(!$complete -and [DateTime]::UtcNow -lt $deadline){
    while($listener.Pending()){
     if(++$accepted -gt 16){throw 'Local metadata server connection bound exceeded'}
     $client=$listener.AcceptTcpClient();$clients+=@{client=$client;header=''}
    }
    foreach($entry in $clients){
     $client=$entry.client;$stream=$client.GetStream()
     if(!$stream.DataAvailable){continue}
     $buf=New-Object byte[] 1024;$n=$stream.Read($buf,0,$buf.Length)
     $entry.header+=[Text.Encoding]::ASCII.GetString($buf,0,$n)
     if($entry.header.Length -gt 16384){throw 'Local metadata request header bound exceeded'}
     if(!$entry.header.Contains("`r`n`r`n")){continue}
     [IO.File]::WriteAllText($Peer,$client.Client.RemoteEndPoint.ToString())
     $reply=[Text.Encoding]::ASCII.GetBytes("HTTP/1.1 200 OK`r`nContent-Length: 2`r`nContent-Type: text/plain`r`nConnection: close`r`n`r`nok")
     $stream.Write($reply,0,$reply.Length);$stream.Flush();$client.Close();$complete=$true;break
    }
    if(!$complete){Start-Sleep -Milliseconds 10}
   }
   if(!$complete){throw ('No complete local HTTP request; accepted='+$accepted)}
  }finally{foreach($entry in $clients){$entry.client.Close()};$listener.Stop()}
 }
 $deadline=[DateTime]::UtcNow.AddSeconds(20)
 while(!(Test-Path $ready)){if([DateTime]::UtcNow -gt $deadline){throw 'Local metadata server readiness timeout'};Start-Sleep -Milliseconds 100}
 $port=[int](Get-Content $ready -Raw);$url="http://127.0.0.1:$port/"
 # Deterministically keep the first connection idle, as speculative preconnect can.
 $idle=[Net.Sockets.TcpClient]::new();$idle.Connect([Net.IPAddress]::Loopback,$port)
 $arguments="--headless=new --no-first-run --no-default-browser-check --disable-background-networking --disable-quic --user-data-dir=`"$data`" --remote-debugging-address=127.0.0.1 --remote-debugging-port=0 --log-net-log=`"$netlog`" --net-log-capture-mode=Default --net-log-max-size-mb=4 about:blank"
 $browser=Start-Process -FilePath $exe[0] -ArgumentList $arguments -PassThru -RedirectStandardOutput (Join-Path $root 'stdout') -RedirectStandardError (Join-Path $root 'stderr');[void]$browser.Handle
 $deadline=[DateTime]::UtcNow.AddSeconds(20);$pages=@();$debugPort=0
 while(!$pages.Count){
  if($browser.HasExited){throw 'Isolated test browser exited before CDP'}
  if([DateTime]::UtcNow -gt $deadline){throw 'Isolated test browser CDP timeout'}
  $portFile=Join-Path $data 'DevToolsActivePort'
  if(Test-Path $portFile){$lines=@(Get-Content $portFile);if($lines.Count){[void][int]::TryParse($lines[0],[ref]$debugPort)}}
  if($debugPort -gt 0){try{$list=Invoke-RestMethod "http://127.0.0.1:$debugPort/json/list" -TimeoutSec 2;$pages=@($list|Where-Object{$_.type -eq 'page'})}catch{}}
  Start-Sleep -Milliseconds 100
 }
 $source=Join-Path $PSScriptRoot '..\..\tools\windows\strategy-tester-v2.ps1';$tokens=$null;$errors=$null
 $ast=[Management.Automation.Language.Parser]::ParseFile((Resolve-Path $source),[ref]$tokens,[ref]$errors)
 $cdp=$ast.Find({param($a)$a -is [Management.Automation.Language.FunctionDefinitionAst] -and $a.Name -eq 'CDP'},$true);Invoke-Expression $cdp.Extent.Text
 $script:cdpID=0;$script:navID=-2;$script:navError='';$script:responses=@{};$script:finished=@{};$script:failed=@{}
 $script:socket=[Net.WebSockets.ClientWebSocket]::new();$cancel=[Threading.CancellationTokenSource]::new(5000)
 try{[void]$script:socket.ConnectAsync([Uri]$pages[0].webSocketDebuggerUrl,$cancel.Token).GetAwaiter().GetResult()}finally{$cancel.Dispose()}
 [void](CDP 'Network.enable' @{});[void](CDP 'Network.setCacheDisabled' @{cacheDisabled=$true});[void](CDP 'Network.setBypassServiceWorker' @{bypass=$true})
 $main=Get-V2MainFrame (CDP 'Page.getFrameTree' @{});$script:navID=$script:cdpID+1;CDP 'Page.navigate' @{url=$url} -NoWait
 $response=$null;$deadline=[DateTime]::UtcNow.AddSeconds(15)
 while($null -eq $response){
  [void](CDP 'Browser.getVersion' @{});Assert-V2SocketLogBound $netlog
  foreach($id in @($script:responses.Keys)){$entry=$script:responses[$id];if($entry.frame_id -eq $main -and $entry.response.url -eq $url -and $script:finished.ContainsKey($id)){$response=$entry.response;break}}
  if([DateTime]::UtcNow -gt $deadline){throw ('Local completed Document response timeout: nav='+$script:navError+'; responses='+$script:responses.Count+'; finished='+$script:finished.Count+'; failures='+($script:failed.Values -join ',')+'; server='+$job.State+'; server_errors='+(@($job.ChildJobs|ForEach-Object{$_.Error}) -join '; '))}
  Start-Sleep -Milliseconds 50
 }
 Check ($response.status -eq 200 -and $response.connectionId -gt 0) 'No real successful CDP connection identity'
 CDP 'Browser.close' @{} -NoWait
 Check ($browser.WaitForExit(10000)) 'Browser did not flush metadata on graceful close';[void]$browser.WaitForExit()
 $binding=Read-V2SocketBinding $netlog ([long]$response.connectionId) ("127.0.0.1:"+$port)
 Check ($binding.status -eq 'BOUND') ('Exact socket binding unavailable: '+$binding.reason)
 Check ($binding.source -eq (Get-Content $peer -Raw).Trim()) 'CDP socket source differs from actual server-observed peer'
 Write-Host ('PASS real Chromium connectionId={0} -> source={1} -> target={2}; independently verified server peer' -f $binding.connection_id,$binding.source,$binding.destination)
}finally{
 [void](Stop-V2TrialBrowser $browser $data)
 if($idle){$idle.Dispose()}
 if($job){Stop-Job $job -ErrorAction SilentlyContinue;Remove-Job $job -Force -ErrorAction SilentlyContinue}
 if(Test-Path $root){Remove-Item $root -Force -Recurse -ErrorAction SilentlyContinue}
}
