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
  var copy=p.Copy();copy.Name="Updated";var json=JsonSerializer.Serialize(copy,WireJson.Options);Assert.Contains("RELAY_REQUIRED",json);Assert.Contains("codex.exe",json);Assert.Contains("sensitive",json);Assert.NotSame(p.PolicyData,copy.PolicyData);
 }
}
