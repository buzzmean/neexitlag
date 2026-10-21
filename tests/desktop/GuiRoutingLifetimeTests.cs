using Neexitlag.Desktop;
using Xunit;
public class GuiRoutingLifetimeTests {
 [Fact] public async Task ClosingStopsRoutingBeforeCompleting() {
  var stopped=false; await GuiRoutingLifetime.StopAsync(true,()=>{stopped=true;return Task.CompletedTask;}); Assert.True(stopped);
 }
 [Fact] public async Task FailureIsNotReportedAsClosed(){await Assert.ThrowsAsync<InvalidOperationException>(()=>GuiRoutingLifetime.StopAsync(true,()=>throw new InvalidOperationException()));}
 [Fact] public async Task AbsentServiceDoesNotBlockClose(){await GuiRoutingLifetime.StopAsync(false,()=>throw new Exception());}
}
