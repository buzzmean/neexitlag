using Neexitlag.Desktop;
using Xunit;
namespace NeExitLag.Tests;
public class BrowserRoutingTests {
 [Fact]public void NoFakePageSuccessOrNativeTransformFromConnect(){
  var model=new DesktopViewModel(new FakeService());
  var record=new BrowserConnection{Id=1,Domain="www.youtube.com",Plan=new(){Policy=new(){Action="DIRECT_PLUS"},Probe="NOT_PROBED"},Result="TCP_ESTABLISHED_HTTP_RESULT_UNKNOWN"};
  model.ApplyBrowser(new(){Records=[record]});
  Assert.Contains("не подтверждён",model.BrowserDomains[0].Route);
  Assert.DoesNotContain("PASS",record.Outcome);
  record.DirectPlusApplied=true;Assert.Equal("DIRECT_PLUS",record.Route);
  record.Plan.Probe="HTTP_SUCCESS";record.Plan.HttpStatus=200;Assert.Contains("страница Chrome не проверена",record.Outcome);
 }
 [Fact]public void FailedProtectedOpenDisplaysBlock(){
  var record=new BrowserConnection{Result="UPSTREAM_OPEN_FAILED",Plan=new(){Policy=new(){Action="RELAY_REQUIRED",RelayId="VPS-01"}}};Assert.Equal("BLOCK",record.Route);
 }
 [Fact]public void ChromeSummaryDoesNotInventOneBrowserWideWinner(){
  var model=new DesktopViewModel(new FakeService());
  model.Apply(new(){Profiles=[new(){Id="chrome",Name="Chrome",Mode="AUTO",Executables=[@"C:\Apps\chrome.exe"]}]});
  BrowserConnection Flow(string domain,string action,bool local=false)=>new(){Domain=domain,Owner=new(){ProfileId="chrome"},Active=true,SentBytes=100,ReceivedBytes=100,DirectPlusApplied=local,Plan=new(){Policy=new(){Action=action,RelayId=action=="RELAY_REQUIRED"?"VPS-01":""}}};
  model.ApplyBrowser(new(){Listening=true,Records=[Flow("chatgpt.com","RELAY_REQUIRED"),Flow("www.youtube.com","DIRECT_PLUS",true),Flow("example.org","DIRECT")]});
  var row=model.Applications.Single();Assert.Contains("по сайтам",row.UserStatus);Assert.Contains("RELAY_REQUIRED",row.Route);Assert.Contains("DIRECT_PLUS",row.Route);Assert.Contains("DIRECT",row.Route);
 }
 [Fact]public void InactiveBrowserTransportAcceptsNullOptionalRecords(){
  var value=System.Text.Json.JsonSerializer.Deserialize<BrowserTransportSnapshot>("{\"listening\":false,\"records\":null}",WireJson.Options)!;
  var model=new DesktopViewModel(new FakeService());model.ApplyBrowser(value);Assert.Empty(model.BrowserDomains);
 }

 [Fact]public async Task ChromePrepareDoesNotPublishAUTOOrEnableCapture(){
  var service=new BrowserService();var vm=new DesktopViewModel(service);
  var p=new Profile{Id="chrome",Mode="DIRECT_PLUS",Executables=[@"C:\Apps\chrome.exe"]};
  var launch=await vm.PrepareChromeAsync(p,@"C:\Users\test\Chrome");
  Assert.Single(service.Fake.Calls);Assert.Equal("browser.prepare",service.Fake.Calls[0].method);
  Assert.Equal("DIRECT_PLUS",p.Mode);Assert.NotEmpty(launch.HealthUrl);
 }
 [Fact]public async Task ChromeCommitDoesNotIgnoreRejectedProfileOrFailedOptimize(){
  var service=new BrowserService();var vm=new DesktopViewModel(service);
  var p=new Profile{Id="chrome",Mode="DIRECT",Executables=[@"C:\Apps\chrome.exe"]};
  service.Fake.Failure=new ServiceException(ServiceFailure.Rejected);
  await Assert.ThrowsAsync<ServiceException>(()=>vm.CommitChromeAsync(p));
  Assert.DoesNotContain(service.Fake.Calls,c=>c.method=="optimize");
  service.Fake.Failure=null;service.Fake.Calls.Clear();service.Fake.State.Optimizing=false;
  await Assert.ThrowsAsync<ChromeActivationException>(()=>vm.CommitChromeAsync(p));
  Assert.Equal(new[]{"profile.set","optimize"},service.Fake.Calls.Select(c=>c.method));
 }
 [Fact]public async Task ChromeRollbackRestoresOrdinaryDIRECTInsteadOfBrokenUnactivatedAUTO(){
  var service=new BrowserService();var vm=new DesktopViewModel(service);
  var p=new Profile{Id="chrome",Mode="AUTO",Executables=[@"C:\Apps\chrome.exe"],LocalStrategyId="working-youtube"};
  await vm.RestoreChromeAsync(p);
  var sent=(Profile)service.Fake.Calls.Single(c=>c.method=="profile.set").payload!;
  Assert.Equal("DIRECT",sent.Mode);Assert.Equal("working-youtube",sent.LocalStrategyId);Assert.Equal("AUTO",p.Mode);
  Assert.DoesNotContain(service.Fake.Calls,c=>c.method=="optimize");
 }
 sealed class BrowserService:IServiceClient{
  public FakeService Fake=new();
  public Task<T> Call<T>(string method,object? payload=null,CancellationToken cancellationToken=default){
   if(method=="browser.prepare"){Fake.Calls.Add((method,payload));return Task.FromResult((T)(object)new BrowserLaunch{HealthUrl="http://neexitlag-health.invalid/ready/nonce"});}
   return Fake.Call<T>(method,payload,cancellationToken);
  }
 }
}
