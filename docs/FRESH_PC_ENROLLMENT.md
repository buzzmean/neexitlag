# VELIQ: пустой Windows-ПК после freeze

## Границы и факты аудита

Работа подготовлена для **существующего 2.56.177.4**, одного ПК владельца.
**Ни SSH, ни install/enroll/provision, ни миграция на production не выполнялись.**
`veliq-swe-01`, PR #8, production БД/ключи/сервисы этой разработкой не изменены.
Скрипты явно отказывают на hostname `veliq-swe-01` и при несовпадении VPS identity.

- `setup-e2e.ps1` создаёт private bundle **на ПК**, запускает `vps-provision`,
  требует совпадения CA существующего VPS и использует bootstrap admin token.
  Для потерянного bundle это неправильная точка входа. Новый bootstrap его не вызывает.
- `vps-provision` создаёт CA/server key, config-signing key, токены и identity.
  `install-backend.sh` копировал CA key также в `/etc/neexitlag/control/ca.key`,
  с правами root:neexitlag-control 0640. Поэтому потеря ПК **сама по себе не
  доказывает потерю CA на VPS**. Его наличие/сроки на production здесь не проверены.
  Ротация подготовлена по явному требованию владельца, не выполняется автоматически.
- `update-vps.ps1` обновляет бинарники с сохранением identity, требует старые
  credentials для live proof; он не восстанавливает утраченную identity/CA.
- `install.ps1` сохраняет SCM `neexitlag`, IPC, owner SID и закрытые ACL. Он
  принимает уже выданные device/trust files. `reinstall-state.ps1` намеренно
  запрещает молча заменить сохранённую identity другим bundle. Эти проверки
  новый bootstrap не обходит; существующая установка сохраняется и отклоняется.
- CP `/api/v1/devices` требует admin token, принимает public key, не CSR.
  `/auth` и certificate renewal проверяют signed device proof и nonce replay.
  Новый root-SSH helper проверяет подпись CSR, регистрирует тот же Ed25519 public
  key и выдаёт совместимый client certificate. Публичный API, ticket wire,
  mTLS, lease renewal и signing identity CP не меняются.

## Где живут ключи и как появляется доверие

На VPS: CA private key **root:root 0600**, `/etc/neexitlag/control/ca.key`;
резервная копия только в root-only каталоге VPS. На ПК CA private key никогда
не создаётся и не передаётся. На ПК генерируются случайные device ID и Ed25519
key, затем подписанный CSR. Приватный device key не покидает ПК.

CP по-прежнему работает непривилегированным. Миграция добавляет systemd
`LoadCredential` и `-ca-key %d/veliq-device-ca-key`: файл на диске остаётся 0600
root-only; systemd даёт **только CP** runtime-копию в изолированном credentials
каталоге на время работы службы. Это не утверждение «CA key недоступен памяти
CP»: текущий CP сам подписывает client certificate/renewal. Для запрета доступа
CP даже к runtime-копии нужен отдельный root signing service, это другой этап.
Ни root private CA, ни CP signing key/admin token/DB password на ПК не передаются.

Публичный CA проверяется по **SHA256 DER certificate**, не по тексту PEM.
Есть два варианта pin:

1. `-CaFingerprint <64 hex>` — ожидаемый fingerprint из доверенного канала владельца;
   допускаются двоеточия и верхний регистр. Несовпадение проверяется на VPS до enrollment transaction и повторно на ПК до
   trusted writes; установка прерывается.
2. Без параметра pin передаётся в enrollment capsule по **аутентифицированному SSH**,
   после проверки host key владельцем. Затем он явно передаётся локальному
   `fresh-pc complete --ca-sha256 ...`; certificate bytes сверяются до любого
   HTTPS/mTLS или установки. Это SSH как начальный источник доверия, а не pin,
   скачанный с ещё недоверенного HTTPS endpoint.

OpenSSH показывает host-key/password prompts в консоли; `StrictHostKeyChecking=ask`,
не `no`. На новом ПК host key нужно сверить с доверенной консолью VPS/записью
владельца; **пароль не заменяет проверку сервера**. `known_hosts` хранится в
репозитории, после freeze подтверждается снова. Секреты не передаются через argv,
git, stdout-логи или transcript. SSH protocol stdout захватывается только в
закрытый `.e2e/fresh-pc/response.json` (это credential file, не лог); файл удаляется
в finally. В нём есть echo token для существующего authenticated probe, не admin token.

## Повторная регистрация и hardware ID

SMBIOS UUID нормализуется и хешируется SHA256 с фиксированным префиксом. Пустые/
нулевые UUID отклоняются; UUID — inventory slot, **не секрет и не авторизация**.
Право заменить слот даёт вход владельца по SSH root (или пользователь с
`sudo -n` для root helper). Общего ключа для ПК нет, предсказуемых ключей нет.

На VPS additive таблица `fresh_machine_slots` содержит current device ID для
hardware hash. Fresh install после freeze создаёт **новую** identity. Helper в
одной PostgreSQL transaction регистрирует новую identity, отзывает предыдущую
в этом слоте и публикует полный подписанный config со всеми revoked IDs.
При DB/config-size/publication ошибке откатывается вся замена. Старый device не
остаётся «воскрешённым». Существующий CP binary можно оставить: schema additive,
таблица существующих devices и формат cert/proof/tickets прежние.

Повтор той же заявки после потерянного SSH-ответа использует тот же локальный
CSR/ключ/ID и не создаёт вторую device row. Ранее отозванная заявка не оживляет
identity. Новые UUID/разные физические ПК имеют отдельные слоты и отзыв.
Первый fresh-slot не может автоматически сопоставить legacy device IDs с hardware,
потому что прежний enroll UUID не хранил. Legacy IDs при необходимости отзываются
существующим owner DELETE `/api/v1/devices?id=...` с VPS; admin token на ПК для этого
не копируется. Сама ротация CA не отзывает записи БД: сохранённый старый device key
может запросить новый cert, пока его ID не revoked. Это не скрывается как «все старые
устройства отозваны». Последующие fresh-slot замены отзывают предшественника автоматически.

История revoked device rows остаётся для аудита и безопасности; бесконечная
retention/GC не реализована. Config ограничен 1 MiB: превышение даёт явный FAIL
без обрезки списка. Для многолетней эксплуатации нужен отдельный retention policy.
На клонах с одинаковым SMBIOS UUID слоты столкнутся — до клуба необходимы
уникальные UUID каждого ПК. Автоматически отличить такие клоны без стабильного
секрета/уникального hardware ID невозможно.

## Один раз: явная миграция CA владельцем

**Не запускать случайно/как часть обычного обновления.** Старые TLS clients после
ротации перестанут доверять серверу. Сейчас VELIQ только на одном пустом ПК;
это maintenance с разрывом старых соединений, не seamless CA renewal.

Подготовленный `deploy/vps/rotate-fresh-pc-ca.py` выполняет:

Для безопасного rollback должны существовать все прежние TLS files, включая
VPS-копию старого CA key. Если ключ потерян **и на VPS**, корректный откат CP после
рестарта без него невозможен: инструмент откажет, нужен отдельный forward-only
план с явным разрешением простоя. Содержимое памяти CP не извлекается.

1. Проверку root/hostname/IP, активных CP/relay/echo, systemd >=247, поддерживаемого
   CP unit/paths/key-id, binary capabilities `/3`/1800s/полные revocations до 1 MiB, ожидаемого старого CA pin и SHA256 helper binary.
2. Генерацию **на VPS** нового root CA и server key/cert с IP SAN VPS+127.0.0.1.
3. Постоянную `/etc/neexitlag.bak-<UTC>-<random>/`: весь config, старые units,
   ownership/modes и SHA256 manifest; каталог 0700, CA backup key 0600.
4. Независимый systemd rollback timer 180 с, остановку CP/relay, атомарные замены
   семи TLS файлов при остановленных readers и CP LoadCredential drop-in.
   CP signing key, admin/heartbeat/echo tokens, DB credentials, relay replay state,
   profiles и firewall не меняются; `/etc/neexitlag` не удаляется/не пересоздаётся.
5. Запуск CP→relay, bounded проверку CA-verified HTTPS8443/healthz, local relay health,
   реальный QUIC4433 mTLS `/3` с fresh signed-proof ticket, TCP+UDP authenticated
   echo через relay. Короткая probe identity отзывается и отзыв публикуется.
   Production lease проверяется как 1800 с, ожидать 30 минут не требуется.
6. Только после всех проверок — commit marker и отмена timer. При исключении —
   восстановление TLS файлов/предыдущего drop-in/modes и старта старых служб.
   Потеря SSH покрыта timer; БД и nonce/replay state **не откатываются**.

Timer переживает потерю SSH, **не внезапную перезагрузку самого VPS**. При power
loss между replace и commit нужен root rollback из постоянной копии; несовпавшие
TLS credentials не дают обхода mTLS, службы остаются неисправны/fail-closed.
Автоматический boot recovery — будущий hardening, не заявляется реализованным.
Локальная копия на том же VPS не является резервной копией от потери его диска.

После отдельного решения владельца migration запускается с Windows **обычным
PowerShell** из `C:\neexitlag` (одна SSH-сессия; пароль вводится интерактивно):

```powershell
& .\tools\windows\bootstrap-fresh-pc.ps1 -VpsIP 2.56.177.4 -SSHUser root -SSHPort 22 -RotateCA -ConfirmCARotation 'ROTATE-CA-2.56.177.4'
```

Можно дополнительно передать `-ExpectedCurrentCAFingerprint <старый DER SHA256>`.
Если его нет, старый pin фиксируется внутри доверенной SSH-сессии перед replace.
Новый public fingerprint и путь backup выводятся, private material не выводится.
Server/client leaf выдаются на 30 дней, CA на 365 дней. Client cert renewal уже
существует в CP; автоматическое продление **server** cert этим bootstrap не добавлено.
До его истечения нужна отдельная maintenance-процедура на VPS. Это не обещание
бессрочной работы без обслуживания сервера.
Флаг `-RotateCA` **никогда** не включается из-за ошибки обычного enrollment.
Повторная ротация уже переведённого на credential unit сервера намеренно отвергается;
инструмент предназначен для этой однократной миграции, не annual CA lifecycle.

Явный rollback на VPS (root, точный путь из вывода; не выполнять здесь):

```bash
python3 /etc/neexitlag.bak-<точный-UTC-random>/rollback.py rollback \
  --backup /etc/neexitlag.bak-<точный-UTC-random> --confirm ROLLBACK-CA-2.56.177.4
```

## Каждый пустой ПК/каждая перезагрузка после freeze

На VPS нужны Linux amd64, Python >=3.9 и systemd >=247. Default enrollment
также проверяет `-version-json` и не обновляет старые binary автоматически.
Предполагаются Git, совместимый Go >=1.25 (у владельца 1.27), **.NET SDK 8**
и Windows OpenSSH Client. SDK проверяется через `dotnet --list-sdks`, не по наличию
`dotnet.exe`. Источник NuGet задаётся приватным repo-local RestoreConfigFile,
глобальные пользовательские NuGet sources не переписываются. Корпоративные
AllSigned/ExecutionPolicy не отключаются; запрещённое политикой выполнение
скриптов требует разрешения IT/подписанной поставки, не автоматического bypass. WinDivert сохраняет
существующие pinned SHA256. Новых библиотек/версий драйвера нет.

PowerShell владельца (обычный при включённом UAC; при проверенном выключенном UAC
допускается его обычный для этого ПК full-admin token, см. режимы ниже):

```powershell
# Клонирование и переход выполняются один раз после каждого freeze.
$repo = Join-Path $env:USERPROFILE 'neexitlag'
git clone --branch codex/work https://github.com/buzzmean/neexitlag.git $repo
Set-Location $repo
# Единственная команда установки после подготовленной migration VPS:
& .\tools\windows\bootstrap-fresh-pc.ps1 -VpsIP 2.56.177.4 -SSHUser root -SSHPort 22
```

Bootstrap строит backend/GUI и Windows/Linux enrollment helper, получает CSR
response за одну SSH-сессию, проверяет CA pin, cert/key/EKU и signed ticket.
**С самого ПК** проверяет CP8443 и real relay TCP/UDP через4433 до установки.
При включённом UAC эскалация под **тем же SID** устанавливает только службу и существующее owned
app-scoped native TCP firewall admission rule. На VPS firewall не меняется.
Elevated-этап сверяет закрытый service.json и хеши скопированных credentials,
передаёт наружу только public proof. Обычный процесс не читает manifest/identity
службы. Проверенные приватные staging-копии identity и echo token удаляются;
установленная identity остаётся SYSTEM/admin-only. Startup ожидание проверяет connected и свежий network snapshot (optional error
читается безопасно), не ждёт quality.health. После UAC обычный процесс ставит/запускает GUI, создаёт browser profiles и повторно
использует прежнюю PAC/recovery/owner-nonce транзакцию Edge, затем Chrome,
проверяет ReadyProfiles и включает Optimize. При включённом UAC GUI не получает elevation.
При доказанно выключенном UAC тот же шаг службы выполняется в текущем процессе;
GUI наследует полные права, как описано ниже.

Это одна команда, но не «zero interaction»: SSH host key/password, при включённом UAC — UAC prompt, а при уже
открытых браузерах могут появиться **существующие** диалоги PAC/restart. Их удаление
(этап в) этой задачей не разрешено и не выполнено. Отмена диалога/неподтверждённый
PAC дают FAIL, а не фиктивный готовый браузер. Чужая policy не перезаписывается;
старый crash-guard остаётся. Случайный SDK/SSH/CA/driver/network/SCM/GUI error не
выдаёт PASS и не вызывает автоматическую ротацию сервера/удаление credentials.

Все входные/staging файлы — под репозиторием (`.e2e/fresh-pc`, `.e2e/windivert`,
`bin`); не нужны bundle/USB/файлы из предыдущей загрузки. Обычные **выходы установки**
в `%ProgramData%\neexitlag`, `%LOCALAPPDATA%\Programs\VELIQ`, shortcut и служба,
а также стандартные caches Go/.NET, естественно создаются вне repo. Если запрет
«никаких файлов вне repo» относится и к этим outputs, Windows service/GUI install
с ним несовместим — это не portable-режим.

SCM/service directory существующей установки не затираются. Для неё требуется
штатный updater/reinstall с прежней identity; fresh bootstrap не служит способом
обойти `reinstall-state.ps1`. При частичной неудачной установке credentials
сохраняются для восстановления, а не удаляются «для чистоты».

## Режимы UAC и права владельца

Bootstrap **только читает** 64-bit HKLM `Policies\System`: DWORD `EnableLUA`
и `ConsentPromptBehaviorAdmin`. Дополнительно через `TOKEN_QUERY` читает
`TokenElevationType`, `TokenElevation`, реальное наличие `TokenLinkedToken`
и `TokenIsRestricted` (класс 40). Проверка не включает привилегии, не меняет токен, UAC,
реестр или ACL установленных файлов. Ошибка/отсутствие обязательных данных — отказ,
не повод считать UAC выключенным. Исключение: для независимо подтверждённого
`TokenElevationTypeDefault=1` отсутствие linked handle или ошибка самого
`TokenLinkedToken` — штатное состояние non-split token. Win32 error сохраняется
в диагностике (не ограничивается кодом 1312); ошибки elevation/restricted/registry
этим исключением не скрываются. Для Full/Limited ошибка linked-query остаётся отказом.

| ПК / токен | Поведение |
|---|---|
| UAC включён, обычный owner или limited split token | `UAC_ENABLED_OWNER`: прежний путь, отдельный elevated ServicePhase после проверки CA/relay; GUI остаётся обычным |
| UAC включён, PowerShell elevated (включая `ConsentPromptBehaviorAdmin=0`, «не спрашивать») | Отказ до build/SSH/enroll. Запустить обычный owner PowerShell |
| UAC выключен (`EnableLUA=0`), Administrators, `TokenElevation=true`, тип Default, linked token отсутствует/его запрос недоступен, `TokenIsRestricted=0` | `UAC_DISABLED_FULL_ADMIN`: служба ставится **в этом же процессе**, без `-Verb RunAs` и без дочернего ServicePhase. GUI запускается с полными правами, **без понижения** |
| Restricted (`runas /trustlevel`), неизвестная/противоречивая policy/token комбинация, либо UAC-off без полного admin token | Отказ; скрипт не пытается чинить токен или обходить запрет |

Режим и безопасные policy/token поля выводятся перед build/SSH. В full-admin
режиме отдельное предупреждение явно сообщает, что и GUI имеет права
администратора. Это узкое исключение из правила unelevated GUI, **явно разрешённое
владельцем для UAC-off клубных ПК**, а не изменение поведения обычных ПК.
Выключать UAC ради VELIQ не нужно и скрипт этого не делает. Изменённое `EnableLUA`
при ещё старом split token (до перезагрузки) не считается подходящим режимом.

При `UAC_TOKEN_PREFLIGHT_REJECTED` выводятся только разрешённые безопасные поля:
`reason` с конкретным не прошедшим условием; `PolicyView`, оба policy value и их
registry kinds; `IsAdministrator`, `ElevationType`, `Elevated`, `Restricted`;
результат linked-query (`ABSENT/PRESENT/ERROR/NOT_RUN`), `HasLinkedToken`, Win32
`LinkedTokenError` и независимые query error codes. Неизвестные значения остаются
`unknown`, а не превращаются в false/0. Частичный отказ чтения реестра не мешает
попытке собрать token evidence и наоборот. Сырые exception/окружение/identity не
сериализуются. Для повторного запуска в той же PS-сессии новый adapter имеет
отдельное имя `FreshPCTokenV3`, старые Add-Type с pointer-sized DWORD ABI не используются.

Классы `TokenElevationType=18`, `TokenElevation=20` (структура из одного DWORD)
и `TokenIsRestricted=40` читаются отдельной P/Invoke-сигнатурой с `out int`,
`TokenInformationLength=sizeof(int)=4` и `out int ReturnLength`. Размер указателя
не используется для DWORD даже на x64. Класс `TokenLinkedToken=19` использует
другую сигнатуру с `out TokenLinkedToken` и размером собственной структуры
(поле `IntPtr`, 8 байт на x64). Успешный DWORD-ответ также требует ReturnLength=4.

Если GetTokenInformation отказал, к snapshot добавляется запись вида
`call=GetTokenInformation; class=TokenElevation(20); error=24; buffer_length=4`.
Это пример формата, не реальный вывод новой сборки с ПК владельца. Ошибка
`ERROR_BAD_LENGTH=24` никогда не даёт подтверждённого elevation/type/restriction:
соответствующее значение остаётся unknown, preflight отклоняется. Она не является
штатным исключением linked-query. При Default и linked error 1312 отсутствие
linked token нормализуется в `HasLinkedToken=False`, но исходные `ERROR/1312`
остаются в диагностике, не заменяются успешным запросом.

Локальная read-only проверка без bootstrap/SSH/install/enroll (из repo):

```powershell
. .\tools\windows\fresh-pc-execution.ps1
$e = Get-FreshExecutionEvidence
Format-FreshExecutionEvidence $e 'READ_ONLY'
Resolve-FreshExecutionMode $e
```

Например, `ENABLE_LUA_NOT_DWORD` означает неверный registry kind,
`RESTRICTED_TOKEN_REFUSED` — ограниченный токен, `UAC_ENABLED_ELEVATED_REFUSED` —
полный admin token при включённом UAC; `TOKEN_RESTRICTION_UNAVAILABLE` — не удалось
доказать отсутствие restrictions. Строка `LinkedTokenQuery=ERROR` сама по себе
не является отказом при Default. Реальный Win32-код прежнего отказа владельца
из старого generic сообщения восстановить нельзя; новый вывод его покажет.

В обоих режимах OwnerSid берётся из текущей Windows identity. Общий installer
повторно проверяет SAME SID и admin token, сохраняет прежние проверки `service.json`,
хешей identity/certificate/echo token и service binary. `install.ps1`, ProgramData
SYSTEM/admin-only ACL и IPC owner-check не меняются; staging ACL по-прежнему
SYSTEM/admin/owner. На UAC-off ПК защита от самого пользователя-администратора
ограничена правами этого пользователя в Windows; скрипт не выдаёт дополнительных
прав и не делает credentials доступными всем.

Скрипты определяют repo по `$PSScriptRoot`, не по текущей папке и не по `C:\neexitlag`.
Поддерживается, например, `C:\Users\<user>\neexitlag`, включая пробелы в имени:
пример клонирования выше использует `$env:USERPROFILE`, создание каталога в корне
C: не требуется. Стандартные install outputs в ProgramData/LocalAppData остаются.

Тесты покрывают матрицу UAC/token (включая consent=0 с включённым UAC), restricted
и missing evidence, выбор ровно одного пути установки с тем же SID, quoting пути
профиля. На Windows CI добавлено настоящее **read-only** чтение policy/token в
PS5.1 и PS7; тесты не изменяют UAC и не выполняют installer/enroll/SSH. Изолированная
матрица не заменяет запуск полного bootstrap на реальном UAC-off ПК владельца.

## Результат и проверка

Финальный PASS требует authenticated service (не demo), backend.ready, проверенных
CP/mTLS/ticket/реальных TCP+UDP echo и actual browser PAC nonce, а не только Running
или TCP ACK. Он **не доказывает**, что Telegram/YouTube/Claude открываются у ISP.
После установки можно выполнить обычную короткую диагностику:

```powershell
& .\tools\windows\diag-no-vpn.ps1 -Only browser,youtube
```

Автотесты: CSR/signature/key/EKU/CA pin и отказ до trusted writes; PostgreSQL
retry/new identity/revocation/atomic publication rollback; Python transactional
CA restore, backup corruption, late watchdog, unit parsing; PS5.1/7 parsing,
StrictMode/status failures/quoting; native C# argument compatibility checks.
Production SSH/password console, systemd credential/rotation на данном VPS и
полный fresh-PC Windows install остаются **непроверенными на боевом окружении**.
Ни один тест не выполняет provision/install/SSH на production.


Локальная проверка разработки: полный `go test -race ./...` с отдельным
PostgreSQL 17, `go vet ./...`, Linux/Windows amd64 build — PASS; 28 Python-тестов
(7 новых fresh-PC), 130 C# core tests, PowerShell 7.4.6 checks — PASS.
WPF и native test host кросс-собраны SDK 8.0.426; native Windows execution и
PowerShell 5.1 локально не выполнялись (добавлены в существующий CI).
Локальный authenticated TLS/mTLS `/3` TCP+UDP fixture — не Internet/VPS PASS.


Регрессия DWORD ABI (после 34e042d): локальные проверки компилируют настоящий
native adapter и через reflection проверяют out Int32, layout linked-token и
размеры; вызывают общий result validator для failed/error=24 и неверного
ReturnLength. Exact owner fixture с elevation=1 и linked error=1312 принимается;
error=24 для любого из обязательных DWORD-запросов отклоняется даже при stale value.
Настоящие Win32-вызовы исполняются только в Windows CI/на ПК; Linux-тесты не
выдаются за Windows runtime PASS.
