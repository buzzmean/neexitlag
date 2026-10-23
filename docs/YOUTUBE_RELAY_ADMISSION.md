# YouTube: SNI-selective DIRECT и отдельный video admission

## Причина 305e2fd

Shared8 позволял opaque Google занять7мест и оставить1под googlevideo.
Двухсекундная очередь уже существовала: она заканчивалась ошибкой, пока старые
opaque sockets держали permits. Native failure mapper скрывал эту локальную
причину за open_io_failed. TLS www.google.com на том жеIP не опровергает SNI
блокировку www.youtube.com. Данные владельца — доказательство именно этого
провайдера; универсальная блокировка Google не утверждается.

## Новый бюджет на один ПК

| Назначение | Активные | Ожидающие | Максимум ожидания | Idle после I/O |
|---|---:|---:|---:|---:|
| Известный googlevideo | 12 | 24 | 2с | 20с |
| Прочий relay Google/opaque | 4 | 8 | 1с | 10с |

Chrome и Edge используют общий бюджет ПК, но **разные video/other pools**.
Пример исходного расчёта12: две медиадорожки × два параллельных запроса =4,
столько же replacement/preload при перемотке =4, ещё4для разных CDN connections
и startup jitter. HTTP/2 может уложить много запросов в один TCP, поэтому это
запас соединений, не число сегментов и не подтверждённый оптимум плеера. Предел
выбран консервативно; нужен двухминутный ретест, а не обещание Internet PASS.

Максимум16Google relay-туннелей ПК, суммарная очередь32 — не384 бесконтрольных
соединения. Глобальный QUIC pool8×48/probe reserve/lease/mTLS не меняется.
До первого I/O idle3с; затем активность Read/Write продлевает срок. Idle closes
освобождают permits ровно один раз и учитываются отдельно; Read возвращает EOF
для scheduled local idle close, не generic relay session failure. Активный
обмен не закрывается таймером; устаревшие timer callbacks отсечены поколением.

При нормальной очереди ждём слот. При timeout/full protected origin остаётся
fail-closed с GOOGLEVIDEO_RELAY_ADMISSION_CAPACITY. Бесконечно держать очередь
нельзя. Отмены вкладок имеют отдельный canceled counter, не admission_failures.
Обычный AUTO origin при local/stream-capacity OPEN failure получает новый DIRECT
plan **до записи payload**. Реальный маршрут виден в native/PAC evidence как
browser_ordinary_capacity_direct. OpenAI/YouTube/Telegram/Anthropic, knownDoH,
protected DNS ambiguity, explicit RELAY_REQUIRED/BLOCK не могут так откатиться.
Неизвестный hidden domain сохраняет ранее утверждённое DIRECT-default ограничение.
Никакого replay/migration established TCP и никакого скрытого DIRECT под RELAY label.

## Нагрузка VPS (C)

VPS1vCPU/2GB: больший предел не увеличивает CPU/полосу сервера. На36ПК абсолютный
предел только Google —576TCP streams, обычное воспроизведение использует меньше.
Например,8Мбит/с видео ×36 =288Мбит/с payload плюс QUIC/TLS overhead; это расчёт,
не измеренная скорость. Одно ядро может стать пределом раньше порта1Гбит/с.
Пилот: сначала1ПК, затем2–3, затем постепенно36; при постоянной загрузке ядра
около85–90%, росте reject/drop или buffering не объявлять пилот успешным и не
поднимать лимиты вслепую. Агент production не проверял и не менял.

PowerShell владельца: следующие команды **только читают** VPS. Запускать во время
видео. Пароль вводится только в SSH; не вставлять в команду/чат.

```powershell
ssh root@2.56.177.4 'top -b -d 2 -n 3 -w 120'
```

```powershell
ssh root@2.56.177.4 'systemctl show neexitlag-relay.service -p ActiveState -p MainPID -p CPUUsageNSec -p MemoryCurrent -p TasksCurrent -p NRestarts'
```

```powershell
ssh root@2.56.177.4 'curl --fail --silent --show-error --max-time 3 http://127.0.0.1:9090/metrics'
```

Нужны neexitlag_active_sessions/active_flows, rx/tx_bytes_total, rejected_flows,
packet_drops/auth_failures. Два снимка bytes за60с показывают реальную payload
скорость; sessions не равны UDP sockets и не равны TCP connections.

```powershell
ssh root@2.56.177.4 'ss -s'
```

```powershell
ssh root@2.56.177.4 'ss -u -a -n sport = :4433'
```

9090 — существующий loopback metrics default; если порт отличается, не менять
конфигурацию/firewall и не выводить содержимое credentials/env. Попросить владельца
указать существующий metrics endpoint. Ни одна команда не перезапускает службы.

## Диагностика клиента

Версия diag-no-vpn.ps1 2026-10-23.2: admission_scope, video12/other4,
video/opaque waiters, failures vs canceled, total/max wait, idle_closed, ordinary
DIRECT fallback attempts, байты/минуту по домену/профилю. Raw opaque admission
failures могут завершиться рабочим DIRECT fallback и не равны video failures.
Counter попыток fallback не доказывает успешный DIRECT — проверять actual flow bytes.
Legacy googlevideo_shared_limit_active теперь video-only compatibility field;
google_admission_total_active показывает сумму, общего shared capacity больше нет.
Старые8и shared описания в предыдущих handoff — история, не текущая архитектура.

## Проверки и ограничения

Детерминированные fixtures: opaque4полностью занят, video12принимает потоки и
ещё4ждут завершения сегментов, без capacity failure; idle/preconnect reclaim,
повторное Close, отмена, hard cap, активность и stale timer generation; ordinary
Google hints vs shared-IP SNI failures; защищённые/explicit правила не понижаются;
ошибка OPEN до payload возвращает настоящий DIRECT plan обычному origin.
Это не проверка YouTube/player/Windows kernel или нагрузки36клубныхПК.
