using System.Diagnostics;

namespace Neexitlag.Desktop;

// Registration is deliberately separate from PAC activation and Optimize.
// This report describes profiles, not browser/network readiness.
public sealed record BrowserStartupEntry(string Browser, string State, string ProfileId, string Mode, string ExePath, string Code);
public sealed record BrowserStartupReport(DateTimeOffset Started, DateTimeOffset Completed, long ElapsedMs, BrowserStartupEntry[] Browsers);

public sealed partial class DesktopViewModel
{
    readonly HashSet<string> startupChromeProfiles = [];
    bool browserStartupComplete;
    public BrowserStartupReport? BrowserStartup { get; private set; }
    public string BrowserStartupSummary => BrowserStartup is null
        ? "Профили браузеров: ожидается подключение к службе"
        : "Профили браузеров: " + string.Join("; ", BrowserStartup.Browsers.Select(b => $"{b.Browser}: {b.State} {b.Mode} {b.Code}"))
          + ". Регистрация не подтверждает PAC или доступность сайтов.";

    // Track Chrome created on startup for quiet PAC preparation on Optimize.
    // Existing DIRECT/custom profiles never enter this set.
    public bool StartupChromeNeedsActivation(Profile profile) => profile.Mode == "DIRECT" && startupChromeProfiles.Contains(profile.Id);

    public async Task EnsureStartupBrowserProfilesAsync(IReadOnlyList<InstalledBrowser> installed, CancellationToken ct = default)
    {
        await gate.WaitAsync(ct);
        try
        {
            if (browserStartupComplete) return;
            Busy = true;
            var started = DateTimeOffset.UtcNow;
            var watch = Stopwatch.StartNew();
            var rows = new List<BrowserStartupEntry>(2);
            foreach (var browser in new[] { "chrome", "edge" })
            {
                var found = installed.FirstOrDefault(b => b.Id == browser && b.AutoSupported);
                if (found is null)
                {
                    rows.Add(new(browser, "NOT_INSTALLED", "", "", "", ""));
                    continue;
                }
                try
                {
                    var profiles = await client.Call<Profile[]>("profiles", cancellationToken: ct);
                    var existing = profiles.FirstOrDefault(p => browser == "chrome" ? BrowserCatalog.IsChrome(p) : BrowserCatalog.IsEdge(p));
                    if (existing is not null)
                    {
                        rows.Add(new(browser, "EXISTING", existing.Id, existing.Mode, existing.Executables.FirstOrDefault() ?? "", ""));
                        continue;
                    }
                    if (!Path.IsPathFullyQualified(found.ExePath) || !found.ExePath.EndsWith(".exe", StringComparison.OrdinalIgnoreCase) || !File.Exists(found.ExePath))
                        throw new InvalidOperationException("Invalid installed browser path");
                    var path = Path.GetFullPath(found.ExePath);
                    var profile = new Profile
                    {
                        Id = "custom-" + Guid.NewGuid().ToString("N"), Name = found.Name,
                        Executables = [path], Mode = "DIRECT", Custom = true
                    };
                    if (browser == "edge")
                    {
                        await RequireEdgeServiceAsync(ct);
                        profile.PolicyData = new() { ["app_preset"] = System.Text.Json.JsonSerializer.SerializeToElement("edge-sites-pending") };
                    }
                    var saved = await client.Call<Profile[]>("profile.set", profile, ct);
                    if (Snapshot is { } snapshot) snapshot.Profiles = saved;
                    foreach (var p in saved)
                    {
                        if (Applications.Any(a => a.Profile.Id == p.Id)) continue;
                        var row = new ApplicationRow(); row.Update(p, null, null); Applications.Add(row);
                    }
                    if (browser == "chrome") startupChromeProfiles.Add(profile.Id);
                    rows.Add(new(browser, "CREATED_PENDING_ACTIVATION", profile.Id, profile.Mode, path, ""));
                }
                catch (OperationCanceledException) when (ct.IsCancellationRequested) { throw; }
                catch (Exception error)
                {
                    // Optional startup work must not clear an authenticated table or
                    // persist raw IPC exception text (which can contain secrets).
                    var code = error is EdgeActivationException edge ? edge.Code
                        : error is ServiceException ipc ? "IPC_" + ipc.Failure.ToString().ToUpperInvariant()
                        : "BROWSER_PROFILE_REGISTRATION_FAILED";
                    rows.Add(new(browser, "FAILED", "", "", found.ExePath, code));
                    RecordGuiFailure(error, "browser.startup." + browser);
                }
            }
            BrowserStartup = new(started, DateTimeOffset.UtcNow, watch.ElapsedMilliseconds, rows.ToArray());
            browserStartupComplete = rows.All(r => r.State is not ("FAILED" or "DEFERRED"));
            Changed(nameof(BrowserStartup)); Changed(nameof(BrowserStartupSummary));
        }
        finally { Busy = false; gate.Release(); }
    }
}
