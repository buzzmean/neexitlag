# VELIQ / NeExitLag — правила работы для Claude Code

Язык общения и комментариев к коммитам: русский.

## Что это
Windows-приложение для маршрутизации трафика по приложениям (игровой пинг + обход блокировок/деградации).
Служба `neexitlag` (SYSTEM) перехватывает пакеты через WinDivert v2 (только IPv4), WPF-клиент VELIQ управляет ею
по named pipe. Маршруты: DIRECT, DIRECT+ (локальные DPI-стратегии), RELAY / RELAY_REQUIRED (QUIC+mTLS до VPS), AUTO.
Первый заказчик: клуб на 36 ПК, ПК откатываются при перезагрузке. Бренд VELIQ, внутренние имена остались `neexitlag`.
Состояние и история: `docs/STATUS.md`, `docs/BLOCKERS.md`, `docs/handoff/` (если есть).

## НЕЛЬЗЯ без явного разрешения владельца
- Менять, деплоить или перезапускать боевой VPS; трогать `/etc/neexitlag`, сертификаты, учётные данные, БД.
- Запускать `tools/windows/setup-e2e.ps1`, `install.ps1`, `update-vps.ps1` (по умолчанию ходит по SSH на боевой VPS),
  `uninstall.ps1`, `tools/vps-provision`, `tools/enroll`.
- Мержить PR #8, делать force-push, переписывать историю.
- Печатать, логировать или коммитить секреты: `ipc.token`, `identity.json`, `device.pem`, ключи, токены, PEM.
- Отключать проверку TLS/сертификатов, ослаблять mTLS, replay-защиту, проверку подписей.
- Добавлять зависимости или менять версию WinDivert (бинарники закреплены по SHA256).

## НЕ переименовывать (совместимость с установленными машинами и VPS)
Служба SCM `neexitlag`; `neexitlag-service.exe`; `%ProgramData%\neexitlag`; `%LOCALAPPDATA%\NeExitLag`;
pipe `neexitlag-v1`; ALPN `neexitlag/*`; Go module `github.com/buzzmean/neexitlag`; mutex; HKCU Run; wire-идентичности.
Бренд VELIQ меняется только в UI, установщике и EXE (`VELIQ.exe`, alias `NeExitLag.exe` оставить).

## Инварианты (ломать нельзя)
- Protected-домены (OpenAI и др.) fail-closed: никакого DIRECT-fallback и утечек.
- Существующие TCP-потоки не переезжают между маршрутами; смена политики касается только новых потоков.
- GUI работает без прав администратора; служба и пользовательские файлы не смешиваются по чужим SID.
- Профили, device identity, owner SID, IPC token, сертификаты и конфиги переживают обновление.
- Сервис не перехватывает собственный трафик и трафик релея; DNS/53 не уходит в relay.
- Ничего не расшифровываем и не подменяем в HTTPS (без MITM).

## Структура
- `core/` — переносимое ядро: `protocol` (wire), `transport` (QUIC-мультиплекс, пул сессий), `security`
  (Ed25519, билеты, подписанный конфиг), `routing`, `flowpolicy`, `optimizer`, `quality`, `measurement`,
  `localstrategy` (DIRECT+), `browserproxy` (Chrome PAC + CONNECT + DNS-recovery), `ipc`, `profiles`, `controlclient`, `telemetry`.
- `platform/windows/` — всё Windows-специфичное: захват WinDivert, идентичность процессов, ACL, сетевой контекст.
- `services/client-service/` — служба (`controller.go` — IPC-обработчик, `browser_*.go`, `hot_attach.go`, `local_auto.go`).
- `services/relay/`, `services/control-plane/` — серверная часть (Linux/PostgreSQL).
- `apps/desktop/` (WPF, .NET 8) и `apps/desktop-core/` (клиентская логика, тестируется отдельно).
- `internal/` — фикстуры и пробы для приёмки; `tools/` — утилиты, лаборатории и PowerShell-скрипты (`tools/windows/`);
  `tests/` — интеграционные, desktop, deployment, windows-tools; `deploy/` — VPS, systemd, Grafana.
- Документы по темам лежат в `docs/` (ADR-001…005, IPC, PROTOCOL, SECURITY, WINDOWS_*, CHROME_*, DIRECT_PLUS*).

## Команды
Окружение: Go 1.25, .NET 8, Windows 10/11 x64 для живого захвата. Большая часть Go-кода собирается и тестируется на Linux;
пакеты с `//go:build windows` (в т.ч. `platform/windows`) требуют Windows.
```
make build            # go build ./...
make test             # go test -race ./...
make verify           # build + test + go vet ./...
go run ./tools/lab    # минимальная UDP/TCP лаборатория
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build ./...   # кросс-сборка Windows (как в CI)
python3 -m unittest discover -s tests/deployment -v
dotnet test tests/desktop/NeExitLag.Client.Tests.csproj --nologo
```
Как в CI на Windows (последовательно, иначе тесты конкурируют за порты):
`go test -p 1 ./core/... ./internal/browserfixture ./platform/windows ./services/client-service ./tests/integration`.
Фаззинг: см. `README.md` (`FuzzDecode`, `FuzzReadOpen`, `FuzzTicket`, `FuzzConfig`, `FuzzFrame`).

## Обновление установленной службы на ПК (только Windows, по просьбе владельца)
1. Закрыть Chrome и GUI («Выход» в трее).
2. `tools/windows/build-e2e.ps1` и `tools/windows/build-desktop.ps1`.
3. `tools/windows/update-service.ps1` — из elevated PowerShell под тем же SID владельца установки.
4. `tools/windows/install-desktop.ps1` (на пользователя).
5. Проверить SHA256: `bin\windows-e2e\neexitlag-service.exe` == `%ProgramData%\neexitlag\neexitlag-service.exe`
   и `Get-Service neexitlag` = Running. Надпись «UPDATE COMPLETED» доказательством НЕ является.
Диагностика: `Invoke-ServiceIPC` из `tools/windows/common.ps1` (`status`, `browser.status`, `capture.diagnostics`,
`quality.diagnostics`). В выводе не раскрывать токены и ключи; для `browser.status` использовать `ConvertTo-Json -Depth 12`.

## Как работать
- Ветка `cleanup/claude` от `codex/direct-plus-flowseal`; маленькие коммиты; тесты до и после каждого шага.
- Сначала план и подтверждение владельца, потом правки. Один вопрос за раз, только если без него нельзя.
- Файл или код удалять, только если доказано, что он не используется (grep по всему репозиторию, workflows, скриптам, документации).
  Всё, что нужно для совместимости (см. раздел «НЕ переименовывать»), не удалять.
- Не выдавать за подтверждённое то, что проверено только в CI или лаборатории: нужен реальный Windows/ISP-тест.
- Не обещать «снижение пинга» без измерений RTT/jitter/loss до реальных игровых серверов.
- Лишние выводы и большие логи не вставлять в контекст; читать файлы по конкретным путям, а не весь репозиторий.
