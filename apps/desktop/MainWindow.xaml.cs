using System.ComponentModel;
using System.Windows;
using System.Windows.Controls;
using System.Windows.Threading;
using Microsoft.Win32;
namespace Neexitlag.Desktop;
public partial class MainWindow : Window
{
    readonly DesktopViewModel model = new(new WindowsServiceClient()); readonly DispatcherTimer timer = new() { Interval = TimeSpan.FromSeconds(2) }; readonly CancellationTokenSource lifetime = new();
    readonly LocalPreferences preferences = LocalPreferences.Load(); readonly System.Windows.Forms.NotifyIcon tray; readonly System.Drawing.Icon trayImage; bool exiting, settingsLoaded;
    public MainWindow()
    {
        InitializeComponent();
 LocalPorts.Text="443";
        DataContext = model;
        Launch.IsChecked = LocalPreferences.AutoStart;
        StartMinimized.IsChecked = preferences.StartMinimized;
        MinimizeToTray.IsChecked = preferences.MinimizeToTray;
        Channel.ItemsSource = new[] { new Choice("stable", Text.Get("Stable")), new Choice("preview", Text.Get("Preview")) };
        Channel.SelectedValue = "stable";
        Fallback.ItemsSource = new[] { new Choice("DIRECT", Text.Get("Direct")) };
        Fallback.SelectedValue = "DIRECT";
        using var iconStream = System.Windows.Application.GetResourceStream(new Uri("pack://application:,,,/Assets/NeExitLag.ico"))!.Stream;
        using var loadedIcon = new System.Drawing.Icon(iconStream);
        trayImage = (System.Drawing.Icon)loadedIcon.Clone();
        tray = new()
        {
            Icon = trayImage,
            Text = "NeExitLag",
            Visible = true
        };
        var menu = new System.Windows.Forms.ContextMenuStrip();
        menu.Items.Add(Text.Get("OpenApp"), null, (_, _) => Dispatcher.Invoke(ShowFromTray));
        var enable = menu.Items.Add(Text.Get("TrayEnable"), null, async (_, _) => await Execute(() => model.OptimizeAsync(true, lifetime.Token)));
        var disable = menu.Items.Add(Text.Get("TrayDisable"), null, async (_, _) => await Execute(() => model.OptimizeAsync(false, lifetime.Token)));
        menu.Items.Add(new System.Windows.Forms.ToolStripSeparator());
        menu.Items.Add(Text.Get("Exit"), null, (_, _) => Dispatcher.Invoke(Exit));
        menu.Opening += (_, _) => { enable.Enabled = model.Available && !model.Busy && !model.Optimizing; disable.Enabled = model.Available && !model.Busy && model.Optimizing; };
        tray.ContextMenuStrip = menu;
        tray.DoubleClick += (_, _) => Dispatcher.Invoke(ShowFromTray);
        Loaded += async (_, _) => { if (preferences.StartMinimized) Hide(); await Refresh(); };
        timer.Tick += async (_, _) => await Refresh();
        timer.Start();
        Closing += OnClosing;
        StateChanged += (_, _) => { if (WindowState == WindowState.Minimized && preferences.MinimizeToTray) Hide(); };
        Closed += (_, _) => { timer.Stop(); lifetime.Cancel(); tray.Visible = false; tray.ContextMenuStrip?.Dispose(); tray.Dispose(); trayImage.Dispose(); System.Windows.Application.Current.Shutdown(); };
    }
    public void ShowFromTray()
    {
        Show();
        if (WindowState == WindowState.Minimized)
            WindowState = WindowState.Normal;
        Activate();
    }
    void Exit()
    {
        exiting = true;
        Close();
    }
    void OnClosing(object? sender, CancelEventArgs e)
    {
        if (!exiting && preferences.MinimizeToTray && (model.Available || WindowsServiceClient.IsRunning()))
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
        catch (Exception e) { model.SetUnavailable(e); }
    }
    async Task Refresh()
    {
        await Execute(() => model.RefreshAsync(lifetime.Token));
        if (!settingsLoaded && model.Snapshot is { } snapshot)
        {
            LoadSettings(snapshot);
            settingsLoaded = true;
        }
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
    LocalStrategyConfig localConfig = new();
    readonly System.Collections.ObjectModel.ObservableCollection<LocalStrategyProfile> localRules = [];
    static string[] Parts(string s)=>s.Split(',',StringSplitOptions.RemoveEmptyEntries|StringSplitOptions.TrimEntries);
    async void LoadLocal(object sender,RoutedEventArgs e)=>await Execute(async()=>{
        localConfig=await model.GetLocalStrategiesAsync(lifetime.Token);LocalEnabled.IsChecked=localConfig.Enabled;
        localRules.Clear();foreach(var p in localConfig.User)localRules.Add(p);LocalGrid.ItemsSource=localRules;
        var m=await model.GetLocalMetricsAsync(lifetime.Token);LocalCounters.Text=string.Format(Text.Russian,Text.Get("LocalCounters"),m.PacketsInspected,m.PacketsModified,m.ExtraPacketsGenerated,m.FlowsDirectPlus,Text.Bytes(m.BytesDirectPlus));
    });
    async void SaveLocal(object sender,RoutedEventArgs e)=>await Execute(async()=>{
        localConfig.Enabled=LocalEnabled.IsChecked==true;localConfig.User=localRules.ToArray();
        localConfig=await model.SaveLocalStrategiesAsync(localConfig,lifetime.Token);
        if(LocalAuto.IsChecked==true && AppGrid.SelectedItem is ApplicationRow row)await model.SetProfileAsync(row.Profile,"AUTO",null,lifetime.Token);
        LocalCounters.Text=Text.Get("LocalSaved");
    });
    void SelectLocal(object sender,SelectionChangedEventArgs e){if(LocalGrid.SelectedItem is not LocalStrategyProfile p)return;LocalApp.Text=p.Application;LocalPorts.Text=string.Join(",",p.Ports);LocalDomains.Text=string.Join(",",p.Domains);LocalIPs.Text=string.Join(",",p.Ips);LocalExcludeDomains.Text=string.Join(",",p.ExcludeDomains);LocalExcludeIPs.Text=string.Join(",",p.ExcludeIps);foreach(ComboBoxItem item in LocalMechanism.Items)if((string)item.Tag==p.Strategy)LocalMechanism.SelectedItem=item;}
    void AddLocal(object sender,RoutedEventArgs e){
        try{var old=LocalGrid.SelectedItem as LocalStrategyProfile;var strategy=(string)((ComboBoxItem)LocalMechanism.SelectedItem).Tag;
        var p=new LocalStrategyProfile{Id=old?.Id??Guid.NewGuid().ToString("N"),Name=System.IO.Path.GetFileNameWithoutExtension(LocalApp.Text),Application=LocalApp.Text,Protocol=strategy.StartsWith("udp-")?"udp":"tcp",Ports=Parts(LocalPorts.Text).Select(ushort.Parse).ToArray(),Domains=Parts(LocalDomains.Text),Ips=Parts(LocalIPs.Text),ExcludeDomains=Parts(LocalExcludeDomains.Text),ExcludeIps=Parts(LocalExcludeIPs.Text),Strategy=strategy,Cutoff=old?.Cutoff??2,Priority=old?.Priority??0,Enabled=true};
        if(old is not null)localRules.Remove(old);localRules.Add(p);LocalGrid.ItemsSource=localRules;LocalGrid.SelectedItem=p;
        }catch{ShowMessage("LocalSelect");}
    }
    void RemoveLocal(object sender,RoutedEventArgs e){if(LocalGrid.SelectedItem is LocalStrategyProfile p)localRules.Remove(p);}
    void TestLocal(object sender,RoutedEventArgs e){
        var picker=new OpenFileDialog{Title=Text.Get("LocalTestScript"),Filter="PowerShell|strategy-tester.ps1",CheckFileExists=true};if(picker.ShowDialog(this)!=true)return;
        var p=new System.Diagnostics.ProcessStartInfo("powershell.exe"){UseShellExecute=true,Verb="runas"};
        foreach(var a in new[]{"-NoProfile","-NoExit","-File",picker.FileName,((string)((ComboBoxItem)LocalMechanism.SelectedItem).Tag).StartsWith("udp-")?"-UDPTarget":"-URL",LocalURL.Text,"-TokenFile",LocalToken.Text})p.ArgumentList.Add(a);
        if(!string.IsNullOrWhiteSpace(LocalCA.Text)){p.ArgumentList.Add("-EndpointCA");p.ArgumentList.Add(LocalCA.Text);}try{System.Diagnostics.Process.Start(p);}catch{ShowMessage("LocalSelect");}
    }
    async void Reconnect(object sender, RoutedEventArgs e) => await Refresh();
    async void ToggleOptimize(object sender, RoutedEventArgs e) => await Execute(() => model.OptimizeAsync(!model.Optimizing, lifetime.Token));
    async void AddApplication(object sender, RoutedEventArgs e)
    {
        var picker = new OpenFileDialog { Title = Text.Get("ChooseExecutable"), Filter = Text.Get("ExecutableFilter"), CheckFileExists = true, Multiselect = false };
        if (picker.ShowDialog(this) == true)
            await Execute(async () => { await model.AddApplicationAsync(picker.FileName, lifetime.Token); });
    }
    void ApplicationSelected(object sender, SelectionChangedEventArgs e)
    {
        if (AppGrid.SelectedItem is ApplicationRow row)
        {
            Mode.SelectedValue = row.Profile.Mode;
            RelayChoice.SelectedValue = row.Profile.RelayId;
        }
    }
    void ModeChanged(object sender, SelectionChangedEventArgs e)
    {
        if (RelayChoice is not null)
            RelayChoice.IsEnabled = (string?)Mode.SelectedValue == "MANUAL";
    }
    async void ApplyMode(object sender, RoutedEventArgs e)
    {
        if (AppGrid.SelectedItem is not ApplicationRow row)
        {
            ShowMessage("SelectApplication");
            return;
        }
        await Execute(async () => { await model.SetProfileAsync(row.Profile, (string?)Mode.SelectedValue ?? "AUTO", (string?)RelayChoice.SelectedValue, lifetime.Token); });
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
            await Execute(() => model.DeleteApplicationAsync(row.Profile, lifetime.Token));
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
            preferences.StartMinimized = StartMinimized.IsChecked == true;
            preferences.MinimizeToTray = MinimizeToTray.IsChecked == true;
            preferences.Save();
            ShowMessage(model.Available ? "SettingsSaved" : "LocalSettingsSaved");
        }
        catch (OperationCanceledException) when (lifetime.IsCancellationRequested) { }
        catch { ShowMessage("SettingsFailed"); }
    }
    void ShowMessage(string key) => PromptWindow.Show(this, key, false);
}
