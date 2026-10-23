# Каталог локального обхода: данные и пределы движка

23.10.2026. Принцип — снимать трафик с VPS подтверждённым локальным обходом,
а не считать любой TCP ACK доказательством работы сайта.

## Происхождение

`core/localstrategy/catalog.json`: все22 general*.bat Flowseal1.10.4,
commit176f56dc3f11ce5915c26abcdde903f93d835b64. Всего119 параметрических rules,
включая основные концепции blockcheck zapret/zapret2 и zapretgui. Для каждого —
источник, revision, путь, TCP80/TCP443/QUIC443, исходные параметры. MIT подтверждены;
у zapretgui лицензия найдена в docs/LICENSE. Notices в third_party/licenses/.

Обновление инвентаря выполняет offline `tools/catalog/import-strategies.py` из
проверенных исходных checkout, без исполнения BAT/shell/Lua/Python upstream.
Бинарники winws, fake-файлы и чужой packet-processing код не копируются.

## Что исполняется собственным движком

После добавления seqovl компиляция даёт13 уникальных TCP-планов из38 совместимых rules. Остальные81
rules явно unsupported, в том числе HTTP/QUIC и не реализованные SYN-data,
fakedsplit/hostfakesplit. Нельзя называть их рабочими, запускать внешнюю
программу или молча заменять другим методом. Каталог сохраняет их для аудита.

Новые собственные возможности: TLS1.3-shaped1250B fake с Google/Yandex SNI,
случайными полями, mixed split markers 1/midsld/sni/sniext и disorder, repeats1–16,
old-sequence badseq, split-seqovl256–1024 с собственным старым TLS-префиксом,
timestamp/MD5-option fooling и bounded autottl по TTL точного
SYN/ACK. Fake независимо от TTL/options всегда целиком в OLD sequence space;
никакой его байт не может попасть в TLS stream origin. TCP receiver отбрасывает ровно объявленный seqovl prefix; standalone fake всё ещё
целиком старый. Original stream/ACK/window/
options сохраняются, IP/TCP checksum пересчитываются. Нет MITM/подмены сертификатов.
Это независимая адаптация параметров, не побитовая копия поведения winws.

Повышенный предел разрешён только явному `browser-catalog` профилю выбранного
браузерного EXE, TCP80/443, cutoff1: максимум16 fake на логический ClientHello и
20 выходных пакетов на входной в сумме. Обычно11fake+3real=14. Обычные DIRECT+
профили и игры сохраняют старый предел5; античит policy не меняется.
На этом этапе новые планы ещё не включают браузерный подбор автоматически.
QUIC fake не заявлен; сохраняется разрешённый QUIC→TCP fallback.

## Проверено

Race tests: каталог/ограничения/случайный own TLS/повторы/cutoff/TTL/MD5/timestamp,
segmented1955B ClientHello и точное восстановление исходного TCP stream. Новые
fake проверены независимым WinDivertHelperCalcChecksums pinned DLL;70вариантов
всех service filters и30 private iSCSI evaluator cases PASS под Wine. Это не
WinDivertOpen/kernel/Windows/ISP proof. Фильтры, driver/pins не менялись.
Два Python inventory tests: все general/provenance и CRLF/caret/no execution PASS.

## Нерешённая граница фоновой проверки CDN

Корень googlevideo обычно403/404 и даёт меньше64KiB. Это не доказательство обхода
заморозки16–20KiB. Native capture без MITM не видит HTTP path и подписанный video
URL; заново открытый TLS имеет другие ключи и не может воспроизвести ciphertext
прикладного запроса. Поэтому строго фоновая передача >64KiB на том же CDN требует
подходящего тестового URL; короткий root не должен подтверждать LOCAL. Вопрос
владельцу: ограниченная проба реального нового LOCAL потока (возможен видимый
post-ServerHello stall) либо RELAY до появления полноценной фоновой пробы.
Без любого origin byte можно безопасно повторить только полный ClientHello;
после первого ответа существующий TLS остаётся pinned, replay запрещён.
