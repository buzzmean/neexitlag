# Выключение браузерной маршрутизации: cached PAC и DIRECT

Аудит регрессии 6adefe5, 21.10.2026. Проверка кода и изолированные тесты;
реальный Chrome/Edge владельца агентом не запускался. VPS не меняется.

## A: CONNECT после выключения

Корневая причина воспроизводима из кода: `openBrowser` делегировал PAC native
upstream по признаку профиля даже при `current.enabled=false`. Уже остановленный
backend отвергал OPEN. Дополнительно planner сохранял protected guard для
выключенного AUTO. Browser PAC остаётся в памяти независимо от удаления registry
value, поэтому этого достаточно для ERR_TUNNEL_CONNECTION_FAILED.

Теперь после настоящей PID/birth/path/SID авторизации browser socket выключенный
режим открывает обычный OS TCP DIRECT без Engine, WinDivert, relay, learning и
повтора application bytes. Транзакционный SYSTEM guard разрешает то же recovery
helper только для своего executable/profile. При повторном Optimize временный
DIRECT grant отзывается до Start capture. При активной оптимизации protected
политика остаётся прежней; это не её fail-open downgrade.

Изменение Optimize/Automatic не закрывает все proxy TCP: существующие соединения
закреплены. Бесполезный listener снимается после трёх минут без сокетов/новых
запросов, активный TCP не закрывается по этому таймеру. Включение умеет заново
привязать прежний порт. Snapshot содержит last_activity/active_connections.

## B: изменение реестра не равно подхвату политики

`browserProxyPolicyRemove` всегда удалял **значение** ProxySettings; сам ключ
Chrome/Edge не удаляется. Гипотеза об удалении ключа не подтвердилась. Причина,
видимая в коде: после удаления не отправлялось user Group Policy notification.
Предыдущий GUI RequestRefresh вызывался только при apply/native enable, не при
снятии политики. В Chromium Windows loader действительно используется
`RegisterGPNotification(..., false)` для user policy:
https://github.com/chromium/chromium/blob/main/components/policy/core/common/policy_loader_win.cc
(исходник прочитан 21.10.2026; это не эмпирический результат конкретного Chrome).

После реального Set/Delete теперь запрашивается RefreshPolicyEx(USER, RP_FORCE).
Повтор записи того же значения, отсутствие value и отказ менять чужое value не
создают notification flood. SYSTEM выбирает только token проверенного owner SID
из WTS, вызывает API на отдельном impersonated thread и возвращает identity;
неудачный Revert не возвращает такой thread планировщику. GUI дополнительно
запрашивает user refresh при переходе Optimizing в false в своей user session.
Нет HKLM policy, запуска gpupdate/браузера, chrome://quit или диалога.

Запрос асинхронный и bounded (один native GP call на процесс); результат доступен
как policy_refresh в browser.status/edge.status. REQUESTED означает только вызов
API, live_uptake=NOT_OBSERVED. Отказ/недоступный owner session не откатывает уже
выполненное удаление owned value и не считается ошибкой активационного кода.

Не гарантируется, что любая версия Chromium немедленно обновит chrome://policy
или сетевую конфигурацию. Для этого нужны реальные Windows измерения. Именно
поэтому A обязателен: пока используется cached PAC, ordinary browser работает
через DIRECT compatibility. После полного закрытия локальных endpoints обычный
сетевой отказ proxy позволяет DIRECT, явно включённый в текущий availability PAC.
Старый сторонний/legacy PAC без DIRECT не позволяет одновременно гарантировать
нулевые listeners и работу браузера, который совсем не перечитывает policy;
такое состояние нужно явно выявлять, а не выдавать за PASS.

## C: SCM host, watchdog и recovery имеют одно имя файла

`neexitlag-service.exe --chrome-recovery` — отдельный owner-GUI child, а не второй
SCM host. Его старый lifecycle завершался только по stdin EOF/GUI Dispose; поэтому
он мог пережить Stop-Service. `-browser-policy-watchdog` — ещё один отдельный
SYSTEM процесс, но он не открывает CONNECT ports. По одному имени EXE и порту
нельзя установить роль PID 20012: нужна проверка SCM ProcessId/flags/parent.

Новый recovery CLI read-only проверяет SCM раз в250мс и сам завершает свой
listener/process при STOP_PENDING/STOPPED/NOT_INSTALLED. UNKNOWN или старт не
считаются остановкой. GUI не перезапускает helper при выключенной оптимизации;
полный SCM stop остаётся штатным exit0. Watchdog не запускает Stopped service:
recovery применяется только к действительно Running/hung host; crash перезапускает
существующая SCM recovery policy. Штатный останов не подменяется control128.

`diagnose-browser-processes.ps1` ничего не меняет: SCM state/PID, ограниченные
process roles/parent/birth/path и loopback ports. CommandLine только разбирается
внутри, никогда не печатается. Start/restart/credentials не вызываются. Новый
status proxy показывает process_role/process_id. Watchdog может оставаться как
процесс без слушателей — это ожидаемо; сам SCM host и его recovery listeners
после штатного stop оставаться не должны. Реальный Stop-Service/SCM на ПК ещё
не проверен. После закрытия всех endpoints cached current PAC использует свой
явный DIRECT fallback; legacy PAC без DIRECT — ограничение из B.
