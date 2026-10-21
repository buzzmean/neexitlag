# Ранний отказ CI и локальная проверка deployment

21.10.2026 владелец предоставил аннотацию GitHub:
> The job was not started because recent account payments have failed or your spending limit needs to be increased. Please check the 'Billing & plans' section in your settings

Это account billing gate до выдачи runner. Аннотация не различает платёж и
spending limit. Исправление — Settings → Billing & plans владельцем аккаунта;
код/permissions/YAML не могут исправить оплату. Артефакты ещё не загружались.
`fresh-credential` был cancelled из-за matrix fail-fast; теперь он отключён,
чтобы после разблокировки оба layout давали независимый результат. Paths и
concurrency сокращают лишние прогоны, не обходят платёжный лимит.

## Linux: тот же реальный systemd/PostgreSQL сценарий локально

Тестовая Linux-машина с Go версии из go.mod, Git, Docker/BuildKit и доступом
к локальному Docker socket. Не запускать на VPS или машине с /etc/neexitlag.
Docker privileged нужен для systemd namespaces; использовать доверенный чистый
checkout на отдельной тестовой машине. Production credentials не требуются.

```bash
bash tests/deployment/run-local-systemd.sh fresh-credential
```

Второй layout:
```bash
bash tests/deployment/run-local-systemd.sh legacy
```

Runner собирает Debian 13 image с systemd/PostgreSQL/Python/OpenSSL, затем
release binaries из чистого checkout. Это тестовые инструменты, не новые
зависимости продукта. Только build image обращается к пакетным репозиториям;
runtime полностью offline (`--network none`), только loopback. Нет host network,
host cgroup bind, bind mounts репозитория/секретов/Docker socket/host /etc.
Root mutations выполняются исключительно внутри disposable filesystem.
Production IP назначается только lo внутри нового namespace; реальных обращений
к VPS нет. Guard требует systemd PID1, root-owned marker, отсутствие других
интерфейсов и default route. GitHub flags не подделываются.

Git archive исключает незакоммиченные файлы, .e2e и credentials. PKI создаётся
заново только для loopback теста, не используется CA пользователя. Cleanup удаляет
конкретный собственный контейнер и временную PKI даже при ошибке; не делает prune
и не останавливает чужие контейнеры. Публичный отчёт:
`.e2e/local-systemd-fresh-credential.json` / `.e2e/local-systemd-legacy.json`.
Не коммитить отчёты/PKI/private logs. Общий exit=0 обязателен; наличие JSON само
по себе PASS не доказывает. Отчёт содержит commit, systemd/PG версии, units,
schema/revocations, факт отзыва probe, rollback injection, без keys/codes/dumps.

Проверяется исходный isolated-upgrade.sh и, для fresh-credential,
isolated-activation.py: root-only CA/LoadCredential, реальные TLS/QUIC /3
TCP/UDP/renewal, 5000 отзывов, восстановление dump в отдельную БД, ошибочный
binary и rollback, замена leaf/failure/rollback, внешний failure после renewal,
успешный повтор schema 2, реальный renewal.service в sandbox и отсутствие активных
rollback timers. Это лаборатория, не Internet/VPS acceptance.

## Windows: локальный native эквивалент activation job

В обычном Windows PowerShell 5.1 из корня клона, с Git/Go/.NET8/Python:
```powershell
& .\tools\windows\test-activation-local.ps1
```

Запускает native activation/helper tests, C# tests/WPF build и fake-SSH fixture.
Не устанавливает службу, не вызывает bootstrap/enroll, не подключается к VPS.
Linux cross-build не заменяет эту проверку. Если Windows runtime недоступен,
результат остаётся неподтверждённым до этого прогона или разблокировки GitHub.
