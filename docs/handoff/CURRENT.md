Старая история — docs/handoff/ARCHIVE.md (читать только при необходимости)

# CURRENT

## Последний итог — Riot + каталог игр,24.10.2026
Riot c625f9b/bd21116: встроенный launcher AUTO,4exe в Riot Games/Riot Client, только TCP443 native; <=2 DNS A перед общей verified LOCAL/RELAY ladder. Per-network/profile/host/IP отрицательный TTL10min, не Cloudflare blanket, no replay после origin byte. Riot не считается browser readiness.
Diag.13 -Only riot: реальные выбранные route/bytes, installed exe names, actual vs foreignSNI<=2 A/host. SNI helper TLS1.2 (PS5.1), это не полная имитация Riot TLS1.3. Новый riot.status routed/club-readonly<=48KiB; поздний LOCAL retry правильно обновляет actual route.
Каталог v1:10игр, встроенный JSON/SHA256, RSA-PSS verifier для будущего owner-pinned update; unsigned imports/network update выключены, production подпись каталога не выпускалась. Монограммы вместо внешних иконок. docs/GAME_CATALOG.md.
Страница «Игры»: поиск/карточки, явные AUTO/DIRECT/DIRECT+/RELAY, копирование existing profile+metadata, repair только Riot. DIRECT+ лишь готовый gameudp с exactpath/guard; остальные не получают выдуманную стратегию. AUTO игр DIRECT; Vanguard/FACEIT/club permissions сохраняются. RELAY warning до выбора, без обещания ping.
Go full race/vet/build Linux+Windows и Windows vet PASS после финального Go-fix; pinned DLL70filters/iSCSI/checksums PASS подWine,5Riot Windows EXE/Wine PASS. C#208/208 иWPF build0errors/4прежнихCS8604 PASS; PS7 fixtures/parser PASS. Native Windows/PS5.1/GUI runtime/Riot login+patch/ISP НЕ проверены.
Pre-push: новые/использованные IPC routed; service.json readers/defaults не менялись; PS5.1-compatible diagnostics/BOM, quoted flags; owner block безgit switch, native exit/hash checks. Экономия: здоровый DIRECT DNS A раньше LOCAL/RELAY. Секреты/драйвер pin/фильтры/iSCSI/игровые guards/VPS/CA/коды/SSH/PR8 не менялись.
Решение: каталог не обходит CLUB_ADMIN_MANAGED, гостевой ClubWindow остаётся прежним; внешний catalog update требует отдельного доверенного public key/release workflow. Обновить service+GUI, проверить обычный GUI/Riot безwinws и -Only riot,browser. Предыдущие итоги дословно вARCHIVE.

## Таблица работ
| Работа / дата | Файлы | Агент / ветка | Статус |
|---|---|---|---|
| CDN small EOF/sibling /24.10.2026 | services/client-service browser meter/selection/video usage/probe evidence + freeze fixture audit; tools/windows diag; tests/docs | Codex/codex/work | В РАБОТЕ: small EOF with retransmits, recent scoped sibling, meaningful trial evidence |
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
| Дизайн VELIQ / 20.10.2026 | site/, brand/, design/; будущие WPF/assets | Claude Code | Сайт согласован; перенос в WPF и новая иконка не сделаны |

## Текущая очередь и открытые вопросы
- Riot запуск/login/patch и фактические имена exe требуют Windows/ISP приёмки; каталог обычного GUI — тоже. Signed external game updates не включены; гостевой ClubWindow не менялся.
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

## Последние команды владельцу — клиент, без VPS/SSH
PowerShell5.1 администратора под прежним owner SID из корня codex/work; Git/Go/.NET8. Chrome/Edge/игры не закрываются скриптом; только GUI VELIQ. GUI после обновления открывается обычным ярлыком (UAC-off full-admin — прежнее исключение).
```powershell
& {
    $ErrorActionPreference='Stop'
    foreach($tool in @('git','go','dotnet')) {
        if(-not (Get-Command $tool -ErrorAction SilentlyContinue)) { throw ('Нужен '+$tool+': Git.Git / GoLang.Go / Microsoft.DotNet.SDK.8 через winget install') }
    }
    & go version; if($LASTEXITCODE){throw 'Go недоступен'}
    & dotnet --version; if($LASTEXITCODE){throw '.NET SDK недоступен'}
    $savedPreference=$ErrorActionPreference
    try {
        $ErrorActionPreference='Continue'
        $branch=(& git branch --show-current | Out-String).Trim()
        if($LASTEXITCODE -or $branch -ne 'codex/work'){throw 'Нужен существующий codex/work; ветка не переключалась'}
        & git pull --ff-only origin codex/work
        $pullExit=$LASTEXITCODE
    } finally { $ErrorActionPreference=$savedPreference }
    if($pullExit){throw 'git pull failed'}
    . .\tools\windows\activation-nuget.ps1
    Initialize-ActivationNuGetSource
    & .\tools\windows\build-e2e.ps1 -DriverDirectory (Join-Path $env:ProgramData 'neexitlag\driver')
    & .\tools\windows\build-desktop.ps1
    & .\tools\windows\Close-VeliqGuiForUpdate.ps1
    & .\tools\windows\update-service.ps1
    & .\tools\windows\install-desktop.ps1
    $pairs=@(
        @('.\bin\windows-e2e\neexitlag-service.exe',"$env:ProgramData\neexitlag\neexitlag-service.exe"),
        @('.\bin\release\VELIQ.exe',"$env:LOCALAPPDATA\Programs\VELIQ\VELIQ.exe")
    )
    foreach($pair in $pairs){if((Get-FileHash $pair[0] -Algorithm SHA256).Hash -ne (Get-FileHash $pair[1] -Algorithm SHA256).Hash){throw ('HASH_MISMATCH '+$pair[1])}}
    . .\tools\windows\common.ps1
    Invoke-ServiceIPC 'status' | Select-Object build,connected,optimizing,capture_active
    Invoke-ServiceIPC 'profiles' | Where-Object {$_.id -eq 'riot-client'} | Select-Object id,name,mode,executables
    Read-Host 'Откройте VELIQ обычным ярлыком: Игры/поиск/Riot. Без winws запустите Riot, проверьте login/patch; затем Enter. Chrome/Edge не закрывать'
    & .\tools\windows\diag-no-vpn.ps1 -Only riot,browser
}
```
Ожидание: Riot AUTO, реальный обмен; исправный второй A DIRECT раньше relay, не всем auth/CDN RELAY. TLS/HTTP404 корня может быть нормальным control endpoint, не доказательством успешного login/patch.
Принимать browser packet/byte evidence за page PASS нельзя. Игры по умолчанию DIRECT; prepared PUBG UDP — отдельное прежнее исключение, античит не трогается. После выключения оптимизации браузеры без restart.
