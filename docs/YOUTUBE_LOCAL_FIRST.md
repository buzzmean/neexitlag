# YouTube LOCAL-first: реализация и приёмка

## Аудит старого relay-first (23.10.2026)

`git blame browser_transport.go:356–371`: явный `youtube_relay_first` появился в
79dfaaa8/ad3c69ce (18.10), family grouping —432899d3 (19.10). Он закрывал реальный
DIRECT_STALL на всех8 DNS адресах, повторные холодные таймауты, SNI-блокировку и
отсутствие подтверждённой рабочей LOCAL стратегии. Ранее relay-first ускорял
страницы по сравнению с повторными DIRECT ожиданиями; это не доказательство,
что провайдер требует relay. Владелец теперь подтвердил локальный обход другим
инструментом; предпочтение relay-first отменено новым решением.

`browserScopedLocal` требовал вручную выбранный LocalStrategyID, global Enabled и
verified baseline именно на этой сети. Автоматический native профиль их не имел;
ветка youtube_relay_first к тому же возвращалась до общих AUTO trials. Наличие
пресета youtube само по себе не включало его для браузера.

Собственный Windows TCP backend умеет bounded ClientHello buffering (32KiB/32
captured segments), split/multisplit, disorder и fake TLS с фиксированным TTL2–16.
Fake лежит целиком в старом TCP sequence space (old-sequence fooling), исключая
попадание fake в настоящий TLS stream даже при достаточном TTL. Handshake/ACK/
sequence/checksum проверяются. Payload fake свой, example.org, без заимствованных
файлов. Существующий preset —7 кандидатов, cutoff1 логического ClientHello.
Это исторический аудит исходного движка. С24.10 все64 TCP443 Flowseal rules
имеют native recipes (fakedsplit/hostfakesplit/SYN-data/TS/MD5 и др.); HTTP/QUIC
остаются unsupported, fake templates свои. Полная эквивалентность winws не заявляется. Новые browser-only кандидаты добавляют2 old-sequence
fake +3 настоящих сегмента с disorder, сохраняя потолок5 пакетов на входной.
Сохранённые пресеты и их версия не изменяются.

QUIC в localstrategy — ограниченный drop распознанных Initial для TCP fallback,
а native браузеры уже отклоняют выбранный UDP443 с bounded ICMP. Fake QUIC для
браузеров не добавляется: UDP payload PUBG не доказывает совместимость с QUIC/TLS.
При TCP fallback новая лестница пытается LOCAL перед RELAY. Игровой UDP/guard,
Vanguard/FACEIT, WinDivert filters и private iSCSI3260 не меняются.

## Контракт новой лестницы

Только browser AUTO, youtube-media, HTTPS443/IPv4. OpenAI/Anthropic/Telegram,
явные пользовательские правила, private destinations, DNS53, disabled и другие
приложения остаются вне неё. Обычный Google/VK остаётся DIRECT.

С 24.10.2026 неподтверждённая подгруппа идёт сразу через authenticated RELAY:
foreground не ждёт нерабочую DIRECT/LOCAL попытку. Подбор выполняется в фоне,
один worker на ПК, max12s на кандидат, не чаще cooldown подгруппы. Worker
проверяет живого owner/текущий профиль/сеть/отсутствие конфликтов, останавливается
при shutdown. Очередь наблюдаемых подгрупп ограничена64, кеш256, plans<=57,
score2048. Независимый тик15s продолжает подбор при долгом RELAY HTTP/2 потоке,
без необходимости нового клиентского handshake. Pause30s→1m→2m→4m→5m max,
после полного sweep продолжается следующий, без permanent RELAY/сброса к первому.

Подгруппы разделены: точный account/page/image host; CDN по sn-пулу googlevideo,
profile/network/port/hint-only остаются в ключе. Этот ключ НЕ меняет общий
DIRECT timeout cache.44 catalog plans +9 legacy кандидатов подключены к AUTO;
порядок по измеренным исходам этой сети/подгруппы. Истёкшие данные не являются
подтверждённым LOCAL. Signed catalog остаётся данными, не удалённой командой.

Фоновый GET с полной проверкой TLS/certificate, без cookies/credentials,
читает до128KiB+1 body. 403/404/204, TCP ACK, certificate/handshake bytes,
частичный body и заморозка16KiB НЕ проходят критерий >64KiB body. Для CDN с24.10 разрешена также ограниченная real-flow проба, когда та же стратегия
уже имеет реальный committed transform на любом YouTube-хосте той же сети/profile/
hint-mode, даже если публичный HTTP body не прошёл >64KiB. Это допуск пробы, не
успех CDN: max1 globally, живой получающий RELAY-сосед, >=3мин/subgroup. Проверки
обоих публичных origins >64KiB остаются более сильным сигналом, но не тупиком.
С24.10 (послеd28dda8) сильный running proof засчитывается без Close: полный
ServerHello, подтверждённый transform, последующая запись браузера, >=1MiB RX
после SH и >=1MiB в >=16 завершённых TLS type23 records, >=3с между первым и
последним наблюдённым application RX, свежий RX<=3с. Это sustained transport,
не доказательство playback/HTTP. Простой после быстрого burst не набирает3с.
TrialSuccesses/StrongPromotions увеличиваются один раз, scope сразу LOCAL10мин.
Глобальный слот освобождён для следующей НЕПОДТВЕРЖДЁННОЙ пробы; established
TLS остаётся на том же socket/egress. Max1 unproven/live RELAY sibling/3мин
на повторные пробы той же подгруппы сохранены. Старый путь из3 более слабых
закрытых sustained потоков (>64KiB каждый) оставлен как альтернатива.

Реальный RST/TLS/IO/WatchMidstream timeout после running promotion отзывает LOCAL
для будущих flows и учитывается PostPromotionFailures/FailedAfterReply; прошлый
сильный успех не стирается. Это НЕ дополнительная TrialFailure уже завершённой
пробы. Малый EOF с0<rx<64KiB, без известного failure, RST/ретрансмитов/alert/
invalid TLS нейтрален (RemoteSmallEOFInconclusive), не success и не bad score.
Зашифрованный Google idle/abandon/alert без MITM не различим: нейтральный исход
с3мин cooldown не утверждает исправность. EOF без ответа, reset и таймаут остаются
отрицательными. Холодный CDN больше не обязан ждать6мин+Close, но без подходящего
нового flow/сильного proof срок LOCAL не гарантируется.

Граница повтора прежняя: только полный ClientHello(+dummy CCS)<=64KiB до ЛЮБОГО
origin byte; early-data/application/partial hello и TLS alert не переигрываются.
LOCAL→RELAY при нулевом ответе укладывается в900ms first-flight timeout; после
ServerHello текущий TLS не переносится. Video12/other4 admission не меняется.

Для ECH однозначная DNS association привязывает LOCAL к точному IP, процессу и
порту с TTL; `TrustedOrigin=false`, segmentation по абсолютным offsets. Это
routing hint, не доказательство внутреннего SNI. Смешанные/неоднозначные/protected
подсказки не разрешают LOCAL. Никакого decrypt/MITM, отключения TLS или глобальной
настройки Chrome Secure DNS.

**Нельзя бесшовно перенести установившийся TLS на другой egress.** После любого
байта origin повтор запрещён, в том числе после alert. Реальный unanswered
outbound midstream stall помечает ступень для нового соединения; старый TLS не
переигрывается. Тишина после8–64KiB в течение4с — только hint для следующего
потока, текущий не закрывается этим признаком: без расшифровки невозможно отличить
«HTTP/2 закончил короткий ответ» от зависшей загрузки. Браузер может повторить
запрос сам; гарантировать отсутствие видимой ошибки после midstream failure
нельзя. Existing TCP остаётся pinned; браузеры не закрываются/не перезапускаются.

## Статус проверок

Контролируемые тесты проверяют настоящий TLS с SAN/certificate validation и
GET200/128KiB, три guarded CDN trial и warm LOCAL; холодный foreground сразу
RELAY. Transform и live-relay counters в portable fixture задаются seam: это
не WinDivert injection/ISP proof. Отдельно проверяются оба public prerequisites,
16KiB stall/truncation, short403/404, TTL/isolation/sweeps/ranking/cooldown/max1,
shutdown worker, mixed shared-IP hints до diagnostic truncation, no replay после
origin byte. Native Windows/PS5.1/реальный player/1080p LOCAL не проверены.

## Обновление и короткая приёмка владельца

Достаточно обновления клиента; для LOCAL-first deployment VPS не нужен. Из корня
репозитория, PowerShell администратора под SID владельца. Браузеры оставьте открытыми.
`Close-VeliqGuiForUpdate` закрывает только VELIQ (не браузеры), освобождая старые
recovery helpers, если они есть. На обычном ПК GUI затем запускайте обычным ярлыком,
на уже подтверждённом UAC-off ПК он остаётся full-admin как раньше. Служба после
обновления сама поднимает AUTO; никаких PAC/restart диалогов этот код не добавляет.

```powershell
git switch codex/work
git pull --ff-only origin codex/work
$m=Get-Content -Raw "$env:ProgramData\neexitlag\service.json" | ConvertFrom-Json
& .\tools\windows\build-e2e.ps1 -DriverDirectory $m.driver.directory
& .\tools\windows\Close-VeliqGuiForUpdate.ps1
& .\tools\windows\update-service.ps1
# В уже открытом Chrome/Edge: новая вкладка YouTube, 1080p2мин, перемотать2–3раза.
& .\tools\windows\diag-no-vpn.ps1 -Only browser,youtube
```

В `YouTube local ladder` сравнить `stage=LOCAL`/имя стратегии, committed transform,
RX и `transport_confirmed`, а также `stage=RELAY` RX до/после. При работающем LOCAL
**новый** video TLS payload через relay должен быть около0; CP/ticket/DNS DoH/echo
трафик и старые pinned relay-сессии в это обещание не входят. Проверить VK/обычный
Google DIRECT, OpenAI RELAY_REQUIRED, затем отключение оптимизации → обычный сайт.
`diag-result.txt` не коммитить. Неподтверждённый transform не выдавать за «точно не
применялся», если ledger потерян; `NOT_CONFIRMED` — отсутствие положительного proof.

Новые release/native Windows CI могут не стартовать из-за известного account
billing gate: `steps=[]` не является выполненными тестами. Доступная проверка
pinned DLL под Wine — компиляция/CRC, не WinDivertOpen/kernel/ISP acceptance.

P0 Hostkey отдельно готовит сервер и автоматически обновляет существующую службу
на ПК запуска (без переустановки/новой identity): docs/RELAY_HOSTKEY_P0.md. Не
запускайте эту remote mutation как обязательный шаг проверки LOCAL-first.

## Общий движок24.10: публичные протоколы Discord
Исправления raw delegation/LSO/native recipes относятся ко всем DPI группам.
Для cdn.discordapp.com — PNG GET по одному TLS до >64KiB; gateway.discord.gg —
публичный WSS Hello+3пары nonce ping/pong; discord.gg — точный HTTPS redirect.
Контрольные проверки явно OK_CONTROL_NOT_BULK, не bulk/page/login PASS.
Googlevideo критерии >64KiB/3success/live sibling не меняются, shared flags OFF.
Telegram/Meta/OpenAI/Anthropic RELAY_REQUIRED и VK/Яндекс DIRECT сохраняются.
