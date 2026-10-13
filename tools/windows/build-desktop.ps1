[CmdletBinding()]
param([string]$Output=(Join-Path $PSScriptRoot '..\..\bin\release'))
$ErrorActionPreference='Stop'
$root=(Resolve-Path (Join-Path $PSScriptRoot '..\..')).Path
if(!(Get-Command dotnet -ErrorAction SilentlyContinue)){throw 'Для сборки установите .NET SDK 8.'}
New-Item $Output -ItemType Directory -Force|Out-Null
$Output=(Resolve-Path $Output).Path
& dotnet publish (Join-Path $root 'apps\desktop\Neexitlag.Desktop.csproj') -c Release -r win-x64 --self-contained true -p:PublishSingleFile=true -p:IncludeNativeLibrariesForSelfExtract=true -p:PublishTrimmed=false -p:DebugType=None -p:DebugSymbols=false -o $Output --nologo
if($LASTEXITCODE){throw 'Не удалось собрать приложение NeExitLag.'}
if(!(Test-Path (Join-Path $Output 'NeExitLag.exe'))){throw 'Исполняемый файл не создан.'}
Write-Host "Готово: $(Join-Path $Output 'NeExitLag.exe')"
Write-Host 'Запускайте приложение от обычного пользователя, которому разрешён доступ к службе.'
