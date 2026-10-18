# YouTube: ограниченный DIRECT → relay и проверка на ПК

Изменения относятся к Chrome/Edge site CONNECT, не к правилам Discord, standalone
ChatGPT/Codex, игровым профилям, mTLS/билетам/lease или VPS. Production не менялась.
Данные владельца: все восемь A-адресов YouTube не завершили прямой TCP connect;
VPS отдаёт HTTP 200. Это отчёт владельца, не новый Internet PASS разработчика.

## Решение и границы доказательств

- `youtube_via_relay=false`: максимум два разных IPv4-адреса из системного DNS, общий
  DIRECT budget 3 с, отдельный connect до 1300 мс. DNS входит в общий бюджет.
  Отказ/timeout/no-route до CONNECT 200 не отдаёт сразу 502: пробуется второй
  адрес, затем authenticated RELAY_REQUIRED VPS-01. Контекст отмены не повторяется.
- `youtube_via_relay=true`: обычный YouTube сразу relay-first. Аварийный fallback
  больше не зависит от флага. Configured DIRECT+, explicit domain rules и
  protected RELAY_REQUIRED не перезаписываются этой лестницей.
- При TLS-застое допускается только полный ClientHello до первого байта ответа,
  без TLS 1.3 early/application data. Сначала закрывается неответивший сокет:
  иначе `routing.Engine` возвращал `ErrFlowExists` на relay OPEN того же tuple.
  На время открытия следующего transport запись сериализована; память first flight
  ограничена 64 KiB. После любого ответа/установления TLS replay и смена egress запрещены.
- Ни TCP ACK, ни первый TLS байт/alert, ни HEAD не являются доказательством
  работы сайта. Очистка failure verdict требует background probe: валидный TLS
  с исходным SNI/сертификатом и настоящий GET с ограниченным чтением ответа.
  Это проверка origin, не доказательство видео/страницы/авторизации.

## Кэш

Ключ IP: `(network fingerprint, profile ID, domain, port, destination IP)`.
Повторные отказы одного IP отмечают только этот адрес; другой адрес не наследует
этот отказ. Два разных отказавших IP в окне 60 с дают доменный relay-first TTL
10 минут. `*.googlevideo.com` сворачивается в семейство `googlevideo.com`, сохраняя
разделение IP, сети, профиля и порта. Плохой адрес пропускается при выборе следующего.

После TTL один caller атомарно получает фоновую перепроверку; foreground остаётся
на relay до подтверждения DIRECT. Сокетов перепроверки не больше четырёх, очередь не
создаётся. Общий бюджет probe 3 с (внутренний TLS/GET read deadline строже); при
неопределённом результате retry через минуту. Кэш содержит максимум 256 записей,
включая адресные и доменные verdict. Это не бессрочная авторизация DIRECT.

`direct_timeout_cache` в **обоих** status IPC — общий кэш. Диагностика содержит
адрес/домен, profile/network, reason, число разных отказавших IP, оставшийся TTL,
next_direct_probe, recheck_in_progress, hit/stall/retry/recovered counters.
Адресный отказ может быть active, пока доменный verdict ещё не создан.

## Установленный googlevideo туннель и лимит

Если DIRECT уже отвечал, а после следующей реальной записи клиентских байт
10 с нет входящих данных, закрывается **этот** туннель и отмечается его домен/IP.
Простой без новой записи, первый ClientHello и безобидный read deadline не считаются
midstream-деградацией. HTTP/TLS payload не читается и не повторяется. Для нового
соединения браузер использует обновлённое решение; старый stream на relay не переносится.
`DIRECT_MIDSTREAM_STALL` виден в browser record/result/failures.

На весь ПК — четыре одновременных CONNECT relay-туннеля `*.googlevideo.com`, общие
для Chrome и Edge. Permit освобождается при ошибке OPEN и ровно один раз при Close.
Пятый ordinary AUTO fallback получает DIRECT с причиной `googlevideo_relay_limit`.
Если ISP не пропускает DIRECT, этот поток всё равно не заработает: ограничитель не
увеличивает пропускную способность. Explicit/protected required rules никогда не
понижаются до DIRECT; их отказ остаётся fail-closed.

`relay_usage` — реальные TCP payload bytes активных браузерных relay-сокетов:
по UTC-минуте, профилю и **реальному** hostname. Probe-трафик, DIRECT и QUIC/mTLS
framing не включены. Это не счётчик всего трафика ПК/провайдера. Хранится 10 минут,
максимум 1024 ключа; IPC показывает последние 64 строки, явно помечая truncation и
cardinality overflow bytes. Counters обновляются во время просмотра, не только после
закрытия; одну и ту же shared статистику Chrome/Edge складывать дважды нельзя.

## Обновление Windows (без VPS)

Закройте Chrome, Edge и VELIQ через «Выход» в трее. Сохраните локальные правки перед
сменой ветки; не используйте reset/clean. В PowerShell **с повышением под тем же SID**:

```powershell
& {
$ErrorActionPreference='Stop'
cd C:\neexitlag
if (@(git status --porcelain).Count) { throw 'Есть локальные правки: сохраните их перед обновлением' }
git fetch origin refs/heads/codex/work:refs/remotes/origin/codex/work
if ($LASTEXITCODE) { throw 'fetch failed' }
git switch codex/work
if ($LASTEXITCODE) { throw 'Не удалось переключить codex/work: не продолжайте' }
git pull --ff-only origin codex/work
if ($LASTEXITCODE) { throw 'pull failed' }
& .\tools\windows\build-e2e.ps1
& .\tools\windows\build-desktop.ps1
& .\tools\windows\update-service.ps1
& .\tools\windows\install-desktop.ps1
$built=(Get-FileHash .\bin\windows-e2e\neexitlag-service.exe).Hash
$installed=(Get-FileHash "$env:ProgramData\neexitlag\neexitlag-service.exe").Hash
if ($built -ne $installed) { throw 'SHA256 установленной службы не совпадает' }
if ((Get-Service neexitlag).Status -ne 'Running') { throw 'Служба не Running' }
}
```

Запустите GUI обычным пользователем, нажмите «Сайты Edge AUTO» (выбор строки больше
не обязателен при одном Edge-профиле). Проверьте `edge://policy` и `edge.status`.
Если активация не удалась, таблица должна остаться заполненной, а код/фаза — появиться
в GUI и `last_error`. «Отключить PAC Edge» заменяет **собственную** политику на guard
даже без профиля; обычные сайты DIRECT, protected закрыты, ключ не удаляется.

С VPN OFF и играющим видео выполните в обычном PowerShell:

```powershell
& .\tools\windows\diag-no-vpn.ps1 -Only youtube
& .\tools\windows\diag-no-vpn.ps1 -Only browser
```

Первый отчёт включает Chrome/Edge records с PID/exe, адресный кэш и minute bytes;
googlevideo берётся из реального record либо используется явно отмеченный
`redirector.googlevideo.com` как контроль. Второй запускает отдельные headless
профили **без proxy launch flags**, то есть через реальную политику браузеров.
Он не является полноценной приёмкой видео/сайта: DOM сам по себе не PASS.
`diag-result.txt` не коммитить. C#/WPF и native Windows должны проверяться CI/ПК;
Linux cross-build не заменяет runtime acceptance.

## QUIC/IPv6: только аудит, политику не меняли

По текущему коду:

1. `capture_windows.go` захватывает outbound IPv4 UDP. Chrome AUTO и Edge с
   `app_preset=edge-sites` при UDP/443 без достоверного домена получают
   RELAY_REQUIRED; никакого guessed YouTube IP → DIRECT разрешения нет.
   Edge **DIRECT** без активной site-полосы может пользоваться native DIRECT UDP/443.
2. `quic_fallback_packets` — подавленные QUIC Initial локальной стратегией
   (`quic.go`: максимум восемь за 3 с), принуждающие browser retry по TCP; это не
   число успешных HTTP/3 ответов. `fake_tls_packets` — локальные fake TLS packets,
   не QUIC/видео success. `direct_reply` — ответы управляемого UDP socket/adaptive
   path; outbound passthrough сам по себе не даёт такого reply evidence.
3. Production PAC/launch не выставляют `--disable-quic`/`QuicAllowed`. CONNECT
   описывает TCP path. Выводить из этого, что браузер обязательно запретил любой
   UDP/WebRTC, нельзя. Тестовые TCP-only helpers могут отдельно ставить disable-quic.
4. Основной redirect IPv4, **но отдельный IPv6 guard уже существует**:
   `capture_windows.go`, `ipv6_guard_windows.go`, `ipv6_guard.go`. Он блокирует
   attributable selected flows, которым нужен неподдерживаемый route/domain decision;
   unprofiled/DIRECT passthrough разрешён. Unknown owner и неподдерживаемый парсинг
   не доказывают domain-aware изоляцию. Это не IPv6 relay implementation и не
   основание заявлять, что предыдущий Edge точно использовал IPv6.
5. По имеющимся данным нельзя доказать, чем именно ранее открывался Edge: возможны
   DIRECT UDP/443 в DIRECT-профиле, IPv6 DIRECT, старое TCP-соединение/иной профиль
   или кэш. Нужна корреляция PID/birth/profile, адреса, протокола и времени.

Read-only команды Windows, пока играет видео:

```powershell
$browserProcesses=@(Get-Process chrome,msedge -ErrorAction SilentlyContinue)
$browserProcesses | Select-Object Id,ProcessName,Path,StartTime
foreach ($browserProcess in $browserProcesses) {
  Get-NetUDPEndpoint -OwningProcess $browserProcess.Id -ErrorAction SilentlyContinue |
    Select-Object OwningProcess,LocalAddress,LocalPort
  Get-NetTCPConnection -OwningProcess $browserProcess.Id -ErrorAction SilentlyContinue |
    Where-Object State -eq Established |
    Select-Object OwningProcess,LocalAddress,LocalPort,RemoteAddress,RemotePort,State
}
netstat -ano
```

UDP endpoint/netstat **не показывают удалённый UDP peer** и сами не доказывают
YouTube/UDP443/HTTP3. Следующий этап после отдельного «да»: bounded diagnostics
capture + Chromium net-export/CDP protocol (`h3`/`h2`), корреляция реального origin и
socket, отрицательная проверка protected traffic, сравнение relay bytes. Не
отправлять полный net-export с URL/cookies владельцу разработки; нужны только
ограниченные метаданные. Безопасный DIRECT QUIC для YouTube потребует достоверной
origin→UDP flow привязки, network/profile-scoped доказательства HTTP/3 и TTL,
обхода ECH/общих CDN без IP-угадывания. Просто разрешить весь UDP443 процесса нельзя.
Реализация этого этапа НЕ начата.

## Наблюдение за существующим VPS (выполняет владелец)

Не меняя конфигурации, службы и firewall, во время просмотра:

```bash
top -b -n 2 -d 1 | head -60
ss -s
ss -uanp
ss -tanp
systemctl status neexitlag-relay --no-pager
journalctl -u neexitlag-relay --since '-10 minutes' --no-pager -n 100
```

Смотреть CPU relay одного ядра, backlog/ошибки/число соединений, клиентские minute
bytes и существующие Prometheus metrics. Порт до 1 Гбит/с не доказывает доступную
производительность одного vCPU. 36 видео могут упереться в CPU/трафик; новый клиентский
лимит — локальный предохранитель, не подтверждённая ёмкость VPS. Серверных команд
в этой работе разработчик не выполнял, VPS deployment не нужен.
