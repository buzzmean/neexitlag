using Neexitlag.Desktop;
using System.Text.Json;
using Xunit;
namespace NeExitLag.Tests;
public sealed class BrowserAutomaticStartTests {
 sealed class Service:IServiceClient{
  public readonly FakeService Inner=new();
  public Task<T> Call<T>(string method,object? payload=null,CancellationToken cancellationToken=default){
   if(method=="optimize")Inner.State.Optimizing=JsonSerializer.SerializeToElement(payload).GetProperty("enabled").GetBoolean();
   return Inner.Call<T>(method,payload,cancellationToken);
  }
 }
 [Fact]public async Task LaunchStartsOnceAndExplicitOffIsNeverUndoneByStartupPolling(){
  var s=new Service();s.Inner.State.Settings.Automatic=true;s.Inner.State.Profiles=[DesktopViewModel.NativeAutoProfile(new(){Id="chrome",Executables=[@"C:\Apps\chrome.exe"]})];
  var vm=new DesktopViewModel(s);await vm.RefreshAsync();await vm.EnsureBrowserOptimizationOnStartupAsync();
  Assert.True(vm.Optimizing);Assert.Single(s.Inner.Calls,c=>c.method=="optimize");
  await vm.OptimizeAsync(false);await vm.EnsureBrowserOptimizationOnStartupAsync();Assert.False(vm.Optimizing);
  Assert.Equal(2,s.Inner.Calls.Count(c=>c.method=="optimize"));
 }
 [Fact]public async Task AutomaticDisabledNeverStartsRouting(){
  var s=new Service();s.Inner.State.Settings.Automatic=false;s.Inner.State.Profiles=[DesktopViewModel.NativeAutoProfile(new(){Id="chrome",Executables=[@"C:\Apps\chrome.exe"]})];
  var vm=new DesktopViewModel(s);await vm.RefreshAsync();await vm.EnsureBrowserOptimizationOnStartupAsync();Assert.DoesNotContain(s.Inner.Calls,c=>c.method=="optimize");
 }
}
