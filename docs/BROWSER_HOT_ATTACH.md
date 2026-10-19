# Браузеры без перезапуска: поэтапная проверка

## Этап (а): регистрация профилей, не активация маршрутов

После первого успешного подключения GUI к службе каталог установленных браузеров
создаёт отсутствующие профили автоматически. Не нужно нажимать «Оптимизировать»
или браузерные кнопки. Chrome получает DIRECT до обычной PAC-активации; Edge —
AUTO с `app_preset=edge-sites-pending`. Это **не** подтверждение готовности AUTO.
Уже существующие профили, режимы, правила, DIRECT+ и полные пути сохраняются.

GUI ничего не записывает в политику браузера в рамках регистрации, не запускает
и не закрывает браузер. Старые кнопки PAC с диалогом/штатным перезапуском пока
сохранены по решению владельца; на этом этапе проверять нужно только регистрацию.
Если служба уже оптимизирует трафик, создание отсутствующего Edge откладывается
(`DEFERRED / ACTIVE_CAPTURE_PENDING_EDGE`): pending AUTO иначе немедленно блокировал
бы его новые внешние соединения. После выключения оптимизации регистрация повторяется.
Отказ дополнительного IPC не очищает подтверждённую таблицу приложений. Повторная
проверка после отказа выполняется не чаще раза в 10 с; после успеха — один раз за запуск GUI.

В «Диагностике» есть сводка регистрации. Ограниченный отчёт без секретов:
`%LOCALAPPDATA%\NeExitLag\browser-startup.json`. В нём время начала/окончания,
длительность, браузер, полный путь, профиль, режим, состояние и безопасный код отказа.
Содержимое ошибки IPC, токены, сертификаты и HTTPS payload туда не записываются.
Отчёт относится к текущему пользователю; проверяйте время, чтобы отличить старый GUI.

## Короткая проверка этапа (а), VPN OFF

Оставьте Chrome и Edge открытыми. Закройте **только VELIQ через «Выход» в трее**.
Если служба уже из `b358970` или новее и `edge.status.activation_version >= 3`,
обновите только GUI (служба и VPS на этом этапе не менялись), в обычном PowerShell:

```powershell
git switch codex/work
if ($LASTEXITCODE) { throw 'Git switch failed' }
git pull --ff-only origin codex/work
if ($LASTEXITCODE) { throw 'Git pull failed' }
git merge-base --is-ancestor 84793bf HEAD
if ($LASTEXITCODE) { throw 'Expected stage-a code/diagnostics not present' }
& .\tools\windows\build-desktop.ps1
if ($LASTEXITCODE) { throw 'GUI build failed' }
& .\tools\windows\install-desktop.ps1
Start-Process "$env:LOCALAPPDATA\Programs\VELIQ\VELIQ.exe"
```

У старой службы `3fec1ae` Edge capability может быть ниже 3: новый GUI честно покажет
`EDGE_IPC_UNSUPPORTED`, а не создаст нераспознаваемый AUTO. Если служба ещё не обновлялась,
после `git pull` сначала выполните в обычном PowerShell `& .\tools\windows\build-e2e.ps1`,
затем в повышенном PowerShell **под тем же SID владельца**, из того же репозитория:

```powershell
& .\tools\windows\update-service.ps1
$built=(Get-FileHash .\bin\windows-e2e\neexitlag-service.exe).Hash
$running=(Get-FileHash "$env:ProgramData\neexitlag\neexitlag-service.exe").Hash
if ($built -ne $running -or (Get-Service neexitlag).Status -ne 'Running') { throw 'Service update verification failed' }
```

После этого выполните GUI-only команды выше в обычном окне. Закрывать браузеры
не требуется; при обновлении службы возможен разрыв сетевых соединений, это не
измерение hot attach. Не запускать setup/install/uninstall/update-vps и не переобучать identity.

На чистой установке с выключенной оптимизацией через несколько секунд должны
появиться обе строки, Edge AUTO pending. **Не удаляйте рабочие профили ради теста.**
Существующий Edge DIRECT не мигрирует автоматически: выбранный ранее режим сохраняется.
Не нажимайте PAC-кнопки для проверки этапа (а): их старый сценарий ещё перезапускает браузер.

```powershell
& .\tools\windows\diag-no-vpn.ps1 -Only browser,youtube
```

Диагностика читает профили из IPC, отчёт GUI, PID и creation time текущей Windows-сессии.
Проверяет сохранение уже открытых процессов за интервал диагностики. Её headless
пробы используют отдельные временные профили; это не доказательство подхвата PAC
открытым окном или полноценной работы сайта. PID дочернего renderer может естественно
исчезнуть; это отдельно от перезапуска основного browser process. Не коммитить diag-result.txt.

## Этап (б): новые TCP, owned PAC остаётся crash-guard

`browser.native.enable {profile_id}` готовит IPv4/TCP capture **до** публикации AUTO,
требует существующий service-owned mandatory PAC того же owner SID. Fresh без guard
получает `NATIVE_OWNED_PAC_GUARD_REQUIRED`, без изменения профиля/политики/браузера.
Это решение владельца: после crash службы WinDivert снимает фильтры. Native-only
не может обещать protected fail-closed при crash; устранение зависимости — отдельный этап.

GUI автоматически пробует этот IPC для Chrome AUTO, созданного GUI pending Chrome и
Edge AUTO / DIRECT `edge-sites-pending`, когда включена оптимизация, и перед Optimize.
Ручной MANUAL/DIRECT+/не-pending DIRECT не меняется. Если guard есть, Edge pending становится
AUTO/native без кнопки и browser quit. Неподдерживаемая/старая служба и отказ capture/guard
не очищают таблицу. Повтор ограничен 30 с. Старые PAC-кнопки, диалоги и quit-путь сохранены;
это откат, удаление которого ещё не разрешено (этап в).

Новый native TCP принимается локально, ClientHello/Host читается bounded worker (32 KiB,
2 с); до решения нет upstream данных. Видимый достоверный SNI/Host использует существующий
TLS/certificate/HTTP GET planner, DNS recovery и сетевой/профильный кэш. ECH outer SNI,
неполный ClientHello и отсутствие имени не считаются доменом: только authenticated
RELAY_REQUIRED к исходному адресу, **никакого DIRECT fallback**. DNS/53 остаётся прежним.
Метаданные/IP/shared-IP не заменяют истинное имя. HTTPS не расшифровывается.

Native enable не вызывает reset/reconfigure/hot attach старых TCP. Уже открытый HTTP/2
может пережить обновление страницы: новый маршрут действует только после появления
нового соединения. Для initial DIRECT silence после отправки данных: через 2 с закрывается
только этот flow, в кэш идёт отказ (YouTube-family сразу TTL 10 мин); никакого replay или
переноса установленного TLS. Браузер может повторить запрос новым соединением. Байт ответа
не очищает verdict и не означает страницу PASS.

`browser.native.status`: capture/Optimize, actual origin/opaque, PID/creation time, policy,
ошибки, actual payload bytes, enable→first upstream open и first bidirectional exchange.
Последние — **не HTTP(S)/page PASS**. Учет максимум 512 flows / 128 activation clocks,
ответ до 60 KiB, усечения/lost видны. Общий native+PAC relay usage: active/opened TCP
и минуты по profile/domain, payload bytes без QUIC overhead и origin probes. Opaque
идёт в `<opaque>`, истинный домен не выдумывается. Googlevideo с известным именем —
общий hard limit 8, очередь 32 / до 2 с, затем admission error без мёртвого DIRECT;
разрешение «потом relay» не делает лимит бесконечным. Opaque видео нельзя достоверно
отнести к googlevideo; оно дополнительно ограничено общим relay pool 8×48. VPS не меняется.

Edge при nonce failure не выдаёт PAC READY: native использует реальный bound listener
и owned fail-closed PAC, либо статический guard (ordinary DIRECT, protected закрыты),
если listener не поднимается. Чужая политика не заменяется. Chrome journal/онлайн PAC
не меняются. Старый owner recovery-процесс восстанавливается из имеющегося состояния.
GUI запрашивает `RefreshPolicy(false)` под owner user, не SYSTEM/machine: Windows Chromium
смотрит user GP notifications. API асинхронный; `REQUESTED` **не доказывает** live uptake.
Отчёт: `%LOCALAPPDATA%\NeExitLag\browser-native-policy-refresh.txt`. Если уже открытый
Edge держит старый мёртвый PAC, отсутствие actual exchange остаётся видимым отказом;
не утверждаем автоматическое исправление этого случая до Windows-проверки владельца.

IPv6 data routing и UDP/443 политика не меняются. IPv6 guard сохраняет fail-closed,
даже когда new-only исключает native TCP из сброса preexisting flows. Не обещаем
поддержку IPv6/QUIC direct/обход coalescing на уровне домена без достоверного имени.

### Обновление и короткая проверка без закрытия Chrome/Edge

Закройте **только GUI** через «Выход» в трее. Сохраните уже установленные owned PAC,
не удаляйте профили или реестр. Обновление службы само разрывает принадлежащие ей sockets;
это отличается от native enable при уже работающей службе и не доказывает continuity.
Из elevated PowerShell **того же SID владельца установки**, в `C:\neexitlag`:

```powershell
$ErrorActionPreference='Stop'
git switch codex/work
if ($LASTEXITCODE) { throw 'git switch failed' }
git pull --ff-only origin codex/work
if ($LASTEXITCODE) { throw 'git pull failed' }
& .\tools\windows\build-e2e.ps1
& .\tools\windows\build-desktop.ps1
& .\tools\windows\update-service.ps1
if ((Get-FileHash .\bin\windows-e2e\neexitlag-service.exe).Hash -ne (Get-FileHash "$env:ProgramData\neexitlag\neexitlag-service.exe").Hash) { throw 'Service hash mismatch' }
Get-Service neexitlag | Select-Object Status
```

Далее обычный owner PowerShell: `& .\tools\windows\install-desktop.ps1`, открыть
`%LOCALAPPDATA%\Programs\VELIQ\VELIQ.exe`, нажать Optimize без PAC-кнопок/закрытия
браузеров. В открытом окне обновить обычную страницу, YouTube и ChatGPT; затем:

```powershell
& .\tools\windows\diag-no-vpn.ps1 -Only browser,youtube
```

Скрипт 2026-10-19.3 сохраняет PID/creation time до/после, native decisions/timing,
relay usage и отдельно headless controls. Не путать headless успех со старым окном.
Существующий PAC может обслуживать запросы без native-record: это совместимый PAC путь,
не native interception PASS. Прислать отчет, **не коммитить diag-result.txt**.

Локально .NET/Windows отсутствуют: C#/WPF, реальный Chrome с owned policy и unchanged
PID/creation проверяются CI, native deferred TCP — Windows socket/mTLS/race burst.
Это не Internet PASS и не доказательство живого policy refresh на ПК владельца.
