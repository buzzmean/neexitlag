# One private interactive code input on a clean ActivationOnly image. Never enroll,
# clone a private device key, print the code, or write a credential into the repo.
[CmdletBinding()]
param([string]$InstallDirectory="$env:ProgramData\neexitlag")
$ErrorActionPreference='Stop'
. (Join-Path $PSScriptRoot 'common.ps1')
Assert-Administrator
foreach($name in @('identity.json','device.pem')){if(Test-Path -LiteralPath (Join-Path $InstallDirectory $name)){throw 'Image staging requires a clean activation-only installation; existing identity preserved'}}
$pending=Join-Path $InstallDirectory 'state\activation-pending'
if((Test-Path -LiteralPath $pending) -and @(Get-ChildItem -LiteralPath $pending -Filter identity.json -Recurse -File).Count -gt 0){throw 'Image contains enrollment key material; preserved. Use a clean activation-only image.'}
$secure=Read-Host 'Введите код активации клуба' -AsSecureString
$ptr=[IntPtr]::Zero
try{
 $ptr=[Runtime.InteropServices.Marshal]::SecureStringToBSTR($secure)
 $code=[Runtime.InteropServices.Marshal]::PtrToStringBSTR($ptr)
 if($code -notmatch '^VQ(-[A-Z2-7]{4}){6}$'){throw 'ACTIVATION_INVALID'}
 $null=Invoke-ServiceIPC 'activation.set' @{code=$code;stage_for_boot=$true} (Join-Path $InstallDirectory 'ipc.token')
 Write-Host 'Код сохранён в защищённом хранилище для следующей загрузки. Сервер ещё не проверял код. Не активируйте образ перед копированием.'
}finally{
 if($ptr -ne [IntPtr]::Zero){[Runtime.InteropServices.Marshal]::ZeroFreeBSTR($ptr)}
 Remove-Variable code -ErrorAction SilentlyContinue
 $secure.Dispose()
}
