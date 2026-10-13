[CmdletBinding()]param()
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'common.ps1')
Assert-Administrator
$service=Get-Service neexitlag -ErrorAction SilentlyContinue
if($service){
 try {
  if($service.Status -ne 'Stopped'){Stop-Service neexitlag;$service.WaitForStatus('Stopped',[TimeSpan]::FromSeconds(20))}
  & sc.exe delete neexitlag|Out-Null
  if($LASTEXITCODE){throw 'SCM deletion failed'}
 }finally{$service.Dispose()}
 $deadline=[DateTime]::UtcNow.AddSeconds(30)
 do {
  $remaining=Get-Service neexitlag -ErrorAction SilentlyContinue
  if(!$remaining){break}
  $remaining.Dispose()
  if([DateTime]::UtcNow -ge $deadline){throw 'SCM deletion is still pending; close Services/other service handles before reinstalling. Preserved files were not removed.'}
  Start-Sleep -Milliseconds 250
 }while($true)
}
Write-Host 'SCM service removed. Credentials, reports and driver files were preserved for explicit rollback/cleanup.'
