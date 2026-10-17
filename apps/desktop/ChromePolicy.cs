using Microsoft.Win32;
using System.Text.Json;
namespace Neexitlag.Desktop;
// Current-user Chrome platform policy only. Never WinINET/WinHTTP/system proxy.
internal static class ChromePolicy {
 const string Key=@"Software\Policies\Google\Chrome";
 static string State=>Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData),"NeExitLag","chrome-proxy-policy.json");
 internal static void Install(string[] args){
  var pac=args.Single(p=>p.StartsWith("--proxy-pac-url=",StringComparison.Ordinal)).Substring("--proxy-pac-url=".Length);
  var json=JsonSerializer.Serialize(new{ProxyMode="pac_script",ProxyPacUrl=pac,ProxyBypassList="<-loopback>"});
  Check();
  using var current=Registry.CurrentUser.CreateSubKey(Key);
  var oldState=File.Exists(State)?File.ReadAllText(State):null;
  // Persist ownership before changing the registry. There is no preexisting
  // policy to overwrite: foreign settings are always refused, never copied.
  Directory.CreateDirectory(Path.GetDirectoryName(State)!);
  var pending=State+".tmp";File.WriteAllText(pending,JsonSerializer.Serialize(new Owned{Value=json}));File.Move(pending,State,true);
  try{current.SetValue("ProxySettings",json,RegistryValueKind.String);}
  catch{if(oldState is null)File.Delete(State);else File.WriteAllText(State,oldState);throw;}
 }
 internal static void Check(){
  using(var machine=Registry.LocalMachine.OpenSubKey(Key))if(new[]{"ProxySettings","ProxyMode","ProxyPacUrl","ProxyServer"}.Any(p=>machine?.GetValue(p)is not null))throw new InvalidOperationException(Text.Get("ChromeProxyManaged"));
  using var current=Registry.CurrentUser.OpenSubKey(Key);
  var present=current?.GetValue("ProxySettings");
  if(present is not null && present is not string)throw new InvalidOperationException(Text.Get("ChromeProxyManaged"));
  var previous=present as string;
  string? owned=null;
  if(File.Exists(State))owned=JsonSerializer.Deserialize<Owned>(File.ReadAllText(State))?.Value;
  if(previous is not null&&previous!=owned||new[]{"ProxyMode","ProxyPacUrl","ProxyServer"}.Any(p=>current?.GetValue(p)is not null))throw new InvalidOperationException(Text.Get("ChromeProxyManaged"));
 }
 internal static void Remove(){
  if(!File.Exists(State))return;
  var owned=JsonSerializer.Deserialize<Owned>(File.ReadAllText(State))?.Value;
  using var current=Registry.CurrentUser.OpenSubKey(Key,true);
  if(current?.GetValue("ProxySettings") is object actual&&(!Equals(actual,owned)))throw new InvalidOperationException(Text.Get("ChromeProxyManaged"));
  current?.DeleteValue("ProxySettings",false);File.Delete(State);
 }
 sealed class Owned {public string Value{get;set;}="";}
}
