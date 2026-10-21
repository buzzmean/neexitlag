namespace Neexitlag.Desktop;
public static class GuiRoutingLifetime {
 // No browser/process control. A failed IPC leaves the GUI open with its error.
 public static Task StopAsync(bool serviceAvailable, Func<Task> disable) => serviceAvailable ? disable() : Task.CompletedTask;
}
