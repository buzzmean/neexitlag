# DIRECT+ v2: Discord / YouTube, реальные измерения

Цель — native DIRECT+ для выбранного exe, без переноса приложения в VPN и без VPS payload там, где локальный вариант действительно работает. Это не подтверждение работы на вашем ISP: реальный Discord login/messages и воспроизведение YouTube проверяются на Windows с Happ OFF. Существующий доказанный RELAY остаётся контролем/fallback.

## Что изучено и что реализовано

Повторно получены актуальные HEAD 2026-10-14: Flowseal `eb16c1419057e36b8a36fbe7c959483450e6ac8b`, upstream zapret `c4d0b1990129dce2b28095277173ee0d7702d304` — совпали с предыдущим review в FLOWSEAL_ZAPRET_ANALYSIS.md. Вновь просмотрены general, все ALT, SIMPLE FAKE, FAKE TLS AUTO, list-general/list-google/ipset, upstream desync.c/readme.md и лицензии.

- general: TCP multisplit с sequence overlap, UDP/443 fake QUIC, отдельный Discord/STUN fake; дополнительные Discord TCP 2053/2083/2087/2096/8443.
- ALT: split/SNI positions, multisplit/multidisorder, fake/badseq/ts, иногда syndata/hostfakesplit/md5sig.
- SIMPLE FAKE: fake TLS + ts, Google hostfakesplit; AUTO TLS: fake modification/rnd/dupsid/SNI + badseq и multidisorder. Это AUTO fake-template, не route scoring.
- TTL/autottl/fooling нужны для неприёма fake сервером; повторять большие packet counts/overlap/game ranges вслепую нельзя.

Перенесены **концепции**, а не winws/C код/бинарные TLS или QUIC templates: конечные ordered chains, SNI-relative split, reverse-order delivery, ранний cutoff, fake old-sequence + отдельный явный TTL и экспериментальное сравнение реальных исходов. MIT notices сохранены в third_party/licenses; небольшие curated domain subsets основаны на указанных Flowseal списках. Никакого list downloader, /0, второго capture pipeline или другого драйвера.

В цепочке максимум три actions: `fake-tls` (только первым), `tcp-multisplit` (1–3 позиции, payload или SNI anchor), `tcp-disorder` (reverse после split). Не более пяти выходных IPv4 packets на один входной. Старые tcp-split, tcp-split-1/5 и udp-fragment-24 совместимы.

**Fake safety:** существующий TCP handle дополнительно наблюдает native inbound SYN/ACK и пропускает его без изменения. Нужны точный flow, ранее наблюдаемый SYN без данных, согласованный ACK, первый полный ClientHello и подходящий captured TTL. Собственный минимальный TLS ClientHello с `www.example.org` размещается целиком **перед** началом application sequence space: fake SEQ + fake length = client ISN + 1. Последний fake byte имеет SEQ ISN, уже потреблённый SYN; корректный TCP receiver отбрасывает весь segment как old/duplicate даже если TTL оказался достаточен до сервера. Real payload не меняется. ACK/server initial sequence проверяются, SYN/FIN/RST/URG не подделываются. Проверены wraparound и независимые IPv4/TCP checksums, включая Windows DLL helper. Это сильнее, чем полагаться только на «TTL должен где-то истечь».

Fake TTL — явный config field 2–16, строго ниже исходного TTL; кандидаты используют 2/5/8. GUI/capture показывают только фактически committed TTL. Нет autottl, глобального TTL или изменения timestamps. Если handshake/ClientHello/sequence неизвестны, весь план проходит DIRECT без fake. На send failure план не засчитывается успешным, оригинал reinject-ится по существующей fail-open политике. Успех reinjection **не** доказывает успех HTTPS.

Поддерживается полный ClientHello внутри одного IPv4 packet <=1500B. Нет TCP TLS reassembly, IPv6 desync, чтения ECH inner SNI, fake QUIC/STUN, MD5/ts/badsum fooling, sequence overlap, syndata, hostfakesplit или гарантии обхода любого DPI. Наблюдаемый outer SNI может участвовать в явном matching; скрытое ECH inner name не угадывается. Fragmented/oversized/offloaded ClientHello может дать честный `no transform`, включая современные браузеры/PQ TLS. Не отключайте TLS verification, Secure DNS или QUIC глобально ради PASS.

## Presets и GUI

1. «Приложения» → «Добавить приложение» → запущенный Discord или Chrome. Выберите **custom строку с полным exe path**, а не встроенное basename правило.
2. Выберите `Discord v2` / `YouTube v2`, отметьте или снимите «HTTP/3/QUIC → TCP fallback», нажмите «Применить пресет DIRECT+».
3. Это materializes 7 versioned scoped candidate profiles и выбирает явный default `fake5`: fake TLS TTL5 + SNI multisplit. Default — **не** измеренный победитель.
4. В «Локальные стратегии» нажмите «Загрузить»: имя/version/candidate, chain, positions и TTL видны в Advanced. Можно явно применить выделенный вариант к выбранному приложению. Редактор одиночного tcp-split не перезаписывает chain. Никаких 10 обязательных ручных полей.

Discord domains включают discord.com, discordapp.com/.net, discordcdn.com, discord.gg, discord.media, activities/status/gift и attachments storage; YouTube — youtube.com, googlevideo.com, ytimg.com, youtube-nocookie, youtubei/googleapis, ggpht и несколько точно scoped frontend/image domains. Apex и children заданы отдельно; неизвестные TCP SNI остаются DIRECT. Domain/IP sets и excludes из v1 сохраняются.

QUIC fallback — исключение **на уровне выбранного exe**, не domain attribution: только новый структурно валидный IPv4 QUIC v1 Initial на UDP destination 443 может включить подавление этого flow на максимум 3 секунды/8 packets. Не существующий UDP flow и не UDP на другом порту. Не все UDP443 — QUIC; malformed Initial проходит DIRECT. Голосовой/STUN/game UDP 19294–19344/50000+ не блокируется. Policy меняется только для новых flows. Выключение engine/conflict/service release возвращает native DIRECT. Counter `quic_fallback_packets` означает намеренно подавленные datagrams, а не доказательство того, что браузер уже выбрал HTTP/2. Это проверяется отдельным response protocol.

`selected_route` сохраняет исторический факт реально применённого DIRECT+; ACK/FIN/cutoff не стирают его. `direct_plus_packets/bytes` и `fake_tls_packets` коммитятся только после полного успешного send plan; неизменённые пакеты считаются direct. Per-flow `preset`, `chain`, `domain` (наблюдаемый Host/outer SNI), `fake_ttl`, relay bytes видны в GUI. Эти bytes не являются измеренной экономией VPS; app payload RELAY=0 не означает отсутствие control-plane/health-probe packets ко VPS.

Исправлен GUI «Нет данных»: truncated snapshot не отменяет положительные факты. Реально открытые relay socket pairs помечаются active, сортируются перед idle history, сохраняются для idle HTTP/2, а service подтверждает PID/creation time при недоступном вспомогательном process query. Нет подмены текущего route настроенным mode. Scoped capture `capture.diagnostics {destination:"IP:port",pid:...}` фильтруется **до** лимита кадра и нужен tester, не ослабляет прежний acceptance.

## Strategy Tester v2 (реальный exe, не перенос probe-exe результатов)

Закройте все процессы тестируемого exe, Happ/VPN OFF. PowerShell от администратора в корне checkout:

```powershell
.\tools\windows\strategy-tester-v2.ps1 `
  -AppPath 'C:\Program Files\Google\Chrome\Application\chrome.exe' `
  -Preset youtube -URL 'https://www.youtube.com/'
```

В GUI есть кнопка «Тестер v2 — реальное приложение / HTTPS». Скрипт лежит в windows-real-e2e-tools. Нужен сам Chrome/совместимый Chromium или Electron, Go не нужен. Если приложение не поддерживает isolated CDP/DevToolsActivePort (включая некоторые Discord Electron builds), автоматический trial честно fails как unsupported; это **не networking FAIL** и не доказательство неработающего пресета. Discord login/messages tester не автоматизирует.

Конечный набор: DIRECT TCP control, 7 candidates (split1, SNI multisplit, SNI disorder, fake TTL2/5 + multisplit, TTL8 + disorder, fake TTL5 only), RELAY TCP, повтор DIRECT и один успешный local near end. Максимум 11 trials, startup/navigation timeout, отдельный свежий browser process и временный browser profile на каждый trial. Никакого debugger attach к вашей обычной сессии. Закрывается только запущенное тестом process tree. Временный профиль удаляется best-effort (при оставшихся файловых блокировках выводится предупреждение); в proof сохраняются только response/route/counter/identity metadata, не HTML, cookie, TLS content или netlog.

Для сравнимости один hostname заранее resolves в public IPv4, который пинится только внутри isolated browser через host-resolver-rules. Это не system DNS rewrite, не замена реального browser DNS acceptance и не обещание работы на других CDN IP. Можно явно передать `-TargetIPv4` из проверенного вами DNS. SNI и hostname certificate validation сохраняются. DIRECT/RELAY controls отключают QUIC **только в диагностических процессах** для TCP baseline; DIRECT+ trial оставляет QUIC enabled и проверяет фактический TCP response. Никакого `--ignore-certificate-errors`.

PASS требует completed HTTPS Document main-frame response 2xx/3xx, secure certificate state, TCP protocol, точных exe/PID/creation/destination, route и actual modification counters; для DIRECT+ — нулевой relay payload. H3 не выдаётся за TCP. HTTP 403/certificate/timeout/нет transform = FAIL с сохранёнными diagnostics, даже если пакеты были отправлены. Не делается fallback к «JSON существует = PASS». Без реального captured target attempt (например, startup/CDP/DNS error) failure не записывается как сетевой результат. Сервер дополнительно требует собственный captured strategy proof для successful chain record.

AUTO использует cache по network fingerprint + config/strategy revision + полному app + **IP:port + protocol**, TTL <=15min, quarantine/ограниченные таблицы. DIRECT с измеренным здоровым исходом выигрывает; иначе лучший пригодный измеренный DIRECT+ по существующему score, затем реально измеренный authenticated RELAY. Общий ping VPS не становится website evidence; для preset-enabled AUTO неизвестный target остаётся DIRECT. Новые DNS/CDN IP требуют новых измерений. Измеренный TCP DIRECT+ также может разрешить QUIC fallback для того же app/IP:443; это не перенос TCP success на UDP relay. Domain-based routing до SYN и перенос существующей TCP connection между routes не добавлены. Нельзя переносить тест probe.exe на Chrome/Discord, или одного IP на весь YouTube. Смена сети/config инвалидирует cache; скрытые application-level ошибки спустя время без нового теста не угадываются из encrypted traffic.

## Обновление service + GUI без переустановки credentials

Скачайте artifacts одного успешного HEAD: `windows-real-e2e-tools.zip` и `NeExitLag-win-x64.zip` в корень checkout. Закройте GUI через «Выход» в трее; запомните его прежний путь. Выполните elevated PowerShell (замените только `$GuiInstalled` на существующий путь):

```powershell
$ErrorActionPreference='Stop'
git switch codex/direct-plus-flowseal
if($LASTEXITCODE){throw 'git switch failed'}
git pull --ff-only origin codex/direct-plus-flowseal
if($LASTEXITCODE){throw 'git pull failed'}
Expand-Archive .\windows-real-e2e-tools.zip .\bin\windows-e2e -Force
Expand-Archive .\NeExitLag-win-x64.zip .\.e2e\gui-update -Force
$GuiInstalled='C:\NeExitLag\NeExitLag.exe' # ваш существующий путь GUI
$replacement=(Resolve-Path .\bin\windows-e2e\neexitlag-service.exe).Path
$guiReplacement=(Resolve-Path .\.e2e\gui-update\NeExitLag.exe).Path
$svc=Get-CimInstance Win32_Service -Filter "Name='neexitlag'"
if(!$svc -or $svc.PathName -notmatch '^"([^"\r\n]+)"'){throw 'Cannot determine installed service executable'}
$installed=$Matches[1]
$old=if($svc.ProcessId -gt 0){Get-Process -Id $svc.ProcessId -ErrorAction SilentlyContinue}
Stop-Service neexitlag
(Get-Service neexitlag).WaitForStatus('Stopped',[TimeSpan]::FromSeconds(30))
if($old -and !$old.WaitForExit(30000)){throw 'Old service has not exited'}
Copy-Item $replacement $installed -Force
Copy-Item $guiReplacement $GuiInstalled -Force
Start-Service neexitlag
(Get-Service neexitlag).WaitForStatus('Running',[TimeSpan]::FromSeconds(30))
```

ProgramData credentials, state и ACL не удаляются/не ослабляются. GUI затем запустите обычным двойным кликом в Проводнике, без elevation. Не запускать winws/GoodbyeDPI параллельно. VPS redeploy не нужен.

## Самый простой REAL TEST, Happ OFF

A. Discord: добавьте реальный Discord.exe из running chooser, полностью закройте его между режимами. Сначала DIRECT (запишите реальный исход, не предполагайте FAIL). Потом preset Discord / MANUAL DIRECT+ `fake5`, запустите и проверьте настоящий login/start/messages. В GUI откройте fresh flows: preset/chain/TTL/modifications, relay bytes=0. Если не работает, сравните конечные варианты / tester если Electron поддерживает CDP. После неуспеха верните MANUAL VPS-01 как уже существующий fallback. Voice проверяется отдельно: fallback не блокирует non-443 UDP, но STUN DPI bypass не реализован.

B. Chrome/YouTube: закройте Chrome, добавьте его полный exe path, DIRECT baseline; затем YouTube preset, QUIC fallback включён, MANUAL DIRECT+. Откройте youtube.com и **реально воспроизведите видео**, подождите и перемотайте; один HTTPS 200 главной страницы этого не доказывает. Проверьте native public IP в том же браузере (например, собственный IP echo / api.ipify.org) — это ваш обычный адрес, не `2.56.177.4`. Сопоставьте видео TCP/UDP flows с PID/exe и zero per-flow relay bytes. Глобальные relay counters могут включать контрольный ChatGPT. QUIC fallback counter может быть 0, если браузер не пытался HTTP/3; не объявляйте QUIC PASS из одного UDP443.

C. ChatGPT: сохраните существующий MANUAL VPS-01 контроль. Если это тот же Chrome.exe, тестируйте режимы последовательно и верните его в RELAY после B; для одновременного контроля используйте другой browser exe с MANUAL VPS-01. Разные browser data directories одного exe не являются разными application identities. AUTO умеет только уже измеренные IP:port, не обещает произвольный per-domain YouTube DIRECT+/ChatGPT RELAY.

При FAIL сохраняйте `.e2e/strategy-v2/*-proof.json`, GUI/capture diagnostics, actual observed mode/identity/protocol. Не пересылайте browser profile/cookies/credentials. В CI реальные external Discord/YouTube/Happ-OFF PASS не фабрикуются.
