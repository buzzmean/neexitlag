# Windows-only native-window smoke; does not claim real network acceptance.
[CmdletBinding()]
param([string]$Executable=(Join-Path $PSScriptRoot '..\..\bin\release\NeExitLag.exe'))
$ErrorActionPreference='Stop'
Add-Type -AssemblyName UIAutomationClient
Add-Type -AssemblyName UIAutomationTypes
$Executable=(Resolve-Path $Executable).Path
$first=$null;$second=$null
try {
 $first=Start-Process $Executable -PassThru
 $deadline=[DateTime]::UtcNow.AddSeconds(30)
 do {$first.Refresh();if($first.HasExited){throw 'Нативное приложение завершилось при запуске'};if($first.MainWindowHandle -ne 0){break};Start-Sleep -Milliseconds 200}while([DateTime]::UtcNow -lt $deadline)
 if($first.MainWindowHandle -eq 0 -or $first.MainWindowTitle -ne 'NeExitLag'){throw 'Нативное окно NeExitLag не открылось'}
 $window=[Windows.Automation.AutomationElement]::FromHandle($first.MainWindowHandle)
 $deadline=[DateTime]::UtcNow.AddSeconds(15)
 do {
  $names=@($window.FindAll([Windows.Automation.TreeScope]::Descendants,[Windows.Automation.Condition]::TrueCondition)|ForEach-Object {$_.Current.Name})
  $missing=@(@('Главная','Приложения','Маршруты','Статистика','Диагностика','Настройки','Локальные стратегии','Повторить подключение')|Where-Object {$_ -notin $names})
  if(!$missing.Count){break};Start-Sleep -Milliseconds 200
 }while([DateTime]::UtcNow -lt $deadline)
 if($missing.Count){throw ('Отсутствуют русские элементы окна: '+($missing -join ', '))}
 if(@($names|Where-Object {$_ -in @('Home','Applications','Routes','Statistics','Diagnostics','Settings')}).Count){throw 'Обнаружены английские элементы интерфейса'}
 $second=Start-Process $Executable -PassThru
 if(!$second.WaitForExit(10000)){throw 'Второй экземпляр не передал активацию первому'}
 $first.Refresh();if($first.HasExited){throw 'Первый экземпляр завершился при активации'}
 Write-Output 'PASS: нативное WPF-окно, русская навигация и единственный экземпляр'
 if(!(Get-Service neexitlag -ErrorAction SilentlyContinue)) {
  $deadline=[DateTime]::UtcNow.AddSeconds(15)
  do {$names=@($window.FindAll([Windows.Automation.TreeScope]::Descendants,[Windows.Automation.Condition]::TrueCondition)|ForEach-Object {$_.Current.Name});if('Служба NeExitLag не установлена' -in $names){break};Start-Sleep -Milliseconds 200}while([DateTime]::UtcNow -lt $deadline)
  if('Служба NeExitLag не установлена' -notin $names){throw 'Отсутствующая служба не отображена по-русски'}
  [void]$first.CloseMainWindow();if(!$first.WaitForExit(10000)){throw 'Окно не завершилось корректно при отсутствующей службе'}
  Write-Output 'PASS: отсутствие службы без падения и корректное завершение'
 }
} finally {
 foreach($process in @($first,$second)){if($null -ne $process){$process.Refresh();if(!$process.HasExited){Stop-Process -Id $process.Id -Force};$process.Dispose()}}
}
