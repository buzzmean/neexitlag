# LOCAL: проверяемый путь пакетов (24.10.2026)

Приёмка владельца f0e181a: relay2160p работает; LOCAL не подтверждён,7фоновых
проверок/6отказов,0CDN trials. Это не доказательство конкретной неисправности TTL
или отсутствия инъекции: прежняя причина объединяла несколько разных отказов.

## Физический путь

`browserOpenUnobserved` → `WinDivertBackend.OpenBrowserLocalTCP` →
`tcpRedirect.dialDelegatedFrom`: служба резервирует точный upstream tuple ДО connect,
связывает его с владельцем/стратегией. NETWORK WinDivert перехватывает SYN/SYNACK и
ClientHello этого собственного сокета по явной delegation; остальные собственные
соединения не становятся LOCAL. `LocalTCPBuffer` собирает bounded ClientHello;
`SendLocal`/buffer deliver отправляют изменённые IPv4/TCP через `t.send`/WinDivertSend.
Это НЕ только socket Write segmentation. Raw TTL/seq/options действительно доступны;
новый драйвер/внешний winws для этого не нужен. Неинъецированный план не является
успехом: counters commit появляются после успешной отправки всего плана.

## Исправления исполнения TCP443

Старый compiler исполнял13 уникальных планов и навязывал old-sequence даже чистым
TS/MD5 стратегиям. Теперь все64 Flowseal TCP443 rules из22 general*.bat имеют
типизированный исполняемый рецепт: fake, multisplit/multidisorder, fakedsplit,
fakeddisorder, hostfakesplit/altorder, SYN-data и их комбинации. Дедупликация всего
каталога даёт44 плана;49 HTTP/QUIC/прочих rules остаются явно unsupported.
С9 legacy кандидатами это53 варианта, помещающихся в64-битную маску подбора.

Raw injection выполняет собственный WinDivertSend; сохраняются реальные байты
TLS, ACK/window/options и относительные sequence original-потока. Fake packet
помечается отдельно, включая interleaved fake/real, с корректными checksum.
Параметры repeats, cutoff1, split markers, zero IP ID, fake list order,
badseq(-10000/override) и badack(-66000), TS delta(-600000), MD5 option19 и seqovl
обрабатываются собственным кодом. При отсутствии исходного TTL/autottl в правиле
не добавляем скрытый TTL2: native fake сохраняет TTL original. Legacy autoTTL
по SYN/ACK остаётся отдельным кандидатом. Это адаптация параметров, не побитовая
копия winws: TLS/STUN/zero модели свои, upstream fake binaries не копируются.

TS/MD5 больше не означают скрытый badseq: fake может иметь текущий sequence;
защита от принятия peer зависит от реально согласованных TS/PAWS либо TCP MD5.
Чистый TS без timestamp option → TIMESTAMP_OPTION_REQUIRED, fake не отправляется.
Только собственный LOCAL socket перед connect включает Winsock TCP_TIMESTAMPS=10;
неподдерживаемая ОС → LOCAL_SOCKET_OPTION_FAILED/TCP_TIMESTAMPS_UNAVAILABLE.
Никаких netsh/global/browser настроек. И это не обещает, что peer примет LOCAL:
истинный TLS/HTTP или WSS probe всё ещё обязателен. SYN-data применяется только
к новому пустому SYN без TFO,16 своих нулевых байт, original sequence; established
потоки не изменяются. MD5 не добавляется, если options не помещаются в TCP header.

LSO first-flight ранее >1500B обходил buffer без transform. Native recipe теперь
принимает bounded ClientHello до MaxHelloBytes, делит оригинальный stream на
≤1500B IP packets, резервирует MD5 header и сохраняет точные wire offsets.
SNI, заканчивающийся точно на конце ClientHello, не отвергается hostfakesplit.
Неподходящий план/размер/бюджет не отправляется частично. Только native browser
TCP443 имеет отдельный потолок64 выходных пакета на первый logical flight;
legacy catalog20/input и обычный DIRECT+5 сохранены. Дальше cutoff: pass-through.

## Диагностика первого этапа

`browser.local.probes`: retained128 strategy×host результатов ВСЕХ групп; отдельный
read-only IPC возвращает последние32 в бюджете48KiB, сообщает omitted_retained.
wire application TX/RX (включая origin TLS), HTTP body bytes, commit, phase/detail,
SYN seen/handshake/server TTL и до64 реально отправленных packet headers:
fake/length/TTL/protocol/TCP flags/relative sequence offset. Ни payload, ни абсолютные
sequence numbers, ни приватные TLS значения не экспортируются. Raw fake bytes
в wire application TX не входят: их видит отдельный список реальных sends.
ORIGIN_SILENT/RST/FREEZE_AFTER_N/EOF_AFTER_N/SHORT_RESPONSE/NO_TRANSFORM/OK разделены;
короткий корректный ответ не назван freeze. OK — проверка собственного HTTPS GET,
не гарантия playback. Отсутствие ledger — неизвестное наблюдение, не успешный send.
Неверный публичный протокол имеет отдельный failure, не выдаётся за freeze.

## Проверяемые риски, ещё не доказанная причина на ПК

При первом SYN OS TCP owner table может ещё не вернуть PID. Native delegation
с OpenDeferred удерживает такой SYN; LOCAL delegation старого кода такой защиты
не имеет и может попасть в self/DIRECT с закреплением решения на весь tuple.
Защита добавлена: LOCAL reservation требует тот же owner proof, что native; pending SYN удерживается до подтверждения владельца. Существующие close/retry тесты сохранены. Unit regression воспроизводит старое неверное решение при owner error; что именно это вызвало на ПК владельца, пока не доказано.
Также фактический commit может отсутствовать из-за handshake/TTL/sequence/guard,
а TLS/HTTP отказ возможен уже после правильно отправленного transform.

QUIC: native browser guard отклоняет selected UDP443 для TCP fallback;
это не fake QUIC. RFC9001 Initial с криптографически верными полями здесь не
реализован; PUBG-shaped fake не доказывает пригодность для браузера. Разрешённый
владельцем fallback сохранён: после него TCP проходит общий DIRECT/LOCAL/RELAY
selector и может использовать подтверждённый LOCAL. Unit проверяет именно
решение fallback, не успешный новый Chrome HTTP/3 → TCP на провайдере. IP/GEO группы не становятся LOCAL, VK/Яндекс остаются DIRECT.
Подробный transport proof и Windows/ISP acceptance ещё нужны.

## Что не проверено

Реальная инъекция на Windows/ISP владельца, PS5.1, LOCAL2160p, Discord WebSocket
и нулевой новый googlevideo relay RX. CP/shared learning не включать для проверки.
VPS/CA/activation/коды/PR8 этим этапом не меняются.

## Новый допуск реального CDN trial (решение владельца24.10)
Committed transform на любом YouTube-хосте запоминается ограниченно/TTL по сети,
профилю, стратегии и hint-mode, даже если public TLS/HTTP неуспешен. Для нового
CDN scope такой proof допускает только trial: max1 globally, живой relay sibling,
cooldown>=3мин/subgroup. LOCAL по-прежнему требует3успешных потоков >64KiB с TLS
и committed transform. Нет proof/relay sibling — нет trial; провал→RELAY/backoff.
Ни обычный origin, ни другая сеть/стратегия/профиль не дают CDN proof. Это не
снимает sustained критерии generic non-CDN origin и не включает CP shared flags.

Полный Go race/vet/build Linux+Windows и Windows vet PASS; checksum/filter DLL
подWine PASS. PS7 diagnostic fixtures PASS. Настоящие Windows capture/PS5.1/ISP
и2160p LOCAL НЕ проверены. Все64 Flowseal TCP443 rules проходят исполнение в packet fixtures; это не proof ISP.

## Общие DPI группы, не только YouTube
Один compiler, LOCAL delegation, buffer, счётчики и кеш применяются ко всем DPI
origins. Discord CDN проверяется несколькими credential-free PNG GET по одному
TLS до >64KiB (≤64 запросов/4с/256KiB, signature PNG+200). Gateway — публичный
TLS WSS Upgrade, проверка RFC6455 accept, Hello и3 свежих nonce ping/pong.
Точный discord.gg допускает ожидаемый HTTPS redirect на discord.com. Последние
два критерия помечены OK_CONTROL_NOT_BULK/BACKGROUND_PUBLIC_*_NOT_BULK_PROOF:
они подтверждают контрольный протокол, не >64KiB, логин, приватный WebSocket или
страницу целиком. Это не ослабляет64KiB Googlevideo и не публикуется как bulk
shared proof. Серверные CP/shared flags не включаются.
Telegram/Meta/OpenAI/Anthropic immutable RELAY floors сохранены; VK/Яндекс DIRECT.

## Лаборатория и границы результата
64 Flowseal recipes: actual packet output, split/order/fooling/checksum/cutoff,
точная сборка оригинального1955B TLS,12KiB offload segmentation, MTU/MD5 и scope
(browser443, не игры/iSCSI). Реальные локальные TLS fixtures: >64KiB CDN PNG,
public WSS nonce-подтверждение, отказ неверного nonce/certificate/redirect.
Pinned DLL compiler/evaluator/checksum запускается под Wine, не в ядре Windows.
Первоначальный owner race и LSO gap — конкретные дефекты кода; что именно
вызывало отказ у владельца, без новой таблицы probe evidence не установлено.
Если остаются TIMESTAMP_OPTION_REQUIRED/LOCAL_SOCKET_OPTION_FAILED — чистый TS
на этой ОС/peer недоступен, подбираются другие варианты. Если transform committed,
но ORIGIN_SILENT/FREEZE_AFTER_N — нужен ISP trace/другая стратегия, не фиктивный PASS.
Нет гарантии, что собственные fake templates дадут ту же реакцию ТСПУ, что winws.
На Windows/провайдере владельца LOCAL2160p, Discord, PS5.1 и новый socket option
пока НЕ проверены. Цель relay RX≈0 для новых googlevideo flows ещё не доказана.

## CDN приёмка6df8a42 —24.10.2026
Владелец подтвердил www.youtube.com/discord.com >128KiB на catalog-c0aa53e48f222aca,
но0/4 CDN trials,3после ответа, googlevideo около7KiB и EOF. Это не proof обхода
CDN: публичный хост и sn-подгруппа могут фильтроваться по-разному.

Исправлен выбор: предыдущий код сначала выбирал следующий каталоговый prior,
и только затем искал proof; успешный другой plan оставался без приоритета.
Теперь same network/profile/hint-mode sustained YouTube proof выше transform-only,
с тем же точным ID; локальный неудачный CDN plan исключён из текущего sweep.
Discord proof не переносится на YouTube. Нельзя менять стратегию текущего trial.
Отдельный TrialAfter сохраняет >=3мин и не сбрасывается completeCDNQualification;
max1/live relay/3success/>64KiB/negative cache и отсутствие replay после origin
byte сохранены. Истёкший witness сам себя не продлевает простым выбором.

C: для fooling=ts badseq/badack — неактивные defaults, seq_delta0/1 ожидаем;
TS меняет существующий TSval на-600000. Для fooling=badseq применяются seq-10000
иack-66000 (либо явный override). MD5 — отдельная option19. TTL сохраняется из
original, если source recipe не задаёт TTL/autottl:128 само по себе НЕ ошибка.
Не добавляем скрытый badseq/TTL и не меняем подтверждённую стратегию.
Diag.10 выводит фактические ts_delta/ack_delta/MD5 option, не абсолютные значения.
Legacy old-sequence и проверка полного ClientHello до любого origin byte прежние.

A: причина конкретного EOF владельца пока НЕ доказана. Старое имя объединяло
remote EOF/reset и локальное закрытие. Теперь отдельные IOFailure, capture FIN/RST,
первый terminal packet/TTL, SYNACK TTL, genuine retransmits и payload counters.
Недавние LOCAL closes отдельно выдаёт browser.local.probes в пределах IPC budget,
не теряя их среди активных RELAY flows. CLIENT_PACKET — локальный upstream socket
(служба delegation), не доказательство закрытия браузером; REMOTE_PACKET может
быть подделан middlebox, разница TTL сама по себе не доказывает DPI. Собственный
Close помечается до закрытия socket, чтобы read-unblock не отравлял health-кеш.
Ни packet telemetry, ни ошибка EOF не разрешают replay/migration established TLS.
Нужно получить новый wire trace с ПК; remote7KiB могут быть TLS/маленьким HTTP
ответом/закрытием, не обязательно freeze после16KiB. Без MITM HTTP-код CDN не виден.

D:16обычных небольших PNG суммарно~21KiB, это ошибка размера выборки, не доказанный
отказ стратегии. Теперь ≤64 credential-free GET на одном TLS до >64KiB, прежние
4с/256KiB wire-cleartext budget;1312B asset воспроизводится fixture. При недостатке
данных PUBLIC_PROBE_BUDGET_NOT_SUSTAINED, не сетевой PASS/не bulk proof.
Пустой301 в generic probe больше не CONNECTION_FAILED. Для точного youtube.com
верифицированный canonical HTTPS→www.youtube.com redirect имеет только CONTROL
proof, не >64KiB/видео; follow на другой origin не выполняется. Корректный короткий
EOF этого control потока не отменяет собственную свежую контрольную проверку;
CDN trial по-прежнему провален при short EOF, отрицательный кеш сохраняется.
i.ytimg.com404 остаётся SHORT_RESPONSE, не локальный transport PASS.
