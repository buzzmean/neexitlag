using Neexitlag.Desktop;
using Xunit;
namespace NeExitLag.Tests;
public sealed class BrowserReadinessTests {
 [Fact] public void NativeReadyDoesNotDependOnPACOrPretendPageWasVerified(){
  var vm=new DesktopViewModel(new PickerService());vm.Apply(new(){Optimizing=true},[]);
  vm.ApplyNativeBrowserReadiness(new(){State="PREPARING"});Assert.Contains("Подготавливаем",vm.BrowserReadinessText);
  vm.ApplyNativeBrowserReadiness(new(){State="READY_NEW_CONNECTIONS",CaptureReady=true});Assert.Contains("новые соединения",vm.BrowserReadinessText);
  vm.Apply(new(){Optimizing=false},[]);Assert.Contains("обычную сеть",vm.BrowserReadinessText);
 }
}
