# CURRENT — состояние проекта (обновляется в конце каждой сессии)

Обновлено: 18.10.2026, Claude Code, ветка `claude/serene-cori-msra1u`, tip на момент записи `9bb43bd`.

## (a) Цель
Чистый Windows 10/11: установил, нажал «Оптимизировать» — заблокированные в РФ сервисы работают без VPN: ChatGPT, Claude, YouTube, Discord, Telegram, браузеры Chrome и Edge. Первый заказчик — клуб на 36 ПК с откатом при перезагрузке.

## (b) Состояние на 18.10.2026
- VPS 2.56.177.4 обновлён по PR #8 (17.10). Сертификат релея и control-plane продлён до 16.01.2027 (следующее продление ~2.01.2027).
- Потери UDP до VPS в покое: 0%.
- На ПК владельца установлена служба из ветки `claude/serene-cori-msra1u`. Проверка: `Get-FileHash bin\windows-e2e\neexitlag-service.exe` (SHA256 должен совпасть с `%ProgramData%\neexitlag\neexitlag-service.exe`). Последний известный хеш начинался с `5A7C3BCB` (сборка до `28841f2`; в status поле `build` покажет коммит новых сборок).
- Службу на ПК нужно пересобрать, чтобы получить коммиты `28841f2`..`9bb43bd` (отпечаток сборки, флаг YouTube).

## (c) Что работает на живом ПК
Telegram, Claude, ChatGPT, Chrome (профиль Chrome создан вручную).

## (d) Открыто
- YouTube: DIRECT без ответа (гипотеза: блокировка по SNI; проверяет секция `sni` в diag). Прозрачный повтор через релей реализован, но за флагом `youtube_via_relay` (по умолчанию выключен) до ответа владельца про канал и лимит трафика VPS.
- Edge без профиля и без политики. B3 сделан только для Chrome (автосоздание профиля). Edge — отдельная задача (своя политика Microsoft\Edge, PAC, recovery-процесс), ждёт «да».
- Discord медленно (headless Chrome: 15.4 с до DOM); причина не найдена.
- DIRECT+ для YouTube выключен по умолчанию; путь привязки к Chrome AUTO не разобран до конца (`ApplyPresetAsync` ставит DIRECT+ на весь профиль).
- Ложный вердикт 8443 в diag-no-vpn.ps1: исправлен в `578a45f`/`28841f2` (OK только при коде curl 0 и времени TLS > 0), на ПК нужен свежий скрипт.
- Не проверено на реальном ПК: все правки после `e436acf`; WPF-слой не компилировался (проверен только `dotnet test` на core-логике); ложные срабатывания порога застоя 2 с.

## (e) Кто чем владеет
| Область | Агент | Ветка | Статус |
|---|---|---|---|
| Всё перечисленное в (g)/журнале (служба, browserproxy, diag, профили) | Claude Code | `claude/serene-cori-msra1u` | сессия закончена 18.10.2026 |
| Edge-«полоса», Discord, DIRECT+ для YouTube | никто | — | ждёт решения владельца |
| PR #8, `codex/direct-plus-flowseal` | не трогать | — | только владелец |

Новые ветки: `claude/<тема>`, `codex/<тема>`, база — `claude/serene-cori-msra1u`.

## (f) Сборка и установка службы на ПК (по CLAUDE.md / docs/AGENT_RULES.md)
Только по просьбе владельца, на его ПК:
1. Закрыть Chrome и GUI («Выход» в трее). `git branch --show-current` = `claude/serene-cori-msra1u`, `git pull`.
2. Обычное окно PowerShell: `& .\tools\windows\build-e2e.ps1`, затем `& .\tools\windows\build-desktop.ps1`.
3. Elevated PowerShell под тем же SID владельца: `& .\tools\windows\update-service.ps1`.
4. Обычное окно: `& .\tools\windows\install-desktop.ps1`.
5. Сверить SHA256 `bin\windows-e2e\neexitlag-service.exe` и `%ProgramData%\neexitlag\neexitlag-service.exe`, `Get-Service neexitlag` = Running.

`powershell -File ...` ломает `$PSScriptRoot` (пустая строка в Join-Path): запускать только через `& .\tools\windows\скрипт.ps1`. Надпись «UPDATE COMPLETED» доказательством не является.
Запрещено: `update-vps.ps1`, `install.ps1`, `uninstall.ps1`, `setup-e2e.ps1`, `tools/vps-provision`, `tools/enroll`.

## (g) Как проверить без VPN
`& .\tools\windows\diag-no-vpn.ps1` (запускать из репозитория, не из старых копий; версия в шапке отчёта — `2026-10-18.5` или новее). Смотреть шапку (версия, коммит, `build` службы), разделы `sni`, `1b`, `1d` (`direct_stalls_total`, `stall_retries_via_relay`), `1f` (`probe_accounting`).
Доказательство: хеши службы совпадают и есть живая проверка на ПК владельца. Зелёный CI, лаборатория и Go-тесты доказательством работы на ПК не являются.
Управление флагом YouTube: `settings.set` с полным объектом Settings и `"youtube_via_relay": true` (по умолчанию выключен).

## (h) Журнал решений
| Дата | Решение | Почему |
|---|---|---|
| 18.10.2026 | C2 (relay-first claude.ai в браузере) не включать | C1 (кэш таймаутов) покрывает |
| 18.10.2026 | C3 (запасной транспорт) только дизайн, не реализовывать | решение владельца |
| 18.10.2026 | Дефолтные relay-first профили: Claude, ChatGPT/Codex, Telegram Desktop; fail-closed | живые тесты показали необходимость |
| 18.10.2026 | DNS/53 и локальные адреса не идут в релей | регрессия после новых профилей |
| 18.10.2026 | Застой DIRECT («TCP есть, ответа нет», порог 2 с) считается отказом; два за 60 с — relay-first на 10 мин | TCP-connect не доказывает работу DIRECT |
| 18.10.2026 | YouTube через relay за флагом `youtube_via_relay`, выключен | нагрузка на VPS (1080p ≈ 4–5 Мбит/с на зрителя), лимиты тарифа неизвестны |
| 18.10.2026 | B3 только для Chrome; Edge отдельно | у службы один PAC и один recovery-exe |
| 18.10.2026 | AUTO: FlapGuard против дребезга маршрутов | 24 переключения за 2,5 часа |

## (i) Последние коммиты ветки (от нового к старому)
- `9bb43bd` флаг `youtube_via_relay` (Go Settings + C# `ServiceSettings`), YouTube в релей только при включении; застои считаются всегда.
- `68e070a` стабилизация теста застоя (выбор записи по ID).
- `28841f2` отпечаток сборки службы (`build` в status), сквозной тест застоя, таблица SNI и шапка версий в diag.
- `a3ab95f` STATUS: застой DIRECT и исключение DNS/локальной сети.
- `91c9016` застой DIRECT-туннеля: счёт как отказ и прозрачный повтор через релей (stallConn, кэш, счётчики).
- `578a45f` diag: честный вердикт 8443, проверка YouTube по SNI.
- `b08de8f` DNS/53 и локальная сеть не уходят в релей у relay-first профилей.
- `9473d72` B3: поиск браузеров и автосоздание профиля Chrome (C# не компилировался).
- `d86de75` учёт проб по маршрутам (`quality.diagnostics.probe_accounting`).
- `c223af7` FlapGuard AUTO.
- `0fc9f97`, `3f75761` DirectTimeoutCache и его подключение к браузерному транспорту.
- `e436acf` дефолтные профили Claude/ChatGPT/Telegram.
- `ac5cf81` автопродление клиентских сертификатов; `673c866`, `f8dbf6b`, `d988b58`, `550a91f`, `df672c3`, `e075960` — надёжность запуска, rate limit, idle TCP, replay/fsync, сетевой снимок.
Подробности — `docs/STATUS.md`, `docs/BLOCKERS.md`.
