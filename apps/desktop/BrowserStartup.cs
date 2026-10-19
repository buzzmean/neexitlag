using System.Text.Json;

namespace Neexitlag.Desktop;

public partial class MainWindow
{
    DateTimeOffset nextBrowserStartup;
    BrowserStartupReport? savedBrowserStartup;
    async Task RegisterStartupBrowsersAsync()
    {
        if (!model.Available || model.Busy || DateTimeOffset.UtcNow < nextBrowserStartup) return;
        nextBrowserStartup = DateTimeOffset.UtcNow.AddSeconds(10);
        await model.EnsureStartupBrowserProfilesAsync(BrowserCatalog.Discover(File.Exists, Environment.GetEnvironmentVariable), lifetime.Token);
        await PrepareNativeBrowsersAsync();
        if (ReferenceEquals(savedBrowserStartup, model.BrowserStartup) || model.BrowserStartup is null) return;
        // Bounded, owner-user report; no policy, tokens, certificates or payloads.
        var folder = Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData), "NeExitLag");
        Directory.CreateDirectory(folder);
        var report = Path.Combine(folder, "browser-startup.json");
        var pending = report + ".pending";
        try
        {
            await File.WriteAllTextAsync(pending, JsonSerializer.Serialize(model.BrowserStartup, WireJson.Options), lifetime.Token);
            File.Move(pending, report, true);
            savedBrowserStartup = model.BrowserStartup;
        }
        finally { if (File.Exists(pending)) File.Delete(pending); }
    }
}
