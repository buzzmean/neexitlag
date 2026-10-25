# Рабочее место VELIQ на любом ПК

Цель: сесть за любой ПК (клубный бездисковый или домашний) и за 10–15 минут получить всё, что нужно
Claude для работы через Remote Control: Git с входом в GitHub, клон `codex/work`, Go/.NET 8/Python,
собранные служба и GUI, установленную службу VELIQ (сервер Hostkey 146.0.77.98), Claude Code с разрешениями и
запущенный Remote Control. Скрипт: `tools/windows/dev-pc.ps1`.

## Один раз: зашифрованный набор `veliq-dev.bin`

На ПК, где SSH к VPS уже работает (ключ лежит в `~\.ssh`), в обычном PowerShell:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File C:\veliq\neexitlag\tools\windows\dev-pc-export.ps1
```

Скрипт спросит новый пароль (от 12 символов) и положит на рабочий стол `veliq-dev.bin`: SSH-ключ,
`known_hosts` и публичный CA VELIQ, зашифрованные AES-256 с ключом из пароля (PBKDF2-SHA256, 600 000
раундов, HMAC проверяется до расшифровки). Файл перетащить в чат проекта VELIQ — без пароля он
бесполезен. Пароль хранить у себя и никуда не писать. Ключ устройства VELIQ, коды активации и токены
в набор не попадают: каждый ПК активируется своим кодом.

## На новом ПК

1. Открыть claude.ai → проект VELIQ, скачать `veliq-dev.bin` в «Загрузки».
2. PowerShell **от администратора** под тем пользователем, который сидит за ПК, вставить:

```powershell
Set-ExecutionPolicy -Scope Process Bypass -Force; $r='C:\veliq\neexitlag'
if(!(Get-Command git -EA 0)){winget install -e --id Git.Git --silent --accept-package-agreements --accept-source-agreements; $env:Path+=";$env:ProgramFiles\Git\cmd"}
if(Test-Path "$r\.git"){git -C $r pull --ff-only}else{git clone -b codex/work https://github.com/buzzmean/neexitlag $r}
& "$r\tools\windows\dev-pc.ps1"
```

Что потребуется от человека, всё в первые минуты:
- войти в GitHub в окне браузера (Git Credential Manager), если ПК чистый;
- ввести пароль от `veliq-dev.bin`;
- в отдельном окне Remote Control войти в Claude (`/login`), если попросит;
- в конце ввести свой код активации в окне VELIQ и нажать «Оптимизировать». Тот же код на том же
  железе занимает то же место (`ACTIVATION_OWNER_RUNBOOK.md`, раздел 3).

Дальше скрипт работает сам: winget-установки, WinDivert с проверкой SHA256, `build-e2e`,
`build-desktop`, установка или обновление службы, проверка SHA256 установленной службы. Лог —
`C:\veliq\logs\dev-pc-*.log` (секреты не выводятся). Повторный запуск безопасен.

## Что ставится в Claude Code (решение владельца 25.10.2026)

`~\.claude\settings.json` дополняется (правила пользователя сохраняются): без вопросов `git`, `go`,
`dotnet`, `python`, `ssh root@146.0.77.98` и `ssh root@2.56.177.4`; force-push запрещён; режим
`acceptEdits`, если другой не задан. Запреты из `docs/AGENT_RULES.md` (изменения на VPS, сертификаты
и т.д. только после «да» владельца) остаются в силе: разрешение на ssh не разрешает менять сервер.

## Чего скрипт не делает

Не ходит по SSH с изменениями, не активирует службу за человека, не удаляет установленную службу
и `%ProgramData%\neexitlag`. Если служба стоит под другим пользователем Windows или папка осталась без
службы, скрипт предупреждает и идёт дальше. Ключи: `-NoKit`, `-SkipService`, `-SkipClaude`,
`-NoRemoteControl`, `-KitFile <путь>`.

## Проверено

`tests/windows-tools/dev-pc-checks.ps1` (шифрование набора, отказ при неверном пароле и подмене,
пин CA) и разбор всех скриптов — в PowerShell 7 на Linux. На живом Windows ПК
скрипт целиком ещё не прогонялся.
