namespace Neexitlag.Desktop;
public sealed class ActivationView {
 public string State {get;set;}="";public string Error {get;set;}="";public bool HasIdentity {get;set;}
 public string CaNotAfter {get;set;}="";public string ServerNotAfter {get;set;}="";
 public string Message=>Error switch {
  "ACTIVATION_INVALID"=>"Неверный код активации. Проверьте код и попробуйте ещё раз.",
  "ACTIVATION_REVOKED" or "ACTIVATION_EXPIRED"=>"Код больше не действует. Обратитесь к администратору или получите новый код.",
  "ACTIVATION_LIMIT"=>"Достигнут лимит компьютеров для этого кода.",
  "MACHINE_DISABLED"=>"Этот компьютер отключён. Обратитесь к администратору.",
  "ACTIVATION_RATE_LIMIT"=>"Слишком много попыток. Подождите немного и повторите.",
  "ACTIVATION_TRUST_NOT_CONFIGURED"=>"Установка не готова к активации. Обновите VELIQ.",
  "HARDWARE_ID_UNAVAILABLE"=>"Не удалось определить компьютер. Проверьте сетевое подключение и повторите.",
  ""=>State=="ACTIVATED"?"Код активации: активирован":State=="ACTIVATING"?"Активируем VELIQ…":"Введите код активации",
  _=>"Нет связи с VELIQ. Проверьте подключение и повторите."
 };
 public string CertificateStatus=>"Срок доверия: "+Date(CaNotAfter)+" · Сертификат подключения: "+Date(ServerNotAfter);
 static string Date(string s)=>DateTimeOffset.TryParse(s,out var d)?d.ToLocalTime().ToString("dd.MM.yyyy")+(d<DateTimeOffset.UtcNow.AddDays(10)?" — требуется продление":""):"нет данных";
}
public sealed class ActivationModel(IServiceClient client){
 public ActivationView Status {get;private set;}=new();
 public async Task RefreshAsync(CancellationToken ct=default){try{Status=await client.Call<ActivationView>("activation.status",cancellationToken:ct);}catch(Exception)when(!ct.IsCancellationRequested){Status=new(){State="UNAVAILABLE",Error="ACTIVATION_UNAVAILABLE"};}}
 public async Task ActivateAsync(string code,CancellationToken ct=default){try{Status=await client.Call<ActivationView>("activation.set",new{code},ct);}catch(Exception e)when(!ct.IsCancellationRequested){Status=new(){State="ACTIVATION_ERROR",Error=KnownError(e.Message)};}}
 static string KnownError(string s){string[] codes=["ACTIVATION_INVALID","ACTIVATION_REVOKED","ACTIVATION_EXPIRED","ACTIVATION_LIMIT","MACHINE_DISABLED","ACTIVATION_RATE_LIMIT","ACTIVATION_TRUST_NOT_CONFIGURED","HARDWARE_ID_UNAVAILABLE"];return codes.FirstOrDefault(c=>s==c||s.EndsWith(": "+c,StringComparison.Ordinal))??"ACTIVATION_UNAVAILABLE";}
}
public sealed partial class DesktopViewModel {
 public ActivationModel? Activation {get;private set;}
 public ActivationModel ActivationClient()=>Activation??=new(client);
}
