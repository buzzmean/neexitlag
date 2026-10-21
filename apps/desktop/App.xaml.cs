using System.Globalization;
using System.IO.Pipes;
using System.Security.Principal;
using System.Windows;
namespace Neexitlag.Desktop;
public partial class App : System.Windows.Application
{
    Mutex? instance; bool ownsMutex; readonly CancellationTokenSource lifetime = new(); string activationPipe = "";
    protected override async void OnStartup(StartupEventArgs e)
    {
        CultureInfo.DefaultThreadCurrentCulture = Text.Russian;
        CultureInfo.DefaultThreadCurrentUICulture = Text.Russian;
        base.OnStartup(e);
        var sid = (WindowsIdentity.GetCurrent().User?.Value ?? Environment.UserName) + "-" + System.Diagnostics.Process.GetCurrentProcess().SessionId;
        activationPipe = "NeExitLag-activate-" + sid;
        instance = new Mutex(true, "Local\\NeExitLag-" + sid, out ownsMutex);
        if (!ownsMutex)
        {
            try
            {
                using var pipe = new NamedPipeClientStream(".", activationPipe, PipeDirection.Out, PipeOptions.Asynchronous);
                using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(3));
                await pipe.ConnectAsync(deadline.Token);
                await pipe.WriteAsync(new byte[] { 1 }, deadline.Token);
            }
            catch { }
            Shutdown();
            return;
        }
        var freshReport = FreshPCSetupArguments(e.Args);
        var window = freshReport is null ? new MainWindow() : new MainWindow(LocalPreferences.Load(), false, true);
        MainWindow = window;
        window.Show();
        _ = ListenForActivation(lifetime.Token);
        if (freshReport is not null) await window.RunFreshPCSetupAsync(freshReport);
    }
    async Task ListenForActivation(CancellationToken ct)
    {
        while (!ct.IsCancellationRequested)
        {
            try
            {
                using var pipe = new NamedPipeServerStream(activationPipe, PipeDirection.In, 1, PipeTransmissionMode.Byte, PipeOptions.Asynchronous | PipeOptions.CurrentUserOnly);
                await pipe.WaitForConnectionAsync(ct);
                var signal = new byte[1];
                using var readDeadline = CancellationTokenSource.CreateLinkedTokenSource(ct);
                readDeadline.CancelAfter(TimeSpan.FromSeconds(2));
                if (await pipe.ReadAsync(signal, readDeadline.Token) == 1) {
                    if(signal[0]==1) await Dispatcher.InvokeAsync(() => ((MainWindow)MainWindow).ShowFromTray());
                    else if(signal[0]==2) await Dispatcher.InvokeAsync(() => ((MainWindow)MainWindow).Exit());
                }
            }
            catch (OperationCanceledException) { }
            catch (IOException) { await Task.Delay(200, ct).ConfigureAwait(false); }
        }
    }
    protected override void OnExit(ExitEventArgs e)
    {
        lifetime.Cancel();
        if (ownsMutex)
            instance?.ReleaseMutex();
        instance?.Dispose();
        lifetime.Dispose();
        base.OnExit(e);
    }
}
