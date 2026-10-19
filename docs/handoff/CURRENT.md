# CURRENT — состояние проекта (обновляется в конце каждой сессии)

Обновлено: 19.10.2026, Codex; рабочая ветка `codex/work`. История передачи Claude сохранена ниже.

## (a) Цель
Чистый Windows 10/11: установил, нажал «Оптимизировать» — заблокированные в РФ сервисы работают без VPN: ChatGPT, Claude, YouTube, Discord, Telegram, браузеры Chrome и Edge. Первый заказчик — клуб на 36 ПК с откатом при перезагрузке.

## (b) Состояние на 19.10.2026
- Владелец подтвердил Chrome/YouTube на Windows build `3fec1ae`; Edge на нём не проходил readiness.
- Причина Edge найдена в коде: guard удалял nonce, выданный предыдущим prepare.
  Новый nonce выдаётся после последнего guard и штатного выхода Edge; служба требует настоящий owner GET.
- Код `983d694`, `6c075ae`, `984c1a5`, `1e90c7f`, `6a64f14`; CI-фикстуры `0aeebcb`.
- VPS, credentials, сертификаты и PR #8 не менялись. QUIC/IPv6 routing не менялись.

## (c) Что подтверждено
- Windows/ISP владельца: Chrome и YouTube на `3fec1ae`. Исправленный Edge на ISP ещё не проверен.
- Linux: полный Go race, build/vet, Windows amd64 cross-build/vet и compile native test binary.
- Изолированный PostgreSQL: control-plane race-тесты, включая 5000 revocations; production код CP не менялся.
- CI `6a64f14`: C# desktop tests, WPF build и actual Edge policy/HTTPS/nonce/outage/restart прошли.
  Это контролируемые Windows origins, не Internet PASS. Последующие CI проверяют фикстуры.

## (d) Открыто
- **Полное удаление ProxySettings при отказе не реализовано:** без PAC нельзя гарантировать
  protected no-DIRECT для подресурсов/ECH/DoH при отказе capture. URLBlocklist не заменяет
  сетевую защиту (Chromium разрешает часть blocked subresources); черновик отвергнут.
  Рабочий PAC автоматически заменяется inline guard: ordinary DIRECT, protected/custom и YouTube
  закрыты до повторной успешной активации. GUI/status явно показывают guard; Mandatory остаётся true.
- При восьми длительных googlevideo туннелях и исчерпании ожидания 2 с новый запрос получает
  явный admission error, не неработающий DIRECT. Очередь максимум 32 workers; требуется короткий ретест видео.
- DNS recovery googlevideo больше не исключён, но несуществующий DNS host/непроверенный TLS/HTTP
  не получают pin/PASS. Для проблемного CDN hostname результат провайдера ещё не проверен.
- Discord latency, QUIC/IPv6 возможность обхода и DIRECT+ tuning: отдельные задачи; код не менялся.

## (e) Кто чем владеет
| Область | Файлы | Агент | Ветка | Статус |
|---|---|---|---|---|
| B3 Chrome: каталог браузеров, автосоздание профиля DIRECT | `apps/desktop-core/BrowserCatalog.cs`, `DesktopViewModel.cs`, `Resources/Strings.resx`, `apps/desktop/ChromeRouting.cs`, `tests/desktop/BrowserCoverageTests.cs` | Claude Code | `claude/serene-cori-msra1u` | передано Codex. ГОТОВО (`9473d72`); C#/WPF-слой не компилировался, проверен только тестами core-логики; на ПК не проверен |
| Edge-полоса (история) | `apps/desktop/EdgeRouting*.cs`, `services/client-service/edge_*.go` | Claude Code | `claude/serene-cori-msra1u` | передано Codex; реализовано в `1ceaf07`, состояние и ограничения см. (j) |
| Застой YouTube (ноль байт после ClientHello): stallConn, кэш, повтор через релей, счётчики | `core/browserproxy/{proxy,stall}.go`, `core/flowpolicy/direct_timeout.go`, `services/client-service/{browser_direct_timeout,browser_transport,controller}.go` и тесты | Claude Code | `claude/serene-cori-msra1u` | передано Codex. ГОТОВО (`91c9016`, `28841f2`); только тесты, на ПК не проверено |
| Флаг `youtube_via_relay` (выключен) | `services/client-service/{controller,browser_direct_timeout}.go`, `apps/desktop-core/Models.cs` | Claude Code | `claude/serene-cori-msra1u` | передано Codex. ГОТОВО (`9bb43bd`) |
| diag-скрипт | `tools/windows/diag-no-vpn.ps1` | Claude Code | `claude/serene-cori-msra1u` | передано Codex. готово (`578a45f`, `28841f2`); дальше только по новым данным владельца |
| Лестница адресов, кэш по (домен, адрес), запасной переход без флага, мидстрим-детектор, ограничитель релея | `core/flowpolicy/direct_timeout.go`, `core/browserproxy/{proxy,stall}.go`, `services/client-service/{browser_direct_timeout,browser_transport}.go`, тесты | Claude Code | — | передано Codex, не начато; план ниже, нужно «да» владельца на старт |
| `diag-no-vpn.ps1 -Only youtube`: расширение (адреса googlevideo, столбец «процесс») | `tools/windows/diag-no-vpn.ps1` | Claude Code | — | передано Codex; базовая версия `ea4792a` готова |
| GUI A и YouTube/браузерный proxy B, Edge | `core/browserproxy/*`, `core/flowpolicy/direct_timeout.go`, `services/client-service/browser_*.go`, `controller.go`, `tools/windows/diag-no-vpn.ps1`, Edge-файлы, `apps/desktop-core/DesktopViewModel.cs`, `apps/desktop/MainWindow.xaml.cs`, новые GUI error-boundary файлы и тесты, `.github/workflows/edge-native.yml`, `tests/desktop/*`, `docs/BROWSER_DIRECT_FALLBACK.md`, `docs/EDGE_SITE_ROUTING.md`, `apps/desktop/MainWindow.xaml`, `apps/desktop-core/Resources/Strings.resx`, `platform/windows/browser_acceptance_windows_test.go`, `services/control-plane/{api,revocation}_test.go` (только фикстуры) | Codex | `codex/work` | ЗАВЕРШЕНО 19.10.2026: исправления закоммичены/запушены, файлы освобождены. Native Chrome/Edge, C#/WPF, Go race с PostgreSQL PASS. Полное снятие PAC остаётся открытым security blocker (см. (d)/(l)); QUIC/IPv6 без изменений |
| Discord медленно, DIRECT+ для YouTube | не определены | никто | — | свободно для Codex после диагностики; Go-файлы застоя (выше) не трогать |
| PR #8, `codex/direct-plus-flowseal` | — | не трогать | — | только владелец |

Строки Claude со статусом «передано Codex» сохранены как история; старые «нужно да»/запреты области не отменяют разрешение владельца на A/B. Новую работу снова помечать перед правками по PROTOCOL.md. QUIC/IPv6 routing менять только после отдельного «да».


Новые ветки: `claude/<тема>`, `codex/<тема>`, база — `claude/serene-cori-msra1u`.

## (f) Сборка и установка службы на ПК (по CLAUDE.md / docs/AGENT_RULES.md)
Только по просьбе владельца, на его ПК:
1. Закрыть Chrome, Edge и GUI («Выход» в трее). Сохранить локальные правки; `git switch codex/work`, `git pull --ff-only origin codex/work`. Полный stop-on-error блок: [BROWSER_DIRECT_FALLBACK.md](../BROWSER_DIRECT_FALLBACK.md#обновление-windows-без-vps).
2. Обычное окно PowerShell: `& .\tools\windows\build-e2e.ps1`, затем `& .\tools\windows\build-desktop.ps1`.
3. Elevated PowerShell под тем же SID владельца: `& .\tools\windows\update-service.ps1`.
4. Обычное окно: `& .\tools\windows\install-desktop.ps1`.
5. Сверить SHA256 `bin\windows-e2e\neexitlag-service.exe` и `%ProgramData%\neexitlag\neexitlag-service.exe`, `Get-Service neexitlag` = Running.

`powershell -File ...` ломает `$PSScriptRoot` (пустая строка в Join-Path): запускать только через `& .\tools\windows\скрипт.ps1`. Надпись «UPDATE COMPLETED» доказательством не является.
Запрещено: `update-vps.ps1`, `install.ps1`, `uninstall.ps1`, `setup-e2e.ps1`, `tools/vps-provision`, `tools/enroll`.

## (g) Как проверить без VPN
`& .\tools\windows\diag-no-vpn.ps1` (запускать из репозитория, не из старых копий; версия в шапке отчёта — `2026-10-19.1` или новее). Смотреть шапку (версия, коммит, `build` службы), разделы `sni`, `1b`, `1d` (`direct_stalls_total`, `stall_retries_via_relay`), `1f` (`probe_accounting`).
Доказательство: хеши службы совпадают и есть живая проверка на ПК владельца. Зелёный CI, лаборатория и Go-тесты доказательством работы на ПК не являются.
Управление флагом YouTube: `settings.set` с полным объектом Settings и `"youtube_via_relay": true` означает relay-first сразу для ordinary YouTube. По умолчанию выключен; аварийный fallback от него не зависит. Configured DIRECT+ и protected rules не заменяются.

## (h) Журнал решений
| Дата | Решение | Почему |
|---|---|---|
| 19.10.2026 | Edge: fresh nonce после последнего guard; confirm без запроса не READY и сам ставит inline guard | Предыдущий guard вызывал Forget и детерминированно удалял challenge; старый native fixture этот шаг пропускал |
| 19.10.2026 | Сохранить protected PAC guard при отказе; не выдавать URLBlocklist за сетевую защиту | URLBlocklistSubresources Chromium разрешает часть подресурсов; полное удаление PAC нарушает no-DIRECT invariant |
| 19.10.2026 | YouTube-family: первый ClientHello stall → TTL 10 мин без TLS повтора на другой IP; video limit 8, очередь 32/2 с | Новые данные владельца указывают на SNI-зависимый отказ; overflow DIRECT давал ноль ответа |
| 19.10.2026 | CI: перенос Edge XAML в ресурсы, native status через JSON, точные default profiles и штатный limiter в DB-фикстуре | Устранены реальные CI отказы без отключения проверок и без изменения production CP |

| 18.10.2026 | A: live PAC только после listener/recovery ACK; при отказе owned static guard, включая отсутствие профиля; AUTO pending для нового msedge | По данным владельца guard записывался до запуска listener, generic GUI catch стирал приложения. Mandatory PAC сохраняется ради protected fail-closed; чужая политика не удаляется |
| 18.10.2026 | B: fallback без флага, два IP/3 с, два разных отказа → домен relay-first TTL 10 мин, максимум четыре googlevideo CONNECT relay на ПК | Владелец подтвердил отсутствие DIRECT TCP ко всем восьми IP. Лестница из восьми адресов не помогает; VPS 1 vCPU/2 GiB, неизвестен лимит трафика |
| 18.10.2026 | Перед retry закрыть неответивший Engine tuple; не повторять TLS early/application data; первый байт/ACK не очищает cache verdict | `ErrFlowExists` мешал повторному OPEN того же tuple; safety и отсутствие ложного HTTP PASS |
| 18.10.2026 | QUIC/IPv6 только аудит: IPv6 guard уже есть, HTTP/3 DIRECT не разрешать по guessed IP | Нужна реальная origin/socket/protocol корреляция; отдельного «да» на изменение политики нет |
| 18.10.2026 | Edge: раздельные listener/PAC/recovery и состояние, запись HKU owner SID только службой; три общих файла изменены только добавлениями в `91d78c7` | Разрешено владельцем после синхронизации с `90d4817`; legacy Edge relay-first не меняется без явного `app_preset=edge-sites`; protected guard сохраняется при отказе/удалении профиля |
| 18.10.2026 | Codex: DNS-fix не переносить повторно; Edge-полоса разрешена, реализация ждёт согласования минимальных правок занятых файлов | DNS-коммиты уже входят в базу. Синхронизация: merge `221045f`. Служба допускает только Chrome и проверяет один Chrome PAC; запись HKCU сейчас выполняет GUI. Нужны точки интеграции в `browser_transport.go`, `controller.go`, `ChromeRouting.cs`; код не менялся, на ПК пока ничего запускать не нужно |
| 18.10.2026 | C2 (relay-first claude.ai в браузере) не включать | C1 (кэш таймаутов) покрывает |
| 18.10.2026 | C3 (запасной транспорт) только дизайн, не реализовывать | решение владельца |
| 18.10.2026 | Дефолтные relay-first профили: Claude, ChatGPT/Codex, Telegram Desktop; fail-closed | живые тесты показали необходимость |
| 18.10.2026 | DNS/53 и локальные адреса не идут в релей | регрессия после новых профилей |
| 18.10.2026 | Застой DIRECT («TCP есть, ответа нет», порог 2 с) считается отказом; два за 60 с — relay-first на 10 мин | TCP-connect не доказывает работу DIRECT |
| 18.10.2026 | YouTube через relay за флагом `youtube_via_relay`, выключен | нагрузка на VPS (1080p ≈ 4–5 Мбит/с на зрителя), лимиты тарифа неизвестны |
| 18.10.2026 | B3 только для Chrome; Edge отдельно | у службы один PAC и один recovery-exe |
| 18.10.2026 | AUTO: FlapGuard против дребезга маршрутов | 24 переключения за 2,5 часа |

### Разбор YouTube от 18.10.2026 (Claude Code, по коду; на ПК не проверено)
Требование владельца: маршрут выбирается автоматически, если какой-то недоступен. Данные: у www.youtube.com 8 A-адресов (142.251.150..157.4), Chrome даёт ERR_CONNECTION_CLOSED за 7,8 с, в Edge то открывалось, то нет. Гипотеза «часть адресов проходит напрямую, часть нет» не доказана.

**1. Зависимость от флага.** `relayAllowedFor` (`services/client-service/browser_direct_timeout.go`) вызывается в `steerTimedOutOrigin` (кэш уводит YouTube на релей только при флаге) и в `stallRetry` (при выключенном флаге возвращает ошибку, туннель закрывается: отсюда ERR_CONNECTION_CLOSED). Только для `youtubeDomain(host)`. Нужно: убрать проверку из обоих мест; флаг = «relay-first сразу, без попытки DIRECT».

**2. Единица кэша.** `DirectTimeoutKey{Network, Application(profile ID), Host, Port}` (`core/flowpolicy/direct_timeout.go:26`), адреса назначения нет; узлы `*.googlevideo.com` сворачиваются в один (`cacheHost`). Адрес выбирает `browserPlan` (`browser_transport.go:215-246`): резолв сам, IPv4 первыми, берётся `candidates[0]`. Один туннель = один адрес. Два застоя на одном плохом адресе за 60 с переводят на релей все 8 адресов на 10 мин. Фоновая проба идёт на адрес запустившего запроса. Второй DIRECT-адрес при застое не пробуется, сразу релей.

**3. Признаки отказа DIRECT.**
- Таймаут подключения: считается (`DeadlineExceeded`, 3 с на попытку в `Engine.OpenTCP`), но этот туннель браузер получает как 502, повтора через релей нет.
- Ноль байт после ClientHello (2 с): считается, повтор через релей.
- RST/EOF до первого байта ответа: считается, повтор через релей.
- Connection refused / нет маршрута: не считается, 502 без повтора.
- Застой посреди передачи (googlevideo): НЕ покрыт, после первого байта туннель неизменяем.

**4. Повтор до ошибки браузера.** Покрыт случай «ClientHello ушёл, ответа нет» (до 64 КБ, один раз, `core/browserproxy/stall.go`). Не покрыт отказ подключения: `core/browserproxy/proxy.go:455` пишет 502; сюда можно добавить повтор через следующий адрес и релей для обычных доменов (не protected, не DIRECT+). Мидстрим повторить нельзя (TLS не перенести): только закрыть туннель и пометить (домен, адрес), чтобы плеер переоткрыл соединение.

**5. Edge.** Общий код: `services/client-service/edge_transport.go:128` вызывает тот же `browserproxy.Listen` с `c.browserPlan`, `c.browserOpen`, `c.browserStall`, `c.browserDirectAnswered`. Кэш раздельный (в ключе ID профиля).

**6. Метрики.** Есть: `direct_stalls_total`, `stall_retries_via_relay`, `stall_retries_failed`, `active_origins`, `relay_steered_flows`, фоновые пробы, `top_origins`; в записях адрес назначения, `stall_retried_via_relay`, `owner.exe`/`profile_id`. Нет: кэша по (домен, адрес), причины отказа, остатка TTL, счёта запасных переходов по адресам.

**План (до 20 строк, исполнять только после «да» владельца):**
1. Ключ кэша = (сеть, профиль, домен, порт, адрес); для googlevideo домен остаётся свёрнутым, адрес сохраняется.
2. Выбор адреса: первый из `candidates`, не помеченный плохим; остальные не трогаем.
3. Лестница при застое: DIRECT на следующий адрес (почти бесплатно для VPS), потом релей; повтор работает всегда, `relayAllowedFor` остаётся только для режима «relay-first сразу».
4. Таймаут и refused при открытии: до 502 пробовать следующий адрес и релей для обычных доменов.
5. Мидстрим-детектор только для googlevideo: клиент отправил запрос, ответа нет ≥10 с: закрыть туннель, пометить пару (домен, адрес).
6. Диагностика: кэш (домен, адрес) с причиной, остатком TTL, числом переходов и перепроверок; счётчик запасных переходов.
7. `diag-no-vpn.ps1 -Only youtube`: адреса из системного DNS и 8.8.8.8, узел googlevideo (из записей, иначе `redirector.googlevideo.com`), столбец «процесс» из `owner.exe`. Базовая версия уже есть (`ea4792a`).
8. Go-тесты: лестница адресов, ключ по адресу, повтор при отказе подключения, мидстрим, флаг off/on, protected без утечек.
`browser_transport.go` в списке закрытых файлов: правка согласуется с владельцем отдельно.

**8. Нагрузка на VPS (оценка по коду, без измерений).** Пул на ПК: до 8 QUIC-сессий по 48 потоков (до 384 потоков, 376 под приложения). Релей: до 128 сессий всего, 64 потока на сессию (потолок 8192 потока). Клуб из 36 ПК при 1-2 сессиях на ПК = 36-72 из 128; при 4+ сессиях на ПК сервер начнёт отказывать. В коде нет лимита байт для TCP (для датаграмм 2000 pps на сессию); предел задают канал и CPU VPS. 1080p ≈ 4-5 Мбит/с на зрителя, 36 зрителей ≈ 160-190 Мбит/с, 1-4 ТБ/мес. Риск: если боевой релей работает в режиме `cidrs` (дефолт `deploy/*/example.env`), адреса Google/YouTube он откажет (Telegram работает, значит скорее всего public-only; проверить read-only). Ограничитель, не блокирующий: на ПК максимум 4 одновременных туннеля googlevideo через релей; сверх лимита туннель продолжает DIRECT (fail-open для обычных доменов), без отказа. Лестница адресов сама снижает нагрузку: на релей идёт только то, что не прошло напрямую.

**Открыто от владельца:** канал (Мбит/с) и месячный лимит трафика VPS, считается ли вход+выход (нужны для решения по флагу `youtube_via_relay`). Результат проверки `claude/merge-edge-trial`: GUI показывает «Не удалось выполнить действие» и пустую таблицу, причину ждём по IPC-выводу (`status`, `profiles`, `browser.status`, `edge.status`, файлы `%LOCALAPPDATA%\NeExitLag\*-recovery.json`); механизм: `MainWindow.Execute` на любое неожиданное исключение вызывает `SetUnavailable` и очищает таблицу.

## (i) Последние коммиты ветки (от нового к старому)
- `9bb43bd` флаг `youtube_via_relay` (Go Settings + C# `ServiceSettings`), YouTube в релей только при включении; застои считаются всегда.
- `68e070a` стабилизация теста застоя (выбор записи по ID).
- `28841f2` отпечаток сборки службы (`build` в status), сквозной тест застоя, таблица SNI и шапка версий в diag.
- `a3ab95f` STATUS: застой DIRECT и исключение DNS/локальной сети.
- `91c9016` застой DIRECT-туннеля: счёт как отказ и прозрачный повтор через релей (stallConn, кэш, счётчики).
- `578a45f` diag: честный вердикт 8443, проверка YouTube по SNI.
- `b08de8f` DNS/53 и локальная сеть не уходят в релей у relay-first профилей.
- `9473d72` B3: поиск браузеров и автосоздание профиля Chrome (C# не компилировался).
- `d86de75` учёт проб по маршрутам (`quality.diagnostics.probe_accounting`).
- `c223af7` FlapGuard AUTO.
- `0fc9f97`, `3f75761` DirectTimeoutCache и его подключение к браузерному транспорту.
- `e436acf` дефолтные профили Claude/ChatGPT/Telegram.
- `ac5cf81` автопродление клиентских сертификатов; `673c866`, `f8dbf6b`, `d988b58`, `550a91f`, `df672c3`, `e075960` — надёжность запуска, rate limit, idle TCP, replay/fsync, сетевой снимок.
Подробности — `docs/STATUS.md`, `docs/BLOCKERS.md`.


## (j) Исторический handoff первой Edge-полосы, 18.10.2026

Данные и номера строк в этом разделе относятся к `1ceaf07`. Актуальные A/B изменения — раздел (k); старые утверждения «не изменены» не описывают новые разрешённые правки.

Синхронизация: `git fetch origin`, merge `221045f` с `90d4817`; конфликт таблицы разрешён по явному решению владельца. DNS-fix повторно не переносился.

- `c017322`: изолированная запись Edge ProxySettings службой в HKU installation owner SID; отказ для чужой политики.
- `9d66683`: новые Edge GUI/recovery/core файлы, Edge discovery, автосоздание одного DIRECT-профиля, отдельные кнопки/таблица и C#-регрессии.
- `91d78c7`: служебная Edge-полоса, nonce/PID/SID readiness, IPC и Windows acceptance. **Все изменения трёх совместных файлов находятся только в этом коммите, +28/-0.**
- `78dc824`, `a18967c`: guard-only restart, честный Edge-статус, защита от удаления профиля до guard; остальные файлы Chrome не тронуты.

Точные добавленные строки **в `91d78c7` и итоговой версии этой сессии**:

| Файл | Строки | Что добавлено |
|---|---|---|
| `services/client-service/browser_transport.go` | 28; 64–66; 626; 669–671; 678–680 | отдельный Edge runtime; Edge owner/profile dispatch; независимое восстановление; отдельная readiness/исключение из Chrome-проверки |
| `services/client-service/controller.go` | 215; 485–496; 512–513 | закрытие Edge; отдельная сериализованная IPC-транзакция и публикация guard; Edge IPC dispatch |
| `apps/desktop/ChromeRouting.cs` | 33; 77 | вызовы Edge coverage и мониторинга |

Вне этих трёх: новые `edge_*.go`, `EdgeRouting*.cs`, Windows `edge_policy*`, Edge-добавление в BrowserCatalog, кнопки/таблица MainWindow.xaml, добавляющая Edge-ветка browser_policy/recovery_guard/ipc_handlers, тесты и новый workflow `edge-native.yml`. `DesktopViewModel.cs`, `Models.cs`, `proxy.go`, `stall.go`, `direct_timeout.go`, `browser_direct_timeout.go`, `youtube_via_relay` не изменены.

Проверено локально: `go test -race -timeout 120s ./...`, `go vet ./...`, `go build ./...`, Windows amd64 cross-build; после дополнительных Edge-правок — focused race/vet и cross-compile Windows test binary. C#/.NET/Edge в Linux-окружении отсутствуют: C# и WPF **локально не компилировались**, native Windows тест не запускался. Добавлены C# regression tests и actual Edge HTTPS/HTTP2 + mTLS relay + outage/restart test (9 контролируемых сайтов, неполные evidence дают FAIL), foreign policy/SID refusal. Никаких Internet PASS или native Edge DIRECT+ PASS не заявлено.

CI: GitHub connector возвращает пустой список запусков для новых Edge-коммитов; `gh run list` получает `Forbidden`. Поэтому CI PASS не подтверждён. Новый workflow: [Edge policy acceptance](https://github.com/buzzmean/neexitlag/blob/codex/work/.github/workflows/edge-native.yml); [запуски ветки](https://github.com/buzzmean/neexitlag/actions?query=branch%3Acodex%2Fwork). Известный старый Windows Verify failure на `c53a4f7` не исправлялся и не перезапускался через PR #8.

Осталось: получить actual C#/WPF/Windows CI; коротко проверить Edge на ПК владельца без VPN. Существующий Edge MANUAL/DIRECT/старый relay-first профиль не мигрирует автоматически — opt-in через «Сайты Edge AUTO». Edge предупредит «Управляется вашей организацией»; «Отключить PAC Edge» освобождает ordinary DIRECT и сохраняет protected guard. Recovery требует GUI; после утраты per-user recovery-state повторная активация проверяет новый процесс, а не доверяет старому ready-флагу. VPS/credentials/certificates/PR #8 не менялись, ветка `codex/direct-plus-flowseal` сохранена.

На ПК: выполнить [обычную пересборку/обновление службы с SHA256-проверкой](../EDGE_SITE_ROUTING.md), включить Edge в VELIQ, VPN OFF, затем `& .\tools\windows\diag-no-vpn.ps1 -Only browser`; `diag-result.txt` не коммитить. До подтверждения Windows CI не считать этот код готовым production-релизом.


### Уточнение владельца и прогресс Codex, 18.10.2026

- Разрешены задачи A/B; обращения за разрешением больше не нужны без блокера.
- Реестр после отказа Edge уже содержал owned `ProxySettings`: первоначальный отказ
  не доказан как ошибка выбора строки. Подтверждены дефекты порядка guard до listener,
  раннего выхода отката без профиля и общего GUI-обработчика, очищавшего таблицу.
- A: `9435dbf`, `0f5788d`, `430d323`: граница ошибок GUI, listener/recovery до live PAC,
  guarded rollback без профиля, phase/code в GUI + `edge.status.last_error`, новый AUTO pending.
  Go-регрессии проходят; .NET/WPF и живой Windows/ISP здесь не исполнялись.
- B: `5895d4b`, `a65e6ab`: адресный/доменный кэш, bounded first-flight и midstream.
  Дальше: интеграция service fallback/лимитера/байт, полный race/vet/build, диагностика.
- Область нового документа `docs/BROWSER_DIRECT_FALLBACK.md` закреплена за Codex:
  семантика отказов/лимитов, read-only QUIC/IPv6-аудит и команды владельцу. Маршруты
  QUIC/IPv6 до отдельного «да» НЕ менять.


## (k) Итог задач A/B Codex, 18.10.2026

Код: `9435dbf`, `0f5788d`, `430d323`, `5895d4b`, `a65e6ab`, `0092c56`,
`79dfaaa`, `d4fedab`, `f46ed9d`, `b50f0e6`, `d86f4a6`, `c21995a`, `a1e7b5f`, `ad3c69c`, `dea7ab7`, `5271669`. Ветка только codex/work,
без rebase/force, без merge PR #8 и без изменения VPS/credentials/certificates.

**A — подтверждённые дефекты и исправления.**
- Generic GUI Execute catch вызывал SetUnavailable, стирал Applications и выключал
  CanControl при ошибке отдельной операции. Теперь Edge rejection/recovery failures
  сохраняют таблицу; настоящая потеря authenticated core status обрабатывается отдельно.
- Кнопка больше не требует предварительно выбранной строки при единственном Edge.
  В реальном отказе registry уже был записан: точная фаза старого отказа без last_error
  не установлена, ошибка выбора строки не выдаётся за доказанную причину этого прогона.
- recovery start → edge.prepare listener/profile/nonce → configuration ACK → guard/quit
  → service live PAC → настоящий headless policy probe → edge.confirm → profile.set AUTO.
  Новая AUTO identity имеет edge-sites-pending; внешние flows fail-closed до activation,
  затем edge-sites. Старый DIRECT не был причиной отказа подготовки: этот режим разрешён.
- Read-only `edge.status.activation_version >= 2` проверяется новым GUI до создания
  pending AUTO/записи PAC, в том числе при ручном добавлении msedge. Старой службе не
  отправляется неизвестный pending-профиль; обновлять нужно service и GUI вместе.
  Capability version не выдаётся за listener/PAC readiness.
- Откат через service writer без профиля, journal ownership, перенос custom protected
  suffixes из старого PAC, строгий guard при повреждённом состоянии. Чужие proxy keys не
  меняются. GUI показывает ordinary DIRECT только если service вернул ordinary_direct.
- Ошибки публикации recovery guard до listener/после handleEdge тоже фиксируются
  в last_error, с безопасным кодом вместо OS path/данных.
- Отдельно различаются `EDGE_IPC_ACCESS_DENIED`, `EDGE_GUARD_STATE_WRITE_FAILED`
  и `EDGE_POLICY_ACCESS_DENIED`: pipe/SID, файл guard и registry не смешиваются.
- last_error всегда присутствует в edge.status; GUI показывает безопасные phase/code
  и русский текст. Ошибка сохраняется в EdgeState после timer refresh; безопасный
  `%LOCALAPPDATA%\NeExitLag\gui-failure.json` содержит phase/code/type/time.
- `%ProgramData%\neexitlag\logs` не создаётся текущим установщиком: slog идёт в stderr;
  SCM журнал — только lifecycle. Команды IPC/GUI diagnostics в EDGE_SITE_ROUTING.md.
  Кто удалил ProxySettings в первоначальном тесте, по имеющимся данным не установлен.

**B — что изменилось.**
- TCP open timeout/refused/no-route: не больше двух разных IPv4 DNS адресов, общий DIRECT
  deadline 3 с, затем relay до 502. Source/probe/attempts отражают фактическое решение.
- Address cache и доменный verdict: два разных IP/60 с → TTL 10 мин; bounded фоновые
  перепроверки с настоящими TLS/certificate/GET, не более четырёх probe sockets.
  Ни ACK, ни первый TLS byte не очищает verdict. Сеть/профиль/домен/порт/IP изолированы.
- Нет флаговой блокировки аварийного fallback; youtube_via_relay теперь только relay-first
  preference. Protected, explicit rules, intranet и configured DIRECT+ не понижаются.
- `direct_attempts` — количество реальных вызовов открытия, а не индекс кандидата:
  при полном cache hit равно нулю. Слот адреса отдельно `direct_address_slot`.
- Silent first flight безопасно повторяется только до ответа; TLS early data запрещены.
  Старый сокет освобождает Engine tuple до открытия нового. Established application
  bytes никогда не мигрируют/не повторяются.
- Midstream googlevideo: после ответа и новой исходящей записи 10 с без входящих байт
  закрывает только этот DIRECT сокет и отмечает домен/IP для следующего connection.
- Четыре googlevideo CONNECT relay tunnels на ПК, сверх ordinary AUTO — DIRECT;
  required/custom protected не понижаются. Если DIRECT заблокирован ISP, overflow может
  не работать — лимит не является обещанием просмотра на 36 ПК.
- Shared relay_usage: реальные active I/O bytes/minute/profile/domain, не probes и не
  QUIC framing. Retention 10 мин/1024 keys/64 IPC rows; truncation/overflow видимы.
- diag-no-vpn.ps1 -Only youtube v2026-10-18.7: Chrome+Edge/PID/exe, IP/domain cache,
  minute bytes, system/8.8.8.8 DNS, реальные googlevideo hostnames и SNI-контроли.
  DOM RECEIVED не выдаётся за полный site PASS.

**Проверки.** Full Go race/vet, Linux build, Windows amd64 build/vet PASS; cross-compiled
service/capture Windows tests; 21 Python deployment regression PASS. Focused тесты:
ошибки listener/ACL/чужой policy/старого guard/удалённого профиля; pending AUTO;
два разных IP и TTL; два молчащих DIRECT сокета → relay; настоящий routing.Engine tuple;
no protected downgrade; cache не очищается по ACK/одному TLS байту; midstream progress,
idle timeout, лимит/release/concurrency, minute bytes/cardinality bounds.

C#/WPF **не компилировались локально** (.NET нет); native Windows/ISP здесь не запускались.
Новые C# регрессии и PowerShell 5.1/7 AST checks включены в Edge CI. Connector для свежих
HEAD возвращает workflow_runs=[], gh API Forbidden; CI PASS НЕ подтверждён:
[Edge workflow](https://github.com/buzzmean/neexitlag/actions/workflows/edge-native.yml),
[запуски codex/work](https://github.com/buzzmean/neexitlag/actions?query=branch%3Acodex%2Fwork).

**Осталось владельцу/следующей сессии.** Получить Windows/C# CI, один короткий Edge
activation и YouTube check после обновления обоих бинарников. Никакого 65-минутного
теста не требуется. QUIC/IPv6: только анализ завершён; код маршрутизации не менялся.
Нужны metadata protocol/socket evidence и отдельное «да» перед реализацией DIRECT QUIC.
Read-only команды Windows/VPS и риски 1 vCPU описаны в BROWSER_DIRECT_FALLBACK.md.
Поддержка IPv6 relay/HTTP3 DIRECT и Internet PASS этой версии НЕ заявляются.

На ПК: закрыть Chrome/Edge/GUI, выполнить один stop-on-error блок обновления из
[BROWSER_DIRECT_FALLBACK.md](../BROWSER_DIRECT_FALLBACK.md#обновление-windows-без-vps),
запустить GUI обычным пользователем → «Сайты Edge AUTO» → VPN OFF →
`& .\tools\windows\diag-no-vpn.ps1 -Only youtube,browser`. Не коммитить diag-result.txt.


Точки изменений общих файлов на code HEAD `5271669`:
- `controller.go:322` — pending AUTO guard; `controller.go:490` — Edge mutation
  transaction, включая новый error-boundary в строках 493–503. Edge IPC dispatch
  дополнен `edge.failure` (только безопасные коды, существующая owner-auth не менялась).
- `browser_transport.go:163` — actual I/O meter/video permit/midstream wrapper;
  `:226` — общий deadline и максимум два IPv4 кандидата только ordinary YouTube;
  `:509`/`:516` — bounded status с cache/usage; `:586`/`:696` — OpenPlan + strict
  ClientHello-only retry для обоих путей восстановления listener.
- `apps/desktop/ChromeRouting.cs` в задачах A/B не менялся; его первоначальные две
  Edge-точки остаются в `91d78c7`. Логика Chrome PAC/YouTube strategy не переписывалась.

Владение освобождено; следующей сессии прочитать (k), сохранить историю Claude и
сначала fetch/pull. Код для Windows ретеста — `5271669`; последующий HEAD содержит
только финальную документацию/версию diag. Полные Go-логи тестов не коммитились.


## (l) Codex: Edge nonce и YouTube SNI, 19.10.2026

- `983d694`: первый SNI stall YouTube-family сразу доменный relay-first TTL 10 мин;
  второй DIRECT адрес допускается только при отказе TCP OPEN, общий budget до 3 с.
  Googlevideo: 8 active, 32 bounded waiters, до 2 с; DIRECT overflow удалён.
- `6c075ae`: DNS recovery/cache для googlevideo, исходный SNI/cert + GET, negative cooldown,
  TTL и сохранение configured DIRECT+. При провале подробная DNS evidence не теряется.
- `984c1a5`: Edge final guard → fresh prepare/nonce → actual policy browser GET → confirm;
  version=3, readiness_state/pending_profile/rollback_guard_active, auto guard при отсутствии GET.
  PID/birth/exe/profile/SID по-прежнему проверяются на фактическом TCP peer, не на GUI PID.
- `1e90c7f`: явный admission failure, очередь и диагностический скрипт `2026-10-19.1`.
- `6a64f14`: Edge XAML подписи в русских resources; строгий ResourceTests сохранён.
- `0aeebcb`: native Chrome читает публичный JSON status вместо type assertion внутреннего wrapper;
  DB-фикстуры ожидают все текущие profiles с точным содержимым и используют реальный limiter.
  CP production/migrations, mTLS/lease/scoring, файлы VPS не изменены.
- Chrome clean install: Defaults/discovery → один DIRECT full-path profile → штатная PAC
  activation → AUTO. Existing profiles не перезаписываются; core тесты прошли CI, WPF build тоже.
- Никакой .NET/Edge локально нет; C#/WPF и actual Windows исполнялись только GitHub CI.
- Owner action: закрыть браузеры/GUI, обновить оба binaries по stop-on-error блоку
  [BROWSER_DIRECT_FALLBACK.md](../BROWSER_DIRECT_FALLBACK.md#обновление-windows-без-vps),
  проверить SHA256 installed==built. Новый GUI обычным SID владельца → «Сайты Edge AUTO»;
  GUI сам делает штатный перезапуск с согласия (Startup Boost может требовать ручного закрытия).
  VPN OFF: `& .\tools\windows\diag-no-vpn.ps1 -Only youtube,browser`.
- Оставшееся требование полного снятия PAC — реальный security blocker, описан в (d) и
  [EDGE_SITE_ROUTING.md](../EDGE_SITE_ROUTING.md#отказ-и-откат); не считать его выполненным.

### CI и проверки кода 0aeebcb

- [Native Chrome PASS](https://github.com/buzzmean/neexitlag/actions/runs/39219040557):
  реальная policy, owner/SID, native capture, DIRECT+, outage/restart; только controlled origins.
- [Edge policy PASS](https://github.com/buzzmean/neexitlag/actions/runs/39218637437) на 6a64f14:
  C# 108 tests, WPF и реальный Edge fresh nonce/HTTPS/recovery. Последующие правки — только фикстуры/документы.
- [Relay release PASS](https://github.com/buzzmean/neexitlag/actions/runs/39219040514),
  [isolated upgrade PASS](https://github.com/buzzmean/neexitlag/actions/runs/39219040504).
- [Verify](https://github.com/buzzmean/neexitlag/actions/runs/39219040508): Linux core PASS;
  Windows build/native bursts/accelerated lease PASS; на момент записи финальные PS/IPC checks ещё идут.
- Локально: `go test -race -timeout 120s ./...` с отдельной PostgreSQL 17 PASS, `go vet ./...`,
  Linux build и Windows amd64 build/vet PASS; native test binaries cross-compiled, deployment Python 21 PASS.
  Тестовый контейнер удалён; ни один command не обращался к production VPS.
- `github_fetch_commit_workflow_runs` фильтрует PR events: его пустой ответ НЕ доказывает отсутствие push CI.
  Читать `/actions/runs?head_sha=...` через read-only GitHub fetch; CLI gh в текущей среде получает Forbidden.
