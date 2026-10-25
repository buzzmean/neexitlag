# Self-contained bootstrap for PCs that roll back on reboot: decrypts the kit
# made by export-kit.ps1 and runs its apply-kit.ps1. Holds no secrets. Put it
# next to veliq-kit.bin and run elevated, as the club user who owns VELIQ:
#   Set-ExecutionPolicy -Scope Process Bypass -Force
#   Unblock-File "$HOME\Downloads\restore-kit.ps1"; & "$HOME\Downloads\restore-kit.ps1"
[CmdletBinding()]
param(
 [string]$KitFile=(Join-Path $PSScriptRoot 'veliq-kit.bin'),
 [switch]$NoLaunch
)
$ErrorActionPreference='Stop'
Set-StrictMode -Version Latest
$principal=[Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()
if(!$principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'Run from an elevated Windows PowerShell.'}
$owner=[Security.Principal.WindowsIdentity]::GetCurrent().User.Value
$KitFile=(Resolve-Path -LiteralPath $KitFile).Path
# Inverse of Protect-KitArchive in export-kit.ps1; the MAC is checked before
# any byte is decrypted.
function Unprotect-KitArchive([string]$Source,[string]$Destination,[string]$Passphrase) {
 $in=[IO.File]::OpenRead($Source)
 try{
  $header=New-Object byte[] 37
  if($in.Length -lt 37+16+32 -or $in.Read($header,0,37) -ne 37 -or [Text.Encoding]::ASCII.GetString($header,0,5) -ne 'VLQK1'){throw 'Not a VELIQ kit file'}
  $kdf=[Security.Cryptography.Rfc2898DeriveBytes]::new($Passphrase,[byte[]]$header[5..20],600000,[Security.Cryptography.HashAlgorithmName]::SHA256)
  $keys=$kdf.GetBytes(64);$kdf.Dispose()
  $body=$in.Length-32
  $mac=[Security.Cryptography.HMACSHA256]::new([byte[]]$keys[32..63])
  [void]$mac.TransformBlock($header,0,37,$null,0)
  $buffer=New-Object byte[] 1048576;$left=$body-37
  while($left -gt 0){$n=$in.Read($buffer,0,[int][Math]::Min($buffer.Length,$left));if($n -le 0){throw 'Kit file truncated'};[void]$mac.TransformBlock($buffer,0,$n,$null,0);$left-=$n}
  [void]$mac.TransformFinalBlock($buffer,0,0)
  $tag=New-Object byte[] 32;[void]$in.Read($tag,0,32)
  $diff=0;for($i=0;$i -lt 32;$i++){$diff=$diff -bor ($tag[$i] -bxor $mac.Hash[$i])}
  $mac.Dispose()
  if($diff){throw 'Wrong passphrase or damaged kit file'}
  $in.Position=37
  $aes=[Security.Cryptography.Aes]::Create();$aes.Key=[byte[]]$keys[0..31];$aes.IV=[byte[]]$header[21..36]
  $out=[IO.File]::Open($Destination,[IO.FileMode]::CreateNew)
  try{
   $crypto=[Security.Cryptography.CryptoStream]::new($out,$aes.CreateDecryptor(),[Security.Cryptography.CryptoStreamMode]::Write)
   $left=$body-37
   while($left -gt 0){$n=$in.Read($buffer,0,[int][Math]::Min($buffer.Length,$left));$crypto.Write($buffer,0,$n);$left-=$n}
   $crypto.FlushFinalBlock();$crypto.Dispose()
  }finally{$out.Dispose();$aes.Dispose()}
 }finally{$in.Dispose()}
}
$secure=Read-Host 'Kit passphrase' -AsSecureString
$passphrase=[Runtime.InteropServices.Marshal]::PtrToStringBSTR([Runtime.InteropServices.Marshal]::SecureStringToBSTR($secure))
# Plaintext credentials exist only in an admin/owner-only directory and are
# deleted once the service has its own copies.
$stage=Join-Path $env:ProgramData ('veliq-kit-'+[Guid]::NewGuid().ToString('N'))
[void][IO.Directory]::CreateDirectory($stage)
try{
 & icacls $stage /inheritance:r /grant:r '*S-1-5-18:(OI)(CI)F' '*S-1-5-32-544:(OI)(CI)F' "*$($owner):(OI)(CI)F"|Out-Null
 if($LASTEXITCODE){throw 'Failed to protect kit staging directory'}
 '===== DECRYPT ====='
 Unprotect-KitArchive $KitFile (Join-Path $stage 'kit.zip') $passphrase
 Expand-Archive -LiteralPath (Join-Path $stage 'kit.zip') -DestinationPath (Join-Path $stage 'kit')
 Remove-Item -LiteralPath (Join-Path $stage 'kit.zip') -Force
 & (Join-Path $stage 'kit\tools\windows\apply-kit.ps1') -KitDirectory (Join-Path $stage 'kit') -NoLaunch:$NoLaunch
}finally{Remove-Item -LiteralPath $stage -Recurse -Force -ErrorAction SilentlyContinue}
