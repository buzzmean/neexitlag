using System.Text.Json;
using Neexitlag.Desktop;
using Xunit;
namespace NeExitLag.Tests;
public sealed class GameUdpTests {
 [Fact]public void ActualDecoysAndRepliesRequiredForActiveMessage(){
 var p=new GameUdpProfile{Name="PUBG",Engine=new(){Enabled=true}};
 Assert.Contains("Готов",p.Summary);
 p.Engine.Flows=[new(){Stage="LOCAL_UDP",Destination="18.157.46.1:7766",FakePackets=24,InboundPackets=0}];Assert.Contains("приманок: 24",p.Summary);Assert.Contains("ответов: 0",p.Summary);Assert.DoesNotContain("пинг",p.Summary);
 p.Reason="DRIVER_ANTICHEAT_COMPATIBILITY_UNVERIFIED";Assert.Contains("не проверена",p.Summary);
 p.Engine.Enabled=false;Assert.Contains("выключен",p.Summary);
 }
 [Fact]public void ReadsSnakeCaseAndDoesNotReportMatchSuccess(){
 var s=JsonSerializer.Deserialize<GameUdpStatus>("""{"profiles":[{"profile_id":"pubg","name":"PUBG","state":"ACTIVE","engine":{"enabled":true,"flows":[{"flow_id":1,"stage":"RELAY","destination":"18.157.46.1:7766","inbound_packets":7,"fake_packets":0}]}}]}""",WireJson.Options)!;
 var p=Assert.Single(s.Profiles);Assert.Equal("pubg",p.ProfileId);Assert.Contains("релей",p.Summary);Assert.Contains("ответов: 7",p.Summary);Assert.DoesNotContain("успеш",p.Summary);
 var vm=new DesktopViewModel(new FakeService());vm.ApplyGameUdp(s);var count=0;vm.GameUdpProfiles.CollectionChanged+=(_,_)=>count++;vm.ApplyGameUdp(s);Assert.Equal(0,count);
 }
 [Fact]public void UnexpectedServiceTextIsNotShownAsRawError(){Assert.DoesNotContain("secret",GameUdpMessages.Reason("secret=password"));}
}
public sealed class GameUdpRejectedTests {
 sealed class Rejected : IServiceClient {
  public Task<T> Call<T>(string method,object? payload=null,CancellationToken cancellationToken=default){Assert.Equal("gameudp.set",method);return Task.FromException<T>(new ServiceException(ServiceFailure.Rejected,"GAME_UDP_ANTICHEAT_COMPATIBILITY_UNVERIFIED"));}
 }
 [Fact]public async Task AnticheatRejectionHasReadableReasonAndKeepsApplications(){
 var vm=new DesktopViewModel(new Rejected());vm.Apply(new FakeService().State);var row=Assert.Single(vm.Applications);
 await vm.SetGameUdpAsync("game",true);Assert.Contains("Vanguard/FACEIT",vm.Notice);Assert.Same(row,Assert.Single(vm.Applications));Assert.Empty(vm.Error);
 }
}
