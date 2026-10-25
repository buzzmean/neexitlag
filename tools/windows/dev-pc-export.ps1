# Один раз на ПК, где уже есть рабочий SSH-ключ к VPS: собирает «набор рабочего места»
# в один зашифрованный паролем файл veliq-dev.bin (SSH-ключ, known_hosts, публичный CA).
# Права администратора не нужны. Файл можно хранить в облаке/чате проекта: без пароля он
# бесполезен. Пароль никуда не пересылать. dev-pc.ps1 на новом ПК находит файл в
# «Загрузках» и спрашивает пароль.
#   powershell -NoProfile -ExecutionPolicy Bypass -File C:\veliq\neexitlag\tools\windows\dev-pc-export.ps1
[CmdletBinding()]
param(
 [string]$OutputFile='',
 [string]$SshDirectory='',
 [string]$CaFile=''
)
$ErrorActionPreference='Stop'
Set-StrictMode -Version Latest
# Публичный CA VELIQ (не секрет): SHA256 по DER. Совпадает с пином в dev-pc.ps1.
$script:DevPcCaPin='622db56fba69041b108fc31b5b3cf81a5ed722faa3110ef2b896b8fac42abd85'

function Test-DevPcCA([string]$Path) {
 if(!(Test-Path -LiteralPath $Path -PathType Leaf)){return $false}
 $text=[IO.File]::ReadAllText($Path)
 if($text -match 'PRIVATE KEY' -or ([regex]::Matches($text,'-----BEGIN CERTIFICATE-----')).Count -ne 1){return $false}
 try{$cert=New-Object Security.Cryptography.X509Certificates.X509Certificate2 $Path}catch{return $false}
 $sha=[Security.Cryptography.SHA256]::Create()
 try{$hash=([BitConverter]::ToString($sha.ComputeHash($cert.RawData))).Replace('-','').ToLowerInvariant()}finally{$sha.Dispose();$cert.Dispose()}
 return $hash -ceq $script:DevPcCaPin
}

# Формат VLQD1: magic | salt(16) | iv(16) | AES-256-CBC | HMAC-SHA256(32) по всему
# предыдущему. PBKDF2-SHA256 600000 раундов -> 32 байта ключа шифрования + 32 байта MAC.
# Обратная функция — Unprotect-DevPcKit в dev-pc.ps1.
function Protect-DevPcKit([string]$Source,[string]$Destination,[string]$Passphrase) {
 $salt=New-Object byte[] 16;$rng=[Security.Cryptography.RandomNumberGenerator]::Create();$rng.GetBytes($salt);$rng.Dispose()
 $kdf=New-Object Security.Cryptography.Rfc2898DeriveBytes ($Passphrase,$salt,600000,[Security.Cryptography.HashAlgorithmName]::SHA256)
 $keys=$kdf.GetBytes(64);$kdf.Dispose()
 $aes=[Security.Cryptography.Aes]::Create();$aes.Key=[byte[]]$keys[0..31];$aes.GenerateIV()
 $out=[IO.File]::Open($Destination,[IO.FileMode]::CreateNew)
 try{
  foreach($part in @([Text.Encoding]::ASCII.GetBytes('VLQD1'),$salt,$aes.IV)){$out.Write($part,0,$part.Length)}
  $crypto=New-Object Security.Cryptography.CryptoStream ($out,$aes.CreateEncryptor(),[Security.Cryptography.CryptoStreamMode]::Write)
  $in=[IO.File]::OpenRead($Source);try{$in.CopyTo($crypto)}finally{$in.Dispose()}
  $crypto.FlushFinalBlock();$crypto.Dispose()
 }catch{$out.Dispose();Remove-Item -LiteralPath $Destination -Force -ErrorAction SilentlyContinue;throw}
 finally{$out.Dispose();$aes.Dispose()}
 $mac=New-Object Security.Cryptography.HMACSHA256 (,[byte[]]$keys[32..63])
 $in=[IO.File]::OpenRead($Destination);try{$tag=$mac.ComputeHash($in)}finally{$in.Dispose();$mac.Dispose()}
 $out=[IO.File]::Open($Destination,[IO.FileMode]::Append);try{$out.Write($tag,0,$tag.Length)}finally{$out.Dispose()}
}

function Read-DevPcNewPassphrase {
 $first=Read-Host 'Придумайте пароль для файла (от 12 символов)' -AsSecureString
 $second=Read-Host 'Повторите пароль' -AsSecureString
 $a=[Runtime.InteropServices.Marshal]::PtrToStringBSTR([Runtime.InteropServices.Marshal]::SecureStringToBSTR($first))
 $b=[Runtime.InteropServices.Marshal]::PtrToStringBSTR([Runtime.InteropServices.Marshal]::SecureStringToBSTR($second))
 if($a -cne $b){throw 'Пароли не совпали'}
 if($a.Length -lt 12){throw 'Пароль короче 12 символов'}
 return $a
}

function Export-DevPcKit {
 # Значения по умолчанию здесь, а не в param(): иначе dot-source ломается без папки «Рабочий стол».
 if(!$OutputFile){$OutputFile=Join-Path ([Environment]::GetFolderPath('Desktop')) 'veliq-dev.bin'}
 if(!$SshDirectory){$SshDirectory=Join-Path $env:USERPROFILE '.ssh'}
 $OutputFile=[IO.Path]::GetFullPath($OutputFile)
 if(Test-Path -LiteralPath $OutputFile){throw "Файл $OutputFile уже есть; удалите его сами или укажите -OutputFile"}
 $owner=[Security.Principal.WindowsIdentity]::GetCurrent().User.Value
 $stage=Join-Path ([IO.Path]::GetTempPath()) ('veliq-dev-'+[Guid]::NewGuid().ToString('N'))
 $kit=Join-Path $stage 'kit'
 [void][IO.Directory]::CreateDirectory($kit)
 & icacls $stage /inheritance:r /grant:r "*$($owner):(OI)(CI)F" '*S-1-5-18:(OI)(CI)F' '*S-1-5-32-544:(OI)(CI)F'|Out-Null
 if($LASTEXITCODE){Remove-Item -LiteralPath $stage -Recurse -Force;throw 'Не удалось закрыть временную папку'}
 try{
  $files=[ordered]@{}
  function Add-KitFile([string]$Source,[string]$Relative){
   if((Get-Item -LiteralPath $Source -Force).Attributes -band [IO.FileAttributes]::ReparsePoint){throw "Ссылки не берём: $Source"}
   $target=Join-Path $kit $Relative;[void][IO.Directory]::CreateDirectory((Split-Path $target))
   Copy-Item -LiteralPath $Source -Destination $target
   $files[$Relative]=(Get-FileHash -LiteralPath $target -Algorithm SHA256).Hash
  }
  '== SSH'
  $keys=@(foreach($name in @('id_ed25519','id_ecdsa','id_rsa')){if(Test-Path -LiteralPath (Join-Path $SshDirectory $name)){$name}})
  if(!$keys.Count){throw "В $SshDirectory нет SSH-ключа (id_ed25519/id_ecdsa/id_rsa)"}
  foreach($name in $keys){
   Add-KitFile (Join-Path $SshDirectory $name) "ssh\$name"
   if(Test-Path -LiteralPath (Join-Path $SshDirectory "$name.pub")){Add-KitFile (Join-Path $SshDirectory "$name.pub") "ssh\$name.pub"}
  }
  foreach($name in @('known_hosts','config')){if(Test-Path -LiteralPath (Join-Path $SshDirectory $name)){Add-KitFile (Join-Path $SshDirectory $name) "ssh\$name"}}
  "ключи: $($keys -join ', ') (содержимое не выводится)"
  '== Публичный CA'
  $root=(Resolve-Path (Join-Path $PSScriptRoot '..\..')).Path
  $candidates=@($CaFile,(Join-Path $root '.e2e\activation-trust\ca.pem'),'C:\veliq\enroll\ca.pem',(Join-Path $env:ProgramData 'neexitlag\cp-ca.pem'),(Join-Path $env:ProgramData 'neexitlag\ca.pem'))
  $ca=$null
  foreach($candidate in $candidates){if($candidate -and (Test-DevPcCA $candidate)){$ca=$candidate;break}}
  if($ca){Add-KitFile $ca 'trust\ca.pem';"CA: $ca (отпечаток совпал)"}else{Write-Warning 'Публичный CA с нужным отпечатком не найден; набор будет без него'}
  $info=[ordered]@{format=1;created_utc=[DateTime]::UtcNow.ToString('o');computer=$env:COMPUTERNAME;files=$files}
  [IO.File]::WriteAllText((Join-Path $kit 'kit.json'),($info|ConvertTo-Json -Depth 4),(New-Object Text.UTF8Encoding $false))
  '== Шифрование'
  $passphrase=Read-DevPcNewPassphrase
  Add-Type -AssemblyName System.IO.Compression.FileSystem
  [IO.Compression.ZipFile]::CreateFromDirectory($kit,(Join-Path $stage 'kit.zip'))
  Protect-DevPcKit (Join-Path $stage 'kit.zip') $OutputFile $passphrase
 }finally{Remove-Item -LiteralPath $stage -Recurse -Force -ErrorAction SilentlyContinue}
 Write-Host "Готово: $OutputFile. Перетащите этот файл в чат проекта VELIQ; пароль в чат не пишите." -ForegroundColor Green
}

if($MyInvocation.InvocationName -eq '.'){return}
Export-DevPcKit
