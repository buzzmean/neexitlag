Старая история — docs/handoff/ARCHIVE.md (читать только при необходимости)

# CURRENT

## Последний итог — CDN retry и ordinary DNS DIRECT,24.10.2026
CDN40a8a00 (merge9a7fd92): EOF callback без meter.Read/Close держал max1; regression воспроизведён иисправлен once-finalization/slot release. До3 fullSH/0<rx<64KiB нейтральных подряд разрешают следующий НОВЫЙ flow без3мин; затем ALL_SMALL_NO_DATA_LIMIT_COOLDOWN. Max1/scoped recent RELAY media<=30s обязательны каждый раз. RST/alert/timeout/post-reply errors→cooldown/откат, не success; послеorigin byte replay запрещён. LateClose не освобождает новую пробу той же группы.
Ordinary: один DIRECT_ADDRESS_FAILED ранее отправлял вобщую лестницу/RELAY до второгоDNS. Тестakashi воспроизвёл. Теперь failed ordinary AUTO сначала<=2 publicIPv4 DIRECT (900ms TLS flights/общий3s; missing native candidates→DNS<=500ms), через неизменённый address stepper Riot. Per-network/profile/host/IP negativeTTL10min; без подсетей. Second verified TLS/HTTP DIRECT иno-replay послеalert/байта покрыты. Только после адресных failures/negative TTL→LOCAL/RELAY. Холодный ordinary DIRECT безspeculative relayGET; действующий LOCAL/явные rules/RELAY floors/DoH/ambiguous/приватные IP не ослаблялись.
Diag.15: trial slot_active/released_at/reason/neutral attempt/streak/cooldown/deny; ordinary recent16из32 retryflows host/address/TCP/TLS/route/reason/TX/RX/retained/omitted всуществующем browser.local.probes<48KB. ServerHello не HTTP PASS; domain negative снимает только прежний verified background probe. Прежние9 ordinary хостов изагрегата не восстановить, нужна новаяdiag.
Проверки: полный Go race/vet иbuild Linux+Windows PASS. Один ранний полныйrace упал наdemo route25s timeout, повторполного прогона PASS (demo21.8s), demo не менялся. PinnedDLL70filters/iSCSI PASS подWine; PS7 diagnostics/legacy/secret allowlist PASS. WindowsEXE CDN/ordinary подWine PASS. Native Windows/PS5.1/ISP экономия иLOCAL2160p послеправок НЕ проверены. C#/WPF/GUI не менялись/не собирались.
Pre-push: existing browser.local.probes/native.status routed/club-readonly; no newIPC/manifest readers/defaults. PS5.1 quotedflags/BOM/nullable fields, tracked-only/native exit/hash owner block сохранены. Игры/guards/Riot-profile/catalog/TLS/mTLS/owner/SID/ACL/profiles/VPS/CA/SSH/коды/PR8 не менялись. Разрешённый merge8a7d70a безконфликтов; af1795e/12652c3 подтянуты ff-only, только app-ui/brand/design/handoff, кодслужбы/клиента не меняют. Предыдущий CDN итог дословно вARCHIVE; область освобождена.

## Таблица работ
| Работа / дата | Файлы | Агент / ветка | Статус |
|---|---|---|---|
| CDN neutral retry + ordinary DIRECT /24.10.2026 | services/client-service browser selection/meter/origin/native; core/browserproxy; tools/windows diag; tests/docs | Codex/codex/work | ОСВОБОЖДЕНО: CDN retry + ordinary DNS DIRECT/diag готовы; native Windows/ISP приёмка впереди |
| Game catalog LF integrity /24.10.2026 | .gitattributes; tests/desktop и windows-tools; verify workflow; docs/handoff | Codex/codex/work | ОСВОБОЖДЕНО: LF/hash/CRLF fixture/C#/WPF готовы; native Windows/PS5.1 впереди |
| CDN small EOF/sibling /24.10.2026 | services/client-service browser meter/selection/video usage/probe/native evidence + edge_transport.go/freeze fixture audit; tools/windows diag; tests/docs | Codex/codex/work | ОСВОБОЖДЕНО: код/лаборатория готовы; Windows/ISP приёмка впереди |
| Riot launcher / 24.10.2026 | core/profiles/localstrategy; services/client-service launcher/native/origin; platform/windows discovery; tools/windows diag; tests/docs | Codex / codex/work | ОСВОБОЖДЕНО: launcher/alternate DNS/LOCAL/diag; owner acceptance впереди |
| Game catalog / 24.10.2026 | apps/desktop-core/GameCatalog* и Resources; apps/desktop Games/MainWindow/VM integration; tests/desktop; docs | Codex / codex/work | ОСВОБОЖДЕНО: каталог/UI/тесты готовы; native Windows/club owner acceptance впереди |
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
| Дизайн VELIQ / 24.10.2026 | `app-ui/` (макеты: экран активации, каталог игр), `brand/copy.md` | Claude Code | ГОТОВО 24.10.2026, область освобождена: в макетах есть экран активации (все состояния и коды ошибок) и каталог игр с режимами и лаунчером Riot; тексты в `brand/copy.md`. Код программы не менялся. Перенос в WPF и новая иконка не сделаны |

## Текущая очередь и открытые вопросы
- Riot запуск/login/patch и фактические имена exe требуют Windows/ISP приёмки; каталог обычного GUI — тоже. Signed external game updates не включены; гостевой ClubWindow не менялся.
- Ordinary akashi: порядок второго DNS DIRECT исправлен; реальный TCP отказ и остальные9 хостов требуют новой diag. При доказанном отказе обоих адресов временный LOCAL/RELAY допустим, не обещать DIRECT при недоступном origin.
- Приёмка CDN: small_eof_inconclusive растёт безfailures; sibling_age/waits иdata_trials показывают допуск; strong_promotions доClose -> LOCAL10min. Нужен новый eligible flow, старый HTTP/2 не переносим. Terminal neutral EOF освобождает max1 сразу; live preconnect безEOF/данных может удерживать его до завершения, не разрываем ради пробы.
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
- CDN (владелец24.10): scoped transform разрешает пробу; public >64KiB не барьер. Max1 НОВЫЙ unproven trial/recent RELAY media<=30s той же сети/profile/HTTPS443/3мин cooldown кроме<=3 neutral short подряд; strong running >=1MiB/16records/3s -> LOCAL10min; альтернатива3weak closed successes. Failure→RELAY будущих flows/backoff.
- Пробы не гарантируют незаметность player retry; реальные post-reply failures считать. Сильный trial не ждёт Close/6мин; без нового eligible flow/proof срок обучения не гарантирован.
- При работающем LOCAL новый video RELAY RX около0; старые pinned relay flows/CP/tickets/DNS DoH не входят в это обещание.
- Video12/other4 клиентские пределы и8×48/probe reserve сохранять; деньги/ресурсы: DIRECT/LOCAL перед VPS, без искусственного page PASS.
- CLAUDE.md20d8bd8 не откатывать/не дублировать; авто-порядок Claude относится к Claude Code.

## Последние команды владельцу — служба/diag, без GUI rebuild/VPS
PowerShell5.1 администратора под прежним owner SID, корень клона codex/work. Нужны Git/Go и установленный pinned driver. Браузеры не закрывать; helper закрывает только VELIQ, после обновления открыть обычным ярлыком.
```powershell
& {
    $ErrorActionPreference='Stop'
    if(-not (Get-Command git -ErrorAction SilentlyContinue)){throw 'Install Git: winget install --id Git.Git -e'}
    if(-not (Get-Command go -ErrorAction SilentlyContinue)){throw 'Install Go: winget install --id GoLang.Go -e'}
    & go version
    if($LASTEXITCODE){throw 'Go is unavailable'}
    $saved=$ErrorActionPreference
    try {
        $ErrorActionPreference='Continue'
        & git --version
        if($LASTEXITCODE){throw 'Git is unavailable'}
        $branch=(& git branch --show-current | Out-String).Trim()
        if($LASTEXITCODE -or $branch -ne 'codex/work'){throw 'Use existing codex/work checkout'}
        $tracked=@(& git status --porcelain --untracked-files=no)
        if($LASTEXITCODE -or $tracked.Count){throw 'Preserve tracked edits before updating; untracked diagnostics ignored'}
        & git pull --ff-only origin codex/work
        if($LASTEXITCODE){throw 'git pull failed'}
    } finally {$ErrorActionPreference=$saved}
    & .\tools\windows\build-e2e.ps1 -DriverDirectory (Join-Path $env:ProgramData 'neexitlag\driver')
    & .\tools\windows\Close-VeliqGuiForUpdate.ps1
    & .\tools\windows\update-service.ps1
    $built=(Get-FileHash '.\bin\windows-e2e\neexitlag-service.exe' -Algorithm SHA256).Hash
    $installed=(Get-FileHash (Join-Path $env:ProgramData 'neexitlag\neexitlag-service.exe') -Algorithm SHA256).Hash
    if($built -ne $installed){throw 'SERVICE_HASH_MISMATCH'}
    . .\tools\windows\common.ps1
    Invoke-ServiceIPC 'status' | Select-Object build,connected,optimizing,capture_active
    Read-Host 'Откройте VELIQ обычным ярлыком, YouTube2160p >=2мин + перемотка; VK/Яндекс в новых вкладках, затем Enter. Chrome/Edge не закрывать'
    & .\tools\windows\diag-no-vpn.ps1 -Only browser,youtube
}
```
Смотреть .15: terminal neutral active=False/slot_released_at; neutral_short_attempt<=3, дальше cooldown>=3мин; data_trials/strong_promotions/LOCAL RX. Ordinary host/address table: второй A DIRECT доRELAY, TLS/negative TTL reason; VK/Яндекс DIRECT вnative records. Новыеeligible TCP нужны, старые HTTP/2 не переносим. Экономию/стабильность подтверждает только ISP приёмка. GUI пересобирать не нужно.
