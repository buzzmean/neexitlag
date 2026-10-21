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
