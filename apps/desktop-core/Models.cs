using System.Text.Json;
using System.Text.Json.Serialization;
namespace Neexitlag.Desktop;
public static class WireJson
{
    public static readonly JsonSerializerOptions Options = new() { PropertyNameCaseInsensitive = true, PropertyNamingPolicy = JsonNamingPolicy.SnakeCaseLower };
}
public sealed class Profile
{
    public string Id { get; set; } = ""; public string Name { get; set; } = ""; public string[] Executables { get; set; } = [];
    public string Mode { get; set; } = "DIRECT"; public string? RelayId
    {
        get; set;
    }
    public bool Custom
    {
        get; set;
    }
    public string? Publisher
    {
        get; set;
    }
    public string? SignatureSha256
    {
        get; set;
    }
    public string[]? ProtocolHints
    {
        get; set;
    }
    public ushort[]? PortHints
    {
        get; set;
    }
    public Profile Copy() => new() { Id = Id, Name = Name, Executables = [.. Executables], Mode = Mode, RelayId = RelayId, Custom = Custom, Publisher = Publisher, SignatureSha256 = SignatureSha256, ProtocolHints = ProtocolHints, PortHints = PortHints };
}
public sealed class Route
{
    public string Id { get; set; } = ""; public double RttMs
    {
        get; set;
    }
    public double JitterMs
    {
        get; set;
    }
    public double Loss
    {
        get; set;
    }
    public double? Score
    {
        get; set;
    }
    public string Status { get; set; } = "unavailable"; public DateTimeOffset ObservedAt
    {
        get; set;
    }
}
public sealed class Capabilities
{
    public bool Udp
    {
        get; set;
    }
    public bool Tcp
    {
        get; set;
    }
    public bool Ipv4
    {
        get; set;
    }
    public bool Ipv6
    {
        get; set;
    }
}
public sealed class BackendStatus
{
    public string Backend { get; set; } = "none"; public bool Ready
    {
        get; set;
    }
    public string? Error
    {
        get; set;
    }
    public Capabilities Capabilities { get; set; } = new();
}
public sealed class ServiceSettings
{
    public bool Automatic
    {
        get; set;
    }
    public string Failover { get; set; } = "DIRECT"; public string UpdateChannel { get; set; } = "stable";
}
public sealed class Snapshot
{
    public string Version { get; set; } = ""; public bool Connected
    {
        get; set;
    }
    public bool Optimizing
    {
        get; set;
    }
    public bool Demo
    {
        get; set;
    }
    public string Primary { get; set; } = ""; public string Secondary { get; set; } = ""; public string Tertiary { get; set; } = "";
    public Route[] Routes { get; set; } = []; public Profile[] Profiles { get; set; } = []; public ulong RxBytes
    {
        get; set;
    }
    public ulong TxBytes
    {
        get; set;
    }
    public string[] ActiveApps { get; set; } = []; public BackendStatus Backend { get; set; } = new(); public JsonElement History
    {
        get; set;
    }
    public ServiceSettings Settings { get; set; } = new(); public string ControlPlane { get; set; } = "";
}
// Existing processes IPC serializes Go Identity fields using their original names.
public sealed class ProcessIdentity
{
    public uint PID
    {
        get; set;
    }
    public string Path { get; set; } = ""; public string Name { get; set; } = ""; [JsonPropertyName("CreationTime")]
    public ulong CreationTime
    {
        get; set;
    }
}
public sealed class Check
{
    public string Name { get; set; } = ""; public string Status { get; set; } = ""; public string Detail { get; set; } = "";
}
public sealed class CaptureEvidence
{
    public bool Truncated
    {
        get; set;
    }
    public CaptureFlow[] Flows { get; set; } = [];
}
public sealed class CaptureFlow
{
    public uint Pid
    {
        get; set;
    }
    public string Exe { get; set; } = ""; public ulong CreationTime
    {
        get; set;
    }
    public string SelectedRoute { get; set; } = ""; public string? RelayId
    {
        get; set;
    }
    public DateTimeOffset LastSeen
    {
        get; set;
    }
}
