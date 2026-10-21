# Коды VELIQ: сценарий владельца после deployment

21.10.2026, по сообщению владельца deployment ecd310a прошёл:
COMPATIBLE_ADDITIVE_SCHEMA2_SECURITY_STATE_PRESERVED, renewal SCHEDULED,
leaf до 20.11.2026, activation_codes=0. Агент на VPS ничего не выполнял.
Ниже команды выполняет только владелец. CA rotation/provision/update повторять
не нужно. Код активации не содержит CA и не заменяет проверенный public trust.

## 1. Создать два кода

На доверенном ПК владельца: PowerShell, обычный пользователь, любая папка.
Не использовать запись терминала/transcript, Tee-Object/перенаправление вывода.
В этой сессии будут один раз показаны секретные коды. Сначала открыть одну SSH
сессию; сверить host key с уже известным ключом сервера, не отключать проверку:

```powershell
ssh root@2.56.177.4
```

Следующие команды — уже в shell VPS под root. Первые две только читают:

```bash
/opt/neexitlag/maintenance/veliq-codes list --ca /etc/neexitlag/control/ca.pem --admin-token-file /etc/neexitlag/control/admin.token
```

```bash
openssl x509 -in /etc/neexitlag/control/ca.pem -noout -fingerprint -sha256
```

Сохранить public CA fingerprint для проверки скачанного public CA ниже.
Создание — это запись в БД, не read-only. Выполнить каждую команду ОДИН раз:

```bash
/opt/neexitlag/maintenance/veliq-codes create --ca /etc/neexitlag/control/ca.pem --admin-token-file /etc/neexitlag/control/admin.token --kind club --limit 40 --name Club-40 --days 90
```

Сразу сохранить `code` и `id` в менеджер паролей с копией вне freeze-ПК
(например, синхронизированном на телефоне владельца). Поле `id` — публичный
идентификатор, `code` — секрет. Не вставлять code в чат/скриншот/командную строку,
не сохранять plaintext в репозитории. `list` не сможет восстановить забытый code.
Клубный код рассчитан на 40 активных машин, не на 40 запусков или CSR.

```bash
/opt/neexitlag/maintenance/veliq-codes create --ca /etc/neexitlag/control/ca.pem --admin-token-file /etc/neexitlag/control/admin.token --kind personal --limit 2 --name Personal-trial --days 7
```

Аналогично сохранить второй code/id. Лимит personal в текущем API ровно 2.
Сроки в этом примере: клуб 90 суток, пробный личный 7 суток. `--days` ограничивает срок приёма новых активаций/enroll. Это не таймер
принудительного отключения уже активированного устройства: ticket/certificate
пути проверяют отзыв device, а не expires_at кода. Для завершения теста явно
использовать revoke ниже; мгновенное закрытие всех существующих сессий не обещается.
На сервере хранится
только SHA256 кода, metadata и слоты; admin token остаётся на VPS. На активированном
клиенте сам код сохраняется в защищённом SYSTEM/Administrators хранилище для
повторного enroll. Нормальный GUI его обратно не читает.

```bash
exit
```

## 2. Пустой Windows-ПК: установить без старой device identity

Git, Go, .NET SDK 8, Python уже установлены. Команды по одной из корня клона в
профиле владельца, например `$HOME\neexitlag`; путь C:\neexitlag не требуется.
Исходники — codex/work. Не выполнять bootstrap-fresh-pc: это другой, SSH-enroll
путь, он не проверяет сценарий активации кодом.

```powershell
git pull --ff-only origin codex/work
```

В обычном PowerShell владельца (при UAC off — существующий full-admin под тем же
пользователем, как на тестовом ПК). Проверить SDK, не только наличие dotnet:

```powershell
dotnet --list-sdks
```

```powershell
. .\tools\windows\activation-nuget.ps1
```

```powershell
Initialize-ActivationNuGetSource
```

```powershell
& .\tools\windows\build-e2e.ps1
```

```powershell
& .\tools\windows\build-desktop.ps1
```

```powershell
& .\tools\windows\fetch-windivert.ps1
```

fetch-windivert предназначен для пустой папки .e2e\windivert; если уже подготовлена,
не удалять/пересоздавать, использовать её pins.json. Скрипт проверяет закреплённые
SHA256 и подпись SYS. Устаревшие/другие драйверы не подставлять.

Создать папку только для ПУБЛИЧНОГО CA:

```powershell
New-Item -ItemType Directory -Force .\.e2e\activation-trust | Out-Null
```

У владельца есть root SSH. Скачать только public ca.pem через SSH с проверенным
host key (пароль вводится в SSH, не в аргументы). Не копировать ca.key/admin.token,
identity/device.pem или весь /etc/neexitlag:

```powershell
scp -o StrictHostKeyChecking=yes root@2.56.177.4:/etc/neexitlag/control/ca.pem .\.e2e\activation-trust\ca.pem
```

На новом ПК сначала установить доверие к SSH host key обычным `ssh root@2.56.177.4`
и сверкой с известным ключом; затем выйти и повторить scp. Отпечаток public CA
вычисляется по DER сертификата (не SHA256 текстового PEM файла):

```powershell
$cert = New-Object Security.Cryptography.X509Certificates.X509Certificate2 ((Resolve-Path .\.e2e\activation-trust\ca.pem).Path)
```

```powershell
$sha = [Security.Cryptography.SHA256]::Create(); try { ([BitConverter]::ToString($sha.ComputeHash($cert.RawData))).Replace('-',':') } finally { $sha.Dispose(); $cert.Dispose() }
```

Он должен совпасть с openssl fingerprint из доверенной SSH-сессии. При несовпадении
остановиться. В текущем развёртывании после ротации один CA обслуживает relay
и control-plane. Нужен также явный cp-ca.pem для HTTPS-клиента control-plane;
одного ca.pem недостаточно для ConnectRemote после активации:

```powershell
Copy-Item .\.e2e\activation-trust\ca.pem .\.e2e\activation-trust\cp-ca.pem
```

Оба файла ПУБЛИЧНЫЕ; содержимое одинаково только для текущей общей CA.
Этот public CA можно включить в доверенный клубный образ/пакет,
тогда SSH на каждом клиенте не нужен. Код и ключи устройства в public пакет не входят.
Публичный CA не скачивать через неподтверждённый HTTPS и не брать из enroll reply.

Для установки службы: elevated PowerShell под ТЕМ ЖЕ владельцем, та же папка клона.
На UAC-off ПК окно уже full-admin; отдельная эскалация не нужна. Служба должна
отсутствовать — это именно чистый ПК; установленную службу не удалять ради теста.

```powershell
$ownerSid = [Security.Principal.WindowsIdentity]::GetCurrent().User.Value
```

```powershell
$pins = Get-Content .\.e2e\windivert\pins.json -Raw | ConvertFrom-Json
```

```powershell
& .\tools\windows\install.ps1 -ActivationOnly -ServiceExecutable .\bin\windows-e2e\neexitlag-service.exe -OwnerSid $ownerSid -TrustDirectory .\.e2e\activation-trust -DriverDirectory .\.e2e\windivert -DllSha256 $pins.dll_sha256 -DriverSha256 $pins.driver_sha256 -ControlPlaneUrl https://2.56.177.4:8443 -ProbeTarget 2.56.177.4:40000 -TCPProbeTarget 2.56.177.4:40000
```

Приватный device key при этом не импортируется. install фиксирует fingerprint
проверенного CA в service.json и сохраняет обычные ACL/owner IPC checks.
В обычном PowerShell того же владельца (при UAC off — существующий режим):

```powershell
& .\tools\windows\install-desktop.ps1
```

```powershell
Start-Process (Join-Path $env:LOCALAPPDATA 'Programs\VELIQ\VELIQ.exe')
```

В GUI должен появиться экран «Введите код активации». Вставить ОДИН из сохранённых
кодов в закрытое поле, нажать «Активировать». При уже существующей legacy identity:
Настройки → «Сменить»; это другой сценарий, не clean-PC. Подождать «Код активации:
активирован», затем включить «Оптимизировать». Не закрывать Chrome/Edge ради активации.

Read-only проверка на ПК, из клона под SID владельца:

```powershell
. .\tools\windows\common.ps1
```

```powershell
Invoke-ServiceIPC 'activation.status' | Select-Object state,error,has_identity,ca_not_after,server_not_after
```

```powershell
Invoke-ServiceIPC 'status' | Select-Object connected,optimizing,demo
```

Успех: state=ACTIVATED, has_identity=True, error пустой, connected=True, demo=False;
после Optimize — optimizing=True. Дополнительно реальное открытие нужного сайта/
приложения и routing diagnostics: ACTIVATED само по себе не доказывает Internet PASS.
CODE_REQUIRED/ACTIVATING/ACTIVATION_ERROR не являются успехом. При отсутствии
активации не выводить activation-code.json/private files; безопасный status выше достаточен.

## 3. Freeze и повторная установка

Если baseline действительно пустой, после reboot программа и введённый код
исчезнут: нужно снова установить ActivationOnly и ввести ТОТ ЖЕ код. Новый код
выпускать не нужно. Для того же (code, hardware hash) новый private key создаётся
на ПК, новый device_id атомарно заменяет предыдущий, прежний device отзывается,
а активный слот остаётся ОДИН. Число строк devices/revoked при этом растёт — это
не число занятых мест. UUID платы, serial и MAC выбранного физического адаптера
должны быть прежними; MachineGuid Windows не используется. Замена/переключение
основного адаптера может изменить hardware hash и занять другое место.

Для клуба без ручного ввода после каждого reboot нужен подготовленный freeze-образ:
ActivationOnly service + GUI + проверенный public CA/driver, БЕЗ device identity.
До сохранения образа, elevated под владельцем, выполнить:

```powershell
& .\tools\windows\stage-club-activation.ps1
```

Ввести клубный код скрыто. Успех staging: CODE_STAGED, это ЕЩЁ НЕ серверная
активация. Не перезапускать службу/ПК и не активировать GUI перед сохранением образа:
первый boot каждого физического ПК должен самостоятельно создать свой ключ.
Образ содержит защищённый код, но не общий private device key. После каждого
freeze boot новый ключ заменяет старый в том же слоте. Администратор образа может
извлечь код; код можно отозвать/ограничить по сроку и числу мест.

## 4. Read-only серверная проверка и освобождение места

PowerShell на ПК владельца → одна SSH-сессия:

```powershell
ssh root@2.56.177.4
```

В shell VPS, root, только чтение:

```bash
/opt/neexitlag/maintenance/veliq-codes list --ca /etc/neexitlag/control/ca.pem --admin-token-file /etc/neexitlag/control/admin.token
```

Подставить публичный ID нужного кода из ответа create/list (НЕ секретный code):

```bash
/opt/neexitlag/maintenance/veliq-codes machines --id CODE_ID --ca /etc/neexitlag/control/ca.pem --admin-token-file /etc/neexitlag/control/admin.token
```

Ответ machines содержит machine_hash, device_id, disabled. occupied seat =
disabled=false. `next_cursor` непустой -> повторить ту же команду с `--after CURSOR`;
не считать одну страницу полным перечнем. При 40/2 слотах обычно одна страница.
Сводка БД без code_hash/секретов (имя БД в текущей установке neexitlag):

```bash
runuser -u postgres -- psql -X -v ON_ERROR_STOP=1 -d neexitlag -c "SELECT c.id,c.kind,c.seat_limit,c.revoked,count(s.machine_hash) FILTER (WHERE NOT s.disabled) AS occupied FROM activation_codes c LEFT JOIN activation_slots s ON s.code_id=c.id GROUP BY c.id ORDER BY c.id;"
```

При повторном freeze/enroll occupied для этого кода не должен увеличиться;
device_id того же machine_hash изменится. Проверить последнюю identity/отзыв:

```bash
runuser -u postgres -- psql -X -v ON_ERROR_STOP=1 -d neexitlag -c "SELECT s.code_id,s.machine_hash,s.device_id,s.disabled,d.revoked FROM activation_slots s JOIN devices d ON d.id=s.device_id ORDER BY s.code_id,s.machine_hash;"
```

ОБЫЧНАЯ ПЕРЕУСТАНОВКА ТОГО ЖЕ ПК: ничего не освобождать вручную.
Для навсегда выведенного/заменённого ПК — отдельная изменяющая команда:

```bash
/opt/neexitlag/maintenance/veliq-codes disable --id CODE_ID --machine MACHINE_HASH --ca /etc/neexitlag/control/ca.pem --admin-token-file /etc/neexitlag/control/admin.token
```

Она освобождает место для ДРУГОГО hardware hash и отзывает действующий device,
но этот hash останется MACHINE_DISABLED для данного кода, включая после переустановки.
В текущем CLI нет enable/reset slot. Не применять disable «на всякий случай» и
не DELETE/UPDATE таблицы вручную. Случайно отключённый ПК потребует другого кода
или отдельной будущей admin-функции; freeze сам блокировку не отменяет.
Повторить machines, проверить disabled=true. Отзыв всей лицензии (изменяет БД,
все её машины теряют разрешение, не использовать вместо освобождения одного места):

```bash
/opt/neexitlag/maintenance/veliq-codes revoke --id CODE_ID --ca /etc/neexitlag/control/ca.pem --admin-token-file /etc/neexitlag/control/admin.token
```

```bash
exit
```

## Подтверждено и границы

Пользователь подтвердил deployment/renewal на VPS, но создание/использование новых
кодов на production пока не проверено. Агент не делал SSH/deployment/enroll на VPS.
Действующий код уже реализует этот путь; runtime изменений для инструкции не нужно.
Тесты/результаты текущей сверки фиксируются в CURRENT.md. Нельзя переносить результаты
fixture на Internet/WinDivert/freeze конкретного ПК без проверки владельцем.

Проверки 21.10.2026: targeted Go race с отдельной PostgreSQL17 — PASS для
core/enrollment, services/control-plane, services/client-service; tools/codes
собран go test (тестовых файлов нет). В том числе 50 freeze replacements/один
слот/отзыв старых keys, HTTPS enroll/lost response, disable/лимит/publication
rollback. HardwareHash stability test — PASS. Обе SQL SELECT выполнены на
временных fixture tables в отдельной PG17; 23 PowerShell блока распарсены,
параметры install.ps1 сверены с реальным ParamBlock; install НЕ запускался.
Собственный PG container остановлен/удалён. Clean-PC GUI/code activation и freeze
на production ещё не проверены владельцем; runtime не менялся.
