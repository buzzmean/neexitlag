using System.Collections.ObjectModel;
using System.Text.Json;
namespace Neexitlag.Desktop;

public sealed class EdgeTransportStatus {
 public bool Listening {get;set;} public bool PolicyReady {get;set;} public bool Truncated {get;set;} public ulong Lost {get;set;}
 public string RecoveryAddress {get;set;}=""; public string[] ReadyProfiles {get;set;}=[];
 public BrowserConnection[] Records {get;set;}=[];
 public BrowserTransportSnapshot Snapshot()=>new(){Listening=Listening,RecoveryAddress=RecoveryAddress,ReadyProfiles=ReadyProfiles,Records=Records,Truncated=Truncated,Lost=Lost};
}
public sealed partial class DesktopViewModel {
 public ObservableCollection<BrowserConnection> EdgeDomains {get;}=[];
 public string EdgeState {get;private set;}="Edge: не настроен";
 public async Task<Profile?> EnsureEdgeProfileAsync(IReadOnlyList<InstalledBrowser> installed,CancellationToken ct=default){
  var edge=installed.FirstOrDefault(b=>b.Id=="edge"&&b.AutoSupported);
  if(edge is null)return null;
  var existing=await client.Call<Profile[]>("profiles",cancellationToken:ct);
  if(existing.Any(BrowserCatalog.IsEdge))return null;
  if(!await AddApplicationAsync(edge.ExePath,ct))return null;
  return Applications.Select(a=>a.Profile).FirstOrDefault(BrowserCatalog.IsEdge);
 }
 public Task<BrowserLaunch> PrepareEdgeAsync(Profile profile,string directory,string recovery,CancellationToken ct=default){
  if(!BrowserCatalog.IsEdge(profile))throw new InvalidOperationException("Выберите профиль Microsoft Edge");
  return client.Call<BrowserLaunch>("edge.prepare",new{profile_id=profile.Id,directory,recovery_address=recovery},ct);
 }
 public Task<JsonElement> EdgePolicyAsync(Profile profile,string operation,CancellationToken ct=default)=>client.Call<JsonElement>("edge.policy",new{profile_id=profile.Id,operation},ct);
 public Task<EdgeTransportStatus> EdgeStatusAsync(CancellationToken ct=default)=>client.Call<EdgeTransportStatus>("edge.status",cancellationToken:ct);
 public async Task ConfirmEdgeAsync(Profile original,BrowserLaunch launch,CancellationToken ct=default){
  if(!launch.HealthUrl.StartsWith("http://neexitlag-health.invalid/ready/",StringComparison.Ordinal))throw new InvalidOperationException("Некорректная проверка готовности Edge");
  var proof=await client.Call<JsonElement>("edge.confirm",new{profile_id=original.Id,health_url=launch.HealthUrl},ct);
  if(!proof.TryGetProperty("ready",out var ready)||ready.ValueKind!=JsonValueKind.True)throw new InvalidOperationException("Edge PAC не подтверждён");
 }
 public async Task CommitEdgeAsync(Profile original,CancellationToken ct=default){
  var copy=original.Copy();copy.Mode="AUTO";
  copy.PolicyData ??= new();copy.PolicyData["app_preset"]=JsonSerializer.SerializeToElement("edge-sites");
  await client.Call<Profile[]>("profile.set",copy,ct);
  await RefreshAsync(ct);
 }
 public async Task RestoreEdgeAsync(Profile original,CancellationToken ct=default){
  // Privileged guard must succeed BEFORE native interception is disarmed.
  await EdgePolicyAsync(original,"guard",ct);
  await RestoreChromeAsync(original,ct); // existing profile-preserving transaction
 }
 public void ApplyEdge(BrowserTransportSnapshot status){
  Replace(EdgeDomains,status.Records.OrderByDescending(r=>r.Started).ThenByDescending(r=>r.Id).Take(32));
  EdgeState=status.Listening?"Edge CONNECT: слушает; AUTO требует подтверждённого PAC":"Edge CONNECT: недоступен; защищённые домены закрыты";
  Changed(nameof(EdgeState));
  foreach(var row in Applications.Where(a=>BrowserCatalog.IsEdge(a.Profile)))row.ApplyEdgeRecords(status);
 }
}
public sealed partial class ApplicationRow {
 internal void ApplyEdgeRecords(BrowserTransportSnapshot status){
  if(!BrowserCatalog.IsEdge(Profile)||Profile.Mode!="AUTO")return;
  var live=status.Records.Where(r=>r.Owner.ProfileId==Profile.Id&&r.Active&&(r.SentBytes>0||r.ReceivedBytes>0)).ToArray();
  route=live.Length==0?Text.Get("NoData"):string.Join(", ",live.Select(r=>r.Route).Distinct());
  UserStatus="Edge · маршруты по доменам · "+route;
  if(status.Truncated||status.Lost>0)UserStatus=Text.Get("PartialRouteEvidence")+" "+UserStatus;
  Tcp=route;Changed("");
 }
}
