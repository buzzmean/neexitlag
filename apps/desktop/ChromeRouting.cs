using System.Diagnostics;
using System.Text.Json;
using System.Windows;
namespace Neexitlag.Desktop;
public partial class MainWindow {
 readonly ChromeRecovery chromeRecovery=new();
 bool chromeActivating,chromeMonitoring;
 static bool IsChrome(Profile profile)=>profile.Executables.Any(p=>p.Replace('\\','/').Split('/').Last().Equals("chrome.exe",StringComparison.OrdinalIgnoreCase));
 async void DisableChromeSites(object sender,RoutedEventArgs e)=>await Execute(async()=>{
  var profiles=model.Applications.Where(r=>r.Profile.Mode=="AUTO"&&IsChrome(r.Profile)).Select(r=>r.Profile.Copy()).ToArray();
  if(profiles.Length==0)ChromePolicy.Recover();
  else{ChromePolicy.Guard(profiles[0]);foreach(var profile in profiles)await model.RestoreChromeAsync(profile,lifetime.Token);}
  model.SetNotice(Text.Get("ChromeSitesRemoved"));
 });
 async void OpenChromeSites(object sender,RoutedEventArgs e){
  if(AppGrid.SelectedItem is not ApplicationRow row){ShowMessage("SelectApplication");return;}
  await Execute(()=>ActivateChromeSites(row));
 }
 async Task EnableOptimizationAsync(){
  var status=await model.ChromeStatusAsync(lifetime.Token);
  foreach(var row in model.Applications.Where(r=>r.Profile.Mode=="AUTO"&&IsChrome(r.Profile)).ToArray()){
   if(status.RecoveryAddress==""||!status.ReadyProfiles.Contains(row.Profile.Id)){await ActivateChromeSites(row);if(!model.Optimizing)return;}
  }
  await model.OptimizeAsync(true,lifetime.Token);
 }
 async Task ProbeChromePolicyAsync(string exe,string healthUrl){
  var directory=Path.Combine(Path.GetTempPath(),"neexitlag-chrome-readiness-"+Guid.NewGuid().ToString("N"));
  Directory.CreateDirectory(directory);
  var info=new ProcessStartInfo(exe){UseShellExecute=false,CreateNoWindow=true,RedirectStandardOutput=true,RedirectStandardError=true};
  foreach(var arg in new[]{"--headless=new","--dump-dom","--timeout=15000","--no-first-run","--user-data-dir="+directory,healthUrl})info.ArgumentList.Add(arg);
  using var process=Process.Start(info)??throw new ChromeActivationException(Text.Get("ChromeActivationFailed"));
  using var deadline=CancellationTokenSource.CreateLinkedTokenSource(lifetime.Token);deadline.CancelAfter(TimeSpan.FromSeconds(20));
  var stdout=ReadAutoOutput(process.StandardOutput,deadline.Token);var stderr=ReadAutoOutput(process.StandardError,deadline.Token);
  try{
   await process.WaitForExitAsync(deadline.Token);
   if(process.ExitCode!=0)throw new ChromeActivationException(Text.Get("ChromeActivationFailed"));
   // Readiness is certified by browser.confirm's server-side nonce, not logs.
   // A descendant retaining a redirected handle must not keep activation open
   // after the actual helper Chrome exited with code 0.
   deadline.Cancel();
   try{await Task.WhenAll(stdout,stderr);}catch(OperationCanceledException)when(process.HasExited&&process.ExitCode==0){}
  }finally{
   deadline.Cancel(); // Also release diagnostic reads after non-zero exit/cancellation.
   // Own disposable empty health profile ONLY, never the user's Chrome/profile.
   if(!process.HasExited){process.Kill(entireProcessTree:true);await process.WaitForExitAsync(CancellationToken.None);}
   try{Directory.Delete(directory,true);}catch(IOException){}catch(UnauthorizedAccessException){}
  }
 }
 async Task ActivateChromeSites(ApplicationRow row){
  if(chromeActivating)return;chromeActivating=true;
  try{await ActivateChromeSitesCore(row);}finally{chromeActivating=false;}
 }
 async Task MonitorChromeTransportAsync(bool wasOptimizing){
  if(chromeActivating||chromeMonitoring||!ChromePolicy.Monitored)return;
  chromeMonitoring=true;
  try{
   try{await chromeRecovery.StartAsync(lifetime.Token);model.ApplyRecovery(await chromeRecovery.SnapshotAsync(lifetime.Token));}
   catch(Exception error)when(!lifetime.IsCancellationRequested && error is (IOException or System.ComponentModel.Win32Exception or UnauthorizedAccessException or OperationCanceledException or ChromeActivationException or JsonException)){model.ApplyRecovery(new());model.SetNotice(Text.Get("ChromeRecoveryUnavailable"));}
   if(!model.Available){ChromePolicy.Suspend(wasOptimizing);return;}
   if(!ChromePolicy.Applied){
    foreach(var p in model.Applications.Where(r=>r.Profile.Mode=="AUTO"&&IsChrome(r.Profile)).Select(r=>r.Profile.Copy()).ToArray()){ChromePolicy.Guard(p);await model.RestoreChromeAsync(p,lifetime.Token);}
    model.SetNotice(Text.Get("ChromeActivationFailed"));return;
   }
   var status=await model.ChromeStatusAsync(lifetime.Token);
   if(!status.Listening){
    ChromePolicy.Suspend(model.Optimizing);
    // The stable PAC already has the owner-scoped recovery CONNECT endpoint.
    // Native capture never sees opaque ordinary bypass traffic; guarded origins
    // have no secondary endpoint. Do not toggle policy or disable other apps.
    model.SetNotice(Text.Get("ChromeRecoveryActive"));
   }else if(ChromePolicy.Suspended){
    var resume=ChromePolicy.Resume();
    if(resume&&!model.Optimizing)await model.RestoreOptimizationAsync(true,lifetime.Token);
   }else{ChromePolicy.RecordOptimization(model.Optimizing);}
  }finally{chromeMonitoring=false;}
 }
 async Task ActivateChromeSitesCore(ApplicationRow row){
  var exe=row.Profile.Executables.FirstOrDefault(p=>p.Replace('\\','/').Split('/').Last().Equals("chrome.exe",StringComparison.OrdinalIgnoreCase));
  if(exe is null||!Path.IsPathFullyQualified(exe)||!File.Exists(exe))throw new InvalidOperationException(Text.Get("SelectChromeAuto"));
  ChromePolicy.Check(); // Refuse managed/foreign settings before asking Chrome to quit.
  async Task<bool> Running(){
   var identities=await model.ChromeProcessesAsync(exe,lifetime.Token);
   foreach(var id in identities){try{using var process=Process.GetProcessById((int)id.PID);if(process.SessionId==Process.GetCurrentProcess().SessionId)return true;}catch(ArgumentException){}}
   return false;
  }
  if(await Running()){
   if(System.Windows.MessageBox.Show(this,Text.Get("ChromeSitesRestart"),"VELIQ",MessageBoxButton.YesNo,MessageBoxImage.Information)!=MessageBoxResult.Yes)return;
   // Chrome's own quit path preserves its profile/session. Never kill Chrome or
   // copy/remove user data. A cancelled/blocked quit refuses activation.
   var quit=new ProcessStartInfo(exe){UseShellExecute=false};quit.ArgumentList.Add("chrome://quit");using var started=Process.Start(quit);
   var deadline=DateTime.UtcNow.AddSeconds(15);while(await Running()&&DateTime.UtcNow<deadline)await Task.Delay(200,lifetime.Token);
   if(await Running())throw new InvalidOperationException(Text.Get("ChromeSitesStillRunning"));
  }
  var original=row.Profile.Copy();var wasOptimizing=model.Optimizing;
  var directory=Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData),"Google","Chrome","User Data");
  BrowserLaunch? launch=null;bool installed=false;
  try{
   ChromePolicy.Guard(original);installed=true;
   // Prepare binds the listener without changing routing, profiles or Optimize.
   await chromeRecovery.StartAsync(lifetime.Token);
   launch=await model.PrepareChromeAsync(original,directory,lifetime.Token,chromeRecovery.Address);
   launch.RecoverySettings.Executable=exe;
   await chromeRecovery.ConfigureAsync(launch.RecoverySettings,lifetime.Token);
   ChromePolicy.Install(launch.Arguments,launch.RecoveryArguments);installed=true;
   // Never restore real user tabs before AUTO is committed: those CONNECTs
   // would race a still-DIRECT profile and get 502 despite a good readiness tab.
   // An empty, isolated headless profile checks the ACTUAL HKCU policy first.
   await model.ActivatePreparedChromeAsync(original,launch,
    ()=>ProbeChromePolicyAsync(exe,launch.HealthUrl),
    ()=>{var info=new ProcessStartInfo(exe){UseShellExecute=false};foreach(var value in launch.UserArguments())info.ArgumentList.Add(value);info.ArgumentList.Add("--restore-last-session");using var process=Process.Start(info)??throw new ChromeActivationException(Text.Get("ChromeActivationFailed"));return Task.CompletedTask;},lifetime.Token);
   model.SetNotice(Text.Get("ChromeSitesLaunched"));
  }catch(Exception error){
   // Keep a mandatory protected-domain PAC guard. Removing PAC altogether can
   // leak OpenAI while capture is unavailable; recovery only releases ordinary
   // origins. Other applications and global Optimize are never disabled here.
   if(!installed)throw new ChromeActivationException(Text.Get("ChromeGuardFailed"));
   ChromePolicy.Guard(original);
   try{await model.RestoreChromeAsync(original,lifetime.Token);if(!wasOptimizing&&model.Optimizing)await model.RestoreOptimizationAsync(false,lifetime.Token);}catch(Exception restoreError){throw new ChromeActivationException(Text.Get("ChromeRollbackFailed")+" "+restoreError.Message);}
   if(!await Running()){var recovery=new ProcessStartInfo(exe){UseShellExecute=false};recovery.ArgumentList.Add("--user-data-dir="+directory);recovery.ArgumentList.Add("--restore-last-session");using var reopened=Process.Start(recovery);}
   throw new ChromeActivationException(Text.Get("ChromeActivationFailed")+" "+error.Message);
  }
 }
}
