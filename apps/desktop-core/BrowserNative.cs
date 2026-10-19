using System.Text.Json;

namespace Neexitlag.Desktop;

public sealed class NativeBrowserStatus
{
    public int Version { get; set; }
    public bool CaptureReady { get; set; }
    public bool Optimizing { get; set; }
    public string ExistingTcp { get; set; } = "";
    public NativeBrowserTiming[] Profiles { get; set; } = [];
    public NativeBrowserRecord[] Records { get; set; } = [];
    public bool Truncated { get; set; }
    public ulong Lost { get; set; }
}
public sealed class NativeBrowserTiming
{
    public string ProfileId { get; set; } = "";
    public long EnableToFirstOpenMs { get; set; } = -1;
    public long EnableToFirstExchangeMs { get; set; } = -1;
}
public sealed class NativeBrowserRecord
{
    public ulong Id { get; set; }
    public BrowserOwner Owner { get; set; } = new();
    public DateTimeOffset Started { get; set; }
    public string Domain { get; set; } = "";
    public FlowPolicyEvidence Policy { get; set; } = new();
    public string OriginProbe { get; set; } = "";
    public string Result { get; set; } = "";
    public string FailureReason { get; set; } = "";
    public ulong SentBytes { get; set; }
    public ulong ReceivedBytes { get; set; }
    public bool Active { get; set; }
    public BrowserConnection Connection() => new() { Id = Id, Owner = Owner, Started = Started, Domain = Domain,
        Plan = new() { Policy = Policy, Probe = OriginProbe }, Result = Result, OpenFailure = FailureReason,
        SentBytes = SentBytes, ReceivedBytes = ReceivedBytes, Active = Active, EvidenceSource = "NATIVE_TCP" };
}
public sealed partial class DesktopViewModel
{
    public static bool NativeBrowser(Profile p) => p.PolicyData is { } data && data.TryGetValue("browser_transport", out var v)
        && v.ValueKind == JsonValueKind.String && v.GetString() == "native";
    public bool NativeBrowserCandidate(Profile p) => (BrowserCatalog.IsChrome(p) || BrowserCatalog.IsEdge(p))
        && (p.Mode == "AUTO" || NativeBrowser(p) || StartupChromeNeedsActivation(p)
            || p.Mode == "DIRECT" && p.PolicyData is { } data && data.TryGetValue("app_preset", out var v) && v.ValueKind == JsonValueKind.String && v.GetString() == "edge-sites-pending");
    public static string NativeBrowserErrorCode(Exception error)
    {
        string[] codes = ["NATIVE_OWNED_PAC_GUARD_REQUIRED", "NATIVE_CAPTURE_NOT_READY", "NATIVE_CAPTURE_NOT_CONFIGURED", "NATIVE_CAPTURE_START_FAILED", "NATIVE_AUTOMATIC_DISABLED", "NATIVE_OWNER_SID_REQUIRED", "NATIVE_BROWSER_PROFILE_REQUIRED", "NATIVE_PROFILE_PERSIST_FAILED", "EDGE_LISTENER_BIND_FAILED", "EDGE_POLICY_OWNERSHIP_CHANGED", "CHROME_LISTENER_BIND_FAILED", "NATIVE_GUARD_ADDRESS_CHANGED"];
        return codes.FirstOrDefault(c => error.Message.Contains(c, StringComparison.Ordinal))
            ?? (error.Message.Contains("unsupported IPC", StringComparison.OrdinalIgnoreCase) ? "NATIVE_IPC_UNSUPPORTED" : "NATIVE_ACTIVATION_FAILED");
    }
    public async Task EnableNativeBrowserAsync(Profile p, CancellationToken ct = default)
    {
        // Endpoint validates an existing owned crash-guard before any mutation.
        // Do not convert a refusal into PAC installation or a browser restart.
        await client.Call<JsonElement>("browser.native.enable", new { profile_id = p.Id }, ct);
        await RefreshAsync(ct);
    }
    public async Task ObserveNativeBrowsersAsync(CancellationToken ct = default)
    {
        if (!Applications.Any(a => NativeBrowser(a.Profile))) return;
        var status = await client.Call<NativeBrowserStatus>("browser.native.status", cancellationToken: ct);
        var records = status.Records.Select(r => r.Connection()).ToArray();
        var snapshot = new BrowserTransportSnapshot { Listening = status.CaptureReady, Records = records, Truncated = status.Truncated, Lost = status.Lost };
        if (Applications.Any(a => BrowserCatalog.IsChrome(a.Profile) && NativeBrowser(a.Profile))) ApplyBrowser(snapshot);
        if (Applications.Any(a => BrowserCatalog.IsEdge(a.Profile) && NativeBrowser(a.Profile)))
        {
            ApplyEdge(snapshot);
            EdgeState = status.CaptureReady ? "Edge native: новые TCP по SNI/Host; opaque только RELAY; owned PAC crash-guard сохранён" : "Edge native: capture не готов; owned PAC crash-guard";
            Changed(nameof(EdgeState));
        }
    }
}
