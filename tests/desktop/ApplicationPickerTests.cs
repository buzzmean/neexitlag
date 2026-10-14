using System.Text.Json;
using Neexitlag.Desktop;
using Xunit;
namespace NeExitLag.Tests;

public sealed class PickerService : IServiceClient
{
    public Snapshot State = new() { Profiles = [] };
    public RunningApplicationList Running = new();
    public readonly List<(string Method, object? Payload)> Calls = [];
    public bool RejectStatus;
    public bool RejectApplicationQuery;
    public Task<T> Call<T>(string method, object? payload = null, CancellationToken cancellationToken = default)
    {
        Calls.Add((method, payload));
        if (method == "status" && RejectStatus) return Task.FromException<T>(new ServiceException(ServiceFailure.Unavailable));
        if (method == "processes" && RejectApplicationQuery && payload is not null && JsonSerializer.SerializeToElement(payload).TryGetProperty("applications", out _))
            return Task.FromException<T>(new ServiceException(ServiceFailure.Rejected));
        if (method == "profile.set") State.Profiles = [.. State.Profiles, ((Profile)payload!).Copy()];
        object result = method switch
        {
            "profiles" or "profile.set" => State.Profiles,
            "status" => State,
            "processes" when payload is not null && JsonSerializer.SerializeToElement(payload).TryGetProperty("applications", out _) => Running,
            "processes" => Array.Empty<ProcessIdentity>(),
            "capture.diagnostics" => new CaptureEvidence(),
            _ => throw new InvalidOperationException(method)
        };
        return Task.FromResult((T)result);
    }
}
public sealed class ApplicationPickerTests
{
    [Fact]
    public async Task PickerUsesBoundedServiceViewAndSearchesApplicationExePathAndPid()
    {
        var service = new PickerService { Running = new() { Applications = [
            new() { Name = "Discord.exe", Path = @"C:\Apps\Discord.exe", Pids = [55], Pid = 55, ProcessCount = 1, DisplayName = "Discord" },
            new() { Name = "chrome.exe", Path = @"C:\Browser\chrome.exe", Pids = [20,21,22,23,24,25,26,27,28,29], Pid = 20, ProcessCount = 10, DisplayName = "Google Chrome" }
        ] } };
        var picker = new DesktopViewModel(service).CreateApplicationPicker();
        await picker.RefreshAsync();
        var query = JsonSerializer.SerializeToElement(service.Calls.Single().Payload);
        Assert.True(query.GetProperty("applications").GetBoolean());
        Assert.Equal(128, query.GetProperty("limit").GetInt32());
        Assert.Equal(new[] { "Discord", "Google Chrome" }, picker.Applications.Select(a => a.Title));
        Assert.Equal(10, picker.Applications[1].ProcessCount);
        foreach (string term in new[] { "GOOGLE", "chrome.exe", @"C:\Browser", "29" })
        {
            picker.Search = term;
            Assert.Equal("chrome.exe", Assert.Single(picker.Applications).Name);
        }
        picker.Search = "not running";
        Assert.Empty(picker.Applications);
        Assert.Equal(Text.Get("RunningEmpty"), picker.Status);
        picker.Search = "";
        service.Running = new() { Applications = [new() { Name = "new.exe", Path = @"C:\new.exe", Pid = 99, Pids = [99], ProcessCount = 1 }], Truncated = true };
        await picker.RefreshAsync();
        Assert.Equal("new.exe", Assert.Single(picker.Applications).Name);
        Assert.True(picker.Truncated);
        Assert.Equal(Text.Get("RunningTruncated"), picker.Status);
    }
    [Fact]
    public async Task FullExecutablePathIsSavedAndAppearsImmediately()
    {
        string path = Path.Combine(Path.GetTempPath(), Guid.NewGuid()+".exe");
        try
        {
            await File.WriteAllBytesAsync(path, [0]);
            var service = new PickerService();
            var desktop = new DesktopViewModel(service);
            await desktop.RefreshAsync();
            // A stalled/unavailable next status poll cannot delay or erase the
            // authoritative profile.set result and the newly inserted row.
            service.RejectStatus = true;
            Assert.True(await desktop.AddApplicationAsync(path));
            var saved = Assert.Single(service.State.Profiles);
            Assert.Equal(Path.GetFullPath(path), Assert.Single(saved.Executables));
            Assert.True(saved.Custom);
            Assert.Equal(saved.Id, Assert.Single(desktop.Applications).Profile.Id);
            Assert.Equal(path, Assert.Single(desktop.Applications).Path);
            Assert.Single(service.Calls, c => c.Method == "status");
        }
        finally { File.Delete(path); }
    }
    [Fact]
    public async Task DuplicateRechecksServiceRatherThanStaleGuiSnapshot()
    {
        string path = Path.Combine(Path.GetTempPath(), Guid.NewGuid()+".exe");
        try
        {
            await File.WriteAllBytesAsync(path, [0]);
            var service = new PickerService();
            var desktop = new DesktopViewModel(service);
            await desktop.RefreshAsync();
            service.State = new() { Profiles = [new() { Id = "existing", Executables = [path], Custom = true }] };
            Assert.False(await desktop.AddApplicationAsync(path));
            Assert.Equal(Text.Get("AlreadyAdded"), desktop.Error);
            Assert.DoesNotContain(service.Calls, c => c.Method == "profile.set");
        }
        finally { File.Delete(path); }
    }
    [Fact]
    public async Task ManualFilePickerCancellationAndSelectionKeepWorking()
    {
        string path = Path.Combine(Path.GetTempPath(), Guid.NewGuid()+".exe");
        try
        {
            await File.WriteAllBytesAsync(path, [0]);
            var service = new PickerService();
            var desktop = new DesktopViewModel(service);
            Assert.Null(await desktop.AddApplicationFromFileAsync(() => null));
            Assert.Empty(service.Calls);
            int opened = 0;
            Assert.True(await desktop.AddApplicationFromFileAsync(() => { opened++; return path; }));
            Assert.Equal(1, opened);
            Assert.Equal(path, Assert.Single(service.State.Profiles).Executables.Single());
            Assert.Single(desktop.Applications);
            Assert.False(await desktop.AddApplicationFromFileAsync(() => path));
            Assert.Single(service.Calls, c => c.Method == "profile.set");
        }
        finally { File.Delete(path); }
    }
    [Fact]
    public async Task UnsupportedApplicationQueryKeepsManualFileFallbackAvailable()
    {
        string path = Path.Combine(Path.GetTempPath(), Guid.NewGuid()+".exe");
        try
        {
            await File.WriteAllBytesAsync(path, [0]);
            var service = new PickerService { RejectApplicationQuery = true };
            var desktop = new DesktopViewModel(service);
            var picker = desktop.CreateApplicationPicker();
            await picker.RefreshAsync();
            Assert.Empty(picker.Applications);
            Assert.NotEmpty(picker.Error);
            Assert.False(picker.Busy);
            Assert.True(await desktop.AddApplicationFromFileAsync(() => path));
            Assert.Equal(path, Assert.Single(service.State.Profiles).Executables.Single());
        }
        finally { File.Delete(path); }
    }
    [Fact]
    public async Task DirectPlusCanStillBeSavedWithoutLosingItsStrategy()
    {
        var service = new FakeService();
        service.State.Profiles[0].LocalStrategyId = "tcp-split-1";
        var desktop = new DesktopViewModel(service);
        await desktop.RefreshAsync();
        Assert.Contains(desktop.Modes, m => m.Value == "DIRECT_PLUS");
        Assert.True(await desktop.SetProfileAsync(service.State.Profiles[0], "DIRECT_PLUS", null));
        var saved = (Profile)service.Calls.Single(c => c.method == "profile.set").payload!;
        Assert.Equal("DIRECT_PLUS", saved.Mode);
        Assert.Equal("tcp-split-1", saved.LocalStrategyId);
        Assert.Equal(Text.Get("DirectPlus"), Text.Mode(saved.Mode));
    }
}
