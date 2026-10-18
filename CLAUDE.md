# VELIQ / NeExitLag: правила
Язык: русский.
## Нельзя без явного разрешения
- Менять или деплоить боевой VPS, трогать /etc/neexitlag, сертификаты, учётные данные.
- Запускать tools/windows/setup-e2e.ps1 и install.ps1 (SSH, provision, enrollment).
- Мержить PR #8, force-push, переписывать историю.
- Печатать или коммитить секреты (ipc.token, identity.json, device.pem, токены).
- Отключать проверку TLS.
## Не переименовывать (совместимость)
Служба `neexitlag`, `neexitlag-service.exe`, %ProgramData%\neexitlag, %LOCALAPPDATA%\NeExitLag, pipe `neexitlag-v1`, ALPN `neexitlag/*`, Go module, mutex, HKCU Run. Бренд VELIQ только в UI и установщике.
## Инварианты
Protected-домены fail-closed, без DIRECT-утечек. Существующие TCP-потоки не переезжают между маршрутами. GUI без прав администратора.
## Порядок работы
Ветка cleanup/claude, маленькие коммиты, тесты до и после уборки. Файл удалять только с доказательством, что он не используется. После обновления службы сравнивать SHA256 собранного и установленного exe, а не верить надписи UPDATE COMPLETED.
## Окружение
Windows 10/11, Go 1.25, .NET 8, WinDivert 2.2.2-A. Клубные ПК откатываются при перезагрузке.
