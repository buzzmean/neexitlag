// Native Windows regression host. Loads the production compiled BAML and runs
// real Window/NotifyIcon handlers on an STA dispatcher, not an XML approximation.
using System.IO;
using System.Runtime.InteropServices;
using System.Windows;
using System.Windows.Controls;
using System.Windows.Controls.Primitives;
using System.Windows.Input;
using System.Windows.Media;
using System.Windows.Media.Imaging;
using System.Windows.Threading;
using Neexitlag.Desktop;
using Application = System.Windows.Application;
using ComboBox = System.Windows.Controls.ComboBox;
using Point = System.Windows.Point;

internal static class Program
{
    static readonly string Output = Path.GetFullPath("bin/native-ui-tests");
    static string settings = "";
    [DllImport("user32.dll")] static extern bool SetCursorPos(int x, int y);
    static void Require(bool condition, string message) { if (!condition) throw new Exception(message); }
    static void Pump()
    {
        for (int i = 0; i < 4; i++)
        {
            var frame = new DispatcherFrame();
            Dispatcher.CurrentDispatcher.BeginInvoke(DispatcherPriority.ApplicationIdle, new Action(() => frame.Continue = false));
            Dispatcher.PushFrame(frame);
            Thread.Sleep(25);
        }
    }
    static void Pass(string label) => Console.WriteLine("PASS " + label);
    static MainWindow Window(LocalPreferences p) => new(p, false, false, settings);
    static void Key(ComboBox box, System.Windows.Input.Key key)
    {
        box.RaiseEvent(new System.Windows.Input.KeyEventArgs(Keyboard.PrimaryDevice, PresentationSource.FromVisual(box)!, 0, key) { RoutedEvent = Keyboard.KeyDownEvent });
        Pump();
    }
    static void SaveImage(FrameworkElement element, string name)
    {
        element.UpdateLayout();
        Require(element.ActualWidth > 0 && element.ActualHeight > 0, "Empty visual: " + name);
        var image = new RenderTargetBitmap((int)Math.Ceiling(element.ActualWidth), (int)Math.Ceiling(element.ActualHeight), 96, 96, PixelFormats.Pbgra32);
        image.Render(element);
        var encoder = new PngBitmapEncoder(); encoder.Frames.Add(BitmapFrame.Create(image));
        using var stream = File.Create(Path.Combine(Output, name + ".png")); encoder.Save(stream);
    }
    static string Color(System.Windows.Media.Brush brush) => ((SolidColorBrush)brush).Color.ToString();
    static void Brush(System.Windows.Media.Brush brush, string hex, string message) => Require(Color(brush) == hex, message + ": " + Color(brush));
    static void Hover(FrameworkElement target)
    {
        var p = target.PointToScreen(new Point(target.ActualWidth / 2, target.ActualHeight / 2));
        Require(SetCursorPos((int)p.X, (int)p.Y), "Cannot position native mouse"); Pump();
        Require(target.IsMouseOver, "Mouse did not reach actual visual");
    }
    sealed class PickerService : IServiceClient
    {
        public Task<T> Call<T>(string method, object? payload = null, CancellationToken cancellationToken = default)
        {
            if (method != "processes") throw new Exception("Unexpected picker request " + method);
            return Task.FromResult((T)(object)new RunningApplicationList { Applications = [new() {
                Name = "long-application.exe", DisplayName = "A long application title that wraps onto multiple lines in the application picker",
                Path = @"C:\Applications\A very long folder name\Another long folder name\Application\bin\long-application.exe",
                Pid = 42, Pids = Enumerable.Range(1000, 20).Select(x => (uint)x).ToArray(), ProcessCount = 20
            }] });
        }
    }
    static void Picker()
    {
        var owner = Window(new() { MinimizeToTray = false }); owner.Show(); Pump();
        var picker = new AddApplicationWindow(owner, new DesktopViewModel(new PickerService()), CancellationToken.None);
        picker.Show(); Pump();
        var grid = (DataGrid)picker.FindName("ProcessGrid");
        Require(double.IsNaN(grid.RowHeight), "Picker must use content-driven row height, not inherited fixed height");
        Require(grid.MinRowHeight == 82, "Picker minimum row height lost");
        // Enrichment is asynchronous; pump until the service metadata is displayed.
        var deadline = System.Diagnostics.Stopwatch.StartNew();
        while (grid.Items.Count == 0 && deadline.Elapsed < TimeSpan.FromSeconds(10)) Pump();
        Require(grid.Items.Count == 1, "Picker metadata did not render"); grid.UpdateLayout();
        var row = (DataGridRow)grid.ItemContainerGenerator.ContainerFromIndex(0);
        Require(row.ActualHeight > 82, "Long title/path/PIDs must expand the row");
        Require(((System.Windows.Controls.Button)picker.FindName("ManualButton")).IsEnabled, "Manual file picker unavailable");
        SaveImage(picker, "add-application"); picker.Close(); owner.Close(); Pump();
        Pass("Add application window opens (production BAML, long title/path/PIDs)");
    }
    static void Lifecycle()
    {
        foreach (bool start in new[] { false, true }) foreach (bool tray in new[] { false, true })
        {
            var prefs = new LocalPreferences { StartMinimized = start, MinimizeToTray = tray };
            prefs.Save(settings);
            // Recreate with persisted preferences, as on GUI restart.
            var w = Window(LocalPreferences.Load(settings)); bool closed = false; w.Closed += (_, _) => closed = true;
            w.Show(); Pump();
            Require(w.TrayVisible == tray, "Startup tray visibility");
            Require(w.IsVisible == !(start && tray), "Startup must hide only in tray mode");
            Require(!start || w.WindowState == WindowState.Minimized, "StartMinimized must minimize");
            Require(w.ShowInTaskbar, "Normal minimized window must remain in taskbar");
            w.TrayIcon.ContextMenuStrip!.Items[0].PerformClick(); Pump(); Require(w.IsVisible && w.WindowState == WindowState.Normal, "Tray Open restore failed");
            w.WindowState = WindowState.Minimized; Pump();
            Require(w.IsVisible == !tray, "Minimize visibility does not match tray setting");
            w.ShowFromTray(); w.Close(); Pump();
            Require(closed == !tray, "X must close when OFF, hide when ON even with service absent");
            if (tray)
            {
                Require(!w.IsVisible && w.TrayVisible, "X tray state");
                w.ShowFromTray(); Pump(); Require(w.IsVisible, "Open after X");
                w.TrayIcon.ContextMenuStrip!.Items[^1].PerformClick(); Pump(); Require(closed && !w.TrayVisible, "Tray Exit must really close and remove icon");
            }
            Pass($"StartMinimized={start} / MinimizeToTray={tray}: startup, minimize, X, Open, Exit, restart");
        }
        var live = Window(new() { MinimizeToTray = true }); live.Show(); Pump();
        live.WindowState = WindowState.Minimized; Pump(); Require(!live.IsVisible, "Initial tray minimize");
        live.SaveLocalPreferences(true, false); Pump();
        Require(live.IsVisible && live.WindowState == WindowState.Minimized && !live.TrayVisible, "Save OFF must immediately restore taskbar window and remove tray");
        var saved = LocalPreferences.Load(settings);
        Require(saved.StartMinimized && !saved.MinimizeToTray, "OFF preference persistence");
        live.SaveLocalPreferences(false, true); Pump();
        Require(!live.IsVisible && live.TrayVisible, "Save ON must immediately enable tray for already minimized window");
        saved = LocalPreferences.Load(settings); Require(!saved.StartMinimized && saved.MinimizeToTray, "ON preference persistence");
        live.Exit(); Pump();
        Pass("MinimizeToTray OFF"); Pass("MinimizeToTray ON"); Pass("StartMinimized combinations"); Pass("Save applies immediately and preferences persist");
    }
    static void Dropdowns()
    {
        var main = Window(new() { MinimizeToTray = false }); main.Show(); Pump();
        var style = (Style)Application.Current.FindResource(typeof(ComboBox));
        foreach (string name in new[] { "Mode", "RelayChoice", "AppPreset", "LocalMechanism", "Fallback", "Channel" })
        {
            string section = name is "Fallback" or "Channel" ? "Settings" : name == "LocalMechanism" ? "LocalStrategies" : "Applications";
            foreach (string page in new[] { "Home", "Applications", "Routes", "Statistics", "Diagnostics", "Settings", "LocalStrategies" })
                ((FrameworkElement)main.FindName(page)).Visibility = page == section ? Visibility.Visible : Visibility.Collapsed;
            var box = (ComboBox)main.FindName(name); box.BringIntoView(); Pump(); box.ApplyTemplate();
            Require(box.Style == style, name + " does not use shared dark style");
            Require(box.Template.FindName("Chrome", box) is Border, name + " native light template still active");
            Require(box.Template.FindName("PART_Popup", box) is Popup, name + " missing popup");
            if (!box.IsEnabled) { Brush(box.Foreground, "#FF9BA9BA", name + " disabled foreground"); }
            SaveImage(box, "actual-" + name);
        }
        main.Close(); Pump();
        var combo = new ComboBox { Width = 300, ItemsSource = new[] { "Normal / selected", "Hover / keyboard", "Third item" }, SelectedIndex = 0 };
        var next = new System.Windows.Controls.Button { Content = "Next focus target" };
        var panel = new StackPanel { Margin = new Thickness(20) }; panel.Children.Add(combo); panel.Children.Add(next);
        var host = new System.Windows.Window { Title = "Native dark ComboBox states", Width = 420, Height = 240, Background = new SolidColorBrush(System.Windows.Media.Color.FromRgb(16,21,28)), Content = panel, WindowStartupLocation = WindowStartupLocation.CenterScreen };
        host.Show(); host.Activate(); Pump();
        SetCursorPos(0, 0); Pump();
        var chrome = (Border)combo.Template.FindName("Chrome", combo);
        Brush(chrome.Background, "#FF191E25", "Normal background"); Brush(combo.Foreground, "#FFEDF2F6", "Normal text");
        SaveImage(host, "combo-normal"); Pass("dark ComboBox normal");
        Hover(combo); Brush(chrome.Background, "#FF28313C", "Hover background");
        SaveImage(host, "combo-hover"); Pass("dark ComboBox hover");
        combo.Focus(); Key(combo, System.Windows.Input.Key.Down); Require(combo.SelectedIndex == 1, "Arrow selection broken");
        Key(combo, System.Windows.Input.Key.F4); Require(combo.IsDropDownOpen, "F4 dropdown keyboard open broken"); Pump();
        var popup = (Popup)combo.Template.FindName("PART_Popup", combo);
        var border = (Border)combo.Template.FindName("PopupBorder", combo);
        Brush(border.Background, "#FF191E25", "Popup background");
        var selected = (ComboBoxItem)combo.ItemContainerGenerator.ContainerFromIndex(combo.SelectedIndex);
        selected.ApplyTemplate(); Brush(selected.Background, "#FF33452C", "Selected background"); Brush(selected.Foreground, "#FFB7F46E", "Selected text");
        SaveImage(border, "combo-popup-selected"); Pass("dark ComboBox selected");
        var unselected = (ComboBoxItem)combo.ItemContainerGenerator.ContainerFromIndex(2); Hover(unselected);
        Brush(unselected.Background, "#FF28313C", "Popup item hover"); SaveImage(border, "combo-popup-hover");
        Key(combo, System.Windows.Input.Key.Escape); Require(!combo.IsDropDownOpen, "Escape must close popup");
        Key(combo, System.Windows.Input.Key.F4); Key(combo, System.Windows.Input.Key.Enter); Require(!combo.IsDropDownOpen, "Enter must close popup");
        combo.Focus(); combo.MoveFocus(new TraversalRequest(FocusNavigationDirection.Next)); Pump(); Require(next.IsKeyboardFocused, "Tab traversal broken");
        combo.Focus(); Pump(); Brush(chrome.BorderBrush, "#FFB7F46E", "Keyboard focus border");
        combo.IsEnabled = false; Pump(); Brush(combo.Foreground, "#FF9BA9BA", "Disabled text"); Brush(chrome.Background, "#FF20262F", "Disabled background");
        Require(Color(combo.Foreground) != Color(chrome.Background), "Disabled foreground equals background");
        SaveImage(host, "combo-disabled"); Pass("dark ComboBox disabled"); Pass("ComboBox keyboard arrows/F4/Enter/Escape/Tab and focus");
        host.Close(); Pump();
    }
    [STAThread]
    static int Main()
    {
        Directory.CreateDirectory(Output); settings = Path.Combine(Path.GetTempPath(), "neexitlag-native-" + Guid.NewGuid() + ".json");
        // Match the published application's pack://application resource root;
        // this console host's entry assembly contains no production assets.
        Application.ResourceAssembly = typeof(App).Assembly;
        var app = new App(); app.InitializeComponent(); app.ShutdownMode = ShutdownMode.OnExplicitShutdown;
        try { Picker(); Lifecycle(); Dropdowns(); Pass("native WPF desktop regression suite"); return 0; }
        catch (Exception e) { Console.Error.WriteLine(e); File.WriteAllText(Path.Combine(Output, "failure.txt"), e.ToString()); return 1; }
        finally { foreach (System.Windows.Window w in app.Windows.Cast<System.Windows.Window>().ToArray()) { if (w is MainWindow m) m.Exit(); else w.Close(); } app.Shutdown(); File.Delete(settings); }
    }
}
