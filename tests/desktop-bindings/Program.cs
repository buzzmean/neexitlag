using Neexitlag.Desktop;
var client=new ServiceClient();
var status=await client.Call<Snapshot>("status");
if(!status.connected||!status.demo||status.routes.Length<1)throw new Exception("Demo service state unavailable");
var profiles=await client.Call<Profile[]>("profiles");
if(profiles.Length<3)throw new Exception("Application profiles unavailable");
var diagnostics=await client.Call<Check[]>("diagnostics");
if(diagnostics.Length<4)throw new Exception("Diagnostics unavailable");
Console.WriteLine($"Desktop binding smoke PASS: {status.routes.Length} measured routes, {profiles.Length} profiles, {diagnostics.Length} diagnostics. Primary={status.Primary}");
