namespace Neexitlag.Desktop;
public sealed partial class DesktopViewModel {
 bool browserAutomaticStartDecided;
 int browserAutomaticStartInFlight;
 void SuppressAutomaticBrowserStartup()=>browserAutomaticStartDecided=true;
 // One automatic start per GUI lifetime. Explicit OFF wins over polling forever
 // in this window; a later GUI launch with Automatic enabled can start again.
 public async Task EnsureBrowserOptimizationOnStartupAsync(CancellationToken ct=default){
  if(browserAutomaticStartDecided||!Available||Busy||Snapshot?.Settings.Automatic!=true||Activation?.Status.CapturePermitted==false||!Applications.Any(a=>a.Profile.Mode=="AUTO"&&NativeBrowser(a.Profile)))return;
  if(Interlocked.CompareExchange(ref browserAutomaticStartInFlight,1,0)!=0)return;
  try{
   if(!Optimizing)await OptimizeAsync(true,ct);
   if(Optimizing)browserAutomaticStartDecided=true;
  }finally{Interlocked.Exchange(ref browserAutomaticStartInFlight,0);}
 }
}
