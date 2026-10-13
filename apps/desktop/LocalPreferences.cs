using System.Text.Json;
using Microsoft.Win32;
namespace Neexitlag.Desktop;
public sealed class LocalPreferences
{
    public bool StartMinimized
    {
        get; set;
    }
    public bool MinimizeToTray { get; set; } = true;
    static string PathName => Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData), "NeExitLag", "desktop.json");
    public static LocalPreferences Load()
    {
        try
        {
            return JsonSerializer.Deserialize<LocalPreferences>(File.ReadAllText(PathName)) ?? new();
        }
        catch { return new(); }
    }
    public void Save()
    {
        Directory.CreateDirectory(Path.GetDirectoryName(PathName)!);
        string temporary = PathName + ".tmp";
        File.WriteAllText(temporary, JsonSerializer.Serialize(this));
        File.Move(temporary, PathName, true);
    }
    public static bool AutoStart
    {
        get
        {
            using var key = Registry.CurrentUser.OpenSubKey(@"Software\Microsoft\Windows\CurrentVersion\Run");
            return key?.GetValue("NeExitLag") != null;
        }
    }
    public static void SetAutoStart(bool enabled)
    {
        using var key = Registry.CurrentUser.CreateSubKey(@"Software\Microsoft\Windows\CurrentVersion\Run");
        if (enabled && Environment.ProcessPath is string path)
            key.SetValue("NeExitLag", '"' + path + '"');
        else
            key.DeleteValue("NeExitLag", false);
    }
}
