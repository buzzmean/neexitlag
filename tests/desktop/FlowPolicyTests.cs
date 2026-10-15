using System.Text.Json;
using Neexitlag.Desktop;
using Xunit;
namespace NeExitLag.Tests;
public class FlowPolicyTests
{
 [Fact] public void NullStrategyArraysAndActionsAreSafe()
 {
  var p=JsonSerializer.Deserialize<LocalStrategyProfile>("""{"ports":null,"domains":null,"ips":null,"exclude_domains":null,"exclude_ips":null,"actions":null}""",WireJson.Options)!;
  Assert.Empty(p.Ports);Assert.Empty(p.Domains);Assert.Empty(p.Ips);Assert.Empty(p.ExcludeDomains);Assert.Empty(p.ExcludeIps);Assert.NotNull(p.Advanced);
  p.Actions=[new LocalStrategyAction{Positions=null!}];Assert.NotNull(p.Advanced);
 }
 [Fact] public void GuiRoundtripPreservesRulesFallbackAndIdentity()
 {
  var p=JsonSerializer.Deserialize<Profile>("""{"id":"openai","name":"ChatGPT","executables":["C:\\Apps\\ChatGPT.exe"],"mode":"AUTO","fallback":"RELAY_REQUIRED","app_preset":"openai","identity":{"root":"C:\\Apps","executable":"chatgpt.exe","helpers":["codex.exe"]},"rules":[{"id":"sensitive","action":"RELAY_REQUIRED","relay_id":"VPS-01"}]}""",WireJson.Options)!;
  var copy=p.Copy();copy.Name="Updated";var json=JsonSerializer.Serialize(copy,WireJson.Options);Assert.Contains("RELAY_REQUIRED",json);Assert.Contains("codex.exe",json);Assert.Contains("sensitive",json);Assert.NotSame(p.PolicyData,copy.PolicyData);Assert.Equal("RELAY_REQUIRED",copy.Fallback);
 }
}
public class AutoStatusTests
{
 [Fact] public void NormalStatusUsesActualCountersAndHidesInternalIds()
 {
  var row=new ApplicationRow();var p=new Profile{Id="chrome",Executables=["C:\\chrome.exe"],Mode="AUTO"};var process=new ProcessIdentity{PID=7,CreationTime=10,Path=p.Executables[0]};
  var flow=new CaptureFlow{Pid=7,CreationTime=10,Exe=process.Path,SelectedRoute="RELAY",RelayId="VPS-01",Strategy="fake5",Chain="TTL=5",LastSeen=DateTimeOffset.UtcNow};var capture=new CaptureEvidence{Flows=[flow]};
  row.Update(p,[process],capture);Assert.Equal(Text.Get("AutoChecking"),row.UserStatus);
  flow.Counters.RelayedBytes=1;row.Update(p,[process],capture);Assert.Equal(Text.Get("ProtectedRoute"),row.UserStatus);Assert.DoesNotContain("VPS-01",row.UserStatus);Assert.DoesNotContain("fake5",row.UserStatus);
  flow.SelectedRoute="BLOCK";flow.RelayRequiredBlocked=true;row.Update(p,[process],capture);Assert.Equal(Text.Get("ProtectedFlowBlocked"),row.UserStatus);
 }
}
