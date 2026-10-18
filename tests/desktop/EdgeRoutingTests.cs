using Neexitlag.Desktop;
using System.Text.Json;
using Xunit;
namespace NeExitLag.Tests;
public class EdgeRoutingTests {
 [Fact] public async Task EdgeCleanInstallCreatesOneDirectProfileAndLeavesChromeUntouched(){
  var dir=Path.Combine(Path.GetTempPath(),"veliq-edge-"+Guid.NewGuid());Directory.CreateDirectory(dir);
  var exe=Path.Combine(dir,"msedge.exe");await File.WriteAllBytesAsync(exe,[0]);
  try{
   var service=new PickerService{State=new(){Profiles=[new(){Id="chrome",Mode="AUTO",Executables=[@"C:\Chrome\chrome.exe"],LocalStrategyId="baseline"}]}};
   var vm=new DesktopViewModel(service);
   var list=new[]{new InstalledBrowser("edge","Microsoft Edge",exe,true)};
   var created=await vm.EnsureEdgeProfileAsync(list);Assert.NotNull(created);Assert.Equal("DIRECT",created!.Mode);
   Assert.Equal(Path.GetFullPath(exe),Assert.Single(created.Executables));
   Assert.Null(await vm.EnsureEdgeProfileAsync(list));
   Assert.Single(service.Calls,c=>c.Method=="profile.set");
   Assert.Equal("AUTO",service.State.Profiles.Single(p=>p.Id=="chrome").Mode);
   Assert.Equal("baseline",service.State.Profiles.Single(p=>p.Id=="chrome").LocalStrategyId);
  }finally{Directory.Delete(dir,true);}
 }
 [Theory][InlineData("AUTO")][InlineData("MANUAL")][InlineData("DIRECT")][InlineData("DIRECT_PLUS")]
 public async Task ExistingEdgeProfileIsNotReplacedOrSilentlyConverted(string mode){
  var service=new PickerService{State=new(){Profiles=[new(){Id="mine",Mode=mode,Executables=[@"D:\Edge\msedge.exe"],RelayId="VPS-01"}]}};
  var vm=new DesktopViewModel(service);
  Assert.Null(await vm.EnsureEdgeProfileAsync([new("edge","Microsoft Edge",@"C:\Edge\msedge.exe",true)]));
  Assert.DoesNotContain(service.Calls,c=>c.Method=="profile.set");Assert.Equal(mode,service.State.Profiles[0].Mode);
 }
 [Fact]public async Task EdgePreparationCannotAcceptChromeProfile(){
  var service=new PickerService();var vm=new DesktopViewModel(service);
  await Assert.ThrowsAsync<InvalidOperationException>(async()=>await vm.PrepareEdgeAsync(new(){Executables=[@"C:\Chrome\chrome.exe"]},"dir","127.0.0.1:1"));
  Assert.Empty(service.Calls);
 }
 [Fact]public void ChromeIdentificationAndPolicyLaunchArgumentsRemainUnchanged(){
  var chrome=new Profile{Executables=[@"C:\Chrome\chrome.exe"]};var edge=new Profile{Executables=[@"C:\Edge\msedge.exe"]};
  Assert.True(BrowserCatalog.IsChrome(chrome));Assert.False(BrowserCatalog.IsEdge(chrome));
  Assert.True(BrowserCatalog.IsEdge(edge));Assert.False(BrowserCatalog.IsChrome(edge));
  Assert.Equal(new[]{"--no-first-run"},new BrowserLaunch{Arguments=["--proxy-pac-url=inline","--no-first-run"]}.UserArguments());
 }
 sealed class EdgeCalls : IServiceClient {
  public readonly PickerService Base=new();public readonly List<string> Order=[];public bool RejectGuard,Ready=true;
  public Task<T> Call<T>(string method,object? payload=null,CancellationToken cancellationToken=default){
   Order.Add(method);
   if(method=="edge.policy"&&RejectGuard)return Task.FromException<T>(new ServiceException(ServiceFailure.Rejected));
   if(method=="edge.policy"||method=="edge.confirm")return Task.FromResult((T)(object)JsonSerializer.SerializeToElement(new{applied=true,ready=Ready}));
   return Base.Call<T>(method,payload,cancellationToken);
  }
 }
 [Fact]public async Task FailedGuardNeverDisarmsEdgeOrMutatesChrome(){
  var service=new EdgeCalls{RejectGuard=true};var vm=new DesktopViewModel(service);
  await Assert.ThrowsAsync<ServiceException>(()=>vm.RestoreEdgeAsync(new(){Id="edge",Mode="AUTO",Executables=[@"C:\Edge\msedge.exe"]}));
  Assert.Equal(new[]{"edge.policy"},service.Order);Assert.Empty(service.Base.Calls);
 }
 [Fact]public async Task EdgeCommitUsesExplicitMarkerAndDoesNotStartGlobalCapture(){
  var service=new EdgeCalls();var vm=new DesktopViewModel(service);
  var original=new Profile{Id="edge",Mode="DIRECT",Executables=[@"C:\Edge\msedge.exe"],PolicyData=new(){["app_preset"]=JsonSerializer.SerializeToElement("openai-web")}};
  await vm.CommitEdgeAsync(original);
  var saved=(Profile)service.Base.Calls.Single(c=>c.Method=="profile.set").Payload!;
  Assert.Equal("edge-sites",saved.PolicyData!["app_preset"].GetString());Assert.Equal("AUTO",saved.Mode);
  Assert.Equal("openai-web",original.PolicyData!["app_preset"].GetString());Assert.Equal("DIRECT",original.Mode);
  Assert.DoesNotContain(service.Order,m=>m=="optimize"||m=="browser.prepare"||m=="browser.confirm");
 }
 [Fact]public async Task FalseEdgeReadinessCannotPublishAuto(){
  var service=new EdgeCalls{Ready=false};var vm=new DesktopViewModel(service);
  await Assert.ThrowsAsync<InvalidOperationException>(()=>vm.ConfirmEdgeAsync(new(){Id="edge"},new(){HealthUrl="http://neexitlag-health.invalid/ready/nonce"}));
  Assert.Equal(new[]{"edge.confirm"},service.Order);
 }

}
