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

Холодный поток: DIRECT → LOCAL → authenticated RELAY. Не более одной LOCAL
стратегии на клиентский handshake; следующий кандидат проверяется в фоне после
cooldown. DIRECT/LOCAL connect budget650ms; ожидание первого TLS ответа400/900ms.
Пока первый ответ не пришёл, повторяется только полный ClientHello (+dummy CCS),
максимум64KiB. TLS early data, application records и partial hello не повторяются.
RELAY шаг сохраняет существующие per-IP recovery, video12 admission и fail-closed.
Флаг старого relay-first предпочтения не запрещает последнюю ступень fallback.

Кэш: network fingerprint + profile + CDN/domain family + port + режим SNI/hint-only, TTL10min,
максимум256 entries. LOCAL подтверждается не TCP ACK/ServerHello: нужны фактический
committed transform, ServerHello, >=64KiB входящих и >=250ms наблюдения. Это только
sustained TLS transport, не HTTP/page/playback PASS. На cache hit новая сессия
сразу использует выученную ступень. Independent probes раз в минуту: real verified
TLS+bounded HTTP GET, без credentials/cookies, отдельный source tuple, shared
четыре browser probe slots; foreground не ждёт. Verified TLS + GET с CDN root
403/404/405 разрешает лишь минутную пробу LOCAL, а не media success: длительный
кэш требует подтверждения реальным потоком. Engine и learning для видимого
SNI/CONNECT и exact-IP ECH/DNS hints разделены, чтобы не заимствовать несовместимую
стратегию. После конечной лестницы кандидатов
— cooldown до TTL. Неподготовленный/conflicting/VPN backend пропускает LOCAL с
явным reason, не выдавая отправку обычных пакетов за desync.

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

Контролируемый TLS+HTTP fixture (самостоятельно созданный сертификат с проверкой
SAN, GET200/128KiB, DIRECT blackhole → LOCAL, повтор из cache) прошёл3 раза: холодный
~0.72–0.75s, warm~0.32–0.34s, RELAY payload0. Transform eligibility в этом portable
тесте задаётся backend seam: это не WinDivert injection/ISP proof. Также проверяются
finite first-flight, TLS alert/early-data no replay, TTL/scope/cache/guard и packet
checksum/old-sequence/output/cutoff. Windows/ISP2мин1080p ещё не проверены.

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
