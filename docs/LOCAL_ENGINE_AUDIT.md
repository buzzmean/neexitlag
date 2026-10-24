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

Реализованы свои fake TLS, split/disorder, old-sequence, TS/MD5 options, bounded
seqovl и autoTTL estimate. Текущий каталог НЕ эквивалентен всем Flowseal general:
из64Flowseal TCP443 rules (75с учётом остальных источников) есть hostfakesplit/fakedsplit/fakeddisorder/SYN-data, которые старый
compiler отвергал. Даже поддержанный mode использует свои fake templates и
обязательный old-sequence safety, отличающийся от чистого ts/md5sig winws. Нельзя
назвать перечисление параметров фактическим исполнением всех75правил.

## Диагностика первого этапа

`youtube_local_ladder.probes`: последние128 strategy×host результатов ВСЕХ групп,
wire application TX/RX (включая origin TLS), HTTP body bytes, commit, phase/detail,
SYN seen/handshake/server TTL и до20 реально отправленных packet headers:
length/TTL/protocol/TCP flags/relative sequence offset. Ни payload, ни абсолютные
sequence numbers, ни приватные TLS значения не экспортируются. Raw fake bytes
в wire application TX не входят: их видит отдельный список реальных sends.
ORIGIN_SILENT/RST/FREEZE_AFTER_N/SHORT_RESPONSE/NO_TRANSFORM/OK разделены;
короткий корректный ответ не назван freeze. OK — проверка собственного HTTPS GET,
не гарантия playback. Отсутствие ledger — неизвестное наблюдение, не успешный send.
Диагностика только наблюдает; routing/criteria этим этапом не ослаблены.

## Проверяемые риски, ещё не доказанная причина на ПК

При первом SYN OS TCP owner table может ещё не вернуть PID. Native delegation
с OpenDeferred удерживает такой SYN; LOCAL delegation старого кода такой защиты
не имеет и может попасть в self/DIRECT с закреплением решения на весь tuple.
Защита добавлена: LOCAL reservation требует тот же owner proof, что native; pending SYN удерживается до подтверждения владельца. Существующие close/retry тесты сохранены. Unit regression воспроизводит старое неверное решение при owner error; что именно это вызвало на ПК владельца, пока не доказано.
Также фактический commit может отсутствовать из-за handshake/TTL/sequence/guard,
а TLS/HTTP отказ возможен уже после правильно отправленного transform.

QUIC: нынешний native browser guard отклоняет selected UDP443 для TCP fallback;
это не fake QUIC и не подтверждённый LOCAL TCP. После fallback TCP проходит тот
же origin selector. IP/GEO группы не становятся LOCAL, VK/Яндекс остаются DIRECT.
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
и2160p LOCAL НЕ проверены. Все TCP443 режимы каталога пока НЕ реализованы.
