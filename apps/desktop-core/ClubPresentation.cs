using System.Text.Json;
namespace Neexitlag.Desktop;
public static class ClubPresentation
{
    public static bool Enabled(string? json)
    {
        try { using var d=JsonDocument.Parse(json??"");return d.RootElement.GetProperty("version").GetInt32()==1&&d.RootElement.GetProperty("enabled").GetBoolean(); }
        catch(Exception e) when(e is JsonException or InvalidOperationException or KeyNotFoundException or FormatException){return false;}
    }
    public static string Summary(bool connected,bool optimizing,bool capture) =>
        connected&&optimizing&&capture ? "VELIQ · защита включена" : "VELIQ · подготовка подключения…";
}
