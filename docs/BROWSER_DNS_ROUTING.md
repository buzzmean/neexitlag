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
