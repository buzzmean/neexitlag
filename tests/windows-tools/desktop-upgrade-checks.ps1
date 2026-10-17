# Isolated per-user GUI migration filesystem/shortcut regression. No real service,
# Chrome policy, credentials, installed GUI, or application process is changed.
$ErrorActionPreference='Stop'
Set-StrictMode -Version Latest
$root=Join-Path ([IO.Path]::GetTempPath()) ('veliq-gui-upgrade-'+[Guid]::NewGuid().ToString('N'))
$source=Join-Path $root 'build';$installed=Join-Path $root 'installed';$links=Join-Path $root 'links'
$runKey='HKCU:\Software\Microsoft\Windows\CurrentVersion\Run'
$hadRunKey=Test-Path $runKey;$oldRun=$null;$hadRun=$false
if($hadRunKey){$item=Get-Item -LiteralPath $runKey;$hadRun='NeExitLag' -in $item.GetValueNames();if($hadRun){$oldRun=$item.GetValue('NeExitLag');$oldKind=$item.GetValueKind('NeExitLag')}}
try{
 foreach($dir in @($source,$installed,$links)){[void][IO.Directory]::CreateDirectory($dir)}
 # These are inert fixture bytes, not an executable or private key.
 [IO.File]::WriteAllText((Join-Path $source 'VELIQ.exe'),'VELIQ GUI fixture v1')
 $legacy=Join-Path $installed 'NeExitLag.exe';[IO.File]::WriteAllText($legacy,'legacy GUI fixture')
 $protected=@('ProgramData/neexitlag/identity.json','ProgramData/neexitlag/state/profiles.json','LocalAppData/NeExitLag/desktop.json','LocalAppData/NeExitLag/chrome-proxy-policy.json','LocalAppData/NeExitLag/chrome-recovery.json')
 $hashes=@{}
 foreach($relative in $protected){$path=Join-Path $root $relative;[void][IO.Directory]::CreateDirectory((Split-Path $path));[IO.File]::WriteAllText($path,"preserve fixture $relative");$hashes[$relative]=(Get-FileHash $path).Hash}
 $shell=New-Object -ComObject WScript.Shell
 $oldLink=Join-Path $links 'NeExitLag.lnk';$link=$shell.CreateShortcut($oldLink);$link.TargetPath=$legacy;$link.Save()
 if(!$hadRunKey){New-Item -Path $runKey -Force|Out-Null}
 Set-ItemProperty -LiteralPath $runKey -Name NeExitLag -Value ('"'+$legacy+'"')
 $script=Join-Path $PSScriptRoot '../../tools/windows/install-desktop.ps1'
 & $script -SourceDirectory $source -InstallDirectory $installed -ShortcutDirectory $links
 $target=Join-Path $installed 'VELIQ.exe'
 if((Get-FileHash $target).Hash -ne (Get-FileHash $legacy).Hash){throw 'Legacy entrypoint differs'}
 if(Test-Path $oldLink){throw 'Owned legacy shortcut not replaced'}
 if($shell.CreateShortcut((Join-Path $links 'VELIQ.lnk')).TargetPath -ine $target){throw 'New shortcut target mismatch'}
 if((Get-ItemProperty -LiteralPath $runKey).NeExitLag -ne ('"'+$target+'"')){throw 'Enabled owned startup not retargeted'}
 [IO.File]::WriteAllText((Join-Path $source 'VELIQ.exe'),'VELIQ GUI fixture v2')
 & $script -SourceDirectory $source -InstallDirectory $installed -ShortcutDirectory $links
 if([IO.File]::ReadAllText($target) -ne 'VELIQ GUI fixture v2'){throw 'Idempotent update failed'}
 foreach($relative in $protected){if((Get-FileHash (Join-Path $root $relative)).Hash -ne $hashes[$relative]){throw "Upgrade changed $relative"}}
 $backups=@(Get-ChildItem (Join-Path $installed 'backups') -Directory)
 if($backups.Count -ne 2 -or !(Test-Path (Join-Path $backups[0].FullName 'NeExitLag.exe'))){throw 'Persistent legacy backup missing'}
 Write-Host 'PASS VELIQ legacy upgrade, identical alias, shortcut, startup, repeat install, preserved identity/profiles/recovery/preferences'
}finally{
 if($hadRun){New-ItemProperty -LiteralPath $runKey -Name NeExitLag -Value $oldRun -PropertyType $oldKind -Force|Out-Null}else{Remove-ItemProperty -LiteralPath $runKey -Name NeExitLag -ErrorAction SilentlyContinue}
 if(!$hadRunKey -and (Test-Path $runKey)){if((Get-Item $runKey).ValueCount -eq 0){Remove-Item -LiteralPath $runKey}}
 Remove-Item -LiteralPath $root -Recurse -Force -ErrorAction SilentlyContinue
}
