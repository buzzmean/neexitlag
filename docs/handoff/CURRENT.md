Старая история — docs/handoff/ARCHIVE.md (читать только при необходимости)

# CURRENT

## Последний итог — Riot launcher,24.10.2026
Добавлен built-in Riot Client AUTO,4 launcher exe внутри Riot Games/Riot Client; games/vgc/vgk/anti-cheat не наследуют профиль. TCP443 new-TCP native/SNI/owner guards, без PAC.
Riot hostname: сначала<=2 DNS IPv4, silent ClientHello -> другой адрес, лишь затем общая verified LOCAL/RELAY selection.10мин per-network/profile/host/IP negative, без permanent/subnet pin; после origin byte replay запрещён. Auth/CDN с рабочим DIRECT остаются DIRECT.
Отдельный riot.status/club read handler и diag.13 -Only riot: выбранный путь/bytes LOCAL/RELAY, actualSNI vs foreignSNI<=2 A/host, TLS verification ON, без credentials. Riot не считается Browser ready.
Go race/vet/build Linux+Windows и Windows vet PASS; PS7 Riot formatter PASS; pinned DLL70filters/iSCSI/checksums PASS подWine; Riot Windows EXE/Wine PASS. Native Windows/PS5.1/Riot login/manifest/ISP НЕ проверены. VPS/CA/коды/SSH не использовались.
Решение: не обобщать один silent IP на Cloudflare subnet; вместо этого пробовать другой DNS A и хранить ограниченный отрицательный endpoint verdict. Подробности docs/LOCAL_ENGINE_AUDIT.md.
Следующий этап: каталог игр/минимальный UI; C#-изменения ещё не входят в Riot commit. Pre-push: riot.status routed/readonly allowlist, manifest readers не менялись, PS5.1-compatible formatter/BOM/quoted flags; существующие ограничения сохранены.

## Таблица работ
| Работа / дата | Файлы | Агент / ветка | Статус |
|---|---|---|---|
| Riot launcher / 24.10.2026 | core/profiles/localstrategy; services/client-service launcher/native/origin; platform/windows discovery; tools/windows diag; tests/docs | Codex / codex/work | ОСВОБОЖДЕНО: launcher/alternate DNS/LOCAL/diag; owner acceptance впереди |
| Game catalog / 24.10.2026 | apps/desktop-core/GameCatalog* и Resources; apps/desktop Games/MainWindow/VM integration; tests/desktop; docs | Codex / codex/work | В РАБОТЕ: встроенные данные/страница/профили, после Riot commit |
| LOCAL engine / 24.10.2026 | core/localstrategy; platform/windows native/DNS/TCP inject; services/client-service browser* и минимальный controller/club IPC dispatch; tools/windows diag; tests/workflow/docs | Codex / codex/work | ОСВОБОЖДЕНО: код/лаборатория готовы; Windows/ISP приёмка не доказана |
| CDN LOCAL recovery / 24.10.2026 | core/localstrategy inspection; core/browserproxy/probe.go; platform/windows TCP evidence; services/client-service browser selection/probes; tools/windows diag; tests/docs | Codex / codex/work | ОСВОБОЖДЕНО: selector/probes/telemetry готовы; требуется owner trace для EOF и CDN success |
| CDN neutral close / 24.10.2026 | services/client-service browser meter/selection/evidence; core/browserproxy TLS metadata/midstream; core/localstrategy inspection; platform/windows TCP evidence; tools/windows diag; tests/docs | Codex / codex/work | ОСВОБОЖДЕНО: реализация/тесты готовы; ISP/Chrome intent и CDN success требуют приёмки |
| Sustained CDN promotion / 24.10.2026 | services/client-service browser meter/selection/evidence и primary fixture; core/browserproxy TLS metadata; tools/windows diag; tests/docs | Codex / codex/work | ОСВОБОЖДЕНО: promotion/release/rollback/diag/fixture готовы; native Windows/ISP приёмка впереди |
| Компактный handoff / 24.10.2026 | docs/handoff/*, docs/AGENT_RULES.md | Codex / codex/work | ОСВОБОЖДЕНО: архив и правила готовы; только документы |
| Экономия VPS / LOCAL каталог и автоподбор 1–7 (23.10.2026) | Детали — [ARCHIVE.md](ARCHIVE.md) | Codex / codex/work | ОСВОБОЖДЕНО; код готов, Windows/ISP gates см. ниже |
| Multi-relay/Hostkey P0 (23.10.2026) | Детали — [ARCHIVE.md](ARCHIVE.md) | Codex / codex/work | ОСВОБОЖДЕНО; код готов, Windows/ISP gates см. ниже |
| YouTube LOCAL-first A–G (23.10.2026) | Детали — [ARCHIVE.md](ARCHIVE.md) | Codex / codex/work | ОСВОБОЖДЕНО; код готов, Windows/ISP gates см. ниже |
| YouTube relay скорость/TTFB (23.10.2026) | Детали — [ARCHIVE.md](ARCHIVE.md) | Codex / codex/work | ОСВОБОЖДЕНО; код готов, Windows/ISP gates см. ниже |
| YouTube admission/ordinary Google/build PS5.1 A–D (23.10.2026) | Детали — [ARCHIVE.md](ARCHIVE.md) | Codex / codex/work | ОСВОБОЖДЕНО; код готов, Windows/ISP gates см. ниже |
| Browser DNS/DoH и opaque recovery A–E (23.10.2026) | Детали — [ARCHIVE.md](ARCHIVE.md) | Codex / codex/work | ОСВОБОЖДЕНО; код готов, Windows/ISP gates см. ниже |
| Срочно: компиляция фильтров WinDivert/iSCSI (22.10.2026) | Детали — [ARCHIVE.md](ARCHIVE.md) | Codex / codex/work | ОСВОБОЖДЕНО; код готов, Windows/ISP gates см. ниже |
| PUBG: scope античита/local-first/status (22.10.2026) | Детали — [ARCHIVE.md](ARCHIVE.md) | Codex / codex/work | ОСВОБОЖДЕНО; код готов, Windows/ISP gates см. ниже |
| Пилот клуба: Setup/image/boot/status A–F (22.10.2026) | Детали — [ARCHIVE.md](ARCHIVE.md) | Codex / codex/work | ОСВОБОЖДЕНО; код готов, Windows/ISP gates см. ниже |
| Browser opaque/ECH DIRECT и Telegram recovery (22.10.2026) | Детали — [ARCHIVE.md](ARCHIVE.md) | Codex / codex/work | ОСВОБОЖДЕНО; код готов, Windows/ISP gates см. ниже |
| Общий игровой UDP обход A–H (22.10.2026) | Детали — [ARCHIVE.md](ARCHIVE.md) | Codex / codex/work | ОСВОБОЖДЕНО; код готов, Windows/ISP gates см. ниже |
| Native Chrome/Edge: нулевой ingress d3aad37 (22.10.2026) | Детали — [ARCHIVE.md](ARCHIVE.md) | Codex / codex/work | ОСВОБОЖДЕНО; код готов, Windows/ISP gates см. ниже |
| Браузеры fresh/hot-attach A–G (22.10.2026) | Детали — [ARCHIVE.md](ARCHIVE.md) | Codex / codex/work | ОСВОБОЖДЕНО; код готов, Windows/ISP gates см. ниже |
| Регрессия выключения A–E (6adefe5) | Детали — [ARCHIVE.md](ARCHIVE.md) | Codex / codex/work | ОСВОБОЖДЕНО; код готов, Windows/ISP gates см. ниже |
| A1–A6, B1, C1–C2, D1–D2, E1–E2 | Детали — [ARCHIVE.md](ARCHIVE.md) | Codex / codex/work | ОСВОБОЖДЕНО; код готов, Windows/ISP gates см. ниже |
| Закрытые работы до 21.10.2026; передачи Claude Code → Codex | [История и исходные строки](ARCHIVE.md) | Claude Code / Codex | История сохранена, активное владение не переносится |
| Дизайн VELIQ / 20.10.2026 | site/, brand/, design/; будущие WPF/assets | Claude Code | Сайт согласован; перенос в WPF и новая иконка не сделаны |

## Текущая очередь и открытые вопросы
- Ordinary akashi: per-IP DIRECT_ADDRESS_FAILED вовлекает ladder раньше второго A; отдельный policy coupling, не исправлен, реальный TCP отказ неизвестен.
- Приёмка CDN: strong_promotions/trial_success растут ДО Close, active slot свободен/занят новой группой; LOCAL TTL10min и новый LOCAL RX основной; поздний failure отзывает будущие opens.2160p/seek, игры DIRECT.
- HTTP80, QUIC fake и Discord voice UDP LOCAL не реализованы; fakedsplit TCP443 реализован, весь Flowseal не заявляется.
- Подбор DPI продолжается ограниченно; пока нет proof — временный authenticated RELAY. Short response не доказывает >64KiB.
- Общий клубный picker и server NIC month/forecast подготовлены, production endpoints/flags выключены: per-PC fallback и monthly UNKNOWN.
- Их включение требует отдельного разрешения на CP/relay и отдельные приватные state dirs; client LOCAL не требует нового deployment.
- Широкие multi-relay A–E отложены; минимальный primary/standby реализован. Балансировка/нагрузочный инструмент — лишь при необходимости.
- Hostkey146.0.77.98 = VPS-PRIMARY принят владельцем; CP/activation и standby2.56.177.4 до12.11.2026. Не повторять P0/CA rotation/коды.
- После12.11 запасной сервер требует решения владельца; новый signed ID сохраняет собственный ticket audience.
- Независимый WFP/driver crash-guard остаётся будущим вариантом B: docs/BROWSER_NATIVE_GUARD.md; не выдавать WinDivert за независимый guard.
- Исторический вопрос полного удаления PAC при capture failure не считать закрытым без проверки актуального lifecycle; owned/чужие policy не смешивать.
- svchost DNS/custom DoH/ECH/IPv6 имеют ограничения: docs/BROWSER_NATIVE_INGRESS.md и BROWSER_ORIGIN_POLICY.md; скрытый домен не считать установленным.
- Installer/подпись/SmartScreen, общий freeze-образ, защита клубных настроек и перенос дизайна в WPF требуют отдельной приёмки/решений владельца.
- Временные test-коды должны быть удалены до релиза; production revoke-tests не запускать: docs/ACTIVATION_OWNER_RUNBOOK.md.

## НЕ ПРОВЕРЕНО / доказательства
- Реальные Windows/PS5.1/WinDivertOpen/kernel injection, ISP LOCAL1080p, Discord voice LOCAL, failover Hostkey и40ПК в последнем коде не проверены.
- Setup, 2–3ПК diskless/freeze, BattlEye/PUBG матчи/RTT, FACEIT/Vanguard совместимость и все ещё не подтверждённые owner gates не объявлять PASS.
- Go race/vet/build Linux+Windows PASS; PG17 изолированный PASS;100Python иPS7 PASS; DLL70filters/iSCSI30/checksums13 подWine PASS — не native Windows.
- C#198/198 иWPF0errors/4прежнихCS8604 после merge; в6–7 C# не менялся и заново WPF не собирали.
- CI1cdb650 Verify40090200087: failure, core/windows/Chrome-native/Edge steps=[]; CI не исполнился. Ранее billing gate подтверждён владельцем; новый текст причины не получен.
- Artifact quota не назначать причиной без evidence: https://github.com/buzzmean/neexitlag/actions/runs/40090200087
- На6df8a42 владелец подтвердил native capture и public128KiB LOCAL(c0aa53), но CDN0/4; Telegram/Anthropic ранее работали. CDN/Discord voice не объявлять PASS.
- Исторические непроверенные gates остаются в ARCHIVE.md; при работе над соответствующей областью искать их точечно, не считать закрытыми.

## Действующие запреты и инварианты
- Все docs/AGENT_RULES.md обязательны: VPS, /etc/neexitlag, БД, CA/credentials/activation не менять и команды там не запускать без явного разрешения.
- PR8/main/codex/direct-plus-flowseal не трогать; merge только по разрешению; force-push/rebase/переписывание истории запрещены.
- Никаких секретов/кодов/PEM/identity/токенов в выводе или git; diag-result/cs-diag/observe-result не коммитить.
- TLS/mTLS/signature/ticket replay/30min lease не ослаблять; каждый новый QUIC session использует свежий one-shot ticket и свой audience.
- Protected fail-closed, DNS53 не уходит в relay, собственный/relay трафик не перехватывать; без MITM.
- Existing TCP/UDP pinned; после любого origin byte нельзя replay/migrate TLS. Потерянный established TCP не переносится на другой egress.
- Новые authenticated opens: standby hold30s и sustained primary recovery30s со свежими echo; saturation/один origin не отравляет весь VPS.
- Браузеры не закрывать/не перезапускать, никаких диалогов/policy reload; GUI без admin на обычном UAC ПК, проверенный UAC-off режим отдельный.
- Сохранять профили/identity/SID/ACL/IPC/config при обновлении; чужую policy не трогать; SCM/IPC/пути/ALPN/module не переименовывать.
- iSCSI3260 bypass сохранять; никаких not (...) в фильтрах, все через pinned DLL compiler; WinDivert pin/зависимости без разрешения не менять.
- Игры AUTO DIRECT; FACEIT/Vanguard guard не ослаблять, PUBG scope отдельно; без инъекции в процессы/файлы/память игры, без внешнего winws.
- CDN (владелец24.10): scoped transform разрешает пробу; public >64KiB не барьер. Max1 НОВЫЙ unproven trial/live RELAY sibling/3мин cooldown; strong running >=1MiB/16records/3s -> LOCAL10min; альтернатива3weak closed successes. Failure→RELAY будущих flows/backoff.
- Пробы не гарантируют незаметность player retry; реальные post-reply failures считать. Сильный trial не ждёт Close/6мин; без нового eligible flow/proof срок обучения не гарантирован.
- При работающем LOCAL новый video RELAY RX около0; старые pinned relay flows/CP/tickets/DNS DoH не входят в это обещание.
- Video12/other4 клиентские пределы и8×48/probe reserve сохранять; деньги/ресурсы: DIRECT/LOCAL перед VPS, без искусственного page PASS.
- CLAUDE.md20d8bd8 не откатывать/не дублировать; авто-порядок Claude относится к Claude Code.

## Последние команды владельцу — обновление клиента, без VPS deployment
PowerShell администратора под прежним owner SID, из корня клона; Git/Go в PATH и установленный pinned driver. Только GUI VELIQ закрывается; Chrome/Edge не закрывать.
```powershell
& {
    $ErrorActionPreference='Stop'
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) { throw 'Install Go: winget install GoLang.Go' }
    & go version
    if ($LASTEXITCODE) { throw 'Install Go: winget install GoLang.Go' }
    if (-not (Get-Command git -ErrorAction SilentlyContinue)) { throw 'Install Git: winget install Git.Git' }
    $gitPreference=$ErrorActionPreference
    try {
        $ErrorActionPreference='Continue'
        $branch=(& git branch --show-current | Out-String).Trim()
        if ($LASTEXITCODE -or $branch -ne 'codex/work') { throw 'Use existing codex/work checkout; no branch was switched' }
        & git pull --ff-only origin codex/work
        $pullExit=$LASTEXITCODE
    } finally { $ErrorActionPreference=$gitPreference }
    if ($pullExit) { throw 'git pull failed' }
    & .\tools\windows\build-e2e.ps1 -DriverDirectory (Join-Path $env:ProgramData 'neexitlag\driver')
    & .\tools\windows\Close-VeliqGuiForUpdate.ps1
    & .\tools\windows\update-service.ps1
    . .\tools\windows\common.ps1
    Invoke-ServiceIPC 'status' | Select-Object build,connected,optimizing,capture_active
    $a=(Get-FileHash '.\bin\windows-e2e\neexitlag-service.exe' -Algorithm SHA256).Hash
    $b=(Get-FileHash "$env:ProgramData\neexitlag\neexitlag-service.exe" -Algorithm SHA256).Hash
    if ($a -ne $b) { throw 'SERVICE_HASH_MISMATCH' }
    Read-Host 'YouTube2160p >=2мин и перемотка в открытом Chrome; затем Enter (Chrome/Edge не закрывать)'
    & .\tools\windows\relay-status.ps1
    & .\tools\windows\relay-traffic.ps1
    & .\tools\windows\diag-no-vpn.ps1 -Only browser,youtube
}
```
Нужны READY/optimizing и новые flows. DPI LOCAL только после proof; VK/Яндекс DIRECT; Telegram/IG/FB/OpenAI/Anthropic RELAY_REQUIRED.
Холодный CDN: strong proof >=1MiB/16 completed records/3s даёт LOCAL10min безClose.2мин не гарантируют создание eligible trial; старые RELAY flows pinned. После обучения новый LOCAL RX должен стать основным.
В diag смотреть strong_promotions/trial_success/trial_result, active_state/age/admission, small_eof_inconclusive/post_promotion_failures, LOCAL failure timing/volume и новый LOCAL RX. TrialFailure — не число всех поздних отказов.
Discord public WSS/redirect CONTROL_NOT_BULK не доказывают авторизованный gateway/voice. Затем выключить оптимизацию → VK без restart, игры DIRECT.
Monthly data пока UNKNOWN; серверные read-only команды — docs/RELAY_TRAFFIC_BUDGET.md, deployment здесь не разрешён.
