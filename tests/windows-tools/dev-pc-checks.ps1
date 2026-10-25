#requires -Version 5.1
# dev-pc.ps1 / dev-pc-export.ps1 without winget, builds, service or network:
# kit round-trip, wrong passphrase/tamper rejection and CA pin.
$ErrorActionPreference='Stop';Set-StrictMode -Version Latest
$tools=Join-Path $PSScriptRoot '..\..\tools\windows'
. (Join-Path $tools 'dev-pc-export.ps1')
$exportPin=$script:DevPcCaPin
. (Join-Path $tools 'dev-pc.ps1')
if($script:DevPcCaPin -cne $exportPin){throw 'CA pin differs between dev-pc.ps1 and dev-pc-export.ps1'}
$temp=Join-Path ([IO.Path]::GetTempPath()) ('veliq-dev-pc-checks-'+[Guid]::NewGuid().ToString('N'))
[void][IO.Directory]::CreateDirectory($temp)
try {
 # Kit encryption round-trip.
 $plain=Join-Path $temp 'plain.bin';$bytes=New-Object byte[] 300000;(New-Object Random 7).NextBytes($bytes);[IO.File]::WriteAllBytes($plain,$bytes)
 $kit=Join-Path $temp 'veliq-dev.bin'
 Protect-DevPcKit $plain $kit 'correct horse battery'
 $out=Join-Path $temp 'out.bin'
 Unprotect-DevPcKit $kit $out 'correct horse battery'
 if((Get-FileHash $plain).Hash -ne (Get-FileHash $out).Hash){throw 'Kit round-trip changed bytes'}
 Remove-Item $out
 $rejected=$false;try{Unprotect-DevPcKit $kit $out 'wrong passphrase!!'}catch{$rejected=$_.Exception.Message -match 'DEV_KIT_WRONG_PASSPHRASE'}
 if(!$rejected -or (Test-Path $out)){throw 'Wrong passphrase accepted or plaintext written'}
 $raw=[IO.File]::ReadAllBytes($kit);$raw[100]=$raw[100] -bxor 1;$tampered=Join-Path $temp 'tampered.bin';[IO.File]::WriteAllBytes($tampered,$raw)
 $rejected=$false;try{Unprotect-DevPcKit $tampered $out 'correct horse battery'}catch{$rejected=$_.Exception.Message -match 'DEV_KIT_WRONG_PASSPHRASE'}
 if(!$rejected -or (Test-Path $out)){throw 'Tampered kit accepted'}
 [IO.File]::WriteAllBytes($tampered,[Text.Encoding]::ASCII.GetBytes('VLQK1'+('x'*200)))
 $rejected=$false;try{Unprotect-DevPcKit $tampered $out 'correct horse battery'}catch{$rejected=$_.Exception.Message -match 'DEV_KIT_FORMAT'}
 if(!$rejected){throw 'Foreign kit format accepted'}

 # CA pin: only the pinned public certificate passes.
 $rsa=[Security.Cryptography.RSA]::Create(2048)
 $request=New-Object Security.Cryptography.X509Certificates.CertificateRequest ('CN=fixture CA',$rsa,[Security.Cryptography.HashAlgorithmName]::SHA256,[Security.Cryptography.RSASignaturePadding]::Pkcs1)
 $request.CertificateExtensions.Add((New-Object Security.Cryptography.X509Certificates.X509BasicConstraintsExtension ($true,$false,0,$true)))
 $cert=$request.CreateSelfSigned([DateTimeOffset]::UtcNow.AddDays(-1),[DateTimeOffset]::UtcNow.AddDays(30))
 $pem="-----BEGIN CERTIFICATE-----`n"+[Convert]::ToBase64String($cert.RawData,[Base64FormattingOptions]::InsertLineBreaks)+"`n-----END CERTIFICATE-----`n"
 $caPath=Join-Path $temp 'ca.pem';[IO.File]::WriteAllText($caPath,$pem)
 if(Test-DevPcCA $caPath){throw 'Unpinned CA accepted'}
 $sha=[Security.Cryptography.SHA256]::Create();$fixturePin=([BitConverter]::ToString($sha.ComputeHash($cert.RawData))).Replace('-','').ToLowerInvariant();$sha.Dispose()
 $savedPin=$script:DevPcCaPin;$script:DevPcCaPin=$fixturePin
 try{
  if(!(Test-DevPcCA $caPath)){throw 'Pinned CA rejected'}
  [IO.File]::WriteAllText((Join-Path $temp 'bad.pem'),$pem+"-----BEGIN PRIVATE KEY-----`nAAAA`n-----END PRIVATE KEY-----`n")
  if(Test-DevPcCA (Join-Path $temp 'bad.pem')){throw 'PEM with a private key accepted'}
  if(Test-DevPcCA (Join-Path $temp 'missing.pem')){throw 'Missing CA accepted'}
 }finally{$script:DevPcCaPin=$savedPin;$cert.Dispose();$rsa.Dispose()}

 # Claude settings: user rules and mode preserved, ours added once, force-push denied.
 $settings=Join-Path $temp '.claude\settings.json'
 Merge-DevPcClaudeSettings $settings
 $first=Get-Content -Raw $settings|ConvertFrom-Json
 if($first.permissions.defaultMode -ne 'acceptEdits'){throw 'defaultMode not set'}
 foreach($rule in @('Bash(git:*)','PowerShell(ssh root@146.0.77.98:*)','Bash(ssh -o BatchMode=yes -o ConnectTimeout=10 root@2.56.177.4:*)')){if(@($first.permissions.allow) -notcontains $rule){throw "Missing allow rule $rule"}}
 if(@($first.permissions.deny) -notcontains 'Bash(git push --force:*)'){throw 'force-push not denied'}
 if(@($first.permissions.allow|Where-Object {$_ -match 'ssh' -and $_ -notmatch '146\.0\.77\.98|2\.56\.177\.4'}).Count){throw 'SSH allowed to an unexpected host'}
 [IO.File]::WriteAllText($settings,'{"model":"x","permissions":{"allow":["Bash(npm:*)"],"defaultMode":"plan"}}')
 Merge-DevPcClaudeSettings $settings;Merge-DevPcClaudeSettings $settings
 $merged=Get-Content -Raw $settings|ConvertFrom-Json
 if($merged.model -ne 'x' -or $merged.permissions.defaultMode -ne 'plan' -or @($merged.permissions.allow) -notcontains 'Bash(npm:*)'){throw 'User settings not preserved'}
 if(@($merged.permissions.allow|Where-Object {$_ -eq 'Bash(git:*)'}).Count -ne 1){throw 'Rules duplicated on rerun'}
 Write-Output 'DEV_PC_CHECKS_PASS'
} finally {Remove-Item -LiteralPath $temp -Recurse -Force -ErrorAction SilentlyContinue}
