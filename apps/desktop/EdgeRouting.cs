using System.Diagnostics;
using System.Text.Json;
using System.Windows;
namespace Neexitlag.Desktop;

public partial class MainWindow {
 readonly EdgeRoutingRecovery edgeRecovery=new();
 bool edgeActivating,edgeMonitoring,edgeLifetimeHook,edgeNeedsActivation;
 static bool EdgeSites(Profile p)=>BrowserCatalog.IsEdge(p)&&p.PolicyData is {} data&&data.TryGetValue("app_preset",out var v)&&v.ValueKind==JsonValueKind.String&&v.GetString()=="edge-sites";
 async Task EnsureEdgeCoverageAsync(IReadOnlyList<InstalledBrowser> installed){
  if(!edgeLifetimeHook){Closed+=(_,_)=>edgeRecovery.Dispose();edgeLifetimeHook=true;}
  try{
   var created=await model.EnsureEdgeProfileAsync(installed,lifetime.Token);
   if(created is not null){await ActivateEdgeSites(created);return;}
   // Existing MANUAL/DIRECT/custom profiles remain unchanged. Previously opted-in
   // AUTO resumes only after its own policy/readiness has been verified.
   foreach(var row in model.Applications.Where(r=>EdgeSites(r.Profile)&&r.Profile.Mode=="AUTO").ToArray()){
    var status=await model.EdgeStatusAsync(lifetime.Token);
    await edgeRecovery.StartAsync(lifetime.Token);
    if(!edgeRecovery.Configured||status.RecoveryAddress!=edgeRecovery.Address||!status.Listening||!status.PolicyReady||!status.ReadyProfiles.Contains(row.Profile.Id))await ActivateEdgeSites(row.Profile.Copy());
   }
  }catch(Exception error)when(!lifetime.IsCancellationRequested){model.SetNotice("Edge AUTO не включён: "+error.Message);}
 }
 async void OpenEdgeSites(object sender,RoutedEventArgs args)=>await Execute(async()=>{
  if(AppGrid.SelectedItem is not ApplicationRow row||!BrowserCatalog.IsEdge(row.Profile))throw new InvalidOperationException("Выберите Microsoft Edge в списке приложений");
  if(!await ActivateEdgeSites(row.Profile.Copy()))return;
  // The regular optimization transaction still validates BOTH browser lanes.
  await model.OptimizeAsync(true,lifetime.Token);
 });
 async void DisableEdgeSites(object sender,RoutedEventArgs args)=>await Execute(async()=>{
  var profiles=model.Applications.Where(r=>EdgeSites(r.Profile)).Select(r=>r.Profile.Copy()).ToArray();
  if(profiles.Length==0){model.SetNotice("Профиль Edge-полосы не найден; PAC не изменялся.");return;}
  foreach(var p in profiles)await model.RestoreEdgeAsync(p,lifetime.Token);
  edgeRecovery.Dispose();edgeNeedsActivation=true;
  model.SetNotice("PAC Edge отключён для обычных сайтов. Защищённые домены остаются закрыты; Chrome не изменён. Перезапустите Edge для применения политики.");
 });
 async Task<bool> EdgeRunning(string exe){
  foreach(var id in await model.ChromeProcessesAsync(exe,lifetime.Token)){
   try{using var p=Process.GetProcessById((int)id.PID);if(p.SessionId==Process.GetCurrentProcess().SessionId)return true;}catch(ArgumentException){}
  }return false;
 }
 async Task<bool> ActivateEdgeSites(Profile original){
  if(edgeActivating)throw new InvalidOperationException("Активация Edge уже выполняется");edgeActivating=true;
  try{return await ActivateEdgeSitesCore(original);}finally{edgeActivating=false;}
 }
 async Task<bool> ActivateEdgeSitesCore(Profile original){
  var exe=original.Executables.FirstOrDefault(p=>p.Replace('\\','/').Split('/').Last().Equals("msedge.exe",StringComparison.OrdinalIgnoreCase));
  if(exe is null||!Path.IsPathFullyQualified(exe)||!File.Exists(exe))throw new InvalidOperationException("Требуется полный путь установленного msedge.exe");
  if(original.Mode is not ("DIRECT" or "AUTO"))throw new InvalidOperationException("Для Edge-полосы выберите DIRECT или AUTO; существующий профиль не изменён.");
  var directory=Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData),"Microsoft","Edge","User Data");
  try{
  // SYSTEM refuses foreign policy and installs a mandatory protected-origin guard
  // before asking the user to close Edge. The GUI never writes policy registry.
  if(await EdgeRunning(exe)){
   if(System.Windows.MessageBox.Show(this,"Для включения маршрутов по сайтам нужно закрыть и снова открыть Edge. В браузере появится «Управляется вашей организацией»: VELIQ задаёт только PAC Edge. Продолжить?","VELIQ",MessageBoxButton.YesNo,MessageBoxImage.Information)!=MessageBoxResult.Yes)return false;
   await model.EdgePolicyAsync(original,"guard",lifetime.Token);
   var quit=new ProcessStartInfo(exe){UseShellExecute=false};quit.ArgumentList.Add("edge://quit");using var launched=Process.Start(quit);
   var deadline=DateTime.UtcNow.AddSeconds(15);
   while(await EdgeRunning(exe)&&DateTime.UtcNow<deadline)await Task.Delay(200,lifetime.Token);
   if(await EdgeRunning(exe))throw new InvalidOperationException("Edge не закрылся. Закройте Edge и фоновые окна Edge; VELIQ не завершает процесс принудительно.");
  }
  await model.EdgePolicyAsync(original,"guard",lifetime.Token);
   await edgeRecovery.StartAsync(lifetime.Token);
   var launch=await model.PrepareEdgeAsync(original,directory,edgeRecovery.Address,lifetime.Token);
   launch.RecoverySettings.Executable=exe;
   await edgeRecovery.ConfigureAsync(launch.RecoverySettings,lifetime.Token);
   await model.EdgePolicyAsync(original,"install",lifetime.Token);
   // Empty health profile uses ACTUAL Edge policy, with no --proxy arguments.
   await ProbeChromePolicyAsync(exe,launch.HealthUrl);
   await model.ConfirmEdgeAsync(original,launch,lifetime.Token);
   await model.CommitEdgeAsync(original,lifetime.Token);
   var start=new ProcessStartInfo(exe){UseShellExecute=false};foreach(var arg in launch.UserArguments())start.ArgumentList.Add(arg);start.ArgumentList.Add("--restore-last-session");
   using var process=Process.Start(start)??throw new InvalidOperationException("Не удалось запустить Edge");
   edgeNeedsActivation=false;
   model.SetNotice("Edge AUTO готов: политика и CONNECT подтверждены Edge. Доступность сайтов проверяется отдельно.");
   return true;
  }catch(Exception error){
   try{await model.RestoreEdgeAsync(original,lifetime.Token);}catch(Exception rollback){throw new InvalidOperationException("Не удалось восстановить Edge guard: "+rollback.Message,error);}
   edgeRecovery.Dispose();edgeNeedsActivation=true;
   if(!await EdgeRunning(exe)){var start=new ProcessStartInfo(exe){UseShellExecute=false};start.ArgumentList.Add("--user-data-dir="+directory);start.ArgumentList.Add("--restore-last-session");using var reopened=Process.Start(start);}
   throw new InvalidOperationException("Edge AUTO не активирован, обычные сайты DIRECT, protected-домены закрыты: "+error.Message,error);
  }
 }
 async Task MonitorEdgeTransportAsync(){
  if(edgeActivating||edgeMonitoring||edgeNeedsActivation||!model.Applications.Any(r=>EdgeSites(r.Profile)&&r.Profile.Mode=="AUTO"))return;
  edgeMonitoring=true;
  try{
   // Recovery remains an owner-user process; it survives service loss and can
   // release only ordinary domains authorized by the service-owned guard file.
   await edgeRecovery.StartAsync(lifetime.Token);
   if(!edgeRecovery.Configured){
    edgeRecovery.Dispose();edgeNeedsActivation=true;
    model.ApplyEdgePolicyState(new(){LastError="Recovery не настроен после сброса локального состояния. Нажмите «Оптимизировать» для повторной проверки Edge."});
    return;
   }
   if(!model.Available){model.ApplyEdge(await edgeRecovery.SnapshotAsync(lifetime.Token));return;}
   var status=await model.EdgeStatusAsync(lifetime.Token);var snapshot=status.Snapshot();
   var recovery=await edgeRecovery.SnapshotAsync(lifetime.Token);
   snapshot.Records=snapshot.Records.Concat(recovery.Records.Where(r=>r.EvidenceSource=="RECOVERY_DIRECT"&&r.Plan.Policy.Action=="DIRECT")).ToArray();
   snapshot.Lost+=recovery.Lost;snapshot.Truncated|=recovery.Truncated;
   model.ApplyEdge(snapshot);model.ApplyEdgePolicyState(status);
   if(!status.Listening)model.SetNotice("Служба Edge CONNECT недоступна: ordinary DIRECT recovery, protected-домены закрыты.");
   else if(!status.PolicyReady&&model.Applications.Any(r=>EdgeSites(r.Profile)&&r.Profile.Mode=="AUTO")){
    foreach(var p in model.Applications.Where(r=>EdgeSites(r.Profile)&&r.Profile.Mode=="AUTO").Select(r=>r.Profile.Copy()).ToArray())await model.RestoreEdgeAsync(p,lifetime.Token);
    model.SetNotice("Edge PAC изменён или недоступен: AUTO остановлен для Edge; Chrome не изменён.");
   }
  }catch(Exception error)when(!lifetime.IsCancellationRequested){model.SetNotice("Edge recovery: "+error.Message);}
  finally{edgeMonitoring=false;}
 }
}
