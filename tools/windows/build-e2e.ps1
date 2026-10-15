[CmdletBinding()]
param([string]$Output=(Join-Path $PSScriptRoot '..\..\bin\windows-e2e'))
$ErrorActionPreference='Stop'
$root=(Resolve-Path (Join-Path $PSScriptRoot '..\..')).Path
New-Item -ItemType Directory -Force $Output|Out-Null
$Output=(Resolve-Path $Output).Path
Push-Location $root
$oldOS=$env:GOOS;$oldArch=$env:GOARCH;$oldCGO=$env:CGO_ENABLED
try {
 $env:GOOS='windows';$env:GOARCH='amd64';$env:CGO_ENABLED='0'
 $targets=@{
  'neexitlag-service.exe'='./services/client-service/cmd/service';'test-game.exe'='./tools/windows-test-client';
  'tcp-selected.exe'='./tools/tcp-test-client';'enroll.exe'='./tools/enroll';'vps-provision.exe'='./tools/vps-provision'
 }
 foreach($name in $targets.Keys){& go build -trimpath -o (Join-Path $Output $name) $targets[$name];if($LASTEXITCODE){throw "Build failed: $name"}}
 Copy-Item (Join-Path $Output 'tcp-selected.exe') (Join-Path $Output 'tcp-control.exe')
 Copy-Item (Join-Path $Output 'tcp-selected.exe') (Join-Path $Output 'tcp-plus.exe')
 $scripts=Join-Path $Output 'tools\windows';New-Item -ItemType Directory -Force $scripts|Out-Null;Copy-Item (Join-Path $PSScriptRoot '*.ps1') $scripts
 $licenses=Join-Path $Output 'third_party\licenses';New-Item -ItemType Directory -Force $licenses|Out-Null;Copy-Item (Join-Path $root 'third_party\licenses\*.txt') $licenses;Copy-Item (Join-Path $root 'third_party\NOTICE.md') (Join-Path $Output 'third_party\NOTICE.md')
 $docs=Join-Path $Output 'docs';New-Item -ItemType Directory -Force $docs|Out-Null;Copy-Item (Join-Path $root 'docs\DIRECT_PLUS.md') $docs;Copy-Item (Join-Path $root 'docs\DIRECT_PLUS_V2.md') $docs;Copy-Item (Join-Path $root 'docs\WINDOWS_APPLICATION_PICKER.md') $docs;Copy-Item (Join-Path $root 'docs\AUTO_FLOW_ROUTING.md') $docs;Copy-Item (Join-Path $root 'docs\ADAPTIVE_UDP_HEALTH.md') $docs;Copy-Item (Join-Path $root 'docs\RELAY_SESSION_POOL.md') $docs
 Copy-Item (Join-Path $Output 'test-game.exe') (Join-Path $Output 'control-client.exe')
 Copy-Item (Join-Path $Output 'test-game.exe') (Join-Path $Output 'udp-plus.exe')
 $linux=Join-Path $Output 'linux';New-Item -ItemType Directory -Force $linux|Out-Null
 $env:GOOS='linux'
 foreach($entry in @(@('relay','./services/relay/cmd/relay'),@('control-plane','./services/control-plane/cmd/control-plane'),@('remote-echo','./tools/remote-echo'),@('probe-client','./tools/windows-test-client'),@('web-test-target','./tools/web-test-target'))){& go build -trimpath -o (Join-Path $linux $entry[0]) $entry[1];if($LASTEXITCODE){throw "Linux build failed: $($entry[0])"}}
 Write-Host "Windows test binaries and Linux deployment binaries ready in $Output"
} finally {$env:GOOS=$oldOS;$env:GOARCH=$oldArch;$env:CGO_ENABLED=$oldCGO;Pop-Location}
