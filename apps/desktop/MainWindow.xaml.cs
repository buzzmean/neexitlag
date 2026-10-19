using System.ComponentModel;
using System.Windows;
using System.Windows.Controls;
using System.Windows.Threading;
using Microsoft.Win32;
namespace Neexitlag.Desktop;
public partial class MainWindow : Window
{
    readonly DesktopViewModel model = new(new WindowsServiceClient()); readonly DispatcherTimer timer = new() { Interval = TimeSpan.FromSeconds(2) }; readonly CancellationTokenSource lifetime = new();
    readonly LocalPreferences preferences; readonly bool shutdownOnClose; readonly string? preferencesPath; readonly System.Windows.Forms.NotifyIcon tray; readonly System.Drawing.Icon trayImage; bool exiting, settingsLoaded, startupApplied;
    public MainWindow() : this(LocalPreferences.Load(), true, true) { }
    // The native test host uses the same BAML, event handlers and NotifyIcon,
    // without starting IPC timers or shutting down the host Application.
    internal MainWindow(LocalPreferences preferences, bool background, bool shutdownOnClose, string? preferencesPath = null)
    {
        this.preferences = preferences;
        this.shutdownOnClose = shutdownOnClose;
        this.preferencesPath = preferencesPath;
        InitializeComponent();
        InitializeGuiFailureLog();
 LocalPorts.Text="443";LocalSplitPosition.Text="1";
        DataContext = model;
        Launch.IsChecked = LocalPreferences.AutoStart;
        StartMinimized.IsChecked = preferences.StartMinimized;
        MinimizeToTray.IsChecked = preferences.MinimizeToTray;
        Channel.ItemsSource = new[] { new Choice("stable", Text.Get("Stable")), new Choice("preview", Text.Get("Preview")) };
        Channel.SelectedValue = "stable";
        Fallback.ItemsSource = new[] { new Choice("DIRECT", Text.Get("Direct")) };
        Fallback.SelectedValue = "DIRECT";
        using var iconStream = System.Windows.Application.GetResourceStream(new Uri("pack://application:,,,/VELIQ;component/Assets/NeExitLag.ico"))!.Stream;
        using var loadedIcon = new System.Drawing.Icon(iconStream);
        trayImage = (System.Drawing.Icon)loadedIcon.Clone();
        tray = new()
        {
            Icon = trayImage,
            Text = "VELIQ",
            Visible = preferences.MinimizeToTray
        };
        var menu = new System.Windows.Forms.ContextMenuStrip();
        menu.Items.Add(Text.Get("OpenApp"), null, (_, _) => Dispatcher.Invoke(ShowFromTray));
        var enable = menu.Items.Add(Text.Get("TrayEnable"), null, async (_, _) => await Execute(EnableOptimizationAsync));
        var disable = menu.Items.Add(Text.Get("TrayDisable"), null, async (_, _) => await Execute(() => model.OptimizeAsync(false, lifetime.Token)));
        menu.Items.Add(new System.Windows.Forms.ToolStripSeparator());
        menu.Items.Add(Text.Get("Exit"), null, (_, _) => Dispatcher.Invoke(Exit));
        menu.Opening += (_, _) => { enable.Enabled = model.Available && !model.Busy && !model.Optimizing; disable.Enabled = model.Available && !model.Busy && model.Optimizing; };
        tray.ContextMenuStrip = menu;
        tray.DoubleClick += (_, _) => Dispatcher.Invoke(ShowFromTray);
        Loaded += async (_, _) => { ApplyStartupState(); if (background) await Refresh(); };
        timer.Tick += async (_, _) => await Refresh();
        if (background) timer.Start();
        Closing += OnClosing;
        StateChanged += (_, _) => { if (WindowState == WindowState.Minimized && preferences.MinimizeToTray) Hide(); };
        Closed += (_, _) => { timer.Stop(); lifetime.Cancel(); chromeRecovery.Dispose(); tray.Visible = false; tray.ContextMenuStrip?.Dispose(); tray.Dispose(); trayImage.Dispose(); if (shutdownOnClose) System.Windows.Application.Current.Shutdown(); };
    }
    internal bool TrayVisible => tray.Visible;
    internal System.Windows.Forms.NotifyIcon TrayIcon => tray;
    void ApplyStartupState()
    {
        if (startupApplied) return;
        startupApplied = true;
        if (preferences.StartMinimized)
            WindowState = WindowState.Minimized; // StateChanged hides only when tray mode is enabled.
    }
    internal void SaveLocalPreferences(bool startMinimized, bool minimizeToTray)
    {
        // Persist first: a failed save must not silently change runtime behavior.
        new LocalPreferences { StartMinimized = startMinimized, MinimizeToTray = minimizeToTray }.Save(preferencesPath);
        preferences.StartMinimized = startMinimized;
        preferences.MinimizeToTray = minimizeToTray;
        tray.Visible = minimizeToTray;
        if (!minimizeToTray && !IsVisible)
        {
            // Disabling tray mode restores a taskbar window immediately.
            Show();
            WindowState = WindowState.Minimized;
        }
        else if (minimizeToTray && WindowState == WindowState.Minimized)
            Hide();
    }
    public void ShowFromTray()
    {
        Show();
        if (WindowState == WindowState.Minimized)
            WindowState = WindowState.Normal;
        Activate();
    }
    internal void Exit()
    {
        exiting = true;
        Close();
    }
    void OnClosing(object? sender, CancelEventArgs e)
    {
        if (!exiting && preferences.MinimizeToTray)
        {
            e.Cancel = true;
            Hide();
        }
    }
    async Task Execute(Func<Task> action)
    {
        try
        {
            await action();
        }
        catch (OperationCanceledException) when (lifetime.IsCancellationRequested) { }
        catch (Exception e) { model.ReportActionFailure(e); }
    }
    bool presetsLoaded,refreshingWindow;
    async Task Refresh()
    {
        if(refreshingWindow)return;refreshingWindow=true;
        try{
        var wasOptimizing=model.Optimizing;
        await Execute(() => model.RefreshAsync(lifetime.Token));
        await Execute(()=>model.ObserveAuxiliaryAsync(RegisterStartupBrowsersAsync,"browser.startup",lifetime.Token));
        await Execute(()=>model.ObserveAuxiliaryAsync(()=>MonitorChromeTransportAsync(wasOptimizing),"browser.monitor",lifetime.Token));
        if(!presetsLoaded&&model.Available){try{AppPreset.ItemsSource=await model.GetPresetsAsync(lifetime.Token);AppPreset.SelectedIndex=0;presetsLoaded=true;}catch(ServiceException){}}
        if (!settingsLoaded && model.Snapshot is { } snapshot)
        {
            LoadSettings(snapshot);
            settingsLoaded = true;
        }
        }finally{refreshingWindow=false;}
    }
    void LoadSettings(Snapshot snapshot)
    {
        Automatic.IsChecked = snapshot.Settings.Automatic;
        Channel.SelectedValue = snapshot.Settings.UpdateChannel;
        Fallback.ItemsSource = snapshot.Backend.Capabilities.Tcp ? new[] { new Choice("DIRECT", Text.Get("Direct")), new Choice("BLOCK", Text.Get("Block")) } : new[] { new Choice("DIRECT", Text.Get("Direct")) };
        Fallback.SelectedValue = snapshot.Settings.Failover == "BLOCK" && snapshot.Backend.Capabilities.Tcp ? "BLOCK" : "DIRECT";
    }
    void Navigate(object sender, RoutedEventArgs e)
    {
        string page = (string)((Button)sender).Tag;
        PageTitle.Text = Text.Get(page);
        foreach (string name in new[] { "Home", "Applications", "Routes", "Statistics", "Diagnostics", "Settings", "LocalStrategies" })
            ((FrameworkElement)FindName(name)).Visibility = name == page ? Visibility.Visible : Visibility.Collapsed;
        if (page == "Settings" && model.Snapshot is { } snapshot)
            LoadSettings(snapshot);
    }
    bool localLoaded;
    LocalStrategyConfig localConfig = new();
    readonly System.Collections.ObjectModel.ObservableCollection<LocalStrategyProfile> localRules = [];
    static string[] Parts(string s)=>s.Split(',',StringSplitOptions.RemoveEmptyEntries|StringSplitOptions.TrimEntries);
    async void LoadLocal(object sender,RoutedEventArgs e)=>await Execute(async()=>{
        localConfig=await model.GetLocalStrategiesAsync(lifetime.Token);localLoaded=true;LocalEnabled.IsChecked=localConfig.Enabled;
        localRules.Clear();foreach(var p in localConfig.User)localRules.Add(p);LocalGrid.ItemsSource=localRules;
        var m=await model.GetLocalMetricsAsync(lifetime.Token);LocalCounters.Text=string.Format(Text.Russian,Text.Get("LocalCounters"),m.PacketsInspected,m.PacketsModified,m.ExtraPacketsGenerated,m.FlowsDirectPlus,Text.Bytes(m.BytesDirectPlus));
    });
    async void SaveLocal(object sender,RoutedEventArgs e)=>await Execute(async()=>{
        if(!localLoaded){ShowMessage("LocalLoadFirst");return;}
        localConfig.Enabled=LocalEnabled.IsChecked==true;localConfig.User=localRules.ToArray();
        localConfig=await model.SaveLocalStrategiesAsync(localConfig,lifetime.Token);
        if(LocalAuto.IsChecked==true && AppGrid.SelectedItem is ApplicationRow row){if(IsChrome(row.Profile))await ActivateChromeSites(row);else await model.SetProfileAsync(row.Profile,"AUTO",null,lifetime.Token);}
        LocalCounters.Text=Text.Get("LocalSaved");
    });
    void SelectLocal(object sender,SelectionChangedEventArgs e){if(LocalGrid.SelectedItem is not LocalStrategyProfile p)return;LocalApp.Text=p.Application;LocalSplitPosition.Text=(p.SplitPosition==0?1:p.SplitPosition).ToString(System.Globalization.CultureInfo.InvariantCulture);LocalPorts.Text=string.Join(",",p.Ports ?? []);LocalDomains.Text=string.Join(",",p.Domains ?? []);LocalIPs.Text=string.Join(",",p.Ips ?? []);LocalExcludeDomains.Text=string.Join(",",p.ExcludeDomains ?? []);LocalExcludeIPs.Text=string.Join(",",p.ExcludeIps ?? []);foreach(ComboBoxItem item in LocalMechanism.Items)if((string)item.Tag==p.Strategy)LocalMechanism.SelectedItem=item;}
    void AddLocal(object sender,RoutedEventArgs e){
        if(!localLoaded){ShowMessage("LocalLoadFirst");return;}
        try{var old=LocalGrid.SelectedItem as LocalStrategyProfile;if(old?.Actions?.Length>0){ShowMessage("ChainAdvancedReadOnly");return;}var strategy=(string)((ComboBoxItem)LocalMechanism.SelectedItem).Tag;
        var p=new LocalStrategyProfile{Id=old?.Id??Guid.NewGuid().ToString("N"),Name=System.IO.Path.GetFileNameWithoutExtension(LocalApp.Text),Application=LocalApp.Text,Protocol=strategy.StartsWith("udp-")?"udp":"tcp",Ports=Parts(LocalPorts.Text).Select(ushort.Parse).ToArray(),Domains=Parts(LocalDomains.Text),Ips=Parts(LocalIPs.Text),ExcludeDomains=Parts(LocalExcludeDomains.Text),ExcludeIps=Parts(LocalExcludeIPs.Text),Strategy=strategy,SplitPosition=strategy=="tcp-split"?ushort.Parse(LocalSplitPosition.Text):((ushort)0),Cutoff=old?.Cutoff??2,Priority=old?.Priority??0,Enabled=true};
        if(old is not null)localRules.Remove(old);localRules.Add(p);LocalGrid.ItemsSource=localRules;LocalGrid.SelectedItem=p;
        }catch{ShowMessage("LocalSelect");}
    }
    void RemoveLocal(object sender,RoutedEventArgs e){if(LocalGrid.SelectedItem is LocalStrategyProfile p)localRules.Remove(p);}
    async void ApplyVariant(object sender,RoutedEventArgs e)
    {
        if(AppGrid.SelectedItem is not ApplicationRow app || LocalGrid.SelectedItem is not LocalStrategyProfile strategy || !app.Profile.Executables.Any(path=>ApplicationRow.Matches(strategy.Application,path))){ShowMessage("SelectApplication");return;}
        var copy=app.Profile.Copy();copy.LocalStrategyId=strategy.Id;
        await Execute(()=>model.SetProfileAsync(copy,"DIRECT_PLUS",null,lifetime.Token));
    }
    void TestLocalV2(object sender,RoutedEventArgs e)
    {
        if(AppGrid.SelectedItem is not ApplicationRow row || AppPreset.SelectedValue is not string preset){ShowMessage("SelectApplication");return;}
        string script=Path.Combine(AppContext.BaseDirectory,"tools","windows","strategy-tester-v2.ps1");
        if(!File.Exists(script)){var picker=new OpenFileDialog{Title=Text.Get("LocalTestV2"),Filter="PowerShell|strategy-tester-v2.ps1",CheckFileExists=true};if(picker.ShowDialog(this)!=true)return;script=picker.FileName;}
        var start=new System.Diagnostics.ProcessStartInfo("powershell.exe"){UseShellExecute=true,Verb="runas"};
        string url=string.IsNullOrWhiteSpace(LocalURL.Text)?(preset=="youtube"?"https://www.youtube.com/":"https://discord.com/app"):LocalURL.Text;
        foreach(var value in new[]{"-NoProfile","-NoExit","-File",script,"-AppPath",row.Profile.Executables[0],"-Preset",preset,"-URL",url})start.ArgumentList.Add(value);
        try{System.Diagnostics.Process.Start(start);}catch{ShowMessage("LocalSelect");}
    }
    void TestLocal(object sender,RoutedEventArgs e){
        var picker=new OpenFileDialog{Title=Text.Get("LocalTestScript"),Filter="PowerShell|strategy-tester.ps1",CheckFileExists=true};if(picker.ShowDialog(this)!=true)return;
        var p=new System.Diagnostics.ProcessStartInfo("powershell.exe"){UseShellExecute=true,Verb="runas"};
        foreach(var a in new[]{"-NoProfile","-NoExit","-File",picker.FileName,((string)((ComboBoxItem)LocalMechanism.SelectedItem).Tag).StartsWith("udp-")?"-UDPTarget":"-URL",LocalURL.Text,"-TokenFile",LocalToken.Text})p.ArgumentList.Add(a);
        if(!string.IsNullOrWhiteSpace(LocalCA.Text)){p.ArgumentList.Add("-EndpointCA");p.ArgumentList.Add(LocalCA.Text);}try{System.Diagnostics.Process.Start(p);}catch{ShowMessage("LocalSelect");}
    }
    async void Reconnect(object sender, RoutedEventArgs e) => await Refresh();
    async void ToggleOptimize(object sender, RoutedEventArgs e) => await Execute(() => model.Optimizing?model.OptimizeAsync(false,lifetime.Token):EnableOptimizationAsync());
    void AddApplication(object sender, RoutedEventArgs e) => new AddApplicationWindow(this, model, lifetime.Token).ShowDialog();
    async void ApplyPreset(object sender,RoutedEventArgs e)
    {
        if(AppGrid.SelectedItem is not ApplicationRow row || AppPreset.SelectedValue is not string preset){ShowMessage("SelectApplication");return;}
        await Execute(()=>model.ApplyPresetAsync(row.Profile,preset,AppQuic.IsChecked==true,lifetime.Token));
    }
    async void QuickPreset(object sender,RoutedEventArgs e)
    {
        if(AppGrid.SelectedItem is not ApplicationRow row){ShowMessage("SelectApplication");return;}
        await Execute(()=>model.ApplyPresetAsync(row.Profile,(string)((Button)sender).Tag,AppQuic.IsChecked==true,lifetime.Token));
    }
    void ApplicationSelected(object sender, SelectionChangedEventArgs e)
    {
        if (AppGrid.SelectedItem is ApplicationRow row)
        {
            Mode.SelectedValue = row.Profile.Mode;
            ProfileFallback.SelectedValue=row.Profile.Fallback ?? "";
            RelayChoice.SelectedValue = row.Profile.RelayId;
        }
    }
    void ModeChanged(object sender, SelectionChangedEventArgs e)
    {
        if (RelayChoice is not null)
            RelayChoice.IsEnabled = (string?)Mode.SelectedValue == "MANUAL";
    }
    bool autoRunning;
    async void UseAuto(object sender, RoutedEventArgs e)
    {
        if (AppGrid.SelectedItem is not ApplicationRow row) { ShowMessage("SelectApplication"); return; }
        if (row.Profile.Executables.Any(p=>p.Replace('\\','/').Split('/').Last().Equals("chrome.exe",StringComparison.OrdinalIgnoreCase))) { await Execute(()=>ActivateChromeSites(row)); return; }
        if(autoRunning)return;autoRunning=true;model.SetNotice(Text.Get("AutoChecking"));
        try { await Execute(async () => {
            await model.SetProfileAsync(row.Profile, "AUTO", row.Profile.RelayId, lifetime.Token);
            await model.OptimizeAsync(true, lifetime.Token);
            var script=Path.Combine(AppContext.BaseDirectory,"tools","windows","auto-optimize.ps1");
            if(!File.Exists(script)) { script=Path.GetFullPath(Path.Combine(AppContext.BaseDirectory,"../../../../../tools/windows/auto-optimize.ps1")); }
            if(!File.Exists(script))throw new InvalidOperationException(Text.Get("AutoToolsMissing"));
            var output=Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData),"NeExitLag","auto-output");Directory.CreateDirectory(output);
            var start=new System.Diagnostics.ProcessStartInfo(Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.System),"WindowsPowerShell","v1.0","powershell.exe")){UseShellExecute=false,CreateNoWindow=true,RedirectStandardOutput=true,RedirectStandardError=true};
            foreach(var value in new[]{"-NoProfile","-ExecutionPolicy","Bypass","-File",script,"-ProfileId",row.Profile.Id})start.ArgumentList.Add(value);
            using var process=System.Diagnostics.Process.Start(start) ?? throw new InvalidOperationException(Text.Get("AutoStartFailed"));
            using var deadline=CancellationTokenSource.CreateLinkedTokenSource(lifetime.Token);deadline.CancelAfter(TimeSpan.FromMinutes(10));
            var stdout=ReadAutoOutput(process.StandardOutput,deadline.Token);var stderr=ReadAutoOutput(process.StandardError,deadline.Token);
            try { await process.WaitForExitAsync(deadline.Token); } catch { if(!process.HasExited)process.Kill(true);throw; }
            await File.WriteAllTextAsync(Path.Combine(output,"latest.stdout.txt"),await stdout,lifetime.Token);
            await File.WriteAllTextAsync(Path.Combine(output,"latest.stderr.txt"),await stderr,lifetime.Token);
            if(process.ExitCode!=0)throw new InvalidOperationException(Text.Get("AutoProbeFailed")+" "+output);
            await model.RefreshAsync(lifetime.Token);
            model.SetNotice(Text.Get("AutoConfigured"));
        }); } finally { autoRunning=false; }
    }
    static async Task<string> ReadAutoOutput(StreamReader reader,CancellationToken ct)
    {
        var saved=new System.Text.StringBuilder();var buffer=new char[4096];bool truncated=false;int n;
        while((n=await reader.ReadAsync(buffer.AsMemory(),ct))>0){int keep=Math.Min(n,1024*1024-saved.Length);if(keep>0)saved.Append(buffer,0,keep);if(keep<n)truncated=true;}
        if(truncated)saved.Append("\n[diagnostic output truncated at 1 MiB]\n");return saved.ToString();
    }
    async void ApplyMode(object sender, RoutedEventArgs e)
    {
        if (AppGrid.SelectedItem is not ApplicationRow row)
        {
            ShowMessage("SelectApplication");
            return;
        }
        if(((string?)Mode.SelectedValue ?? "AUTO")=="AUTO"&&IsChrome(row.Profile)){await Execute(()=>ActivateChromeSites(row));return;}
        var copy=row.Profile.Copy();copy.Fallback=(string?)ProfileFallback.SelectedValue ?? "";
        await Execute(async () => { await model.SetProfileAsync(copy, (string?)Mode.SelectedValue ?? "AUTO", (string?)RelayChoice.SelectedValue, lifetime.Token); });
    }
    async void DeleteApplication(object sender, RoutedEventArgs e)
    {
        if (AppGrid.SelectedItem is not ApplicationRow row)
        {
            ShowMessage("SelectApplication");
            return;
        }
        if (!row.Profile.Custom)
        {
            ShowMessage("DefaultProfile");
            return;
        }
        if (PromptWindow.Show(this, "DeleteConfirm", true))
            await Execute(async()=>{if(IsChrome(row.Profile))ChromePolicy.Guard(row.Profile);await model.DeleteApplicationAsync(row.Profile,lifetime.Token);});
    }
    async void RunDiagnostics(object sender, RoutedEventArgs e) => await Execute(() => model.DiagnoseAsync(lifetime.Token));
    async void SaveSettings(object sender, RoutedEventArgs e)
    {
        try
        {
            if (model.Available && !await model.SaveServiceSettingsAsync(new()
            {
                Automatic = Automatic.IsChecked == true,
                Failover = (string?)Fallback.SelectedValue ?? "DIRECT",
                UpdateChannel = (string?)Channel.SelectedValue ?? "stable"
            }, lifetime.Token))
                return;
            LocalPreferences.SetAutoStart(Launch.IsChecked == true);
            SaveLocalPreferences(StartMinimized.IsChecked == true, MinimizeToTray.IsChecked == true);
            ShowMessage(model.Available ? "SettingsSaved" : "LocalSettingsSaved");
        }
        catch (OperationCanceledException) when (lifetime.IsCancellationRequested) { }
        catch { ShowMessage("SettingsFailed"); }
    }
    void ShowMessage(string key) => PromptWindow.Show(this, key, false);
}
