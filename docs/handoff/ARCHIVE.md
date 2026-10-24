## Последний итог — running CDN success/release,24.10.2026
Наd28 владелец получил196MB реального googlevideo LOCAL(c0aa53), но success ждал Close и держал slot331s. Исправлено: сильный sustained proof засчитывается на живом потоке, слот свободен для другой пробы, established TLS не переносится.
Критерий: fullSH/committed/postSH write, >=1MiB после SH и в>=16 завершённых outer type23 records, RX-span>=3s и freshRX<=3s. Scope сразу LOCAL10min; max1 НОВАЯ неподтверждённая проба/live RELAY sibling/3min cooldown сохранены. Это transport, не playback proof.
Поздний настоящий failure отзывает будущий LOCAL: PostPromotionFailures/FailedAfterReply и SUCCESS_REVOKED_AFTER_FAILURE отдельно от уже засчитанного успеха. Поздний Close не освобождает чужой новый slot.
Малый EOF(0<rx<64KiB) без known failure/RST/retransmits/alert/invalidTLS нейтрален: не bad score/Tried/failed_after_reply и не success. Намерение Google без MITM не установлено; реальные RST/timeout/alert отрицательны.
Diag.12: strong_promotions/post_promotion_failures/small_eof_inconclusive/trial_result; completed app records/bytes/span; отдельные последние LOCAL failure timings/volume, включая midstream stall6s. Такой timeout закрывает socket, browser retry/rebuffer возможен, replay запрещён.
Ordinary akashi: один per-IP DIRECT failure включает общую ladder до проверки второго адреса, defaultRELAY без LOCAL proof. Отдельный policy coupling, не исправлялся; реальная причина TCP-отказа без трассы неизвестна.
P0 typed-nil fixture не имела PID/path/appID, поэтому игнорировала pinnedRELAY и зависела от доступности1.1.1.1. На57396ee с доступным loopback FAIL10/10 Linux+Wine; исправленная identity/sentinel PASS20/20 Linux race+Windows EXE/Wine. Go1.27/native Windows здесь не запускались.
Полные Go race/vet/build Linux+Windows и Windows vet PASS; обе Windows fixture EXE собраны. PinnedDLL70filters/iSCSI30/checksum44plans PASS подWine; новые CDN Windows EXE/Wine и PS7 fixtures PASS; Native Windows/PS5.1/новый ISP promotion/TTL/revoke/multi-subgroup НЕ проверены. VPS/CA/SSH/коды/CP flags/пакетные рецепты/фильтры/игры не менялись.
Тест реального TLS с certificate validation:1.25MiB/3.4s -> promotion безClose, допуск другой группы, второй GET по тому же TLS (opens=1). Transform portable seam — не WinDivert/ISP proof.
Pre-push: существующие IPC dispatch проверены; новых методов/manifest readers нет; PS5.1 syntax/BOM/quoted native flags сохранены, JSON/внешние tools не менялись. Команды безgit switch, с exit/hash checks. DIRECT/LOCAL передRELAY, protected/games без изменений.
Аудит: docs/LOCAL_ENGINE_AUDIT.md; прежний итог перенесён дословно в ARCHIVE. Решение24.10 владельца: сильный running proof заменяет ожидание terminal/3weak successes;3weak closed success остаются альтернативой.


---

## Последний итог — neutral close/admission,24.10.2026
Владелец57396ee: все first-close CLIENT FIN+ACK, без RST/ретрансмитов; CDN1trial/12мин,0success. Это не доказанный отказ LOCAL.
CLIENT_INITIATED_CLOSE с rx>0 теперь neutral, не failure/failed_after_reply/negative score; без sustained не success. TrialInconclusive/ClientInitiatedTrialCloses отдельно; реальные RST/timeout/TLS failures сохраняются.
Терминальный Read при обоих FIN/reset завершает учёт max1 сразу; поздний Close не трогает чужой trial. Живой H2 не обрывается/не переезжает ради новой пробы.
TrialAfter>=3мин сохранён, фоновый Retry больше не gate qualified реального flow. По подгруппе reason/count/time допуска, active slot age/lastRX/state; TRIAL_PENDING не гарантирует, что Chrome создаст новый TCP.
Diag.11: LOCAL/RELAY Hello→ServerHello avg/max, TCP open, transform send/reassembly, Hello→close/lastRX→close, wire RX после SH, outer application record/half-close. HTTP/H2/Chrome intent без MITM не виден.
Midstream watchdog отмечает timeout ДО Close, чтобы его собственный FIN не скрывал failure. Пассивная idle gap не отравляет strategy; реальные timeout/after-reply failures остаются отрицательными.
Полные Go race/vet/build Linux+Windows и Windows vet PASS; обе Windows fixture EXE собраны; pinned DLL70filters/iSCSI30/checksum44plans подWine PASS; PS7 PASS. Native Windows/PS5.1/Chrome intent/ISP CDN success НЕ проверены здесь. Fake recipes/repeats/TTL/guards, VPS/CA/коды/SSH/CP flags/PR8 не менялись.
Аудит/ограничения: docs/LOCAL_ENGINE_AUDIT.md (после57396ee). Предыдущий итог перенесён дословно в ARCHIVE, ничего из истории не удалено.


---

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


## Последний итог — LOCAL native engine,24.10.2026
`5319850`: точная strategy×host диагностика; `4568dd4`: LOCAL owner race guard и CDN trial по scoped transform.
`1cdb650`:64 Flowseal TCP443 rules реально исполняются в fixtures,44 уникальных плана+9legacy;49 HTTP/QUIC/прочих rules явно unsupported.
Raw WinDivertSend уже существовал; исправлены LSO>1500, native TS/MD5/badseq semantics, fake/split/disorder/host/SYN-data и bounded64 outputs/flight. Без external winws/глобальных настроек.
Pure TS требует timestamp negotiation; собственный LOCAL socket включает Winsock option10, отказ виден в diag. Templates свои, побитовая эквивалентность winws НЕ заявлена.
Общие DPI группы: Discord CDN PNG >64KiB/один TLS; публичный gateway WSS nonce ping/pong и .gg redirect — отдельный CONTROL_NOT_BULK proof, не login/page/voice PASS.
Diag.9/browser.local.probes: actual fake/TTL/seq delta/headers, TX/RX/body, status/failure и отдельные silence/RST/freeze/EOF/protocol/NO_TRANSFORM причины; все service groups.
Полные Go race/vet/build Linux+Windows+Windows vet PASS; обе Windows fixture EXE собраны; DLL70filters/iSCSI30/checksums всех catalog plans подWine PASS; PS7 fixtures PASS.
Нативные Windows/PS5.1/новый socket option/ISP2160p/Discord LOCAL НЕ проверены; видео RELAY RX≈0 ещё НЕ доказано. CI не выдавать за PASS без job steps.
VPS/CA/activation/коды/CP flags/PR8 не трогали; браузеры не закрывали. Аудит/ограничения: docs/LOCAL_ENGINE_AUDIT.md; предыдущие итоги дословно в ARCHIVE.


## LOCAL диагностика — 24.10.2026 (работа продолжается)
Владелец f0e181a:2160p RELAY работает, LOCAL0/CDN trials0; ordinary тоже имеет отказы. Исправления общие для DPI групп, IP/GEO relay и VK/Яндекс DIRECT сохраняются.
Пункт1: browser.local.probes (отдельный read-only IPC из-за64KiB лимита), bounded128 recent/<=48KiB выдача, strategy×host/group/result/TX/RX/body/commit/SYN/handshake/TTL/actual packet headers. Payload не сохраняется.
Diag.7 — таблица всех групп; новые PS7/PS5.1 CI gates. Это диагностика, не LOCAL PASS; PS5.1 и ISP здесь не выполнялись.
Пункты2–5 в работе: raw WinDivert send уже есть; проверить owner race LOCAL delegation, actual catalog semantics и допуск CDN по реальному transform. CP/shared flags не включать.
Пункт1 проверки: полный Go race/vet/build Linux+Windows PASS; actual pinned DLL filters70/iSCSI30/catalog checksum13 подWine PASS, не kernel. PS7 probes/legacy allowlist PASS; native PS5.1 НЕ проверен. IPC inventory/общая нормализация не менялись; новый метод включён в dispatch и club read-only.
Пункт2/5: LOCAL reserved SYN теперь требует owner proof как native; прежний DIRECT self bypass при owner lookup race запрещён. CDN trial допущен по committed transform того же scoped плана, критерий3успеха/>64KiB и live relay sibling сохранён. Пункт3 (все TCP443 семантики) ещё не реализован; аудит docs/LOCAL_ENGINE_AUDIT.md.
Пункты2/5 gates: полный Go race/vet/build Linux+Windows+Windows vet PASS; DLL70filters/iSCSI30/checksums13 подWine PASS, PS7 PASS. Старые delegation-close тесты сохранены. На ISP root cause/LOCAL2160p и native PS5.1 НЕ доказаны.


## Сжатие handoff — 24.10.2026
Было3163строки/392819байт (2813 в задании — прежний размер); история закрытых этапов и старые итоги перенесены дословно в ARCHIVE.md.
Новые блоки сверху; предыдущий технический итог тоже архивирован; действующие ограничения/открытые вопросы перенесены в краткий статус.
Только документы, тестовые прогоны не запускались. Правила завершения задачи и pre-push checklist добавлены; код/скрипты не менялись.
Итоговый размер CURRENT: 105строк (≤150). Команд владельцу для этой docs-задачи нет.


## Последний итог — документы, 24.10.2026
Архив `46c0b97`: CURRENT3163строки/392819байт →118строк;185исторических блоков перенесены дословно, новые сверху.
Эта задача: правила короткого handoff, обязательный pre-push checklist и соразмерные проверки; код/CLAUDE.md не менялись.
Предыдущий технический итог теперь тоже в ARCHIVE. Тестовые прогоны не запускались: только документы, проверены diff/сохранность истории/лимит строк.
Последний код `6f0312d`, учёт `aac01db`; Hostkey relay/1080p принят владельцем, LOCAL video НЕ подтверждён.
Подробные текущие ограничения/очередь/доказательства ниже; на VPS ничего не выполнялось.


# ARCHIVE — дословная история handoff

Последний итог Codex (24.10.2026 MSK): очередь экономии VPS, код опубликован по этапам; область освобождена.
Владелец подтвердил Hostkey relay/1080p; LOCAL video пока НЕ подтверждён.
Пункты1a–1c и каталог2 сохранены коммитами b213fa5/c84f7c1/3799f75,
de8e5d7/08a9eb5/a3ec712/77566d6. Обычный merge20d8bd8 разрешён владельцем
и выполнен9135519: только9строк CLAUDE.md, конфликтов нет, без rebase/force.
Пункт3 подключает catalog plans к bounded selection и guarded CDN trials;
Пункт4 signed advisory coordination подготовлен и изолированно проверен;
production CP ещё без endpoint, общего picker на40ПК там пока нет.
Пункт5: общий HTTPS origin policy/selector проверен; HTTP80/Discord UDP LOCAL
не реализованы, прежний voice не изменялся. Пункт6 aac01db: учёт проверен; пункт7 (этот коммит): primary/standby hysteresis и live diagnostics.
Решение CDN принято: обе фоновые проверки www.youtube.com/i.ytimg.com >64KiB,
затем max1 LOCAL trial при реально получающих данные RELAY googlevideo siblings,
cooldown несколько минут/подгруппу;3успешных >64KiB потоков перед LOCAL,
сбой→RELAY новых соединений, экспоненциальная пауза, post-reply failures в diag.
Это ограничение риска, не гарантия незаметного повтора сегмента плеером.
Production/CA/activation/credentials/PR8 агентом не запускались. Windows/PS5.1/
WinDivertOpen/ISP/LOCAL1080p не проверены; проверки merge записаны ниже.

Исторические блоки: новые сверху; устаревшие инструкции не являются текущим разрешением.
Искать точечно через rg; актуальные ограничения — CURRENT.md и AGENT_RULES.md.

# Также Discord, Instagram, X, Telegram Web, ChatGPT/Claude, rutracker, VK/Яндекс.
& .\tools\windows\diag-no-vpn.ps1 -Only browser,youtube
```

Нужны enabled optimization/READY и реальные новые browser flows, не только
профили/настроенный адрес. В diag.6: service-separated LOCAL/RELAY counters и
стратегия, trial costs/streak, причина relay/cache, primary physical endpoint и
байты/health/probe timestamp, client session/day. VK/Яндекс DIRECT, IP/GEO группы
Telegram/ChatGPT/Claude/Instagram RELAY_REQUIRED; DPI группы LOCAL только после
положительной проверки. Discord voice LOCAL не считать принятым по TCP.
Для холодного googlevideo guard требует3законченных успешных проб с интервалом
>=3мин: двух минут недостаточно для первоначального подтверждения. Подбор идёт
автоматически, пока основной video получает RELAY; visible/post-reply failures
считать явно. После подтверждения проверить новые потоки1080p2мин+seek: LOCAL RX
должен стать основным, новые video RELAY RX около0; старые pinned relay потоки,
DNS DoH/CP/tickets в это обещание не входят. При неудаче — RELAY и следующий
bounded trial, не объявлять LOCAL PASS. Затем выключить оптимизацию и проверить
обычный сайт без перезапуска браузеров. Диагностику/коды/секреты не коммитить.
# В уже открытом Chrome/Edge: YouTube1080p2мин и перемотка после подтверждения LOCAL.
### ТОЧНЫЕ команды владельцу (ПК; без закрытия Chrome/Edge и без VPS deployment)

PowerShell администратора под прежним SID, из корня клона codex/work (путь любой).
Закрывается только VELIQ GUI, чтобы освободить прежние recovery helpers. GUI код
не менялся в6–7, build-desktop/install-desktop не нужны. Служба обновляется с
сохранением identity/settings/owner; на время её перезапуска managed sockets могут
переподключиться — не обещать непрерывность существующего TCP при service restart.
Обычный GUI затем открыть прежним ярлыком; на UAC-off режиме права как раньше.

```powershell
git switch codex/work
git pull --ff-only origin codex/work
git rev-parse HEAD
$m=Get-Content -Raw "$env:ProgramData\neexitlag\service.json" | ConvertFrom-Json
& .\tools\windows\build-e2e.ps1 -DriverDirectory $m.driver.directory
& .\tools\windows\Close-VeliqGuiForUpdate.ps1
& .\tools\windows\update-service.ps1
. .\tools\windows\common.ps1
Invoke-ServiceIPC 'status' | Select-Object build,connected,optimizing,capture_active
Invoke-ServiceIPC 'local.catalog' | ConvertTo-Json -Depth 7
& .\tools\windows\relay-status.ps1
& .\tools\windows\relay-traffic.ps1
### Решения24.10.2026 / границы готовности
- Экономия: только sustained verified LOCAL; неподтверждённые DPI группы временно
  RELAY и продолжают bounded подбор. Игры/DNS53/TLS/mTLS/lease/replay не менялись.
- Клуб: shared hints advisory, peer подтверждает локально; без нового CP endpoint
  явно per-PC fallback, общего picker на текущем production обещать нельзя.
- Учёт: client payload session/day доступен после client update; server NIC month
  и forecast подготовлены с explicit gaps. Flags default-off; production monthly
  пока UNKNOWN. Его включение/shared endpoint — отдельное разрешение на CP/relay
  и приватные state dirs. Повторный deployment для LOCAL клиента НЕ нужен.
- Failover: только новые authenticated opens, минимум30с hold/30с sustained return,
  fresh echo proof, no load-balancing. Старые TCP/UDP pinned до close/death;
  потерянный established TCP не восстанавливается прозрачно на другом egress.
- Оставшиеся ограничения: собственный TCP443 каталог не является полной копией
  Flowseal; неподдержанные HTTP80/QUIC/fakedsplit варианты явно отмечены данными.
  QUIC→TCP прежний; Discord WebSocket участвует в HTTPS selector, voice UDP LOCAL
  не реализован, прежний adaptive voice не менялся. Short response не >64KiB proof.
  LOCAL1080p/ISP/40ПК/PS5.1/native WinDivertOpen здесь не подтверждены.
- Hostkey P0 принят владельцем ранее; не требовать add-relay/CA rotation/codes заново.
  Старый2.56.177.4 доступен как standby до12.11 по данным владельца, не гарантия
  дальнейшей доступности. Серверы/сертификаты/credentials/activation/PR8 не трогались.

### 7 — minimal primary/standby (24.10.2026 MSK; в проверке)
Runtime hold30s + continuous fresh authenticated primary recovery30s, без
load-balancing, без миграции/replay existing TCP/UDP. Signed config refresh
не сбрасывает hold; single session/capacity/origin error не отравляет здоровый
pool. Новые OPEN bounded1.5s→other auth pool; оба fail→fail-closed, не DIRECT.
Per-relay address/live endpoint/probe RTT+timestamp/health/active/bytes/failure
class/switch reason. relay-status.ps1 теперь только live IPC, не требует старого
Go cache reader или повторного add-relay/SSH. Video12/other4 и8×48/probe reserve
сохранены. Детали/ограничения — docs/RELAY_PRIMARY_FAILOVER.md.
Первые focused race10x и PS7/diag allowlist PASS; полный проход ещё идёт.

Пункт7 final: основной Hostkey/standby через signed registry, replacement standby
с собственным ID без installer rebuild. Aggregate load=1 не снимает provider;
реальное здоровье своего pool/echo отдельно. Повторение одного good snapshot не
доказательство восстановления: нужны более новые positive probe observations.
Focused race10x с failure/capacity/origin/config/load/UDP/old TCP cases PASS.
Финальные go test -race -p1 ./..., vet, build Linux/Windows и Windows vet PASS.
Новый Windows platform test EXE собран; actual pinned DLL filters70/iSCSI30/
catalog checksums13 PASS подWine. PS7 status/traffic/catalog/selection/legacy/
StrictMode fixtures PASS;100Python deployment PASS. Не native driver/PS5.1/ISP.
C# исходники не менялись: после разрешённого merge ранее198/198 +WPF0errors,
4 прежних CS8604 PASS; в6–7 заново WPF не собирали.
Isolated PG17 полного CP6 PASS12.462s; контейнер удалён, CP7 не менялся.
GitHub Verify для aac01db:40070545974, core/windows failure со steps=[]:
https://github.com/buzzmean/neexitlag/actions/runs/40070545974
CI тесты не исполнялись; новый billing текст не получен, artifact quota не
назначена причиной. Ни Go gates, ни Wine не заменяют native Windows acceptance.

### 6 — client payload и signed server NIC budget (24.10.2026 MSK; в проверке)
Новый read-only traffic.status/relay-traffic.ps1/diag.5: session/UTC-day по profile,
service/subgroup/reason, logical relay alias. Engine observer после release lock,
DIRECT/LOCAL не считаются.512keys/day/session,2дня, overflow/omitted явные, файл
atomic1MiB, flush вне packet lock, device epoch не в IPC; image counters другого
устройства игнорируются. Root network/identity/профили не меняются.
Server TCP metrics теперь live per Write (раньше только завершение io.Copy).
Optional Linux NIC sampler и отдельный CP signed relay-traffic endpoint/TTL:
relay-specific прежний heartbeat token bound кID, без health/config/schema writes.
CP интегрирует raw counter deltas сам, duplicate report не удваивает; scopes32relay/
2month до64rows. Hostkey3TB, old limitunknown; partial/reset/stale/boundary/gaps и
неизвестная billing reset date явные, предупреждение80%/forecast, не route block.
CP/relay flags defaultdisabled, units/production не менялись. Для server monthly
нужны отдельное разрешённое обновление и собственные service state directories;
клиентский day counter доступен без этого, monthly UNKNOWN не выдаётся за0used.
Raw NIC включает overhead/SSH/другие процессы, payload totals не invoice.
Документация/только read-only owner commands — docs/RELAY_TRAFFIC_BUDGET.md.
Проверки/commit — после полного прохода. Пункт7 ещё не выполнялся.

Пункт6 final gates: полный go test -race -p1 ./... PASS; Go vet и build Linux/
Windows + Windows vet PASS; isolated PostgreSQL17 полный control-plane race
12.462s PASS (после pg_isready, без production);100Python deployment PASS; PS7
traffic/legacy/StrictMode fixtures PASS. Actual pinned DLL filters70/iSCSI30/
catalog checksum13 повтор PASS подWine (EXE пункта5, capture source не менялся).
C# не менялся. Нет native Windows/PS5.1/ISP/40ПК/monthly production proof.
Учёт не пишет schema/config/identity и не меняет маршруты; optional server meter
не включён. Переходим к7 без ожидания.

### 5 — общий HTTPS origin policy/selector (24.10.2026 MSK; в проверке)
Единый versioned registry/source/date для DPI/RELAY групп в core/originpolicy;
старые BrowserRelayGroup/KnownBrowserRelayGroup/protectedBrowserDomain читают
данные, required security floors не ослабляются. Registry optional в signed
catalog, default data release2 вместо1 (schema1); публикация next>=3, regression
old1→new2 без equivocation. Обычные VK/Google/Яндекс DIRECT; известные DPI origins
и индивидуально отказавшие обычные HTTPS origins используют ТОТ ЖЕ worker/cache/
engine, без дублирования. Динамические signed GEO enforced в planner/open/native
hints/DNS; required9th hint проверяется до truncation. Club learning передаёт
только reviewed DPI scopes, не неизвестные browsing names; peer локально проверяет.
Статистика service-separated: YouTube не смешивается с Discord/X/rutracker.
Это TCP443/IPv4 этап: HTTP80/browser QUIC fake unsupported; Discord adaptive UDP
voice НЕ изменялся и LOCAL для него НЕ доказан/не реализован. Без подходящего
public body >64KiB LOCAL не объявляется. Подробности docs/BROWSER_ORIGIN_POLICY.md.
Проверки и commit фиксируются после полного прохода; VPS/CA/activation не трогались.
Пункт5 проверки: полный Go race PASS; unit data floor/nested override/suffix /
ordinary defaults /dynamic signed routing /required9th hint /scope reserved IP
PASS; old signed catalog release1→release2 regression PASS. Полный isolatedPG17
CP race11.281s PASS после явного pg_isready; первый запуск был ошибочно начат до
готовности fixture и дал connection refused, код/тесты ради этого не менялись.
Container удалён. Go vet/Linux/Windows build+vet PASS;100Python deployment иPS7
fixtures PASS. Последнее reserved-IP изменение проверено subset и полным race;
сборки/проверку pinned DLL повторить до публикации. Native Windows/PS5.1/ISP и
Discord voice LOCAL не проверены. Пункты6–7 пока НЕ выполнены.
Пункт5 publication gates: final full Go race PASS; повтор Linux/Windows build+vet
PASS; pinned actual DLL test EXE пересобран с текущим кодом, filters70/iSCSI30/
catalog checksum13 PASS подWine. PS7+100Python PASS. Все result markers/logs вне
repo; нативный WinDivertOpen/PS5.1/ISP/player не запускались. C# не менялся.
TCP-часть5 готова к проверке, HTTP80/QUIC fake/Discord UDP LOCAL остаются явными
ограничениями (не выключать прежнюю adaptive voice логику). Пункты6–7 следующие.

### 4 — signed club coordination (24.10.2026 MSK; подготовка)
Добавлен отдельный advisory local-learning domain/API: device proof bound к hash
command, nonce/revocation через существующий ConsumeProof, signed scoped TTL90s.
CP HMAC cohort от фактического peer /24 или /48 (не XFF), никаких hardware/user
метаданных. Один leased picker/cohort+subgroup+hint, candidate10min; peers только
локально подтверждают известный ID, не включают LOCAL по одному signed сообщению.
Строгие prerequisites/trials3 сохранены. Shared failure не ослабляет protected.
Файл CP0600 atomic<=256KiB/512entries, явный -local-learning-file, без SQL migrations.
Не настроено/legacy404/outage → диагностируемый per-PC fallback, не обещать общий
picker в текущем production. Клиентский polling<=1req/2s/ПК, picker30s/peer2m/
confirmed5m, max16scopes/проход; не блокирует foreground. Не было VPS deployment.
Core40-member/persistence/re-election/proof scope/TTL/failure tests PASS; isolated
PG17 полный CP race10.752s PASS (включая nonce replay/revocation/schema3 invariant),
service shared/real TLS subset3x PASS. Full final checks фиксируются перед push.
Реальный40ПК/freeze/Windows/ISP/shared CP не проверен. Пункты5–7 ещё не готовы.
Подробный протокол/ограничения — docs/SHARED_LOCAL_LEARNING.md.
Пункт4 final: полный Go race повтор PASS; Go vet/Linux build/Windows build+vet
PASS;100Python deployment PASS; PS7 shared/legacy/StrictMode fixtures PASS; pinned
DLL70filters/iSCSI30/catalog checksum13 PASS подWine (packet/filter source не
менялся). Один предыдущий full race имел demo25s route timeout при параллельной
компиляции, последний full PASS; причина flake НЕ доказана, тест не менялся.
Дополнительные regressions сохраняют3CDN successes при warm flow и запрещают
продлевать CDN media deadline одними public prerequisites. Изолированный PG
container после проверки удалён; production DB/activation/CA не запускались.
CI пункта3 c7ef8bd Verify40045013456: core/windows failure со steps=[]; CI не
исполнился. https://github.com/buzzmean/neexitlag/actions/runs/40045013456
Новые annotations не получены, artifact quota не назначена причиной.

### 3 — subgroup selection и согласованные CDN trials (24.10.2026 MSK)
Пункт2 опубликован43c7917 после разрешённого merge9135519. Пункт3: отдельные
account/page/image host и sn-CDN ключи, network/profile/port/hint-only isolation.
13own catalog +9legacy plans включены в background selection; статистика этой
сети/подгруппы ранжирует prior. Cold foreground RELAY сразу, не ждёт таймаут.
Worker max1/ПК, max12s/candidate, timer15s продолжает recheck без новых flows,
backoff30s..5m, полный sweep повторяется, owner/policy/network/guard rechecked.
Кеш256, pending64, engines256, scores2048, qualifications256, stats128+overflow.
Transfer GET separate API: TLS/cert/HTTP2/1.1, body<=128KiB+1, >64KiB actual body,
short403/404/204/16KiB-freeze не успех. Для CDN обязателен same-strategy проход
ОБОИХ www.youtube.com и i.ytimg.com (отдельные source tuple/evidence). Затем max1
live LOCAL trial при data-receiving RELAY video sibling того же профиля,
min3m/subgroup;3consecutive >64KiB+ServerHello+posthello write+transform/no failure
до Close→LOCAL. Поломка сразу RELAY для новых flows, after-reply failures считаются.
Смешанные YT hints разных подгрупп не разрешают LOCAL; uniform test до truncation.
TLS после origin byte не replay/migrate; video12/other4, games/anti-cheat/iSCSI и
WinDivert filters не меняются. Три trial требуют минимум6мин плюс длительность
их потоков, нельзя обещать cold2min LOCAL PASS или отсутствие цены проб.
Проверки пункта3: полный повтор Go race PASS (включая unchanged replay/lease/game
regressions), Go vet/Linux build/Windows build+vet PASS. Три controlled real
TLS+HTTP128KiB CDN trial→warm LOCAL PASS; short/16KiB stall/cert/204 rejected,
max1/cooldown/streak/mixed ninth hint/worker shutdown PASS; повтор subset3x PASS.
PS7 diag trial-cost/StrictMode и legacy browser fixtures PASS. Pinned DLL helper
filters70/iSCSI30/catalog checksum13 повторяется без packet/filter source changes.
Native Windows/PS5.1/WinDivertOpen/owner ISP/actual new strategy LOCAL video НЕ
проверены. Пункты4–7 ещё НЕ выполнены, текущее обучение локально этому ПК.

### Решение владельца и синхронизация (24.10.2026 MSK)
Обычный merge20d8bd8 выполнен9135519 без конфликтов (только CLAUDE.md).
Владелец разрешил guarded CDN trials: предварительно та же стратегия на
www.youtube.com и i.ytimg.com с verified TLS/HTTP и >64KiB, max1 trial глобально,
только при живых RELAY media siblings, раз в несколько минут на подгруппу,
3consecutive sustained flows перед LOCAL; деградация → RELAY новых потоков,
next strategy/backoff. Считать сорвавшиеся после ответа trials отдельно.
Существующие TCP pinned, TLS/owner/anti-cheat guards не ослаблять.
Исторические записи о блокере выше отражают прежнее состояние, блокер снят.
После merge повторены: полный Go race PASS; C#198/198 PASS; WPF cross-build
0errors/4 прежних CS8604 warnings;100Python deployment PASS; PS7 owner preflight,
каталог и browser-path fixtures PASS; pinned DLL70filters/iSCSI30/catalog13checksum
PASS подWine. Linux/Windows сборки и vet после merge PASS. Native Windows,
PS5.1 и ISP этим не подтверждены. Код Chrome/Edge GUI не менялся.

### 1c supplemental — PS5.1 native argv (24.10.2026 MSK)
Python -c preflight больше не содержит embedded double quotes: Python string
literals используют single quotes внутри PS double-quoted аргумента. Это исключает
известное удаление внутренних double quotes legacy PS5.1 native marshaller.
Fixture проверяет фактический -c argv и отсутствие опасных quotes; PS7 real Python
+Store alias/tracked-only checks PASS. Native PS5.1 всё ещё не исполнялся.
Дополнение пока локальное до разрешённого merge;3799f75 на origin содержит
предыдущую quote-форму. Повторный deployment VPS не нужен.

Дополнительная проверка публикации: read-only git merge-tree для local HEAD и
remote20d8bd8 показывает чистое объединение, но actual merge НЕ выполнялся.
CI уже pushed3799f75: https://github.com/buzzmean/neexitlag/actions/runs/40022209321
core/windows failure со steps=[]; не исполнилось, не CI PASS. Для локального
каталога CI ещё нет, потому что push заблокирован. Последний supplemental PS7
preflight/argv/каталог/browser fixtures и Go IPC inventory PASS; C# не менялся.
Ignore observe-result ограничен форматами отчётов, не скрывает будущие .ps1/.go.

# Только после публикации пункта2; до неё метод local.catalog ещё отсутствует:
Invoke-ServiceIPC 'local.catalog' | ConvertTo-Json -Depth 7
& .\tools\windows\diag-no-vpn.ps1 -Only browser,youtube
```

При будущей приёмке3–7: уже открытые Chrome/Edge, YouTube1080p2мин/перемотка,
Discord/Instagram/X/TelegramWeb/rutracker/VK/Яндекс. Требовать реальные LOCAL/RELAY
bytes и >64KiB health, не только страницы/ACK; monthly VPS budget пока не реализован,
команды выдумывать нельзя. Интернет/новую стабильность без owner ISP не объявлять.

### Передача на блокере публикации (24.10.2026 MSK)
Следующий шаг — только после ответа о merge: fetch/review, обычный merge конкретного
20d8bd8 (только CLAUDE.md), повтор обязательных checks при merge, push codex/work.
Не force/rebase, не трогать ветку Claude или PR8. Новое чужое изменение перед
push сначала проверить; конфликт не решать без владельца. Все source changes
пункта2 сохранены коммитами, незакоммиченных секретов/diag нет.

После публикации каталог доступен для инспекции, но 3–7 ещё предстоит реализовать.
Первая задача3 — новый subgroup key отдельно от DirectTimeoutHost, transfer probe
>64KiB вместо32769B, непрекращающийся bounded backoff, ranking и подключение13plans.
Текущая старая foreground/provisional логика ещё НЕ изменена. Не считать этот
checkpoint приёмкой LOCAL и не просить владельца заново deploy VPS/коды/CA.
Общий network learning4, общий browser ladder/data blocklist5, daily/monthly
budget6 и hysteresis7 пока отсутствуют. relay.primary.status даёт session payload,
НЕ месячный расход Hostkey. Полный приёмочный блок для3–7 пока неприменим.

**Команды владельцу, только обновление/проверка клиента, без VPS**. Корень repo,
PowerShell администратора под SID владельца, Chrome/Edge не закрывать. До merge
это ставит только уже опубликованные1a–1c (remote20d8bd8), не LOCAL каталог2. После
merge проверять фактический HEAD. GUI VELIQ закрывается штатным helper, браузеры нет.
C# не менялся: повторная GUI-сборка не нужна, установка/activation не повторяются.

```powershell
git switch codex/work
git pull --ff-only origin codex/work
git rev-parse HEAD
$m=Get-Content -Raw "$env:ProgramData\neexitlag\service.json" | ConvertFrom-Json
& .\tools\windows\build-e2e.ps1 -DriverDirectory $m.driver.directory
& .\tools\windows\Close-VeliqGuiForUpdate.ps1
& .\tools\windows\update-service.ps1
. .\tools\windows\common.ps1
Invoke-ServiceIPC 'relay.primary.status' | ConvertTo-Json -Depth 8
### 2 — signed data delivery / совместимость (24.10.2026 MSK)
core/localcatalog: отдельная подпись local-catalog, версия/TTL48h limit, persistent
anti-rollback/equivocation cache. CP GET public signed data / admin PUT atomic data
file (-local-catalog-file); без файла PUT503. Config wire не расширялся, лимиты
64KiB catalog и1MiB revocations не менялись. Ни SQL/migrations/enroll/revoke, ни CA
не затронуты. Старый production CP ещё404; клиент использует bundled data,
poll раз в час не блокирует routing/GUI. Около49KiB/check/~47MiB/day40PC; это
metadata, фактическая экономия video пока0/не доказана. local.catalog +diag.1
показывают available plans, не LOCAL PASS. Club read-only эти/relay diagnostics
разрешены; settings/optimize и прочие mutations по-прежнему закрыты для guest IPC.
Signed/tamper/expiry/domain/rollback/disk/atomic admin tests, oldCP404 клиентский
путь, dispatcher235literals/57methods, PS7 fixed-fields fixtures PASS. Полный
isolated PostgreSQL17 CP race PASS10.806s, service full race PASS13.617s после
seqovl. После seqovl полный повтор Go race/vet/Linux+Windows builds/Windows vet PASS. Один
предшествующий полный прогон упал в demo25s route timeout; изолированные current
и baseline27f2a8f оба3/3 PASS. Причину единичного timeout не доказали, тест не менялся.
Пункты3–7 ещё не реализованы. Пока не разрешён merge remote20d8bd8, новая работа
сохранена локально, push de8e5d7/08a9eb5 отклонён без force. Нужен ответ владельца
на merge и на CDN real-flow vs strict-background verification; без ответа
для переработки пункта3 сохраняется требование >64KiB; старый foreground/provisional
код ещё не переписан. Не объявляем этот этап исправлением LOCAL video выбора.

Точный аудит к пункту3: local cache key cacheHost→DirectTimeoutHost сводит accounts,
www и googlevideo к одному youtube.com. Старый transfer probe body32769B не может
проверить >64KiB; short403/provisional не является success. После полного sweep
Retry=Until прекращает выбор до TTL, затем порядок сбрасывается. Эти три места
нужно переработать вместе с отдельными subgroup keys; пока НЕ исправлены.

### 2 — seqovl расширение (24.10.2026 MSK)
Добавлен browser-only bounded split-seqovl256–1024: собственный полный TLS prefix
перед первым настоящим сегментом, только в старом sequence range. Receiver model
проверяет trim ровно объявленного prefix и сохранение всех application bytes;
старые standalone-fake security tests не ослаблены. MTU<=1500/cutoff1/budget20,
legacy ceiling5 прежний. Теперь119rules →38compatible/13unique own plans,81explicit
unsupported. Core race PASS, повтор actual pinned DLL70filters/iSCSI30/checksums
PASS подWine. Это не Windows/ISP proof. Push пока заблокирован remote20d8bd8;
запрошено разрешение обычного merge, не rebase/force.

### 2 — каталог/packet foundation (23.10.2026; ещё не полный пункт2)
Flowseal1.10.4/176f56dc:22 general*.bat,119rules,4MIT sources. zapretgui MIT
найдена docs/LICENSE, notices добавлены. Offline data importer не исполняет scripts.
6own TCP plans/17rules компилируются;102 (в т.ч.QUIC/HTTP/SYNdata/fakedsplit/seqovl)
явно unsupported, не выдаются за рабочие. Own TLS1.3 fake1250B, repeats<=16, mixed
markers/multisplit/disorder, oldseq + TS/MD5, autottl по exact SYN/ACK. Ограниченный
browser-catalog ceiling20/input/cutoff1; legacyDIRECT+ ceiling5 и game guards
не менялись. Пока каталог не включён в AUTO, это foundation, не ISP PASS.
Go core/platform race PASS; pinned DLL filters70/iSCSI30/checksums PASS underWine;
2Python inventory tests PASS; native kernel/PS5.1/ISP не проверены.
Пункт3 имеет вопрос: CDN root403/404 <64KB, Native не видит signed video URL.
Нужен ответ владельца: реальный bounded LOCAL trial с риском post-hello error либо
строгий RELAY до подходящего фонового URL. Пока зависимое поведение не меняется.
Signed CP update/shared stats/подгруппы/budget/failover ещё НЕ выполнены.

### 1c — tracked checkout / Python / private output (23.10.2026)
add-relay проверяет только tracked changes, untracked owner reports не мешают.
diag default → %LOCALAPPDATA%\VELIQ\Diagnostics\diag-result.txt (TEMP fallback),
legacy diag-result/cs-diag/observe-result patterns игнорируются Git. Python preflight
проверяет реальный3.9+, пропускает WindowsApps Store alias без запуска, поддерживает
py -3 и даёт winget install Python.Python.3.12 при отсутствии. Никаких empty native
arguments; prefix argv сохраняется. PS7 fixture (real temp Git repo, untracked ok,
tracked rejected, Store alias rejected, real Python) PASS; parser/browser StrictMode
PASS, workflow YAML parsed. PS5.1 fixture добавлен в CI, локально native5.1 нет.
VPS/owner PC не запускались; повторный deployment не требуется.

### 1b — общий runtime manifest (23.10.2026)
Manifest/strict Load/Normalize вынесены в internal/clientconfig; service alias
и relay-client используют один код. ConfigureActivation/Club/ConnectRemote также
нормализуют программно переданный manifest. Legacy missing probe_token_file →
sibling probe.token; missing probe_target → host существующего CP:40000. Явные
settings не меняются, файлы/identity не создаются, service.json не переписывается.
Race tests defaults/idempotence/no-write/explicit-settings + tool no-enroll +
activation/club/dispatcher subset PASS. Workaround service.relayproof.json больше
не нужен; actual owner probe ещё не выполнен. VPS не запускался.

### 1a — IPC primary dispatch (23.10.2026)
Воспроизведён unsupported relay.primary.status: subhandler был, main switch нет.
Добавлен route и read-only regression. Static AST inventory проверяет56 dispatched
методов, subhandlers и234 literal calls из Windows PS/desktop; early club.status
route распознаётся, guard-only cases не считаются dispatch. Race subset PASS.
Hostkey повторно деплоить не нужно, достаточно client update. VPS не запускался.

### Приёмка владельца и новая очередь LOCAL/catalog (23.10.2026)
Владелец подтвердил27f2a8f: Hostkey VPS-PRIMARY работает, 221opens/~60ms,1080p
идеально через relay. LOCAL fake5x2-disorder к accounts.youtube.com ORIGIN_SILENT;
1 отказ ошибочно повлиял на youtube.com scope, relay22conn/~181MB за2мин.
Это НЕ LOCAL PASS. Владельцем подтверждён локальный обход winws на этой сети.
Временный owner workaround service.relayproof.json заполнял probe defaults;
никакие identity/токены из него не копировать в handoff/git.
Новая очередь1a IPC dispatch,1b общий manifest defaults,1c checkout/Python/output,
2 каталог/техники с provenance/license,3 bounded subgroup selection,4 signed club
learning,5 general browser ladder/data blocklist,6budget,7primary/standby.
Главный экономический принцип добавлен по прямому указанию владельца.

# Отключить оптимизацию в GUI: обычный сайт должен работать; браузер не закрывать.
```

Ожидаемый факт приёмки: youtube-media stage LOCAL с именем стратегии и реальным
committed transform/transport confirmation, стабильное видео и почти0 НОВОГО
YouTube TLS payload RELAY. Старые pinned relay connections, CP/tickets/DNS DoH/echo
могут продолжать relay traffic. При неработающем LOCAL покажутся failure reason и
RELAY: это корректная деградация, не локальный PASS. P0 actual endpoint/ticket ID
смотреть в relay-status connections, configured_primary — только intent.
Нативные Windows/PS5.1/WinDivertOpen/OpenSSH/systemd/Hostkey/reboot/freeze/ISP/player
ещё не проверены; no dialogs/restart добавлением этого кода не требуются.

# Проверить VK/google.com DIRECT, OpenAI RELAY_REQUIRED. Затем диагностика:
& .\tools\windows\diag-no-vpn.ps1 -Only browser,youtube
# Уже открытый Chrome/Edge: новая вкладка YouTube, 1080p2мин, перемотать2–3раза.
# Альтернатива P0: обновить ТОЛЬКО клиент, без SSH/VPS (браузеры не закрываются).
$m=Get-Content -Raw "$env:ProgramData\neexitlag\service.json" | ConvertFrom-Json
& .\tools\windows\build-e2e.ps1 -DriverDirectory $m.driver.directory
& .\tools\windows\Close-VeliqGuiForUpdate.ps1
& .\tools\windows\update-service.ps1
# P0: сначала только read-only, затем одна команда production-переноса владельцем.
& .\tools\windows\add-relay-vps.ps1 -Host 146.0.77.98 -Name hostkey-ams -PromotePrimary -Inspect
& .\tools\windows\add-relay-vps.ps1 -Host 146.0.77.98 -Name hostkey-ams -PromotePrimary
& .\tools\windows\relay-status.ps1
### Итог P0 + YouTube LOCAL-first / решение и передача (23.10.2026)
Безопасный путь P0 — отдельная relay audience и update существующей службы, а не
подмена адреса двух replay caches за одним ID. Прямого fallback из RELAY_REQUIRED
нет. Никакого скрытого автообновления остальных ПК, балансировки40ПК/Timeweb,
per-relay video pools или throughput tool в P0 не заявлено; A–E остаются очередью.
Новый leaf90d ещё без отдельного auto-renew; старый CP/activation/echo зависим
от2.56.177.4 до12.11.2026. Hostkey3TB≈9.3Mbps среднего непрерывного трафика.
YouTube LOCAL — собственные fake(old seq+TTL)+multisplit/disorder, два fake/ceiling5
в отличие от полного winws general. QUIC→TCP разрешённый fallback сохранён.
Перенос после ServerHello/16–20KiB freeze без MITM невозможен: старый сокет не
переигрывается; последующие потоки учитывают failure/hint. 403/404 CDN root только
provisional1min, не success; фактические transform+ServerHello+64KiB+250ms даютTTL10min.
Portable verified TLS+GET128KiB fixture×3: cold~0.72–0.75s, warm~0.32–0.34s, relay0;
transform seam mocked, реальную работу desync у провайдера это НЕ доказывает.

Проверки финального исходного `3a2f39c`: полный `go test -race -p 1 ./...`,
Linux/Windows vet/build PASS; focused YouTube race×3 PASS. До последней service-only
поправки также полный isolated PostgreSQL17 control-plane race PASS13.032s,
98deployment Python tests PASS3.617s, actual pinned DLL filter compiler70variants
и private iSCSI evaluator30cases + repeated-fake checksum PASS (Wine, не kernel).
PS7 YouTube/browser-path/StrictMode fixtures PASS после последней поправки.
Windows filters, driver version/pins, game UDP/anti-cheat policy, TLS/mTLS/replay и
production lease1800s не менялись. Секреты/diag-result/cs-diag/observe-result не коммитились.
CI: https://github.com/buzzmean/neexitlag/actions/runs/40015935949
и https://github.com/buzzmean/neexitlag/actions/runs/40015936148
и https://github.com/buzzmean/neexitlag/actions/runs/40015935881 — steps=[]; НЕ CI PASS.

**ТОЧНЫЕ команды владельцу**: корень репозитория, PowerShell5.1 администратора
под SID владельца. Chrome/Edge оставить открытыми. Блок P0 изменяет VPS только
при запуске владельцем; агент его не запускал. Сначала Inspect, затем deployment.
P0 уже обновляет службу этим же исходным кодом на ПК запуска; следующий LOCAL-only
блок нужен вместо P0, если сейчас требуется обновить только клиент без VPS.

```powershell
git switch codex/work
git pull --ff-only origin codex/work
### YouTube LOCAL-first — финальная проверка фоновой пробы (23.10.2026)
CDN root403/404/405 после verified TLS+GET допускает только минутную LOCAL trial,
не подтверждённый media cache. Длительный TTL требует реального transformed
ServerHello/64KiB/250ms. Иначе alternate candidate оставался на relay даже при
рабочем LOCAL TLS. Разделены engine/cache keys для visible SNI и hint-only:
совпадение host/IP не должно подменять exact-IP absolute split SNI-scoped engine.
Регрессии этих двух случаев race×3 PASS; повторный полный Go race/vet, Linux/Windows build и Windows vet PASS. PS7/StrictMode YouTube/browser diagnostics fixtures PASS; native Windows/PS5.1/ISP не проверены.

### YouTube LOCAL-first F — bounded диагностика и команды (23.10.2026)
diag-no-vpn2026-10-23.4: youtube-media DIRECT/LOCAL/RELAY, candidate/transform confirmation, TX/RX, scoped cache/recheck/transition reasons, explicit omitted recent records. Отдельно primary/standby authenticated identity и реальный TCP peer из P0 provider; старый VPS-01 pool не выдаётся за Hostkey pool. PS5.1/pwsh workflow gates и safe/StrictMode fixtures добавлены. `NOT_CONFIRMED` не равен доказанному NO_TRANSFORM при утраченном ledger.
P0 add-relay one-command проверяет active identity/owner, до SSH собирает клиент с настоящим pinned filter compiler; после remote proof/commit обновляет только существующую службу, ждёт connected+optimizing+backend.ready+signed primary. GUI закрывается same-owner pipe только если мешает его legacy recovery helper; браузеры/игры не закрываются. relay-status показывает реальные1s payload deltas и peer endpoint, не только подписанный план. Native PS5.1/Hostkey execution ещё не проверены.

### YouTube LOCAL-first B–E — transport integration (23.10.2026)
Общий browser opener вместо безусловного youtube_relay_first: DIRECT → scoped LOCAL → RELAY только после неуспеха; cold650ms dial/400ms reply, LOCAL650/900ms, максимум одна local candidate на handshake. Новые auto engines не требуют ручного preset/baseline; подтверждённый configured baseline остаётся первой LOCAL стратегией. Complete ClientHello-only replay до ANY origin byte, без 0RTT/application/partial hello; consumer/writer gates и одноразовая передача управления ребёнку устраняют гонку вложенных retry. TLS alert/partial ответ не скрывается/не переигрывается, только negative verdict для следующего потока. Старые TLS после ответа не мигрируют.
Кэш256, network/profile/domain-family/port, TTL10min; фоновый verified TLS+GET/minute, отдельный source tuple и4shared probe slots, конечный sweep кандидатов. LOCAL success требует native committed transform + ServerHello +64KiB +250ms; это transport, не playback proof. 8–64KiB/4s без прогресса — явно слабый hint для нового соединения, без закрытия текущего потока этим признаком; idle HTTP/2 нельзя точно отличить без MITM. Обычный unanswered outbound midstream watchdog остаётся осторожным. Однозначные YT hints или all-YT hint set могут разрешить exact-IP/app scoped absolute split; mixed/protected hints остаются relay, TrustedOrigin=false. Все hint candidates проверяются до диагностической обрезки8.
Native/PAC сообщают фактическую текущую pre-ServerHello ступень; capture pipe обновляет только evidence от optional policy reporter, не выполняет reroute. OpenAI/Telegram/Anthropic/правила/DNS53/private/IPv6/игры вне лестницы, QUIC→TCP сохранён. Portable TLS+HTTP200/128KiB cold→LOCAL/warm cache ×3 PASS (~0.72–0.75s/~0.32–0.34s, relay0), но transform seam в этом тесте mocked: реальный Windows/ISP desync/1080p ещё НЕ ПРОВЕРЕН. Unit finite replay×10, alert/partial/early data/no replay, scope/TTL/cache/guard/baseline и native DLL checksum/70filters+iSCSI evaluator PASS (DLL под Wine, не kernel injection).

### P0 — security уточнение, supersedes a808ad0 (23.10.2026)
Не использовать ранний a808ad0 как deployment: одинаковая audience VPS-01 на двух физических узлах означает независимые replay caches. Выбран уникальный server ID VPS-PRIMARY на146.0.77.98; старый VPS-01/2.56.177.4 остаётся неизменным. Билеты/lease renewal фабрики всегда запрашиваются для соответствующего node.ID, cross-audience replay отвергается. Root helper блокирует prepare/commit старого ID. CA/коды/client certs не меняются.
Без переустановки, но с однократным обновлением EXE существующей службы: add-relay-vps.ps1 теперь делает это в той же команде (owner admin SID, pinned DLL gate до remote mutation, штатное same-owner GUI close при update, без закрытия браузеров/игр). Signed primary подключён как локальный provider VPS-01; новый OPEN может выбрать authenticated standby при unavailable/open failure, никогда DIRECT. Уже возвращённый TCP/UDP остаётся на своём provider; при самом обновлении службы обычное краткое отключение proxy потоков неизбежно, их продолжение не обещается. Other legacy PCs требуют того же штатного обновления службы/образа: скрытого автоапдейтера нет. Это минимальный P0, не реализация балансировки/RTT/load/A–E.
`relay.primary.status` и relay-status.ps1 показывают реальный выбранный authenticated server ID, последний TCP session peer endpoint и payload TX/RX (UDP endpoint не выдумывается). Проверены Go race primary/standby new-open fallback, сохранение ранее открытого сокета и replay/cross-audience rejection; PS7 parser/PS5.1 subset и Python RPC fault fixtures PASS. Реальный Windows service update/Hostkey deployment/Ubuntu/SSH ещё НЕ ПРОВЕРЕНЫ. Все production действия — только владельцу, docs/RELAY_HOSTKEY_P0.md переписан под безопасный вариант. Полные финальные проверки общей ветки выполняются перед завершением LOCAL-first.

P0 дополнительная регрессия: пустой signed registry не должен публиковать typed-nil `*transport.Pool` в routing.Relay. Исправлены primary/standby nil guards; защищённое открытие возвращает ошибку, не panic/DIRECT. Race test ×3 PASS.

P0 final ownership: подготовка после потерянного ACK/DB failure повторяет owned pending reservation; commit/rollback не перетирают чужую heartbeat rotation. PRIMARY address move требует новой audience. Leaf/ownership unit race PASS; production не запускался.

### YouTube LOCAL-first B (packet часть,23.10.2026)
Browser-only fake5x2/fake8x2 + multisplit/disorder: собственные2 old-sequence TLS fake, ceiling5 outputs/input, cutoff1. Existing preset IDs/revisions не менялись. Учтён реальный multi-segment HelloBuffer путь (он раньше выдавал только один fake); tests подтверждают оба fake, sequence/checksum, точное сохранение исходного TLS stream и bounds/cutoff. `go test -race ./core/localstrategy` PASS. QUIC fake/autottl/md5sig не имитируются; выбран разрешённый QUIC→TCP. Ни driver/filter, ни game UDP guard не менялись этим шагом. Browser ladder integration и real Windows/ISP acceptance ещё в работе.

### YouTube LOCAL-first A — аудит (23.10.2026)
Предыдущий relay-first79dfaaa8/ad3c69ce/432899d3 скрывал подтверждённые DIRECT/SNI stalls и отсутствие browser verified baseline, но bypass локально не проверял. Browser-scoped LOCAL требовал manual selected preset/baseline и потому не включался на fresh native. Реально есть TCP fake(old-sequence+TTL)/multisplit/disorder и QUIC→TCP; полного Flowseal/autottl/md5sig/fakeQUIC нет. Новый owner request разрешает bounded browser-only автоматическую лестницу. Подробности/границы повторов — docs/YOUTUBE_LOCAL_FIRST.md. Установленный TLS после origin bytes не мигрирует; бесшовный midstream retry без MITM невозможен и не обещается.

### P0 Hostkey — подготовка переноса (23.10.2026)
Новые root-only relay fleet helper/RPC, `add-relay-vps.ps1 -PromotePrimary`, signed-cache `relay-status.ps1`; docs/RELAY_HOSTKEY_P0.md. Без client reinstall смена существующего VPS-01 в signed registry, poll30s, existing TCP pinned. Сохранён будущий seed CP без его restart; новый server-key только на Hostkey, CSR подписывается существующим CA; local health → PC authenticated QUIC/mTLS /3 + UDP echo + verified HTTPS GET → publication → fresh heartbeat. Atomic DB CAS/publish/rollback без security restore; независимые10min timer/boot recovery; owned credential checks/idempotent retry, hash/CRLF normalization.
Проверено локально: relay-admin/client Go race; PostgreSQL17 isolated atomic publication/CAS/nonce preservation/DB-error rollback PASS;8 Python RPC/fault tests PASS; targeted vet и Windows tools build PASS; PS7 AST/PS5.1 subset PASS. Настоящие PS5.1/SSH/systemd/UFW/Hostkey/40ПК НЕ ПРОВЕРЕНЫ. Production агентом не выполнялся. Старый relay остаётся для явного rollback, автоматический standby/балансировка — A–E после LOCAL-first. Legacy cache показывает назначенный endpoint, но actual endpoints старых TCP не инструментированы — не выдавать за миграцию всех потоков.
Команды владельцу (из репозитория, тот же пользователь; это явные production actions только владельца):
```powershell
& .\tools\windows\add-relay-vps.ps1 -Host 146.0.77.98 -Name hostkey-ams -PromotePrimary -Inspect
& .\tools\windows\add-relay-vps.ps1 -Host 146.0.77.98 -Name hostkey-ams -PromotePrimary
& .\tools\windows\relay-status.ps1
```
При провале journal path печатается безопасно; содержимое journal не присылать (private heartbeat/undo). Для возврата registry: та же add-relay команда с `-Rollback`. CA/client identities/activation/коды/PR8 не менялись. P0 код готов, remote acceptance владельца ещё нужен; далее YouTube LOCAL-first.

### Решение владельца 23.10.2026: порядок P0 → YouTube LOCAL-first → multi-relay A–E
Сначала только подготовка переноса логического VPS-01 на146.0.77.98 с сохранением CA/кодов/клиентских сертификатов. Затем LOCAL-first для youtube-media. Остальная multi-relay архитектура отложена. Агент production не запускает. Предыдущая TTFB задача завершена и её область освобождена; новый YouTube запрос явно разрешает browser_transport/diag интеграцию.

### Опубликованный performance-код и CI

Код/наблюдатель/инструкция запушены: `506b20f`, `f107204`, `b44aeb8`.
Последний полный race exit0 (`video-last-race`), последние vet/build Linux/Windows exit0;
PS7 empty-ledger и10s observer fake-IPC first2empty→8samples PASS. Предыдущие demo
FAIL сохранены в рисках выше, базовый1bb001f воспроизводит общий deadline.
CI `b44aeb8`: Verify https://github.com/buzzmean/neexitlag/actions/runs/39988493361
(core/windows: failure,0steps), Chrome https://github.com/buzzmean/neexitlag/actions/runs/39988493370
(driver/policy: failure,0steps), Edge https://github.com/buzzmean/neexitlag/actions/runs/39988493382
(failure,0steps). Ни один из этих Windows jobs не исполнил код; не CI/Internet PASS.
Этот заключительный handoff-коммит меняет только документ, код тот же b44aeb8.
VPS deployment НЕ нужен; следующее действие владельца — обновить только клиент
по docs/YOUTUBE_RELAY_PERFORMANCE.md и записать120s CSV на уже открытом браузере.

### Итог проверок performance (23.10.2026)

- Последний полный `GOMAXPROCS=2 go test -race -p 1 ./...` exit0; перед ним два поздних полного FAIL в demo deadline25s. Targeted new tests и полный затронутый service race PASS. Demo на новом коде отдельно3/3 PASS; на чистом1bb001f1/3 deadline FAIL (diagnostics HTTP),2/3 PASS. Это не безусловно стабильный suite, production failover/test deadline не изменялись.
- Финальные Linux vet/build и Windows vet/cross-build exit0. Девять pure Windows service EXE tests под Wine10 PASS; не Windows kernel/WinDivert/PS5.1/ISP/player PASS. C#/WPF не затронуты и не пересобирались.
- PS7 browser-path и observer main-loop/CSV/StrictMode с изолированным fake IPC PASS; empty initial video ledger поддержан. PS5.1 предусмотрен Verify, локально недоступен. YAML parse PASS.
- GitHub Verify f107204 https://github.com/buzzmean/neexitlag/actions/runs/39985233531: core/windows failure с steps=[]; раньше владелец подтвердил billing gate, новую annotation не прочли. Никакого CI PASS.
- Журнал решений: downstream credit увеличен только на клиенте после real relay lab; лимиты admission/pool/сервер/авторизация не менять. TCP/TLS streams не переиспользуются между независимыми соединениями; shared QUIC session уже reuse. Нужны измерения данных на ПК для полного объяснения minute-long buffering; 2–3с/1080p/36PC не обещать.
- Владельцу: client update/hash check/без browser restart — docs/YOUTUBE_RELAY_PERFORMANCE.md, затем `& .\tools\windows\observe-video-relay.ps1 -DurationSeconds 120` во время нового1080p video и `& .\tools\windows\diag-no-vpn.ps1 -Only browser,youtube`. VPS обновлять не нужно; агент VPS не запускал.

## 23.10.2026: YouTube relay performance, шаг 3

Добавлены pool-phase diag и read-only120s observer CSV с delta rate/TTFB/wait/open (baseline/reset честные, без browser/service/VPS mutations), PS7 main-loop+AST fixtures. Живые видеотуннели приоритетнее закрытых preconnect; PAC/native IPC budget с явным omitted. Полный race повторён после кода; финальные vet/build/Windows EXE проверки и CI состояние — итог ниже. Документ docs/YOUTUBE_RELAY_PERFORMANCE.md содержит аудит buffers/pool/session reuse, измеренные результаты/ограничения и точные owner команды.

## 23.10.2026: YouTube relay performance, шаг 2

Подтверждён локальный downstream flow-control bottleneck: старый client receive max256KiB/1MiB на real relay при60msRTT ограничивал bulk. Только client advertise: initial256KiB/1MiB, max2MiB/8MiB (auto-tune, pool8 = максимум64MiB stream credit). Server defaults/stream limits/auth/lease unchanged, текущий VPS совместим. Pool recent OPEN теперь разделяет acquire (включая fresh-ticket/session setup) и stream+origin TCP ACK; возраст сессии показывает reuse. Real mTLS relay bulk SHA2561/4 streams, same ticket-authenticated pool bulk/reuse, security budget unit — targeted race PASS. Эмпирическая полная причина 2fps/ISP ещё не установлена; Windows playback не проверен.

## 23.10.2026: YouTube relay performance, шаг 1

Добавлены bounded per-video wire TTFB/admission/OPEN/rates/gaps, 128 записей/24 в snapshot с omitted; native/PAC реальный upstream, не HTTP/FPS proof. Старые схемы UNINSTRUMENTED, без payload/URL/секретов. Targeted service Go race и PS7 browser-path fixtures PASS. Owner подтвердил 1bb001f: видео играет, admission failures=0, старт/рывки остаются. Полное расследование/транспортные проверки продолжаются; production не запускался.

## 23.10.2026 — YouTube starvation D: итог и передача

Коммиты: build1c06fb5, A03eba78, Bbf936f2, C4f38988; D — этот коммит.
Причины: shared8занятopaque, ожидание2с не помогает при удерживаемых permits;
local admission скрыт какopen_io_failed; exactIP/family failure путал SNI YouTube
и обычныйGoogle. Native observed ordinary hints теперь DIRECT, при реальном
собственном отказе scopedcache; YouTube aliases/knownDoH/ambiguity protected.
Дополнение владельца ChromeWebStore: capacity обычного AUTO => фактический DIRECT
plan до payload; protected/explicitRequired не откатываются. No established replay.

D: test-browser-dns -DurationSeconds120 -RequireGoogleDirect требует named video
relay + VK/ordinaryGoogle DIRECT, zero video failure delta, dedicated scope;
неизменность main browser PID/birth и службы. Свежие bytes existingHTTP2 считаются
по delta, старые cumulative/reset не дают proof. Три маршрута требуют один
Chrome profile, не смесь разных браузеров. Canceled отдельно. Видео/перемотку
наблюдает владелец: UI не управляется/не читается, page/playback PASS не выдуман.

Итог: полный go test -race -p1 ./... PASS на финальном Go source; go vet./... PASS;
Linux/Windows amd64 build PASS; Windows platform/service vet PASS; оба Windows test
EXE compile PASS. Восемь Windows admission/idle/concurrency/scope unit cases под Wine PASS,
не Windows kernel/driver/Chrome. Настоящий pinned DLL:70compile variants, private
SAN evaluator/negative/limits и DNS worker fixture под Wine PASS; driver не открыт.
PS7 build-cleanup/quotedflags/installed sentinel и browser-path/120sec acceptance
projection/counter/delta/parser fixtures PASS. NativePS5.1/настоящийWindows и
ISP/YouTube2мин/36PC нагрузка НЕ ПРОВЕРЕНЫ. D также ограничил otherGoogle admission/idle webTCP443, чтобы не трогать
mtalk/push5228; known protected control/API остаётся вне ordinary4 и под
существующим global bounded transport, не обещан общий cap16 на весь relay.
Короткий .example fixture был заменён
на валидное publicDNS имя, vendor-public-name check не ослаблен. C#не менялся.

CI C4f38988: Verify39977101529, core/windows failure с steps=0; новый CIне исполнил
код. Известный accountbilling gate ранее подтверждён владельцем, текущая аннотация
отдельно не прочитана; CI PASS не заявлен. VPS/SSH/production/certificates/credentials/
DB/PR8/другие ветки не тронуты, deployment VPS не требуется.

Владельцу из клона admin sameSID: git pull; startup baseline с открытыми Chrome/Edge;
Close-VeliqGuiForUpdate(толькоGUI); build-e2e с ProgramData driver (обязателен DLL
PASS, старые bin удалены); update-service; SHA256True; запустить GUI; status READY;
test-browser-dns -DurationSeconds120 -RequireGoogleDirect, играть видео/перемотать,
открыть VK/Google/WebStore в старомChrome; diag-no-vpn -Only browser,youtube.
Затем startup ready/disabled/exit и новыеVKвкладки. Команды по одной и read-only SSH
CPU/metrics/ss — docs/YOUTUBE_RELAY_ADMISSION.md. Отдельные12video/4other +32waiting
и20/10сidle — стартовый расчёт, не измерение ёмкости1vCPU/VPS. Полевая приёмка
остаётся открытой, файловая область освобождена.

## 23.10.2026 — YouTube starvation C

Bbf936f2: dedicated video12/other4 и capacity DIRECT перед payload (обычные AUTO,
не protected). C: diag2026-10-23.2 показывает scope, оба пула, оба waiters, failure
vs cancel, idle cleanup, ожидание и attempts обычного DIRECT fallback. Старое
shared поле помечено compatibility video-only; отдельная сумма16 ясна. C# не менялся.
docs/YOUTUBE_RELAY_ADMISSION.md: расчёт bounded budget, оговорка про отсутствие
реальных player/36PC измерений; read-only SSH top/systemctl/metrics9090/ss команды
владельцу. Ни одна команда на VPS агентом не выполнялась. PS7 StrictMode/old-service
snapshot/safe fields fixture PASS; PS5.1 native проверки ещё не запускались.
Далее D, область занята.


## 23.10.2026 — YouTube starvation B

A03eba78 запушен. B: dedicated googlevideo12 + otherGoogle4 (32waiters суммарно,
2с/1с budget), без конкуренции с opaque за video slots. Глобальный pool8×48 и
security не менялись. Idle3с доI/O,20сvideo/10сother после; stale callback generation
не закрывает working stream, local idle EOF учитывается отдельно. Native/PAC
показывают local admission code вместо open_io_failed. Cancellation отдельна.
Владелец добавил ChromeWebStore: local/stream capacity OPEN failure обычного AUTO
origin возвращает реальный DIRECT plan до payload. Protected/DoH/explicitRequired
и protected DNS ambiguity не имеют fallback. Не мигрировать successful streams.
Go race core/service PASS; player-like controlled16flows при opaque4 и других
admission/idle/fallback regressions x3 PASS, это не YouTube/player Internet PASS.
Исправление meter nil accounting в zero-value fixture покрыто. Общие race/vet/build
проверки идут; native Windows/ISP ещё не выполнялись. Далее C/D, область занята.

## 23.10.2026 — YouTube starvation A

Сборка исправлена 1c06fb5. A: ordinary Google DNS/outer hints не наследуют чужой
Google-family/exactIP failure; несколько только ordinary hints также DIRECT.
YouTube API/CDN aliases, protected ambiguity, DoH и explicit BLOCK не ослаблены.
Реальный отказ DIRECT ordinary hint set имеет собственный network/profile/IP/
hint-hash cache, TTL2мин/max256; unrelated SNI не отравляет ordinary origins.
Targeted Go race и vet service PASS; native Windows/ISP ещё не проверены.
Далее B/C/D, область занята.

## 23.10.2026 — YouTube starvation: исправление сборки

PS5.1: native -test.run/-test.v/-test.timeout теперь quoted literal arguments.
build-e2e перед любыми проверками удаляет только известные generated Windows/Linux
бинарники, при любой последующей ошибке удаляет частичную новую сборку. Config/
identity/installed-like directories защищены; неизвестные файлы не удаляются.
PS7 isolated early/late failure + installed sentinel/AST argument fixture PASS.
Настоящий PS5.1 повторяется владельцем/CI; работа на VPS не выполнялась.
Далее A/B/C/D, область занята.

## 23.10.2026 — browser DNS A–E: передача владельцу

Коммиты: A06feff6, B9984a10, Ccf0fe52, D36a08a5; E — этот коммит.
E: docs/BROWSER_DNS_ROUTING.md содержит update/read-only команды по одной.
Новый test-browser-dns.ps1 наблюдает90с NEW native flows, фиксированные безопасные
поля, max512наблюдений, неизменность основных Chrome/Edge PID/birth из startup
baseline. Отдельно требует известный googlevideo RELAY_REQUIRED с ServerHello/bytes
и VK DIRECT. Opaque/PAC/5-byte alert/empty snapshots не дают положительный verdict;
страница/видео и полная no-leak proof не выдумываются, lost delta видна. Existing
HTTP/2 reused => честно NOT_OBSERVED, не изменение существующего egress.

Итог локальных проверок: полный go test -race -p1 ./... PASS; go vet -p1 ./... PASS;
Linux и Windows amd64 go build -p1 ./... PASS; Windows platform/service vet PASS;
оба Windows test EXE скомпилированы. Настоящий pinned WinDivert2.2.2 DLL под Wine:
70compile variants + privateSAN evaluator + negative syntax/complexity + DNS
UDP reply/bounded workers fixture PASS. Драйвер не открывался. PS7 StrictMode
browser-path/DNS evidence/acceptance parser и browser-startup fixtures PASS.
C# и WPF не изменялись. Native Windows/PS5.1/Chrome/ISP/video не выполнялись.
Полный race был запущен до окончания правки новой D-test fixture и один раз
остановился на её ошибке типов; после исправления весь suite повторён и PASS.

CI D36a08a5: Verify39874687251 и GameUDP39874687440 завершились failure без шагов
(steps=0 у Linux/Windows), не тестовый FAIL кода и не CI PASS. Известный account
billing blocker ранее подтверждён владельцем; повторный актуальный текст аннотации
не прочитан. Новых зависимостей проекта нет, WinDivert pins не менялись.
VPS/сертификаты/credentials/DB/production config/PR8/другие ветки не тронуты.

Ограничения: только attributable browser-owned IPv4 DNS53; Windows DNS Client
svchost не перенаправляется глобально. Known transit DoH — relay (разрешённый
fallbackB), без forced auto→plaintext, без наблюдения содержимого и TTL encrypted
ответов. CustomDoH+ECH/sharedIP identity не гарантируется. Google group требует
2реальных DIRECT failures/10с, TTL2мин, не доменная идентификация; known video и
opaqueGoogle делят8слотов/32waiters/2с. Protected известные hints/домены fail-closed,
DNS53 не relayed, только HTTPS443 DoH. Игры и ordinary VK/Яндекс DIRECT сохранены.

Владельцу: открытые Chrome/Edge оставить; из клона admin sameSID git pull; startup
-Phase baseline; Close-VeliqGuiForUpdate (толькоGUI); build-e2e с установленным
ProgramData\neexitlag\driver; update-service; SHA256 True; запустить установленный
GUI; status READY; test-browser-dns -DurationSeconds90 + новые YouTube/VK вкладки,
play/seek; diag-no-vpn -Only browser,youtube; startup ready/disabled/exit и новые
VK вкладки. Всё по одной команде в docs/BROWSER_DNS_ROUTING.md. Browser/Internet PASS
пока не заявлен. Все A–E закончены, файловая область освобождена.

## 23.10.2026 — browser ordinary DIRECT/diagnostics D

C запушен cf0fe52, ветка codex/work. D: combined fixtures Chrome/Edge DNS hints +
Google failures => VK/Яндекс DIRECT; protected RELAY_REQUIRED; отключение/другая
сеть не обучают failure cache; игры/чужие DNS не перехватываются. DNS snapshot
показывает только activeTTL,16hint rows + omitted; diag2026-10-23.1 показывает
DNS paths/TTL, failure reasons и known/opaque Google shared8. Исправлен optional
Prop для пустого PSCustomObject под StrictMode. Targeted Go race и PS7 fixture
PASS; native Windows/PS5.1/ISP не выполнены. Полный race повторяется перед E.
Далее E, область занята.


## 23.10.2026 — browser opaque DIRECT recovery C

C: точный IP:порт cache не покрывал следующий Google IP; первый TLS alert/header
ошибочно прекращал наблюдение. Теперь полный ServerHello (минимум38-byte body),
1200мс budget после ClientHello, точные EOF/RST/alert/Write failure, callback cache
до закрытия. Новые соединения только: ни replay, ни migration DIRECT→relay.
Два разных реальных Google TCP443 отказа за10с => network/profile relay-first2мин,
64group cap, не по одному ASN membership. VK/Яндекс/5228 не включаются.
Opaque Google делит8слотов/32waiters/2с с googlevideo, отдельный счётчик неизвестных
доменов; hostname/видео PASS по IP не выдумываются. После disable/stop cache не учится.
Targeted race3пакета PASS; дополнительные TLS timer/polling/header/alert/ServerHello
регрессии count3 PASS; vet и Windows build проверяются перед push. TLS recovery
fixtures заменены с пустого ServerHello на структурно полный; no-byte replay и
no-replay-after-any-origin-byte assertions сохранены. Реальный Windows/ISP не проверен.
Далее D/E, область занята.

## 23.10.2026 — browser Secure DNS B

B: известные DoH endpoints из Chromium snapshot (точные IP/hostname, TCP443) selected native browsers всегда RELAY_REQUIRED; explicit BLOCK, остальные приложения/disabled DIRECT сохранены. IPv6 guard без ослабления. Не ломаем strict Secure DNS преднамеренными RST: выбран разрешённый резерв B, транзитный DoH через relay. Automatic→DNS53 fallback НЕ принуждается и не заявлен доказанным; неизвестный custom DoH+ECH/shared IP не определяется без MITM. Содержимое/TTL encrypted DoH НЕ наблюдается. No-restart/no-settings path; Go targeted race/vet/Windows build PASS, native Windows/ISP ещё не проверены. Далее C–E, область занята.

## 23.10.2026 — browser DNS A

A: IPv4 DNS/53 только атрибутированных native Chrome/Edge перехватывается в UDP/TCP. Защищённые имена — HTTPS DoH через существующий authenticated relay, остальные — wire exchange исходному резолверу. Raw DNS/53 в relay не отправляется. Системный svchost DNS не присваивается браузеру; чужие приложения не затрагиваются. Бюджет 2с, 16 workers/streams, ответы UDP 512..1232 с TC→TCP, 8KiB TCP, TTL/CNAME/ID validation и browser/network hints (1024 адреса). Pending B/C/D/E. Go targeted race/vet и Windows build PASS; настоящий DLL compiler/evaluator 70 случаев + Windows packetWriter worker/capacity fixture под Wine PASS, не Windows driver/ISP PASS. Область ещё занята.

Последний итог Codex 23.10.2026: YouTube starvation305e2fd исправлен в codex/work.
Build1c06fb5, A03eba78, Bbf936f2, C4f38988, D — этот коммит. Dedicated video12/other4,
ordinary Google DIRECT hints/capacity fallback before payload, protected unchanged,
idle/queue/typed diagnostics. Полные local Go race/vet/build Linux/Windows PASS;
PS7 fixtures и Windows EXE8cases/pinned DLL70variants под Wine PASS (не native
Windows/PS5.1/player/36PC PASS). Команды — docs/YOUTUBE_RELAY_ADMISSION.md.
VPS/credentials/certificates/PR8 не тронуты, deployment VPS не требуется.
Файловая область освобождена; ждём двухминутную приёмку владельца без browser restart.

Последний итог Codex 23.10.2026: relay video performance — код и измерения готовы.
506b20f: bounded per-video TTFB/wait/open/rate; f107204: client QUIC receive
initial256KiB/1MiB, max2MiB/8MiB, server defaults unchanged и OPEN phases.
Контрольный local mTLS relay60ms RTT: одиночный23.4→55.0Mbit/s;4stream65.6→68.1.
Это не причина всей минуты2fps и не Windows/ISP/1080p PASS. VPS deployment не нужен.
Финальный полный Go race PASS, Linux/Windows vet/build PASS, PS7 fixtures и9Windows
EXE unit под Wine PASS; повторные demo timeout также воспроизведены на1bb001f,
фикстура не ослаблялась. Observer120s/обновление — docs/YOUTUBE_RELAY_PERFORMANCE.md.
VPS/credentials/сертификаты/PR8 не тронуты. Область освобождена; ждём короткий
замер playback/CSV владельца без закрытия браузеров.

Последний итог Codex 23.10.2026: очередь P0 Hostkey → YouTube LOCAL-first завершена
в коде, codex/work. P0 safe audience/provider `e9cf0b5`, root retry/CAS `7b3168c`;
YouTube packet `be4dc64`, ladder `b86588e`, diag `64ab653`, CDN/hint fix `3a2f39c`.
Старый relay/CP/activation/CA сохранены; production агентом НЕ запускался.
Переход Hostkey требует однократного обновления службы на каждом legacy ПК/образе;
одна owner-команда ставит relay и обновляет существующую службу ПК запуска.
Лестница YouTube DIRECT → scoped LOCAL → authenticated RELAY; TLS не мигрирует
после любого origin byte, midstream hint только для новых соединений.
Финальные Go race/vet/Linux+Windows build+Windows vet PASS; PG17 security/CAS PASS;
98Python PASS; PS7/StrictMode PASS; pinned WinDivert DLL70filters/iSCSI30cases/fake
checksums под Wine PASS. Это НЕ native Windows/PS5.1/kernel/Hostkey/ISP/1080p PASS.
CI `3a2f39c`: Verify40015935949, Chrome40015936148, Edge40015935881 — failure,
все соответствующие jobs steps=[]; тесты там не исполнились. Ранее billing block
подтверждён владельцем, новые annotations недоступны; artifact quota не объявляется
причиной. C# не менялся, нового C#/WPF PASS не заявляем.
Области освобождены; multi-relay A–E отложены, не реализованы. Точные команды и
ограничения — заключительный раздел ниже и docs/YOUTUBE_LOCAL_FIRST.md,
docs/RELAY_HOSTKEY_P0.md. PR8/main/Claude ветки, VPS/credentials/CA не тронуты.

## Решение 22.10.2026 — обязательный компилятор фильтров WinDivert

- Заменить `not (A or B)` в исключении SAN на эквивалентные отдельные тесты по де Моргану: pinned 2.2.2 не разбирает отрицание группы. Не исключать весь порт 3260: это разрешило бы обход защищённого HTTPS при таком локальном bind.
- Проверять каждый production-фильтр на правильном слое через настоящий pinned DLL до генерации службы; не принимать SKIP за PASS. В runtime проверять/сериализовать перед WinDivertOpen, сохранять точную причину/offset/length.
- Истинная проверка DLL выполнена под временным Wine 10 (runtime удалён после проверки); без открытия драйвера и без новых зависимостей проекта. Windows kernel и ISP acceptance остаются за владельцем; CI billing ранее блокировал jobs до steps, зелёный Windows CI не заявлен.
- Владельцу: PowerShell администратора из репозитория → `git pull --ff-only origin codex/work` → `& .\tools\windows\Close-VeliqGuiForUpdate.ps1` → `& .\tools\windows\build-e2e.ps1` (обязателен WINDIVERT_FILTERS_PASS) → `& .\tools\windows\update-service.ps1` → status/backend. Chrome/Edge/игру не закрывать; GUI не менялся. Подробные команды по одной — docs/WINDIVERT_FILTER_VALIDATION.md.

## 22.10.2026 (MSK) — PUBG: финальная передача

Коммиты66332d9 (scope/packet mutex),a2f6dab (local-first),5a4f364(status/GUI/diag).
Дополнительно SCM открыт только SC_MANAGER_CONNECT + SERVICE_QUERY_STATUS:
старый mgr.Connect/OpenService просил ALL_ACCESS и мог получать ложный access
failure у защищённого драйвера. Никакого start/stop/config античит-служб нет.
После этой Windows-only правки повторены Windows build/vet/platform test EXE compile.
Полный go test -race -p1 ./... PASS (без внешнейDB, gated Windows/PG cases могут
быть SKIP); go vet./... PASS; Linux/Windows go build./... PASS; Windows platform/
service vet + оба Windows test EXE compile PASS. Demo actual22.338s PASS.
C#198/198 PASS; WPF cross-build4прежнихCS8604/0errors. PS7 parsing/StrictMode/game
report/club fixtures PASS (native Windows ACL явно SKIP), Python installer7/7PASS.
Реальный WinDivert/SCM process observation,FACEIT/BattlEye совместимость, загрузка
3–5матчей, работа GUI наWindows не проверены. CI PASS не заявлен (billing).
Не утверждаем исправление2610 до owner retest. Нет GeoIP/pin: неизвестный public
PUBG endpoint local-first, confirmed exact DIRECT приоритетен, existing tuple pinned.

Команды по одной — docs/GAME_UDP.md (последний раздел). Admin same-owner из клона:
git pull; Close-VeliqGuiForUpdate; build-e2e; build-desktop; update-service;
install-desktop; запуск установленногоVELIQ. Затем gameudp.status и diag-game-udp
-DurationSeconds180 до следующего матча; не запускать одновременно winws/zapret.
FACEIT/Vanguard службы не отключать. Ожидается PUBG ACTIVE/ready=true, initial_policy
LOCAL_FIRST_UNCONFIRMED_ENDPOINT; WAITING_NEW_FLOW до трафика и24fake первых2packets,
плюс реальные replies/relay counters. Game ACTIVE сам по себе не матч PASS.
VPS/сертификаты/credentials/PR8/другие ветки не тронуты. Область освобождена.

## 22.10.2026 (MSK) — PUBG status/GUI/приёмка

PUBG ready => ACTIVE с отдельным activity=WAITING_NEW_FLOW/LOCAL_DECOYS_OBSERVED;
это readiness, не фиктивные packets/матч PASS. При guard/capture/optimize проблеме
status сразу пересчитывает gate, без задержки1с tick. GUI обычный и club показывает
понятную причину/ожидание, incomplete priming не объявляется успешным. В club-status
добавлены activity/initial_policy; readonly capture.changes(.marker)/diagnostics
разрешены в club IPC, mutation gameudp.set/optimize по-прежнему запрещены.
Диагностика показывает initial_policy/selection_reason и уточняет informational
DriverAntiCheat. docs/GAME_UDP.md обновлён, команды по одной; same-owner/session
Close-VeliqGuiForUpdate закрывает только GUI для обновления, без kill браузеров/игры.
Go race3пакета/vet+Windows vet PASS; C#198/198, WPF cross-build0errors/4старыхCS8604;
PS7 parser/StrictMode/read-only fixture PASS. Реальный WinDivert/SCM/BattlEye матч
не выполнялся; CI PASS не заявлен. Дальше финальный full race/build gates.


## 22.10.2026 (MSK) — PUBG local-first

Только selected TslGame.exe: constructor задаёт LOCAL_FIRST_UNCONFIRMED_ENDPOINT
при discovery, custom enable и восстановлении game-udp.json v2. Первый новый
публичный IPv4 UDP tuple сразу получает K12/N2 (24 decoys перед2оригиналами),
без предварительного6с DIRECT trial. Exact confirmed DIRECT cache этой сети/
exe/address:port имеет приоритет, TTL сохранён. Legacy app-wide DIRECT не служит
авторизацией другого endpoint и не тормозит переход к следующим local вариантам.
Не определяем страну по IP: неизвестный российский endpoint тоже local-first,
оригиналы DIRECT без VPS; fake success не выдаётся за доказанный pure DIRECT.
Остальные игры DIRECT-first; existing tuples pinned, bounded ladder/budget/cutoff
и fallback после local alternatives сохранены. Selection_reason/initial_policy
в evidence, нет постоянных IP pins. Детерминированный Bind+send fixture проверяет
порядок12fake/original×2, последующий pass-through и relay_packets=0.
Go race3пакета PASS; Windows cross-build PASS. РеальныйПК/матч/BattlEye не проверен.

## 22.10.2026 (MSK) — PUBG: область античит-проверки

Причина e5e0127: updateGamesGate и gameudp.set глобально запрещали все игры при
DriverAntiCheat!=empty. Теперь PUBG/BattlEye и остальные непричастные профили не
наследуют FACEIT/Vanguard; loaded/stopped/active client различаются. Установка
сервиса со StartType=System при Stopped не означает activity. CS2 opt-in: loaded
FACEIT driver без client/service допускается, active FACEIT driver+client/service
оставляет compatibility gate. Сам факт матча FACEIT из этих metadata не доказываем;
точная интеграция с матчами отсутствует. Valorant/FACEIT exe остаются gated.
Anti-cheat SCM-query failure отдельно от реального WinDivert conflict check;
PUBG не блокируется из-за недоступного чужого драйвера. REFLECT/process-list/capture/
network/VPN guards сохранены. Toolhelp/SCM query вынесен из packet-path mutex,
startup pending и защита от stale result после reset покрыты race tests.
Go race core/gameudp+platform/windows+client-service PASS; Windows platform test
EXE cross-compile PASS. Windows SCM/настоящие FACEIT/BattlEye/матчи не запускались.
VPS/сертификаты/PR8 не менялись. Дальше PUBG local-first и GUI/diag.

## 22.10.2026 — клубный пилот F: передача владельцу

A399f03d, B2be19c8, Cdb200c4, D4351360, E0051660, уточнения420d052.
Полная инструкция docs/CLUB_PILOT.md: one-command build, private code input,
Setup /CLUBIMAGE, seal без identity,2–3разных overlay/freeze, readonly сводка,
замена pending file/offline image, owner-only revoke, rollback исходным image.
Новый test-club-local.ps1 без installation/SSH/enroll: core/club Go native,
PS fixtures,C# иWPF; optional -Race лишь при уже имеющемся Windows race toolchain.
Публичный Setup без клубного кода; частный образ содержит защищённый code, не ключи.
Одинаковый SID допускается только локально, overlay у каждого ПК изолирован.
Полный admin guest не защищается приложением, отдельного VELIQ password нет;
для настоящего запрета настроек нужен стандартный guest при том же owner SID.
Автоматизировать API/commit Langame без сведений невозможно; это действие владельца
в панели boot-провайдера. Runtime identity в техобраз обратно не сохранять.
Офлайн helper не трогает SCM adminPC. Старый бинарник один не откатывать: старый
LoadManifest не знает club fields; откат всего исходного sealed image.

Финальные доступные gates: Go test -race -p1 ./... PASS (disposable PostgreSQL,
с cache где применимо), go vet ./... PASS, Linux/Windows amd64 go build ./... PASS,
Windows platform/service vet и оба Windows test EXE compile PASS. PG3машины×3boot,
HTTP CSR5freeze c одним request ID/lostresponse,50replacement/seat/revoke PASS.
C#196/196 PASS; WPF cross-build PASS,4прежних CS8604. Python installer7/7 и
PS7 build/package/image/task/rotation/parser fixtures PASS. Ранний demo timing
timeout и последующие успешные повторы описаны выше; сам тест не ослаблялся.
Disposable PostgreSQL container и наш baseline worktree удалены. Native Windows
Inno/SCM/ACL/task,реальный HWND/tray,iSCSI boot/overlay/PUBG античит/36PC нагрузка
не исполнялись. CI PASS не заявлен (известный billing block владельца).
VPS/сертификаты/credentials/PR8/main/ветка Claude не трогались.

Владельцу: PowerShell админа из корня клона: git pull --ff-only origin codex/work;
& .\tools\windows\build-setup.ps1. Затем ровно по docs/CLUB_PILOT.md, сначала2–3ПК.
Область освобождена, дальнейшее ожидание owner Windows/image acceptance.

Дополнительная граница SYSTEM task: публичный Setup tree/все boot dependencies
и их parent directories явно owner=Administrators, запись SYSTEM/Admin only,
owner SID только RX. Future standard guest не сохраняет owner WRITE_DAC после
снятия membership Administrators. Reparse tree отвергается до ACL mutations.
Native Windows temp-tree fixture проверяет owners/protected ACL/no guest write;
на Linux этот участок явно SKIP, PS7 parser/остальные fixtures и Python7/7 PASS.
Это не расширение прав на ProgramData/IPC и не замена Windows admin boundary.
Новые поправки только installer/PS/docs, Go/C# код после последних PASS не менялся.

## 22.10.2026 — клубный пилот: уточнения после проверки границ

Offline -Image code replacement не вызывает SCM административного ПК; legacy
DIRECT browser без explicit BrowserAutoDisabled переводится в AUTO native (иначе
старый fresh/pending цикл возвращается). Явный admin opt-out сохранён. Club probe
берёт стандартный probe.token при изначально пустом ActivationOnly manifest.
Storage exclusion ограничен направлением: local source port3260 НЕ позволяет
обойти protected policy к публичному443; official DLL evaluator имеет эти cases.
HTTP CSR freeze fixture теперь повторяет один staged request ID из общего image
между пятью загрузками, свежие keys/один slot/revoke сохраняются. PG race PASS.
Go targeted race/Windows test EXE cross-compile и PS7 mocks PASS.
Полный race первый раз: demo route selection 25с timeout, повтор count3 имел1fail;
baseline0ea5284 count3 PASS за66с. Полный повтор race -p1 затем PASS без правки
самого demo/таймаутов; это наблюдаемая timing-нестабильность, не скрываем её.


## 22.10.2026 — клубный пилот E: замена кода без переустановки

New-ClubCodeFile.ps1 сохраняет введённый владельцем код в private SYSTEM/Admin file
(не argv/лог/публичный Setup). Set-ClubActivationCode.ps1 заменяет только protected
activation-code.json.pending, свежий request ID, без удаления identity/history.
Live режим перезапускает только службу для сброса revoked-runtime состояния;
Image режим сохраняет остановленную службу и требует отсутствие runtime identity,
boot marker/private pending keys. Freeze: менять sealed base/offline original,
не запекать identity от техрежима; если платформа не умеет offline edit, новый
чистый image. Автоматический API Langame неизвестен, не имитируем его.
Изменение локального overlay потеряется после freeze; базовый файл тоже обновить.
PS7 parser/pending-record/mock rotation/image-stop/write-failure restoration PASS.
Реальная Windows ACL/SCM/code replacement и production revoke не выполнялись.

## 22.10.2026 — клубный пилот D: одна сводка загрузки

club.status и owner-readonly ProgramData/neexitlag/club-status.json (каждые 10с):
activation/relay/native browser/game состояния без machine hash/ключей/кодов.
Короткий authenticated TCP+UDP echo через существующий VPS-01 pool с nonce/HMAC,
3с на протокол, начальный jitter 0–29с; успешный повтор раз в5мин, неудача30–59с.
Baseline PASS означает только session/capture+реальный echo; страницы и матчи
UNVERIFIED, native bidirectional отдельно, игры без трафика не выдают PASS.
Get-ClubStatus.ps1 — одна локальная read-only команда; optional список PC читает
только готовый public status по существующему admin-share, не включает SMB/WinRM,
не просит пароль и не меняет firewall. Старый/недоступный файл явно STALE/UNAVAILABLE.
Go nonce/TCP/UDP/status race и PS7 checks PASS; WPF cross-build PASS, 4 прежних warnings.
Live network echo здесь только локальная фикстура; production VPS не вызывался.

## 22.10.2026 — клубный пилот C: сервисные defaults и простой GUI

ConfigureClub до activation/IPC создаёт отсутствующие Chrome/Edge AUTO native по
доверенным полным путям из образа, сохраняет существующий admin opt-out, включает
Automatic; PUBG использует существующий discovery/default-enabled без смены ladder.
Клубный GUI автоматически скрыт в трее, показывает только состояние; нет технических
экранов, отключения/выхода/изменения маршрутов. Внешний IPC club owner — read-only
allowlist (неизвестные методы тоже запрещены), внутренние activation/optimize работают.
Обычный не-клубный GUI/IPC не изменены. Только Windows admin правит manifest/service.
Go targeted race PASS; C# 196/196 PASS; WPF cross-build PASS (4 прежних CS8604).
Живой WPF/SCM/обычный guest на реальном Windows пока не проверены. Full-admin/UAC-off
пользователя приложение не ограничивает; для защиты настроек нужен стандартный гость.

## 22.10.2026 — клубный пилот B: чистый общий образ

Setup /CLUBIMAGE=1 /CLUBCODEFILE=<private file> запечатывает ActivationOnly:
без identity/device key/probe token, SYSTEM boot task разрешает enrollment лишь
после следующей загрузки. Служба Automatic, на этапе запечатывания остановлена.
Foreign task и существующая identity не удаляются/не перезаписываются. Код только
в защищённом staged-файле образа, публичный Setup его не содержит.
Одинаковый SID не используется в machine hash; local named pipe rejects remote.
Уникальные hardware UUID/MAC + свежий ключ каждого ПК, повторные freeze boot
заменяют ключ/отзывают старый, повторно используют slot. PostgreSQL disposable:
3 машины × 3 загрузки => 3 seats, 3 live/6 revoked; existing 50 replacement и
HTTPS activation integration race PASS. PS7 image/foreign task gates PASS.
Частный iSCSI TCP/3260 явно исключён из native network filter и policy; собственный
storage-профиль не затягивает boot SAN в relay. Windows platform tests cross-compile
PASS; actual official WinDivertHelperEvalFilter проверяется командой build-setup
на Windows перед упаковкой. Здесь DLL evaluator/SCM/Inno/freeze не исполнялись.
Критично: каждому ПК нужен изолированный writable overlay, не общий live state.
Полный admin guest не может быть ограничен приложением; нужен стандартный гость.

## 22.10.2026 — клубный пилот A: сборка Setup

Новая одна команда: `& .\tools\windows\build-setup.ps1`. Чистый tracked HEAD,
.NET SDK8, существующий NuGet helper, approved Inno6.4.3 (winget только если нет),
Authenticode publisher/version, public CA constraints/validity; сборка всех
компонентов и pinned WinDivert; изолированный staging, VELIQ-Setup.exe/SHA256SUMS
в bin/installer. Code/identity/keys не входят в payload; secret filenames rejected.
Python installer 4/4 и PS7 public-CA/parser/package-tamper fixtures PASS.
Реальная Inno компиляция/установка Windows не выполнялась; нужен ПК владельца.
На VPS ничего не выполнялось. Дальше image/boot B и club/status C–F.

## 22.10.2026 — итог opaque/DNS/Telegram/startup

Коммиты: fbab84f (DIRECT default и DNS hints), 49ecb96 (preflight IP до browser TLS),
5c8f3be (startup helper), 885d836 (protected hint поверх IP DIRECT rule),
2804508 (unknown failures только сеть/профиль/IP/порт, не пустой домен).
Область освобождена; рабочее дерево чистое до этой записи.

Финальные проверки на коде 2804508 + уточнение protected outer/DNS aliases: `go test -race -p 1 ./...`, `go vet ./...`,
`go build ./...`, Windows amd64 `go build ./...`, Windows platform/service vet,
compile platform+service Windows test EXE — PASS. PS7 parser + StrictMode fixtures
startup/disable/path/TLS evidence — PASS; реальный PS5.1/Windows SCM/Chrome/Edge/ISP
в этой среде не запускались. C#/WPF код не менялся и повторно не собирался.
Дополнительно: race count=3/5 TLS cert+GET, large fragmented ClientHello exact bytes,
no replay after first origin byte, DNS TTL/network/profile/ambiguity/owner-error gates.
Chrome/Verify CI на 2804508: failure ДО шагов, все четыре jobs steps=[] (GitHub API);
https://github.com/buzzmean/neexitlag/actions/runs/39808274015
https://github.com/buzzmean/neexitlag/actions/runs/39808273984
CI PASS не заявлен; native Windows runtime не исполнялся.

Решение/ограничение: требуемый владельцем opaque DIRECT default не может дать
строгий no-leak для полностью скрытого protected ECH+DoH inner на неизвестном/общем
IP. DNS hints НЕ объявлены доказанным inner SNI; распознанные protected связи
fail-closed, unknown_policy сообщает ограничение. Нельзя считать это релизной
гарантией скрытого protected no-leak. Подробности docs/BROWSER_OPAQUE_ROUTING.md.
VPS/сертификаты/credentials/PR8 не трогались; relay deployment не требуется.

ПК: из корня клона `git pull --ff-only origin codex/work`; закрыть только GUI VELIQ,
браузеры оставить открытыми; `& .\tools\windows\build-e2e.ps1`, при обновлении GUI
`& .\tools\windows\build-desktop.ps1`; elevated SAME OWNER SID
`& .\tools\windows\update-service.ps1`; обычный owner
`& .\tools\windows\install-desktop.ps1`, запустить VELIQ ярлыком.
Новые вкладки VK/Google/YouTube/claude.ai/Telegram;
`& .\tools\windows\diag-no-vpn.ps1 -Only browser,youtube`.
Только при проверке неизменных PID: до VELIQ `& .\tools\windows\test-browser-startup.ps1 -Phase baseline`,
затем новые вкладки и `& .\tools\windows\test-browser-startup.ps1 -Phase ready`.
Существующие HTTP/2/TCP не переносятся. Реальные страницы/провайдер и отсутствие
SSL_PROTOCOL_ERROR требуют нового короткого owner retest; Internet PASS не заявляем.

Дополнительная регрессия финального коммита: несколько Telegram DNS aliases на
одном IP не скрывают flow-specific protected outer. Он может выбрать matching
кандидат только для relay/recovery, не DIRECT; DNS ambiguity остаётся в evidence,
inner SNI не объявляется известным. Без outer конкретное имя не угадывается.
Targeted race count=3 и повторные полные gates PASS.

## 22.10.2026 — unknown origin: кэш только по IP

Отдельный фикс после ревью: пустое имя не допускается в доменный timeout cache.
Иначе ошибки разных unknown IP могли бы обучить общий пустой домен и снова
превратить браузер в full-relay. Opaque DIRECT OPEN имеет собственный bounded
1300ms admission и relay retry только после ошибки и до app payload; owner errors
не запускают retry/learning. Первый ответ/таймаут opaque отслеживается отдельно,
ошибка хранится только сеть/профиль/IP/порт 2 мин. Regression race count=3 PASS.

## 22.10.2026 — дополнительная защита DNS hint

При ревью добавлена отдельная регрессия: явное DIRECT IP-правило не может
ослабить protected association, в том числе ambiguous shared-IP; BLOCK остаётся
приоритетным. Для обычных сайтов явные правила сохраняются. Targeted race PASS.
Startup helper запушен 5c8f3be; Telegram шаг — 49ecb96. Общие проверки идут.

## 22.10.2026 — browser startup, шаг 3

Telegram/preflight запушен отдельным коммитом (см. git log).
PowerShell startup helper: help/examples, точная & .\tools\windows\... команда;
без Phase — BROWSER_STARTUP_USAGE до CIM/IPC, без интерактивного ввода.
Bare name без пути не может быть перехвачен скриптом: shell не запускает файл;
это прямо объяснено в BROWSER_HOT_ATTACH.md. Регрессия реально вызывает helper
без параметров в изолированном каталоге. Никаких browser restart/policy changes.

## 22.10.2026 — browser opaque, шаг 2 Telegram

Шаг 1 запушен fbab84f. Для relay-first DNS/outer hint восстановлен вход в
origin health; до передачи browser ClientHello — ограниченная независимая
DoH+TLS/HTTP проверка IP, TTL/singleflight через существующий DNSRecovery.
Пять уже полученных origin bytes не отбрасываются и TLS не переигрывается.
Изолированная real TLS+GET регрессия .40 bad/.99 good: отдельный probe, затем
два настоящих TLS/HTTP app соединения к .99, cache hit без повторной DNS,
cert/SNI проверены; race count=3 PASS. VPS не выполнялся; ISP/Chrome runtime не проверен.

## 22.10.2026 — browser opaque, шаг 1

Владелец подтвердил native hot attach 29fdeb5: Chrome accepted240/opened236,
233 exchanges. Причина тотального VPS: любой ECH/GREASE стирал SNI и выбирал
RELAY_REQUIRED. Native теперь имеет запрошенный DIRECT default для неизвестного
origin; известные SNI/защищённые DNS/outer associations — relay-required.
Read-only OS DNS cache (256 записей, реальный остаточный TTL <=300с, network scope),
неоднозначный IP не объявляется конкретным доменом; provenance/ошибки в native diag .4.
Детали и неизбежное ограничение ECH+DoH/shared-IP no-leak — docs/BROWSER_OPAQUE_ROUTING.md.
Desktop/protected app policy, игры, PAC, TLS/mTLS и VPS не менялись.
Проверки: Go race browserproxy/platform/service PASS; Windows amd64 cross-build PASS.
Windows/ISP/DNS cmdlet runtime не проверены. Следом Telegram first-flight/preflight и PS helper.

### 22.10.2026: игровой UDP — окончательные проверки и передача

Runtime 4ea4645; итоговый commit только документации. Полный post-fix Go race (serial -p1), vet, Linux/Windows all-build, platform/windows и client-service Windows amd64 test EXE compile PASS. C#191/191 и WPF0 errors/4 старых CS8604; PS7 game-udp-checks parsing/StrictMode/read-only/bounded evidence PASS, workflow YAML PASS. Последние общие race прогоны PASS; ранее найдены флейки старых tools/demo (25s deadline, 1/3 также на исходном 5d61ab8) и один local HTTPS timeout под конкурирующей сборкой (отдельно -race -count=5 PASS). Тесты/таймауты не ослаблялись. Новые Windows runtime тесты собраны, не запущены локально.

CI Game UDP: https://github.com/buzzmean/neexitlag/actions/runs/39801168382 ; Linux job 119901926568, Windows 119901927532, оба failure/steps=[] (подтверждено jobs+steps API). CI PASS/driver/game PASS не заявлен. Исторический account billing gate владелец уже показывал; текущие annotations через доступный connector отдельно не проверены. Workflow без artifact dependency, VPS/credentials не использует.

Дополнения владельца приняты явно: механизм общий, PUBG default-on, остальные игры одной кнопкой; RELAY разрешён последней ступенью, включая PUBG (новое дополнение заменяет прежнее «без relay»). Состояние и обучение scoped к игре/сети; healthy DIRECT endpoint исключён из ненужного relay и не удерживает другой endpoint DIRECT; cadence и global amplification bounded; unknown pre-existing flow DIRECT, без replay. Смена сети прекращает priming/training старого flow, не его трафик/egress. Никаких обещаний совместимости BattlEye/Vanguard/FACEIT, снижения пинга или загрузки матча по transport counters.

Владельцу: из клона codex/work под тем же owner SID обновить только клиент через build-e2e/build-desktop/update-service/install-desktop (GUI закрыть только для замены EXE). Полные команды по одной, SHA256, Clear-DnsClientCache, Frankfurt DNS-hints, 3–5 новых матчей и выключенный контроль — docs/GAME_UDP.md. Без Happ/winws; service/PUBG запускают штатно, другие приложения не перезапускают. Проверка: `& .\tools\windows\diag-game-udp.ps1 -DurationSeconds 300`. Это сбор evidence, не автоматический match PASS. Следующий вариант — только следующий новый UDP tuple; игра не принуждается к reconnect. На VPS/ПК агент ничего не выполнял, PR8 не merge, force-push не было. Область файлов освобождена.

### 22.10.2026: игровой UDP — смена сети

Дополнительный точечный fix: network epoch инвалидирует оставшийся first-flight/обучение старых UDP flows, не переносит их маршрут и не смешивает health разных сетей; stale classifier fingerprint отвергается. Targeted race core/gameudp + platform/windows + client-service PASS после fix; общий post-fix Go gate выполняется. Код GUI/приёмки 022ad01; Game UDP CI https://github.com/buzzmean/neexitlag/actions/runs/39800023841 linux/windows failure steps=[] — runtime не исполнялся. Без VPS/credentials/PR8.

### 22.10.2026: игровой UDP — GUI и приёмка

GUI: PUBG default-on, CS2/Dota2/Apex opt-in и выбранный exact-path custom game одной кнопкой, реальные stage/fake/reply counters и понятные ошибки античита. C#191/191 PASS, WPF build 0 errors/4 прежних CS8604, PS7 parsing/StrictMode/report fixture PASS, workflow YAML PASS. Новые Windows native dispatch fixtures собраны, НЕ исполнены. Добавлены monotonic status, reset guard без копирования live mutex, проверка protected helper ancestry и separate healthy DIRECT endpoint cache (QoS не удерживает плохой игровой endpoint DIRECT, рабочий endpoint не наследует relay). Latest targeted race PASS. Новый workflow Game UDP не требует artifacts/VPS; GitHub Verify D https://github.com/buzzmean/neexitlag/actions/runs/39797683423 core/windows завершены failure steps=[] — не CI PASS. Общие Go race/vet/Linux+Windows builds проходили, финальный post-change прогон выполняется отдельно. Команды обновления/контроль 3–5 матчей: docs/GAME_UDP.md. Реального Windows/BattlEye/игры proof пока нет; VPS/сертификаты/PR8 не тронуты.

### 22.10.2026: игровой UDP — строгий first-flight и cadence

Дополнительно закрыты повторное priming после idle packet cache (native endpoint/tombstone), неизвестный старый tuple остаётся DIRECT, rolling cadence ловит редкие replies при продолжающемся outbound, полное K×N требуется для learning. Добавлены monotonic per-flow metadata snapshots, ранняя запись evidence при relay-open failure, IPC 56KiB bound даже с длинными путями/всеми cache entries. BattlEye/EAC presence — отдельная диагностика, не ложный foreign WinDivert conflict; Vanguard/FACEIT game gate сохранён. Steam manifest проверяет appid. Targeted race PASS; Windows EXE compile и Linux/Windows builds/vet PASS, последние GUI/общие gates завершаются. Нестабильность прежнего demo 25s доказана на исходном 5d61ab8: 1/3 failure, и текущем: 1/3; тест не ослаблен. Последующий общий race PASS; финальный прогон будет записан после GUI. Реальная Windows/игра/античит не проверены.

### 22.10.2026: игровой UDP — интеграция захвата и профилей

Готова интеграция per-game ladder с WinDivert: выбор и pin маршрута до UDP admission; свой first-flight sender, реальные inbound cadence/counters, отдельная sniff/REFLECT диагностика конфликтов и SCM/process-name guard Vanguard/FACEIT (без process memory). Steam PUBG default-on, CS2/Dota2/Apex opt-in, arbitrary exact-path game profiles через gameudp.set. Остальные игры/приложения и UDP53/443/LAN/IPv6 не меняются. Race core/gameudp+platform/windows+client-service PASS; targeted vet и Windows all-build/test-EXE compile PASS. Реальные драйвер/античит/игра/Windows seam execution НЕ проверены. VPS не тронут. Следующий этап GUI и owner diagnostics.

### 22.10.2026: уточнение владельца — общий игровой UDP механизм
Готово ядро генерации `eb4c265`. По новому явному разрешению владельца добавлена лестница DIRECT → LOCAL_UDP (6 bounded вариантов) → RELAY, отдельно на игру/сеть; изменение касается только следующих новых UDP tuples. Работающий DIRECT сохраняется, старые потоки не перепримируются/не переезжают. Общий бюджет приманок 256 пакетов/с на ПК; собственный потолок K16/N4. Кэш с TTL, finite exhaustion/cooldown. Новый engine race suite PASS. Интеграция WinDivert/профили/GUI ещё в работе; реальный Windows/античит/матч PASS не заявлен. PUBG opt-in default подтверждён владельцем, Vanguard/FACEIT default-off до runtime compatibility. VPS не тронут.

### PUBG A: ядро локальной приманки и bounded first-flight

Собственная QUIC-v1-подобная генерация crypto/rand на поток (1250 bytes, C8/v1/DCID8); default K12/N2. Отдельный потолок K16/N4/payload1400, flow128/network16 и 256 fakes/s глобально, никаких relay/API зависимостей. Ядро уже содержит bounded cadence learning для подключения этапом D: burst не healthy, immutable вариант текущего потока, 6 кандидатов, cooldown. Scope только точный exe/IPv4 public UDP >=1024. Go race package PASS; WinDivert/ПК/BattlEye ещё не подключены и не проверены.

### CI итог actual-driver gate 7de94e0

GitHub API после push: https://github.com/buzzmean/neexitlag/actions/runs/39759360991 — chrome-policy и browser-native-driver завершились failure за 2–3 секунды, у обоих `steps=[]`. Следовательно ни checkout, ни Chrome, ни WinDivert тест не запускались; это не результат приёмочного теста. Verify https://github.com/buzzmean/neexitlag/actions/runs/39759361435 и Edge https://github.com/buzzmean/neexitlag/actions/runs/39759361292 также failure. Точная свежая annotation недоступна через connector; раньше owner подтвердил billing/spending gate, нельзя приписывать этот отказ quota artifacts. Реальный Windows gate остаётся блокером A/C/F. На VPS/ПК агент не выполнял действий. Этот заключительный commit меняет только handoff и порядок owner-команд в документации; runtime код — 7de94e0.

### Нулевой ingress d3aad37 — F: actual WinDivert browser gate, 22.10.2026

Новая TestWindowsNativeBrowserDriverHotAttach (отдельный Windows CI job) использует production NewWinDivert и подписанный driver, настоящий Chrome/Edge до Controller, profile publication на активном backend, реальный mTLS/одноразовые tickets/QUIC relay и проверку HTTP body/egress. Подставной ingress закрыт, packetWriter/loopback seam отсутствуют. Требует positive native accepted/opened/exchanges и production SYN/accept counters; неизменный browser PID/birth. На установленном ПК fixture откажется до firewall изменения; только одноразовая VM. Это контролируемый HTTP транспорт, не Internet/HTTPS/SCM PASS.

Проверки: полный `go test -race -p 1 ./...` PASS; `go vet ./...`, Linux build, Windows vet/build PASS; Windows platform/client test EXE compile PASS; C#187/187 PASS; WPF 0 errors/4 прежних CS8604; PS7 actual failed-ready prerequisites + parser PASS; YAML PASS. Локальная среда Linux: Windows/Chrome157/ICMP fallback/пользовательский ISP не исполнены. Последний доступный CI 7da5529, https://github.com/buzzmean/neexitlag/actions/runs/39757234963 — failure, steps=[]; не ошибка тестового шага и не artifact quota. Новый actual-driver job после push нужен, не объявлен PASS.

Решение: не ослаблять owner/SID/пути/TLS или protected routing ради ненулевого счётчика. Не удалять !impostor и не менять игры без фактической атрибуции пакетов. Диагностика отделяет capture armed от обмена и показывает точную границу, включая IPv4 inventory независимо от Engine. Реальную причинность d3aad37 и C/F оставлять открытыми до Windows доказательства. Команды обновления и проверки по одной — docs/BROWSER_NATIVE_INGRESS.md; браузеры не закрывать. На VPS/ПК владельца агент ничего не запускал.

### Нулевой ingress d3aad37 — A/C: native QUIC → TCP

Для AUTO/native Chrome/Edge UDP443 теперь fail-closed с однократным ICMP port-unreachable (только IP+UDP заголовки) вместо отдельного opaque QUIC relay-потока без native TCP counters. Это явно запрошенный TCP fallback; DNS53, legacy PAC, игры и другие приложения сохраняют прежнюю политику. Go race регрессии QUIC/Native/Socket/BrowserPolicy PASS; полный Windows build PASS; actual Windows ICMP→TCP и сайты пока НЕ проверены.

### Нулевой ingress d3aad37 — A: воспроизведён dual-stack дефект

SOCKET CONNECT с IPv4-mapped IPv6 адресами давал ключ ::ffff:IPv4, а NETWORK IPv4 SYN — IPv4. Ассоциация utility/NetworkService PID терялась. Новый TestSocketDualStackChromeOwnerMatchesIPv4SYN сначала FAIL на старом коде, после Unmap PASS; PID/birth reuse по-прежнему отвергается. Focused race PASS. Это подтверждённый дефект кода, но его причинность именно на ПК владельца ещё не доказана; реальный Windows driver gate впереди.

### Нулевой ingress d3aad37 — B: диагностика границ

Добавлены bounded SOCKET/TCP-SYN/reflection/UDP443/IPv6 события, фактические фильтры, коды owner-table ошибок и opt-in browser.native.attribution: текущие IPv4 TCP сокеты Chrome/Edge и текущая классификация независимо от routing.Engine. diag версия 2026-10-22.2. Метаданные без payload/commandline/секретов, IPC bounded; пакеты игр не семплируются. Focused Go race, Windows cross-build, PS7 parse PASS. Причина отказа на ПК ещё не доказана; реальные owner/driver наблюдения нужны.

### Нулевой ingress d3aad37 — E

Приёмка disabled/exit/stopped теперь явно сообщает, что ready не прошёл, вместо побочного BASELINE_REQUIRED. Proof ready сохраняется только после успешной проверки каждого браузера и PID/birth. PS7 fixture/parser PASS; PS5.1/реальный ПК не выполнены.

### Нулевой ingress d3aad37 — D

GUI больше не выдаёт armed capture за готовый браузер: нужны native bidirectional counters для каждого AUTO/native профиля. Общий browser_policy health использует тот же факт, а не PAC readiness. Go focused race PASS; C#187/187 PASS. Настоящий Chrome/WinDivert ещё не выполнен; продолжается аудит A–C/F.

### Браузеры G: owner acceptance и итог 22.10.2026

Read-only test-browser-startup.ps1: baseline обоих уже открытых браузеров до GUI; затем native AUTO + оба native-only exchanges + PID/birth/path; OFF/Exit/Stopped delegates строгому прежнему helper. One-way/PAC/один браузер/пустой status не дают proof. Вывод TRANSPORT_ONLY_NOT_PAGE_PASS, настоящий HTTP/TLS/page outcome владелец проверяет по странице. Новые команды и fresh/update/freeze без manual Reload policy — BROWSER_AUTOMATIC_ACCEPTANCE.md.

Найден дополнительный fresh-install bottleneck: app-scoped native TCP reflection firewall admission раньше добавлялся лишь отдельным флагом updater/bootstrap, а ActivationOnly install его не ставил. Install и update теперь по умолчанию проверяют тот же узкий service EXE/TCP rule. Нет глобального firewall disable, расширения правил, запуска capture до активации или изменения чужих policy. Reinstall ACL-only fixture явно отключает сетевую часть, которая отдельно проверяется mocks.

Проверки: полный последовательный Go race ./... PASS, Go vet/Linux build/Windows amd64 cross-build PASS; Windows browser test binary compile PASS. C#187/187, WPF0errors/4 старых CS8604 PASS. Python deployment90/90 PASS. PS7 evidence/firewall/process mock и parse PASS; WindowsPS5.1 предусмотрен workflow, здесь не запускался. YAML parse/diff check PASS. DB-dependent tests без отдельного PostgreSQL не выдаются за новый PG proof. Новая реальная Chrome/Edge fixture с socket-reflection seam подключена CI, но тут не выполнялась; WinDivert/SCM и ISP владельца не проверены.

Решение: native main path для новых IPv4 TCP, SNI/Host only; opaque ECH/UDP443 authenticated RELAY_REQUIRED без DIRECT fallback при активной оптимизации, DNS53 DIRECT. Unsupported IPv6 guard остаётся, IPv6 support не заявлен. PAC только compatibility/advanced; fresh policy не требуется. Browser PID не меняется, старые TCP pinned. Не обещать произвольную новую вкладку как новое TCP (HTTP/2/QUIC reuse). Не обещать kernel fail-closed при смерти WinDivert процесса: отдельного kernel guard нет, прежний owner-authorized availability contract описан честно. Brave/Яндекс/OperaGX/Firefox обнаруживаются, auto runtime support до отдельных Windows fixtures не заявлен.

Владельцу на установленном227160d: admin прежний SID, корень клона, закрыть только VELIQ; git pull --ff-only origin codex/work; build-e2e; build-desktop; update-service; install-desktop; SHA256 equalsTrue. Затем до запуска GUI оставить Chrome/Edge открытыми: test-browser-startup -Phase baseline, запустить VELIQ, новые YouTube/Telegram/claude вкладки, -Phase ready. OFF → новая vk.ru → -Phase disabled; ON → Exit → новая vk.ru → -Phase exit. Не закрывать браузеры/не Reload policy/не менять профили руками. На production VPS/ПК агент ничего не выполнял.

Последняя G-сверка: browser.status/edge.status без proxy теперь сообщают NATIVE_NEW_TCP / NATIVE_READY_NEW_CONNECTIONS отдельно от listening/policy_ready=false; NOT_PREPARED PAC не выдаётся за native failure. Safe native cause сохраняется в GUI journal без exception payload. Regression cases проверены; новые C#187/187.

Финальная гонка OFF: browser.native.enable проверяет service enabled под сериализацией activationMu; запоздалый native prepare после OFF возвращает NATIVE_OPTIMIZATION_DISABLED и не стартует capture. Auto-start rejection показывается нейтрально с safe specific cause в журнале; generic red error не используется. C#187/187 и WPF/native fixture compile0errors/4 прежнихwarnings PASS. Один полный race при параллельной WPF-сборке получил demo_test.go:81 context deadline25s; пороги/код demo не менялись. Два изолированных race повтора PASS (20.39с/21.40с), затем полный последовательный race ./... PASS. Причина одиночного wall-clock превышения не доказана; результат не скрыт и не объявлен сетевым регрессом. Final vet/Linux/Windows/native test compile PASS.

Предыдущий итог Codex: регрессия выключения браузеров A–E исправлена и проверена доступными локальными тестами. A `6b3e9d4`, B `c5450a6`, C `387e642`, D `b69eb8a`, E — этот commit. Область освобождена. Go race/vet/Linux/Windows build PASS, C#180/180, WPF/native fixture compile0 errors/4 прежних CS8604. Реальные Windows/SCM/GP uptake/ISP ещё не проверены; ни VPS, ни ПК владельца агент не запускал. Команды — docs/BROWSER_DISABLE_RECOVERY.md.

Последний итог Codex 22.10.2026: fresh/автоматический native browser A–G. A ba38329, B 3d11370, C 6a4c07f, D 6b20bcd, E f655cdf, F 2dd06b0, G — этот commit. AUTO/PAC круг разорван; native новые TCP не ждут policy uptake. Core/source tests PASS, не actual Windows/Chrome157/ISP PASS. Область освобождена; реальная приёмка — docs/BROWSER_AUTOMATIC_ACCEPTANCE.md. VPS/credentials/PR8 не затрагивались.

Последний итог Codex 22.10.2026: d3aad37 ZERO native ingress. D `6098ded`, E `b328fb5`, B `7da5529`, A dual-stack `ff78065`, A QUIC→TCP `c845dd8`, F `7de94e0`; итог CI — этот commit. Найден воспроизводимый дефект IPv4-mapped SOCKET key; не доказано, что это причина именно на ПК владельца. Native accepted/page/Chrome157/ISP PASS **не заявлен**. Новая actual-driver VM фикстура закрывает прежнюю seam-дыру, Windows test EXE собран, но не выполнен. Полный Go race/vet/build Linux/Windows и C#187/WPF PASS. CI кода 7de94e0: run 39759360991, оба Windows jobs (включая browser-native-driver) failure с steps=[]; actual driver не исполнен. Следующий шаг — реальный Windows gate и безопасный diag 2026-10-22.2; docs/BROWSER_NATIVE_INGRESS.md. Область освобождена, A/C/F остаются НЕ ПРИНЯТЫ без runtime proof. VPS/credentials/сертификаты/PR8 не тронуты.

Последний итог Codex 22.10.2026: общий игровой UDP механизм готов в коде, `codex/work`. Коммиты `eb4c265`, `def5946`, `17b1c69`, `fd8f7b9`, `022ad01`, `4ea4645` (последний runtime). DIRECT → bounded local priming → authenticated RELAY только после предыдущих неудач; immutable новые tuples, healthy DIRECT endpoints отдельно, network epoch, PUBG default-on/остальные opt-in, Vanguard/FACEIT gate. Приманки собственные, без zapret/Flowseal. Реальные counters/GUI/инкрементальная приёмка и ограничения — docs/GAME_UDP.md. Финальный post-fix `go test -race -p 1 ./...`, vet, build Linux/Windows, Windows test EXE compile PASS; C#191/WPF0errors+4 прежних CS8604; PS7/StrictMode/YAML PASS. Windows runtime/WinDivert injection/BattlEye/матчи/RTT НЕ ПРОВЕРЕНЫ. Game UDP CI кода 4ea4645: https://github.com/buzzmean/neexitlag/actions/runs/39801168382 — linux/windows failure, steps=[]; не CI PASS. Ранее владельцем подтверждён billing gate; текущую annotation отдельно не читали. Область освобождена. VPS/credentials/сертификаты/PR8 не тронуты; ветки Claude/direct-plus-flowseal не менялись.

Последний итог Codex 22.10.2026: срочный CAPTURE_START_FAILED e3dc65c — исправление в этом коммите. Причина: WinDivert 2.2.2 не поддерживает `not (...)`; настоящий pinned DLL сообщает parse error offset=117/length=574. De Morgan сохраняет private iSCSI kernel bypass без обхода protected TCP с локальным портом 3260. Все 7 service-фильтров компилируются до сборки EXE и каждого WinDivertOpen; обязательный Windows gate без SKIP, AST inventory, реальные DLL compiler/evaluator/limits regressions. DLL под Wine: 70 комбинаций + 30 iSCSI matches + negative/limits PASS; это не Windows kernel PASS. Финальный полный Go race/vet/build Linux/Windows и Windows vet, PS7 fixture/YAML/7 installer tests PASS. Первый race под параллельными сборками: DNS 1s и demo deadline FAIL; DNS отдельно x3 и полный спокойный повтор PASS, их код не менялся. Реальные Windows/WinDivertOpen/PUBG матчи ещё НЕ ПРОВЕРЕНЫ. Область освобождена; команды — docs/WINDIVERT_FILTER_VALIDATION.md. VPS/сертификаты/credentials/PR8 не тронуты.

### Браузеры F: native Chrome/Edge fixture и регрессии

E f655cdf. Windows fixture: реальный браузер до контроллера, без PAC/proxy flags, original Host/native opener + authenticated mTLS/tickets relay, разный response source, HTTP body, PID/birth, OFF/host-close; только packet-reflection seam, НЕ actual WinDivert/SCM/ISP proof. Windows fixture compile PASS; workflow подключён. Найден и исправлен auto-start нового GUI после предыдущего Exit при Automatic=true; явный OFF в текущем окне не отменяется, Automatic=false не обходится. C#185/185, WPF0errors/4 старых warnings; полный Go race/vet/Linux/Windows gates перед commit. Реальный Windows/Chrome157 недоступен; CI outcome проверяется после final push. CP/relay/VPS не менялись.

### Браузеры E: автоматический UX без технических PAC кнопок

D 6b20bcd. Обычный AUTO/Optimize использует native; PAC/Origin-probe moved в collapsed advanced diagnostics, legacy путь сохранён опционально. Нет фонового silent PAC prepare. Legacy AUTO сохраняет strategy/rules/custom preset/relay ID при adoption; DIRECT opt-out не сбрасывается. Startup rejection показывает нейтральный статус, safe IPC code остаётся в журнале; retry2s, ready сбрасывается при connected/capture loss. C#183/183 PASS; WPF0errors/4 прежних warnings PASS. Реальный Windows UI/Chrome157 не выполнялся.

### Браузеры D: выключение native и совместимость cached PAC

C 6a4c07f. OFF/Exit/host stop для fresh native не требует proxy или registry uptake; Classify новых TCP/UDP возвращает DIRECT, AUTO намерение сохраняется. Переход из native в пользовательский DIRECT использует ReconfigureNewConnections, а не reset established TCP; явный opt-out сохранён. Старый cached PAC DIRECT/idle/SCM recovery guard остаётся из 227160d. Core/service/host targeted race PASS, C# browser disable/startup/readiness проверены перед commit; настоящий SCM/Chrome/ISP не выполнялся. VPS/PR8 не изменены.

### Браузеры C: readiness и native timing

B 3d11370. Native readiness state/reason/configured_profiles/guard отделены от PAC и page verification; при отсутствии capture/relay — PREPARING, при готовности — READY_NEW_CONNECTIONS, при OFF — DISABLED. GUI показывает подготовку/готовность новых соединений, diag2026-10-22.1 показывает фактические native ingress/timers. Публикация профиля при включённом capture начинает epoch до первого flow. Focused native/Fresh race PASS, C#183/183 PASS, WPF0errors/4 прежнихCS8604, PS parse PASS. Реальное время Chrome157/ISP не измерено; VPS не тронут.

### Браузеры B: native не ждёт PAC

A ba38329; classifier native AUTO исключён из browser_policy_suspended bypass, automatic OFF остаётся DIRECT. GUI native prepare не ждёт PolicyReady и не вызывает новый PAC для native; fresh/repair проверяют native, legacy compatibility сохранена. Старые native journals не переустанавливают policy, owned cleanup сохраняет foreign. SNI protected/opaque ECH/UDP443 fail-closed, DNS53 DIRECT; IPv6 guard остаётся без заявления поддержки. Service/platform race PASS, C#182/182 PASS, WPF0errors/4 старых CS8604. Chrome157/WinDivert/ISP здесь не выполнялись; подробности BROWSER_AUTOMATIC_START.md. Далее C — готовность и сроки native.

### Браузеры fresh A: AUTO без PAC-круга

Новые Chrome/Edge startup profiles — native AUTO; старый default DIRECT/pending мигрирует без замены ID/path; явный opt-out сохраняется. Discovery читает App Paths/StartMenuInternet и стандартные пути, Opera GX отдельно; дополнительные браузеры обнаруживаются, runtime support не заявлен. Native profile admission зависит от owner/capture, а не PAC presence. C#182/182 PASS, focused service race PASS; WPF проверяется перед commit. Нативный Windows/Chrome157/ISP не выполнялся. Следующий пункт B — снять PAC-зависимость из orchestration/classifier и проверить native TCP ingress. VPS/PR8 не тронуты.

### Регрессия выключения — E и окончательные проверки

Global Optimize/Automatic off явно отделён от profile DIRECT/pending: compatibility grant только для глобального OFF; active protected pending не заимствует DIRECT. Регрессия проверяет отзыв disabled grant до Backend.Start. Windows freeze/guard fixture использует настоящий SID текущего Windows test owner. Новая изолированная Chrome/Edge acceptance с настоящими registry PAC, owner/PID/birth и HTTP body проверяет active → disable → host close без browser restart/manual policy reload; backend seam не выдаётся за WinDivert/SCM proof. Фикстура скомпилирована, runtime Windows здесь не выполнялся. Workflow подключён.

Добавлены read-only `check-browser-disable.ps1` и fixture для ролей/пустого списка/секретного command line/парсинга; helper проверяет метаданные и явно НЕ выдаёт STRUCTURE_PASS за Internet PASS. GUI сохраняет ACTIVATED/identity при browser error и сразу показывает прямую сеть после OFF, одинаковый следующий status не мерцает.

Финальный последовательный `go test -race -p 1 -count=1 ./...` PASS, `go vet ./...`, Linux build и Windows amd64 cross-build PASS; Windows service native test binary compile PASS. C#180/180 PASS; WPF/native fixture cross-build0 errors/4 прежних CS8604. PS7 mock/parse PASS; Windows PS5.1 и реальный Chromium fixture здесь не запускались. DB-dependent skips не считаются новым PostgreSQL proof; CP/relay/БД production не менялись. Исторический GitHub billing block не исправляется кодом; новый CI outcome проверяется отдельно после push.

Решение A–E: удаление policy не равно live uptake. После реального изменения посылается штатный user GP refresh, его результат диагностируется отдельно от наблюдения Chromium. Пока cached PAC приходит в живой listener, owner-validated OFF запрос идёт OS DIRECT; затем idle listener закрывается через3мин без новых запросов/активных соединений. При SCM stop owner recovery завершает listener по read-only SCM state; современный availability PAC имеет явный DIRECT. Для старого/чужого cached PAC без DIRECT гарантировать одновременно отсутствие всех listener и сеть при отсутствии policy uptake нельзя; ограничение явно описано, не скрыто. Роль старого PID20012 без owner process snapshot не доказана.

Владельцу: администратор под прежним SID, корень репозитория; закрыть только VELIQ перед обновлением, Chrome/Edge оставить. По одной команде: `git pull --ff-only origin codex/work`, `& .\tools\windows\build-e2e.ps1`, `& .\tools\windows\build-desktop.ps1`, `& .\tools\windows\update-service.ps1`, `& .\tools\windows\install-desktop.ps1`; проверить SHA256. Затем `check-browser-disable.ps1 -Stage baseline`, GUI OFF → новая vk.ru вкладка → Stage disabled; ON → tray Exit → новая vk.ru → Stage exit; снова GUI ON, read-only diagnose-browser-processes, Stop-Service → Stage stopped и новая vk.ru. Полная последовательность команд — docs/BROWSER_DISABLE_RECOVERY.md. Без manual Reload policies, без browser restart. На VPS/ПК агент ничего не выполнял; PR#8 не изменялся.

### Регрессия выключения — D

Поздняя browser подготовка теперь проверяет generation перед следующим IPC: disable/exit/state change отменяет старый ticket, запрещены late silent.apply и rollback поверх новой команды. Background не начинает lane при выключенном/неподключённом режиме. Edge/Chrome phase/code остаются в GUI журнале; UI — нейтральный статус, без ложной активации/Exception type; transient Notice сбрасывается на успехе, historical diagnostic сохраняется. UI-контракты обновлены по новому требованию, packet/security checks не ослаблены. C#178/178 PASS, WPF/native fixture compile0 errors/4 прежних CS8604. Реальный GUI на Windows не запускался.

### Регрессия выключения — C

По коду обнаружен отдельный --chrome-recovery owner child того же EXE, прежде живший до GUI stdin EOF; watchdog CONNECT не слушает и Stopped SCM не перезапускает. Реальная роль PID20012 без owner read-only вывода не установлена. Новый child проверяет read-only SCM и завершает listener при STOP_PENDING/STOPPED (250мс), deliberate host stop публикует DIRECT до удаления policy и Close. Добавлены process_role/process_id и diagnose-browser-processes.ps1 без вывода command line; watchdog допустимо остаётся без TCP. Race stop/guard/watchdog PASS, PS parse PASS, Windows host/native test compile PASS. Настоящий SCM/браузер после Stop-Service ещё не выполнялся.

### Регрессия выключения — B

Удаляется только ProxySettings, не registry key. Подтверждён RegisterGPNotification в Chromium loader; после реальной мутации теперь async bounded RefreshPolicyEx(USER,RP_FORCE) под verified owner token (SYSTEM WTS impersonation, dedicated thread); GUI сообщает user refresh при Optimizing=false. Неуспешный API не выдаётся за live uptake: policy_refresh.state/at/requests/live_uptake. Coalescing не теряет последнюю мутацию; native registry fixture проверяет key/foreign/no-op. Go race queue/service PASS, Windows native fixture compile PASS, WPF0 errors/4 прежних warnings. Реальное удаление из chrome://policy ещё не проверено, A остаётся обязательной защитой.

### Регрессия выключения — A

Исправлен реальный путь: native PAC делегировался в уже остановленный WinDivert; off теперь owner-validated OS DIRECT без Engine/relay/learning и без protected rule при явном выключении. Recovery получает только service-owned disabled grant с SID/path проверками; active protection/deny остаются. Убран mass CloseConnections при Optimize/Automatic switch, established TCP pinned. Idle listener уходит через3мин без сокетов, rebind проверен. Core/service race PASS, targeted real TCP echo/auth/idle PASS, Windows service cross-build PASS. Реальный Windows/Chrome пока не выполнялся; остановка secondary recovery — пункт C.

### Заключительные проверки

`go test -race -p 1 -count=1 ./...` PASS при повторном последовательном прогоне;
`go vet ./...`, Linux `go build ./...` и Windows amd64 cross-build PASS.
C#174/174 PASS; WPF/native fixture cross-build0 errors/4 прежних CS8604.
Python installer4/4 и branding3/3 PASS; doc links/diff check PASS. Нативный
Windows host и отдельный PostgreSQL здесь не запускались; DB-dependent cases
не выдаются за свежую PostgreSQL acceptance. Ранее проведённые проверки отмечены
отдельно в истории; новый runtime outcome на ISP владельца не заявляется.

Первый заключительный Go race получил один timing failure:
TestDirectOpenSharesBudgetRatherThanMultiplyingTimeoutByAddresses, 1.18с при
пороге0.5с, direct=1/relay=1. Тест/пороги/production код не менялись. После
завершения параллельных сборок — отдельный race20/20 PASS (3.573с), затем полный
повтор PASS. Это не доказанный runtime regression; точная причина одиночного
wall-clock overrun не установлена, результат не скрыт. Подробное прежнее
доказательство lease test50x до/после — testing/LEASE_REVOCATION_RACE.md.

### Итог очереди A1–E2, 21.10.2026

Опубликованы A1 `3e3efa8`, A2 `6e6d02b`, A3 `8b66e20`, A4 `a694689`,
A5 `7a7be82`, A6 `83c0c63`, B1 `5cb7f89`; разрешённые merge `cee7d64` и
`9dca1e6` не затрагивают runtime. Исправление точного revocation test `a5cfba4`;
C1 `15e91d7`, C2 `a7ce326`, D1 `46c5ceb`, D2 `26306aa`, E1 `bd5de8c`;
дополнительные WPF Closing/confirmed shutdown fixes `8659b8b` и `ac73e62`. E2 — завершающий документ этого commit.

Не подтверждены на реальном Windows: SCM crash/hang/watchdog/freeze boot,
визуальное отсутствие мерцания/новый header, фактическое время startup listener,
новый googlevideo recovery на провайдере, Setup ISCC/install/upgrade/uninstall.
Точная причина всех старых трёх минут listener и всех ORIGIN_DNS_FAILED без
исходных ответов не доказана; устранены конкретные воспроизводимые пути.
ExitLag observer не измеряет RTT (пассивный OS API его не отдаёт), UDP remote/path
не обещаются. Старый Windows/tray ICO пока не заменён утверждённым SVG export.

Решения владельца перед выпуском: production signing/public trust для Setup,
native Windows acceptance на одноразовой VM/ПК и проверка нового ICO при упаковке;
отдельное разрешение/бюджет на WFP вариант B. Setup пока не готов к массовой
раздаче; подпись, artifact/Release не публиковались. Игра AUTO остаётся DIRECT.
На production VPS/ПК не выполнялось ничего, PR#8/credentials не менялись.

CI последнего runtime `8659b8b`: Verify 39668762315 (core/windows), Relay release
39668762336 — failure, все jobs steps=[]; E1 Edge 39668181579 также steps=[].
Это не провал выполненных тестов и не CI PASS. Ранее владельцем подтверждена
account billing block; новые jobs не дошли ни до тестов, ни до artifact upload.
Ссылки: https://github.com/buzzmean/neexitlag/actions/runs/39668762315 ,
https://github.com/buzzmean/neexitlag/actions/runs/39668762336 .

Владельцу: PowerShell администратора под прежним SID, корень репозитория.
Выйти только из VELIQ GUI; Chrome/Edge не закрывать. По одной команде:
`git pull --ff-only origin codex/work`, `& .\tools\windows\build-e2e.ps1`,
`& .\tools\windows\build-desktop.ps1`, `& .\tools\windows\update-service.ps1`,
`& .\tools\windows\install-desktop.ps1`; затем сравнить SHA256 service binary,
проверить SCM Auto/Running и IPC status, открыть VELIQ, проверить обычную и
защищённую вкладку, `& .\tools\windows\diag-no-vpn.ps1 -Only browser,youtube`.
Не выполнять revoke-tests для проверки текущего активированного тестовым кодом ПК.

### E2 — оценка независимого guard

WFP_GUARD_DESIGN.md: только архитектура, подпись/стоимость/этапы и ворота
failure injection; новый driver/WDK/dependencies не реализованы. Текущий A имеет
owner-authorized DIRECT при отказе инфраструктуры; watchdog не называется
kernel fail-closed guard. Для варианта B требуется отдельное решение владельца.
Ссылки проверены, runtime/VPS не менялись.

### A1 — подтверждённое отключение при выходе

Общий OptimizeAsync проглатывал IPC failure через Act, а Available=false не доказывает отсутствия службы. Выход теперь всегда пытается отключить routing отдельным методом, который требует успешный ответ optimizing=false и не проглатывает rejection/cancellation/unavailable. Только подтверждённые SCM NotInstalled/Stopped позволяют выход без ACK. Native fixture получает явную тестовую отсутствующую службу, без реального IPC; добавлены настоящие WPF failure-path assertions. C#174/174 PASS; WPF/native fixture cross-build0 errors/4 прежних CS8604; реальный Windows runtime не выполнен.

### A1 — дополнительная проверка завершения GUI

Найден синхронный путь StopAsync при недоступной службе: Close внутри Closing WPF мог быть отклонён. Финальный Close теперь ставится в Dispatcher после завершения cleanup, native lifecycle fixture проверяет X/tray и отсутствие синхронного reentry. WPF и native fixture cross-build PASS (0 errors,4 прежних CS8604); реальный Windows runtime не выполнялся. Браузеры и маршруты не менялись.

### E1

Импортированы четыре утверждённых SVG из 2dfdb48 байт-в-байт, source SHA manifest и WPF drawing без новой зависимости; header VELIQ. До импорта составлен BRAND_INVENTORY.md, внутренние SCM/IPC/пути/identity не меняются. Python 3/3, C#165/165, WPF0 errors/4 прежних CS8604 PASS; native визуальная/upgrade проверка недоступна. Старый ICO оставлен до отдельной проверенной упаковки, это не полная миграция Windows shell icons.

### D2

Только ZAPRET_OPTION_DESIGN.md: опора на закреплённый локальный upstream audit от14.10, экономия relay как гипотеза; один capture-owner, bounded adapter/state, критерии TLS/HTTP/ISP, conflicts/лицензии. Новых рабочих стратегий по памяти нет, external binary/dependencies не добавлялись; runtime/VPS/драйвер не менялись.

### D1

Только GAME_ROUTER_DESIGN.md: РФ ingress/common egress, authenticated multipath UDP/dedup, pinned TCP/NAT, hysteresis, bounded resources, WFP и матрица EAC/BattlEye/VAC/FACEIT, P0–P2/измеримые критерии. Текущие игры DIRECT и production не менялись; никакие latency/anti-cheat claims не подтверждены.

### C2

Добавлен пассивный OS endpoint observer ExitLag: exact PID/name, creation time, bounded snapshots/CSV/время, TCP remote и только UDP local, без памяти/injection/сетевых probes. RTT честно NOT_MEASURED, смена endpoint не доказательство маршрута. Pure PS fixtures/parse PASS, Windows CI fixture добавлен в Verify. Реальный ExitLag/Windows не запускался; смотрите EXITLAG_OBSERVATION.md.

### C1

Воспроизведён отказ DNS recovery при проверенном TLS и HTTP403/404 на корне googlevideo. Введено address_verified отдельно от успешного HTTP verified; узкая CDN область, relay-only cached plan и TTL сохранены, не page PASS. DNS/scoping/CA/DIRECT+ regressions; diag 2026-10-21.1. Точная причина старых owner ORIGIN_DNS_FAILED не доказана без ответа; Windows/ISP не выполнялись. См. GOOGLEVIDEO_DNS_RECOVERY.md.

### Разрешённое исправление теста отзыва (21.10)

До изменений: bdbd452 — 14/50 device_revoked и 1/50 fixture Dial deadline; 83c0c63 — 21/50 device_revoked. После: 50/50 PASS, точный remote RevokedCode, no-renewal и production Client reason проверены. Только tests/integration/lease_test.go и docs; production поведение/таймауты не изменены. Повторный полный go test -race -p 1 -count=1 ./... без параллельных сборок PASS; vet PASS; C#165/165; WPF0 errors/4 прежних CS8604. Первый общий run получил отдельный midstream 60ms flake; отдельно20/20 PASS, код/таймаут не менялся. Подробности docs/testing/LEASE_REVOCATION_RACE.md. Публикуются сохранённые B1/merge, далее C1–E2.

## Остановка после разрешённого merge (21.10.2026)

Владелец разрешил обычный merge df2a4bc и потребовал остановиться при сбое
проверок, не исправляя молча. Сохранены локальные коммиты:

- B1 `5cb7f89` (ещё НЕ опубликован).
- `cee7d64`: merge df2a4bc, только CURRENT.md, без конфликтов.
- `9dca1e6`: сохранены появившиеся после df2a4bc изменения сайта до a01b36f;
  только website/, brand/*.md, design/README.md, design/tools/build-site.py,
  CURRENT.md. Никакого force-push/переписывания. Сравнение с 5cb7f89 для apps/,
  services/, core/, platform/, tools/, installer/ и lease_test.go — пустое.

Повторные проверки: go vet ./... PASS; C# 165/165 PASS; WPF cross-build PASS
(0 ошибок, 4 прежних CS8604); Python installer contracts 4 PASS; синтаксис
генератора сайта PASS. Полный go test -race -p 1 ./... — FAIL:
TestLeaseRenewalReplayAndDeviceBinding, tests/integration/lease_test.go:375,
Application error 0xb (remote): device_revoked. Это запрос после verifier.Update
со списком отзыва. Тест требует отрицательный ACK; periodic lease guard в
services/relay/server.go вправе первым закрыть сессию как revoked. Файлы теста,
relay, transport/security не изменены слияниями. Другие опубликованные локальные
проверки не заменяют этот failed gate; native Windows не выполнялся.

Код/тест НЕ исправлялся после сбоя. Предложение владельцу: отдельная регрессия,
допускающая отрицательный ACK либо закрытие с точным RevokedCode; никакого
принятия произвольной сетевой ошибки/timeout или успешного renewal. До решения
владельца B1/merge не пушить, C1–E2 не начинать. Первое незавершённое действие:
решение по failed gate, затем повтор проверки и обычный push сохранённых коммитов.
На VPS и ПК ничего не выполнялось. Команды обновления ПК пока не выдаются:
опубликованный клиент остаётся A1–A6, installer/source merge ещё локальны.

### B1

Добавлен offline Inno Setup 6.4.3 payload/compiler workflow, SHA256 и public CA pin, install/update/uninstall без browser control, same-SID owner, экран активации существующего GUI; тихий club staging из ACL-protected файла через stage_for_boot. Код не в EXE/args/logs. Unsigned/signing оставлен решением владельца. Python 4 контрактных теста и PS7 functional integrity fixture/parse PASS (driver hashes mocked), WPF cross-build; настоящий ISCC/Setup/Windows UAC/freeze не проверен. Production trust/подпись/Release не публиковались; при отсутствии public vars CI только non-distributable fixture. Детали INSTALLER_BUILD.md.

### A6

Добавлен TEST_CODE_RETIREMENT.md: точные runtime/API/CLI пути, негативные тесты после удаления, сохранение 003 и security history; предупреждение в runbook, что массовый отзыв отключает привязанные ПК. Только docs; никакого production revoke. Блок A завершён в коде; реальная Windows/freeze/SCM проверка остаётся владельцу.

### A5

GUI различает отзыв кода, отзыв конкретного компьютера (только по подписанному конфигу), истечение, места, неверный ввод, сеть, storage/identity conflict; только русские сообщения без сырого ответа/кода. Отзыв также блокирует повторную автоматическую подготовку PAC и останавливает backend после снятия policy. Server API и информация публичного перебора не расширены. Go/C# локально; real revocation на ПК не выполнялся.

### A4

install уже создавал Automatic; update сохранял выставленный владельцем Manual. После успешного обновления теперь явно Automatic и readback SCM StartMode=Auto; при rollback возвращается прежний режим. Pure PS mocks PASS, реальный SCM/boot не выполнялся.

### A3

Найден невосстанавливаемый путь: bind listener при Restore выполнялся однократно; таймер только проверял отсутствие. Добавлен повтор bind каждого подготовленного endpoint раз в tick без замены живых соединений, sanitized bind reason, listener_attempts и first-ready timestamps. GUI повторяет подготовку pending lane через 2с и показывает «Подготавливаем браузер…», пока relay готов, но браузер нет. Точная причина всех 3 минут на ПК без лога не доказана; реальное время Windows не измерено.

### A2

Убран Busy из фонового polling; нет Changed(пусто) для Apply/строк приложений; сравнение отображаемых значений и коллекций не перерисовывает неизменные данные. Добавлены регрессии идентичного status и изменения только Optimize. Реальное визуальное отсутствие мерцания на Windows ещё не проверено.

## Очередь 21.10 — A1

Закрытие GUI (крестик, включая сворачивание в трей, и Выход) сначала отправляет optimize=false; при ошибке окно не считается закрытым. Production Close и signed device revocation снимают owned PAC до слушателей/relay/backend. SYSTEM watchdog сохранён; nil status/crash, зависание, freeze и штатный stop покрыты моделями. Реальный kill процесса/SCM и питание на Windows не проверены.

## (ar) PAC после activation: startup, достоверный статус и независимый watchdog

21.10.2026: owner подтвердил server/client771c303, schema3/test и ACTIVATED.
Не готов browser lane: connected=false/optimizing=false/backend.ready=false,
но mandatory PAC был записан; Stop-Service удаляет его правильно. Исправления:

- 3675533: реальный bounded CONNECT+TLS+HTTP healthcheck; отдельная AvailabilityPAC
  обёртка с DIRECT для каждого хоста. Legacy pure PAC API остаётся для совместимости,
  production ConfigureActivation включает owner-authorized availability contract.
- 2f33228: backend стартует без круговой PAC предпосылки, первый неудачный startup
  не теряется (capture retry2..30с), явное off отменяет pending retry. DriverError
  сохраняется в backend.error, connection_reason показывает CAPTURE_START_FAILED.
  Snapshot.connected больше не подменяет transport readiness одним Backend.Ready;
  GUI не подменяет его фактом успешного IPC. RelayAuthenticated/Backend/Optimizing
  отдельно, warming_up quality не proxy для transport readiness.
- Перед PAC: ACT state, enabled+Automatic, Backend.Ready, authenticated pool,
  два настоящих CONNECT и certificate-verified CP healthz status=ok. Fixed-target
  local capability не позволяет произвольный proxy/owner bypass, не выходит в
  status/logs, не считается browser ingress/page PASS. Контроль10с кэш/15с lease.
- Снятие owned policy при off/потере readiness/смене browser mode и каждом boot.
  В packet path только atomics/срок допуска: при stale health новые AUTO browser
  flows DIRECT, desktop MANUAL и игровой AUTO->DIRECT сохраняют старые правила.
  GUI DIRECT очищает native auto-флаг, чтобы таймер не переактивировал браузер.
- Отдельный SYSTEM process -browser-policy-watchdog: bounded local IPC/journal
  cleanup. Три IPC failures у известной живой службы -> проверка SCM path/PID и
  custom128 recovery; обычный Stop/Shutdown exit0, без autorestart. Неизвестный
  PID имеет startup grace30с, запросы ограничены минутой. WinDivert handles
  освобождает SCM recovery, браузеры не завершаются. Никаких WFP/driver изменений.
- Завершающий коммит регистрирует task VELIQ-BrowserPolicyWatchdog при install/update
  (AtStartup+AtLogOn, SYSTEM); update останавливает только эту owned task, откат
  сохраняет прежний EXE/ACL/credentials. Uninstall убирает только owned task/PAC.
  Windows CI и test-activation-local включают cmd/service custom-control тесты;
  PS task fixture полностью подменяет cmdlets/sc.exe, не выполняет установку.

Проверки: Go race core/browserproxy/client-service/cmd/platform PASS; ветка игр
проверена существующими unit regressions. vet ./..., Linux build ./..., Windows
cross-build ./... и compile native Windows test EXE PASS. C#153/153, WPF0 ошибок,
4 прежних CS8604. Node реально исполняет PAC, HTTPS fixture проверяет CONNECT и
сертификаты; native registry/custom128 tests добавлены, локально не исполнялись.
Первый параллельный race core/proxy дал timeout в existing RecoveryTransport
fixture во время сборок; отдельный и завершающий последовательные прогоны PASS,
проверки/тайм-ауты не ослаблены. CI для конечного HEAD ещё не подтверждён.

Решение владельца (последний запрос, п.B/C): доступность при отказе proxy важнее
прежнего обязательного fail-closed PAC. Поэтому infrastructure failure может
отправить ранее protected browser domain DIRECT; никогда не заявлять прежнюю
гарантию при crash. Рабочий RELAY_REQUIRED tunnel не получает upstream DIRECT
retry, TLS/mTLS/replay/lease30мин, VPS/credentials/certificates не менялись.

Ограничения: нулевой сетевой перерыв при полном suspend/hang службы не гарантируется.
Polling/IPC timeout, до10с SCM stop и5с restart delay имеют ненулевую задержку.
Старый freeze image может восстановить старый PAC ещё до запуска SCM/task — нужно
однократно обновить образ, а не обещать исправление кодом до его запуска.
Реальный Win32 driver failure на ПК неизвестен без backend.error; устранён
воспроизводимый startup cycle, driver/pins/AV на ПК ещё не доказаны. ScheduledTasks
под Langame, SCM fault injection, Chrome/Edge Internet и freeze boot требуют
нативной проверки. End-user dialogs/browser kill/restart не добавлялись.
Owner update: обычные build-e2e/build-desktop, update-service/install-desktop
под тем же SID, проверка SHA и task/status. Подробные команды по одной — последний
раздел docs/ACTIVATION_OWNER_RUNBOOK.md. VPS/PR#8 не запускались/не менялись.

Дополнение (ar), 21.10.2026: ac3a698 исправляет именно реальный обработчик
EnableOptimizationAsync (кнопка/трей), не только legacy helper: backend-first
BrowserOptimizationSequence, при Automatic off PAC не подготавливается; native
подготовка ждёт PolicyReady без red-error повторов. 153 C# теста PASS, WPF build
0 errors/4 прежних warnings, desktop-native fixture cross-build0 errors/0 warnings.
Windows fixture явно включает Automatic, чтобы проверка exclusivity не стала
пустой после нового UI guard. Реальный isolated PostgreSQL17 + TLS CSR/freeze/
lost-response activation integration PASS (2.243с); контейнер удалён после теста.
Завершающий commit содержит SYSTEM task integration/install/update/uninstall,
PS mock contracts, CI filters и owner runbook. Сервера 771c303 это не касается.

## (aq) Временный короткий test code и безопасное завершение теста

21.10.2026: owner разрешил известное слово test только для controlled testing.
914fb35 — отдельный завершённый hardware/capture/PAC fix. Новая команда
veliq-codes create-test --value test требует admin token на VPS; нет default
credential, только явная admin операция создаёт строку kind=test, [TEST],
2 seats/14 суток. SHA256 only, прежние IP/code hash buckets/CSR/mTLS сохранены.
revoke-tests атомарно отзывает все test codes + связанные devices + signed config,
не удаляет историю и не затрагивает club/personal. Повтор безопасен; известный
отозванный/expired literal никогда не оживляется повторным create-test.

Добавочная migration003 schema2→3 меняет только constraint/версию, не теряет
данные/salt/revocations. Safe deployment/verify признают schema3 отдельным
compatible verdict. В тестах сохранены downgrade/retry старых001/002. До
релиза сначала revoke-tests и проверка доставки revocations, затем удаление
runtime CodeHash exception/API/CLI/test feature. Миграцию/schema compatibility
оставить для истории установок; destructive downgrade/pg_restore не делать.

Проверки: Go race core/enrollment, platform/windows, client-service/cmd,
control-plane и tools/codes/fresh-pc с отдельным PostgreSQL17 PASS;
Linux + Windows build ./... PASS; vet ./... PASS; Python deployment90 PASS;
C#147 PASS, WPF0 errors/4 старых CS8604. Первый одновременный запуск двух
CP suites превысил deadlines существующих50-freeze/5000-revocation fixtures;
последовательный полный повтор CP PASS (45с), бюджеты/проверки не ослаблялись.
Нативный Windows и полный systemd VM сценарий schema3 не запускались, GitHub CI
для последнего коммита ещё не подтверждён. Ничего не выполнялось на production
VPS/ПК владельца, PR#8 не тронут. Команды обновления, создания test, GUI/IPC
активации и массового отзыва — docs/ACTIVATION_OWNER_RUNBOOK.md, последний раздел.
14 суток — срок новых enroll; для прекращения теста нужен явный revoke-tests.
Известное слово не секрет: rate-limit не предотвращает захват двух мест тем,
кто его знает. Новые server/CLI/client бинарники обязательны; старые test отвергают.

## (ap) ActivationOnly: hardware и сеть до активации

21.10.2026: причины — обязательный board serial и unconditional Automatic/GUI
browser policy до enroll. Исправлены independent CIM queries/physical adapter
selection, v3 fallback без manufacturer placeholders, safe hardware IPC/GUI.
До активации WinDivert и browser mutations запрещены. Штатный stop/uninstall
удаляет только exact journal-owned ProxySettings; crash сохраняет guard. Уже
активированная установка при ожидании CP также сохраняет guard. CA/ACL/identity,
игры и маршрутизация не изменены. Локально race (5 пакетов, реальный отдельный
PostgreSQL), vet, Windows cross-build; C# 147/147, WPF 0 ошибок/4 прежних CS8604.
Windows runtime и failure/recovery на ПК ещё не подтверждены. Обновление без
переустановки — docs/ACTIVATION_OWNER_RUNBOOK.md, последний раздел.
VPS/ПК владельца не запускались, PR #8 не тронут.

Решение 21.10.2026: штатная остановка означает намеренный выход из routing и
очищает owned PAC; crash/init failure не разрешают снимать fail-closed guard.

## (ao) 21.10.2026 — сценарий владельца после deployment кодов
- По сообщению владельца ecd310a успешно развёрнут: compatible schema2/security
  preserved, renewal SCHEDULED, leaf20.11, activation_codes=0. Это owner evidence,
  не SSH-проверка агента. Следующий шаг — создание/активация тестовых кодов.
- Добавлен docs/ACTIVATION_OWNER_RUNBOOK.md: owner SSH/veliq-codes club40 и personal2,
  one-time code output/менеджер паролей вне freeze, CA fingerprint, clean-PC
  ActivationOnly install (ca.pem + cp-ca.pem), GUI activation, safe IPC status,
  freeze и stage-for-boot image, read-only inventory/SQL, disable/revoke.
- Runtime изменений не потребовалось. При обычной переустановке НЕ disable:
  прежний hardware/code slot переиспользуется, old device revoked. Disable
  освобождает место для другого hardware и оставляет MACHINE_DISABLED старому;
  enable/reset в CLI нет. Не делать DELETE/UPDATE security records вручную.
- Уточнено ограничение expires_at/--days: запрещает новые enroll, но не является
  автоматическим отзывом существующих devices/tickets; для окончания теста revoke.
  Пустой freeze baseline сам VELIQ не восстановит: повторный install+тот же code;
  unattended boot требует ActivationOnly+CODE_STAGED без device private key в image.
- Targeted go test -race (PG17) core/enrollment/control-plane/client-service PASS;
  tools/codes compile, hardware stable hash PASS. Обе runbook SELECT проверены на
  temporary fixture tables PG17, 23 PS blocks parsed, install params сверены.
  Собственный PG container удалён. На ПК/VPS install/enroll/codes не выполнялись.
- Owner команды по одной: docs/ACTIVATION_OWNER_RUNBOOK.md. Не запускать старый
  bootstrap-fresh-pc для проверки code enrollment: это SSH admin enrollment.
  Actual production code activation/freeze/browser acceptance ещё не доказаны.

- Уточнение владельца: после Stop-Service Chrome ERR_PROXY_CONNECTION_FAILED. Область расширена на owned policy cleanup при штатном stop/uninstall и GUI mutation gates до активации; чужие policy сохраняются, crash guard отдельно от intentional stop.

## (an) 21.10.2026 — NuGet на чистом тестовом ПК
- Владелец подтвердил test-activation-local.ps1 на c9c8af1: native Windows Go
  platform/windows/client-service/fresh-pc PASS, C#145/145, WPF 0 errors/4 CS8604,
  PS5.1 fake SSH fixture PASS. Это не отдельный native прогон 5adc8e5/ecd310a.
- ecd310a — только ownership CURRENT.md для этой задачи; production код равен
  e6dbe3f. Windows SSH/deploy scripts после c9c8af1 не менялись; production engine
  fixes 06ce3d7/5adc8e5 покрыты чистыми Linux full fresh/legacy из (am).
- На чистом ПК не было NuGet sources (NU1100). Runner теперь до C# вызывает
  activation-nuget.ps1: SDK list source, official URL enabled -> no-op, missing ->
  add nuget.org в user config. Чужие sources не заменяются, disabled/name-conflict
  явно диагностируются без печати потенциально секретных feed URLs.
- Добавлен real SDK regression с isolated NuGet.Config: empty, unrelated feed,
  repeated run, official URL под другим именем, conflict, disabled — PASS на
  PS7/Linux; parser всех трёх PS файлов PASS. Регрессия включена в local runner.
  Новая Windows runtime проверка helper ещё не выполнена. Runtime/Go/VPS вне правок.
- SDK/install/download retry не добавлялись: единичный timeout testhost60s у
  владельца прошёл со второго раза, это отдельный сетевой сбой. Зависимости не
  менялись; нужен установленный .NET SDK8, runtime-only по-прежнему недостаточен.
- Команда с ПК: git pull --ff-only origin codex/work; затем
  & .\tools\windows\test-activation-local.ps1. Это тесты/build, не deployment.


## (am) 21.10.2026 — CI billing gate / offline Debian 13 acceptance
- Владелец предоставил точную аннотацию GitHub: job not started because recent
  account payments failed OR spending limit needs increasing. Это billing gate,
  не YAML/runner/permissions и не artifact upload quota; fresh matrix cancelled
  после legacy через fail-fast. Различить платёж и лимит может только владелец
  в Settings → Billing & plans. Агент не меняет платёжные настройки.
- gh получает HTTP 403 в текущей managed среде; GitHub connector смог повторить
  activation run 39568106123 (attempt 2), снова отказ до steps. Endpoints
  annotations/billing для коннектора недоступны; аннотация получена от владельца.
- Подготовлен offline local Docker runner: настоящий Debian 13/systemd 257/PG17,
  отдельные PID/network/cgroup namespaces, network none/только lo, без bind mounts
  host /etc/repo/socket/credentials, только git archive и тестовые loopback PKI.
  Тот же isolated-upgrade/activation сценарий; guard требует root+systemd PID1+
  image marker+нет других интерфейсов/route. Product deps/TLS не менялись.
- Local Windows equivalent: tools/windows/test-activation-local.ps1, только тесты
  и build, PS5.1+Go+.NET8+Python; без install/enroll/VPS. На Linux не исполняется.
- Production-upgrade paths ограничены relevant кодом, docs/site не запускают
  тяжёлые jobs; concurrency отменяет устаревшие прогоны; fail-fast=false даёт
  независимые результаты layout. Это снижение расхода, не обход billing gate.
- Работа продолжается: local actual scenario и найденные deployment regressions.

- Первый local runner ff70fd3 остановился ДО deployment: source archive в /tmp
  исчез при boot/mount/tmpfiles systemd. Staging перенесён в /var/tmp, добавлено
  ожидание systemd running/degraded и непустой failure JSON даже до source copy.
  Docker/PKI cleanup выполнен; ни VPS, ни host /etc не менялись.

- Local Debian13 выявил две ошибки тестового образа: не был установлен D-Bus
  (systemd-run не видел system bus), /run Docker tmpfs был private, поэтому
  credential mount из дочернего namespace не попадал к PID1. Проверка dummy
  LoadCredential до изменения READABLE=False, после make-rshared /run внутри
  контейнера READABLE=True. В image добавлены D-Bus и entrypoint с shared
  container /run; host /run не монтируется. CA ACL/product не ослаблялись.

- Base isolated-upgrade уже прошёл на real systemd257/PG17: 5000 revocations,
  /3 TCP+UDP/два renewal, restore dump в отдельную БД, wrong-binary rollback.
  В activation fixture исправлена инициализация relay_health перед wait_health;
  ip addr replace допускает диагностический повтор в том же private namespace.
  Engine apply после rollback теперь сохраняет исходную ошибку/command/exit code,
  вместо маскирующего UpgradeError. Python deployment: 88 PASS. Чистый полный
  повтор после этих исправлений ещё выполняется; exploratory apply PASS не
  выдаётся за CI/production proof.

- Найдена настоящая взаимная блокировка первого fresh recovery: installer держит
  transaction_lock, arm ждёт systemctl enable --now recovery, recovery блокируется
  на тех же locks. Recover теперь nonblocking как основной production updater:
  live-transaction-owner не запускает rollback; automatic rollback по-прежнему
  ждёт lock. Добавлены реальные flock и CLI regression; Python: 90 PASS.

- Итог 21.10: clean fresh-credential И legacy на 5adc8e5 завершились exit=0,
  Debian13/systemd257.13/PG17.11. Fresh: полный deployment/leaf renewal, реальный
  leaf failure и outer failure injection/rollback, retry schema2, CA unchanged,
  temporary active probes=0, renewal.service success, timer active, rollback
  timers отсутствуют. Legacy: base binary transaction/5000 revocations/real /3
  TCP+UDP/2 renewal/restore dump в отдельную БД/bad binary rollback PASS.
- Дополнительные проверки: Python 90; real-PG Linux activation Go race (четыре
  пакета), go build/vet ./...; Windows amd64 service/helper/codes cross-build;
  C# 145; WPF Windows targeting compile (4 nullable warnings, 0 errors); PS7
  parsing/quoting PASS. Native Windows НЕ выполнен, GitHub steps отсутствуют.
- CI: activation 39585387093, systemd matrix 39585387157 — отказ до шагов, billing
  gate владельца. После Billing & plans нужны rerun/native Windows; кодом оплату
  не исправить. Оба собственных контейнера/PKI удалены. VPS/PR8 не трогались.
- Решение: до отдельного mutating deployment сначала только verify-activation-vps
  read-only с ПК владельца; затем test-activation-local.ps1. Подробные команды,
  результаты/границы и local Linux runner: docs/CI_LOCAL_ACCEPTANCE.md. Полный
  production допуск не объявлен. Область Codex освобождена; работа остановлена
  на внешнем billing gate/native Windows подтверждении, не на Linux сценарии.


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


## (h) Журнал решений

- 21.10.2026, очередь A1–E2: по разрешению владельца каждый пункт опубликован отдельно; merge сайта не меняет runtime. Отдельно разрешённый lease test принимает только точный device_revoked, без ослабления ACK/security checks. E1 импортирует утверждённые SVG без миграции внутренних имён. E2 — только проект; независимый WFP driver/signing требует нового решения, текущий PAC/watchdog не выдаётся за kernel guard. Нативные ограничения и gate failures сохраняются в итоговой записи ниже.

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

## (a) Цель
Чистый Windows 10/11: установил, нажал «Оптимизировать» — заблокированные в РФ сервисы работают без VPN: ChatGPT, Claude, YouTube, Discord, Telegram, браузеры Chrome и Edge. Первый заказчик — клуб на 36 ПК с откатом при перезагрузке.

## Очередь владельца A1–E2 (21.10.2026)

Владелец подтвердил e03ff8a: чистый ПК, активация, relay/PAC READY, Chrome/Edge/Telegram работают; отключение возвращает браузеры в исходное состояние.

| Работа | Файлы | Агент | Ветка | Статус |
|---|---|---|---|---|
| Экономия VPS / LOCAL каталог и автоподбор 1–7 (23.10.2026) | docs/AGENT_RULES и handoff; services/client-service controller/IPC/browser/local/relay; core localstrategy/browserproxy/controlclient/security/flowpolicy/originpolicy/locallearning/routing/traffic; platform/windows local packet code/tests; tools relay-client/windows/catalog; services/control-plane/relay metrics/config; tests/workflows/docs | Codex | codex/work | ОСВОБОЖДЕНО:1–4/TCP443часть5/6–7 опубликованы; Go local gates PASS. LOCAL1080p/Windows/PS5.1/40ПК не подтверждены; HTTP80/QUIC fake/Discord UDP LOCAL не реализованы; optional CP endpoints production не включены |
| Multi-relay/Hostkey P0 (23.10.2026) | новые deploy/vps relay fleet scripts; tools/windows/add-relay-vps, relay-status, relay-speed; core/optimizer/controlclient; services/control-plane relay registry; services/client-service remote/relay selection (не TTFB файлы); tests; docs | Codex | codex/work | ОСВОБОЖДЕНО: P0 подготовлен/запушен с unique VPS-PRIMARY audience; реальный deployment только владелец; A–E следующая очередь, не выполнены |
| YouTube LOCAL-first A–G (23.10.2026) | services/client-service browser transport/local ladder/meter; core/localstrategy; platform/windows local TCP audit/tests; tools/windows diag; tests; docs | Codex | codex/work | ОСВОБОЖДЕНО: audit/packet/ladder/cache/diag готовы; local gates PASS, native Windows/ISP/1080p не проверены; existing TLS pinned |
| YouTube relay скорость/TTFB (23.10.2026) | services/client-service browser meter/native/transport; core/transport; services/relay; platform/windows TCP forwarding (аудит); tools/windows diag/observe-video; tests/verify workflow/docs | Codex | codex/work | ОСВОБОЖДЕНО: client receive credit/OPEN phases/per-video metrics/observer готовы; Windows/ISP playback не подтверждён |
| YouTube admission/ordinary Google/build PS5.1 A–D (23.10.2026) | services/client-service browser policy/admission/meter; core/browserproxy Plan/errors; tools/windows build/filter/diag/acceptance; tests/docs | Codex | codex/work | ОСВОБОЖДЕНО: build/A/B/C/D запушены; локальные проверки PASS; ждёт Windows/ISP 2мин video retest |
| Browser DNS/DoH и opaque recovery A–E (23.10.2026) | core/browserproxy DNS/первые байты; platform/windows DNS/capture/domain; services/client-service browser*; tests/diag/docs | Codex | codex/work | ОСВОБОЖДЕНО: A–E реализованы/запушены; ждёт реальной Windows/ISP приёмки; svchost/custom DoH ограничения задокументированы |
| Срочно: компиляция фильтров WinDivert/iSCSI (22.10.2026) | platform/windows capture/loader/storage/filter tests; tools/windows build/filter check; Windows CI; docs | Codex | codex/work | СВОБОДНО: grouped NOT исправлен; все фильтры обязательны к проверке DLL; actual Windows capture ждёт владельца |
| PUBG: scope античита/local-first/status (22.10.2026) | core/gameudp; platform/windows game UDP guard; services/client-service game_udp; desktop GameUdp; tests/docs | Codex | codex/work | СВОБОДНО: scope/local-first/GUI готовы; реальные Windows/PUBG матчи ждут владельца |
| Пилот клуба: Setup/image/boot/status A–F (22.10.2026) | installer/*; tools/windows club/build/status; services/client-service club/activation/startup/settings IPC; desktop club/tray; tests/docs | Codex | codex/work | СВОБОДНО: A–F подготовлены/запушены; реальный Setup/2–3ПК freeze pilot ожидает владельца |
| Browser opaque/ECH DIRECT и Telegram recovery (22.10.2026) | core browserproxy/domaininspect; platform/windows domain/DNS; services/client-service browser*; tools/windows browser startup/diag; tests/docs | Codex | codex/work | КОД/ЛОКАЛЬНЫЕ GATES ГОТОВЫ; область освобождена. fbab84f, 49ecb96, 5c8f3be, 885d836, 2804508. Windows/ISP и ECH+DoH/shared-IP ограничения — новый handoff ниже |
| Общий игровой UDP обход A–H (22.10.2026) | core/gameudp/*; platform/windows UDP capture/reply/discovery/REFLECT/conflicts; services/client-service game_udp/controller/local_auto (additive conflict metadata); apps/desktop{,-core}/GameUdp и UI; tools/windows game UDP/diag; tests/workflows/docs | Codex | codex/work | КОД ГОТОВ; область освобождена. PUBG default-on; остальные opt-in. DIRECT → local → authenticated relay, новые tuples; античит gate. Local gates PASS; actual Windows/driver/BattlEye/3–5 матчей/RTT не проверены, CI steps=[] |
| Native Chrome/Edge: нулевой ingress d3aad37 (22.10.2026) | platform/windows capture/process/TCP/UDP и tests; services/client-service browser/status; desktop readiness; tools/windows диагностика/приёмка; workflows; docs | Codex | codex/work | КОД И ДИАГНОСТИКА ГОТОВЫ, область освобождена; подтверждён unit dual-stack bug; Windows gate не выполнен, A/C/F не приняты |
| Браузеры fresh/hot-attach A–G (22.10.2026) | services/client-service/browser*, edge*, ipc; platform/windows TCP/UDP domain capture и browser discovery; desktop Browser*/Edge*/Chrome* и UI; tests; tools/windows acceptance; docs/ | Codex | codex/work | ЗАВЕРШЕНО В КОДЕ A–G; область освобождена. Native/PAC circle, startup/readiness/UI/disable/acceptance; Go race/vet/build, C#187 и WPF PASS; actual Windows/Chrome157/SCM/ISP не выполнены. VPS/PR8 не тронуты |
| Регрессия выключения A–E (6adefe5) | services/client-service/browser*, edge*, cmd/service/*, controller lifecycle; core/browserproxy/*; platform/windows/*policy*; desktop recovery/status/startup и тесты; tools/windows watchdog/acceptance; docs/ | Codex | codex/work | ЗАВЕРШЕНО В КОДЕ A–E; область освобождена: DIRECT compatibility, GP notification, SCM/recovery role, честный GUI, Windows acceptance fixture и owner helper; реальный Windows/ISP ретест нужен; production не выполнялся |
| A1–A6, B1, C1–C2, D1–D2, E1–E2 | browser lifecycle/remote/activation/controller и тесты; desktop/core модели/UI/активация и тесты; tools/windows install/update/diagnostics; новый installer/ и workflow; browser DNS и тесты; новый пассивный observer; docs/; brand/assets только после инвентаризации | Codex | codex/work | ЗАВЕРШЕНО В КОДЕ/ДОКУМЕНТАХ 21.10.2026; область освобождена. Native Windows/SCM/freeze/Setup/ISP gates остаются непроверенными; WFP только проект. Production не выполнять |

Последняя сверка Codex 21.10.2026: владелец подтвердил production 771c303/schema3 и создание test; на чистом ПК activation=ACTIVATED/has_identity=True. Выявлен browser PAC при connected/optimizing/backend=False; штатная Stop-Service очистка работает. Текущее исправление lifecycle/GUI/watchdog только в codex/work, реальный Windows/freeze/SCM ретест ещё нужен. Агент production/ПК не запускал; сервер обновлять не требуется. Подробности — (ar), docs/ACTIVATION_OWNER_RUNBOOK.md (последний раздел).

Предыдущий итог Codex 21.10.2026: очередь A1–E2 опубликована в codex/work (реализация/документы, не Internet/Windows PASS). Дальнейших пунктов очереди нет; необходимы native приёмка, проверка Setup и решения по подписи/будущему WFP. Область освобождена; см. заключительную сводку и команды ниже. На VPS/ПК агент ничего не выполнял.

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


Обновлено: 20.10.2026, Codex; `codex/work`, база 6d91aea (fast-forward сайта Claude сохранён; browser работа от 0d20db1). Владелец подтвердил TLS/HTTP отказ исходного Telegram IP с VPS и успех независимого кандидата с тем же SNI. Реализован bounded per-IP first-flight recovery для relay-first, negative TTL и диагностика .2; детали — (ab), docs/TELEGRAM_TLS_DIAGNOSTICS.md. На production VPS ничего не выполнялось; Владелец подтвердил per-IP recovery на ПК. Этап (в): вариант A, тихий owned PAC без browser restart, реализация и ограничения — (ac); PR #8 не тронут.

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

## (r) Первая загрузка подтверждённых browser групп — 19.10.2026

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

### Подтверждение владельца / границы

- e9e31ac: новые вкладки в открытых Chrome/Edge работают, Chrome PID прежний,
  игры DIRECT. Это подтверждённый ISP результат предыдущей сборки.
- Текущая сборка не объявляется новым ISP PASS. Owner diag не коммитился.
- PAC остаётся единственным независимым crash-guard; этап (в) не начат.
  Online PAC вызывает CONNECT, поэтому browser native-only SYN не возникает.
  PAC DIRECT для всех сайтов после crash снял бы protected fail-closed.
  Нужен отдельный устойчивый kernel/WFP guard; watchdog/реестр не атомарная замена.
  Подробный безопасный следующий этап — docs/BROWSER_HOT_ATTACH.md.

## (q) Cold-start relay retry и native-only — 19.10.2026

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

## (p) PAC/native и Edge readiness, 19.10.2026

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

### Обновление Windows

Точный stop-on-error блок: `docs/BROWSER_HOT_ATTACH.md`, раздел «Обновление и короткая
проверка без закрытия Chrome/Edge». Закрыть только GUI; elevated owner PowerShell:
`build-e2e.ps1`, `build-desktop.ps1`, `update-service.ps1`, SHA256; обычный owner:
`install-desktop.ps1`, открыть VELIQ, Optimize, обновить страницы в уже открытых окнах.
Проверка: `& .\tools\windows\diag-no-vpn.ps1 -Only browser,youtube` (версия 2026-10-19.3).
VPS/credentials/certificates/PR #8/ветка direct-plus-flowseal не менялись.


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

## (g) Как проверить без VPN
`& .\tools\windows\diag-no-vpn.ps1` (запускать из репозитория, не из старых копий; версия в шапке отчёта — `2026-10-19.1` или новее). Смотреть шапку (версия, коммит, `build` службы), разделы `sni`, `1b`, `1d` (`direct_stalls_total`, `stall_retries_via_relay`), `1f` (`probe_accounting`).
Доказательство: хеши службы совпадают и есть живая проверка на ПК владельца. Зелёный CI, лаборатория и Go-тесты доказательством работы на ПК не являются.
Управление флагом YouTube: `settings.set` с полным объектом Settings и `"youtube_via_relay": true` означает relay-first сразу для ordinary YouTube. По умолчанию выключен; аварийный fallback от него не зависит. Configured DIRECT+ и protected rules не заменяются.

## (f) Сборка и установка службы на ПК (по CLAUDE.md / docs/AGENT_RULES.md)
Только по просьбе владельца, на его ПК:
1. Браузеры оставить открытыми; закрыть только GUI («Выход» в трее) для замены recovery EXE. Сохранить локальные правки; `git switch codex/work`, `git pull --ff-only origin codex/work`. Полный stop-on-error блок: [BROWSER_DIRECT_FALLBACK.md](../BROWSER_DIRECT_FALLBACK.md#обновление-windows-без-vps).
2. Обычное окно PowerShell: `& .\tools\windows\build-e2e.ps1`, затем `& .\tools\windows\build-desktop.ps1`.
3. Elevated PowerShell под тем же SID владельца: `& .\tools\windows\update-service.ps1`.
4. Обычное окно: `& .\tools\windows\install-desktop.ps1`.
5. Сверить SHA256 `bin\windows-e2e\neexitlag-service.exe` и `%ProgramData%\neexitlag\neexitlag-service.exe`, `Get-Service neexitlag` = Running.

`powershell -File ...` ломает `$PSScriptRoot` (пустая строка в Join-Path): запускать только через `& .\tools\windows\скрипт.ps1`. Надпись «UPDATE COMPLETED» доказательством не является.
Запрещено: `update-vps.ps1`, `install.ps1`, `uninstall.ps1`, `setup-e2e.ps1`, `tools/vps-provision`, `tools/enroll`.

## (e) Кто чем владеет
| Область | Файлы | Агент | Ветка | Статус |
|---|---|---|---|---|
| PAC readiness/lifecycle после активации, owner fail-open fallback | `services/client-service/{activation*,browser_*,edge_*,controller.go,ipc_handlers.go,remote.go}`, `services/client-service/cmd/service/*`, `core/browserproxy/{pac*,proxy*,recovery*}`, `platform/windows/*policy*`, `apps/desktop-core/{Models,DesktopViewModel,Browser*}.cs`, `apps/desktop/{Browser*,Chrome*,Edge*,FreshPCSetup}.cs`, `tools/windows/{install,update-service,uninstall}*.ps1`, новые watchdog файлы/тесты, `.github/workflows/activation.yml`, `tools/windows/test-activation-local.ps1`, desktop/Windows/Go tests, docs/ACTIVATION_OWNER_RUNBOOK.md, CURRENT.md | Codex | `codex/work` | ГОТОВО 21.10.2026: core3675533 и runtime2f33228, регистрация независимого watchdog/CI/инструкции в завершающем коммите. Go race/vet/Linux+Windows build PASS; C#153, WPF0 errors/4 прежних CS8604, pure PS fixtures PASS. Native Windows/SCM/freeze/Internet не проверены; область освобождена; DIRECT при отказе инфраструктуры явно разрешён владельцем, VPS не выполнять |
| Временный код test: hash-only, 2 места/14 суток, bulk revoke | `core/enrollment/{code,test_code}*.go`, `services/control-plane/{codes*,store.go,migrations/003*}`, `tools/codes/*`, `tools/fresh-pc/upgrade_test.go` (schema3 expectation), `.github/workflows/activation.yml` (CLI tests), `deploy/vps/{deploy-activation,inspect-activation}.py` (schema3 compatibility), `tests/deployment/*activation*`, docs/ACTIVATION_OWNER_RUNBOOK.md, docs/handoff/CURRENT.md | Codex | `codex/work` | ГОТОВО 21.10.2026: create-test --value test, 2 места/14 суток, hash-only, IP/code limits, atomic revoke-tests, additive schema3; тесты PostgreSQL/race, Python 90 PASS; область освобождена, до релиза удалить runtime feature, VPS/Windows runtime не проверены |
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
| Дизайн VELIQ: дизайн-система, сайт, макеты программы, бренд-материалы | только новые папки `design/`, `website/`, `app-ui/`, `brand/`; строка в `CURRENT.md` | Claude Code | `codex/work` | ГОТОВО 21.10.2026, область освобождена: сайт переведён на общую подачу (без Langame, технического режима, образа, пароля, трея, времени запуска, размера файла, расположения серверов), тексты `brand/copy.md`, `brand/club-deck.md` согласованы. Ранее: ГОТОВО 20.10.2026 (в т. ч. улучшения сайта: страница vpn.html, блок «Что видит сервер», клубы, выбор устройства на загрузке). Код программы, службы и VPS не менялся. Описание и допущения — `design/README.md`; перенос в WPF и новая иконка в `apps/desktop/Assets/` ещё не сделаны |
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
| Ранний отказ CI и локальная воспроизводимая systemd/leaf проверка | `.github/workflows/{activation,production-upgrade}.yml`, `tests/deployment/*`, deployment fixes по найденным регрессиям, docs | Codex | `codex/work` | ГОТОВО 21.10.2026; область освобождена; fresh + legacy offline systemd/PG PASS на 5adc8e5, 90 Python/Go race/build/vet/C# PASS; native Windows недоступен локально; GitHub billing gate, не объявлены PASS; VPS не выполнять |
| NuGet clean-PC preflight тестового runner + Windows PASS владельца | `tools/windows/{test-activation-local,activation-nuget}.ps1`, `tests/windows-tools/activation-nuget-checks.ps1`, docs/CI_LOCAL_ACCEPTANCE.md, CURRENT.md | Codex | `codex/work` | ГОТОВО 21.10.2026; область освобождена; NuGet helper/real SDK config regression PASS на PS7/Linux; native Windows результат нового helper ещё не проверен; VPS вне области |
| Сценарий владельца: выпуск кодов / ActivationOnly / freeze / слоты | `docs/ACTIVATION_OWNER_RUNBOOK.md`, `docs/ENROLLMENT_CODES.md`, CURRENT.md (runtime только если найдена необходимость) | Codex | `codex/work` | ГОТОВО 21.10.2026; область освобождена; только docs, targeted race/PG17/SQL/PS parsing PASS, production не выполнялся |
| ActivationOnly: hardware fallback/diagnostics и нулевой capture до активации | `core/enrollment/{code,hardware}*.go`, `platform/windows/enrollment*.go`, `services/client-service/{activation*,controller.go,ipc_handlers.go,browser_native.go,browser_silent.go,browser_transport.go}`, `services/client-service/cmd/service/{main,startup}*.go`, `apps/desktop-core/Activation.cs`, `apps/desktop/{BrowserStartup,ChromeRouting,EdgeRouting}.cs`, `platform/windows/*policy*`, `services/client-service/browser_policy_cleanup*.go`, `tools/windows/uninstall.ps1`, desktop/activation tests, docs/ACTIVATION_OWNER_RUNBOOK.md, CURRENT.md | Codex | `codex/work` | ГОТОВО 21.10.2026: hardware fallback UUID+MAC, capture/PAC activation gate, owned cleanup при штатном stop/uninstall; isolated Go race с PostgreSQL PASS, C# 147 PASS, WPF 0 errors/4 CS8604. Windows cross-build PASS, native runtime/ПК ретест не выполнен; область освобождена |
| Discord медленно, DIRECT+ для YouTube | не определены | никто | — | свободно для Codex после диагностики; Go-файлы застоя (выше) не трогать |
| PR #8, `codex/direct-plus-flowseal` | — | не трогать | — | только владелец |

Строки Claude со статусом «передано Codex» сохранены как история; старые «нужно да»/запреты области не отменяют разрешение владельца на A/B. Новую работу снова помечать перед правками по PROTOCOL.md. QUIC/IPv6 routing менять только после отдельного «да».


Новые ветки: `claude/<тема>`, `codex/<тема>`, база — `claude/serene-cori-msra1u`.

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

## (c) Что подтверждено
- Windows/ISP владельца: Chrome и YouTube на `3fec1ae`. Исправленный Edge на ISP ещё не проверен.
- Linux: полный Go race, build/vet, Windows amd64 cross-build/vet и compile native test binary.
- Изолированный PostgreSQL: control-plane race-тесты, включая 5000 revocations; production код CP не менялся.
- CI `6a64f14`: C# desktop tests, WPF build и actual Edge policy/HTTPS/nonce/outage/restart прошли.
  Это контролируемые Windows origins, не Internet PASS. Фикстуры 0aeebcb прошли Linux/Windows Verify, native Chrome, relay release и isolated upgrade; ссылки в (l).

## (b) Состояние на 19.10.2026
- Владелец подтвердил Chrome/YouTube на Windows build `3fec1ae`; Edge на нём не проходил readiness.
- Причина Edge найдена в коде: guard удалял nonce, выданный предыдущим prepare.
  Новый nonce выдаётся после последнего guard и штатного выхода Edge; служба требует настоящий owner GET.
- Код `983d694`, `6c075ae`, `984c1a5`, `1e90c7f`, `6a64f14`; CI-фикстуры `0aeebcb`.
- VPS, credentials, сертификаты и PR #8 не менялись. QUIC/IPv6 routing не менялись.

# CURRENT — состояние проекта (обновляется в конце каждой сессии)


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

