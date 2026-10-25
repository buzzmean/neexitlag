# Рабочее место VELIQ на любом ПК (в т.ч. бездисковом Langame) одной командой.
# PowerShell от администратора, под тем пользователем, который сидит за ПК:
#   & C:\veliq\neexitlag\tools\windows\dev-pc.ps1
# Ставит Git/Go/.NET 8/Python, собирает службу и GUI, ставит/обновляет службу VELIQ
# (активация кодом в окне VELIQ, сервер Hostkey), ставит Claude Code с разрешениями
# и запускает Remote Control. Повторный запуск безопасен: готовое пропускается.
# Секреты (SSH-ключ) приходят только из зашифрованного veliq-dev.bin (dev-pc-export.ps1).
# Ничего не меняет на VPS, не удаляет установленную службу и её файлы.
[CmdletBinding()]
param(
 [string]$KitFile='',
 [switch]$NoKit,
 [switch]$SkipService,
 [switch]$SkipClaude,
 [switch]$NoRemoteControl,
 [string]$ControlPlaneUrl='https://146.0.77.98:8443',
 [string]$ProbeTarget='146.0.77.98:40000',
 [string]$GitName='buzzmean',
 [string]$GitEmail='255260433+buzzmean@users.noreply.github.com'
)
$ErrorActionPreference='Stop'
Set-StrictMode -Version Latest
$script:DevPcCaPin='622db56fba69041b108fc31b5b3cf81a5ed722faa3110ef2b896b8fac42abd85'
$script:DevPcVpsHosts=@('146.0.77.98','2.56.177.4')

function Write-DevPcStage([string]$Text) {
 $elapsed=if($script:DevPcStarted){[int]([DateTime]::UtcNow-$script:DevPcStarted).TotalSeconds}else{0}
 Write-Host ('== ['+[TimeSpan]::FromSeconds($elapsed).ToString('mm\:ss')+'] '+$Text) -ForegroundColor Cyan
}

function Update-DevPcPath {
 $parts=@([Environment]::GetEnvironmentVariable('Path','Machine'),[Environment]::GetEnvironmentVariable('Path','User'),(Join-Path $env:USERPROFILE '.local\bin'),(Join-Path $env:ProgramFiles 'Git\cmd'))
 $env:Path=(@($parts|Where-Object {$_})) -join ';'
}

function Test-DevPcCA([string]$Path) {
 if(!(Test-Path -LiteralPath $Path -PathType Leaf)){return $false}
 $text=[IO.File]::ReadAllText($Path)
 if($text -match 'PRIVATE KEY' -or ([regex]::Matches($text,'-----BEGIN CERTIFICATE-----')).Count -ne 1){return $false}
 try{$cert=New-Object Security.Cryptography.X509Certificates.X509Certificate2 $Path}catch{return $false}
 $sha=[Security.Cryptography.SHA256]::Create()
 try{$hash=([BitConverter]::ToString($sha.ComputeHash($cert.RawData))).Replace('-','').ToLowerInvariant()}finally{$sha.Dispose();$cert.Dispose()}
 return $hash -ceq $script:DevPcCaPin
}

# Обратная к Protect-DevPcKit из dev-pc-export.ps1; MAC проверяется до расшифровки.
function Unprotect-DevPcKit([string]$Source,[string]$Destination,[string]$Passphrase) {
 $in=[IO.File]::OpenRead($Source)
 try{
  $header=New-Object byte[] 37
  if($in.Length -lt 37+16+32 -or $in.Read($header,0,37) -ne 37 -or [Text.Encoding]::ASCII.GetString($header,0,5) -ne 'VLQD1'){throw 'DEV_KIT_FORMAT'}
  $kdf=New-Object Security.Cryptography.Rfc2898DeriveBytes ($Passphrase,[byte[]]$header[5..20],600000,[Security.Cryptography.HashAlgorithmName]::SHA256)
  $keys=$kdf.GetBytes(64);$kdf.Dispose()
  $body=$in.Length-32
  $mac=New-Object Security.Cryptography.HMACSHA256 (,[byte[]]$keys[32..63])
  [void]$mac.TransformBlock($header,0,37,$null,0)
  $buffer=New-Object byte[] 1048576;$left=$body-37
  while($left -gt 0){$n=$in.Read($buffer,0,[int][Math]::Min($buffer.Length,$left));if($n -le 0){throw 'DEV_KIT_TRUNCATED'};[void]$mac.TransformBlock($buffer,0,$n,$null,0);$left-=$n}
  [void]$mac.TransformFinalBlock($buffer,0,0)
  $tag=New-Object byte[] 32;[void]$in.Read($tag,0,32)
  $diff=0;for($i=0;$i -lt 32;$i++){$diff=$diff -bor ($tag[$i] -bxor $mac.Hash[$i])}
  $mac.Dispose()
  if($diff){throw 'DEV_KIT_WRONG_PASSPHRASE'}
  $in.Position=37
  $aes=[Security.Cryptography.Aes]::Create();$aes.Key=[byte[]]$keys[0..31];$aes.IV=[byte[]]$header[21..36]
  $out=[IO.File]::Open($Destination,[IO.FileMode]::CreateNew)
  try{
   $crypto=New-Object Security.Cryptography.CryptoStream ($out,$aes.CreateDecryptor(),[Security.Cryptography.CryptoStreamMode]::Write)
   $left=$body-37
   while($left -gt 0){$n=$in.Read($buffer,0,[int][Math]::Min($buffer.Length,$left));$crypto.Write($buffer,0,$n);$left-=$n}
   $crypto.FlushFinalBlock();$crypto.Dispose()
  }finally{$out.Dispose();$aes.Dispose()}
 }finally{$in.Dispose()}
}

# Разрешения Claude Code (решение владельца 25.10.2026): git/go/dotnet/python и SSH
# только к двум VPS проекта без вопросов; force-push запрещён. Правила пользователя сохраняются.
function Get-DevPcClaudeRules {
 $allow=New-Object Collections.Generic.List[string]
 foreach($tool in @('Bash','PowerShell')){
  foreach($command in @('git','go','dotnet','python','py','gofmt')){$allow.Add("${tool}(${command}:*)")}
  foreach($vps in $script:DevPcVpsHosts){
   $allow.Add("${tool}(ssh root@${vps}:*)")
   $allow.Add("${tool}(ssh -o BatchMode=yes root@${vps}:*)")
   $allow.Add("${tool}(ssh -o BatchMode=yes -o ConnectTimeout=10 root@${vps}:*)")
  }
 }
 $deny=@(foreach($tool in @('Bash','PowerShell')){foreach($flag in @('--force','-f','--force-with-lease')){"${tool}(git push ${flag}:*)"}})
 return @{allow=@($allow);deny=@($deny)}
}

function Merge-DevPcClaudeSettings([string]$Path) {
 $rules=Get-DevPcClaudeRules
 if(Test-Path -LiteralPath $Path){$settings=Get-Content -Raw -LiteralPath $Path|ConvertFrom-Json}else{$settings=New-Object psobject}
 if($null -eq $settings.PSObject.Properties['permissions']){$settings|Add-Member -NotePropertyName permissions -NotePropertyValue (New-Object psobject)}
 $permissions=$settings.permissions
 foreach($kind in @('allow','deny')){
  $current=@();if($null -ne $permissions.PSObject.Properties[$kind]){$current=@($permissions.$kind)}
  $merged=@($current)+@($rules[$kind]|Where-Object {$current -notcontains $_})
  if($null -ne $permissions.PSObject.Properties[$kind]){$permissions.$kind=$merged}else{$permissions|Add-Member -NotePropertyName $kind -NotePropertyValue $merged}
 }
 if($null -eq $permissions.PSObject.Properties['defaultMode']){$permissions|Add-Member -NotePropertyName defaultMode -NotePropertyValue 'acceptEdits'}
 [void][IO.Directory]::CreateDirectory((Split-Path $Path))
 [IO.File]::WriteAllText($Path,($settings|ConvertTo-Json -Depth 8),(New-Object Text.UTF8Encoding $false))
}

function Install-DevPcPackage([string]$Id,[scriptblock]$Present,[string[]]$Extra=@()) {
 if(& $Present){Write-Host "$Id — уже есть";return}
 Write-Host "$Id — ставлю..."
 & winget install --exact --id $Id --source winget --silent --accept-package-agreements --accept-source-agreements --disable-interactivity @Extra
 # -1978335189 = APPINSTALLER_CLI_ERROR_UPDATE_NOT_APPLICABLE (уже установлен).
 if($LASTEXITCODE -and $LASTEXITCODE -ne -1978335189){throw "winget не смог поставить $Id (код $LASTEXITCODE)"}
 Update-DevPcPath
 if(!(& $Present)){throw "$Id установлен, но не найден в PATH; откройте новое окно PowerShell и запустите скрипт ещё раз"}
}

function Find-DevPcKit {
 if($KitFile){return (Resolve-Path -LiteralPath $KitFile).Path}
 $folders=@((Join-Path $env:USERPROFILE 'Downloads'),[Environment]::GetFolderPath('Desktop'),'C:\veliq')
 try{$folders+=[Environment]::ExpandEnvironmentVariables((Get-ItemProperty 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Explorer\User Shell Folders').'{374DE290-123F-4565-9164-39C4925E467B}')}catch{}
 $found=@(foreach($folder in ($folders|Select-Object -Unique)){if($folder -and (Test-Path -LiteralPath $folder)){Get-ChildItem -LiteralPath $folder -Filter 'veliq-dev*.bin' -File -ErrorAction SilentlyContinue}})
 if(!$found.Count){return $null}
 return ($found|Sort-Object LastWriteTimeUtc -Descending|Select-Object -First 1).FullName
}

function Install-DevPcSshFile([string]$Source,[string]$Owner) {
 $sshTarget=Join-Path $env:USERPROFILE '.ssh'
 [void][IO.Directory]::CreateDirectory($sshTarget)
 $name=Split-Path $Source -Leaf
 $destination=Join-Path $sshTarget $name
 if(Test-Path -LiteralPath $destination){
  if((Get-FileHash -LiteralPath $destination).Hash -ne (Get-FileHash -LiteralPath $Source).Hash){Write-Warning "~\.ssh\$name уже есть и отличается от набора — оставлен как есть"}
  return
 }
 Copy-Item -LiteralPath $Source -Destination $destination
 if($name -notmatch '\.pub$' -and $name -notin @('known_hosts','config')){
  # OpenSSH не примет ключ, который могут читать другие пользователи.
  & icacls $destination /inheritance:r /grant:r "*$($Owner):F" '*S-1-5-18:F' '*S-1-5-32-544:F'|Out-Null
  if($LASTEXITCODE){throw 'Не удалось закрыть права на SSH-ключ'}
 }
}

function Import-DevPcKit([string]$Path,[string]$Repo,[string]$Owner) {
 $stage=Join-Path $env:ProgramData ('veliq-dev-'+[Guid]::NewGuid().ToString('N'))
 [void][IO.Directory]::CreateDirectory($stage)
 try{
  & icacls $stage /inheritance:r /grant:r "*$($Owner):(OI)(CI)F" '*S-1-5-18:(OI)(CI)F' '*S-1-5-32-544:(OI)(CI)F'|Out-Null
  if($LASTEXITCODE){throw 'Не удалось закрыть временную папку'}
  $zip=Join-Path $stage 'kit.zip'
  for($attempt=1;;$attempt++){
   $secure=Read-Host "Пароль от $(Split-Path $Path -Leaf)" -AsSecureString
   $passphrase=[Runtime.InteropServices.Marshal]::PtrToStringBSTR([Runtime.InteropServices.Marshal]::SecureStringToBSTR($secure))
   try{Unprotect-DevPcKit $Path $zip $passphrase;break}
   catch{
    Remove-Item -LiteralPath $zip -Force -ErrorAction SilentlyContinue
    if($_.Exception.Message -notmatch 'DEV_KIT_WRONG_PASSPHRASE'){throw "Файл набора повреждён или не тот: $($_.Exception.Message)"}
    if($attempt -ge 3){throw 'Пароль не подошёл 3 раза'}
    Write-Warning 'Пароль не подошёл, попробуйте ещё раз'
   }finally{$passphrase=$null}
  }
  $kit=Join-Path $stage 'kit'
  Add-Type -AssemblyName System.IO.Compression.FileSystem
  [IO.Compression.ZipFile]::ExtractToDirectory($zip,$kit)
  $info=Get-Content -Raw -LiteralPath (Join-Path $kit 'kit.json')|ConvertFrom-Json
  if($info.format -ne 1){throw 'Неизвестная версия набора'}
  foreach($entry in $info.files.PSObject.Properties){
   $file=Join-Path $kit $entry.Name
   if($entry.Name -match '\.\.' -or !(Test-Path -LiteralPath $file -PathType Leaf) -or (Get-FileHash -LiteralPath $file -Algorithm SHA256).Hash -ne $entry.Value){throw "Набор повреждён: $($entry.Name)"}
  }
  $ssh=Join-Path $kit 'ssh'
  if(Test-Path -LiteralPath $ssh){foreach($file in Get-ChildItem -LiteralPath $ssh -File){Install-DevPcSshFile $file.FullName $Owner};Write-Host 'SSH-ключ установлен в ~\.ssh (содержимое не выводится)'}
  $ca=Join-Path $kit 'trust\ca.pem'
  if(Test-Path -LiteralPath $ca){Install-DevPcCA $ca $Repo}
 }finally{Remove-Item -LiteralPath $stage -Recurse -Force -ErrorAction SilentlyContinue}
}

function Install-DevPcCA([string]$Source,[string]$Repo) {
 if(!(Test-DevPcCA $Source)){throw "CA $Source не совпал с пином 62:2D:B5…; файл не используется"}
 $trust=Join-Path $Repo '.e2e\activation-trust'
 [void][IO.Directory]::CreateDirectory($trust)
 foreach($name in @('ca.pem','cp-ca.pem')){$target=Join-Path $trust $name;if(!(Test-DevPcCA $target)){Copy-Item -LiteralPath $Source -Destination $target -Force}}
 Write-Host 'Публичный CA на месте (отпечаток совпал)'
}

function Find-DevPcCA([string]$Repo) {
 foreach($candidate in @((Join-Path $Repo '.e2e\activation-trust\ca.pem'),'C:\veliq\enroll\ca.pem',(Join-Path $env:USERPROFILE 'Downloads\ca.pem'),(Join-Path $env:ProgramData 'neexitlag\cp-ca.pem'))){
  if(Test-DevPcCA $candidate){return $candidate}
 }
 return $null
}

function Install-DevPcService([string]$Repo,[string]$Owner) {
 $tools=Join-Path $Repo 'tools\windows'
 $built=Join-Path $Repo 'bin\windows-e2e\neexitlag-service.exe'
 $installDir=Join-Path $env:ProgramData 'neexitlag'
 $installed=Join-Path $installDir 'neexitlag-service.exe'
 if(Get-Service neexitlag -ErrorAction SilentlyContinue){
  $manifest=Get-Content -Raw -LiteralPath (Join-Path $installDir 'service.json')|ConvertFrom-Json
  if([string]$manifest.owner_sid -ne $Owner){Write-Warning 'Служба VELIQ установлена другим пользователем Windows — не трогаю. Войдите под ним или попросите Claude переустановить.';return}
  if((Get-FileHash -LiteralPath $built).Hash -eq (Get-FileHash -LiteralPath $installed).Hash){Write-Host 'Служба уже этой версии'}
  else{
   & (Join-Path $tools 'Close-VeliqGuiForUpdate.ps1')
   & (Join-Path $tools 'update-service.ps1') -ServiceExecutable $built
  }
 }elseif(Test-Path -LiteralPath $installDir){
  Write-Warning "$installDir есть, а службы нет. Файлы не трогаю (там может быть ключ устройства); спросите Claude."
  return
 }else{
  $trust=Join-Path $Repo '.e2e\activation-trust'
  if(!(Test-DevPcCA (Join-Path $trust 'ca.pem'))){Write-Warning 'Нет публичного CA (он приходит в veliq-dev.bin). Служба не установлена; положите файл в «Загрузки» и запустите скрипт ещё раз.';return}
  $driver=Join-Path $Repo '.e2e\windivert'
  $pins=Get-Content -Raw -LiteralPath (Join-Path $driver 'pins.json')|ConvertFrom-Json
  & (Join-Path $tools 'install.ps1') -ActivationOnly -ServiceExecutable $built -OwnerSid $Owner -TrustDirectory $trust -DriverDirectory $driver -DllSha256 $pins.dll_sha256 -DriverSha256 $pins.driver_sha256 -ControlPlaneUrl $ControlPlaneUrl -ProbeTarget $ProbeTarget -TCPProbeTarget $ProbeTarget
 }
 & (Join-Path $tools 'install-desktop.ps1') -SourceDirectory (Join-Path $Repo 'bin\release')
 $same=(Get-FileHash -LiteralPath $built).Hash -eq (Get-FileHash -LiteralPath $installed).Hash
 $status=(Get-Service neexitlag).Status
 Write-Host "служба: $status; SHA256 собранной и установленной совпадает: $same"
 if(!$same -or [string]$status -ne 'Running'){throw 'Проверка службы не прошла'}
 $gui=Join-Path $env:LOCALAPPDATA 'Programs\VELIQ\VELIQ.exe'
 # explorer.exe запускает GUI без прав администратора, как и положено.
 if(!@(Get-Process -Name VELIQ -ErrorAction SilentlyContinue).Count){Start-Process explorer.exe -ArgumentList ('"'+$gui+'"')}
 $state='UNKNOWN'
 try{. (Join-Path $tools 'common.ps1');$state=[string](Invoke-ServiceIPC 'activation.status').state}catch{}
 if($state -ne 'ACTIVATED'){$script:DevPcTodo+='Введите свой код в окне VELIQ и нажмите «Оптимизировать» (тот же код, что раньше: место не тратится).'}
}

function Start-DevPcRemoteControl([string]$Repo) {
 # Отдельное окно, чтобы облачный Claude подключился, пока здесь идёт сборка.
 $command="Set-Location -LiteralPath '$($Repo.Replace("'","''"))'; claude remote-control; if(`$LASTEXITCODE){Write-Host 'Нужен вход в Claude: в открывшемся claude введите /login, после входа /exit — Remote Control запустится сам.' -ForegroundColor Yellow; claude; claude remote-control}"
 Start-Process powershell.exe -ArgumentList @('-NoExit','-NoProfile','-ExecutionPolicy','Bypass','-Command',$command)
}

function Invoke-DevPc {
 $script:DevPcStarted=[DateTime]::UtcNow
 $script:DevPcTodo=@()
 $repo=(Resolve-Path (Join-Path $PSScriptRoot '..\..')).Path
 $base=Split-Path $repo
 $logs=Join-Path $base 'logs';[void][IO.Directory]::CreateDirectory($logs)
 Start-Transcript -Path (Join-Path $logs ('dev-pc-'+(Get-Date -Format 'yyyyMMdd-HHmmss')+'.log'))|Out-Null
 try{
  Write-DevPcStage 'Проверка прав'
  $identity=[Security.Principal.WindowsIdentity]::GetCurrent()
  if(!([Security.Principal.WindowsPrincipal]$identity).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){throw 'Запустите PowerShell от имени администратора'}
  $owner=$identity.User.Value
  $shell=@(Get-CimInstance Win32_Process -Filter "Name='explorer.exe'"|ForEach-Object{(Invoke-CimMethod -InputObject $_ -MethodName GetOwnerSid).Sid}|Select-Object -Unique)
  if($shell.Count -and $shell -notcontains $owner){throw 'PowerShell повышен под другим пользователем. Запустите его от администратора под тем, кто сидит за ПК.'}
  if(!(Get-Command winget -ErrorAction SilentlyContinue)){throw 'Нет winget (App Installer). Поставьте «Установщик приложений» из Microsoft Store и запустите снова.'}
  Update-DevPcPath

  # Всё, что требует человека, — в начале; дальше можно отойти.
  Write-DevPcStage 'Набор рабочего места (SSH-ключ, CA)'
  if(!$NoKit){
   $kit=Find-DevPcKit
   if($kit){Import-DevPcKit $kit $repo $owner}
   else{$script:DevPcTodo+='Нет veliq-dev.bin в «Загрузках»: SSH к серверам и установка службы пропущены. Скачайте файл из проекта и запустите скрипт ещё раз.'}
  }
  if(!(Test-DevPcCA (Join-Path $repo '.e2e\activation-trust\ca.pem'))){$ca=Find-DevPcCA $repo;if($ca){Install-DevPcCA $ca $repo}}

  Write-DevPcStage 'Программы (winget)'
  Install-DevPcPackage 'Git.Git' {[bool](Get-Command git -ErrorAction SilentlyContinue)}
  Install-DevPcPackage 'GoLang.Go' {[bool](Get-Command go -ErrorAction SilentlyContinue)}
  Install-DevPcPackage 'Microsoft.DotNet.SDK.8' {(Get-Command dotnet -ErrorAction SilentlyContinue) -and @(& dotnet --list-sdks 2>$null|Where-Object {$_ -match '^8\.'}).Count}
  Install-DevPcPackage 'Python.Python.3.12' {[bool](Get-Command py -ErrorAction SilentlyContinue)} @('--scope','machine','--override','/quiet InstallAllUsers=1 PrependPath=1 Include_launcher=1')

  Write-DevPcStage 'Git'
  if(!(& git config --global user.name)){& git config --global user.name $GitName}
  if(!(& git config --global user.email)){& git config --global user.email $GitEmail}
  & git config --global core.autocrlf false
  Write-Host "репозиторий: $repo, ветка $((& git -C $repo branch --show-current))"

  if(!$SkipClaude){
   Write-DevPcStage 'Claude Code'
   if(!(Get-Command claude -ErrorAction SilentlyContinue)){
    # Официальный установщик Claude Code (ставит в ~\.local\bin).
    Invoke-RestMethod https://claude.ai/install.ps1|Invoke-Expression
    Update-DevPcPath
    if(!(Get-Command claude -ErrorAction SilentlyContinue)){throw 'Claude Code не установился'}
   }else{Write-Host 'Claude Code — уже есть'}
   Merge-DevPcClaudeSettings (Join-Path $env:USERPROFILE '.claude\settings.json')
   Write-Host 'Разрешения Claude Code: git/go/dotnet/python и ssh к 146.0.77.98, 2.56.177.4 без вопросов; force-push запрещён'
   if(!$NoRemoteControl){Start-DevPcRemoteControl $repo;$script:DevPcTodo+='В окне Remote Control войдите в Claude, если попросит (один раз).'}
  }

  Write-DevPcStage 'NuGet и WinDivert'
  . (Join-Path $PSScriptRoot 'activation-nuget.ps1')
  Initialize-ActivationNuGetSource
  $driver=Join-Path $repo '.e2e\windivert'
  if(!(Test-Path -LiteralPath (Join-Path $driver 'pins.json'))){
   if(Test-Path -LiteralPath $driver){throw "$driver есть без pins.json; проверьте папку вручную"}
   & (Join-Path $PSScriptRoot 'fetch-windivert.ps1') -Destination $driver
  }

  Write-DevPcStage 'Сборка службы (build-e2e)'
  & (Join-Path $PSScriptRoot 'build-e2e.ps1') -DriverDirectory $driver
  Write-DevPcStage 'Сборка GUI (build-desktop)'
  & (Join-Path $PSScriptRoot 'build-desktop.ps1')

  if(!$SkipService){
   Write-DevPcStage 'Служба и GUI VELIQ'
   try{Install-DevPcService $repo $owner}catch{Write-Warning "Служба: $($_.Exception.Message)";$script:DevPcTodo+='Служба VELIQ не поставилась — попросите Claude посмотреть лог.'}
  }

  Write-DevPcStage 'Готово'
  foreach($vps in $script:DevPcVpsHosts){
   if(Test-Path -LiteralPath (Join-Path $env:USERPROFILE '.ssh')){
    & ssh -o BatchMode=yes -o ConnectTimeout=8 "root@$vps" 'true' 2>$null
    Write-Host ("ssh root@${vps}: "+$(if($LASTEXITCODE -eq 0){'OK'}else{'нет доступа'}))
   }
  }
  if($script:DevPcTodo.Count){Write-Host 'Осталось сделать вам:' -ForegroundColor Yellow;foreach($item in $script:DevPcTodo){Write-Host " - $item" -ForegroundColor Yellow}}
  else{Write-Host 'Всё готово. Claude может работать на этом ПК.' -ForegroundColor Green}
 }finally{Stop-Transcript|Out-Null}
}

if($MyInvocation.InvocationName -eq '.'){return}
Invoke-DevPc
