using System.Collections.ObjectModel;
using System.ComponentModel;
using System.Runtime.CompilerServices;
using System.Text.Json;
namespace Neexitlag.Desktop;
public abstract class Observable : INotifyPropertyChanged
{
    public event PropertyChangedEventHandler? PropertyChanged;
    protected void Changed([CallerMemberName] string? property = null) => PropertyChanged?.Invoke(this, new(property));
    protected bool Set<T>(ref T field, T value, [CallerMemberName] string? property = null)
    {
        if (EqualityComparer<T>.Default.Equals(field, value))
            return false;
        field = value;
        Changed(property);
        return true;
    }
}
public sealed record Choice(string Value, string Label);
public sealed record RouteRow(string Id, string Name, string Status, string Rtt, string Jitter, string Loss, string Score, string Role, QualityMetrics? Quality = null)
{
    public string Burst => Quality is null ? Text.Get("NoData") : Quality.MaxLossBurst.ToString(Text.Russian);
    public string Stability => Quality is null ? Text.Get("NoData") : Text.Percent(Quality.Stability);
    public string SampleCount => Quality is null ? Text.Get("NoData") : Quality.SampleCount.ToString(Text.Russian);
    public string Median => Quality?.PercentilesAvailable == true ? Text.Milliseconds(Quality.MedianRtt / 1e6) : Text.Get("NoData");
    public string P95 => Quality?.PercentilesAvailable == true ? Text.Milliseconds(Quality.P95Rtt / 1e6) : Text.Get("NoData");
}
public sealed record DiagnosticRow(string Name, string Status, string Detail);
public sealed record MeasurementRow(string Time, string Route, string Rtt, string Jitter, string Loss);
public sealed class ApplicationRow : Observable
{
    public Profile Profile { get; private set; } = new(); public string Name => Profile.Name; public string Path => executablePath; public string Mode => Text.Mode(Profile.Mode);
    string processStatus = Text.Get("NoData"), route = Text.Get("NoData"), executablePath = Text.Get("NoData"); public string ProcessStatus => processStatus; public string Route => route;
    public void Update(Profile profile, ProcessIdentity[]? processes, CaptureEvidence? capture)
    {
        Profile = profile;
        var matched = processes?.Where(p => profile.Executables.Any(exe => Matches(exe, p.Path))).ToArray();
        executablePath = string.Join(", ", profile.Executables.Select(exe => exe.Contains('\\') || exe.Contains('/') ? exe : matched?.FirstOrDefault(p => Matches(exe, p.Path))?.Path ?? $"{exe} ({Text.Get("NoData")})"));
        processStatus = matched is null ? Text.Get("NoData") : Text.Get(matched.Length > 0 ? "Active" : "NotRunning");
        var routes = capture?.Flows.Where(f => matched?.Any(p => p.PID == f.Pid && p.CreationTime == f.CreationTime) == true && profile.Executables.Any(e => Matches(e, f.Exe)) && DateTimeOffset.UtcNow - f.LastSeen < TimeSpan.FromSeconds(10)).Select(f => f.SelectedRoute == "RELAY" ? Text.Route(f.RelayId) : f.SelectedRoute == "DIRECT" ? Text.Route("DIRECT") : Text.Get("NoData")).Distinct().ToArray();
        route = routes is { Length: > 0 } ? string.Join(", ", routes) : Text.Get("NoData");
        Changed("");
    }
    public static bool Matches(string executable, string path)
    {
        var e = executable.Replace('\\', '/');
        var p = path.Replace('\\', '/');
        return string.Equals(e, p.Contains('/') && !e.Contains('/') ? p[(p.LastIndexOf('/') + 1)..] : p, StringComparison.OrdinalIgnoreCase);
    }
}
public sealed class DesktopViewModel : Observable
{
    readonly IServiceClient client; readonly SemaphoreSlim gate = new(1, 1); public Snapshot? Snapshot
    {
        get; private set;
    }
    bool available, busy; string connection = Text.Get("Connecting"), error = "", notice = "";
    public bool Available
    {
        get => available; private set
        {
            if (Set(ref available, value))
                Changed(nameof(CanControl));
        }
    }
    public bool Busy
    {
        get => busy; private set
        {
            if (Set(ref busy, value))
                Changed(nameof(CanControl));
        }
    }
    public bool CanControl => Available && !Busy;
    public string Connection
    {
        get => connection; private set => Set(ref connection, value);
    }
    public string Error
    {
        get => error; private set => Set(ref error, value);
    }
    public string Notice
    {
        get => notice; private set => Set(ref notice, value);
    }
    public ObservableCollection<ApplicationRow> Applications { get; } = []; public ObservableCollection<RouteRow> Routes { get; } = []; public ObservableCollection<Choice> Relays { get; } = [];
    public ObservableCollection<DiagnosticRow> Diagnostics { get; } = []; public ObservableCollection<MeasurementRow> Measurements { get; } = []; public ObservableCollection<string> Changes { get; } = [];
    public ObservableCollection<Choice> Modes { get; } = [new("AUTO", Text.Mode("AUTO")), new("DIRECT", Text.Mode("DIRECT")), new("MANUAL", Text.Mode("MANUAL"))];
    public bool Optimizing => Available && Snapshot?.Optimizing == true; public string OptimizeLabel => Text.Get(Optimizing ? "DisableOptimization" : "EnableOptimization"); public string OptimizationStatus => Text.Get(!Available ? "NoData" : Optimizing ? "OptimizationOn" : "OptimizationOff");
    public string Primary => Available ? Text.Route(Snapshot?.Primary) : Text.Get("NoData");
    public string ActiveApps => !Available ? Text.Get("NoData") : Snapshot?.ActiveApps.Length > 0 ? string.Join(", ", Snapshot.ActiveApps.Select(id => Snapshot.Profiles.FirstOrDefault(p => p.Id == id)?.Name ?? id)) : Text.Get("NoActiveApps");
    public string Rx => Available && Snapshot is { } s ? Text.Bytes(s.RxBytes) : Text.Get("NoData"); public string Tx => Available && Snapshot is { } s ? Text.Bytes(s.TxBytes) : Text.Get("NoData");
    Route? Current => Snapshot?.Routes.FirstOrDefault(r => r.Id == Snapshot.Primary); public string CurrentLatency => Metric(Current, r => Text.Milliseconds(r.RttMs));
    public string DirectLatency => Metric(Snapshot?.Routes.FirstOrDefault(r => r.Id == "DIRECT"), r => Text.Milliseconds(r.RttMs));
    public string RelayLatency => Metric(Snapshot?.Routes.Where(r => r.Id != "DIRECT" && Fresh(r)).OrderBy(r => r.Id == Snapshot.Primary ? 0 : 1).ThenBy(r => r.Score ?? double.MaxValue).FirstOrDefault(), r => Text.Milliseconds(r.RttMs)); public string Jitter => Metric(Current, r => Text.Milliseconds(r.JitterMs)); public string Loss => Available && Current?.Quality?.NetworkSampleCount > 0 ? Text.Percent(Current.Quality.ProbeLoss) : Metric(Current, r => Text.Percent(r.Loss));
    public string CurrentBurst => Available && Current?.Quality is { } q ? q.MaxLossBurst.ToString(Text.Russian) : Text.Get("NoData");
    public string CurrentStability => Available && Current?.Quality is { } q ? Text.Percent(q.Stability) : Text.Get("NoData");
    public string CurrentSamples => Available && Current?.Quality is { } q ? q.SampleCount.ToString(Text.Russian) : Text.Get("NoData");
    public DesktopViewModel(IServiceClient client)
    {
        this.client = client;
    }
    public static bool Fresh(Route route) => route.Status == "healthy" && route.ObservedAt != default && DateTimeOffset.UtcNow - route.ObservedAt < TimeSpan.FromSeconds(90) && route.ObservedAt <= DateTimeOffset.UtcNow.AddSeconds(5);
    string Metric(Route? route, Func<Route, string> format) => Available && route is not null && Fresh(route) ? format(route) : Text.Get("NoData");
    static void Replace<T>(ObservableCollection<T> target, IEnumerable<T> source)
    {
        var next = source.ToArray();
        if (target.SequenceEqual(next))
            return;
        target.Clear();
        foreach (var item in next)
            target.Add(item);
    }
    public async Task RefreshAsync(CancellationToken ct = default)
    {
        if (!await gate.WaitAsync(0, ct))
            return;
        Busy = true;
        try
        {
            var status = await client.Call<Snapshot>("status", cancellationToken: ct);
            if (status.Demo)
                throw new ServiceException(ServiceFailure.Incompatible);
            ProcessIdentity[]? processes = null;
            CaptureEvidence? capture = null;
            // Auxiliary diagnostics must not prevent a valid service connection.
            try
            {
                processes = await client.Call<ProcessIdentity[]>("processes", new
                {
                    executables = status.Profiles.SelectMany(p => p.Executables).Distinct(StringComparer.OrdinalIgnoreCase).ToArray()
                }, ct);
            }
            catch (ServiceException) { }
            try
            {
                capture = await client.Call<CaptureEvidence>("capture.diagnostics", cancellationToken: ct);
            }
            catch (ServiceException) { }
            Apply(status, processes, capture);
            Error = "";
        }
        catch (OperationCanceledException) when (ct.IsCancellationRequested) { }
        catch (Exception e) { SetUnavailable(e); }
        finally { Busy = false; gate.Release(); }
    }
    public void Apply(Snapshot status, ProcessIdentity[]? processes = null, CaptureEvidence? capture = null)
    {
        Snapshot = status;
        Available = true;
        Connection = Text.Get("Connected");
        if (status.Backend.Capabilities.Tcp && !Modes.Any(m => m.Value == "BLOCK"))
            Modes.Add(new("BLOCK", Text.Mode("BLOCK")));
        if (!status.Backend.Capabilities.Tcp)
        {
            var block = Modes.FirstOrDefault(m => m.Value == "BLOCK");
            if (block is not null)
                Modes.Remove(block);
        }
        if (capture?.Truncated == true)
            capture = null;
        foreach (var old in Applications.Where(a => !status.Profiles.Any(p => p.Id == a.Profile.Id)).ToArray())
            Applications.Remove(old);
        foreach (var profile in status.Profiles)
        {
            var row = Applications.FirstOrDefault(a => a.Profile.Id == profile.Id);
            if (row is null)
            {
                row = new();
                Applications.Add(row);
            }
            row.Update(profile, processes, capture);
        }
        Replace(Routes, status.Routes.Select(r => new RouteRow(r.Id, Text.Route(r.Id), Text.Get(Fresh(r) ? "Available" : "Unavailable"), Metric(r, x => Text.Milliseconds(x.RttMs)), Metric(r, x => Text.Milliseconds(x.JitterMs)), r.Quality?.NetworkSampleCount > 0 ? Text.Percent(r.Quality.ProbeLoss) : Metric(r, x => Text.Percent(x.Loss)), Fresh(r) && r.Score.HasValue ? r.Score.Value.ToString("F1", Text.Russian) : Text.Get("NoData"), r.Id == status.Primary ? Text.Get("Primary") : r.Id == status.Secondary || r.Id == status.Tertiary ? Text.Get("Backup") : "", r.Quality)));
        var availableRelays = status.Routes.Where(r => r.Id != "DIRECT" && Fresh(r)).Select(r => new Choice(r.Id, Text.Route(r.Id))).ToArray();
        if (!Relays.SequenceEqual(availableRelays))
            Replace(Relays, availableRelays);
        RenderHistory(status.History);
        Changed("");
    }
    void RenderHistory(JsonElement history)
    {
        var rows = new List<MeasurementRow>();
        var events = new List<string>();
        if (history.ValueKind != JsonValueKind.Object)
        {
            Replace(Measurements, rows);
            Replace(Changes, events);
            return;
        }
        if (history.TryGetProperty("measurements", out var list) && list.ValueKind == JsonValueKind.Array)
            foreach (var m in list.EnumerateArray().TakeLast(80).Reverse())
            {
                bool measured = m.GetProperty("available").GetBoolean();
                rows.Add(new(m.GetProperty("observed_at").GetDateTimeOffset().ToLocalTime().ToString("HH:mm:ss"), Text.Route(m.GetProperty("route_id").GetString()), measured ? Text.Milliseconds(m.GetProperty("rtt").GetDouble() / 1e6) : Text.Get("NoData"), measured ? Text.Milliseconds(m.GetProperty("jitter").GetDouble() / 1e6) : Text.Get("NoData"), measured ? Text.Percent(m.GetProperty("loss").GetDouble()) : Text.Get("NoData")));
            }
        if (history.TryGetProperty("changes", out var changes) && changes.ValueKind == JsonValueKind.Array)
            foreach (var c in changes.EnumerateArray().TakeLast(80).Reverse())
                events.Add($"{c.GetProperty("At").GetDateTimeOffset().ToLocalTime():HH:mm:ss}  {Text.Route(c.GetProperty("From").GetString())} → {Text.Route(c.GetProperty("To").GetString())}");
        Replace(Measurements, rows);
        Replace(Changes, events);
    }
    public void SetUnavailable(Exception e)
    {
        Notice = "";
        Available = false;
        Snapshot = null;
        Connection = Text.Failure(e);
        Error = Connection;
        Applications.Clear();
        Routes.Clear();
        Relays.Clear();
        Measurements.Clear();
        Changes.Clear();
        Diagnostics.Clear();
        Changed("");
    }
    async Task<bool> Act(Func<Task> action, CancellationToken ct)
    {
        await gate.WaitAsync(ct);
        Busy = true;
        try
        {
            await action();
            Error = "";
            return true;
        }
        catch (OperationCanceledException) when (ct.IsCancellationRequested) { return false; }
        catch (Exception e)
        {
            if (e is ServiceException failure && failure.Failure != ServiceFailure.Rejected)
                SetUnavailable(e);
            else
                Error = Text.Failure(e);
            return false;
        }
        finally { Busy = false; gate.Release(); }
    }
    public async Task OptimizeAsync(bool enabled, CancellationToken ct = default)
    {
        if (await Act(async () => { await client.Call<Snapshot>("optimize", new { enabled }, ct); }, ct))
            await RefreshAsync(ct);
    }
    public async Task<bool> AddApplicationAsync(string path, CancellationToken ct = default)
    {
        if (!System.IO.Path.IsPathFullyQualified(path) || !path.EndsWith(".exe", StringComparison.OrdinalIgnoreCase) || !File.Exists(path))
        {
            Error = Text.Get("InvalidExecutable");
            return false;
        }
        if (Snapshot?.Profiles.Any(p => p.Executables.Any(e => string.Equals(e, path, StringComparison.OrdinalIgnoreCase))) == true)
        {
            Error = Text.Get("AlreadyAdded");
            return false;
        }
        var profile = new Profile { Id = "custom-" + Guid.NewGuid().ToString("N"), Name = System.IO.Path.GetFileNameWithoutExtension(path), Executables = [path], Mode = "AUTO", Custom = true };
        bool ok = await Act(async () => { await client.Call<Profile[]>("profile.set", profile, ct); }, ct);
        if (ok)
            await RefreshAsync(ct);
        return ok;
    }
    public async Task<bool> SetProfileAsync(Profile original, string mode, string? relay, CancellationToken ct = default)
    {
        if (mode == "BLOCK" && Snapshot?.Backend.Capabilities.Tcp != true)
        {
            Error = Text.Get("BlockUnsupported");
            return false;
        }
        if (!new[] { "AUTO", "DIRECT", "MANUAL", "BLOCK" }.Contains(mode))
        {
            Error = Text.Get("CommandRejected");
            return false;
        }
        if (mode == "MANUAL" && !(Snapshot?.Routes.Any(r => r.Id == relay && r.Id != "DIRECT" && Fresh(r)) == true))
        {
            Error = Text.Get("NoRelay");
            return false;
        }
        var copy = original.Copy();
        copy.Mode = mode;
        copy.RelayId = mode == "MANUAL" ? relay : null;
        bool ok = await Act(async () => { await client.Call<Profile[]>("profile.set", copy, ct); }, ct);
        if (ok)
        {
            Notice = Text.Get("ProfileSaved");
            await RefreshAsync(ct);
        }
        return ok;
    }
    public async Task DeleteApplicationAsync(Profile profile, CancellationToken ct = default)
    {
        if (!profile.Custom)
        {
            Error = Text.Get("DefaultProfile");
            return;
        }
        if (await Act(async () => { await client.Call<Profile[]>("profile.delete", new { id = profile.Id }, ct); }, ct))
            await RefreshAsync(ct);
    }
    public async Task<bool> SaveServiceSettingsAsync(ServiceSettings settings, CancellationToken ct = default)
    {
        if (settings.Failover == "BLOCK" && Snapshot?.Backend.Capabilities.Tcp != true)
        {
            Error = Text.Get("BlockUnsupported");
            return false;
        }
        return await Act(async () => { await client.Call<ServiceSettings>("settings.set", settings, ct); }, ct);
    }
    public async Task DiagnoseAsync(CancellationToken ct = default)
    {
        await Act(async () =>
        {
            var checks = await client.Call<Check[]>("diagnostics", cancellationToken: ct);
            var s = await client.Call<Snapshot>("status", cancellationToken: ct);
            var rows = new List<DiagnosticRow> { new(Text.Get("ServiceName"), Text.Get("Working"), Text.Get("ServiceWorking")), new(Text.Get("DriverName"), Text.Get(s.Backend.Ready ? "Working" : string.IsNullOrEmpty(s.Backend.Error) ? "NotAvailable" : "Error"), Text.Get(s.Backend.Ready ? "DriverReady" : string.IsNullOrEmpty(s.Backend.Error) ? "DriverIdle" : "DriverFailure")) };
            foreach (var c in checks)
            {
                string key = c.Name switch
                {
                    "Control Plane" => "ControlPlane",
                    "Relay" => "RelayCheck",
                    "Latency / Loss" => "LatencyCheck",
                    _ => ""
                };
                var name = key != "" ? Text.Get(key) : new[] { "UDP", "TCP", "MTU", "DNS" }.Contains(c.Name) ? c.Name : Text.Get("Check");
                var detail = c.Status == "not configured" ? Text.Get("DiagnosticUnconfigured") : c.Status is "healthy" or "measured" ? Text.Get(c.Name == "MTU" ? "DiagnosticMtu" : c.Name == "DNS" ? "DiagnosticDns" : "DiagnosticHealthy") : Text.Get("DiagnosticUnavailable");
                rows.Add(new(name, Text.Status(c.Status), detail));
            }
            var direct = checks.FirstOrDefault(c => c.Name == "UDP");
            rows.Add(new(Text.Get("DirectCheck"), Text.Status(direct?.Status ?? "not configured"), Text.Get(direct?.Status == "healthy" ? "DiagnosticHealthy" : direct?.Status == "not configured" || direct is null ? "DiagnosticUnconfigured" : "DiagnosticUnavailable")));
            try
            {
                var drops = await client.Call<DropDiagnostics>("quality.diagnostics", cancellationToken: ct);
                foreach (var metric in drops.Capture ?? [])
                {
                    rows.Add(RenderDrop("WinDivert", metric.Key, metric.Value));
                }
                foreach (var route in drops.Transport)
                    foreach (var metric in route.Value)
                    {
                        rows.Add(RenderDrop(route.Key, metric.Key, metric.Value));
                    }
                if (!drops.RemoteRelayAvailable)
                    rows.Add(new(Text.Get("RemoteDrops"), Text.Get("NoData"), Text.Get("RemoteDropsHint")));
            }
            catch (ServiceException) {/* Older IPC v1 service has no optional endpoint. */}
            Replace(Diagnostics, rows);
        }, ct);
    }
    static DiagnosticRow RenderDrop(string scope, string key, DropMetric metric)
    {
        string resource = key switch
        {
            "capture_queue_drops" => "CaptureQueueDrops",
            "worker_saturation_drops" => "WorkerDrops",
            "transport_queue_drops" => "TransportDrops",
            "quic_decode_drops" => "DecodeDrops",
            "unknown_flow_drops" => "UnknownFlowDrops",
            "relay_socket_send_drops" => "SocketDrops",
            "reply_reinjection_failures" => "ReinjectDrops",
            "capture_kernel_queue_drops" => "KernelDrops",
            "capture_receive_errors" => "CaptureErrors",
            "policy_drops" => "PolicyDrops",
            _ => "CaptureDrops"
        };
        string detail = !metric.Available ? Text.Get("NoData") : metric.Count.ToString(Text.Russian) + " " + Text.Get(metric.Unit == "events" ? "Events" : "Packets") + (metric.Fraction.HasValue ? " (" + Text.Percent(metric.Fraction.Value) + ")" : "; " + Text.Get("FractionUnknown"));
        return new(scope + " · " + Text.Get(resource), Text.Get(!metric.Available ? "NotAvailable" : metric.Count > 0 ? "Error" : "Working"), detail);
    }

}
