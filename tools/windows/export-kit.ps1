# Captures a working installation into one passphrase-encrypted kit file for PCs
# that roll back on reboot (diskless club images). Run once, elevated, as the
# installation owner, with VELIQ exited from the tray. The kit holds this
# device's credentials: it may travel through a cloud chat only encrypted, the
# passphrase never travels with it, and it must never be used on another PC.
[CmdletBinding()]
param(
 [string]$OutputFile=(Join-Path ([Environment]::GetFolderPath('Desktop')) 'veliq-kit.bin'),
 [string]$InstallDirectory="$env:ProgramData\neexitlag",
 [string]$GuiDirectory=(Join-Path $env:LOCALAPPDATA 'Programs\VELIQ'),
 [string]$SshDirectory=(Join-Path $env:USERPROFILE '.ssh'),
 [switch]$NoState,
 [switch]$NoSsh
)
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'common.ps1')
Assert-Administrator
function Read-KitPassphrase {
 $first=Read-Host 'Kit passphrase (12+ characters)' -AsSecureString
 $second=Read-Host 'Repeat passphrase' -AsSecureString
 $a=[Runtime.InteropServices.Marshal]::PtrToStringBSTR([Runtime.InteropServices.Marshal]::SecureStringToBSTR($first))
 $b=[Runtime.InteropServices.Marshal]::PtrToStringBSTR([Runtime.InteropServices.Marshal]::SecureStringToBSTR($second))
 if($a -cne $b){throw 'Passphrases differ'}
 if($a.Length -lt 12){throw 'Passphrase must have at least 12 characters'}
 return $a
}
# Format VLQK1: magic | salt(16) | iv(16) | AES-256-CBC ciphertext | HMAC-SHA256(32)
# over everything before it. PBKDF2-SHA256 600000 rounds gives 32 bytes of
# encryption key and 32 bytes of MAC key. Must match restore-kit.ps1.
function Protect-KitArchive([string]$Source,[string]$Destination,[string]$Passphrase) {
 $salt=New-Object byte[] 16;$rng=[Security.Cryptography.RandomNumberGenerator]::Create();$rng.GetBytes($salt)
 $kdf=[Security.Cryptography.Rfc2898DeriveBytes]::new($Passphrase,$salt,600000,[Security.Cryptography.HashAlgorithmName]::SHA256)
 $keys=$kdf.GetBytes(64);$kdf.Dispose();$rng.Dispose()
 $aes=[Security.Cryptography.Aes]::Create();$aes.Key=[byte[]]$keys[0..31];$aes.GenerateIV()
 $out=[IO.File]::Open($Destination,[IO.FileMode]::CreateNew)
 try{
  $magic=[Text.Encoding]::ASCII.GetBytes('VLQK1')
  foreach($part in @($magic,$salt,$aes.IV)){$out.Write($part,0,$part.Length)}
  $crypto=[Security.Cryptography.CryptoStream]::new($out,$aes.CreateEncryptor(),[Security.Cryptography.CryptoStreamMode]::Write)
  $in=[IO.File]::OpenRead($Source);try{$in.CopyTo($crypto)}finally{$in.Dispose()}
  $crypto.FlushFinalBlock();$crypto.Dispose()
 }catch{$out.Dispose();Remove-Item -LiteralPath $Destination -Force -ErrorAction SilentlyContinue;throw}
 finally{$out.Dispose();$aes.Dispose()}
 $mac=[Security.Cryptography.HMACSHA256]::new([byte[]]$keys[32..63])
 $in=[IO.File]::OpenRead($Destination);try{$tag=$mac.ComputeHash($in)}finally{$in.Dispose();$mac.Dispose()}
 $out=[IO.File]::Open($Destination,[IO.FileMode]::Append);try{$out.Write($tag,0,$tag.Length)}finally{$out.Dispose()}
}
$owner=[Security.Principal.WindowsIdentity]::GetCurrent().User.Value
$manifest=Get-Content -Raw -LiteralPath (Join-Path $InstallDirectory 'service.json')|ConvertFrom-Json
if([string]$manifest.owner_sid -ne $owner){throw 'Run as the installation owner: service.json owner_sid differs from the current user'}
if((Get-Service neexitlag -ErrorAction Stop).Status -ne 'Running'){throw 'Service is not running; export only a verified working installation'}
$OutputFile=[IO.Path]::GetFullPath($OutputFile)
if(Test-Path -LiteralPath $OutputFile){throw 'Output file exists; refusing to overwrite an existing kit'}
$passphrase=Read-KitPassphrase
# Plaintext staging lives only in an admin/owner-only directory and is deleted.
$stage=Join-Path $env:ProgramData ('veliq-kit-'+[Guid]::NewGuid().ToString('N'))
$kit=Join-Path $stage 'kit'
[void][IO.Directory]::CreateDirectory($kit)
& icacls $stage /inheritance:r /grant:r '*S-1-5-18:(OI)(CI)F' '*S-1-5-32-544:(OI)(CI)F' "*$($owner):(OI)(CI)F"|Out-Null
if($LASTEXITCODE){Remove-Item -LiteralPath $stage -Recurse -Force;throw 'Failed to protect kit staging directory'}
try{
function Copy-KitFile([string]$Source,[string]$Relative){
 $destination=Join-Path $kit $Relative
 [void][IO.Directory]::CreateDirectory((Split-Path $destination))
 if((Get-Item -LiteralPath $Source -Force).Attributes -band [IO.FileAttributes]::ReparsePoint){throw 'Kit export refuses reparse points'}
 Copy-Item -LiteralPath $Source -Destination $destination
}
'===== BINARIES ====='
Copy-KitFile (Join-Path $InstallDirectory 'neexitlag-service.exe') 'bin\neexitlag-service.exe'
Copy-KitFile (Join-Path $GuiDirectory 'VELIQ.exe') 'bin\VELIQ.exe'
# The GUI runs AUTO through scripts shipped beside it; install-desktop.ps1 takes
# them from bin\tools\windows.
$guiScripts=Join-Path $GuiDirectory 'tools\windows'
if(!(Test-Path -LiteralPath $guiScripts)){$guiScripts=$PSScriptRoot}
foreach($file in Get-ChildItem -LiteralPath $guiScripts -Filter '*.ps1' -File){Copy-KitFile $file.FullName "bin\tools\windows\$($file.Name)"}
foreach($name in @('common.ps1','reinstall-state.ps1','install.ps1','install-desktop.ps1','apply-kit.ps1')){Copy-KitFile (Join-Path $PSScriptRoot $name) "tools\windows\$name"}
'===== DRIVER ====='
foreach($name in @('WinDivert.dll','WinDivert64.sys')){Copy-KitFile (Join-Path $manifest.driver.directory $name) "driver\$name"}
'===== CREDENTIALS ====='
# Copied, never printed. ipc.token is not exported: a fresh install issues a new one.
Copy-KitFile $manifest.device_identity_file 'secrets\device\identity.json'
Copy-KitFile $manifest.device_certificate 'secrets\device\device.pem'
Copy-KitFile $manifest.trust_file 'secrets\trust\trust.json'
Copy-KitFile $manifest.ca_file 'secrets\trust\ca.pem'
if($null -ne $manifest.PSObject.Properties['control_plane_ca_file']){Copy-KitFile $manifest.control_plane_ca_file 'secrets\trust\cp-ca.pem'}
if($null -ne $manifest.PSObject.Properties['probe_token_file']){Copy-KitFile $manifest.probe_token_file 'secrets\probe.token'}
if(!$NoSsh){
 '===== SSH ====='
 # The owner's VPS key travels only inside the encrypted kit.
 foreach($name in @('id_ed25519','id_ed25519.pub','known_hosts')){
  $path=Join-Path $SshDirectory $name
  if(Test-Path -LiteralPath $path){Copy-KitFile $path "secrets\ssh\$name"}
 }
}
if(!$NoState){
 '===== STATE ====='
 # Profiles, settings and learned strategies. Not carried over: the browser
 # guard, the signed config cache, and browser-transport.json, whose ready
 # Chrome profiles would skip re-activation while the HKCU policy is gone.
 foreach($name in @('profiles.json','settings.json','local-strategies.json','auto-learning.json')){
  $path=Join-Path $manifest.state_directory $name
  if(Test-Path -LiteralPath $path){
   try{[void](Get-Content -Raw -LiteralPath $path|ConvertFrom-Json)}catch{throw "State file is not valid JSON: $name"}
   Copy-KitFile $path "state\$name"
  }
 }
 $prefs=Join-Path $env:LOCALAPPDATA 'NeExitLag\desktop.json'
 if(Test-Path -LiteralPath $prefs){Copy-KitFile $prefs 'desktop\desktop.json'}
}
'===== MANIFEST ====='
$hashes=[ordered]@{}
foreach($file in Get-ChildItem -LiteralPath $kit -Recurse -File -Force|Sort-Object FullName){$hashes[$file.FullName.Substring($kit.Length+1)]=(Get-FileHash -LiteralPath $file.FullName -Algorithm SHA256).Hash}
$info=[ordered]@{format=1;created_utc=[DateTime]::UtcNow.ToString('o');computer=$env:COMPUTERNAME;owner_sid=$owner;control_plane_url=$manifest.control_plane_url;probe_target=$manifest.probe_target;tcp_probe_target=[string]$manifest.tcp_probe_target;dll_sha256=$manifest.driver.dll_sha256;driver_sha256=$manifest.driver.driver_sha256;files=$hashes}
[IO.File]::WriteAllText((Join-Path $kit 'kit.json'),($info|ConvertTo-Json -Depth 4),[Text.UTF8Encoding]::new($false))
'===== ENCRYPT ====='
Compress-Archive -Path (Join-Path $kit '*') -DestinationPath (Join-Path $stage 'kit.zip')
Protect-KitArchive (Join-Path $stage 'kit.zip') $OutputFile $passphrase
}finally{Remove-Item -LiteralPath $stage -Recurse -Force}
# The bootstrap holds no secrets; send it next to the kit.
Copy-Item -LiteralPath (Join-Path $PSScriptRoot 'restore-kit.ps1') -Destination (Join-Path (Split-Path $OutputFile) 'restore-kit.ps1') -Force
Write-Host "Kit written: $OutputFile ($($hashes.Count) files) and restore-kit.ps1 next to it. Keep the passphrase outside the chat."
