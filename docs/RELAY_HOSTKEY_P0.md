# Перенос VPS-01 на Hostkey (P0)

Подготовка кода не означает выполненный deployment. Агент к production не подключался.
Control plane, activation, CA, клубный код и существующие клиентские сертификаты
остаются на 2.56.177.4. Новый сервер146.0.77.98 получает только relay, собственный
ключ и server-only сертификат от того же CA. Закрытый CA не покидает старый VPS.

## Как работает перенос

Установленный клиент получает relay адрес из подписанного конфига каждые30с;
браузерные protected правила закреплены за логическим VPS-01. Поэтому P0 меняет
его адрес, а не требует переустановки/нового кода. Изменение фабрики пула прекращает
приём новых потоков старыми сессиями. **Существующие TCP/UDP не мигрируют:** пока
соединение открыто, оно может продолжать пользоваться старым сервером.

Старый relay не останавливается; P0 сохраняет его как запасной для явного отката.
Автоматический выбор между разными relay ID, балансировка40ПК и per-relay video
пулы — отложенные A–E. P0 не обещает эти функции старому клиенту.

Root-only helper делает compare-and-publish relay row + config в одной транзакции
с новой монотонной версией. Ни миграции БД, ни enroll/revoke-tests, ни откат security
таблиц не выполняются. При ошибке адрес возвращается новым подписанным config.

На CP сохраняются private journals в `/var/lib/veliq-relay-fleet/` (не выводить их
содержимое: там heartbeat/undo). Независимый10min timer и boot recovery откатывают
незавершённую публикацию. Финализация отменяет timer. Будущий seed адрес CP меняется
owned systemd drop-in без рестарта CP; его TLS/LoadCredential пути сохраняются.
Heartbeat нового сервера имеет свежий токен. Старый relay продолжает получать
revocations/config даже после смены heartbeat token.

На новом сервере install идёт через существующий relay-install.sh; SHA256 каждого
артефакта сверяется. Неизвестная существующая установка/изменённые credentials
отвергаются. UFW не включается/не сбрасывается: только если уже active, добавляется
4433/udp. Внешний firewall провайдера должен пропускать этот порт.

После local health выполняется **с ПК** fresh-ticket QUIC mTLS /3, authenticated UDP
к существующему echo и HTTPS GET с полноценной проверкой публичного сертификата и
HTTP/DNS ответа. Только затем публикуется адрес и ждётся свежий heartbeat.
При провале новый сервер может остаться установленным, но registry откатывается;
его ключи/файлы не удаляются. Повтор безопасен только для owned установки.

## Владелец: PowerShell5.1, папка репозитория, тот же владелец установки

Нужны Git, Go, Python3, OpenSSH, активированная локальная identity для proof,
root доступ к обоим VPS (новый по `~/.ssh/id_ed25519`; старый может запросить пароль).
`accept-new` принимает только первый ключ; изменённый host key отвергается. При
первом подключении разумно предварительно сверить SSH fingerprint через консоль
провайдера. Ключи/пароли/коды не отправлять в чат.

Перенос (явная production-команда владельца; агент её не выполняет):
```powershell
& .\tools\windows\add-relay-vps.ps1 -Host 146.0.77.98 -Name hostkey-ams -PromotePrimary
```
Проверка установленного подписанного cache и пула после30–60с:
```powershell
& .\tools\windows\relay-status.ps1
```
Она честно показывает **назначенный адрес новых соединений**, active/draining
сессии. Legacy IPC не содержит удалённый endpoint каждого QUIC socket; нельзя
выдать этот вывод за доказательство, что все старые TCP уже покинули прежний VPS.

Read-only registry/heartbeat inspection (загружает временный helper в `/var/tmp`,
читает БД; не меняет registry, systemd, CA или journal):
```powershell
& .\tools\windows\add-relay-vps.ps1 -Host 146.0.77.98 -Name hostkey-ams -PromotePrimary -Inspect
```
Явный откат (существующие TCP также остаются pinned):
```powershell
& .\tools\windows\add-relay-vps.ps1 -Host 146.0.77.98 -Name hostkey-ams -PromotePrimary -Rollback
```

Server leaf выдаётся на90суток, ограничен сроком CA. Его автоматическое продление
на отдельном узле пока не установлено: это ограничение P0, не вечный сертификат.
CP/activation/echo остаются единой зависимостью старого VPS: продление/перенос до
12.11.2026 — отдельная обязательная работа владельца. 3ТБ/месяц — примерно9.3Мбит/с
непрерывно;1Гбит/с порт не гарантирует40 одновременных1080p и не снимает квоту.

## Проверки и границы

Unit certificate CSR/privilege restrictions, atomic signed PostgreSQL publication,
CAS/rollback/security nonce preservation и failure injection RPC проверяются локально.
PowerShell AST/PS7 и Linux/Windows build не заменяют реальный PS5.1/OpenSSH/systemd/
UFW/Hostkey proof. Эти production проверки выполняет только владелец указанной
командой. Не заявляем Hostkey/Internet/40ПК PASS до этого запуска.
