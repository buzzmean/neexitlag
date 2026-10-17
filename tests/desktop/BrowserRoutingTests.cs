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
}
