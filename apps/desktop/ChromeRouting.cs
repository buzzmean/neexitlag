using System.Diagnostics;
using System.Windows;
namespace Neexitlag.Desktop;
public partial class MainWindow {
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
   if(!status.ReadyProfiles.Contains(row.Profile.Id)){await ActivateChromeSites(row);if(!model.Optimizing)return;}
  }
  await model.OptimizeAsync(true,lifetime.Token);
 }
 async Task ActivateChromeSites(ApplicationRow row){
  var exe=row.Profile.Executables.FirstOrDefault(p=>p.Replace('\\','/').Split('/').Last().Equals("chrome.exe",StringComparison.OrdinalIgnoreCase));
  if(exe is null||!Path.IsPathFullyQualified(exe)||!File.Exists(exe))throw new InvalidOperationException(Text.Get("SelectChromeAuto"));
  ChromePolicy.Check(); // Refuse managed/foreign settings before asking Chrome to quit.
  async Task<bool> Running(){
   var identities=await model.ChromeProcessesAsync(exe,lifetime.Token);
   foreach(var id in identities){try{using var process=Process.GetProcessById((int)id.PID);if(process.SessionId==Process.GetCurrentProcess().SessionId)return true;}catch(ArgumentException){}}
   return false;
  }
  if(await Running()){
   if(System.Windows.MessageBox.Show(this,Text.Get("ChromeSitesRestart"),"NeExitLag",MessageBoxButton.YesNo,MessageBoxImage.Information)!=MessageBoxResult.Yes)return;
   // Chrome's own quit path preserves its profile/session. Never kill Chrome or
   // copy/remove user data. A cancelled/blocked quit refuses activation.
   var quit=new ProcessStartInfo(exe){UseShellExecute=false};quit.ArgumentList.Add("chrome://quit");using var started=Process.Start(quit);
   var deadline=DateTime.UtcNow.AddSeconds(15);while(await Running()&&DateTime.UtcNow<deadline)await Task.Delay(200,lifetime.Token);
   if(await Running())throw new InvalidOperationException(Text.Get("ChromeSitesStillRunning"));
  }
  var original=row.Profile.Copy();
  var directory=Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData),"Google","Chrome","User Data");
  BrowserLaunch? launch=null;bool installed=false;
  try{
   ChromePolicy.Guard(original);installed=true;
   // Prepare binds the listener without changing routing, profiles or Optimize.
   launch=await model.PrepareChromeAsync(original,directory,lifetime.Token);
   ChromePolicy.Install(launch.Arguments);installed=true;
   var info=new ProcessStartInfo(exe){UseShellExecute=false};foreach(var value in launch.Arguments)info.ArgumentList.Add(value);info.ArgumentList.Add("--restore-last-session");info.ArgumentList.Add(launch.HealthUrl);
   using var process=Process.Start(info)??throw new ChromeActivationException(Text.Get("ChromeActivationFailed"));
   await model.VerifyChromeAsync(original,launch,lifetime.Token);
   await model.CommitChromeAsync(original,lifetime.Token);
   model.SetNotice(Text.Get("ChromeSitesLaunched"));
  }catch(Exception error){
   // Keep a mandatory protected-domain PAC guard. Removing PAC altogether can
   // leak OpenAI while capture is unavailable; recovery only releases ordinary
   // origins. Other applications and global Optimize are never disabled here.
   if(!installed)throw new ChromeActivationException(Text.Get("ChromeGuardFailed"));
   ChromePolicy.Guard(original);
   try{await model.RestoreChromeAsync(original,lifetime.Token);}catch(Exception restoreError){throw new ChromeActivationException(Text.Get("ChromeRollbackFailed")+" "+restoreError.Message);}
   if(!await Running()){var recovery=new ProcessStartInfo(exe){UseShellExecute=false};recovery.ArgumentList.Add("--user-data-dir="+directory);recovery.ArgumentList.Add("--restore-last-session");using var reopened=Process.Start(recovery);}
   throw new ChromeActivationException(Text.Get("ChromeActivationFailed")+" "+error.Message);
  }
 }
}
