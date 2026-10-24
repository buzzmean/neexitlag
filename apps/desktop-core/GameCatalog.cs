using System.Collections.ObjectModel;
using System.Reflection;
using System.Security.Cryptography;
using System.Text.Json;
namespace Neexitlag.Desktop;

public sealed record GameDefinition(string Id, string Name, string Icon, string[] Executables, string Launcher, string AntiCheat, string DefaultMode, string UdpPreset, bool LauncherNeedsBypass);
public sealed record GameCatalogData(int Version, GameDefinition[] Games);
public static class GameCatalog
{
    public const string BundledSha256 = "65409160B1CE9EF67F1DFD5724D8D5FB762B5D989D091EB9363964D26FB32BE7";
    // No unsigned imports/network updates. Catalog trust follows the installed
    // assembly. A public update envelope requires an owner-pinned signing key.
    public static GameCatalogData Bundled() {
        using var stream=typeof(GameCatalog).Assembly.GetManifestResourceStream("Neexitlag.Desktop.Resources.GameCatalog.json") ?? throw new InvalidDataException("GAME_CATALOG_MISSING");
        using var buffer=new MemoryStream();stream.CopyTo(buffer);
        return Parse(buffer.ToArray(),BundledSha256);
    }
    public static GameCatalogData Parse(byte[] json,string expectedSha256) {
        if(json.Length>65536 || !Convert.ToHexString(SHA256.HashData(json)).Equals(expectedSha256,StringComparison.OrdinalIgnoreCase))throw new InvalidDataException("GAME_CATALOG_INTEGRITY_FAILED");
        var c=JsonSerializer.Deserialize<GameCatalogData>(json,new JsonSerializerOptions{PropertyNameCaseInsensitive=true}) ?? throw new InvalidDataException("GAME_CATALOG_INVALID");
        if(c.Version!=1 || c.Games is null || c.Games.Length is <1 or >64 || c.Games.Select(g=>g.Id).Distinct().Count()!=c.Games.Length)throw new InvalidDataException("GAME_CATALOG_VERSION_OR_DUPLICATE");
        foreach(var g in c.Games)if(g.Id.Length is <1 or >32 || g.Name.Length is <1 or >100 || g.DefaultMode!="DIRECT" || g.Executables is null || g.Executables.Length is <1 or >4 || g.Executables.Any(e=>e.Length>120||e.Contains('/')||e.Contains('\\')||!e.EndsWith(".exe",StringComparison.OrdinalIgnoreCase)))throw new InvalidDataException("GAME_CATALOG_SCOPE_INVALID");
        return c;
    }
    public static string RiotActivity(JsonElement status) {
        if(status.ValueKind==JsonValueKind.Object && status.TryGetProperty("records",out var records) && records.ValueKind==JsonValueKind.Array)
            foreach(var r in records.EnumerateArray())
                if(r.TryGetProperty("received_bytes",out var rx) && rx.TryGetUInt64(out var bytes) && bytes>0 && r.TryGetProperty("policy",out var p) && p.TryGetProperty("action",out var a) && a.GetString() is "DIRECT_PLUS" or "RELAY_REQUIRED" or "RELAY")
                    return "Лаунчер: наблюдается LOCAL/RELAY обмен; запуск ещё не подтверждён";
        return "Лаунчер: ожидаем подтверждённый обход; DIRECT может быть достаточен";
    }
    public static GameCatalogData VerifySigned(byte[] json,byte[] signature,RSA ownerPinnedKey) {
        if(json.Length>65536 || !ownerPinnedKey.VerifyData(json,signature,HashAlgorithmName.SHA256,RSASignaturePadding.Pss))throw new InvalidDataException("GAME_CATALOG_SIGNATURE_INVALID");
        return Parse(json,Convert.ToHexString(SHA256.HashData(json)));
    }
}
public sealed class GameCard(GameDefinition definition):Observable {
    public GameDefinition Definition {get;}=definition;
    public string Id=>Definition.Id;public string Name=>Definition.Name;public string Icon=>Definition.Icon;
    public string Launcher=>Definition.Launcher;public string AntiCheat=>"Античит: "+Definition.AntiCheat;
    public bool CanRepair=>Definition.LauncherNeedsBypass;
    public string Compatibility=>Definition.AntiCheat.Contains("Vanguard")||Definition.AntiCheat.Contains("FACEIT") ? "Совместимость обхода с Vanguard/FACEIT не проверена. Автоматически игру не трогаем." : "Снижение пинга не подтверждено измерениями. По умолчанию DIRECT.";
    string mode="DIRECT",status="Обычный маршрут; изменения только по вашему выбору";
    public string Mode {get=>mode;set=>Set(ref mode,value);}
    public string Status {get=>status;set=>Set(ref status,value);}
    public Profile? Profile {get;set;}
}
public sealed class GameCatalogViewModel(IServiceClient client):Observable {
    public event Action<Exception,string>? Failed;
    static string FailureText(Exception e)=>e is ServiceException {Message:"CLUB_ADMIN_MANAGED"}?"Настройками клуба управляет администратор":e is ServiceException {Message:"GAME_UDP_ANTICHEAT_COMPATIBILITY_UNVERIFIED"}?GameUdpMessages.Reason(e.Message):Text.Failure(e);
    readonly GameCard[] all=GameCatalog.Bundled().Games.Select(g=>new GameCard(g)).ToArray();
    public ObservableCollection<GameCard> Games {get;}=[];
    public Choice[] Modes {get;}=[new("AUTO","AUTO"),new("DIRECT","DIRECT"),new("DIRECT_PLUS","DIRECT+"),new("MANUAL","RELAY")];
    string search="",notice="Каталог не обещает снижение пинга. RELAY может его увеличить.";bool busy;
    public string Search {get=>search;set{if(Set(ref search,value))Filter();}}
    public string Notice {get=>notice;private set=>Set(ref notice,value);}
    public bool Busy {get=>busy;private set{if(Set(ref busy,value))Changed(nameof(CanInteract));}}
    public bool CanInteract=>!Busy;
    public void Filter(){Games.Clear();foreach(var g in all.Where(g=>g.Name.Contains(Search.Trim(),StringComparison.OrdinalIgnoreCase)||g.Launcher.Contains(Search.Trim(),StringComparison.OrdinalIgnoreCase)))Games.Add(g);}
    public async Task RefreshAsync(CancellationToken ct=default){
        if(Busy)return;Busy=true;
        try {
            var profiles=await client.Call<Profile[]>("profiles",cancellationToken:ct);
            string riotActivity="Лаунчер: диагностика пока недоступна";try{riotActivity=GameCatalog.RiotActivity(await client.Call<JsonElement>("riot.status",cancellationToken:ct));}catch(ServiceException){}
            GameUdpStatus? udp=null;try{udp=await client.Call<GameUdpStatus>("gameudp.status",cancellationToken:ct);}catch(ServiceException){}
            foreach(var g in all){g.Profile=profiles.FirstOrDefault(p=>p.Executables.Any(e=>g.Definition.Executables.Any(n=>ApplicationRow.Matches(n,e))));g.Mode=g.Profile?.Mode??g.Definition.DefaultMode;g.Status=g.Profile is null?"DIRECT · профиль появится при выборе режима":"Профиль: "+Text.Mode(g.Profile.Mode);var gameUdp=udp?.Profiles.FirstOrDefault(p=>p.ProfileId==g.Profile?.Id);if(gameUdp is not null){if(gameUdp.Engine.Enabled && g.Profile?.Mode is "DIRECT" or "AUTO")g.Mode="DIRECT_PLUS";g.Status+=" · "+gameUdp.Summary;}if(g.CanRepair)g.Status+=" · "+riotActivity;}
            Filter();Notice="Игры и лаунчеры имеют разные профили. Существующие соединения сохраняют маршрут.";
        }catch(Exception e){Failed?.Invoke(e,"gamecatalog.refresh");Notice="Каталог сохранён. Служба недоступна: "+FailureText(e);}finally{Busy=false;}
    }
    public async Task ApplyAsync(GameCard g,CancellationToken ct=default){
        if(Busy)return;Busy=true;
        try {
            if(!Modes.Any(m=>m.Value==g.Mode))throw new InvalidOperationException("Неизвестный режим");
            var p=g.Profile?.Copy()??new Profile{Id="game-"+g.Id,Name=g.Name,Executables=[..g.Definition.Executables],Custom=true};
            if(g.Mode=="DIRECT_PLUS"){
                if(g.Definition.UdpPreset.Length==0||p.Executables.Length!=1||!p.Executables[0].Contains('\\')){Notice="Для DIRECT+ нужен обнаруженный игровой UDP-профиль с полным путём. Совместимость и локальная стратегия должны быть проверены.";return;}
                if(p.Mode is not ("DIRECT" or "AUTO")){Notice="Сначала сохраните DIRECT или AUTO; текущий маршрут не изменён.";return;}
                await client.Call<GameUdpStatus>("gameudp.set",new{profile_id=p.Id,enabled=true},ct);
                g.Status="Игровой UDP-обход включён; маршрут приложения сохранён. Загрузка матча ещё не проверена.";return;
            }
            if(g.Mode=="MANUAL"){
                var status=await client.Call<Snapshot>("status",cancellationToken:ct);
                var relay=status.Routes.FirstOrDefault(r=>r.Id!="DIRECT"&&r.Status is "healthy" or "degraded");
                if(relay is null){Notice="Нет доступного релея. Профиль не изменён.";return;}p.RelayId=relay.Id;
            }else p.RelayId=null;
            p.Mode=g.Mode;
            await client.Call<Profile[]>("profile.set",p,ct);
            if(g.Mode=="DIRECT"&&g.Definition.UdpPreset.Length>0&&g.Profile is not null)await client.Call<GameUdpStatus>("gameudp.set",new{profile_id=p.Id,enabled=false},ct);
            g.Profile=p;g.Status="Сохранено: "+Text.Mode(p.Mode);Notice=p.Mode=="MANUAL"?"RELAY выбран явно. Для доступных напрямую серверов это может увеличить пинг.":"Маршрут сохранён для новых соединений; AUTO игр остаётся DIRECT.";
        }catch(Exception e){Failed?.Invoke(e,"gamecatalog.apply");Notice="Режим не подтверждён: "+FailureText(e);}finally{Busy=false;}
    }
    public async Task RepairLauncherAsync(GameCard g,CancellationToken ct=default){
        if(!g.CanRepair||Busy)return;Busy=true;
        try {var profiles=await client.Call<Profile[]>("profiles",cancellationToken:ct);var p=profiles.FirstOrDefault(p=>p.Id=="riot-client")?.Copy();if(p is null){Notice="Нужна обновлённая служба с профилем Riot Client.";return;}p.Mode="AUTO";p.RelayId=null;await client.Call<Profile[]>("profile.set",p,ct);g.Status="Riot Client: AUTO включён; DIRECT → рабочий DNS-адрес → LOCAL/RELAY. Проверка запуска ещё нужна.";Notice="Профиль игры и античит не изменены.";}catch(Exception e){Failed?.Invoke(e,"gamecatalog.launcher");Notice="Не удалось настроить лаунчер: "+FailureText(e);}finally{Busy=false;}
    }
}
