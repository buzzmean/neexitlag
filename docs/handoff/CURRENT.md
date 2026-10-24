Старая история — docs/handoff/ARCHIVE.md (читать только при необходимости)

# CURRENT

## Последний итог — CDN trial/EOF,24.10.2026
Владелец6df8a42 подтвердил www.youtube.com/discord.com128KiB LOCAL(c0aa53…), но CDN0/4/3post-reply failures: успех googlevideo всё ещё НЕ доказан.
Исправлен selector: same network/profile/hint-mode sustained YouTube proof выше transform-only/catalog prior; точный Qualified ID виден в diag; failed CDN candidate не форсируется снова в текущем sweep.
Фон не меняет binding/health активного trial; TrialAfter отдельно сохраняет >=3мин независимо от requalification; max1/live RELAY sibling/3sustained CDN success сохранены. Existing TCP pinned, после origin byte replay запрещён.
Diag.10: фактические ACK/TS deltas/MD5; bounded LOCAL close table с EOF/reset, FIN/RST направлением/первым TTL/SYNACK TTL/ретрансмитами, отдельно от active RELAY records и в пределах IPC budget.
Для fooling=ts seq_delta0/1 и original TTL128 ожидаемы: badseq/badack неактивны; ts_delta должен быть-600000. Генерация fake/old-sequence legacy не менялась.
16PNG давали~21KiB: теперь≤64 GET/один TLS/4с/256KiB до >64KiB. Пустой301 не IO failure; canonical youtube.com redirect — CONTROL_NOT_BULK, не CDN/video PASS. i.ytimg404 не успех.
A: конкретный EARLY_EOF на ПК пока НЕ объяснён; новая packet evidence нужна, чтобы отличить remote FIN/RST/локальное закрытие/короткий ответ. TTL не доказывает DPI, HTTP CDN без MITM не виден.
Полные Go race/vet/build Linux+Windows и Windows vet PASS; обе Windows fixture EXE собраны, pinned DLL70filters/iSCSI30/checksum44plans подWine PASS; PS7 PASS. Windows kernel/PS5.1/ISP CDN не выполнялись здесь. GUI/CP/shared flags/VPS/CA/коды/SSH/PR8 не трогали.
Аудит: docs/LOCAL_ENGINE_AUDIT.md (CDN приёмка6df8a42); прежний итог дословно в ARCHIVE. IPC использует прежний browser.local.probes dispatch; manifest/defaults не менялись, DIAG read-only, RELAY не форсируется вместо рабочего LOCAL.

## Таблица работ
| Работа / дата | Файлы | Агент / ветка | Статус |
|---|---|---|---|
| LOCAL engine / 24.10.2026 | core/localstrategy; platform/windows native/DNS/TCP inject; services/client-service browser* и минимальный controller/club IPC dispatch; tools/windows diag; tests/workflow/docs | Codex / codex/work | ОСВОБОЖДЕНО: код/лаборатория готовы; Windows/ISP приёмка не доказана |
| CDN LOCAL recovery / 24.10.2026 | core/localstrategy inspection; core/browserproxy/probe.go; platform/windows TCP evidence; services/client-service browser selection/probes; tools/windows diag; tests/docs | Codex / codex/work | ОСВОБОЖДЕНО: selector/probes/telemetry готовы; требуется owner trace для EOF и CDN success |
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
- Приёмка CDN нового кода: сначала scoped c0aa53 proof и >=3min trials; при EOF читать LOCAL close/actual fooling table. После LOCAL —2160p2мин/seek; игры DIRECT.
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
- CDN (решение владельца24.10): тот же scoped transform на любом YouTube host разрешает только пробу; public >64KiB больше не обязательный барьер; max1 real LOCAL trial с живым RELAY sibling, >=3мин cooldown/подгруппу,3успеха подряд; сбой→RELAY новых flows/backoff.
- Пробы не гарантируют незаметность player retry; post-reply failures считать. Для холодного CDN двух минут недостаточно (минимум6мин+завершение проб).
- При работающем LOCAL новый video RELAY RX около0; старые pinned relay flows/CP/tickets/DNS DoH не входят в это обещание.
- Video12/other4 клиентские пределы и8×48/probe reserve сохранять; деньги/ресурсы: DIRECT/LOCAL перед VPS, без искусственного page PASS.
- CLAUDE.md20d8bd8 не откатывать/не дублировать; авто-порядок Claude относится к Claude Code.

## Последние команды владельцу — обновление клиента, без VPS deployment
PowerShell администратора под прежним owner SID, из корня клона; Git/Go в PATH и установленный pinned driver. Только GUI VELIQ закрывается; Chrome/Edge не закрывать.
```powershell
& {
    $ErrorActionPreference='Stop'
    & go version
    if ($LASTEXITCODE) { throw 'Install Go: winget install GoLang.Go' }
    git switch codex/work
    if ($LASTEXITCODE) { throw 'git switch failed' }
    git pull --ff-only origin codex/work
    if ($LASTEXITCODE) { throw 'git pull failed' }
    $m=Get-Content -Raw "$env:ProgramData\neexitlag\service.json" | ConvertFrom-Json
    & .\tools\windows\build-e2e.ps1 -DriverDirectory $m.driver.directory
    & .\tools\windows\Close-VeliqGuiForUpdate.ps1
    & .\tools\windows\update-service.ps1
    . .\tools\windows\common.ps1
    Invoke-ServiceIPC 'status' | Select-Object build,connected,optimizing,capture_active
    $a=(Get-FileHash '.\bin\windows-e2e\neexitlag-service.exe' -Algorithm SHA256).Hash
    $b=(Get-FileHash "$env:ProgramData\neexitlag\neexitlag-service.exe" -Algorithm SHA256).Hash
    if ($a -ne $b) { throw 'SERVICE_HASH_MISMATCH' }
    Read-Host 'YouTube2160p >=6мин/seek в уже открытом Chrome; Discord/CDN, Telegram/ChatGPT/Claude, VK/Яндекс; затем Enter'
    & .\tools\windows\relay-status.ps1
    & .\tools\windows\relay-traffic.ps1
    & .\tools\windows\diag-no-vpn.ps1 -Only browser,youtube
}
```
Нужны READY/optimizing и новые flows. DPI LOCAL только после proof; VK/Яндекс DIRECT; Telegram/IG/FB/OpenAI/Anthropic RELAY_REQUIRED.
Холодный CDN:3trial/>64KiB с cooldown>=3мин (≥6мин+закрытие),2мин просмотра недостаточно для гарантии LOCAL. После обучения —2мин2160p/seek и новый googlevideo LOCAL RX основной.
В diag смотреть LOCAL probes commit/actual packets/ошибки, group stage/strategy/LOCAL-RELAY RX и failed_after_reply; старые pinned RELAY flows не считать регрессией.
Discord public WSS/redirect CONTROL_NOT_BULK не доказывают авторизованный gateway/voice. Затем выключить оптимизацию → VK без restart, игры DIRECT.
Monthly data пока UNKNOWN; серверные read-only команды — docs/RELAY_TRAFFIC_BUDGET.md, deployment здесь не разрешён.
