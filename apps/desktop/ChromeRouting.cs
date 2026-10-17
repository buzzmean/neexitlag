using System.Diagnostics;
using System.Windows;
namespace Neexitlag.Desktop;
public partial class MainWindow {
 void DisableChromeSites(object sender,RoutedEventArgs e){try{ChromePolicy.Remove();model.SetNotice(Text.Get("ChromeSitesRemoved"));}catch{ShowMessage("ChromeProxyManaged");}}
 async void OpenChromeSites(object sender,RoutedEventArgs e){
  if(AppGrid.SelectedItem is not ApplicationRow row){ShowMessage("SelectApplication");return;}
  await Execute(()=>ActivateChromeSites(row));
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
  await model.SetProfileAsync(row.Profile,"AUTO",row.Profile.RelayId,lifetime.Token);
  await model.OptimizeAsync(true,lifetime.Token);
  var profile=model.Applications.First(p=>p.Profile.Id==row.Profile.Id).Profile;
  var directory=Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData),"Google","Chrome","User Data");
  var launch=await model.PrepareChromeAsync(profile,directory,lifetime.Token);
  ChromePolicy.Install(launch.Arguments);
  var info=new ProcessStartInfo(exe){UseShellExecute=false};foreach(var value in launch.Arguments)info.ArgumentList.Add(value);info.ArgumentList.Add("--restore-last-session");using var process=Process.Start(info)??throw new InvalidOperationException(Text.Get("AutoStartFailed"));
  model.SetNotice(Text.Get("ChromeSitesLaunched"));
 }
}
