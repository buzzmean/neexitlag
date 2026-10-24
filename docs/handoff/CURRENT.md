Старая история — docs/handoff/ARCHIVE.md (читать только при необходимости)

# CURRENT

## Последний итог — LF и хеши ресурсов,24.10.2026
Причина воспроизведена в отдельном Git checkout с core.autocrlf=true/core.eol=crlf: GameCatalog.json 3177байт/135CR вместо3042байт/LF;9из10 старых GameCatalogTests FAIL GAME_CATALOG_INTEGRITY_FAILED. После LF210/210 C# PASS, WPF cross-target Release0errors/4прежнихCS8604.
.gitattributes: GameCatalog.json,4assets/brand SVG сSHA изSOURCES.json, embedded LOCAL catalog/origin registry закрепленыLF; ZIP/DLL/SYS binary (WinDivert pins). Хеш GameCatalog и проверка подписей не менялись; содержимое ресурсов/игры/guards/Riot не менялись. Runtime-signed JSON сериализуется в bytes; installer hashes считаются при сборке, не hardcoded text pins.
Новые GameCatalogIntegrityTests: embedded bytes LF/exact SHA иотказ CRLF поSHA/RSA-PSS. resource-checkout-checks.ps1 воспроизводит старый CRLF, затем проверяет7LF/3binary приautocrlf=true; добавлен вVerify Windows PS5.1+PS7. PS7PASS, branding3/3PASS втомжеcheckout. Уже извлечённый CRLF Git может не переписать после сменыattrs: owner block повторно извлекает проверенные tracked resources.
НЕ ПРОВЕРЕНО: native Windows/PS5.1/GUI запуск наПК; Linux checkout эмулирует CRLF, WPF только cross-target. CI Windows не объявляетсяPASS. Go/routing/service не менялись — race/vet/WinDivert прогоны неприменимы.
Pre-push: IPC/manifest defaults неприменимы(не менялись); PS5.1-compatible args/ASCII source, JSON exchange отсутствует; реальные git/dotnet версии иexit checks, tracked-only cleanliness. DIRECT/LOCAL/RELAY, подписи, TLS/mTLS, SID/ACL/profiles, VPS/CA/SSH/коды/PR8 не затронуты. Решение24.10: сохранять точные bytes через attrs, не нормализовать подписанный ввод. Предыдущий CDN итог дословно вARCHIVE.

## Таблица работ
| Работа / дата | Файлы | Агент / ветка | Статус |
|---|---|---|---|
| CDN neutral retry + ordinary DIRECT /24.10.2026 | services/client-service browser selection/meter/origin/native; core/browserproxy; tools/windows diag; tests/docs | Codex/codex/work | В РАБОТЕ: slot release, bounded neutral retries, второй DNS DIRECT |
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
| Дизайн VELIQ / 20.10.2026 | site/, brand/, design/; будущие WPF/assets | Claude Code | Сайт согласован; перенос в WPF и новая иконка не сделаны |

## Текущая очередь и открытые вопросы
- Riot запуск/login/patch и фактические имена exe требуют Windows/ISP приёмки; каталог обычного GUI — тоже. Signed external game updates не включены; гостевой ClubWindow не менялся.
- Ordinary akashi: per-IP DIRECT_ADDRESS_FAILED вовлекает ladder раньше второго A; отдельный policy coupling, не исправлен, реальный TCP отказ неизвестен.
- Приёмка CDN: small_eof_inconclusive растёт безfailures; sibling_age/waits иdata_trials показывают допуск; strong_promotions доClose -> LOCAL10min. Нужен новый eligible flow, старый HTTP/2 не переносим. Live short/preconnect безEOF/данных может удерживать max1 до завершения; не разрываем его ради пробы.
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
- CDN (владелец24.10): scoped transform разрешает пробу; public >64KiB не барьер. Max1 НОВЫЙ unproven trial/recent RELAY media<=30s той же сети/profile/HTTPS443/3мин cooldown; strong running >=1MiB/16records/3s -> LOCAL10min; альтернатива3weak closed successes. Failure→RELAY будущих flows/backoff.
- Пробы не гарантируют незаметность player retry; реальные post-reply failures считать. Сильный trial не ждёт Close/6мин; без нового eligible flow/proof срок обучения не гарантирован.
- При работающем LOCAL новый video RELAY RX около0; старые pinned relay flows/CP/tickets/DNS DoH не входят в это обещание.
- Video12/other4 клиентские пределы и8×48/probe reserve сохранять; деньги/ресурсы: DIRECT/LOCAL перед VPS, без искусственного page PASS.
- CLAUDE.md20d8bd8 не откатывать/не дублировать; авто-порядок Claude относится к Claude Code.

## Последние команды владельцу — только GUI/ресурсы, без службы/VPS
Обычный PowerShell5.1 прежнего owner (UAC-off full-admin допустим), корень существующего клона codex/work. Нужны Git/.NET8. Передpull разрешённый reset только GameCatalog.json; прочие tracked правки сохраняются и останавливают блок. Chrome/Edge не закрывать; helper закрывает только VELIQ. После установки открыть VELIQ обычным ярлыком и страницу «Игры».
```powershell
& {
    $ErrorActionPreference='Stop'
    if(-not (Get-Command git -ErrorAction SilentlyContinue)){throw 'Install Git: winget install --id Git.Git -e'}
    if(-not (Get-Command dotnet -ErrorAction SilentlyContinue)){throw 'Install .NET 8: winget install --id Microsoft.DotNet.SDK.8 -e'}
    & dotnet --version
    if($LASTEXITCODE){throw 'dotnet is unavailable'}
    $saved=$ErrorActionPreference
    try {
        $ErrorActionPreference='Continue'
        & git --version
        if($LASTEXITCODE){throw 'git is unavailable'}
        $branch=(& git branch --show-current | Out-String).Trim()
        if($LASTEXITCODE -or $branch -ne 'codex/work'){throw 'Use existing codex/work checkout; no branch switched'}
        & git checkout -- apps/desktop-core/Resources/GameCatalog.json
        if($LASTEXITCODE){throw 'Catalog checkout failed'}
        $tracked=@(& git status --porcelain --untracked-files=no)
        if($LASTEXITCODE -or $tracked.Count){throw 'Preserve other tracked edits before update; untracked diagnostics ignored'}
        & git pull --ff-only origin codex/work
        if($LASTEXITCODE){throw 'git pull failed'}
        $resources=@(& git ls-files -- apps/desktop-core/Resources/GameCatalog.json core/localstrategy/catalog.json core/originpolicy/registry.json 'assets/brand/*.svg')
        if($LASTEXITCODE -or $resources.Count -ne 7){throw 'Expected seven pinned tracked resources'}
        foreach($path in $resources){Remove-Item -LiteralPath $path -ErrorAction Stop}
        & git checkout -- @resources
        if($LASTEXITCODE){throw 'Resource checkout failed; restore only the seven listed paths using git checkout -- <path>'}
        & git ls-files --eol -- @resources
        if($LASTEXITCODE){throw 'EOL check failed'}
    } finally {$ErrorActionPreference=$saved}
    & .\tests\windows-tools\resource-checkout-checks.ps1
    . .\tools\windows\activation-nuget.ps1
    Initialize-ActivationNuGetSource
    & dotnet test tests/desktop/NeExitLag.Client.Tests.csproj --nologo
    if($LASTEXITCODE){throw 'C# tests failed'}
    & .\tools\windows\build-desktop.ps1
    & .\tools\windows\Close-VeliqGuiForUpdate.ps1
    & .\tools\windows\install-desktop.ps1
    $built=(Get-FileHash '.\bin\release\VELIQ.exe' -Algorithm SHA256).Hash
    $installed=(Get-FileHash (Join-Path $env:LOCALAPPDATA 'Programs\VELIQ\VELIQ.exe') -Algorithm SHA256).Hash
    if($built -ne $installed){throw 'GUI_HASH_MISMATCH'}
    Write-Host 'GUI updated. Open VELIQ using its normal shortcut, then Games. Browsers stay open.'
}
```
Ожидается: resource fixturePASS;210C#PASS; GameCatalog.json i/lf w/lf attr/text eol=lf; «Игры» загружается без GAME_CATALOG_INTEGRITY_FAILED. Службу/driver/identity не обновляем. Старые CDN/ISP gates ивсе открытые вопросы выше сохраняются; команды предыдущей CDN приёмки — точечно вARCHIVE поe36e98c/short EOF.
