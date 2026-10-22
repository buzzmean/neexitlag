# DNS и непрозрачные browser flows

## A. Собственные DNS-сокеты Chrome/Edge

Для выбранного AUTO/native Chrome/Edge IPv4 UDP/TCP к порту 53 (включая
loopback) обслуживает служба. PID/birth/path/SID/profile/network перепроверяются
для каждой транзакции. Другие процессы, DNS Client Windows (`svchost.exe`), игры,
режим DIRECT и выключенная оптимизация сохраняют исходный путь.

Для YouTube/googlevideo/ytimg/ggpht, Telegram, Anthropic и OpenAI запрос идёт
по HTTPS DoH через существующий authenticated VPS transport. Серверу не нужны
новые программы или конфигурация. DNS/53 через VPS не пересылается; проверка
TLS DoH не отключается. Для остальных имён сохраняется исходный DNS-резолвер
и протокол запроса. Не меняются настройки DNS Windows/браузера.

Один IN-вопрос, не AXFR/IXFR/ANY; 8KiB максимальное сообщение; 16 одновременных
UDP-обработчиков и TCP DNS streams; 16 транзакций, без неограниченной очереди;
2с бюджет транзакции. Переполнение/отказ protected DoH даёт SERVFAIL, без DIRECT
DNS fallback. UDP учитывает EDNS (512..1232 байта), большие ответы возвращают
TC для повторения через TCP. TCP ограничен 64 запросами/5с простоя.

Ответ сохраняет A/AAAA/HTTPS/SVCB/DNSSEC; ECH-параметры не удаляются. Проверяются
ID/question/класс/CNAME; HTTP Age вычитается из TTL. Только фактически выданные
A/AAAA дают подсказки (network, profile, IP) с TTL не дольше ответа/5мин.
1024 адреса, до 8 имён на адрес. Совпавшие shared IP остаются неоднозначными;
наличие protected имени разрешает только ужесточить маршрут. Подсказка DNS не
выдаётся за доказанный внутренний ECH SNI. Обновление системного DNS-cache
не стирает browser-scoped ответы.

## Ограничения, которые нельзя скрывать

- Трафик Windows DNS Client нельзя надёжно приписать вызывающему браузеру
  только по PID сокета `svchost`. Он не перенаправляется целиком.
- HTTPS DoH шифрует вопрос и ответ. Без MITM их нельзя прочитать из TCP-потока.
- Скрытый домен на shared IP нельзя доказать по одной принадлежности адреса.
- Уже открытые TCP не переезжают; браузерные соединения и процессы не убиваются.
- IPv6 browser interception остаётся в прежнем режиме ограничений; эта DNS
  реализация обслуживает IPv4 сокеты. Нельзя заявлять полное владение DNS.

## Проверки

Проверяются TLS DoH с правильным/чужим CA, A/AAAA/HTTPS/SVCB, ID и CNAME/TTL,
UDP TC/TCP framing/cancellation, scoped hints и их ограничение, protected без
plaintext fallback, ordinary DNS и disabled/unselected без перехвата.
Windows packetWriter fixture проверяет построение ответа и отказ при saturation.
Все фильтры дополнительно проверены настоящим pinned DLL под Wine. Это не
проверка WinDivert kernel на Windows и не Internet/YouTube PASS.

## B. Secure DNS браузера

Выбран безопасный резерв B: **известные DoH endpoints — RELAY_REQUIRED**, без
DIRECT fallback, без намеренных RST и без подмены DNS-провайдера пользователя.
Тихо и надёжно определить automatic/secure-only режим работающего браузера
по зашифрованному сокету нельзя. Блокировка DoH не гарантирует откат на DNS/53
и способна сломать secure-only. Поэтому откат браузера НЕ объявляется доказанным.
DNS/53, если браузер его использует, обслуживается этапом A; существующие TCP
DoH не мигрируют. Новые известные DoH идут через relay в том же браузере.

Список — адреса/имена провайдеров из Chromium `net/dns/public/doh_provider_entry.cc`,
снимок 23.10.2026, SHA256 `cd0c73995ebe946154fb44ddbff4237bab378ac8329d52e5561f74b45ee689bf`,
и стандартные Cloudflare/NextDNS aliases. Только точные IP и hostname, TCP/443;
не весь CDN/ASN. IPv6 остаётся под существующим selected-browser IPv6 guard,
IPv6 relay support не заявляется. Чужие приложения и выключенная оптимизация
сохраняют обычный путь. Explicit BLOCK сохраняется; DIRECT правило не разрешает
обход этого известного resolver endpoint.

Custom DoH на неизвестном shared IP с ECH нельзя надёжно узнать без расшифровки
или настройки браузера. Ответы транзитного DoH не читаются, их TTL/имена не
выдумываются. Это ограничение явно остаётся; C наблюдает реальные DIRECT failures.

## C. Реальный отказ непрозрачного DIRECT

Старый кэш запоминал только точный (сеть, профиль, IP:порт), поэтому очередной
адрес Google обходил прошлый отказ. Первый TLS alert/даже заголовок в 5 байт
останавливал наблюдение как «ответ». Ошибка первого Write с n=0 не записывала
отказ. Callback публиковался после закрытия, что позволяло reconnect обогнать его.

Теперь TLS-flow ждёт пассивно наблюдаемый полный ServerHello (не TCP ACK).
Silent/incomplete handshake: бюджет 1200мс после успешной записи; ранний
EOF/RST/alert/неверный TLS или Write failure — точный failure_reason. Кэш
публикуется до нашего закрытия. Preconnect без данных, явное закрытие владельцем,
здоровый ServerHello и короткий polling timeout не дают ложного failure.
Запись/ответ не меняются и не воспроизводятся на другом egress.

Для ограниченного каталога Google IPv4 service ranges два разных DIRECT
TCP443 отказа за 10с включают relay-first на 2мин только для той же пары
(network, browser profile). TLS failures и реальные TCP-open failures учитываются;
один IP, UDP/53/5228, DNS/owner errors — не доказательство блокировки всей группы.
Каталог сам по себе не выбирает relay и не доказывает домен. Никакого
переписывания назначения/SNI по принадлежности сети. Максимум 64 scoped groups;
точные endpoint failures остаются отдельно (256 записей). VK/Яндекс не входят.

Opaque Google recovery делит прежние 8 admission slots с известным googlevideo,
32 waiters/2с; очереди не расширяются и DIRECT fallback не добавляется.
`googlevideo_active_tunnels` считает только известные video origins;
`opaque_google_active_tunnels` — неизвестные домены Google;
`googlevideo_shared_limit_active` — их общую загрузку. Ноль известного googlevideo
не доказывает отсутствие видео в opaque relay. Реальная приёмка должна показать
hostname evidence; без него результат по googlevideo — UNKNOWN, не PASS.

## D. DIRECT для обычных сайтов и диагностика

При известных DNS-подсказках VK/Яндекс остаются DIRECT. При неизвестном домене
они также не наследуют Google failure group. Неизбранные процессы и игры,
выключенная оптимизация и DNS чужого процесса не получают browser DNS handler.
Поздний callback закрываемого flow не обучает кэш после выключения или смены сети.

`diag-no-vpn.ps1 -Only browser,youtube`, версия 2026-10-23.1, отдельно показывает:
`browser DNS broker`, пути `AUTHENTICATED_RELAY_DOH`/`SYSTEM_DNS`, типы запросов,
количество выданных подсказок; до16 активных profile/network/IP/domain/expiry
подсказок и число пропущенных строк; реальную причину/состояние Google failures;
раздельные known-googlevideo/opaque-Google admissions и общий прежний предел8.
Старый сервис/отсутствующие поля помечаются UNINSTRUMENTED. Вывод не содержит
содержимого DNS/TLS, токенов или ключей. DNS/bytes не объявляются страницей PASS.
