using System.Buffers.Binary;
using System.IO;
using System.IO.Pipes;
using System.Net.Http;
using System.Text;
using System.Text.Json;
using System.Text.Json.Serialization;
namespace Neexitlag.Desktop;
public sealed class ServiceClient
{
    private static readonly JsonSerializerOptions Json = new() { PropertyNameCaseInsensitive = true };
    private readonly string tokenFile = Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.CommonApplicationData), "neexitlag", "ipc.token");
    private readonly Uri? demo; private readonly string? demoToken;
    public ServiceClient()
    {
        var args = Environment.GetCommandLineArgs();
        for (int i = 1; i < args.Length - 1; i++)
        {
            if (args[i] == "--demo-url")
            {
                var uri = new Uri(args[i + 1]);
                if (uri.Scheme != "http" || !uri.IsLoopback)
                    throw new InvalidOperationException("Demo endpoint must be loopback HTTP");
                demo = uri;
            }
            if (args[i] == "--demo-token-file")
                demoToken = File.ReadAllText(args[i + 1]).Trim();
        }
    }
    public async Task<T> Call<T>(string method, object? payload = null)
    {
        using var timeout = new CancellationTokenSource(TimeSpan.FromSeconds(5));
        string token = demo != null ? demoToken ?? throw new InvalidOperationException("Demo token is required") : File.Exists(tokenFile) ? (await File.ReadAllTextAsync(tokenFile, timeout.Token)).Trim() : throw new InvalidOperationException("Service not installed: IPC token is missing. Run the Windows service installer.");
        var id = Guid.NewGuid().ToString("N");
        var request = JsonSerializer.SerializeToUtf8Bytes(new
        {
            version = 1,
            id,
            sent_at = DateTimeOffset.UtcNow,
            token,
            method,
            payload
        });
        if (request.Length > 65536)
            throw new InvalidOperationException("Request is too large");
        byte[] reply;
        if (demo != null)
        {
            using var client = new HttpClient(new HttpClientHandler { AllowAutoRedirect = false }) { Timeout = TimeSpan.FromSeconds(5) };
            using var content = new StringContent(Encoding.UTF8.GetString(request), Encoding.UTF8, "application/json");
            using var result = await client.PostAsync(new Uri(demo, "/api/ipc"), content, timeout.Token);
            result.EnsureSuccessStatusCode();
            reply = await result.Content.ReadAsByteArrayAsync(timeout.Token);
            if (reply.Length > 65536)
                throw new InvalidOperationException("Service response is too large");
        }
        else
        {
            using var pipe = new NamedPipeClientStream(".", "neexitlag-v1", PipeDirection.InOut, PipeOptions.Asynchronous);
            try
            {
                await pipe.ConnectAsync(2000, timeout.Token);
            }
            catch (TimeoutException) { throw new InvalidOperationException("Service stopped or unavailable: named pipe connection timed out."); }
            var header = new byte[4];
            BinaryPrimitives.WriteUInt32BigEndian(header, (uint)request.Length);
            await pipe.WriteAsync(header, timeout.Token);
            await pipe.WriteAsync(request, timeout.Token);
            await pipe.FlushAsync(timeout.Token);
            await pipe.ReadExactlyAsync(header, timeout.Token);
            int length = checked((int)BinaryPrimitives.ReadUInt32BigEndian(header));
            if (length < 1 || length > 65536)
                throw new InvalidOperationException("Invalid service response length");
            reply = new byte[length];
            await pipe.ReadExactlyAsync(reply, timeout.Token);
        }
        using var document = JsonDocument.Parse(reply);
        var root = document.RootElement;
        if (root.GetProperty("version").GetInt32() != 1)
            throw new InvalidOperationException("Service incompatible version; update client and service together");
        if (root.TryGetProperty("id", out var responseId) && responseId.GetString() != id)
            throw new InvalidOperationException("Service response ID mismatch");
        if (!root.GetProperty("ok").GetBoolean())
            throw new InvalidOperationException(root.GetProperty("error").GetString() ?? "Service request failed");
        return root.GetProperty("data").Deserialize<T>(Json) ?? throw new InvalidOperationException("Invalid service response data");
    }
}
public sealed class Profile
{
    public string id { get; set; } = ""; public string name { get; set; } = ""; public string[] executables { get; set; } = []; public string mode { get; set; } = "DIRECT"; public string? relay_id
    {
        get; set;
    }
    public bool custom
    {
        get; set;
    }
    public string? publisher
    {
        get; set;
    }
    public string? signature_sha256
    {
        get; set;
    }
    public string[]? protocol_hints
    {
        get; set;
    }
    public ushort[]? port_hints
    {
        get; set;
    }
    [JsonIgnore] public string Executable => string.Join(", ", executables);
}
public sealed class Route
{
    public string id { get; set; } = ""; public double rtt_ms
    {
        get; set;
    }
    public double jitter_ms
    {
        get; set;
    }
    public double loss
    {
        get; set;
    }
    public double? score
    {
        get; set;
    }
    public string status { get; set; } = "unavailable"; [JsonIgnore] public string RTT => status == "healthy" ? $"{rtt_ms:F1} ms" : "—"; [JsonIgnore] public string Jitter => $"{jitter_ms:F1} ms"; [JsonIgnore] public string Loss => $"{loss:P1}";
}
public sealed class BackendStatus
{
    public string backend { get; set; } = "none"; public bool ready
    {
        get; set;
    }
    public string? error
    {
        get; set;
    }
}
public sealed class Snapshot
{
    public bool connected
    {
        get; set;
    }
    public bool optimizing
    {
        get; set;
    }
    public bool demo
    {
        get; set;
    }
    public string Primary { get; set; } = "DIRECT"; public string Secondary { get; set; } = ""; public string Tertiary { get; set; } = ""; public Route[] routes { get; set; } = []; public Profile[] profiles { get; set; } = []; public ulong rx_bytes
    {
        get; set;
    }
    public ulong tx_bytes
    {
        get; set;
    }
    public string[] active_apps { get; set; } = []; public BackendStatus backend { get; set; } = new(); public JsonElement history
    {
        get; set;
    }
    public JsonElement settings
    {
        get; set;
    }
}
public sealed class Check
{
    public string name { get; set; } = ""; public string status { get; set; } = ""; public string detail { get; set; } = "";
}
