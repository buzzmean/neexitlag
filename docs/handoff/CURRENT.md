Старая история — docs/handoff/ARCHIVE.md (читать только при необходимости)
План до релиза и порядок этапов R1–R7 — docs/ROADMAP.md (читать в начале каждой задачи).

# CURRENT

## Последний итог — M1–M5/C1–C6,25.10.2026
Фаза0 0b505da; база07e65b0 поверх8b7b2e2. CLAUDE/AGENT_RULES/protocol/wiki прочитаны; wiki неISP facts. Прежнийитог дословно вARCHIVE.
M1 peer-decoded localQUIC: IP→пустойSNI, legacy/1, ticket/3+/2, TLS1.3, JA4-like параметры безCH bytes/secret; Go1.25.3, нефакт fingerprints Chrome/ISP. M2 pool session_creation attempts/created60s/max/concurrent/start+replacement+capacity+rotation/factory vs handshake. Lab2sessions/max1 concurrent, не40ПК.
M3 network-check.exe изbuild-e2e: system/UDP/TCP/verifiedDoT+DoH(IP dial)/AA/AD/TC/RCode/ID-question, <=2IP TCP/TLSrealforeign/publicHTTP128KiB4s. Existing controls only, no creds/redirects/TLS bypass. Diag.16 measurement revision2: активныепробы-Only layers/default, browser+youtubepassive. NXDOMAIN+aa/timeout/reset≠доказаннаяподмена/SYNdrop/DPI.
M4/M5 boundedretainedLOCAL census ts/payload/SYN negotiatedUNKNOWN/required, relativeearlyretrans/RST/latewireRXfails; безSYN hotattachUNKNOWN. НельзяHTTPredirect внутриTLS, raw/absolute seq неэкспортируются. C1 permanent instrumentation/tests. C2/C3/C4 отложены донастоящихM4/earlyfail/nonbrowserDNS данных; purets безопцииужеskip, netsh ненужен. Лестница/кадрирование/guards/max1/3мин/no-replay не менялись.
C5 толькоdocs/RELAY_TRANSPORT_EVOLUTION.md: additiveALPN/version/SNIownSAN/TCP+TLS flagsOFF/compatmatrix/backup+rollback/securitynonce preservation. Нетдоказанногоrelayfiltering → кода/flags/serverdeployment нет. C6 GGC_AS20485 communityrange label only, дваIPотложены(решениеowner/нетapp); SteamValve/BF6ports metadata вcatalog сновымpin+LF, BF6 неrunnablecard/exeневыдуман, все10defaultDIRECT иguards/Riot неизменны.
Проверки: полныйGo race/vet/build Linux+Windows PASS; realQUIC CH+authenticatedpool recovery; pinnedDLL70filters/iSCSI/trace подWine PASS; PS7measurement/local/browser/youtube PASS; C#211/211 иWPF0errors/4прежнихCS8604, LF/hash/CRLF negative PASS. NativeWindows/PS5.1/ISP/DoT/реальнаяэкономия НЕ проверены. Отчёт/ограничения по каждомуC — docs/research/NETWORK_MEASUREMENTS_2026-10.md.
Prepush: existingRPC routed/clubreadonly; no manifestreader/default/newdeps; PSquotedflags/BOM/UTF8JSON/trackedonly/nativeexit/hash checks. НичегонаVPS/SSH/CA/кодах/activation/PR8 не запускалось. Область освобождена; следующийшаг — owner measurements, неавтоматическийdeploy/rotation.

## Таблица работ
| Работа / дата | Файлы | Агент / ветка | Статус |
|---|---|---|---|
| Фазы0/1: измерения relay/DNS/LOCAL /25.10.2026 | core/transport; core/localstrategy inspection; services/client-service telemetry; tools/windows diag; catalog informational fields; tests/docs | Codex/codex/work | ОСВОБОЖДЕНО: фаза0 0b505da, фаза1 metadata/design; ISP/native Windows gates открыты |
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
| Дизайн VELIQ / 25.10.2026 | `website/roadmap.html`, блок ROADMAP в `design/tools/build-site.py`, `website/assets/site.css` | Claude Code | ГОТОВО 25.10.2026, область освобождена: страница «План развития» оформлена маршрутом из трёх этапов; тексты приведены к правилам `brand/copy.md` (без расположения серверов, названий игр, «образа», обещания 4K). Код программы не менялся |

## Текущая очередь и открытые вопросы
- R1 замер Claude Code 25.10.2026 17:09–17:21 (PC07, служба=95e3117: peer proof хранится в stateyoutube-cdn-peer.json 24 ч): ролик A (Big Buck Bunny) — промоция, файл записан; Restart-Service; ролик B (LG OLED) — PeerRestored, PeerProvisional=22, googlevideo RELAY 0 МБ, LOCAL 232 МБ; в 17:16 у владельца пропал интернет (воспроизведение встало на 91 с, RELAY/ordinary отказы — из-за этого). Повтор C (Costa Rica, 3 мин, после восстановления сети): 2160p без остановок, googlevideo RELAY 0 МБ, LOCAL 534 МБ, RELAY youtube 20 МБ только www.youtube.com. Критерий R1 «2160p 2 мин: googlevideo RELAY ≈ 0» выполнен при наличии proof <24 ч; на бездисковом ПК state после перезагрузки сбрасывается — первая сессия снова начинает с RELAY до первой промоции. Discord LOCAL не проверен.
- R1 замеры Claude Code 25.10.2026 16:30–16:43 (PC07, служба=13b11a7: peer proof не отзывается BACKGROUND_*/не-CDN отказами), чистые профили Chrome: (1) Big Buck Bunny 4K60 — 6 мин 2160p, 1 dropped; LOCAL 333 МБ / RELAY 121 МБ (до первой промоции, одна подгруппа), Strong+Burst промоции. (2) LG OLED 4K60 + 2 автоперехода — LOCAL 600 МБ / RELAY 6 МБ (только www.youtube.com), PeerProvisional=26, PeerProvisionalFailures=0, PeerRevocations=0; один автоперешедший ролик ABR опускал до 1080p/1440p ~2 мин без остановок. Итого 3 замера: 88% / 73% / 99% LOCAL. Ограничение: peer proof живёт 10 мин от промоции, после — новые подгруппы снова RELAY до новой пробы. Причина PeerProvisional=0 на 7513906: 43 фоновых BACKGROUND_* отказа отзывали proof и ставили Retry/Tried подгруппам.
- R1 замер Claude Code 25.10.2026 16:19–16:26 (PC07, служба=7513906 peer proof, чистый профиль Chrome, тот же Costa Rica 4K60): 2160p 5 мин без остановок (1 dropped), затем автопереход на следующий ролик 1440p. NIC RX +1184 МБ; LOCAL youtube-media catalog-c0aa53e48f222aca +908 МБ (9 conn, 3 fail); RELAY youtube +126 МБ (TRIAL_PENDING sn-u15hn5-c5 85, sn-4g5lznly 30). ~88% LOCAL. selection: Trials=2, StrongPromotions=1, PeerProvisional=0 — выигрыш дала промоция той подгруппы, что несла основной поток; peer proof ни разу не сработал (вероятно, остаток шёл по старым H2 без нового TCP). Нужны ещё замеры, прежде чем считать R1 закрытым.
- R1 замер Claude Code 25.10.2026 16:04–16:09 (PC07, служба=7e8ab4c, отдельный профиль Chrome, Costa Rica 4K60 VP9, 2160p подтверждён getStatsForNerds 3840x2160@60): ~5 мин воспроизведения, одна просадка буфера ~10 с, 68/17162 dropped. NIC RX +1032 МБ; RELAY youtube-media +897 МБ (sn-u14g55-5c LOCAL_HEALTH_NOT_CONFIRMED 494 + TRIAL_PENDING 118, sn-aj5go5-5p TRIAL_PENDING 277); LOCAL youtube-media +28 МБ (~3%). selection: Trials=2 (было 0), DataTrials=1, StrongPromotions=1, BurstPromotions=0; records/cache в browser.native.status пустые. Вывод: неоднозначность снята, лестница работает, но экономии нет — продвижение одной подгруппы не переносится на новые sn-кластеры, основной поток остаётся в TRIAL_PENDING/LOCAL_HEALTH_NOT_CONFIRMED. PR #11 не мержить как закрытие R1.
- R1 правка Claude Code 25.10.2026 (ветка claude/project-thread-6jv5o3, PR в codex/work): ECH-гипотеза по коду не подтвердилась (opaque YouTube тоже идёт в лестницу HintOnly). Причина: быстрый LOCAL отдаёт сегмент <3 с и простаивает, strong-критерий (span≥3с) не срабатывает, проба держит max1 слот до FIN сервера (~120 с). Добавлено: SUSTAINED_CDN_BURST_COMPLETE_PROMOTED — ≥1MiB/16 records после ServerHello, затем ≥7 с тишины без записи клиента, без RST/alert/failure (проверка и на read-timeout); счётчик burst_promotions в диагностике. Отказ после продвижения отзывает LOCAL как раньше. Windows/ISP НЕ ПРОВЕРЕНО. Замер 683ee50: 4K ~730 МБ через RELAY, trials=0 — основной поток opaque (GREASE ECH: googlevideo только как outer SNI) + DNS-набор IP ambiguous → OriginAmbiguous → лестница недоступна. Добавлено ECH_YOUTUBE_MEDIA_OUTER_IN_DNS_SET: outer googlevideo:sn-* из YouTube-only DNS-набора без protected снимает ambiguity (браузерная лестница RELAY→LOCAL-проба).
- R1 замер владельца 25.10.2026 (PC07, служба=066b839, diag .16 rev2, YouTube 2160p): googlevideo RELAY ≈206 МБ на rr5---sn-oj5hn5-5t (source=browser_opaque_evidence, LOCAL-проба не создавалась); единственная CDN-проба sn-oj5hn5-5a: 3,29 МБ LOCAL за ~1 с, затем 120 с тишины, 7 client_retransmits → засчитана SUCCESS_WEAK_CLOSED. Гипотеза: GREASE ECH (0xfe0d) в Chrome → ECH_OUTER_SNI_ONLY → opaque (reassembly.go helloHasECH). Промпт Codex выдан владельцу (Claude Code), там же фикс Windows-теста core/locallearning (POSIX-права, learning_test.go:80).
- M1–M5: собрать реальный .16 measurement revision2 snapshot/ClientHello наGoПК; ts/circular/nonbrowserDNS/новыйtransport менять только послеevidence. Старыеproxy/LOCAL/ordinary поля сохранены.
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

## Последние команды владельцу — клиент/каталог/измерения, безVPS
PowerShell5.1 администратора под прежним ownerSID, кореньcodex/work; Git/Go/.NET8/pinned installed driver. Браузеры не закрывать. Helper закрывает толькоVELIQ, GUI затемоткрыть обычнымярлыком; rebuild GUI нужен из-заcatalog. Активныеlayers<=60s/host дополнительно кстарымdiag.
```powershell
& {
    $ErrorActionPreference='Stop'
    $install=@{git='winget install --id Git.Git -e';go='winget install --id GoLang.Go -e';dotnet='winget install --id Microsoft.DotNet.SDK.8 -e'}
    foreach($tool in @('git','go','dotnet')){if(!(Get-Command $tool -ErrorAction SilentlyContinue)){throw ('Install: '+$install[$tool])}}
    & go version
    if($LASTEXITCODE){throw 'Go unavailable'}
    & dotnet --version
    if($LASTEXITCODE){throw '.NET SDK unavailable'}
    $saved=$ErrorActionPreference
    try {
        $ErrorActionPreference='Continue'
        & git --version
        if($LASTEXITCODE){throw 'Git unavailable'}
        $branch=(& git branch --show-current|Out-String).Trim()
        if($LASTEXITCODE -or $branch -ne 'codex/work'){throw 'Use existing codex/work checkout'}
        $tracked=@(& git status --porcelain --untracked-files=no)
        if($LASTEXITCODE -or $tracked.Count){throw 'Preserve tracked changes before update'}
        & git pull --ff-only origin codex/work
        if($LASTEXITCODE){throw 'git pull failed'}
    } finally {$ErrorActionPreference=$saved}
    & .\tests\windows-tools\layer-measurement-checks.ps1
    & .\tests\windows-tools\browser-path-checks.ps1
    & go test ./core/transport '-run' '^TestRelayClientHelloObservedOnWire$' '-v'
    if($LASTEXITCODE){throw 'Local QUIC ClientHello fixture failed; no VPS was used'}
    & .\tools\windows\build-e2e.ps1 -DriverDirectory (Join-Path $env:ProgramData 'neexitlag\driver')
    & .\tools\windows\build-desktop.ps1
    & .\tools\windows\Close-VeliqGuiForUpdate.ps1
    & .\tools\windows\update-service.ps1
    & .\tools\windows\install-desktop.ps1
    $built=(Get-FileHash '.\bin\windows-e2e\neexitlag-service.exe' -Algorithm SHA256).Hash
    $installed=(Get-FileHash (Join-Path $env:ProgramData 'neexitlag\neexitlag-service.exe') -Algorithm SHA256).Hash
    if($built -ne $installed){throw 'SERVICE_HASH_MISMATCH'}
    Get-Service neexitlag | Select-Object Status
    . .\tools\windows\common.ps1
    Invoke-ServiceIPC 'status' | Select-Object build,connected,optimizing,capture_active
    Read-Host 'Откройте VELIQ обычным ярлыком; YouTube >=2мин с перемоткой, Discord, VK/Яндекс в новых вкладках. Затем Enter. Chrome/Edge не закрывать'
    & .\tools\windows\diag-no-vpn.ps1 -Only browser,youtube,layers -LayerHosts @('www.youtube.com','discord.com','i.ytimg.com')
}
```
Смотреть session_creation пристарте/обычнойработе/восстановлении (неkillрадипробы); layer_audit retained/unknown отдельноот0; DNS encryptedcontrast/TLS/HTTP stage иGGC label. НетдоказательстваISP фильтра→C2–C5 неактивироватьвслепую. Файлdiag-resultне коммитить, секретынеприсылать. NativePS5.1 checks иpinnedDLL gate исполняются вownerblock/build-e2e; здесьтолькоPS7/Wine.
