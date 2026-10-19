# CURRENT — состояние проекта (обновляется в конце каждой сессии)

Обновлено: 19.10.2026, Codex; ветка `codex/work`. Игры на ec72beb подтверждены владельцем: CS DIRECT, relay_flows=0. Browser paths/Edge runtime и Windows updater — код `1db6be1`; Linux/Windows Verify, actual Chrome/native, Edge policy, release и isolated upgrade CI PASS. Нового ISP PASS нет; короткий ретест владельца с открытыми браузерами ещё нужен. PAC guard сохранён, этап (в) не начат. Подробности и команды — (p), docs/BROWSER_HOT_ATTACH.md. История Claude сохранена.

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
  Это контролируемые Windows origins, не Internet PASS. Фикстуры 0aeebcb прошли Linux/Windows Verify, native Chrome, relay release и isolated upgrade; ссылки в (l).

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
| Этап (а): профили Chrome/Edge при запуске GUI и диагностика; PAC/restart без изменений | `apps/desktop/BrowserStartup*.cs`, `apps/desktop/MainWindow.xaml.cs`, `apps/desktop/ChromeRouting.cs` (только сохранение старого activation path), `apps/desktop-core/BrowserStartup*.cs`, `apps/desktop-core/BrowserCatalog.cs`, `apps/desktop-core/DesktopViewModel.cs` (создание/диагностика), `apps/desktop/MainWindow.xaml`, `tests/desktop/*`, `tests/desktop-bindings/*`, `tools/windows/diag-no-vpn.ps1`, `docs/BROWSER_HOT_ATTACH.md`, `docs/handoff/CURRENT.md` | Codex | `codex/work` | Этап (а) ГОТОВ 19.10.2026 (`c556ac6`), область освобождена, проверка ПК PASS (4 мс, PID/start time Chrome неизменны); локальные Go/Python PASS, C#/WPF, Edge policy и полный Linux/Windows Verify CI PASS. Этап (б) отдельно, (в) только после проверки владельца; VPS/QUIC/IPv6 не менять |
| Этап (б): native browser AUTO и автоматическое восстановление Edge | `services/client-service/browser_*.go`, `edge_*.go`, `controller.go`, `ipc_handlers.go`, `flow_policy.go`, `core/profiles/profile.go` (backward-compatible browser_transport), `platform/windows/{boundary,hot_attach,ipv6_guard,deferred_tcp,tcp_capture_windows,browser}*.go`, browser acceptance/тесты, `tools/route-test/upgrade_test.go` (только fresh-ticket/clock фикстура), `apps/desktop{,-core}/Browser*.cs`, `EdgeRouting*.cs`, `ChromeRouting.cs`, `DesktopViewModel.cs`, `MainWindow.xaml{,.cs}`, `tests/desktop/*`, `tools/windows/diag-no-vpn.ps1`, `.github/workflows/*native.yml`, `docs/BROWSER_HOT_ATTACH.md`, `CURRENT.md` | Codex | `codex/work` | ГОТОВО 19.10.2026 (`b3dcd1c`), область освобождена; этап (б) ждёт проверки ПК владельца с открытыми браузерами и existing owned PAC. Этап (в) не начат; VPS, QUIC/IPv6 вне правок |
| UDP latency / аудит очередей вне выбранных маршрутов | `platform/windows/capture_queue*.go` (новые), `capture_windows.go`, `tcp_capture_windows.go`, `ipv6_guard_windows.go` (только замеры), `boundary.go` (статус), `services/client-service/controller.go` (новый read-only IPC), `capture_queue_test.go` (новый), `tools/windows/diag-no-vpn.ps1`, `.github/workflows/verify.yml` (только новые Windows tests), `docs/UDP_CAPTURE_AUDIT.md` (новый), `CURRENT.md` | Codex | `codex/work` | ПРИОСТАНОВЛЕНО 19.10.2026: владелец подтвердил 28025 UDP packets CS через VPS; причина — маршрут application default RELAY. Код фильтров/очередей не менялся, область освобождена |
| Игры/неизвестные приложения: безопасный DIRECT по умолчанию и явное RELAY предупреждение | `core/profiles/profile{,_test}.go`, `services/client-service/{controller,flow_policy}.go`, новые `automatic_relay*.go`, `services/client-service/*_test.go` (регрессии routing), `apps/desktop-core/{DesktopViewModel,RelayWarning}.cs`, `apps/desktop/{MainWindow.xaml,RelayWarning}.cs`, `apps/desktop-core/Resources/Strings.resx`, `tests/desktop/*`, `tools/windows/diag-no-vpn.ps1`, `tests/windows-tools/profile-routing-checks.ps1` (новый), `.github/workflows/verify.yml` (PS fixtures), `docs/GAME_DIRECT.md` (новый), `CURRENT.md` | Codex | `codex/work` | ГОТОВО 19.10.2026 (`50f750f`, `060bc19`, `2815bd0`), область освобождена; владелец подтвердил исходный AUTO. Старые игровые AUTO безопасны, новые обычные приложения DIRECT; явный MANUAL и browser/protected presets сохранены. Проверено владельцем на ec72beb: cs2 relay_flows=0, игровой flow DIRECT; баг пинга закрыт |
| Browser paths: PAC/native, Edge readiness, ошибки origin | `services/client-service/browser_*.go`, `edge_*.go`, `controller.go` (только browser hooks), `core/browserproxy/*` (метрики/ошибки), `apps/desktop{,-core}/BrowserNative*.cs`, `EdgeRouting*.cs`, `Models.cs` (добавочные поля), `tests/desktop/*`, `tests/windows-tools/*`, `tools/windows/diag-no-vpn.ps1`, `tests/desktop/BrowserStartupTests.cs` (только cleanup собственной фикстуры), `.github/workflows/{chrome-native,verify}.yml` (read-only warmup/fixtures), `tools/windows/{native-tcp-firewall,update-service}.ps1` (явная opt-in Windows app-scoped admission и rollback; не VPS), browser Windows acceptance, `platform/windows/{browser_windows,browser_other,tcp_capture_windows,boundary,deferred_tcp}.go` (только exact delegated upstream, без изменения фильтра), `docs/BROWSER_HOT_ATTACH.md`, `CURRENT.md` | Codex | `codex/work` | ГОТОВО 19.10.2026, код 1db6be1, область освобождена: PAC/native раздельно, exact WinDivert handoff, Edge readiness и opt-in app-scoped firewall; CI PASS, ISP ретест ещё нужен. Игры/PAC guard сохранены, этап (в) не начат |
| Cold-start browser retry / native-only audit | `core/browserproxy/{proxy,stall,path_metrics}*.go`, `core/browserproxy/*_test.go`, `services/client-service/browser_*.go`, `services/client-service/controller.go` (read-only browser IPC), `platform/windows/{tcp_capture_windows,browser*,native_firewall*}.go`, Windows browser tests, `internal/browserfixture/*` (cold-start fixture), `tools/windows/diag-no-vpn.ps1`, `tests/windows-tools/*`, `.github/workflows/{chrome-native,verify}.yml` (tests), `docs/BROWSER_HOT_ATTACH.md`, `CURRENT.md` | Codex | `codex/work` | В РАБОТЕ: e9e31ac проверен владельцем (оба браузера, PID неизменен, игры DIRECT). Аудит native-only без снятия PAC; cold DIRECT stall → relay, фактический firewall status. Этап (в) не начинать |
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
| 19.10.2026 | Windows native TCP reflection требует app-scoped inbound admission; update-service получает явный -NativeTcpFirewall с owned rollback | Driver trace показала SYN без listener SYN/ACK; узкое правило восстановило nonce. Production helper проверен на отдельном EXE. Firewall не отключается, чужие правила/GPO не обходятся; PAC guard сохранён |
| 19.10.2026 | Edge native OPEN без ответа не считается готовым туннелем; PAC nonce readiness показывать отдельно | Успешный upstream OPEN не доказывает завершённый local handoff или загрузку страницы |
| 19.10.2026 | Считать PAC ingress и native upstream отдельно; CONNECT authority позволяет открыть известный план в native worker до ACK, без второго ClientHello | Online PAC забирает все browser origins на loopback, внешнего SYN браузера нет. Нельзя выдавать PAC bytes или RefreshPolicy REQUESTED за native uptake; прежняя защита opaque без metadata сохраняется |
| 19.10.2026 | PLAN, OPEN и pre-response stall retry имеют отдельные bounded budgets | Общий deadline истекал во время DNS/TLS/HTTP planning; новый OPEN/retry получал остаток или уже expired context. Старые 35 ошибок потеряны и не могут быть приписаны одному дефекту |
| 19.10.2026 | Игры и неизвестные AUTO не получают global/learned relay; новый обычный профиль DIRECT; явный MANUAL требует подтверждения GUI | CS с подтверждённым AUTO шёл через VPS: 28025 UDP packets. Глобальный probe winner не измеряет качество пути к игровому серверу; browser/protected/Discord presets и явные правила сохраняются |
| 19.10.2026 | Диагностика различает configured default и фактические flow counters с PID/creation time | Browser usage не показывает игровой UDP; source=profile не определяет автора профиля. Старые файлы не хранят автора/время override; отсутствие flow в bounded snapshot не доказывает отсутствие relay |
| 19.10.2026 | Immediate-renewal test использует уникальные ID и короткое разнесение test ticket issues | Loopback issuer с нулевой задержкой и coarse Windows clock давал ложный deadline-extension FAIL; production lease/security и assertions не меняются |
| 19.10.2026 | Native activation отделена logical flow ID; публикация нового профиля fences UDP, сохраняет все TCP | Windows timestamps могут совпасть; старый поток не должен давать ложный activation PASS, старый DIRECT UDP/443 не должен переживать AUTO |
| 19.10.2026 | Windows burst echo target вынесен в отдельный процесс | Independent dial подтвердил loopback refusal вне admission/pool; не снижены 160/256 concurrency и строгие assertions, production параметры не меняются |
| 19.10.2026 | Этап (б): owned PAC сохраняется как независимый crash-guard; fresh без него автоматически не активировать | Владелец подтвердил: WinDivert снимает фильтры при crash службы; native-only не обеспечивает fail-closed после crash. Удаление зависимости — отдельный этап |
| 19.10.2026 | Разрешены этапы (а) регистрация профилей при запуске, (б) native новых соединений; удаление диалога/quit/перезапуска (в) только после проверки владельца | Рабочий PAC остаётся прежним до этапа (в); регистрация не выдаёт readiness/PASS. Existing TCP egress не меняется, ECH/DoH не разрешают guessed DIRECT |
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
  Windows build/native bursts/accelerated lease и финальные PS/IPC checks PASS; весь Verify завершён success.
- Локально: `go test -race -timeout 120s ./...` с отдельной PostgreSQL 17 PASS, `go vet ./...`,
  Linux build и Windows amd64 build/vet PASS; native test binaries cross-compiled, deployment Python 21 PASS.
  Тестовый контейнер удалён; ни один command не обращался к production VPS.
- `github_fetch_commit_workflow_runs` фильтрует PR events: его пустой ответ НЕ доказывает отсутствие push CI.
  Читать `/actions/runs?head_sha=...` через read-only GitHub fetch; CLI gh в текущей среде получает Forbidden.


## (m) Этап (а): автосоздание профилей при запуске GUI, 19.10.2026

- Код `c556ac6`: после первого успешного IPC-подключения GUI регистрирует Chrome/Edge
  без кнопок, записи PAC, запуска/закрытия браузеров и включения Optimize.
- Chrome DIRECT до старой PAC-активации; новый Edge AUTO / `edge-sites-pending`.
  Существующие режимы, пути, правила, identity и DIRECT+ сохраняются. Повторные вызовы
  сериализованы, отсутствие браузера не создаёт фиктивный профиль, optional IPC fail
  не очищает таблицу. Создание отсутствующего Edge при уже включённой оптимизации
  отложено (`ACTIVE_CAPTURE_PENDING_EDGE`), чтобы pending fail-closed не ломал браузер
  при простом открытии GUI; после выключения Optimize повтор не чаще раза в 10 с.
- Регистрация диагностируется в GUI и bounded owner-user `browser-startup.json`
  (время, профиль, режим, safe code; raw IPC ошибки и секреты не пишутся).
- `diag-no-vpn.ps1` 2026-10-19.2: IPC-профили и отчёт GUI доступны с `-Only browser,youtube`,
  PID/creation time существующих браузеров сравниваются до/после диагностики. Это не
  подтверждение принятия PAC открытым окном; headless остаётся отдельной проверкой.
- Локально: Linux build/vet, Windows amd64 cross-build, 21 deployment test PASS;
  полный `go test -race -p 1 -timeout 120s ./...` с изолированным PostgreSQL 17 PASS.
  Первый параллельный race дал timeout DNS/TLS recovery fixture; последовательный
  полный повтор прошёл без изменения Go-кода. Временный PostgreSQL удалён.
- .NET и Windows локально отсутствуют: C#/WPF локально НЕ компилировались. Добавлены
  C# regression tests и Windows-only запуск Chrome/Edge ДО регистрации с проверкой
  того же PID/start time. CI и Internet/ISP владельца — разные доказательства.
- PAC/диалог/quit/restore-last-session не удалены. Дельта `ChromeRouting.cs` сохраняет
  старую активацию для Chrome, который теперь зарегистрирован раньше, при старте GUI.
  Точки интеграции: `ChromeRouting.cs:26–29` (выбор созданного startup-профиля),
  `MainWindow.xaml.cs:116` (регистрация после успешного Refresh); restart-код не менялся.
  Механизм AUTO без PAC, enable-to-first-success, объединённые native/PAC relay counters
  и последующая googlevideo admission-политика — этап (б), не заявлены готовыми.
  Этап (в) ждёт отдельной проверки владельца. VPS/сертификаты/TLS/mTLS/PR #8 не менялись.
- Инструкция GUI-only обновления и границы проверки: [BROWSER_HOT_ATTACH.md](../BROWSER_HOT_ATTACH.md).
  Оставить Chrome/Edge открытыми, обновить/перезапустить только GUI, не нажимать PAC-кнопки
  для проверки регистрации; затем `& .\tools\windows\diag-no-vpn.ps1 -Only browser,youtube`.

### CI этапа (а)

- Полный код/скрипт `84793bf`: [Edge policy acceptance PASS](https://github.com/buzzmean/neexitlag/actions/runs/39229227048)
  (C# core, заранее работающие Chrome/Edge с теми же PID/start time, WPF, PS5.1/7 parser,
  Edge real policy/HTTPS/recovery). Native Chrome [PASS на `c556ac6`](https://github.com/buzzmean/neexitlag/actions/runs/39229047809),
  следующие изменения затрагивали diagnostics/docs, не Chrome runtime.
- [Verify `84793bf`](https://github.com/buzzmean/neexitlag/actions/runs/39229227087):
  полный PASS: Linux core и Windows, включая native WPF, Go, ускоренный lease и PS-регрессии.
  Финальный handoff после `84793bf` меняет только документы, не исполняемый код/скрипт.
- DNS recovery fixture после первого load-timeout дополнительно прошла `-race -count=3`
  без кэша; production Go и таймауты не менялись.
- Инструкция учитывает старую службу: если capability Edge <3, сначала обновить только
  локальную Windows-службу через build-e2e/update-service с SHA256-проверкой, затем GUI.
  Browser процессы закрывать не надо; соединения при service update могут прерваться.


## (n) Этап (б), 19.10.2026 — реализация и проверка

- `0e89144`, `fc75207`: native AUTO новых IPv4 TCP, тот же TLS/certificate/HTTP GET
  planner, DNS recovery и сетевой/профильный кэш. Opaque/ECH → RELAY_REQUIRED;
  внешний IP не угадывается по DNS/outer SNI. Нет replay или migration старого TLS.
- Решение владельца: owned mandatory PAC остаётся независимым crash-guard. Fresh
  без него автоматически не активируется (`NATIVE_OWNED_PAC_GUARD_REQUIRED`).
  Gate проверяется и для `profile.set`. Удаление guard — отдельный этап.
- `96b8a2f`, `b6792e5`, `673624a`: latency от начала IPC, first open/exchange только
  новых flow ID. Одних timestamp недостаточно: Windows может выдать одинаковое время.
  Backend prepare не сбрасывает старые TCP браузера даже до публикации native-профиля.
- `7d47e2d`: новая backend `ReconfigureNewConnections` использует прежний UDP fence,
  закрывает только старые UDP mappings выбранного браузера и обновляет cache epoch,
  **сохраняет все установленные TCP** и не запускает TCP-table reconnect sweep.
  Идемпотентный monitor не публикует/не сбрасывает UDP каждые 30 с.
- Edge DIRECT/AUTO `edge-sites-pending` → AUTO/native автоматически при готовых guard
  и capture. Ошибка nonce не даёт fake PAC READY: остаётся bound listener/owned policy
  или static fail-closed guard. Старые owner/recovery/config/identity сохранены.
- `701237f`, `6cf7e0b`: различаются отказ native IPC и auxiliary recovery/refresh;
  таблица приложений сохраняется. Явное «Отключить PAC» отключает native, сохраняет
  protected guard, не включает его снова таймером. Старые PAC/dialog/quit кнопки
  работают как opt-in rollback. Этап (в) **не начат**.
- Метрики: actual origin/opaque, PID/birth, route/reason/error, payload bytes;
  enable→first open/exchange — **не HTTP/page PASS**. Bounded 512 records / 128 clocks,
  IPC до 60 KiB; truncation/lost видимы. Native+PAC relay usage: active/opened TCP,
  bytes/minute по profile/domain; QUIC overhead/probes не входят. Opaque = `<opaque>`.
- Known googlevideo: hard 8, очередь 32 / 2 с, затем admission error без DIRECT.
  Opaque видео не выдумывается как googlevideo; действует общий pool 8×48.
- Реальные QUIC/IPv6 маршруты не менялись. Только сохранена IPv6 fail-closed проверка
  при new-only TCP-флаге; legacy interception/прочие приложения покрыты регрессиями.

### Верификация / оставшееся

- Локально: полный Go race с isolated PostgreSQL, Linux build/vet, Windows amd64
  cross-build и native platform test compilation PASS; Python deployment 21 PASS.
  .NET и Windows локально отсутствуют. В первых прогонах были существующие timing
  failures lease-revocation ACK и demo failover; повторный полный race PASS.
- Runtime `673624a`: [Chrome native](https://github.com/buzzmean/neexitlag/actions/runs/39238593704)
  и [Edge/C#/WPF](https://github.com/buzzmean/neexitlag/actions/runs/39238593692) PASS.
  Chrome уже запущен при native IPC enable: PID/creation unchanged, owned PAC сохранён.
  Это real controlled HTTPS/HTTP2/mTLS/driver fixture, не Internet PASS.
- `c2455a4` подтвердил отказ test target: `upstream_dial_failed / connectex refused`,
  independent dial тоже refused, Accept error отсутствует, relay pool имел свободную
  ёмкость. `f1a0363` изолирует echo destination в отдельный дочерний процесс; тот же
  единственный target и все 160/256 одновременно открытые sockets сохранены.
  Production capacity/timeouts не менялись; assertions no BLOCK/no DIRECT/continuity
  не ослаблены. Точная внутренняя причина Windows loopback refusal не доказана.
- Предыдущий [Verify `f1a0363`](https://github.com/buzzmean/neexitlag/actions/runs/39255204947)
  прошёл browser/driver/race/burst/lease/IPC, но упал на последнем Go-шаге:
  `TestUpgradeProofRealProductionLeaseImmediateRenewals` требует strictly increasing
  deadline при двух мгновенных renewals, а fresh ID формировался из часов.
  `b3dcd1c` меняет только fixture: ID по счётчику и 20 мс между ticket issues
  (ограничено context). Все assertions сохранены; production 30-min lease,
  security/replay и upgrade tool runtime не меняются.
- [Verify `b3dcd1c`](https://github.com/buzzmean/neexitlag/actions/runs/39257168380):
  Linux полный race/vet/build/cross-build и Windows native Go/race/bursts,
  C#/WPF, desktop smoke, PowerShell и ускоренная lease-проверка PASS.
- [Relay release](https://github.com/buzzmean/neexitlag/actions/runs/39257168514) и
  [isolated upgrade](https://github.com/buzzmean/neexitlag/actions/runs/39257168393) PASS.
- Финальный handoff — docs-only commit через GitHub API после потери соединения
  со средой разработки; код совпадает с проверенным `b3dcd1c`. Локальный checkout
  следующей сессии сначала синхронизировать `git fetch origin` /
  `git merge --ff-only origin/codex/work` (без rebase/force).
- Существующий PAC может обслуживать новые запросы без native records: это PAC path,
  **не native interception PASS**. `RefreshPolicy(false)` означает user GP request,
  не live uptake. Cached dead PAC открытого Edge требует проверки на ПК владельца.
  Старое HTTP/2-соединение остаётся pinned; reload может его повторно использовать.

### Обновление Windows

Точный stop-on-error блок: `docs/BROWSER_HOT_ATTACH.md`, раздел «Обновление и короткая
проверка без закрытия Chrome/Edge». Закрыть только GUI; elevated owner PowerShell:
`build-e2e.ps1`, `build-desktop.ps1`, `update-service.ps1`, SHA256; обычный owner:
`install-desktop.ps1`, открыть VELIQ, Optimize, обновить страницы в уже открытых окнах.
Проверка: `& .\tools\windows\diag-no-vpn.ps1 -Only browser,youtube` (версия 2026-10-19.3).
VPS/credentials/certificates/PR #8/ветка direct-plus-flowseal не менялись.


## (o) Игровой AUTO → DIRECT, 19.10.2026

- Данные владельца: CS 120 мс с Optimize / 25 мс без; scoped capture показал
  28025 UDP packets через VPS-01, source=profile. Владелец подтвердил **mode=AUTO**.
  Встроенный cs2 AUTO введён в `fa7cdc1`; global optimizer выбирал релей по probe
  target, не игровому серверу. Это подтверждённый путь ошибочного relay, а не
  доказанная задержка WinDivert. Создателя возможного local override старые данные
  не хранят: нельзя приписывать его ручному действию GUI только по source=profile.
- `50f750f`: свежий cs2 DIRECT (Steam уже DIRECT); старые игровые/неизвестные AUTO
  игнорируют global relay/backups и generic learned relay, включая защиту в engine
  selector. Профили/кэш не удалены. MANUAL и явные rules остаются opt-in;
  browser per-domain, OpenAI/Anthropic/Telegram/Discord protected политики сохранены.
- `060bc19`: обычное добавление из процессов/файла создаёт DIRECT с полным путём.
  GUI подтверждает MANUAL предупреждением о возможном росте пинга; «Нет» не меняет
  профиль и не вызывает IPC. Browser startup — отдельный согласованный путь.
- `2815bd0`: optional read-only IPC profiles.routing (built_in/local_override/custom,
  policy defaults, bounded 64); diagnostics 2026-10-19.4 раздел «Процессы с профилем
  RELAY», реальный scoped flow PID/birth/group, explicit truncation/lost. Старый
  сервис помечает AUTO UNKNOWN_LEGACY_AUTO, не выдаёт ложный DIRECT/PASS.
  PS5.1/7 StrictMode fixtures, C# регрессии и исправленные expected DIRECT включены.
- Файлы: core/profiles/profile.go; services/client-service/{automatic_relay*,controller,
  flow_policy,local_strategy_test,controller_test}.go; desktop-core DesktopViewModel,
  RelayWarning и ресурсы; desktop MainWindow.xaml.cs/RelayWarning; tests/desktop;
  diag-no-vpn.ps1, tests/windows-tools/profile-routing-checks.ps1, verify.yml,
  docs/GAME_DIRECT.md. WinDivert/queues/relay runtime не менялись.

### Проверки

- Локально: targeted Go race service/profiles PASS, go vet, Linux build и Windows
  amd64 cross-build PASS; отдельные route-test/demo race PASS. Полный локальный
  race в нескольких прогонах упирался в существующий demo 25-second route-selection
  timeout под нагрузкой; чистый полный локальный PASS не заявляется. Таймауты и
  assertions не ослаблялись. В Linux CI полный race прошёл.
- .NET/PowerShell/Windows локально отсутствуют; C#/WPF локально НЕ компилировались.
  Проверки выполняются Windows CI, а не выдаются за тест на ISP владельца.
- Проверенный код: `2815bd04ad46669f8504780039cfce530094fadf`.
- [Verify](https://github.com/buzzmean/neexitlag/actions/runs/39266607496):
  Linux core и Windows PASS: полный race/vet/build/cross-build, C#/WPF,
  PS5.1/7, native driver/Chrome, 160/256 bursts, ускоренный lease, IPC/CLI checks.
- [Edge policy/C#/WPF](https://github.com/buzzmean/neexitlag/actions/runs/39266607404),
  [relay release](https://github.com/buzzmean/neexitlag/actions/runs/39266607426),
  [isolated upgrade](https://github.com/buzzmean/neexitlag/actions/runs/39266607408) PASS.
- Финальный handoff меняет только документы; следующий checkout обновлять без
  rebase/force. VPS, credentials, сертификаты, PR #8 и direct-plus-flowseal не менялись.

### ПК владельца

Сразу на установленной версии: «Приложения» → Counter-Strike 2 → DIRECT → «Применить»;
Steam и остальные игровые профили оставить DIRECT. Встроенный профиль не удалять.
Не менять Optimize/браузеры/защищённые приложения. Смена UDP egress может требовать
переподключения к игровому серверу, не перезапуска остальных приложений.
Обновлять службу и GUI штатными build-e2e/build-desktop/update-service/install-desktop
под прежним owner SID, проверять SHA256 службы; Chrome/Edge закрывать не требуется.
Обычный PowerShell, C:\neexitlag во время матча:
`& .\tools\windows\diag-no-vpn.ps1 -Only apps,ipc`.
Сравнить mode/default и фактический игровой flow DIRECT, затем пинг с Optimize on/off.
Новый реальный пинг и отсутствие задержки WinDivert пока не измерены.


## (p) PAC/native и Edge readiness, 19.10.2026

### Подтверждённая причина и граница выводов

- Игры на ec72beb: владелец подтвердил cs2 relay_flows=0, игровой DIRECT; пинг-баг закрыт.
- Владелец подтвердил новые вкладки Chrome/Edge без restart на 4ba1bfe, но не native
  interception. Online PAC направляет все origins в CONNECT/loopback: внешнего SYN
  браузера нет, поэтому native timing мог оставаться NOT_OBSERVED при работающих сайтах.
  Исторический GP refresh LIVE_UPTAKE_NOT_OBSERVED не является live evidence.
- Native Edge сохранял pending preset; GUI заменял PAC snapshot пустым native snapshot.
  Это отдельные ошибки состояния/отображения, исправлены без выдуманного PAC nonce READY.
- Записи 35 UPSTREAM_OPEN_TIMEOUT / 3 OPEN_FAILED **не сохранились**, владелец это
  подтвердил. Достоверно назвать их домены или приписать все одному дефекту невозможно.
- Найден реальный дефект budget: header/PLAN/OPEN делили deadline, длительный PLAN
  оставлял OPEN и stall retry истёкший контекст. Теперь бюджеты разделены и ограничены.

### Runtime и диагностика

- df72208 / 82d0700 / 50237a9: cumulative PAC counters/clocks, bounded failure_details
  domain/profile/PLAN-or-OPEN/route/destination/source/code/count/last_at; stage durations,
  typed relay/ACK/authorization errors, исправлен двойной UPSTREAM prefix, IPC bounds.
  OPEN max 3 с; stall retry — собственный bounded budget от lifetime этого соединения.
  Это не 3-секундный lifetime established socket и не неограниченный retry.
- 0ef71c3 / 99b7c21 / e41ae43: exact reserved service upstream owner-validated CONNECT
  делегируется в native WinDivert OPEN worker. Известный план открывается до SYN ACK,
  без ожидания второго ClientHello и без повторного планирования/probes. Исходные ошибки
  actual OPEN возвращаются frontend вместо искусственного kernel-connect timeout.
- ff093e1: отказ frontend handoff закрывает только его exact unaccepted upstream;
  accepted/foreign/legacy flows не затрагиваются. Нет orphan, удерживающего Engine key.
- Метки PAC_CONNECT/PAC_RECOVERY, PAC_NATIVE_UPSTREAM, NATIVE_TCP раздельны. Первые
  две — ingress, третья — native upstream того же запроса. Нельзя суммировать их как
  два application request или объявлять hybrid native-only без PAC.
- 530d8ac / b4be6b8 / 1e86b94 / 1322d21: GUI сохраняет PAC records при пустом native,
  разделяет Chrome/Edge evidence, BLOCK отображается BLOCK; probe error не page FAIL.
- Native Edge после успешной публикации edge-sites-pending → edge-sites; status
  transport_mode/native armed/actual observation отдельно от pac_nonce_readiness.
  1223f25: upstream OPEN без обмена = NATIVE_UPSTREAM_OPENED_NO_EXCHANGE; только
  actual bidirectional bytes дают NATIVE_TUNNEL_OBSERVED. Page PASS из байтов не выводится.
  owned_guard_installed отдельно от static rollback_guard_active.
- diag 2026-10-19.5: отдельные PAC/native counters, first OPEN/exchange и scope часов,
  ingress, Edge state, failure aggregates по доменам, loss/truncation. Старые сервисы
  не получают выдуманные метрики. Native -1/отсутствующая дата = NOT_OBSERVED.
- mtalk/stats/update endpoints могут быть фоновыми/non-HTTP. Ошибка альтернативного
  origin probe не доказывает отказ реального CONNECT/страницы. Повторяемый actual
  OPEN/BLOCK нужного ресурса требует проверки; GUI помечает источник, не скрывает отказ.

### Windows Firewall: найденная и проверенная причина handoff timeout

- Реальная driver-трасса: SYN перехвачен, actual upstream открыт, reflection send
  успешен, но listener SYN/ACK отсутствовал. Узкое TEST-NET/порт правило Windows
  Firewall восстановило handshake и nonce. После prepare listener меняет порт:
  fixture admission теперь привязана к текущему порту, не только объекту backend.
- Первый program-scoped эксперимент из временной Go test-директории был отвергнут
  API (87), затем удалён; это не доказательство причины. Итоговый production helper
  проверен отдельно на копии тестового EXE с нормальным полным installed-style путём.
- 1db6be1: update-service.ps1 -NativeTcpFirewall — явный opt-in. Правило только inbound
  TCP для полного пути установленного neexitlag-service.exe, edge traversal Block;
  не общий порт/все программы и не изменение VPS. Порт динамический, scope по EXE.
  WinDivert по-прежнему отбрасывает физический inbound на listener; accept требует
  exact native mapping. Никакого открытого proxy endpoint не добавлено.
- Helper проверяет owned marker/scope и effective ActiveStore; foreign rules и GPO
  не обходятся. Новый rule удаляется при отказе обновления, старый бинарник/ACL
  откатывается; уже существующий проверенный rule сохраняется. Без флага updater
  firewall не меняет. Identity/IPC/owner SID/профили/PAC/сертификаты не меняются.

### Тесты и CI

- Локально: полный race изменённых browserproxy/client-service/platform пакетов PASS;
  go vet ./..., Linux build, Windows amd64 cross-build, native test binary compile PASS.
  Локально .NET/PowerShell/Windows нет — C#/WPF/WinDivert выполняются только CI.
- Полный локальный race ранее дал UDP impairment deadline и demo route-selection
  timeout. Isolated UDP повтор PASS, demo повтор timeout; эти пути/assertions не менялись.
  Актуальный Linux CI полный go test -race -timeout 120s ./... PASS.
- [Chrome policy native на 1db6be1](https://github.com/buzzmean/neexitlag/actions/runs/39301567566): SUCCESS. Same main PID/creation,
  fresh controlled HTTPS nonce + actual PAC_NATIVE_UPSTREAM OPEN/bidirectional bytes;
  simultaneous HTTPS/HTTP2 DIRECT, native DIRECT+ disorder, protected REQUIRED relay,
  blocked unknown relay, outage/recovery/restart. Не localhost как Internet PASS.
- В том же run TestNativeBrowserDriverHandshake реально проверяет production app-scoped
  helper на отдельной копии EXE, native TCP handshake/nonce и owned cleanup. Ни SCM,
  ни реальные credentials не устанавливаются. PS5.1/7 mocks проверяют opt-in, scope,
  idempotence, foreign/API/policy failure и rollback.
- [Edge policy acceptance на 0789d5d](https://github.com/buzzmean/neexitlag/actions/runs/39299854054): SUCCESS (actual Edge policy/nonce,
  HTTPS/outage/restart); status/GUI Edge/native regressions включены в текущий Verify.
- [Verify 39301567351](https://github.com/buzzmean/neexitlag/actions/runs/39301567351): SUCCESS core + windows. Полный Go race/vet, Linux/Windows builds, PostgreSQL regressions, 130 C# tests, WPF/native GUI smoke, repeated TCP bursts, accelerated lease, actual Chrome/WinDivert и IPC/tool regressions PASS.
- [Release 39301567455](https://github.com/buzzmean/neexitlag/actions/runs/39301567455) / [isolated upgrade 39301567464](https://github.com/buzzmean/neexitlag/actions/runs/39301567464): SUCCESS.
  Это изолированные CI, не deployment. VPS не трогался.

### Передача владельцу

Этап (в) НЕ начат: PAC, диалоги и legacy quit/recovery сохранены; fresh без owned
PAC guard автоматически не активируется. Deferred SNI/boundary/ECH policy в итоговой
дельте не ослаблены. Existing TCP/H2 не мигрируют и могут штатно переиспользоваться.
VPS, credentials/сертификаты, game/UDP policy, фильтры, default route, PR #8 и ветка
codex/direct-plus-flowseal не менялись. Нового ISP PASS Chrome/Edge нет.

Обновлять службу и GUI под прежним owner SID, закрыть только GUI, браузеры оставить
открытыми. Сборки и полный блок — docs/BROWSER_HOT_ATTACH.md; elevated updater теперь
вызывать с явным -NativeTcpFirewall. Сверить SHA256 installed/built service и Running.
Затем обычный owner PowerShell в C:\neexitlag после новых вкладок/обновления страниц:
`& .\tools\windows\diag-no-vpn.ps1 -Only browser,youtube`.
