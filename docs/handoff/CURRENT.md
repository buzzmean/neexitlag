# CURRENT — состояние проекта (обновляется в конце каждой сессии)

Последняя передача Codex 20.10.2026: коды активации/CSR после freeze и leaf timer реализованы, область освобождена; результаты и команда — (af). Production не обновлён: SSH из среды недоступен, реальные коды не созданы.

Обновлено: 20.10.2026, Codex; `codex/work`, база 6d91aea (fast-forward сайта Claude сохранён; browser работа от 0d20db1). Владелец подтвердил TLS/HTTP отказ исходного Telegram IP с VPS и успех независимого кандидата с тем же SNI. Реализован bounded per-IP first-flight recovery для relay-first, negative TTL и диагностика .2; детали — (ab), docs/TELEGRAM_TLS_DIAGNOSTICS.md. На production VPS ничего не выполнялось; Владелец подтвердил per-IP recovery на ПК. Этап (в): вариант A, тихий owned PAC без browser restart, реализация и ограничения — (ac); PR #8 не тронут.

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
| Cold-start browser retry / native-only audit | `core/browserproxy/{proxy,stall,path_metrics}*.go`, `core/browserproxy/*_test.go`, `services/client-service/browser_*.go`, `services/client-service/controller.go` (read-only browser IPC), `platform/windows/{tcp_capture_windows,tcp_state,browser*,native_firewall*}.go`, `platform/windows/native_firewall_probe.ps1` (embedded read-only probe), Windows browser tests, `internal/browserfixture/*` (cold-start fixture), `tools/windows/diag-no-vpn.ps1`, `tests/windows-tools/*`, `.github/workflows/{chrome-native,verify}.yml` (tests), `docs/BROWSER_HOT_ATTACH.md`, `CURRENT.md` | Codex | `codex/work` | ГОТОВО 19.10.2026, файлы освобождены: ae4fac1/f21bf68/bc79d6c; race gate 659c865. Cold retry и actual Windows regression PASS, firewall read-only и честные native-only counters; все-site native-only остаётся security blocker при PAC как единственном crash-guard. CI upload quota, подробности (q); этап (в) не начат |
| Relay-first подтверждённых browser групп / first-byte timing | `core/flowpolicy/direct_timeout*.go`, `core/browserproxy/{proxy,stall,path_metrics}*.go`, новые group/timing файлы и тесты, `services/client-service/browser_*.go`, `services/client-service/controller.go` (только read-only timing IPC), `services/client-service/stall_end_to_end_test.go`, `internal/browserfixture/fixture.go` (только SAN тестовых origins), `tools/windows/diag-no-vpn.ps1`, `tests/windows-tools/browser-path-checks.ps1`, `platform/windows/browser_cold_windows_test.go`, `.github/workflows/chrome-native.yml` (тесты), `docs/BROWSER_HOT_ATTACH.md`, `CURRENT.md` | Codex | `codex/work` | ГОТОВО 19.10.2026, область освобождена: 5a3bd24 / 5f16ed5 / 432899d. Group relay-first, scoped timing IPC и actual Windows холодный benchmark PASS. PAC/лимит 8/VPS/game policy без изменений; общие CI упираются в artifact quota, ISP ретест владельца ещё нужен |
| Telegram/Anthropic browser relay-first и no_reply причины | `core/flowpolicy/browser_group.go`, `core/browserproxy/{proxy,first_byte}*.go`, `core/browserproxy/path_metrics_test.go` (только ожидание обоих I/O counters), новые browser close-evidence файлы/тесты, `services/client-service/browser_*.go`, `tools/windows/diag-no-vpn.ps1`, `tests/windows-tools/browser-path-checks.ps1`, `.github/workflows/chrome-native.yml` (тесты), `docs/BROWSER_HOT_ATTACH.md`, `CURRENT.md` | Codex | `codex/work` | ГОТОВО 20.10.2026: 5f413c6 / 88a6f61 / 5c0d2ec, async test fix 5867936; область освобождена. Actual Chrome/Edge tests PASS, CI upload quota; ISP/no_reply breakdown ждёт короткой проверки владельца. PAC/guard/owner-check/TLS/игры/лимит 8 сохранены |
| Fresh-PC CSR enrollment / подготовленная CA migration | `core/freshpc/*` (новые), `tools/fresh-pc/*` (новые), `services/control-plane/fresh_enrollment*.go` (новые), `store.go` (выделение publication transaction helper), `tools/windows/bootstrap-fresh-pc*.ps1` (новые), `deploy/vps/{rotate-fresh-pc-ca,fresh-pc-ssh}.py` (новые), `apps/desktop/App.xaml.cs` (bootstrap entry only), `apps/desktop/FreshPCSetup.cs` (new, reuse existing PAC transaction), `tests/desktop-native/{Program,FreshPCSetupChecks}.cs`, `tests/windows-tools/fresh-pc-checks.ps1`, `.github/workflows/verify.yml` (early tests), tests/CI fresh enrollment, `docs/FRESH_PC_ENROLLMENT.md`, CURRENT.md | Codex | `codex/work` | ГОТОВО 20.10.2026; область освобождена. По требованию владельца один итоговый commit, ownership включён в него. Только isolated tests; VPS/veliq-swe-01/PR #8 не тронуты; production Windows/SSH/systemd не проверены |
| Дизайн VELIQ: дизайн-система, сайт, макеты программы, бренд-материалы | только новые папки `design/`, `website/`, `app-ui/`, `brand/`; строка в `CURRENT.md` | Claude Code | `codex/work` | ГОТОВО 20.10.2026, область освобождена (в т. ч. улучшения сайта: страница vpn.html, блок «Что видит сервер», клубы, выбор устройства на загрузке). Код программы, службы и VPS не менялся. Описание и допущения — `design/README.md`; перенос в WPF и новая иконка в `apps/desktop/Assets/` ещё не сделаны |
| Fresh-PC: UAC-disabled execution | `tools/windows/bootstrap-fresh-pc.ps1`, новый `tools/windows/fresh-pc-execution.ps1`, `tests/windows-tools/fresh-pc-checks.ps1`, `docs/FRESH_PC_ENROLLMENT.md`, CURRENT.md | Codex | `codex/work` | ГОТОВО 20.10.2026; область освобождена; один итоговый коммит по требованию владельца. UAC-off поддержка и isolated tests; actual Windows UAC-off/full install не проверены; production install/enroll не запускались |
| Fresh-PC: Default linked-query regression | `tools/windows/{fresh-pc-execution,bootstrap-fresh-pc}.ps1`, `tests/windows-tools/fresh-pc-checks.ps1`, `docs/FRESH_PC_ENROLLMENT.md`, CURRENT.md | Codex | `codex/work` | ГОТОВО 20.10.2026; область освобождена; один коммит по требованию владельца. Exact owner facts + Default query errors covered; actual owner Windows retest не выполнен; без VPS |
| Fresh-PC: DWORD token ABI fix | `tools/windows/fresh-pc-execution.ps1`, `tests/windows-tools/fresh-pc-checks.ps1`, `docs/FRESH_PC_ENROLLMENT.md`, CURRENT.md | Codex | `codex/work` | ГОТОВО 20.10.2026; область освобождена; один итоговый commit по требованию владельца. ABI/reflection/result-validator и exact owner fixture PASS; actual owner Windows не проверен, без VPS |
| Fresh-PC: VPS binary preparation without old identity | `deploy/vps/{fresh-pc-ssh,fresh-vps-upgrade,production-upgrade,rotate-fresh-pc-ca}.py`, `tools/windows/bootstrap-fresh-pc.ps1`, `tools/fresh-pc/*`, deployment/PS/Go tests, docs/FRESH_PC_ENROLLMENT.md, CURRENT.md | Codex | `codex/work` | ГОТОВО 20.10.2026; область освобождена. Один итоговый commit по требованию владельца. Только isolated tests; production SSH/upgrade/rotation не выполнялись |
| Fresh-PC PAC readiness | `apps/desktop{,-core}/Fresh*`, `ChromeRouting.cs`, `BrowserNative.cs`, browser readiness/status/errors, `tools/windows/bootstrap-fresh-pc.ps1`, desktop/Go/PS tests, FRESH_PC_ENROLLMENT.md, CURRENT.md | Codex | `codex/work` | ГОТОВО 20.10.2026; область освобождена; один итоговый commit по требованию владельца; локальные проверки, реальный Windows/ISP ретест ещё нужен; VPS/CA/enroll не запускались |
| Этап (в): независимый guard / проверка реализуемости | `docs/BROWSER_NATIVE_GUARD.md` (новый), `docs/BROWSER_HOT_ATTACH.md`, CURRENT.md; runtime этапа (в) до подтверждения guard не менять | Codex | `codex/work` | BLOCKED_CURRENT_BACKEND 20.10.2026; аудит завершён, область освобождена; только docs в одном commit по пункту 4 владельца; runtime этапа (в) не менять до реализации/проверки независимого guard |
| Telegram main origin: безопасная TLS-диагностика и read-only VPS proof | `core/browserproxy/{tls_wire*,proxy,close_evidence}.go`, `services/client-service/browser_native{,_pac}.go`, новый `browser_tls_test.go`, `platform/windows/telegram_hello_test.go`, `tests/integration/telegram_tls_test.go`, `tools/windows/diag-no-vpn.ps1`, `deploy/vps/diagnose-telegram-tls.py`, deployment/PS tests, docs/TELEGRAM_TLS_DIAGNOSTICS.md, CURRENT.md | Codex | `codex/work` | ГОТОВО 20.10.2026; область освобождена; один итоговый commit по требованию владельца; read-only script подготовлен, на VPS не запускался; Internet причина/Windows ретест остаются открытыми |
| Relay-first: per-IP TLS silence recovery | `core/browserproxy/{origin_*,proxy,dns,dns_test,tls_wire,tls_wire_test}.go`, `services/client-service/browser_{transport,admission,known_group,native,native_pac,origin,origin_test}.go`, `tests/integration/relay_origin_test.go`, PS tests/diag-no-vpn.ps1, docs/TELEGRAM_TLS_DIAGNOSTICS.md, CURRENT.md | Codex | `codex/work` | ГОТОВО 20.10.2026; область освобождена; один итоговый commit по требованию владельца; actual Windows/ISP ещё не проверен; VPS не тронут |
| Этап (в), вариант A: тихий owned PAC без browser restart | `services/client-service/browser_silent*.go`, browser/edge transport/controller; `platform/windows/*policy*`; desktop Browser/Chrome/Edge/Fresh routing; bootstrap/update/diag; acceptance/tests/docs | Codex | `codex/work` | ГОТОВО 20.10.2026; область освобождена; один итоговый commit включает ownership по требованию владельца; WFP/VPS/игры вне задачи; Windows uptake не измерен локально |
| Windows freeze-фикстура: настоящий SID + post-push CI handoff | `services/client-service/browser_silent_test.go`, `docs/BROWSER_SILENT_POLICY.md`, CURRENT.md | Codex | `codex/work` | ГОТОВО 20.10.2026; область освобождена; дополнительный commit явно разрешён владельцем; production runtime без правок |
| Коды активации club/personal, повторный CSR после freeze, leaf renewal | `core/enrollment/*`, control-plane/migrations/API/store, client-service activation/manifest/startup, desktop Activation*, tools/codes, fresh hardware adapter, deploy/vps certificate renewal + tests/docs | Codex | `codex/work` | ГОТОВО 20.10.2026, область освобождена; 4cdf2e5 / 82f6f84 / 7ea0769 / b84bdb3 / 0d3883b. Local checks и focused Linux/Windows CI PASS; финальные CI ссылки ниже. SSH недоступен, production/codes не изменены; deployment передан владельцу; veliq-swe-01/PR #8 вне области |
| Отказ activation deployment после preflight | `deploy/vps/{production-upgrade,fresh-vps-upgrade,deploy-activation,activation-release-ssh,renew-server-cert,inspect-activation}.py`, `tools/windows/{activation-ssh,verify-activation-vps,update-activation-vps}.ps1`, deployment/PS tests, `.github/workflows/{activation,production-upgrade}.yml`, `docs/ENROLLMENT_CODES.md`, CURRENT.md | Codex | `codex/work` | ГОТОВО 20.10.2026, область освобождена; 70 Python tests и PS parsing/fresh checks PASS. Actual Windows redirected process/systemd matrix — в CI после push; production не выполнялся. Ownership включён в отдельный fix-коммит по требованию владельца |
| Отказ READ_ONLY_SNAPSHOT: причины, Windows CRLF/BOM, общий deployment guard | `deploy/vps/{activation-diagnostics,inspect-activation,production-upgrade,deploy-activation,activation-release-ssh}.py`, `tools/windows/{verify-activation-vps,activation-ssh}.ps1`, deployment/PS tests, docs | Codex | `codex/work` | ГОТОВО 21.10.2026; область освобождена; неверный default baseline (38 hex) воспроизведён и исправлен; безопасная общая диагностика, CRLF/BOM и rollback packaging покрыты тестами; VPS не выполнялся |
| Leaf renewal после activation deploy, повтор schema 2 и safe rollback | `tools/fresh-pc/{main.go,upgrade_test.go}`, `deploy/vps/{renew-server-cert,deploy-activation,activation-release-ssh,activation-diagnostics,inspect-activation}.py`, deployment/PS tests, workflows, docs | Codex | `codex/work` | РЕАЛИЗОВАНО 21.10.2026; область освобождена; CI после push; VPS не выполнять; security-monotonic rollback, схема 2 сохраняется явно; полный DB restore не реализован (см. (al)) |
| Ранний отказ CI и локальная воспроизводимая systemd/leaf проверка | `.github/workflows/{activation,production-upgrade}.yml`, `tests/deployment/*`, deployment fixes по найденным регрессиям, docs | Codex | `codex/work` | В РАБОТЕ 21.10.2026; VPS не выполнять; изолированный Docker/VM разрешён владельцем |
| Discord медленно, DIRECT+ для YouTube | не определены | никто | — | свободно для Codex после диагностики; Go-файлы застоя (выше) не трогать |
| PR #8, `codex/direct-plus-flowseal` | — | не трогать | — | только владелец |

Строки Claude со статусом «передано Codex» сохранены как история; старые «нужно да»/запреты области не отменяют разрешение владельца на A/B. Новую работу снова помечать перед правками по PROTOCOL.md. QUIC/IPv6 routing менять только после отдельного «да».


Новые ветки: `claude/<тема>`, `codex/<тема>`, база — `claude/serene-cori-msra1u`.

## (f) Сборка и установка службы на ПК (по CLAUDE.md / docs/AGENT_RULES.md)
Только по просьбе владельца, на его ПК:
1. Браузеры оставить открытыми; закрыть только GUI («Выход» в трее) для замены recovery EXE. Сохранить локальные правки; `git switch codex/work`, `git pull --ff-only origin codex/work`. Полный stop-on-error блок: [BROWSER_DIRECT_FALLBACK.md](../BROWSER_DIRECT_FALLBACK.md#обновление-windows-без-vps).
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

- 20.10.2026: владелец разрешил дополнительный test/docs commit после 5d47ff0. Windows freeze-фикстура получает SID фактического процесса через существующий discovery/token API, а не синтетический S-1-test; CI ошибки/неподтверждённый live uptake записаны без ложного PASS. Force-push/rewrite не используются.


- 20.10.2026: по явному решению владельца выбран вариант A этапа (в): quiet owned PAC + user policy refresh, никаких browser restart/dialog/process control. Независимый WFP guard — будущий B. Registry configuration и реальный owner ingress не смешиваются; pending browser uptake не валит bootstrap. Один итоговый commit включает ownership по требованию владельца.


- 20.10.2026: после owner proof Telegram DNS добавлен общий bounded origin-IP health/retry, не IP hardcode. Replay только полного ClientHello до первого origin byte; ранний alert/partial reply влияет лишь на следующие connections. DNS через HTTPS/relay, не новый VPS resolver API. Детектор на actual QUIC boundary отличает session/auth failures. Один итоговый commit включает ownership по требованию владельца.

- 20.10.2026, Telegram TLS: не подменять origin IP и не replay браузерный handshake на основании пяти байтов. Добавлена пассивная framing диагностика на PAC и actual relay boundary; вывод первых 16 байт маскирует payload. Владелец отдельно запускает read-only `.40`/`.99` TLS/GET proof; до этого причина не считается установленной. Один итоговый commit включает ownership по явному требованию владельца.

20.10.2026: этап (в) разрешён, UX с browser restart запрещён. Пункт 4 владельца:
при отсутствии доказанного независимого fail-closed guard остановить runtime,
не снимать PAC/owned checks и не маскировать проблему watchdog polling. Новый
проект WFP enforcement/redirect/signing/native failure injection нужен до выпуска;
подробности — [BROWSER_NATIVE_GUARD.md](../BROWSER_NATIVE_GUARD.md). Существующий
legacy restart path не принимается как временный пользовательский сценарий.

20.10.2026: после успешной установки identity восстановление browser lane — локальный
`-RepairBrowserReadiness`; никаких повторных RotateCA/enroll. Fresh-регистрация
использует DIRECT staging при активной оптимизации и требует AUTO + PAC/owner proof
для каждого установленного браузера. Native timer не вмешивается в PAC activation.
Старый диалог/штатный restart сохраняется, этап (в) не разрешён.
| Дата | Решение | Почему |
|---|---|---|
| 20.10.2026 | Explicit binary preparation перед CA: local VPS-only proof новой временной identity, persistent journal/revoke, backup + timer/boot recovery. Исправлен default_lease_seconds вместо lease_seconds | Потерянный PC bundle не нужен для server-side proof; прежний capability-check мог ложно обвинять уже совместимый relay |
| 20.10.2026 | Token classes 18/20/40: out int + sizeof(int)=4, linked token: отдельная pointer-sized структура; failed/error=24 не успех | Владелец подтвердил TokenElevation ERROR_BAD_LENGTH при IntPtr.Size=8 в прежнем adapter |
| 20.10.2026 | Default token: любая ошибка linked-query штатна; elevation/restriction/registry по-прежнему обязательны. Диагностика — фиксированные поля и точный reason | В b882d64 non-1312 linked-query исключение скрывалось generic отказом; владелец подтвердил Default/elevated/unrestricted/UAC-off |
| 20.10.2026 | Дизайн VELIQ в `design/`, `website/`, `app-ui/`, `brand/`; позиционирование общее: без числа серверов и ПК, сервисы как примеры, для игр — «оптимальный маршрут» | Решение владельца; тексты — `brand/copy.md` |
| 20.10.2026 | UAC-off: GUI может наследовать full-admin только при EnableLUA=0 + Default token без linked/restricted; служба в текущем процессе, SID/ACL прежние | Явное исключение владельца для freeze/Langame; при включённом UAC elevated bootstrap по-прежнему запрещён |
| 20.10.2026 | Fresh-PC: SSH-authorized CSR, root-only CA storage, hardware-slot replacement и atomic full publication; CA rotation только по явному confirmation | Старый bundle потерян; нельзя раздавать CA/admin token или молча перепровиживать VPS |
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


## (q) Cold-start relay retry и native-only — 19.10.2026

### Подтверждение владельца / границы

- e9e31ac: новые вкладки в открытых Chrome/Edge работают, Chrome PID прежний,
  игры DIRECT. Это подтверждённый ISP результат предыдущей сборки.
- Текущая сборка не объявляется новым ISP PASS. Owner diag не коммитился.
- PAC остаётся единственным независимым crash-guard; этап (в) не начат.
  Online PAC вызывает CONNECT, поэтому browser native-only SYN не возникает.
  PAC DIRECT для всех сайтов после crash снял бы protected fail-closed.
  Нужен отдельный устойчивый kernel/WFP guard; watchdog/реестр не атомарная замена.
  Подробный безопасный следующий этап — docs/BROWSER_HOT_ATTACH.md.

### Изменения маленькими коммитами

- ae4fac1: full Close planned CONNECT теперь синхронно освобождает actual native
  upstream/Engine tuple. Exact delegation generation не даёт закрыть чужой/заменённый
  поток; half-close и pinning прежние. Старое поведение с silent origin сохраняло
  Engine key до ответа/FIN и давало ErrFlowExists при relay retry того же CONNECT.
- Остаточная коллизия получает LOCAL_FLOW_RESERVATION_IN_USE, не open_io_failed,
  и не считается отказом VPS/сети/качества. Routing/security/tickets не изменены.
- f21bf68: native_only / pac_native_upstream раздельные per-activation счётчики,
  first OPEN/exchange clocks; NATIVE_ONLY_OBSERVED только actual NATIVE_TCP с
  двусторонним обменом. Hybrid не создаёт этот статус; legacy = uninstrumented.
  Суммы переживают bounded eviction, IPC прежнего ограниченного размера.
- bc79d6c: browser.native.firewall — служба read-only проверяет installed EXE,
  owned PersistentStore/ActiveStore и GPO; query failure = UNKNOWN, не PASS/MISSING.
  diag 2026-10-19.6 выводит фактическое правило и ingress отдельно. Проверка не
  вызывается из OPEN/capture и сама не меняет firewall. На ПК владельца актуальное
  APPLIED ещё нужно прочитать; cold collision в CI доказана даже при APPLIED.
- 659c865: новые lifecycle/ingress unit race-тесты выполняются отдельным ранним
  Windows gate, не теряются за build/artifact шагами полного Verify.

### Проверки и CI

- Локально Go 1.25.3: go test -race -timeout 120s ./... PASS (включая integration
  и demo), go vet ./... PASS, Linux build PASS, Windows amd64 cross-build PASS,
  native Windows test binary compile PASS. Локально Windows/.NET/PowerShell нет.
- [Chrome native на bc79d6c](https://github.com/buzzmean/neexitlag/actions/runs/39307168170):
  actual driver, embedded firewall APPLIED, cold-cache same-CONNECT retry и actual
  Chrome policy/HTTPS/HTTP2/outage/recovery PASS. Пять холодных origins дают verified
  TLS+HTTP nonce, 5 stalls / 5 successful retries / 0 failed retries / 0 local collisions;
  protected no DIRECT. Старый kernel-only Close воспроизвёл ErrFlowExists.
  Run целиком FAILURE только из-за upload-artifact storage quota.
- [Edge на bc79d6c](https://github.com/buzzmean/neexitlag/actions/runs/39307168085):
  C# regressions/WPF, real Edge policy/HTTPS/relay/recovery и Chrome/Edge owner race
  PASS. Run целиком FAILURE на artifact quota. Это controlled origins, не ISP PASS.
- [Verify bc79d6c](https://github.com/buzzmean/neexitlag/actions/runs/39307168003):
  PS5.1/7 browser-path и native-firewall regressions PASS; desktop-native UI PASS.
  Windows job остановился на artifact quota до последующих Windows Go/lease шагов:
  их нельзя считать выполненными этим run. Linux full race с PostgreSQL, vet и оба fuzz шага PASS.
- [Isolated upgrade bc79d6c](https://github.com/buzzmean/neexitlag/actions/runs/39307168304): SUCCESS.
- [Новый Windows race gate / Chrome](https://github.com/buzzmean/neexitlag/actions/runs/39307613274),
  [Verify](https://github.com/buzzmean/neexitlag/actions/runs/39307613031) на 659c865:
  новый Windows race gate PASS (planned full-close/foreign generation, ingress proof,
  local collision classification); actual Chrome/WinDivert/cold-cache повтор PASS,
  embedded firewall APPLIED. Run целиком FAILURE только upload quota. В Verify
  на bc79d6c оба job FAILURE на upload quota, Linux full race/vet/fuzz PASS.
  Смотреть test steps и общий вывод отдельно; artifact quota не исправлена.
  Workflow assertions/upload не ослаблялись; старые artifacts не удалялись.
- Release bc79d6c дополнительно получил CDP/INSTRUMENTATION_FAILURE при запуске
  Linux Chrome (DevToolsActivePort не появился за 25 с); тот же полный набор race
  прошёл в Verify и локально. Это отказ инструментирования до browser traffic,
  не доказательство ошибки маршрута; повторный release на 659c865 выполняется.

### Журнал решений / передача

19.10.2026: не превращать PAC_NATIVE_UPSTREAM в native-only доказательство. При
нынешнем единственном PAC guard полный native-only auto не активировать: fail-closed
важнее нового имени статуса. Архитектуру независимого guard вынести отдельно.
19.10.2026: закрытие planned native upstream должно быть синхронным до retry, а не
ждать remote FIN. Причина воспроизведена при действующем app-scoped firewall;
увеличение timeout/смена VPS/повторное ослабление firewall её не исправляют.
19.10.2026: исчерпание GitHub Artifact storage явно сохраняется как CI blocker;
не удалять прежние сборки и не делать upload continue-on-error ради зелёного badge.

Область Codex освобождена; Chrome/Edge/game policy, фильтры, UDP, TLS/mTLS, VPS,
сертификаты, credentials, production lease, PR #8 и direct-plus-flowseal не трогались.
Для новой сборки повтор нужен только короткий: Chrome/Edge не закрывать, только GUI;
обновить службу тем же owner SID по docs/BROWSER_HOT_ATTACH.md с -NativeTcpFirewall.
GUI код не менялся; переустановка GUI для этих исправлений не обязательна.
Обычный PowerShell владельца из C:\neexitlag после новых вкладок/обновления страницы:
`& .\tools\windows\diag-no-vpn.ps1 -Only browser,youtube`.


## (r) Первая загрузка подтверждённых browser групп — 19.10.2026

### Причина и изменение

Владелец ea4afdf: сайты открываются, но 21 DIRECT_STALL / 21 relay retry — время
первого запроса не сократилось. Fix full Close лечил потерю запроса, не ожидание.
Ключ timeout cache оставался exact hostname (исключение — googlevideo), поэтому
новые CDN/YouTube имена повторяли DIRECT. Кроме того, learned confirmed DIRECT
мог обходить timeout cache; pending DIRECT plans не перепроверяли его перед OPEN.

- 5a3bd24: фиксированная `youtube-media` группа с canonical cache host youtube.com
  (youtube/googlevideo/ytimg/ggpht и явные YouTube aliases); не весь Google/Google APIs.
  Первый реальный silent ClientHello подтверждает группу только для network/profile/
  port. Обычные домены exact-host, старый two-distinct-address механизм сохранён.
  Кэш раньше learned DIRECT + повторная проверка непосредственно перед OPEN.
- Relay REQUIRED, без DIRECT fallback; если relay недоступен, нет попытки DIRECT.
  Explicit rules, working configured DIRECT+, DNS/53 и opaque/protected semantics
  сохранены. Более широкая CDN failure group не расширяет DIRECT+ eligibility.
- Один background probe на ключ после 10 мин, глобально прежние 4 probes/3 с.
  В группе хранится фактически отказавший probe_host; новый TLS/HTTP GET recheck
  проверяет его, а не случайный рабочий sibling. Network/owner/Optimize проверяются
  до/после I/O. Без verified success cached traffic не возвращается DIRECT.
- 5f16ed5: bounded first-byte rows и cold TLS/HTTP benchmark. Фикстура содержит
  только дополнительные SAN в эфемерном тестовом сертификате, не production cert.
- 432899d: optional timings вынесены в `browser.first-byte` / `edge.first-byte`;
  `browser.status`/`edge.status` не расходуют routing evidence budget на новую сводку.
  Первый CI обнаружил truncation из-за смешивания этих данных; assertions не
  ослаблялись. Новая регрессия проверяет независимый bounded report и полный status.
- diag 2026-10-19.7: по network/profile/group/port/PAC ingress режимы DIRECT_FIRST,
  RELAY_FIRST и DIRECT_PLUS; replies/no_reply, avg/min/max/last и observed before/after.
  Отсутствующие samples = NOT_OBSERVED, старый IPC = UNAVAILABLE/UNINSTRUMENTED.
  Измеряется frontend TCP accept → первый upstream byte (включая TLS), не HTTP
  page load/скорость скачивания. PAC native upstream не дублирует frontend sample.

### Ограничения / безопасность

Volatile cache не переносится между запусками службы. Полностью неизвестная сеть
с пустым кэшем требует первого доказательства; автоматического ISP seed нет.
Профили Chrome и Edge независимы. Уже открытые до verdict TCP остаются pinned;
новая группировка применяется к новым OPEN, а не мигрирует существующие потоки.
Предусмотренный ClientHello-only retry до ответа сервера не расширялся на TLS
application/early data. PAC/crash-guard остаётся, этап (в) не начат.

Googlevideo 8 active на ПК, 32 waiters, максимум 2 с; затем явная admission error,
не девятый tunnel и не DIRECT. Session pool, byte/minute/per-domain counters и
probe reserve без увеличения. Это не глобальное ограничение 36 ПК: реальная
нагрузка VPS 1 vCPU не измерялась, при массовом видео остаётся риск CPU/трафика.
Игры/UDP policy, credentials, production сертификаты/VPS, PR #8 и protected-ветка
не менялись. Код GUI не менялся; обязательной переустановки GUI нет.

### Проверки

- Финальный локальный `go test -race -p 1 -timeout 120s ./...` PASS, `go vet ./...`
  PASS, Linux build PASS, Windows amd64 cross-build PASS, native test binary
  compile PASS. Локально нет Windows/PowerShell/.NET; native/PS/C# проверены CI.
- [Chrome/native/race на 432899d](https://github.com/buzzmean/neexitlag/actions/runs/39313482881):
  test steps PASS. Actual signed WinDivert, mTLS/tickets/relay, cold-cache TLS+HTTP
  nonce, unchanged running Chrome PID/creation, protected no DIRECT, HTTP/2,
  configured DIRECT+ и outage/recovery/restart прошли. Run целиком FAILURE только
  из-за GitHub upload-artifact storage quota; этот статус не переименовывать PASS.
- Холодный benchmark в этом run: ровно 1 DIRECT_FIRST reply YouTube группы —
  207.34 мс до первого byte; 9 RELAY_FIRST replies — среднее 6.29 мс, min 5.41,
  max 7.04. Все 7 первых имён получили verified TLS/HTTP nonce, 0 failed retry,
  0 local tuple collisions, protected no DIRECT. Это isolated CI, не ISP speedup.
- [Verify](https://github.com/buzzmean/neexitlag/actions/runs/39313482957): Windows
  PS5.1/7 browser-path/first-byte tests PASS, desktop native UI PASS; Windows job
  остановился на quota до последующих Go/lease steps. Linux full race с PostgreSQL
  PASS, vet и оба fuzz шага PASS; оба job FAILURE только upload quota.
  Пропущенные Windows шаги не считать выполненными.
- [Edge](https://github.com/buzzmean/neexitlag/actions/runs/39313482914): actual
  Edge policy/HTTPS/relay/recovery, C#/WPF и desktop/owner race PASS; run целиком
  FAILURE только upload quota.
- [Isolated upgrade](https://github.com/buzzmean/neexitlag/actions/runs/39313482886): SUCCESS.
  [Release build](https://github.com/buzzmean/neexitlag/actions/runs/39313483012):
  full Linux race/vet/builds PASS; run FAILURE только upload quota, публикация
  artifacts неуспешна.

### Журнал решений / передача

19.10.2026: распространять доказанную сетевую неисправность на фиксированную
YouTube family, а не начинать ladder каждого hostname; не переносить verdict
на другую сеть/профиль/порт и не считать все сайты Google заблокированными.
19.10.2026: time-to-first-byte сводка имеет собственный IPC budget; увеличение
телеметрии не должно ломать complete evidence и ослаблять acceptance assertions.
19.10.2026: измеренные CI миллисекунды не являются результатом провайдера владельца.
Нужно короткое сравнение новых соединений при VPN OFF; VPS для этого не менять.

Обновить службу прежним owner SID по docs/BROWSER_HOT_ATTACH.md с -NativeTcpFirewall,
закрыв только GUI, Chrome/Edge оставить открытыми. После новых вкладок/видео:
`& .\tools\windows\diag-no-vpn.ps1 -Only browser,youtube` из обычного PowerShell
в C:\neexitlag. Не удалять кэш/PAC/профили/identity. Scope освобождён.


## (s) Telegram/Anthropic relay-first и no_reply (20.10.2026)

- Причина DIRECT_FIRST: browser origin проходил общий AUTO learner, application preset не покрывал браузерный поддомен. Новое правило: точные Telegram/Anthropic семьи, normal active AUTO → RELAY_REQUIRED/VPS-01, KNOWN_RELAY_FIRST, без DIRECT trial/fallback. Ordinary domains/DNS53/YouTube local strategy scope сохранены. Явный BLOCK не ослабляется.
- Сохранено bounded verified DNS recovery через authenticated relay при destination OPEN отказе; auth/session failure не вызывает DNS probes. Проверка TLS/SNI/HTTP обязательна; данные приложения не replay-ятся. Новый guard список/offline-гарантии не добавлялись.
- no_reply получает первый наблюдаемый I/O cause до cleanup, request/origin payload presence без анализа содержимого. Client EOF может быть half-close; отсутствие payload только preconnect candidate. Старые 34/75 не классифицируются задним числом. Нужен свежий diag владельца; исправление реальных YouTube сбоев без причин не заявляется.
- Actual Chrome CI первого коммита выявил starvation DNS recovery: новый путь занимал общий AUTO probe semaphore, заполненный фоновыми origins. Исправление использует существующий отдельный DNSRecovery limit=2/singleflight/cooldown; добавлен regression с заполненными 4 ordinary slots. Тесты Chrome не ослаблялись.
- Повторный actual Chrome подтвердил работу HTTPS после DNS repair, но выявил перезапись Probe/source у queued requests с уже verified DNS pin. Сохранены DNS_RELAY_HTTP_VERIFIED/TTL evidence; добавлен regression queued plan. Native Chrome assertion не ослаблен.
- Диагностика .8: причины/legacy unknown, групповые before/after, безопасные fixed fields; ответы по-прежнему не являются page PASS. Отдельный timing IPC не вытесняет routing evidence.
- Код: `5f413c6`, `88a6f61`, `5c0d2ec`. Локально PASS: полный `go test -race -p 1 -timeout 120s ./...`, scoped race после последних правок, vet, Linux/Windows amd64 build, Windows compile test binaries browserproxy/client-service, 21 deployment Python test. PostgreSQL integration локально требует отдельную БД (CP код не менялся). Windows/.NET/PowerShell локально нет; CI: PS5.1/7 diag .8 и C# core PASS. Actual browser результаты ниже.
- ПК: штатно обновить только службу (GUI не менялся), браузеры не закрывать, PAC/identity не удалять. Новые вкладки Telegram/Claude/YouTube, VPN OFF; обычный owner PowerShell C:\neexitlag: `& .\tools\windows\diag-no-vpn.ps1 -Only browser,youtube`.
- VPS, credentials, certificates и PR #8 не затронуты; этап (в) не начат.

20.10.2026: финальная проверка runtime `5c0d2ec`:
- [Chrome native](https://github.com/buzzmean/neexitlag/actions/runs/39392197652): все tests PASS, включая ошибочный DNS Telegram, owner/SID/ECH, HTTP(S) и холодный YouTube; только upload-artifact FAIL (storage quota).
- [Edge](https://github.com/buzzmean/neexitlag/actions/runs/39392197692): тесты/WPF PASS, только upload-artifact FAIL (storage quota).
- [Verify](https://github.com/buzzmean/neexitlag/actions/runs/39392197744): Linux PostgreSQL/race/vet/fuzz и Windows PS5.1/7/native C# PASS. Upload-artifact FAIL (storage quota); следующие зависимые Windows шаги пропущены, полного green workflow нет. Удаление чужих артефактов/ослабление upload не выполнялись.
- Повтор локального полного race выявил асинхронную гонку TestPACPathsOpenBudgetAndFirstExchange: RX мог быть учтён раньше TX после net.Pipe echo. `5867936` ждёт оба счётчика, не меняя ни одной строгой assertion/таймаута/production логики. 30 повторов race + затем весь Go race/vet/Linux/Windows build PASS.
- CI после test-only правки: [Verify](https://github.com/buzzmean/neexitlag/actions/runs/39392793966), [Chrome](https://github.com/buzzmean/neexitlag/actions/runs/39392793977), [Edge](https://github.com/buzzmean/neexitlag/actions/runs/39392793978) — завершён: все исполнившиеся проверки PASS, только upload-artifact FAIL (storage quota). Chrome/Edge actual acceptance и Linux полный Go race/vet с PostgreSQL PASS; Verify Windows PS5.1/7 и native C# PASS, последующие Windows шаги после upload пропущены. Полный workflow green НЕ заявляется.
- Решение: известные Telegram/Anthropic группы не требуют foreground DIRECT trial; новое DNS repair не должно зависеть от занятых обычных AUTO probes. Старые no_reply не объявлять preconnect/PASS без фактов. Изменение crash-guard/этап (в) не входит в задачу.

20.10.2026: handoff завершён, область Codex освобождена. Функциональный код и тесты
запушены в codex/work, PR #8/ветка direct-plus не менялись. Итоговая инструкция ПК:
после штатного обновления службы (Chrome/Edge не закрывать) открыть новые вкладки
Telegram/Claude/видео при VPN OFF и выполнить
`& .\tools\windows\diag-no-vpn.ps1 -Only browser,youtube`.
Нужен именно новый breakdown: старые no_reply не объявлены ни preconnect, ни
исправленным Internet failure. Реальная скорость/работа сайта у провайдера ещё не доказана.


## (t) Fresh-PC enrollment после freeze — подготовка, НЕ deployment

20.10.2026, Codex, база `f51ec66`; реализация и handoff входят в один коммит
«Подготовить fresh-PC регистрацию и подтверждаемую ротацию CA» (`git log -1 --stat`).
Перед push выполнен fast-forward до `c67ff21`: новые дизайн-папки и обе записи Claude сохранены; код не пересекался.
Владелец явно потребовал один коммит, поэтому предварительная ownership-пометка
из PROTOCOL.md включена в итоговый коммит, отдельный служебный commit не делался.

- Новый `bootstrap-fresh-pc.ps1`: SDK/build/pinned driver, один host-verified SSH,
  новый ключ на ПК и CSR, явный DER SHA256 pin до доверия и регистрации,
  реальный signed-ticket/mTLS `/3` TCP+UDP probe, same-owner UAC install,
  unelevated GUI и прежняя PAC/nonce транзакция. Старый setup/provision не вызывается.
- CA key остаётся root:root 0600 на VPS и в root-only backup. CP получает private
  runtime-copy через systemd LoadCredential; это не изоляция от памяти CP.
  Migration требует отдельного явного `-RotateCA`/confirmation, проверяет старый pin,
  сохраняет `/etc/neexitlag.bak-*`, ставит независимый rollback timer, проверяет
  CP TLS/QUIC `/3`/TCP+UDP и откатывается при отказе. На VPS не запускалось.
- Hardware hash — inventory, не authentication. После нового freeze — новый ключ
  и ID; SSH-авторизованная замена slot отзывает прежний ID и публикует полный
  signed config в одной DB-транзакции. Legacy IDs без slot не удаляются молча.
  Signing/admin/echo/DB identity и replay state не вращаются/не восстанавливаются.
- Минимальные точки интеграции: CP `store.go` выделяет transaction publication;
  `App.xaml.cs` — explicit fresh setup entry; desktop-native runner и CI добавляют
  проверки. Routing/Chrome policy implementation/игры не менялись, этап (в) не начат.
- Локально PASS: полный Go race с отдельным PostgreSQL 17 (включая 5000 revocations),
  vet, Linux и Windows amd64 build; настоящий fixture TLS/mTLS QUIC `/3` + signed
  ticket + TCP/UDP echo; атомарный DB rollback. Python: 28 tests, включая 7 fresh-PC;
  PowerShell 7.4.6: parsing/StrictMode/pin/deadline; C# core: 130 tests; WPF/native
  cross-build с SDK 8.0.426. Windows PS5.1/native execution — CI, не локальный PASS.
- Непроверено: SSH password console Windows, реальная CA migration/systemd credential
  на данном VPS, полностью пустой Windows с UAC/GUI/PAC. Никаких production SSH,
  enroll/install/provision здесь не выполнялось. Старые CI страдают artifact quota;
  новый green workflow до завершения не заявляется.
- Ограничения: старый CA key должен существовать на VPS для безопасного rollback;
  если он потерян и там, скрипт откажет. Server cert 30 дней: автоматическое
  server-cert renewal не добавлено. Power-loss требует ручного backup rollback.
  Root SSH + UAC интерактивны; прежние PAC/restart dialogs остаются по инвариантам.
- Владельцу сейчас НЕ выполнять ротацию/установку на production без отдельного
  разрешения. Будущая обычная команда после подготовки VPS, owner PowerShell
  из `C:\neexitlag`: `& .\tools\windows\bootstrap-fresh-pc.ps1 -VpsIP 2.56.177.4 -SSHUser root -SSHPort 22`.
  Первичная CA migration отдельно описана в `docs/FRESH_PC_ENROLLMENT.md` и требует
  явного confirmation; никаких файлов старого bundle не нужно.


## (u) Fresh-PC bootstrap на UAC-off клубном ПК

20.10.2026, база `c0725d6`; один итоговый commit
«Поддержать fresh-PC bootstrap при подтверждённо выключенном UAC».
Ownership-пометка включена в этот же commit по прямому требованию владельца.

- `fresh-pc-execution.ps1`: read-only 64-bit DWORD EnableLUA/ConsentPromptBehaviorAdmin,
  TOKEN_QUERY (elevation/type/linked) и restricted detection. Consent=0 не считается
  выключенным UAC; missing/unknown/inconsistent evidence и trustlevel отклоняются.
- `bootstrap-fresh-pc.ps1`: общий Install-FreshService с повторной SAME owner SID/admin
  проверкой. При UAC-off нет RunAs/дочернего ServicePhase, GUI наследует полные права
  с явным предупреждением. При UAC-on обычный parent + прежний elevated child.
  install.ps1, ProgramData ACL и IPC owner-check не менялись.
- Root repo вычисляется от PSScriptRoot; hardcoded C:\neexitlag в скриптовой цепочке
  нет. Документ использует USERPROFILE; тест проверяет путь с пробелами в isolated
  Users/Test Owner fixture. Нет ни требования создать каталог в корне C:, ни запрета
  user-profile clone. CA pin/enroll/probe/preserved-state safeguards не менялись.
- Локально: PowerShell 7.4.6 parser + compiled native adapter, policy/token matrix,
  restricted/UAC-on elevated refusal, SAME SID dispatch, SDK/pin/status regressions
  и path fixture PASS; deployment Python 28 tests PASS; scoped `go test -race ./core/freshpc ./tools/fresh-pc` PASS. Windows CI выполняет настоящий
  read-only policy/token probe в PS5.1 и PS7; полная UAC-off установка и именно ПК
  владельца локально не доступны, PASS для них не заявляется.
- VPS, CA rotation, production enrollment, install.ps1 и provision здесь не запускались.
  PR #8 и direct-plus-flowseal не трогались. UAC/registry не менялись; зависимостей нет.
- Будущая команда из клонированного repo в профиле: `& .\tools\windows\bootstrap-fresh-pc.ps1 -VpsIP 2.56.177.4 -SSHUser root -SSHPort 22`.
  Это enrollment на production при реальном запуске, не локальная проверка; подготовка
  CA migration по (t) остаётся отдельным подтверждаемым действием.


## (v) UAC-off Default token / linked-query false rejection

20.10.2026, база `b882d64`; один commit
«Исправить Default token preflight и показать причины UAC-отказа».
Ownership включён в этот же commit по прямому требованию владельца.

- В b882d64 C# adapter бросал исключение при linked-token error !=1312. Тогда
  `Get-FreshExecutionEvidence` не возвращал snapshot и bootstrap печатал generic
  UAC_TOKEN_PREFLIGHT_REJECTED. Это реальный путь ложного отказа; конкретный Win32
  код на ПК владельца ранее не сохранён и не заявляется установленным.
- Теперь read-only collector сохраняет независимые результаты 18/20/40/19, policy
  Registry64/DWord и membership. Default допускает linked-query ERROR/ABSENT,
  но не PRESENT/NOT_RUN. Full/Limited error, restricted, UAC-on elevated,
  unknown elevation/restriction и плохая policy по-прежнему отклоняются.
- Safe diagnostic allowlist показывает все входы, query errors и точное условие
  отказа, без raw exceptions/окружения/секретов. Partial errors не скрывают уже
  собранные факты. Native adapter V2 не конфликтует с загруженным старым Add-Type.
- Локально PS7.4.6: exact owner fixture UAC=0/consent=0/type=1/elevated=1/restricted=0
  с linked errors 5/87/1008/1312 и null handle PASS; restricted/limited/UAC-on
  elevated отказы, матрица и безопасная partial diagnostics PASS; C# adapter
  компилируется. Deployment Python 28 tests и scoped Go race (core/freshpc, tools/fresh-pc) PASS. PS5.1/native token reading — существующий CI; actual owner ПК
  и полный fresh install здесь не выполнялись. Service dispatch/SID/ACL/IPC,
  CA pin/TLS/mTLS/SSH/enrollment/production logic не менялись.
- VPS, CA rotation и production enroll не запускались. PR #8/force-push запрещены
  и не выполнялись. Дополнительных dependencies нет.
- Владелец может безопасно проверить локально без VPS: `. .\tools\windows\fresh-pc-execution.ps1; $e=Get-FreshExecutionEvidence; Format-FreshExecutionEvidence $e 'READ_ONLY'; Resolve-FreshExecutionMode $e`.


## (w) Исправление DWORD token ABI / ERROR_BAD_LENGTH

20.10.2026, база `34e042d`; один commit
«Исправить четырёхбайтовый ABI запросов Windows token».
Ownership включён в итоговый commit по прямому требованию владельца.

- Причина подтверждена владельцем: на x64 предыдущий Dword() передавал
  IntPtr.Size=8 в GetTokenInformation(TokenElevation=20), Win32 API
  возвращал ERROR_BAD_LENGTH=24, Elevated оставался unknown. Это ошибка adapter,
  не состояния UAC. Старое предположение (v) о конкретном linked error владельца
  не подтверждалось; теперь есть точная ошибка elevation из ПК.
- Классы 18/20/40 используют отдельный P/Invoke out int + sizeof(int)=4. Класс19 —
  out TokenLinkedToken с IntPtr и Marshal.SizeOf этой структуры. Generic unmanaged
  buffer убран. Native adapter V3 исключает reuse уже загруженного V2.
- CheckedDwordResult сохраняет ошибку native call и отвергает неверный ReturnLength;
  error24 никогда не даёт успешное обязательное evidence, даже со stale value=1.
  Default/linked error1312 остаётся штатным, HasLinkedToken=false, ERROR/code сохранены.
- Безопасная диагностика failed GetTokenInformation содержит call, класс/номер,
  error и фактическую передаваемую длину. Не сериализуются ключи/токены/окружение.
- Локально: PS7.4.6, C# adapter compile/reflection ABI, result-validator error24,
  exact owner fixture UAC0/type1/elevated1/restricted0/linked1312 и существующие
  restricted/UAC-on elevated/limited отказы PASS. Native Win32 queries — Windows
  CI; реальный ПК владельца и полный bootstrap новой сборкой ещё не проверены.
  Python deployment 28 tests и scoped Go race (core/freshpc, tools/fresh-pc) PASS.
- ACL/SID/IPC/CA/enroll/install/rotation не менялись и не запускались.
  VPS/PR #8/direct-plus-flowseal не тронуты, force-push нет, новых dependencies нет.
- Безопасная проверка на ПК без SSH/enroll: `. .\tools\windows\fresh-pc-execution.ps1; $e=Get-FreshExecutionEvidence; Format-FreshExecutionEvidence $e 'READ_ONLY'; Resolve-FreshExecutionMode $e`.


## (x) VPS preparation без потерянной PC identity — только разработка

20.10.2026, база `a86d5e5`; один commit
«Подготовить VPS upgrade без старой device identity».
Ownership-пометка включена в тот же commit по прямому требованию владельца.

- Устранена найденная ошибка contract: relay публикует `default_lease_seconds`,
  а SSH/CA checks искали `lease_seconds`. Поэтому old error сам по себе не
  доказывал старые binaries. Read-only report показывает реальные capabilities;
  совместимые binaries не обновляются. Контракт проверен Go+Python regressions.
- `bootstrap-fresh-pc.ps1 -PreflightVps`: один SSH, Python engine/report в памяти,
  без builds/CSR/enroll/backup/staging/locks на VPS. Отдельные flags
  `-UpgradeVpsBinaries -ConfirmVpsUpgrade UPGRADE-VPS-2.56.177.4`; совместить
  read-only с upgrade/rotation нельзя. Флаги rotation подтверждаются отдельно.
- `fresh-vps-upgrade.py` использует production transaction engine. Минимальные hooks
  base engine сохраняют прежний updater. Backup `/etc/neexitlag.bak-UTC-12hex`
  root-only: full config/units, binaries, PG custom dump/globals, SHA/state/recovery.
  Независимые timer/boot guard и shared old-updater mutex; незавершённая прежняя
  transaction блокирует новую. Additive Store.Migrate после dump и до replacements;
  relay запускается раньше CP/full publication. CA/keys/config не вращаются тут.
- Root-only helper modes upgrade-migrate/local-probe/cleanup используют существующий
  CA 0600/0640 без преждевременного требования нового credential dropin. Новая
  временная identity только на VPS; CP/QUIC loopback, echo к собственному assigned
  public IP с сохранением public-only. Root-only fsync journal позволяет revoke
  после kill/SSH loss/boot; private probe state удаляется. Никаких старых PC credentials.
- Commit после настоящего mTLS `/3` ticket TCP/UDP echo, отзыва+полной publication,
  applied cache, health/ports/process SHA. Никакого pg_restore/replay rewind при
  rollback; при несовместимости или недоступности security state службы fail-closed.
  Legacy path >998 rows или config >48000 bytes требует отдельного reviewed plan.
  После последующей смены CA/config ручной binary rollback тоже не автоматический.
- Локально PASS: полный Go race с отдельным PostgreSQL17, включая 5000 revocations
  и новый cleanup/publication failure/retry; vet, Linux/Windows amd64 build;
  40 Python tests (12 новых), PS7.4.6 parser/confirmation/mode/one-SSH checks.
  Real pinned TLS/mTLS signed-ticket TCP/UDP fixture PASS. Systemd/SSH production
  сценарий и реальный ПК новой сборкой НЕ проверены. C#/GUI/routing code не менялись.
- Документ FRESH_PC_ENROLLMENT.md содержит две будущие команды и точный backup
  recovery. Для одного ПК owner inventory attestation не притворяется discovery
  36 клиентов. Только владелец запускает после review; разработчик не выполнял
  SSH/upgrade/CA rotation/install/enroll на 2.56.177.4 или veliq-swe-01.
- Будущий read-only запуск из repo на ПК: `& .\tools\windows\bootstrap-fresh-pc.ps1 -VpsIP 2.56.177.4 -SSHUser root -SSHPort 22 -PreflightVps`.
  Mutating combined command с ДВУМЯ confirmation tokens — в документе, не запускался.


## (y) Fresh-PC browser/PAC readiness — локальное исправление

20.10.2026, база `9a355b6`; один commit по требованию владельца.

- Владелец сообщил: production CA rotation committed, desktop ChatGPT/Codex/Claude
  работают; browser lane не готова. Разработчик VPS не трогал. Повторять CA/enroll
  для browser failure нельзя. Историческое исключение Chrome было потеряно в
  старом GUI/PS report; точную его причину на ПК без нового кода не утверждаем.
- Кодовые дефекты: startup deferred Edge при уже активном capture, fresh setup
  тихо пропускал его; Chrome cancel возвращался как Task success; report сохранял
  только exception class и PS скрывал даже phase. Native preparation timer не
  исключал chromeActivating/edgeActivating и мог менять profile/guard во время
  legacy PAC nonce transaction. Теперь mutual exclusion и Windows WPF regression.
- Explicit fresh/repair регистрирует отсутствующие Chrome/Edge как DIRECT staging
  с полными exe paths, не отключая другие приложения и не вводя pending AUTO в
  работающий capture. Existing ID/rules/mode сохранены до явной активации;
  MANUAL/DIRECT+ не переписываются. Edge перед Chrome; каждый installed browser
  обязан получить AUTO, listening, policy_ready и свой ready_profiles proof.
  Нет браузеров / отмена / ambiguous profile не получают PASS.
- Chrome: listeners/recovery до guard/quit, окончательный nonce после guard/quit,
  реальный headless policy GET, confirm, AUTO commit, затем восстановление вкладок.
  При отказе profile rollback и owned protected guard остаются; foreign policy не
  меняется. Диалог и штатный quit/restart НЕ удалены; этап (в) не начат.
- Chrome browser.status добавляет безопасные last_error/readiness_state без raw
  ошибок/путей/nonce; Go regressions bind failure / missing proof / prepare при
  active capture. GUI сохраняет fixed phase/code/browser + bounded state allowlist.
  PS печатает PAC_READINESS_FAILED и browser/edge.status даже при optional полях,
  отличает GUI early exit, timeout и malformed report; deadline sample не теряется.
- `bootstrap-fresh-pc.ps1 -RepairBrowserReadiness` — только существующие service/
  identity; owner/token проверки прежние, без SSH/build/enroll/rotation/install.
  Запускается после обновления GUI/службы и выхода только VELIQ через трей.
  Комбинации с server/ServicePhase flags отвергаются до каких-либо действий.
  Инструкция — docs/FRESH_PC_ENROLLMENT.md, раздел browser/PAC repair.
- Локально: C# 135 unit tests PASS; WPF/native-host cross-build PASS (прежние
  nullable warnings в recovery); PS7 parser/token/mode/PAC redaction/timeout/
  early-exit/boundary fixtures PASS; deployment Python 40 PASS; Go race, vet и
  Linux/Windows amd64 builds проверены. Actual Windows WPF regression добавлена в CI; локально только cross-build.
  Bootstrap на owner UAC-off ПК/Internet после fix НЕ проверен.
- Не менялись TLS/mTLS/owner SID/ACL/WinDivert/маршрутизация игр/production VPS,
  credentials/CA/сертификаты/PR #8. diag-result/cs-diag/observe-result не добавлены.
- ПК: обновить службу/GUI существующими update scripts, затем
  `& .\tools\windows\bootstrap-fresh-pc.ps1 -VpsIP 2.56.177.4 -SSHUser root -RepairBrowserReadiness`.
  Принять старые dialogs штатного browser restart; при отказе не повторять enrollment.


## (z) Этап (в): остановка на независимом guard, только документация

20.10.2026, база `2c4f75f`; один итоговый commit по требованию владельца.

- Цель утверждена: одна кнопка/автостарт, без dialogs/quit/browser restart/user
  reconnect/PowerShell. Новые flows native, старые TCP PINNED_UNTIL_CLOSED.
- Найден security blocker текущей реализации, применён прямой пункт 4 задания.
  WinDivert 2.2.2 (версия из pins/download helper): dynamic WFP session, lifetime
  handles; cleanup может reinject не просроченные NETWORK queue/work queue packets,
  destroy удаляет callouts/закрывает engine. Watchdog, реагирующий ПОСЛЕ crash,
  не делает запрет атомарным. Второй процесс/duplicate handle сам по себе не
  доказывает boot/last-handle/guard-crash безопасность.
- Persistent user-mode WFP/Firewall BLOCK возможен без собственного драйвера,
  но не перенаправляет исходный connect в существующий NETWORK reflector.
  SOCKET layer block не позволяет изменить/inject connect. Production WFPBackend
  — явная unavailable заглушка, как и описывает ADR-001. Нельзя назвать такой
  block работоспособным браузером или снять owned PAC check ради false READY.
- Рекомендация — подписанный ALE connect-redirection/enforcement callout с
  постоянными per-app+owner v4/v6 rules, независимым broker/watchdog, generation/
  heartbeat, bounded admission и защитой boot/recovery. WFP сам не раскрывает
  ECH/H2/H3 origin; coalescing/DNS/opaque правила требуют отдельного no-leak gate.
  Shared PID/tuple/impostor не являются доверенным permit; games и desktop apps
  должны остаться вне browser guard scope.
- Новый docs/BROWSER_NATIVE_GUARD.md содержит подтверждённые места кода и
  upstream/Microsoft первичные источники, контрпримеры, архитектуру, ограничения,
  безопасную миграцию уже прочитанного PAC, diagnostics proposal и failure-injection
  acceptance. Proposed states/codes/≤3s/≤2s/≤10s — не существующий IPC/измерения.
- Runtime, GUI, bootstrap, diag-скрипт и WinDivert pins не менялись. Диалоги старого
  кода НЕ объявлены временным продуктовым решением; этап (в) НЕ выполнен.
  Снятие PAC до нового guard не разрешено. Нужны WDK/signing/Windows kernel test
  возможности и решение владельца о новом driver scope; нет оснований объявлять
  stage (в) безопасно реализованным на имеющихся компонентах.
- Проверка — чтение соответствующих исходников и официальных docs, существование
  relative links/якорных методов и git diff --check. Нового guard нет: kill/hang/
  boot/native browser acceptance НЕ выполнялись и PASS не заявляется. Go/C#/CI
  runtime suite не перезапускалась для docs-only изменения.
- Production VPS, CA, сертификаты, credentials, PR #8, игровые и desktop policies
  не тронуты. diag-result/cs-diag/observe-result не добавлялись.
- На ПК НИЧЕГО НЕ ЗАПУСКАТЬ: только документация. Новый install/repair с browser
  restart не предлагается для выполнения этапа (в). Дальше — решение о WFP scope,
  не очередной ручной browser test с перезапуском.


## (aa) Telegram main origin: TLS framing и read-only воспроизведение

База `c22c480`; пользовательский runtime `2c4f75f`. Работа этой итерации — диагностика,
не изменение маршрутизации и не заявленный Telegram fix/PASS.

- Проверены PAC CONNECT -> native delegation -> QUIC stream -> relay TCP copy.
  Подтверждённого обрезания/перекодирования SNI/ALPN/ECH/ClientHello не найдено.
- `openKnownBrowserGroup` восстанавливает DNS только после TCP OPEN rejection,
  не после последующего TLS failure. Проверка `.40` против `.99` ещё нужна:
  отсутствуют основания автоматически назначать постоянный IP или replay TLS.
- Новые `tls_wire` поля: реальные размеры первых восьми TLS records, длина первого
  ClientHello, complete/incomplete, actual accepted writes и reads до client delivery.
  Пять байтов `TLS_HEADER_ONLY` не объявляются alert. Полный plaintext alert имеет
  числовой level/description. Payload первых 16 байт маскируется; полные Hello,
  random/session IDs/ECH/HTTP/ключи не сохраняются.
- Инструментированы PAC и actual native relay boundary, включая PAC_NATIVE_UPSTREAM.
  Счётчики двух точек нельзя суммировать как независимые flows. Metadata bounded;
  прежние IPC record limits/truncated/lost остаются в силе. Игры/guard/policy без изменений.
- `deploy/vps/diagnose-telegram-tls.py`: только OpenSSL/curl public GET с правильным
  SNI/hostname и проверкой системной CA; `.40`/`.99` — кандидаты, не pins.
  Есть обычный ClientHello, optional ML-KEM (SKIP на старом OpenSSL), большой
  Hello с unknown extensions (~1953 байта в local OpenSSL), HTTP/1.1 и HTTP/2.
  Время/вывод ограничены; raw trace/body/header не печатаются. Никаких installs,
  служебных identity, ключей, /etc/neexitlag, systemd/firewall/credentials изменений.
- `diag-no-vpn.ps1` версия `2026-10-20.1`; выводит framing и masked response prefix
  отдельно для PAC и actual upstream. Старые данные не реконструируются.
- Регрессии: реальный QUIC+mTLS+verified origin TLS/GET с большим ML-KEM Hello,
  fragment writes по 73 байта, opaque ECH/GREASE-like stream byte equality,
  5-byte header + EOF delivery, partial-write accounting, redaction/bounds,
  native pre-read ML-KEM при complete и segment-limit, read-only script parsing.
- Проверки PASS: новые Go race-регрессии в browserproxy/platform/windows/client-service/integration;
  48 Python deployment tests (включая actual local OpenSSL large Hello); PowerShell 7
  StrictMode/парсинг; go vet ./...; Linux/Windows amd64 builds; Windows native test compile.
  Полный go test -race -p 1 -timeout 120s ./... НЕ зелёный: последний прогон падает
  в существующем TestPoolTCPBurstContinuity/320 (OPEN timeout). Ранее full-run
  также ловил revocation ACK race и demo deadline; standalone lease и demo PASS.
  Чистый baseline c22c480 demo PASS за 24.1 с при лимите 25 с. Эти тесты и transport/
  lease/demo production-код не менялись; причина burst-отказа не доказана, CI PASS
  не заявляется. Не увеличивать production timeouts ради зелёного результата.
- Не проверено: actual owner Chrome/Edge/ISP и direct OpenSSL/curl на VPS. Ни одного
  production SSH/HTTP/TCP вызова из среды разработки, Telegram Internet PASS нет.

Владельцу из корня репозитория (обычный PowerShell, пароль только в SSH prompt):
`Get-Content -Raw .\deploy\vps\diagnose-telegram-tls.py | ssh -T -p 22 root@2.56.177.4 'python3 - --confirm-read-only'`.
После обновления службы и нового Chrome/Edge соединения:
`& .\tools\windows\diag-no-vpn.ps1 -Only browser,youtube`.
Полная интерпретация и ограничения — `docs/TELEGRAM_TLS_DIAGNOSTICS.md`.


## (ab) Relay-first origin-IP health/recovery

База `042869d`; owner read-only VPS proof принят как причина Telegram main-host:
TCP OPEN успешен, TLS на исходном DNS IP не работает, независимый IP с тем же SNI
проходит certificate/GET. Никаких IP pins/изменений DNS Windows/VPS.

- Новый OriginHealth: scope network/profile revision/domain/port/relay/IP,
  negative 60 с (1024 entries), TTL expiry -> lazy real retry. DNS answers TTL
  <=5 мин (256 entries), singleflight, 2 workers, error cooldown.
- Новый originConn: три application OPEN максимум, 900 мс silence/retry OPEN,
  4 с с первого flight; только complete ClientHello/CCS <=64 КиБ. Никакого replay
  application/early-data records или после одного origin byte. Ранний alert/EOF
  с байтами -> отрицательный кэш для следующих flows; частичный TLS reply не висит
  бесконечно. Уже работающие TLS/HTTP flows pinned, без смены relay egress.
- При native PAC детектор оборачивает actual QUIC stream; report-ref переносит
  статистику в PAC, не теряя session/auth error class. Эти ошибки не отравляют IP.
- Все ограниченные DNS ответы до 64 IP участвуют в выборе; ordinary route probes
  сохраняют 3 адреса. Independent Google+Cloudflare DoH объединяются через relay
  HTTPS/443, TLS проверяется. /3 не умеет hostname OPEN; ОС DNS VPS не вызывался,
  deployment не требуется. DNS/53 не relayed.
- Protected/known и выбранные REQUIRED browser origins: только прежний relay/fallback;
  IP BLOCK исходного policy snapshot сохраняется, owner/network checks на каждом OPEN.
  Proof/HTTPStatus старого IP не переносится на новый, byte/ACK не объявляется PASS.
- Диаг .2: addresses_tried/selected/negative_skips/dns/replayed bytes/attempt reasons;
  PAC+native один flow, не суммировать. GUI/игры/DIRECT+/guard/lease/credentials без правок.
- Тесты: focused Go race (origin/DNS/metadata/native PAC + actual QUIC/mTLS TLS/HTTP) PASS;
  `go vet ./...`, Linux и Windows `go build ./...`, Windows service tests compile PASS;
  48 Python deployment tests и PowerShell browser-path checks PASS. Полный
  `go test -race -p 1 ./...` НЕ PASS: TestLeaseRenewalReplayAndDeviceBinding
  (device_revoked), TestPool160SimultaneousRequiredFlowsAndFreshTickets
  (OPEN/admission timeout). Оба отдельно повторены с race: PASS, 2.213 с.
  Эти фикстуры/lease/pool не менялись; причину нестабильности общий прогон не доказал.
- Не проверено: настоящий Windows WinDivert + Chrome/Edge после изменения и ISP;
  никакого production SSH/HTTP/deploy/CA действия из среды разработки. Первый flow
  с ранним origin alert может FAIL согласно запрету replay после первого байта;
  следующий connection использует negative cache. Не обещать обратного/Internet PASS.

Владельцу после обычной сборки/обновления службы: открыть Telegram Web в Chrome/Edge,
затем `& .\tools\windows\diag-no-vpn.ps1 -Only browser,youtube` и смотреть .2
`origin IP recovery` (выбранный IP/попытки), TLS framing и реальный результат страницы.


## (ac) Вариант A этапа (в): тихая PAC policy без browser restart

- Удалены MessageBox/restart/quit/restore-last-session, ожидание/launch браузеров из
  Chrome/Edge activation и fresh/repair. Recovery helper остаётся служебным EXE,
  с прежними owner SID/ACL/guard; Chrome/Edge процессы продукт не трогает.
- Служебные browser.silent.apply / edge.silent.apply отдельно разрешают configured
  lane после listener/recovery/mandatory PAC; никаких synthetic nonce/ReadyProfiles.
  AUTO сохраняется только после successful apply. Chrome legacy journal мигрирует
  при совпадении expected PAC (разный порядок JSON не конфликт). Чужую policy не
  переписываем. Новая Edge staging DIRECT безопасна при уже активном capture.
- Restore на каждом SCM boot идёт до CP/network initialization: стабильный listener
  и owned policy восстанавливаются из ACL-защищённого journal. Неподключённый owner
  hive не создаём; bounded maintenance retry через 5 с после logon. HKLM не меняем.
- GUI user RefreshPolicy(false) — запрос, не доказательство. Silent state отдельно:
  POLICY_PENDING_BROWSER_REFRESH, PAC_REQUEST_OBSERVED, POLICY_OR_LISTENER_UNAVAILABLE.
  Для owner ingress есть timestamp/count/apply_to_request_ms. Nonce и page PASS
  не подделываются. Bootstrap/fresh repair не падает лишь из-за отсутствия browser
  uptake; реальные отказы listener/ACL/policy всё ещё FAIL. Update выводит safe status.
- Chrome/Edge GUI + diag .3 показывают pending/observed честно. Клубный образ должен
  содержать policy/identity/config и штатный автозапуск: если freeze стирает policy
  и state, нельзя обещать применение до запуска браузера. Live uptake не гарантирован
  для всех Chromium: при отсутствии подхвата — следующий естественный запуск,
  без просьбы/диалога. Existing TCP pinned, reload может переиспользовать HTTP/2.
- Windows CI добавлен actual running Chrome/Edge + registry/user-GP/CDP HTTP ingress
  test без proxy launch flags, с PID/creation-time и elapsed-ms либо явным pending.
  Это не Internet/TLS PASS. Исторический HTTPS/mTLS acceptance сохранён отдельно.
- Тесты: полный `go test -race -p 1 -timeout 120s ./...` PASS (integration 43.621 с);
  дополнительный race silent/owner/policy PASS, `go vet ./...`, Linux/Windows
  `go build ./...`, Windows service tests compile PASS. C# 139 tests PASS,
  WPF cross-build PASS (4 прежних CS8604 warnings, 0 errors), 52 Python tests PASS,
  PowerShell fresh/browser-path checks + parser PASS. Actual Windows uptake
  acceptance добавлен в CI, локально Linux не выполнялся. GitHub CLI Actions API
  недоступен из среды (403), CI PASS до наблюдения не заявлять.
- Не проверено: actual Windows/ISP в этой итерации, live policy uptake время, клубный
  boot с браузером в автозапуске. Production VPS/credentials/CA/certs не тронуты;
  игры и relay/lease policy без изменений. PR #8 не тронут.

Владельцу: обычное обновление службы/GUI с открытыми браузерами; для необязательного
отчёта `& .\tools\windows\diag-no-vpn.ps1 -Only browser,youtube`. В программе
только «Оптимизировать»/автозапуск, никаких browser restart диалогов.


Post-push CI 5d47ff0: Edge real HTTPS/mTLS step PASS; silent Edge uptake за ~30 с
не наблюдался (PID/creation unchanged). Chrome CDP instrumentation unavailable
за 10 с -> новый acceptance FAIL, не Internet failure. Verify Linux упал в старом
OpenSSL diagnostics test, Verify Windows прерван artifact quota до Go tests.
Выявлена native-only ошибка новой freeze-фикстуры: S-1-test нельзя использовать
в Windows ACL publication. Локально подготовлен minimal actual-user-SID fixture
fix, portable race и Windows compile PASS. Владелец явно разрешил дополнительный
test/docs commit 20.10.2026: только actual SID для Windows freeze-фикстуры и
результаты CI, без переписывания истории. Исправление не меняет production-код,
policy/guard/ACL contract, CDP readiness и маршруты. Нативный Windows запуск этого
исправления локально не выполнялся; cross-compile не является runtime PASS.


## (af) Коды активации и автопродление server leaf — 20.10.2026

База: 165e026 (план/владение), fast-forward сайта владельца 50ae3cb сохранён.
Добавлена migration 002, публичный code+CSR enroll, club/personal seat model,
admin API/CLI, атомарные замена ключа/отзыв/полный signed config. В БД только hash
кода; hardware v2 — общий HMAC UUID/board/MAC, legacy fresh UUID v1 сохранён
ради существующих slots, v2 включается публичной солью. HW spoofing не является
TPM-аттестацией; код из образа доступен администратору машины.

Служба активирует в фоне с backoff, держит код и новый private key под SYSTEM/admin
ACL, повторяет потерянный ответ тем же CSR, не стирает рабочую identity при неудачной
смене кода. В образ — только public trust и отложенный код, не device private key.
GUI: маскированный ввод, безопасные ошибки, смена кода/сроки в Настройках.
Существующий legacy enroll и remote/renew путь сохранены; маршруты не менялись.

Проверено локально: полный go test -race -p 1 ./... (первый прогон старого demo
timeout, отдельный и повторный полный прогон PASS); vet, Linux build и Windows
cross-build; 145 C# tests, WPF build (старые CS8604), 61 deployment Python tests;
PS parsing + fresh/reinstall checks. Windows ACL regression скомпилирован, runtime
только в Windows CI. COM desktop-upgrade проверка на Linux недоступна.
PostgreSQL 17 — отдельный localhost/tmpfs container, не production.
50 freeze replacements <5s занимают 1 seat; настоящий TLS fixture проверяет lost
response/CSR/idempotency/новые ключи, отзыв/limit/DB failure/publication rollback.

Server leaf timer: те же CA/SAN, 30d leaf, порог 10d, root-only durable backup,
independent rollback/boot recovery, HTTPS и локальный mTLS QUIC /3 TCP+UDP proof.
Временный proof device имеет durable journal для отзыва после crash. CA не ротируется.
Production coordinator использует существующий binary upgrade без старого device
credential, SHA256, pg_dump, health proof, затем leaf/timer/LoadCredential.

Решение владельца: при rollback сохранять nonce, revocations, новые devices и
совместимую additive schema; pg_dump — только backup, не автоматический pg_restore.
Подготовленные coordinator/leaf transactions проверены fixtures, но production
systemd/SSH не выполнялись. SSH к VPS из этой среды недоступен: этапы 4/5 не
выполнены, реальные club/personal codes не созданы, leaf на VPS не продлён.
Ожидается CI; команды владельцу и rollback — docs/ENROLLMENT_CODES.md.

Следующий шаг после зелёного CI полного commit: из checkout codex/work выполнить
`& .\tools\windows\update-activation-vps.ps1 -ExpectedCommit (git rev-parse HEAD).Trim() -ConfirmDeployment 'ACTIVATE-CODES-2.56.177.4'`.
Это отдельное изменяющее VPS действие владельца, не обычный bootstrap/CA rotation.
После deployment проверить существующий клиент, затем создать тестовые коды на VPS
через root-only veliq-codes (одноразовый stdout, без transcript/файлов), проверить
masked activation и повторную загрузку на ПК. Реального Windows/ISP PASS не заявляем.

Журнал решений 20.10.2026: сохранять актуальные security-записи при откате (явное
уточнение владельца); public runtime probe credential выдаётся только после
валидного enrollment, root CA key/admin token клиенту не передаются; без SSH
останавливаться на подготовленном deployment, не выдавать выдуманные рабочие коды.


Дополнение к (af): implementation 4cdf2e5 запушен. При ревью смены кода обнаружен
старый runtime-флаг connected после замены identity: activationAttempt останавливает
прежний remote, и теперь менеджер обязательно создаёт новый remote. Регрессия
на настоящем HTTPS/CSR проверяет смену кода, закрытие прежнего worker group и
появление нового без перезапуска службы. Targeted race PASS.
CI 4cdf2e5: isolated systemd/PostgreSQL upgrade PASS (run 39484588909), Verify core
остановился в старой Telegram OpenSSL-фикстуре до новых Go tests: fixture не имела
production -ign_eof и могла выходить на закрытом stdin до ClientHello. Исправлена
только фикстура, assertions сохранены; локально 61 deployment tests PASS.
Остальные CI ещё выполняются; production deployment по-прежнему не выполнялся.

CI: добавлен focused activation.yml без upload artifacts: isolated PostgreSQL/race,
Go vet/build, deployment tests; Windows native SYSTEM/admin ACL, activation contracts,
145 C# tests/WPF и PS5.1 parsing. Это отделяет новые проверки от прежнего quota-blocked
общего Verify; не заменяет Internet/production acceptance.

CI уточнение 7ea0769: реальный ClientHello фикстуры 713 bytes на OpenSSL runner,
а не ожидаемые >=1500. Одного -ign_eof недостаточно: версии ограничивают число
custom extensions по-разному. В test-only ALPN добавлены 8 bounded protocol names
для фактического большого ClientHello; размер/framing assertions не ослаблены,
production TLS/probe не изменены. Причина прежнего CI отказа уточнена по safe summary.

CI b84bdb3: focused Activation and leaf renewal PASS на Linux+Windows
(run 39485374085); реальные native protected ACL tests, activation contracts,
145 C# tests/WPF и PS5.1 PASS. Isolated production upgrade PASS (39485374001).
General Verify Windows остановился на upload-artifact quota (39485374049), не
ошибке новых тестов. Полный общий CI зелёным не объявляется; deployment не выполнен.

Дополнительная regression смены кода: неверный новый код не должен перезапускать
работающие remote workers/потоки. connected теперь проверяет реальное состояние
worker group, а не сбрасывается от одного GUI wake. HTTPS regression последовательно
проверяет invalid change (старый worker жив) и valid change (старый закрыт, новый
создан). Полный client-service race PASS (6.606 s); остальные маршруты не менялись.
Документировано maintenance-ограничение leaf timer: CP/relay restart может кратко
прервать relay-сессии; это не изменение same-session lease renewal.

Финальная передача (af): код 0d3883b, область освобождена, новых правок не планируется.
- https://github.com/buzzmean/neexitlag/actions/runs/39485374085 — focused Linux/Windows PASS (b84bdb3).
- https://github.com/buzzmean/neexitlag/actions/runs/39485839538 — isolated production upgrade PASS (0d3883b).
- https://github.com/buzzmean/neexitlag/actions/runs/39485839584 — focused CI 0d3883b, выполнялся на момент записи.
- Общий Verify/Edge: старое ограничение artifact quota; все workflows зелёными не названы.
Последняя передача дополнительно проверяет чистый checkout и untracked files до
сборки server/helper: Go VCS modified не должен обнаруживаться лишь после SSH.
Ignored `.e2e/`/`bin/` допустимы; отчёты держать вне release checkout, не коммитить.

Production/живой Windows activation/freeze/systemd leaf renewal не проверены.
Из этой среды нет разрешённого SSH/TCP-доступа к VPS, поэтому остановка ровно на
передаче команд владельцу (пункт 4 задания), а не deployment или выдаче фиктивных
кодов. Следующий owner-run — только после принятия CI, существующую identity не
переустанавливать и CA не ротировать. Отзыв/nonce при rollback сохраняются.


## (ag) Activation upgrade: root-only CA layout и диагностируемый отказ

20.10.2026, база 3f31c1f. Владелец запускал deployment сам: SSH пароль со второй
попытки принят, достигнут VPS_PREFLIGHT. Binaries пока сообщали 2c06089, schema 1,
CA migration уже выполнена (91 LoadCredential). Старый общий код ошибки не хранит
точный исторический exit/stage; без read-only снимка нельзя объявить VPS неизменённым.

Воспроизведена несовместимость старого preflight: direct `runuser ... test -r`
для CA key root:600; затем неизвестный 91 drop-in и `%d` CA argument. Исправлено
строгим признанием только owned LoadCredential layout, сравнением runtime key с
root key (приватно), seed-free override 94 после 91/92. Root CA ACL не расширяется.
Неизвестные directives/specifiers/drop-ins по-прежнему блокируют обновление.
Full candidate validation выполняется ДО постоянных backup/units/service stop.

Новый activation SSH transport не загружает bootstrap/enroll; сохраняет один
host-verified сеанс и интерактивный пароль. Нет ложного `no installation performed`.
Удалённый failure report: stage/step, command/exit_code, allowlisted stderr, backup,
vps_modified и rollback. Неизвестные слова/PEM/ошибки не пересылаются как raw stderr.
Backup выводится сразу после создания; guard unit тоже входит в backup/rollback.
Исходная причина сохраняется даже при отдельной ошибке rollback. SQL restore нет,
nonce/revocations/добавочная схема сохраняются согласно решению владельца.

Read-only: `& .\tools\windows\verify-activation-vps.ps1` из обновлённого клона.
Скрипт ничего не собирает и не staging/enroll/provision/restart на VPS; Python
исполняется из памяти, SQL в BEGIN READ ONLY. Сверяет installed/running hashes,
baseline commit, schema/activation tables, PID/start-time, units/drop-ins, timers,
backup phases, CA root/runtime доступ. Snapshot не доказывает отсутствие всех
исторических DB writes; baseline и deployment artifacts показываются раздельно.
Старый скрипт мог создать внешний backup/guard и перезапустить службы при rollback
ещё до замены binaries/БД. Этот факт не скрывается под «ничего не установлено».

Локально: 70 Python deployment tests PASS; bash syntax, PS parser/quoting + fresh-PC
regressions PASS. Go/C# production код не менялся. Добавлен Windows fake ssh.exe
process test на loopback: redirected stdout/exit=23, без сети/паролей; локально
Windows runtime недоступен. Isolated systemd CI теперь имеет legacy и fresh-credential
(root:600 + LoadCredential) варианты, оба проверяют update/rollback без ACL downgrade.
CI запускается после commit; результаты проверять по SHA в Actions.

На боевом VPS, в его БД, units или ключах из среды разработки ничего не выполнялось.
Первая команда владельцу — read-only verify; повторный deployment только после
разбора результата и CI нового полного SHA. PR #8/маршруты/credentials не менялись.
Область освобождена, одно отдельное исправление в codex/work без rewrite/force-push.

CI после 05cdfd1: оба actual isolated systemd/PostgreSQL прогона (legacy и fresh
LoadCredential/root:600) PASS, run 39492338314. Новый Windows SSH fixture выявил
UTF-8 BOM в stdin .NET Framework: literal payload fail не был распознан fixture.
Транспорт теперь явно StandardInputEncoding=UTF8Encoding(false), fixture отвергает
любой неизвестный input (не принимает его как success). Это отдельный небольшой
follow-up без rewrite; на VPS ничего не выполнялось. Локально PS parse/quoting PASS;
actual native Windows rerun — после push, фактический результат отдельно в Actions.

Уточнение stdin: StandardInputEncoding отсутствует в .NET Framework reference source.
В 053e858 применение было без проверки; следующий маленький compatibility fix
проверяет PSObject property и не использует его в PS5.1. Сервер json.loads(bytes)
принимает optional BOM (существующий успешный путь owner SSH); native fixture теперь
читает raw stdin StreamReader с BOM detection, как сервер, и проверяет оба exit path.
На Core свойство позволяет задать UTF-8 явно. История не переписывается.

Коррекция диагноза Windows fixture (09b7fa4 CI): BAD_FIXTURE_INPUT_LENGTH=0.
Причина оказалась не BOM, а оставшийся @args после переименования hashtable в
$sshFixtureArguments. В production updater аргументы передаются именованно, этот
дефект относится к новому тесту. Исправлены оба splat; транспорт дополнительно
отвергает пустые параметры до Process.Start без prompt/SSH. Предположение о BOM
не является установленной причиной исходного VPS отказа. Conditional encoding
на Core и BOM-aware raw reader fixture совместимы с Python JSON и Framework.
CA/systemd исправление неизменно; оба actual isolated systemd variants PASS.


## (ak) 21.10.2026 — READ_ONLY_SNAPSHOT / неверный default baseline
- Фикс отдельным коммитом в codex/work; история сайта Claude до 40ca824 сохранена fast-forward.
- Причина RuntimeError воспроизведена на точном Python source: baseline по умолчанию
  имел 38 символов, пропущено `7d`; верный полный hash
  `2c06089f61c7b7a27cdf575f5cea4b834992426d` получен из git. Это проверяется тестом;
  malformed baseline по-прежнему отклоняется. В PowerShell default проверяется до SSH.
- Общий formatter read-only/update: code/step/exit/command, безопасные категории stderr,
  bounded traceback только file/line, без source/locals/сырого exception. Read-only
  psql stderr сохраняется приватно до классификации; stdout виден и при SSH error.
- CRLF/BOM проверяются на реальных модулях в памяти; отдельная Windows PS5.1 fixture
  исполняет точный launcher через локальный Python/pipe, без SSH/VPS. EnvironmentFile
  BOM/CRLF/quoted address принимается; неверный/повторный адрес отклоняется.
- Модуль диагностики входит в SHA256 release manifest и постоянный backup rollback.
- Локально: 76 deployment Python tests PASS; PS parsing/quoting PASS.
  Windows5.1 native pipe/child process — в CI после push, локально Linux.
- Никаких команд на VPS, enroll, сертификатах, БД/сервисах не выполнялось.
  Реальное состояние Debian/VPS не проверено. PR #8 не тронут.
- Владельцу, PowerShell из корня клона: `git pull --ff-only origin codex/work; & .\tools\windows\verify-activation-vps.ps1` (только read-only).
- Решение: CRLF/BOM не объявляем причиной исторического отказа; доказанный дефект —
  38-символьный baseline и потеря кода исключения. Unknown exceptions остаются
  санированными, а не выводятся целиком ради диагностики.

- Post-push CI 0274127: isolated systemd/PostgreSQL legacy + fresh-credential PASS
  ([run](https://github.com/buzzmean/neexitlag/actions/runs/39544537587)). Windows
  native Go, 145 C# tests, WPF, parse и все assertions pipe/child process PASS;
  job ошибочно failed из-за ожидаемого Python exit=1, оставшегося в LASTEXITCODE.
  Отдельный маленький follow-up очищает его только после всех assertions;
  неожиданный exit по-прежнему приводит к Check/throw. История не переписывается.

- Follow-up 9489f5e: сброс только ожидаемого LASTEXITCODE после проверок фикстуры.
  Workflow dispatch через доступный GitHub CLI отклонён HTTP 403; добавлен
  отсутствовавший path-trigger на саму Windows activation-фикстуру, чтобы её
  изменения всегда проверялись CI автоматически, без ручного dispatch.


## (al) 21.10.2026 — LEAF_RENEWAL_FAILED / безопасный повтор schema 2
- Причина подтверждена по исходникам: renewal передаёт сертификатный journal
  /var/backups/neexitlag-certificates/<UTC>-<8hex>/probe.id, а Go-helper 5bbf629
  разрешал только /etc/neexitlag.bak-<UTC>-<12hex>/probe.id. Локальная проба
  отказывала до Register; stderr терялся внешним subprocess, mixed-case code
  превращался в SAFE_OPERATION_FAILED. Точные два формата теперь проверяются,
  отдельный read-only helper contract вызывается до любых изменений.
- Все child-процессы activation/leaf: stdout/stderr/exit санированы, фиксированные
  codes и file/line trace, без свободного текста/ключей/кодов. Для binary engine
  включён такой же observer. Итоговый снимок binaries/schema/tables/units/timers/
  certificates передаётся parent; при невозможности снимка — явная причина.
- Read-only verify показывает срок leaf/CA, состояние renewal service/timer,
  безопасные временные probe ID/даты/отзыв и последние версии публикаций.
- Schema 2 additive и совместима с baseline SQL 001; старый бинарник не реализует
  activation API. +1/+1 согласуется с revoke binary proof; точная причинность
  +19 по одному snapshot не доказана. Повтор сохраняет соль/nonce/отзывы.
- Выбран путь доведения deployment вперёд из schema 2. Полный byte-for-byte
  rollback живой БД несовместим с сохранением новых revocations/replay; его НЕ
  реализуем и НЕ заявляем. Security-monotonic rollback восстанавливает runtime/
  units/config/leaf, сохраняет additive schema/security rows. Полная DB атомарность
  требует отдельного maintenance/security-epoch дизайна. Это открытое ограничение.
- Renewal CLI/rollback сериализуются; успешный commit/rollback останавливает
  transient timer. StateDirectory создаёт private каталог systemd sandbox;
  deploy проверяет реальный service due=false до commit. CA не ротируется.
- Локально: 81 Python tests, helper Go race/vet, Linux/Windows build и PS parser
  PASS (PG-тесты локально пропущены без DB). В CI добавлены native Windows helper
  contracts, PG baseline/schema2 совместимость, полноценный systemd activation/
  leaf/failure-injection/повтор в disposable fresh-credential VM.
- Подробно и read-only команда: docs/ACTIVATION_DEPLOYMENT_RECOVERY.md.
  Владелец сначала запускает из корня клона `git pull --ff-only origin codex/work`,
  затем `& .\tools\windows\verify-activation-vps.ps1`. Никаких mutating VPS
  команд агент не выполнял. Deployment и установка renewal до 19.11.2026
  16:05:54 МСК остаются за владельцем после отдельного разрешения.

- Post-push f01012f: локально поднят отдельный PostgreSQL 17 container, выполнены
  `go test -race -p 1 ./tools/fresh-pc ./services/control-plane` с настоящей БД —
  PASS, включая schema2/001/retry и revoke publication. Собственный container
  остановлен; production не затронут.
- GitHub CI f01012f не начал тестовые шаги: activation Linux/Windows и isolated
  legacy/fresh jobs завершились за ~2 с, steps=[], job logs отсутствуют (404).
  Причина раннего отказа инфраструктуры не установлена; нельзя объявлять CI PASS.
  [Activation](https://github.com/buzzmean/neexitlag/actions/runs/39568106123),
  [systemd](https://github.com/buzzmean/neexitlag/actions/runs/39568106065).
  Полный новый systemd/leaf сценарий и Windows на этом изменении НЕ выполнены;
  перед любым mutating deployment требуется успешный повтор этих jobs.
