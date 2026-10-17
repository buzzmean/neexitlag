using System.Collections.ObjectModel;
namespace Neexitlag.Desktop;
public sealed class ChromeActivationException(string message) : Exception(message) { }
public sealed class BrowserLaunch { public string HealthUrl {get;set;}=""; public string[] Arguments {get;set;}=[]; public string Address {get;set;}=""; public string Scope {get;set;}=""; }
public sealed class BrowserTransportSnapshot {public string[] ReadyProfiles {get;set;}=[];public Dictionary<string,ulong> Failures {get;set;}=[];public bool Listening {get;set;}public bool Truncated {get;set;}public ulong Lost {get;set;}BrowserConnection[] records=[];public BrowserConnection[] Records {get=>records;set=>records=value??[];}}
public sealed class BrowserOwner {public uint Pid {get;set;} public ulong CreationTime {get;set;} public string Exe {get;set;}="";public string ProfileId {get;set;}="";}
public sealed class BrowserPlan {public FlowPolicyEvidence Policy {get;set;}=new();public string Probe {get;set;}="";public int HttpStatus {get;set;}}
public sealed class BrowserConnection {
 public string OpenFailure {get;set;}="";public ulong Id {get;set;}public BrowserOwner Owner {get;set;}=new();public string Domain {get;set;}="";public BrowserPlan Plan {get;set;}=new();public string Result {get;set;}="";public ulong SentBytes {get;set;}public ulong ReceivedBytes {get;set;}public bool Active {get;set;}
 public bool DirectPlusApplied {get;set;}
 public string Route=>(Result=="UPSTREAM_OPEN_FAILED"||Result=="ORIGIN_PLAN_FAILED")?"BLOCK":Plan.Policy.Action=="DIRECT_PLUS"&&!DirectPlusApplied?"DIRECT+ (не подтверждён)" :Plan.Policy.Action=="RELAY_REQUIRED"?"RELAY_REQUIRED · "+Plan.Policy.RelayId:Plan.Policy.Action;
 public string Reason=>Plan.Policy.Reason;
 public string Outcome=>OpenFailure!=""?OpenFailure:Plan.Probe=="HTTP_SUCCESS"?$"Origin HTTP {Plan.HttpStatus}; страница Chrome не проверена":Plan.Probe;
 public string Traffic=>$"↑ {Text.Bytes(SentBytes)} / ↓ {Text.Bytes(ReceivedBytes)}";
}
public sealed partial class DesktopViewModel {
 public ObservableCollection<BrowserConnection> BrowserDomains {get;}=[];
 public void ApplyBrowser(BrowserTransportSnapshot status){
  // Only actual transport records are shown; no assumed success from TCP ACKs.
  Replace(BrowserDomains,status.Records.OrderByDescending(r=>r.Id).Take(32));
  foreach(var app in Applications)app.ApplyBrowserRecords(status);
 }
 public async Task VerifyChromeAsync(Profile profile,BrowserLaunch launch,CancellationToken ct=default){
  if(!launch.HealthUrl.StartsWith("http://neexitlag-health.invalid/ready/",StringComparison.Ordinal))throw new ChromeActivationException(Text.Get("ChromeActivationFailed"));
  for(int i=0;i<100;i++){
   try{await client.Call<System.Text.Json.JsonElement>("browser.confirm",new{profile_id=profile.Id,health_url=launch.HealthUrl},ct);return;}
   catch(ServiceException e) when(e.Failure==ServiceFailure.Rejected){await Task.Delay(200,ct);}
  }
  var diagnostics=await ChromeStatusAsync(ct);
  throw new ChromeActivationException(Text.Get("ChromeActivationFailed")+" "+string.Join(", ",diagnostics.Failures.Where(p=>p.Value>0).Select(p=>p.Key).Take(8)));
 }
 public async Task CommitChromeAsync(Profile original,CancellationToken ct=default){
  var copy=original.Copy();copy.Mode="AUTO";
  await client.Call<Profile[]>("profile.set",copy,ct);
  var status=await client.Call<Snapshot>("optimize",new{enabled=true},ct);
  if(!status.Optimizing)throw new ChromeActivationException(Text.Get("ChromeActivationFailed"));
  await RefreshAsync(ct);
 }
 public async Task RestoreChromeAsync(Profile original,CancellationToken ct=default){
  var copy=original.Copy();if(copy.Mode=="AUTO"){
   copy.Mode="DIRECT";
   if(copy.PolicyData is {} data && data.TryGetValue("rules",out var rules) && rules.ValueKind==System.Text.Json.JsonValueKind.Array && rules.GetArrayLength()>0){
    data["browser_recovery_rules"]=rules.Clone();data["rules"]=System.Text.Json.JsonSerializer.SerializeToElement(Array.Empty<object>());
   }
  }
  await client.Call<Profile[]>("profile.set",copy,ct);await RefreshAsync(ct);
 }
 public Task<BrowserTransportSnapshot> ChromeStatusAsync(CancellationToken ct=default)=>client.Call<BrowserTransportSnapshot>("browser.status",cancellationToken:ct);
 public Task<ProcessIdentity[]> ChromeProcessesAsync(string exe,CancellationToken ct=default)=>client.Call<ProcessIdentity[]>("processes",new{executables=new[]{exe}},ct);
 public Task<BrowserLaunch> PrepareChromeAsync(Profile profile,string directory,CancellationToken ct=default){
  if(!profile.Executables.Any(p=>string.Equals(p.Replace('\\','/').Split('/').Last(),"chrome.exe",StringComparison.OrdinalIgnoreCase)))throw new InvalidOperationException(Text.Get("SelectChromeAuto"));
  return client.Call<BrowserLaunch>("browser.prepare",new{profile_id=profile.Id,directory},ct);
 }
}

public sealed partial class ApplicationRow {
 internal void ApplyBrowserRecords(BrowserTransportSnapshot status){
  if(!status.Listening||Profile.Mode!="AUTO"||!Profile.Executables.Any(p=>p.Replace('\\','/').Split('/').Last().Equals("chrome.exe",StringComparison.OrdinalIgnoreCase)))return;
  var live=status.Records.Where(r=>r.Owner.ProfileId==Profile.Id&&r.Active&&(r.SentBytes>0||r.ReceivedBytes>0)).ToArray();
  route=live.Length==0?Text.Get("NoData"):string.Join(", ",live.Select(r=>r.Route).Distinct());
  UserStatus=Text.Get("ChromeRoutesByDomain")+" · "+route;
  if(status.Truncated||status.Lost>0)UserStatus=Text.Get("PartialRouteEvidence")+" "+UserStatus;
  Tcp=route;Changed("");
 }
}
