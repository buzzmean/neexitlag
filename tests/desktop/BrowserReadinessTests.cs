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
 [Fact] public void PreparationJournalRetainsSafeNativeCauseWithoutExceptionPayload(){
  var vm=new DesktopViewModel(new PickerService());vm.Apply(new(){Optimizing=true},[]);
  vm.ReportBrowserPreparationFailure(new ServiceException(ServiceFailure.Rejected,"NATIVE_CAPTURE_START_FAILED secret=private"),"browser.native.enable");
  Assert.Equal("NATIVE_CAPTURE_START_FAILED",vm.LastGuiFailure!.Code);
  Assert.DoesNotContain("private",vm.Notice);Assert.DoesNotContain("NATIVE_",vm.Notice);Assert.Empty(vm.Error);
 }
}
