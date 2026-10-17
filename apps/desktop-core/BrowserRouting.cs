using System.Collections.ObjectModel;
namespace Neexitlag.Desktop;
public sealed class BrowserLaunch { public string[] Arguments {get;set;}=[]; public string Address {get;set;}=""; public string Scope {get;set;}=""; }
public sealed class BrowserTransportSnapshot {public bool Listening {get;set;}public bool Truncated {get;set;}public ulong Lost {get;set;}public BrowserConnection[] Records {get;set;}=[];}
public sealed class BrowserOwner {public uint Pid {get;set;} public ulong CreationTime {get;set;} public string Exe {get;set;}="";public string ProfileId {get;set;}="";}
public sealed class BrowserPlan {public FlowPolicyEvidence Policy {get;set;}=new();public string Probe {get;set;}="";public int HttpStatus {get;set;}}
public sealed class BrowserConnection {
 public ulong Id {get;set;}public BrowserOwner Owner {get;set;}=new();public string Domain {get;set;}="";public BrowserPlan Plan {get;set;}=new();public string Result {get;set;}="";public ulong SentBytes {get;set;}public ulong ReceivedBytes {get;set;}public bool Active {get;set;}
 public bool DirectPlusApplied {get;set;}
 public string Route=>(Result=="UPSTREAM_OPEN_FAILED"||Result=="ORIGIN_PLAN_FAILED")?"BLOCK":Plan.Policy.Action=="DIRECT_PLUS"&&!DirectPlusApplied?"DIRECT+ (не подтверждён)" :Plan.Policy.Action=="RELAY_REQUIRED"?"RELAY_REQUIRED · "+Plan.Policy.RelayId:Plan.Policy.Action;
 public string Reason=>Plan.Policy.Reason;
 public string Outcome=>Plan.Probe=="HTTP_SUCCESS"?$"Origin HTTP {Plan.HttpStatus}; страница Chrome не проверена":Plan.Probe;
 public string Traffic=>$"↑ {Text.Bytes(SentBytes)} / ↓ {Text.Bytes(ReceivedBytes)}";
}
public sealed partial class DesktopViewModel {
 public ObservableCollection<BrowserConnection> BrowserDomains {get;}=[];
 public void ApplyBrowser(BrowserTransportSnapshot status){
  // Only actual transport records are shown; no assumed success from TCP ACKs.
  Replace(BrowserDomains,status.Records.OrderByDescending(r=>r.Id).Take(32));
 }
 public Task<ProcessIdentity[]> ChromeProcessesAsync(string exe,CancellationToken ct=default)=>client.Call<ProcessIdentity[]>("processes",new{executables=new[]{exe}},ct);
 public Task<BrowserLaunch> PrepareChromeAsync(Profile profile,string directory,CancellationToken ct=default){
  if(profile.Mode!="AUTO"||!profile.Executables.Any(p=>string.Equals(p.Replace('\\','/').Split('/').Last(),"chrome.exe",StringComparison.OrdinalIgnoreCase)))throw new InvalidOperationException(Text.Get("SelectChromeAuto"));
  return client.Call<BrowserLaunch>("browser.start",new{profile_id=profile.Id,directory},ct);
 }
}
