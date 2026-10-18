# Edge: отдельная полоса сайтов VELIQ

Реализована отдельная полоса Microsoft Edge: свой IPv4 loopback CONNECT listener,
inline PAC, проверка готовности, recovery-процесс и сохранённое состояние.
Chrome PAC, Windows proxy, DNS Windows, SCM, credentials и сертификаты не меняются.
Сервер VPS не требует обновления.

## Активация и существующие профили

«Оптимизировать» обнаруживает установленный `msedge.exe`. Если профиля Edge ещё нет,
создаётся один custom-профиль с полным путём и начальным DIRECT. AUTO публикуется
только после реальной проверки Edge. Существующий профиль (в том числе старый
OpenAI relay-first, MANUAL, DIRECT, DIRECT+) автоматически не заменяется.
Для существующего DIRECT/AUTO выберите его в «Приложения» → «Сайты Edge AUTO».
Для MANUAL/DIRECT+ сначала явно выберите подходящий режим; эти профили не мигрируют молча.

Явная активация записывает `app_preset: edge-sites` в этот профиль. Без этой метки
старые Edge AUTO/OpenAI профили сохраняют прежний relay-first/fail-closed контракт.
Одновременно поддерживается один Edge site-профиль владельца; второй не заменяет
PAC уже активного профиля. Chrome работает независимо, одновременно с Edge.

Последовательность: согласие закрыть Edge → protected guard → штатный `edge://quit`
(без убийства пользовательского процесса) → recovery ACK → listener/nonce → запись
PAC службой → headless Edge с пустым временным профилем и **без proxy launch flags**
→ подтверждение nonce/политики/PID/SID → AUTO → открытие пользовательской сессии.
Фоновые процессы Startup Boost могут помешать штатному выходу: тогда активация
отказывается продолжать и сообщает причину; чужие процессы не завершаются.

## Привилегии и политика

GUI остаётся обычным пользовательским процессом. Только служба выполняет `edge.policy`
через существующий authenticated owner IPC и пишет:

`HKEY_USERS\<installation owner SID>\Software\Policies\Microsoft\Edge\ProxySettings`

Это JSON REG_SZ с `ProxyMode=pac_script`, inline `ProxyPacUrl` и
`ProxyPacMandatory=true`. HKLM/чужие HKU/HKCU Google Chrome не записываются.
Чужие machine/user proxy policy, отдельные ProxyMode/PacUrl/Server/Bypass значения
и неверный SID приводят к отказу. Произвольные PAC/ключи реестра через IPC не принимаются:
служба сама создаёт онлайн PAC и guard. До записи ведётся ограниченный журнал собственных
значений, чтобы сбой между файловой записью и реестром был восстанавливаемым.

В Edge появится стандартное «Управляется вашей организацией» / “Managed by your
organization”. `edge://policy` должен показывать `ProxySettings` в scope пользователя.
Это управление только proxy-policy Edge; VELIQ не читает содержимое страниц и не
подменяет TLS. Машинная политика другой организации не перезаписывается.

## Маршруты, readiness и диагностика

Используется существующий origin planner и transport: обычные сайты DIRECT-first,
неизвестные сайты — ограниченный AUTO с настоящими TLS/HTTP-пробами; protected OpenAI
только RELAY_REQUIRED. Рабочий локальный DIRECT+ для YouTube применяется только при
имеющейся подтверждённой конфигурации для Edge/сети. Новый профиль такой результат
не выдумывает. Флаг `youtube_via_relay` и логика застоя остаются прежними.
Перенос состоявшегося TCP-туннеля на другой egress не выполняется.

PAC разделяет proxy endpoint по доменам, ограничивая HTTP/2 coalescing между origin.
CONNECT передаёт TLS непрозрачно. HTTPS через PAC использует TCP; это не доказательство
полной native IPv6, UDP/QUIC или WebRTC поддержки Edge. DNS/53 остаётся прежним.
Loopback CONNECT исключён штатным WinDivert `!loopback`; внешние сокеты службы
исключаются существующей проверкой собственного процесса.

`edge.status`: `listening`, `policy_ready`, `ready_profiles`, `address`,
`recovery_address`, `last_error`, `records`, `failures`, `truncated`, `lost`.
Записи содержат реальные PID/creation time/exe/profile, домен, policy/route/reason,
байты, результат открытия. UI отдельно показывает домены и состояние Edge.
Origin HTTP 200 — только проверка origin, не доказательство работоспособности всей страницы.

Состояние службы: `edge-transport.json` рядом с локальными профилями в защищённом
каталоге установки. GUI/recovery: `%LOCALAPPDATA%\NeExitLag\edge-recovery.json`.
Сохраняются свои loopback-порты. Ошибка восстановления Edge диагностируется отдельно
и не останавливает восстановление Chrome. Существующие пути identity/ACL не мигрируют.

## Отказ и откат

«Отключить PAC Edge» устанавливает служебный guard, затем снимает AUTO только с
Edge-профиля. Обычные сайты идут DIRECT, защищённые OpenAI/явные guarded origins и
YouTube без подтверждённого транспорта закрыты. Для обновления политики перезапустите
Edge. Guard намеренно остаётся в реестре: стандартная надпись управления останется.
Полное снятие protected guard не является кнопкой безопасного отката.

При остановке службы recovery-процесс GUI сохраняет ordinary DIRECT только для
разрешённых профилем доменов. Protected домены не получают DIRECT fallback. После
старта службы возвращается тот же listener/PAC без переноса установленных потоков.
При утрате per-user recovery-state (например, после rollback Windows) повторное
«Оптимизировать» заново проверяет и активирует Edge; старый ready-флаг службы сам
по себе не подтверждает новый recovery-процесс.
Recovery требует запущенного GUI; если закрыты и GUI, и служба, онлайн PAC остаётся
fail-closed, а для освобождения обычных сайтов надо запустить службу/GUI и отключить PAC Edge.
Чужую политику VELIQ не удаляет даже при ошибке активации.

## Обновление на ПК владельца

Закройте GUI через трей и браузеры. В обычном PowerShell из репозитория:

```powershell
git status --short
git fetch origin
git switch codex/work
git pull --ff-only origin codex/work
& .\tools\windows\build-e2e.ps1
& .\tools\windows\build-desktop.ps1
```

При ошибке любой команды остановитесь; не запускайте старые binaries дальше.
Elevated PowerShell **под тем же SID владельца установки**:

```powershell
& .\tools\windows\update-service.ps1
Get-Service neexitlag
Get-FileHash .\bin\windows-e2e\neexitlag-service.exe
Get-FileHash "$env:ProgramData\neexitlag\neexitlag-service.exe"
```

Служба должна быть Running, SHA256 должны совпасть. В обычном PowerShell:

```powershell
& .\tools\windows\install-desktop.ps1
```

Откройте VELIQ, нажмите «Оптимизировать»; существующий Edge-профиль активируйте явно
кнопкой «Сайты Edge AUTO». Подтвердите штатный перезапуск Edge. В `edge://policy`
проверьте Microsoft Edge ProxySettings; настройки Chrome должны остаться прежними.

С VPN полностью OFF запустите существующую headless-диагностику обоих браузеров:

```powershell
& .\tools\windows\diag-no-vpn.ps1 -Only browser
```

Она открывает **изолированные пустые профили** Chrome и Edge для claude.ai, chatgpt.com,
YouTube, Telegram Web, Discord. Не коммитьте `diag-result.txt`. `PAGE LOADED`/DOM/title
не доказывают проигрывание видео, авторизацию, все API/WebSocket или отсутствие утечек.
Сопоставьте результат со свежими `edge.status` records; `ERR_*`, EMPTY PAGE и TIMEOUT
считать отказом, а не PASS. Для Edge IPC из обычного PowerShell:

```powershell
. .\tools\windows\common.ps1
Invoke-ServiceIPC 'edge.status' | ConvertTo-Json -Depth 12
```

Не запускать VPS/update-vps/setup-e2e/enroll/install/uninstall для этого обновления.

## Автоматические проверки

- Go race: DNS/bootstrap, legacy Edge semantics, separate Chrome/Edge state,
  refusal of unproved AUTO, listener failure, guard journal, guard-only restart.
- C#: discovery/dedup/full path, profile preservation, explicit marker, false ACK,
  guard failure before profile mutation, unchanged Chrome contract, no fake page PASS.
- `.github/workflows/edge-native.yml`: C# core tests, WPF build, actual Edge policy,
  девять simultaneous HTTPS/HTTP2 fixture-сайтов, реальный mTLS relay, protected
  source check, ordinary DIRECT recovery и service restart без замены браузера;
  отдельный отказ для чужой proxy policy и неверного SID.
- Existing Chrome tests/workflow остаются включены. Native Edge DIRECT+ packet
  transform и пользовательский ISP не объявляются проверенными этой фикстурой.

В Linux-окружении Codex .NET/Edge отсутствуют: C# локально не компилировался,
actual Windows acceptance здесь не выполнялся. CI должен подтвердить их отдельно.
