using System.Collections.ObjectModel;
namespace Neexitlag.Desktop;

public sealed class GameUdpStatus
{
    public GameUdpProfile[] Profiles { get; set; } = [];
    public GameUdpGuard Guard { get; set; } = new();
    public string Error { get; set; } = "";
}
public sealed class GameUdpGuard { public string DriverAntiCheat {get;set;}=""; public string Error {get;set;}=""; public bool Conflict {get;set;} }
public sealed class GameUdpProfile
{
    public string ProfileId {get;set;}="";
    public string Name {get;set;}="";
    public string State {get;set;}="";
    public string Reason {get;set;}="";
    public string AntiCheat {get;set;}="";
    public GameUdpEngine Engine {get;set;}=new();
    public string ButtonLabel=>Engine.Enabled?"Выключить UDP-обход":"Включить UDP-обход";
    public string Summary {
        get {
            if (!Engine.Enabled) return "Игровой UDP-обход выключен. Обычный маршрут приложения сохранён.";
            if (Reason!="") return "UDP-обход приостановлен: "+GameUdpMessages.Reason(Reason);
            var f=Engine.Flows.OrderByDescending(f=>f.FlowId).FirstOrDefault();
            if(f is null)return "Готов для новых UDP-потоков: DIRECT → локальный обход → релей.";
            if(f.State=="TRANSPORT_FAILED")return "Игровой поток: "+GameUdpMessages.Reason(f.Reason);
            return f.Stage switch {
                "LOCAL_UDP" when f.FakePackets>0=>"Обход активен · "+f.Destination+" · приманок: "+f.FakePackets+" · ответов: "+f.InboundPackets,
                "LOCAL_UDP"=>"Приманка не отправлена · "+GameUdpMessages.Reason(f.Reason),
                "RELAY"=>"Игровой поток через релей · "+f.Destination+" · ответов: "+f.InboundPackets,
                _=>"DIRECT · "+f.Destination+" · ответов: "+f.InboundPackets
            };
        }
    }
}
public sealed class GameUdpEngine { public bool Enabled {get;set;} public GameUdpFlow[] Flows {get;set;}=[]; }
public sealed class GameUdpFlow {
    public ulong FlowId {get;set;} public string Stage {get;set;}="";public string Destination {get;set;}="";
    public ulong FakePackets {get;set;} public ulong FakeBytes {get;set;} public ulong OutgoingPackets {get;set;} public ulong InboundPackets {get;set;} public string State {get;set;}=""; public string Reason {get;set;}="";
}
public static class GameUdpMessages {
    public static string Reason(string code)=>code switch {
        "DISABLED_BY_USER"=>"выключен пользователем",
        "OPTIMIZATION_DISABLED"=>"оптимизация выключена",
        "GAME_RELAY_OPEN_FAILED"=>"не удалось открыть релей; DIRECT-подмены нет",
        "GAME_RELAY_TRANSPORT_FAILED"=>"соединение с релеем потеряно; DIRECT-подмены нет",
        "GAME_UDP_LADDER_EXHAUSTED"=>"варианты исчерпаны; ждём следующего окна подбора",
        "GAME_PROFILE_OVERRIDDEN_OR_REMOVED"=>"обычная политика профиля имеет приоритет",
        "UDP_CONFLICT_CHECK_PENDING"=>"проверяем другие сетевые перехватчики",
        "UDP_CONFLICT_DETECTOR_UNAVAILABLE"=>"не удалось безопасно проверить сетевые перехватчики",
        "GAME_UDP_CONFIG_INVALID"=>"настройки режима повреждены; приманки отключены",
        "ACTIVATION_REQUIRED"=>"нужна активация",
        "CAPTURE_NOT_READY"=>"сетевой перехват ещё не готов",
        "EXTERNAL_VPN_ACTIVE"=>"обнаружен внешний VPN",
        "FOREIGN_UDP_INTERCEPTOR" or "FOREIGN_UDP_INTERCEPTOR_OR_OBSERVER_PENDING"=>"работает другой сетевой перехватчик; двойная обработка отключена",
        "GAME_UDP_ANTICHEAT_COMPATIBILITY_UNVERIFIED" or "DRIVER_ANTICHEAT_COMPATIBILITY_UNVERIFIED"=>"совместимость с Vanguard/FACEIT не проверена; режим отключён",
        "NETWORK_FINGERPRINT_PENDING"=>"проверяем текущую сеть",
        _=>code.Length<=80 && code.All(c=>c is >='A' and <='Z' or >='0' and <='9' or '_')?code:"сбой службы; подробности в диагностике"
    };
}
public sealed partial class DesktopViewModel {
    public ObservableCollection<GameUdpProfile> GameUdpProfiles {get;}=[];
    string? lastGameUdp;
    public void ApplyGameUdp(GameUdpStatus status){var wire=System.Text.Json.JsonSerializer.Serialize(status);if(wire==lastGameUdp)return;lastGameUdp=wire;GameUdpProfiles.Clear();foreach(var p in status.Profiles.OrderBy(p=>p.Name,StringComparer.CurrentCultureIgnoreCase))GameUdpProfiles.Add(p);}
    public async Task RefreshGamesAsync(CancellationToken ct=default)=>ApplyGameUdp(await client.Call<GameUdpStatus>("gameudp.status",cancellationToken:ct));
    public async Task SetGameUdpAsync(string profileId,bool enabled,CancellationToken ct=default){
        try {ApplyGameUdp(await client.Call<GameUdpStatus>("gameudp.set",new{profile_id=profileId,enabled},ct));Notice=enabled?"Игровая лестница включена для новых UDP-потоков. Релей используется только после неудачных DIRECT и локальных вариантов. Проверка загрузки матча ещё нужна.":"Игровой UDP-обход выключен.";}
        catch(ServiceException e){ReportAuxiliaryFailure(e,"gameudp.set");Notice="Не удалось изменить игровой UDP-обход: "+GameUdpMessages.Reason(e.Message);}
    }
}
public sealed partial class DesktopViewModel {
    public void ReportGameUdpSelectionRequired()=>Notice="Выберите игровой профиль с полным путём .exe. Для готовых Steam-профилей достаточно кнопки в разделе «Игровой UDP-обход».";
}
