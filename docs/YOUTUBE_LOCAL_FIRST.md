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
Полного Flowseal general нет: autottl, TCP MD5 signature, timestamp/badsum,
полноценный fakedsplit/hostfakesplit не реализованы. Называть текущие кандидаты
эквивалентом winws нельзя. Новые browser-only кандидаты добавляют2 old-sequence
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
DIRECT timeout cache.13 catalog plans +9 legacy кандидатов подключены к AUTO;
порядок по измеренным исходам этой сети/подгруппы. Истёкшие данные не являются
подтверждённым LOCAL. Signed catalog остаётся данными, не удалённой командой.

Фоновый GET с полной проверкой TLS/certificate, без cookies/credentials,
читает до128KiB+1 body. 403/404/204, TCP ACK, certificate/handshake bytes,
частичный body и заморозка16KiB НЕ проходят критерий >64KiB body. Для CDN
та же стратегия должна пройти ОБА www.youtube.com/ и публичную большую картинку
на i.ytimg.com; если картинка недоступна/короче порога, CDN остаётся RELAY.
Это косвенные prerequisites, не проверка подписанного video URL.

Затем max1 LOCAL CDN trial одновременно на весь ПК, только при открытом
получающем данные >64KiB RELAY video tunnel того же профиля (последний RX <5s),
не чаще1раза/3мин на подгруппу. Три последовательных trial-потока с ServerHello,
последующей записью браузера, >64KiB RX, committed transform и без наблюдаемого
сбоя до закрытия переводят новые потоки на LOCAL с TTL10мин. Очень быстрый
настоящий ответ не замедляется искусственной паузой. Долгий trial держит слот
до своего закрытия; это сознательное ограничение риска и может замедлить подбор.
Ошибки→следующий кандидат/RELAY новых потоков/backoff, счётчик failed_after_reply
показывает цену проб. Свежий CDN не станет LOCAL за2мин: только cooldown требует
не менее6мин на3пробы плюс время закрытия потоков. Длительный ручной тест не
нужен для unit/integration; реальную работу стратегий доказывает только ISP.

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
