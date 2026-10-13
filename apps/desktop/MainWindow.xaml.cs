using System.Windows;
using System.Windows.Controls;
using System.Windows.Media;
using System.Windows.Shapes;
using System.Windows.Threading;
using Microsoft.Win32;
namespace Neexitlag.Desktop;
public partial class MainWindow : Window
{
    private readonly ServiceClient client = new(); private readonly DispatcherTimer timer = new() { Interval = TimeSpan.FromSeconds(2) }; private Snapshot? snapshot; private bool busy; private string page = "Home";
    public MainWindow()
    {
        InitializeComponent();
        Launch.IsChecked = Registry.CurrentUser.OpenSubKey(@"Software\Microsoft\Windows\CurrentVersion\Run")?.GetValue("neexitlag") != null;
        Loaded += async (_, _) => await Refresh();
        timer.Tick += async (_, _) => await Refresh();
        timer.Start();
        Closed += (_, _) => timer.Stop();
    }
    private void Navigate(object sender, RoutedEventArgs e)
    {
        page = ((Button)sender).Tag.ToString()!;
        PageTitle.Text = page;
        foreach (var name in new[] { "Home", "Applications", "Routes", "Statistics", "Diagnostics", "Settings" })
            ((FrameworkElement)FindName(name)).Visibility = name == page ? Visibility.Visible : Visibility.Collapsed;
    }
    private async Task Refresh()
    {
        if (busy)
            return;
        busy = true;
        try
        {
            var s = await client.Call<Snapshot>("status");
            snapshot = s;
            Connection.Text = s.connected ? s.demo ? "Connected · Local lab" : "Connected · Windows service" : "Disconnected";
            Backend.Text = s.demo ? "Portable demo: real relay traffic, no Windows interception" : s.backend.ready ? "WinDivert active · IPv4 UDP · TCP / IPv6 DIRECT" : s.backend.error ?? "Capture backend unavailable; networking is DIRECT";
            Primary.Text = s.Primary;
            Standby.Text = $"Standby: {s.Secondary} / {s.Tertiary}";
            Optimize.Content = s.optimizing ? "Pause optimization" : "Optimize";
            var direct = s.routes.FirstOrDefault(r => r.id == "DIRECT");
            var selected = s.routes.FirstOrDefault(r => r.id == s.Primary);
            DirectRTT.Text = direct?.RTT ?? "—";
            OptimizedRTT.Text = selected?.RTT ?? "—";
            Jitter.Text = selected?.Jitter ?? "—";
            Loss.Text = selected?.Loss ?? "—";
            Traffic.Text = $"{s.rx_bytes:N0} / {s.tx_bytes:N0} B";
            ActiveApps.Text = s.active_apps.Length == 0 ? "None" : string.Join(", ", s.active_apps);
            if (page != "Applications" || AppGrid.SelectedItem == null)
                AppGrid.ItemsSource = s.profiles;
            RouteGrid.ItemsSource = s.routes;
            if (page != "Settings")
            {
                Automatic.IsChecked = s.settings.GetProperty("automatic").GetBoolean();
                Fallback.SelectedIndex = s.settings.GetProperty("failover").GetString() == "BLOCK" ? 1 : 0;
                Channel.SelectedIndex = s.settings.GetProperty("update_channel").GetString() == "preview" ? 1 : 0;
            }
            Plot(s);
            Error.Text = "";
        }
        catch (Exception e) { Connection.Text = "Disconnected"; Error.Text = e is OperationCanceledException ? "Service timeout: networking remains DIRECT." : e.Message; }
        finally { busy = false; }
    }
    private async Task Act(Func<Task> action)
    {
        if (busy)
            return;
        busy = true;
        try
        {
            await action();
            Error.Text = "";
        }
        catch (Exception e) { Error.Text = e.Message; }
        finally { busy = false; }
    }
    private async void ToggleOptimize(object sender, RoutedEventArgs e)
    {
        await Act(async () => { snapshot = await client.Call<Snapshot>("optimize", new { enabled = !(snapshot?.optimizing ?? false) }); });
        await Refresh();
    }
    private async void AddApplication(object sender, RoutedEventArgs e)
    {
        var picker = new OpenFileDialog { Title = "Choose application executable", Filter = "Windows applications (*.exe)|*.exe", CheckFileExists = true };
        if (picker.ShowDialog() != true)
            return;
        var profile = new Profile { id = "custom-" + Guid.NewGuid().ToString("N"), name = System.IO.Path.GetFileNameWithoutExtension(picker.FileName), executables = [picker.FileName], mode = "AUTO", custom = true };
        await Act(async () => { AppGrid.ItemsSource = await client.Call<Profile[]>("profile.set", ProfileRequest(profile)); });
    }
    private static object ProfileRequest(Profile p) => new { p.id, p.name, p.executables, p.mode, p.relay_id, p.custom, p.publisher, p.signature_sha256, p.protocol_hints, p.port_hints };
    private async void ApplyMode(object sender, RoutedEventArgs e)
    {
        if (AppGrid.SelectedItem is not Profile p)
            return;
        p.mode = ((ComboBoxItem)Mode.SelectedItem).Content.ToString()!;
        if (p.mode == "MANUAL")
        {
            if (RouteGrid.SelectedItem is Route r && r.id != "DIRECT")
                p.relay_id = r.id;
            else
                p.relay_id = snapshot?.routes.FirstOrDefault(r => r.id != "DIRECT" && r.status == "healthy")?.id;
            if (p.relay_id == null)
            {
                Error.Text = "Choose an available relay before using MANUAL mode.";
                return;
            }
        }
        await Act(async () => { AppGrid.ItemsSource = await client.Call<Profile[]>("profile.set", ProfileRequest(p)); });
    }
    private async void RunDiagnostics(object sender, RoutedEventArgs e)
    {
        await Act(async () => { Checks.ItemsSource = await client.Call<Check[]>("diagnostics"); });
    }
    private async void SaveSettings(object sender, RoutedEventArgs e)
    {
        await Act(async () => { await client.Call<System.Text.Json.JsonElement>("settings.set", new { automatic = Automatic.IsChecked == true, failover = ((ComboBoxItem)Fallback.SelectedItem).Content.ToString(), update_channel = ((ComboBoxItem)Channel.SelectedItem).Content.ToString() }); using var run = Registry.CurrentUser.CreateSubKey(@"Software\Microsoft\Windows\CurrentVersion\Run"); if (Launch.IsChecked == true && Environment.ProcessPath is string path) run.SetValue("neexitlag", '"' + path + '"'); else run.DeleteValue("neexitlag", false); });
    }
    private void Plot(Snapshot s)
    {
        Chart.Children.Clear();
        if (!s.history.TryGetProperty("measurements", out var measurements) || measurements.ValueKind != System.Text.Json.JsonValueKind.Array)
            return;
        MeasurementGrid.ItemsSource = measurements.EnumerateArray().TakeLast(16).Reverse().Select(x => new {
            Time = x.GetProperty("observed_at").GetDateTime().ToLocalTime().ToString("T"),
            Route = x.GetProperty("route_id").GetString(),
            RTT = x.GetProperty("available").GetBoolean() ? $"{x.GetProperty("rtt").GetDouble() / 1e6:F1} ms" : "—",
            Jitter = $"{x.GetProperty("jitter").GetDouble() / 1e6:F1} ms",
            Loss = $"{x.GetProperty("loss").GetDouble():P1}"
        }).ToArray();
        var all = measurements.EnumerateArray().Where(x => x.GetProperty("available").GetBoolean()).ToArray();
        double width = Math.Max(600, Chart.ActualWidth), max = Math.Max(1, all.Select(x => x.GetProperty("rtt").GetDouble() / 1e6).DefaultIfEmpty(1).Max());
        var ids = all.Select(x => x.GetProperty("route_id").GetString()).Distinct().ToArray();
        Color[] colors = [Color.FromRgb(183, 244, 110), Color.FromRgb(112, 183, 255), Color.FromRgb(244, 185, 110), Color.FromRgb(181, 148, 239)];
        for (int n = 0; n < ids.Length; n++)
        {
            var points = all.Where(x => x.GetProperty("route_id").GetString() == ids[n]).ToArray();
            var line = new Polyline { Stroke = new SolidColorBrush(colors[n % colors.Length]), StrokeThickness = 2 };
            for (int i = 0; i < points.Length; i++)
                line.Points.Add(new Point(20 + (width - 40) * i / Math.Max(1, points.Length - 1), 200 - 170 * (points[i].GetProperty("rtt").GetDouble() / 1e6) / max));
            Chart.Children.Add(line);
            var label = new TextBlock { Text = ids[n], Foreground = line.Stroke };
            Canvas.SetLeft(label, 20 + n * 150);
            Canvas.SetTop(label, 5);
            Chart.Children.Add(label);
        }
        if (s.history.TryGetProperty("changes", out var changes) && changes.ValueKind == System.Text.Json.JsonValueKind.Array)
            Changes.ItemsSource = changes.EnumerateArray().Select(c => $"{c.GetProperty("At").GetDateTime():T}  {c.GetProperty("From").GetString()} → {c.GetProperty("To").GetString()}").ToArray();
    }
}
