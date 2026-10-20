using System.IO;
using Neexitlag.Desktop;
internal static class FreshPCSetupChecks
{
    internal static void Run()
    {
        static void Require(bool ok) { if (!ok) throw new Exception("Fresh-PC argument contract failed"); }
        static void Reject(Action action) { bool failed=false;try{action();}catch(ArgumentException){failed=true;}Require(failed); }
        var folder=Path.Combine(Path.GetTempPath(),"veliq-fresh-arguments-"+Guid.NewGuid());
        Directory.CreateDirectory(folder);
        try
        {
            var report=Path.Combine(folder,"report.json");
            Require(App.FreshPCSetupArguments([]) is null);
            Require(App.FreshPCSetupArguments(["--existing-normal-argument"]) is null);
            Require(App.FreshPCSetupArguments(["--fresh-pc-setup",report])==report);
            Reject(()=>App.FreshPCSetupArguments(["--fresh-pc-setup"]));
            Reject(()=>App.FreshPCSetupArguments(["--fresh-pc-setup","relative.json"]));
            File.WriteAllText(report,"owner file");
            Reject(()=>App.FreshPCSetupArguments(["--fresh-pc-setup",report]));
            Require(File.ReadAllText(report)=="owner file");
        }
        finally { Directory.Delete(folder,true); }
        Console.WriteLine("PASS fresh-PC explicit GUI entry and refusal of unsafe/foreign report path; ordinary GUI args unchanged");
    }
}
