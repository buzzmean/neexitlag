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
}
