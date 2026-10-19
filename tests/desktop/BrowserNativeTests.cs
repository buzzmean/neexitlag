using System.Text.Json;
using Neexitlag.Desktop;
using Xunit;

namespace NeExitLag.Tests;
public sealed class BrowserNativeTests
{
    sealed class NativeService : IServiceClient
    {
        public readonly PickerService Inner = new();
        public bool Reject;
        public int NativeCalls;
        public Task<T> Call<T>(string method, object? payload = null, CancellationToken cancellationToken = default)
        {
            if (method == "browser.native.enable")
            {
                NativeCalls++;
                if (Reject) return Task.FromException<T>(new ServiceException(ServiceFailure.Rejected, "NATIVE_OWNED_PAC_GUARD_REQUIRED"));
                return Task.FromResult((T)(object)JsonSerializer.SerializeToElement(new { native = true }));
            }
            return Inner.Call<T>(method, payload, cancellationToken);
        }
    }
    [Theory]
    [InlineData("chrome.exe")]
    [InlineData("msedge.exe")]
    public async Task NativeEnableIsOneIPCThenRefreshNoPrepareQuitOrProfileRewrite(string exe)
    {
        var service = new NativeService(); var p = new Profile { Id = "browser", Executables = [exe], Mode = "AUTO" };
        service.Inner.State.Profiles = [p];
        var vm = new DesktopViewModel(service); await vm.RefreshAsync();
        await vm.EnableNativeBrowserAsync(p);
        Assert.Single(vm.Applications); Assert.Equal(1, service.NativeCalls);
        Assert.DoesNotContain(service.Inner.Calls, c => c.Method is "browser.prepare" or "edge.prepare" or "profile.set" or "optimize");
    }
    [Fact]
    public async Task GuardRefusalKeepsApplicationsAndDoesNotInstallPolicy()
    {
        var service = new NativeService { Reject = true }; service.Inner.State.Profiles = [new() { Id = "edge", Executables = ["msedge.exe"], Mode = "DIRECT" }];
        var vm = new DesktopViewModel(service); await vm.RefreshAsync();
        await Assert.ThrowsAsync<ServiceException>(() => vm.EnableNativeBrowserAsync(service.Inner.State.Profiles[0]));
        Assert.True(vm.Available); Assert.Single(vm.Applications);
        Assert.DoesNotContain(service.Inner.Calls, c => c.Method.Contains("policy") || c.Method.Contains("prepare") || c.Method == "profile.set");
    }
    [Fact]
    public async Task ExplicitDisableClearsNativeFlagInsteadOfReenablingNative()
    {
        var service=new NativeService();var p=new Profile{Id="chrome",Executables=["chrome.exe"],Mode="AUTO",PolicyData=new(){["browser_transport"]=JsonSerializer.SerializeToElement("native")}};
        service.Inner.State.Profiles=[p];var vm=new DesktopViewModel(service);await vm.RefreshAsync();
        await vm.DisableNativeBrowserAsync(p);
        var saved=Assert.Single(service.Inner.State.Profiles);
        Assert.Equal("DIRECT",saved.Mode);Assert.False(DesktopViewModel.NativeBrowser(saved));Assert.Equal(0,service.NativeCalls);
        Assert.False(vm.NativeBrowserCandidate(saved));
    }
    [Fact]
    public void NativeErrorsShowOnlyKnownCodeWithoutPayloadOrSecrets()
    {
        Assert.Equal("NATIVE_OWNED_PAC_GUARD_REQUIRED", DesktopViewModel.NativeBrowserErrorCode(new Exception("NATIVE_OWNED_PAC_GUARD_REQUIRED token=private")));
        Assert.Equal("NATIVE_ACTIVATION_FAILED", DesktopViewModel.NativeBrowserErrorCode(new Exception("token=private")));
    }
    [Fact]
    public void NativeCandidatesPreserveExplicitModesAndPendingEdgeCanConverge()
    {
        var vm = new DesktopViewModel(new PickerService());
        foreach (var mode in new[] { "DIRECT", "MANUAL", "DIRECT_PLUS" })
            Assert.False(vm.NativeBrowserCandidate(new() { Executables = ["chrome.exe"], Mode = mode }));
        var edge = new Profile { Executables = ["msedge.exe"], Mode = "DIRECT", PolicyData = new() { ["app_preset"] = JsonSerializer.SerializeToElement("edge-sites-pending") } };
        Assert.True(vm.NativeBrowserCandidate(edge));
        edge.PolicyData["browser_transport"] = JsonSerializer.SerializeToElement("native");
        Assert.True(DesktopViewModel.NativeBrowser(edge)); Assert.True(DesktopViewModel.NativeBrowser(edge.Copy()));
        Assert.False(vm.NativeBrowserCandidate(new() { Executables = ["discord.exe"], Mode = "AUTO" }));
    }
}
