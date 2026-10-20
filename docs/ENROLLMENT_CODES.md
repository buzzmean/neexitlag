# Коды активации VELIQ

## Схема и совместимость

Дополнительная миграция `002_activation_codes.sql`: activation_codes (случайный ID,
SHA256 высокоэнтропийного кода UNIQUE, kind club/personal, seat_limit, name,
expires_at, revoked); activation_slots (code_id + machine_hash PRIMARY KEY,
device_id, disabled). Слот остаётся отключённым до явного решения администратора,
повтор после freeze его не оживляет. Код хранится в БД только как hash; audit и
метрики используют случайный code_id, не hash/code. Админские devices API и
root-SSH fresh_machine_slots остаются совместимыми. Новые слоты привязаны к коду:
нельзя подставить hw hash, чтобы отозвать чужое SSH/legacy устройство.

`VQ-` + 24 случайных base32 символа (120 бит), группы по 4. Пример формата с
заполнителями, не рабочий ключ: VQ-XXXX-XXXX-XXXX-XXXX-XXXX-XXXX. Public release
не содержит ни кодов, ни administrator token, ни private CA. Обычное перечисление
админом возвращает ID/метаданные, никогда исходный код/hash. Создание возвращает
код один раз; tools/codes не пишет его в файл/лог, только явно в терминал владельца.

## API

* POST /api/v1/enroll: code + существующий CSR request (device_id, machine_hash,
  csr), без admin token. Проверка подписи CSR и Ed25519/subject/limits — существующая
  core/freshpc. Никакой выдачи private device key, CA key или echo/admin token.
* Admin bearer: /api/v1/codes POST create / GET list; /api/v1/codes/{id}/machines
  GET; /api/v1/codes/{id}/revoke POST; /api/v1/codes/{id}/machines/{hash}/disable POST.
  CLI читает admin token из файла, не argv; код вводится/возвращается без логов.
* Известные ошибки: ACTIVATION_INVALID, ACTIVATION_REVOKED, ACTIVATION_EXPIRED,
  ACTIVATION_LIMIT, MACHINE_DISABLED, ACTIVATION_RATE_LIMIT, ACTIVATION_UNAVAILABLE.
  Ошибки/логи не включают тело запроса, код, hash кода или CSR.

Транзакция берёт существующий advisory lock публикации и блокировку кода. Лимит
считается по слотам, не по накопленным revoked devices. Повтор того же CSR
идемпотентен. Новая identity того же (code,hardware) атомарно заменяет старую,
отзывает её и публикует полный signed config; DB/publication failure откатывает всё.
Новый сертификат формируется до commit, но выдаётся только после успешного commit.
Старые отозванные заявки не воскресают. Отзыв кода и отключение слота публикуются
той же транзакцией; existing auth/ticket/session renewal сохраняет revoked checks.

## Идентификатор машины

Старый SSH путь использовал SHA256 UUID v1. Сохраняется поддержка его существующих
слотов; не переводим их молча. Общий модуль hardware v2 принимает SMBIOS UUID,
серийник платы и MAC основной UP физической карты (выбор по route metric, стабильная
сортировка). MachineGuid не используется. Normalize + HMAC-SHA256 с публичной
installation/deployment солью (не случайной после каждой загрузки); на сервер
уходит только 64 hex. Соль поставляется вместе с доверенной публичной конфигурацией,
не загружается из неподтверждённого HTTPS. Fresh tooling получает тот же v2 helper
как опцию, legacy v1 остаётся для совместимости. Сырые значения не логируются.

HW ID — inventory, не удалённая аппаратная аттестация. Обладатель клубного кода
может скопировать его из образа или подделать hw hash: полностью закрыть это без
TPM/постоянного секрета вне freeze нельзя. Меры: SYSTEM/admin ACL, индивидуальные
ключи, лимит мест, IP/code/global rate limits, срок кода, отключение слотов и отзыв
кода. Администратор ПК может извлечь код; это честное ограничение, не обещание DRM.
Ротация hardware/MAC может создать новый слот, который админ должен освободить.

## Клиент и GUI

Служба принимает код по owner-authenticated IPC и хранит его в ProgramData только
SYSTEM/Administrators; GUI не читает файл обратно. Публичные CA/signing trust и
endpoint заранее установлены/проверены; TOFU и отключение TLS запрещены.
Ключ создаётся на ПК. При отсутствии действующей identity и наличии кода служба
в фоне делает CSR enroll, сохраняет certificate с прежними ACL и подключается;
временные сбои повторяются с bounded exponential backoff. Startup/IPC не блокируются.
Повтор потерянного ответа использует тот же CSR/ключ. Старые зарегистрированные
устройства без кода используют прежний ConnectRemote/renew путь.

GUI: «Введите код активации» / «Активировать», masked input; в Настройках только
«Код активации: активирован / Сменить». Нет кодов в exception/telemetry/diagnostics.
Никаких публичных обещаний про число серверов, игровых маршрутов или размер клуба.

## Server certificate renewal

Root systemd timer ежедневно проверяет leaf; за 10 суток выдаёт новый server leaf
на 30 суток тем же CA, с прежними SAN (в том числе VPS IP и 127.0.0.1). CA/ключи
identity не ротируются. Проверяет срок/CA signature/SAN, делает root-only постоянный
backup, ставит независимый rollback timer, атомарно заменяет CP/relay TLS files,
перезапускает CP -> relay. Проверка HTTPS8443 и настоящего QUIC4433 mTLS /3 через
краткоживущую локальную probe identity с обязательным отзывом. При любой ошибке
восстанавливает старые TLS files; не продлевает expired CA, не ослабляет TLS.
Диагностика показывает публичные CA/server NotAfter и предупреждение до истечения,
не ключи. Смена CA отдельно, в этот этап не входит.

## Выпуск и остановки

Сначала isolated PostgreSQL/race, client/GUI/renewal fake-clock tests, vet/build и
Windows compile. Деплой только при зелёных проверках, через существующий upgrade
с pg_dump, постоянным backup и rollback. Разрешение владельца распространяется на
2.56.177.4, не veliq-swe-01. Если SSH недоступен, выполняются только подготовка и
точные команды владельцу; тестовые club/personal коды до реального deployment
не создаются и не выдаются как работающие. На ПК не обещать PASS без owner retest.
