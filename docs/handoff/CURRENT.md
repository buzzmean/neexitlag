Старая история — docs/handoff/ARCHIVE.md (читать только при необходимости)

# CURRENT

## Последний итог — CDN media admission/identity, ordinary evidence,25.10.2026
База48a3e8f;8b7b2e2 docs/research подтянут ff-only, CLAUDE/wiki прочитаны без принятия гипотез за ISP facts. Прежний итог дословно вARCHIVE.
A: приOPEN нельзя знать encrypted request size; после TLS app-write менять egress запрещено. Теперь LOCAL CDN trial требует недавний RELAY RX>1MiB того жеhost/распознанногоsn-cluster (network/profile/443/<=30s), плюс прежние transform/sibling/max1/cooldown/neutral limits. rrN не classifier; unclassified не общийcluster. Без новогоeligible TCP проб может не быть, H2 не ломаем, экономия не доказана.
B: CDN LOCAL failure сохраняет failures/after-reply/cooldown, но negativeTTL10min ограничен host/IP/strategy+network/profile/subgroup(max256); qualification/score всей группы не отравляется. Плохойendpoint→RELAY новыхflows, другие могутLOCAL. IP/гео/DPI причина из203байт+silence не установлена.
C: .15 не отличает новыйtrial того жеhost отстарого. Добавлены trial_id/record_id/active_host/IP/started/result/matches; terminal/Close/promotion сверяют exposureID. Дополнительно воспроизведён hard-callback безRead/Close, который держалslot: теперь release сразу, counters не скрыты; старое завершение не снимает новыйtrial.
D: budget сначала удалялвсюordinaryтаблицу, script ложноNONE_RECORDED — воспроизведено/исправлено. Приоритет compact evidence, retained/omitted ифактическиеordinary_ladder hosts/IP/profile/stage. Старый ordinary_direct_fallback_attempts считал толькоcapacity, неDNS! .16 отдельныеDNS attempts/scope denials; guards явнообъясняют отказ(ambiguous/noTLS/rule...). Маршруты ordinary/игр/Riot не переписаны; рабочий2йDNS DIRECT тестами сохранён. Какиеименно12 старыхflows — изагрегата не узнать; доupdate owner block выводит retained records.
Проверки: полный Go race/vet/build Linux+Windows PASS; WindowsEXE CDN/ordinary/full TLS HTTP tests подWine PASS; pinnedDLL всефильтры/iSCSI PASS подWine; PS7 legacy/identity/bulk/omission/secret allowlist PASS. Native Windows/PS5.1/Chrome/ISP LOCAL иреальнаяэкономия НЕ проверены. GUI/C#/WPF не менялись/не собирались.
Pre-push: existing browser.local.probes/native.status routed/club-readonly, no newRPC/manifest readers/defaults; PS5.1-compatible syntax/BOM/quotedflags/optional fields/owner tracked-only/native exit/hash checks. TLS/mTLS/owner/SID/ACL/profiles/no-replay/protected/filters/игры/guards/Riot-profile/catalog/VPS/CA/SSH/коды/PR8 не менялись. Разрешён только клиентский update, не deployment. Область освобождена; docs/YOUTUBE_LOCAL_FIRST.md иLOCAL_ENGINE_AUDIT.md содержат обоснование/ограничения.

## Таблица работ
| Работа / дата | Файлы | Агент / ветка | Статус |
|---|---|---|---|
| CDN identity/media evidence + ordinary dispatch /25.10.2026 | services/client-service browser selection/meter/native/ordinary; tools/windows diag; tests/docs | Codex/codex/work | ОСВОБОЖДЕНО: media admission/endpoint/identity иevidence готовы; native Windows/ISP впереди |
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
- Ordinary .16: сверить compact host/IP ladder rows, DNS DIRECT attempts иscope denials. .15 NONE_RECORDED могло означатьусечение; старыйcapacity counter неDNS. VK/IP/гео/DPI причины не установлены, безданных не менятьполитику.
- Приёмка CDN .16: требуется >1MiB recent RELAY того жеhost/sn-cluster; сверить exposureIDs, data_trials/strong_promotions/LOCAL RX иendpoint TTL. Новый TCP обязателен; старый H2/preconnect безterminal не разрывать, success не гарантирован.
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
- CDN (владелец24.10): scoped transform разрешает пробу; public >64KiB не барьер. Max1 НОВЫЙ unproven trial/bulk>1MiB того жеhost/sn-cluster +recent RELAY media<=30s той же сети/profile/HTTPS443/3мин cooldown кроме<=3 neutral short подряд; strong running >=1MiB/16records/3s -> LOCAL10min; альтернатива3weak closed successes. Failure→RELAY будущих flows/backoff.
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
    . .\tools\windows\common.ps1
    try {
        $before=Invoke-ServiceIPC 'browser.native.status'
        $lp=$before.PSObject.Properties['youtube_local_ladder']
        if($lp){
            $rp=$lp.Value.PSObject.Properties['records']
            if($rp){$rp.Value | Where-Object {$p=$_.PSObject.Properties['service'];$p -and $p.Value -eq 'ordinary'} | Select-Object domain,destination,profile,stage,reason,TX,RX | Format-Table -AutoSize}
        }
    } catch {Write-Host 'Previous ordinary history unavailable; update continues'}
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
Смотреть .16: active_trial_id/record_id противID закрытия, active_host/IP; media_rx>1MiB иSCOPED_BULK_RELAY_NOT_OBSERVED; data_trials/LOCAL RX иendpoint negatives. Ordinary DNS attempts/scope_denials/retained/omitted/ladder hosts показывают реальныйпуть. ФильтрацияIP/гео/DPI безизмерений не доказана. Передupdate печатается стараяretained ordinary история, не весьaggregate. GUI не пересобирать; 2мин не гарантируют eligible новыйTCP/LOCAL.
