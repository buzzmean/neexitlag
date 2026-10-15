[CmdletBinding()]
param([string]$Output=(Join-Path $PSScriptRoot '..\..\bin\release'))
$ErrorActionPreference='Stop'
$root=(Resolve-Path (Join-Path $PSScriptRoot '..\..')).Path
if(!(Get-Command dotnet -ErrorAction SilentlyContinue)){throw 'Install .NET SDK 8 to build the desktop app.'}
New-Item $Output -ItemType Directory -Force|Out-Null
$Output=(Resolve-Path $Output).Path
& dotnet publish (Join-Path $root 'apps\desktop\Neexitlag.Desktop.csproj') -c Release -r win-x64 --self-contained true -p:PublishSingleFile=true -p:IncludeNativeLibrariesForSelfExtract=true -p:PublishTrimmed=false -p:DebugType=None -p:DebugSymbols=false -o $Output --nologo
if($LASTEXITCODE){throw 'NeExitLag desktop build failed.'}
if(!(Test-Path (Join-Path $Output 'NeExitLag.exe'))){throw 'NeExitLag executable was not created.'}
Write-Host "Ready: $(Join-Path $Output 'NeExitLag.exe')"
Write-Host 'Run the desktop app as a standard user authorized to access the service.'

$scripts=Join-Path $Output 'tools\windows'
New-Item -ItemType Directory -Force $scripts|Out-Null
Copy-Item (Join-Path $PSScriptRoot '*.ps1') $scripts -Force
