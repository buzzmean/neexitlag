using System.Text.Json;

namespace Neexitlag.Desktop;

public partial class App
{
    // Explicit bootstrap only. Normal GUI behavior/dialogs stay unchanged.
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
        try
        {
            await model.RefreshAsync(lifetime.Token);
            if (!model.Available) throw new IOException("Служба недоступна");
            phase = "profiles";
            await model.EnsureStartupBrowserProfilesAsync(BrowserCatalog.Discover(File.Exists, Environment.GetEnvironmentVariable), lifetime.Token);
            // Reuse the same recovery child, owner SID, actual browser policy
            // nonce and rollback transaction as the existing GUI buttons.
            // Edge commits its PAC without enabling Optimize. Chrome's legacy
            // commit enables Optimize, so prepare Edge first to avoid racing an
            // unprepared installed Edge lane during that transaction.
            phase = "edge_pac";
            foreach (var profile in model.Applications.Select(r => r.Profile.Copy()).Where(BrowserCatalog.IsEdge).ToArray())
                if (!await ActivateEdgeSites(profile)) throw new IOException("Активация Edge отменена");
            phase = "chrome_pac";
            foreach (var row in model.Applications.Where(r => IsChrome(r.Profile)).ToArray())
                await ActivateChromeSites(row);
            phase = "optimize";
            await EnableOptimizationAsync();
            await model.RefreshAsync(lifetime.Token);
            if (!model.Optimizing) throw new IOException("Оптимизация не включена");
            phase = "verify_pac";
            foreach (var profile in model.Applications.Select(r => r.Profile).Where(p => IsChrome(p) || BrowserCatalog.IsEdge(p)))
            {
                if (BrowserCatalog.IsEdge(profile))
                {
                    var status = await model.EdgeStatusAsync(lifetime.Token);
                    if (!status.Listening || !status.PolicyReady || !status.ReadyProfiles.Contains(profile.Id)) throw new IOException("Готовность PAC Edge не подтверждена");
                }
                else
                {
                    var status = await model.ChromeStatusAsync(lifetime.Token);
                    if (!status.Listening || !ChromePolicy.Applied || !status.ReadyProfiles.Contains(profile.Id)) throw new IOException("Готовность PAC Chrome не подтверждена");
                }
            }
            pass = true;
            phase = "ready";
        }
        catch (Exception error)
        {
            // Fixed safe report, never serialize IPC payloads/certs/tokens or an
            // arbitrary exception. Existing activation already installs guard.
            errorClass = error.GetType().Name;
            model.ReportActionFailure(error);
        }
        finally
        {
            var pending = report + ".pending";
            await using (var file = new FileStream(pending, FileMode.CreateNew, FileAccess.Write, FileShare.None))
                await JsonSerializer.SerializeAsync(file, new { pass, stage = phase, error_class = errorClass, scope = "owned PAC/nonce readiness; not Internet page PASS" });
            File.Move(pending, report);
            timer.Start();
        }
    }
}
