# Read-only Windows policy/token inspection. No elevation, token replacement or
# registry writes. Keep the decision pure so unsafe combinations can be tested.
function Resolve-FreshExecutionMode($Evidence) {
 if($Evidence.Restricted){throw 'RESTRICTED_TOKEN_REFUSED: do not use runas /trustlevel'}
 if($Evidence.EnableLUA -notin @(0,1) -or $Evidence.ConsentPromptBehaviorAdmin -notin @(0,1,2,3,4,5) -or $Evidence.ElevationType -notin @(1,2,3)){
  throw 'UAC_OR_TOKEN_EVIDENCE_INVALID'
 }
 if($Evidence.EnableLUA -eq 0 -and $Evidence.IsAdministrator -and $Evidence.Elevated -and $Evidence.ElevationType -eq 1 -and !$Evidence.HasLinkedToken){
  return 'UAC_DISABLED_FULL_ADMIN'
 }
 if($Evidence.EnableLUA -eq 1 -and !$Evidence.IsAdministrator -and !$Evidence.Elevated -and (($Evidence.ElevationType -eq 3 -and $Evidence.HasLinkedToken) -or ($Evidence.ElevationType -eq 1 -and !$Evidence.HasLinkedToken))){
  return 'UAC_ENABLED_OWNER'
 }
 throw 'ELEVATED_OR_INCONSISTENT_TOKEN_REFUSED: use ordinary owner PowerShell with UAC enabled; full-admin mode requires EnableLUA=0, default token and no linked token'
}

function Get-FreshExecutionEvidence {
 if($env:OS -ne 'Windows_NT'){throw 'WINDOWS_TOKEN_QUERY_REQUIRED'}
 if($null -eq ('Veliq.FreshPCToken' -as [type])){
  Add-Type -TypeDefinition @'
using System;
using System.ComponentModel;
using System.Runtime.InteropServices;
namespace Veliq {
 public sealed class FreshPCToken {
  public int ElevationType { get; private set; }
  public bool Elevated { get; private set; }
  public bool HasLinkedToken { get; private set; }
  public bool Restricted { get; private set; }
  [DllImport("kernel32.dll")] static extern IntPtr GetCurrentProcess();
  [DllImport("kernel32.dll", SetLastError=true)] static extern bool CloseHandle(IntPtr handle);
  [DllImport("advapi32.dll", SetLastError=true)] static extern bool OpenProcessToken(IntPtr process, uint access, out IntPtr token);
  [DllImport("advapi32.dll", SetLastError=true)] static extern bool GetTokenInformation(IntPtr token, int info, IntPtr buffer, int size, out int needed);
  [DllImport("advapi32.dll")] static extern bool IsTokenRestricted(IntPtr token);
  public static FreshPCToken Read() {
   IntPtr token;
   // TOKEN_QUERY only: no SeDebugPrivilege, AdjustTokenPrivileges or linked-token impersonation.
   if(!OpenProcessToken(GetCurrentProcess(), 8, out token)) throw new Win32Exception(Marshal.GetLastWin32Error());
   IntPtr buffer=Marshal.AllocHGlobal(IntPtr.Size);
   try {
    int needed;
    if(!GetTokenInformation(token, 18, buffer, IntPtr.Size, out needed)) throw new Win32Exception(Marshal.GetLastWin32Error());
    int type=Marshal.ReadInt32(buffer);
    if(!GetTokenInformation(token, 20, buffer, IntPtr.Size, out needed)) throw new Win32Exception(Marshal.GetLastWin32Error());
    bool elevated=Marshal.ReadInt32(buffer)!=0;
    bool linked=false;
    if(GetTokenInformation(token, 19, buffer, IntPtr.Size, out needed)) {
     IntPtr other=Marshal.ReadIntPtr(buffer);
     linked=other!=IntPtr.Zero;
     if(linked) CloseHandle(other);
    } else {
     int error=Marshal.GetLastWin32Error();
     // A default (non-split) token has no linked logon session. Any other
     // query failure is UNKNOWN, never evidence that UAC is disabled.
     if(error!=1312) throw new Win32Exception(error);
    }
    return new FreshPCToken { ElevationType=type, Elevated=elevated, HasLinkedToken=linked, Restricted=IsTokenRestricted(token) };
   } finally { Marshal.FreeHGlobal(buffer); CloseHandle(token); }
  }
 }
}
'@
 }
 $base=[Microsoft.Win32.RegistryKey]::OpenBaseKey([Microsoft.Win32.RegistryHive]::LocalMachine,[Microsoft.Win32.RegistryView]::Registry64)
 try{
  $key=$base.OpenSubKey('SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System',$false)
  if($null -eq $key){throw 'UAC_POLICY_MISSING'}
  try{
   foreach($name in @('EnableLUA','ConsentPromptBehaviorAdmin')){
    if($key.GetValueKind($name) -ne [Microsoft.Win32.RegistryValueKind]::DWord){throw 'UAC_POLICY_TYPE_INVALID'}
   }
   $lua=[int]$key.GetValue('EnableLUA');$consent=[int]$key.GetValue('ConsentPromptBehaviorAdmin')
  }finally{$key.Dispose()}
 }finally{$base.Dispose()}
 $token=[Veliq.FreshPCToken]::Read()
 $identity=[Security.Principal.WindowsIdentity]::GetCurrent()
 try{$admin=([Security.Principal.WindowsPrincipal]::new($identity)).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)}finally{$identity.Dispose()}
 return [pscustomobject]@{EnableLUA=$lua;ConsentPromptBehaviorAdmin=$consent;IsAdministrator=$admin;Elevated=$token.Elevated;ElevationType=$token.ElevationType;HasLinkedToken=$token.HasLinkedToken;Restricted=$token.Restricted}
}

function Invoke-FreshServicePhase([string]$Mode,[string]$Sid,[scriptblock]$InProcess,[scriptblock]$Elevate) {
 switch -Exact ($Mode) {
  'UAC_DISABLED_FULL_ADMIN' { & $InProcess $Sid; return }
  'UAC_ENABLED_OWNER' { & $Elevate $Sid; return }
  default { throw 'UNVERIFIED_SERVICE_EXECUTION_MODE' }
 }
}
