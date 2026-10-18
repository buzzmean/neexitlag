# CURRENT — состояние проекта (обновляется в конце каждой сессии)

Обновлено: 18.10.2026, Claude Code передал работу Codex; ветка `claude/serene-cori-msra1u`.

## (a) Цель
Чистый Windows 10/11: установил, нажал «Оптимизировать» — заблокированные в РФ сервисы работают без VPN: ChatGPT, Claude, YouTube, Discord, Telegram, браузеры Chrome и Edge. Первый заказчик — клуб на 36 ПК с откатом при перезагрузке.

## (b) Состояние на 18.10.2026
- VPS 2.56.177.4 обновлён по PR #8 (17.10). Сертификат релея и control-plane продлён до 16.01.2027 (следующее продление ~2.01.2027).
- Потери UDP до VPS в покое: 0%.
- На ПК владельца установлена служба из ветки `claude/serene-cori-msra1u`. Проверка: `Get-FileHash bin\windows-e2e\neexitlag-service.exe` (SHA256 должен совпасть с `%ProgramData%\neexitlag\neexitlag-service.exe`). Последний известный хеш начинался с `5A7C3BCB` (сборка до `28841f2`; в status поле `build` покажет коммит новых сборок).
- Службу на ПК нужно пересобрать, чтобы получить коммиты `28841f2`..`9bb43bd` (отпечаток сборки, флаг YouTube).

## (c) Что работает на живом ПК
Telegram, Claude, ChatGPT, Chrome (профиль Chrome создан вручную).

## (d) Открыто
- YouTube: DIRECT без ответа (гипотеза: блокировка по SNI; проверяет секция `sni` в diag). Прозрачный повтор через релей реализован, но за флагом `youtube_via_relay` (по умолчанию выключен) до ответа владельца про канал и лимит трафика VPS.
- Edge-полоса реализована в `codex/work` (коммиты `c017322`, `9d66683`, `91d78c7`, `78dc824`, `a18967c`): отдельные PAC/CONNECT/recovery, owner-SID policy через службу, guarded rollback. Реальная Windows/ISP-проверка и C#/WPF-компиляция пока НЕ подтверждены; см. (j) и `docs/EDGE_SITE_ROUTING.md`.
- Discord медленно (headless Chrome: 15.4 с до DOM); причина не найдена.
- DIRECT+ для YouTube выключен по умолчанию; путь привязки к Chrome AUTO не разобран до конца (`ApplyPresetAsync` ставит DIRECT+ на весь профиль).
- Ложный вердикт 8443 в diag-no-vpn.ps1: исправлен в `578a45f`/`28841f2` (OK только при коде curl 0 и времени TLS > 0), на ПК нужен свежий скрипт.
- Не проверено на реальном ПК: все правки после `e436acf`; WPF-слой не компилировался (проверен только `dotnet test` на core-логике); ложные срабатывания порога застоя 2 с.

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
| GUI A и YouTube/браузерный proxy B, Edge | `core/browserproxy/*`, `core/flowpolicy/direct_timeout.go`, `services/client-service/browser_*.go`, `controller.go`, `tools/windows/diag-no-vpn.ps1`, Edge-файлы, `apps/desktop-core/DesktopViewModel.cs`, `apps/desktop/MainWindow.xaml.cs`, новые GUI error-boundary файлы и тесты, `docs/BROWSER_DIRECT_FALLBACK.md`, `docs/EDGE_SITE_ROUTING.md` | Codex | `codex/work` | В РАБОТЕ 18.10.2026: A — диагностика и план; B — два DIRECT адреса/≤3 с, relay-first TTL 10 мин, лимит googlevideo=4, поминутные байты. QUIC/IPv6 только аудит до отдельного «да» |

| Discord медленно, DIRECT+ для YouTube | не определены | никто | — | свободно для Codex после диагностики; Go-файлы застоя (выше) не трогать |
| PR #8, `codex/direct-plus-flowseal` | — | не трогать | — | только владелец |

Для Codex, без пересечения: Edge-«полоса» (после «да» владельца) или разбор Discord/DIRECT+ в `core/localstrategy/` и `services/client-service/browser_flow*.go`. Перед стартом пометить строку здесь и сделать `git pull`.


Новые ветки: `claude/<тема>`, `codex/<тема>`, база — `claude/serene-cori-msra1u`.

## (f) Сборка и установка службы на ПК (по CLAUDE.md / docs/AGENT_RULES.md)
Только по просьбе владельца, на его ПК:
1. Закрыть Chrome и GUI («Выход» в трее). `git branch --show-current` = `claude/serene-cori-msra1u`, `git pull`.
2. Обычное окно PowerShell: `& .\tools\windows\build-e2e.ps1`, затем `& .\tools\windows\build-desktop.ps1`.
3. Elevated PowerShell под тем же SID владельца: `& .\tools\windows\update-service.ps1`.
4. Обычное окно: `& .\tools\windows\install-desktop.ps1`.
5. Сверить SHA256 `bin\windows-e2e\neexitlag-service.exe` и `%ProgramData%\neexitlag\neexitlag-service.exe`, `Get-Service neexitlag` = Running.

`powershell -File ...` ломает `$PSScriptRoot` (пустая строка в Join-Path): запускать только через `& .\tools\windows\скрипт.ps1`. Надпись «UPDATE COMPLETED» доказательством не является.
Запрещено: `update-vps.ps1`, `install.ps1`, `uninstall.ps1`, `setup-e2e.ps1`, `tools/vps-provision`, `tools/enroll`.

## (g) Как проверить без VPN
`& .\tools\windows\diag-no-vpn.ps1` (запускать из репозитория, не из старых копий; версия в шапке отчёта — `2026-10-18.5` или новее). Смотреть шапку (версия, коммит, `build` службы), разделы `sni`, `1b`, `1d` (`direct_stalls_total`, `stall_retries_via_relay`), `1f` (`probe_accounting`).
Доказательство: хеши службы совпадают и есть живая проверка на ПК владельца. Зелёный CI, лаборатория и Go-тесты доказательством работы на ПК не являются.
Управление флагом YouTube: `settings.set` с полным объектом Settings и `"youtube_via_relay": true` (по умолчанию выключен).

## (h) Журнал решений
| Дата | Решение | Почему |
|---|---|---|
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


## (j) Codex — Edge-полоса, 18.10.2026

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
