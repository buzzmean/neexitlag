# Русский нативный Windows-клиент: проверка

Среда: Linux x64, .NET SDK 8.0.425, Go 1.25.1, PostgreSQL 17.
Ветка: `codex/windows-native-russian-ui`, поверх `codex/windows-real-e2e` / PR #3.

| Проверка | Результат |
|---|---|
| `dotnet build` из корня | PASS, 0 warnings / 0 errors |
| `dotnet test --no-build` из корня | PASS, 36 tests, 0 skipped |
| Self-contained win-x64 single-file publish | PASS, `bin/release/NeExitLag.exe` |
| Проверка PE | x86-64, Windows GUI subsystem, asInvoker, PerMonitorV2 |
| `go test -race -count=1 -timeout 120s ./...` | PASS, с настоящим PostgreSQL |
| `go vet ./...` | PASS |
| Все Go-пакеты: Windows amd64 CGO0 build | PASS |
| PowerShell parser | PASS, build-desktop и desktop-smoke включены |
| Настоящее окно, DPI, трей, HKCU startup в этой среде | NOT RUN: здесь нет Windows desktop runtime |
| GUI + уже установленная Windows Service на машине оператора | NOT RUN здесь |

.NET tests включают настоящий обмен через NamedPipeServerStream с частичным
чтением кадров, проверку токена/ID/версии, отклонение oversized кадра, timeout
соединения и чтения, явную отмену, отсутствие службы, DTO создания процесса.
Модельные тесты покрывают реальные payload форматы команд, доступные/устаревшие
серверы, запрет неподдерживаемого BLOCK, сохранение метаданных профиля,
добавление существующего файла, защиту от дублирования профиля и повторного PID,
русские ресурсы/XAML, историю/счётчики, исчезновение службы и повторное подключение,
сохранение UI-выбора при неизменившихся snapshots. Фикстуры не являются E2E.

В Windows CI добавлен запуск настоящего опубликованного EXE с UI Automation:
окно NeExitLag, русская навигация, активация первого экземпляра вторым,
состояние отсутствующей службы и корректное завершение. Здесь smoke только
синтаксически проверен; результат Windows workflow нужно смотреть в GitHub.
CI также сохраняет опубликованный EXE как артефакт `NeExitLag-win-x64`.

Production GUI не содержит HTTP/browser transport. HTTP adapter старого демо
остаётся только в tests/desktop-bindings. Маршрутизация, WinDivert, relay,
process discovery и measurement принадлежат Go-службе. Добавлены совместимый
фильтр IPC processes и исправление private-CA health-check; протокол IPC остаётся v1.

Оператор ранее доказал ручной simultaneous DIRECT/RELAY E2E. Это отдельное
свидетельство маршрутизации, а не runtime-проверка нового WPF-окна.

[Сборка одной командой и запуск](../apps/desktop/README.md).
