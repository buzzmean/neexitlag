namespace Neexitlag.Desktop;

public partial class MainWindow
{
    bool nativeBrowsersActivating;
    readonly Dictionary<string, DateTimeOffset> nativeBrowserRetry = [];
    readonly HashSet<string> nativeBrowserPolicyRefreshed = [];
    async Task PrepareNativeBrowsersAsync(bool forOptimize = false)
    {
        if (nativeBrowsersActivating || !model.Available || model.Busy || (!forOptimize && !model.Optimizing)) return;
        nativeBrowsersActivating = true;
        try
        {
            foreach (var p in model.Applications.Select(a => a.Profile.Copy()).Where(model.NativeBrowserCandidate).ToArray())
            {
                // Idempotent enable repairs an owned Edge guard/listener after service restart.
                if (!forOptimize && nativeBrowserRetry.TryGetValue(p.Id, out var at) && at > DateTimeOffset.UtcNow) continue;
                nativeBrowserRetry[p.Id] = DateTimeOffset.UtcNow.AddSeconds(30);
                try {
                    await model.EnableNativeBrowserAsync(p, lifetime.Token);
                    // Restore existing owner-scoped recovery only; never create PAC
                    // or a new browser profile, and never launch/quit a browser.
                    if (BrowserCatalog.IsEdge(p)) {
                        if (!edgeLifetimeHook) { Closed += (_,_) => edgeRecovery.Dispose(); edgeLifetimeHook=true; }
                        if (File.Exists(Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData), "NeExitLag", "edge-recovery.json")))
                            await edgeRecovery.StartAsync(lifetime.Token);
                    } else if (ChromePolicy.Monitored) await chromeRecovery.StartAsync(lifetime.Token);
                    if (nativeBrowserPolicyRefreshed.Add(p.Id))
                    {
                        var refresh = BrowserNativePolicy.RequestRefresh();
                        var folder = Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData), "NeExitLag");
                        Directory.CreateDirectory(folder);
                        await File.WriteAllTextAsync(Path.Combine(folder, "browser-native-policy-refresh.txt"), refresh, lifetime.Token);
                    }
                }
                catch (OperationCanceledException) when (lifetime.IsCancellationRequested) { throw; }
                catch (Exception error)
                {
                    model.RecordGuiFailure(error, "browser.native.enable");
                    model.SetNotice("browser.native.enable: native не активирован. Нужны обновлённая служба, WinDivert и существующий owned PAC crash-guard. Fresh без guard не активируется; старые кнопки PAC доступны. Профили и окна сохранены.");
                }
            }
            await model.ObserveAuxiliaryAsync(() => model.ObserveNativeBrowsersAsync(lifetime.Token), "browser.native.status", lifetime.Token);
        }
        finally { nativeBrowsersActivating = false; }
    }
}
