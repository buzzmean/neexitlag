using System.Text.Json;

namespace Neexitlag.Desktop;

public partial class App
{
    // Explicit quiet bootstrap; policy uptake is reported separately.
    internal static string? FreshPCSetupArguments(string[] args)
    {
        if (!args.Contains("--fresh-pc-setup")) return null;
        if (args.Length != 2 || args[0] != "--fresh-pc-setup" || !Path.IsPathFullyQualified(args[1]))
            throw new ArgumentException("Invalid fresh-PC setup arguments");
        var path = Path.GetFullPath(args[1]);
        var parent = Path.GetDirectoryName(path)!;
        if (!Directory.Exists(parent) || File.Exists(path) || (File.GetAttributes(parent) & FileAttributes.ReparsePoint) != 0)
            throw new ArgumentException("Fresh-PC report must be a new file in a real owner directory");
        return path;
    }
}

public partial class MainWindow
{
    internal async Task RunFreshPCSetupAsync(string report)
    {
        timer.Stop();
        var phase = "service";
        bool pass = false;
        string errorClass = "";
        string code = "BROWSER_SETUP_FAILED";
        string browser = "both";
        try
        {
            await model.RefreshAsync(lifetime.Token);
            if (!model.Available) throw new IOException("Служба недоступна");
            phase = "profiles";
            var profiles = await model.PrepareFreshBrowserProfilesAsync(BrowserCatalog.Discover(File.Exists, Environment.GetEnvironmentVariable), lifetime.Token);
            // Reuse owner-scoped recovery/listeners and quiet owned policy writes.
            // No browser launch, nonce helper, process wait or restart.
            // Edge commits its PAC without enabling Optimize. Chrome's legacy
            // commit enables Optimize, so prepare Edge first to avoid racing an
            // unprepared installed Edge lane during that transaction.
            phase = "edge_pac"; browser = "edge";
            foreach (var profile in profiles.Where(BrowserCatalog.IsEdge))
                if (!await ActivateEdgeSites(profile)) throw new ChromeActivationException("Активация Edge отменена", "edge_pac", "BROWSER_ACTIVATION_CANCELLED");
            phase = "chrome_pac"; browser = "chrome";
            foreach (var row in model.Applications.Where(r => IsChrome(r.Profile) && profiles.Any(p => p.Id == r.Profile.Id)).ToArray())
                if (!await ActivateChromeSites(row)) throw new ChromeActivationException("Активация Chrome отменена", "chrome_pac", "BROWSER_ACTIVATION_CANCELLED");
            phase = "optimize";
            // Fresh setup configures the owned PAC lanes first. Native adoption remains
            // the ordinary GUI path AFTER both guards/proofs exist.
            await model.OptimizeAsync(true, lifetime.Token);
            await model.RefreshAsync(lifetime.Token);
            if (!model.Optimizing) throw new ChromeActivationException("Оптимизация не включена", "optimize", "BROWSER_OPTIMIZE_FAILED");
            phase = "verify_pac";
            foreach (var profile in profiles)
            {
                browser = BrowserCatalog.IsEdge(profile) ? "edge" : "chrome";
                await model.VerifyFreshBrowserAsync(profile.Id, BrowserCatalog.IsEdge(profile), lifetime.Token);
                if (IsChrome(profile) && !(await model.ChromeStatusAsync(lifetime.Token)).PolicyReady)
                    throw new ChromeActivationException("Owned Chrome policy absent", "verify_pac", "CHROME_POLICY_NOT_OWNED");
            }
            pass = true;
            phase = "ready"; code = "OK"; browser = "both";
        }
        catch (Exception error)
        {
            // Fixed safe report, never serialize IPC payloads/certs/tokens or an
            // arbitrary exception. Existing activation already installs guard.
            errorClass = error.GetType().Name;
            code = FreshBrowserDiagnostics.ErrorCode(error);
            if (error is ChromeActivationException chrome) phase = chrome.Phase;
            if (error is EdgeActivationException edge) phase = edge.Phase;
            model.ReportActionFailure(error);
        }
        finally
        {
            var states = await model.FreshBrowserDiagnosticsAsync(CancellationToken.None);
            var pending = report + "." + Guid.NewGuid().ToString("N") + ".pending";
            try {
                await using (var file = new FileStream(pending, FileMode.CreateNew, FileAccess.Write, FileShare.None))
                    await JsonSerializer.SerializeAsync(file, new { pass, stage = phase, code, browser, states, error_class = errorClass, scope = "owned PAC configuration; browser uptake may be pending; not Internet page PASS" });
                File.Move(pending, report);
            }
            finally { if (File.Exists(pending)) File.Delete(pending); timer.Start(); }
        }
    }
}
