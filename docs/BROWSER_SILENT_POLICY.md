# Вариант A: тихий owned PAC Chrome и Edge

Решение владельца, 20.10.2026: независимый WFP guard отложен (вариант B в
BROWSER_NATIVE_GUARD.md). PAC остаётся обязательным crash-guard. Пользователь
не получает диалогов о перезапуске, программа не закрывает окна/процессы, не
открывает браузер, не вызывает `chrome://quit` / `edge://quit` и не восстанавливает
вкладки. Одна кнопка «Оптимизировать», автозапуск или fresh/repair используют
один тихий путь. Это не реализация native-only и не снятие PAC-защиты.

## Последовательность

1. GUI регистрирует установленные Chrome/Edge как DIRECT staging (не игру и не
   неизвестный процесс). Новая staging Edge теперь допустима при активном capture.
2. Под owner SID запускается прежний recovery helper **служебного EXE**, не браузер.
   `browser.prepare` / `edge.prepare` поднимает loopback CONNECT listener. GUI
   конфигурирует recovery тем же путём с теми же SID/ACL/guard-файлом.
3. `browser.silent.apply` / `edge.silent.apply` записывает обязательную inline PAC
   **службой**, только в `HKU\<installation owner SID>\Software\Policies\Google\Chrome`
   или `...\Microsoft\Edge`. HKCU SYSTEM, HKLM, WinINET/WinHTTP не меняются.
   Чужие ProxySettings/раздельные proxy values и HKLM policy отвергаются.
   Миграция старого Chrome PAC допускается только при совпадении прежнего expected
   PAC; различие порядка JSON-полей не делает нашу policy чужой.
4. Только после успешной записи профиль становится AUTO. Nonce, ReadyProfiles и
   подтверждённый HTTP(S) результат не подделываются. Разрешение настроенной lane
   отделено от старого `browser.confirm` / `edge.confirm`: эти методы остаются
   совместимыми и всё ещё требуют owner nonce. Любой реальный CONNECT заново
   проверяет PID, creation time, exe, SID, профиль и текущую политику.
5. GUI запрашивает `RefreshPolicy(false)` — обновление пользовательской GP. Это
   асинхронный запрос ОС, не API управления вкладками или соединениями браузера.
   Реестр и Chromium watcher могут обеспечить подхват; успешный RefreshPolicy
   **не доказывает** эффективный proxy. Машинную policy не обновляем.
6. После рестарта/boot служба сначала восстанавливает стабильный порт и owned PAC,
   затем подключает Control Plane и запускает обычную оптимизацию. Если owner
   hive ещё не загружен, не создаётся фиктивный HKU\SID; повтор выполняется раз в
   5 секунд после logon, пока служба работает. Ошибка чужой policy остаётся ошибкой,
   чужая policy не перезаписывается. Recovery GUI автоматически восстанавливает
   прежний owner-процесс при входе пользователя.

## Честные гарантии

* На продуктовых путях нет запуска/закрытия/kill/ожидания Chrome или Edge и
  запроса к пользователю. Идентификаторы существующих процессов не меняются.
* Запись policy, готовность listener и **подхват работающим браузером — разные факты**.
  `silent_activation.state=POLICY_PENDING_BROWSER_REFRESH` означает ожидание. GUI
  сообщает: «если браузер не перечитает policy, маршрут применится при следующем
  запуске браузера». Программа не просит сделать этот запуск и не выполняет его.
* Нет гарантированного времени подхвата конкретной версии Chrome/Edge: новое
  соединение может пользоваться ещё старой эффективной policy. До наблюдения
  нельзя заявлять работающую PAC-защиту этих окон или Internet PASS. Это честное
  ограничение принятого варианта A, а не независимый guard для ещё не прочитанной
  policy. Нативный capture и действующий owned PAC не отключаются.
* После подхвата PAC protected CONNECT не имеет DIRECT fallback, включая отказ
  службы; прежние recovery/owner-check/mandatory правила сохранены. Никакого MITM,
  отключения TLS/mTLS, подмены сертификатов или machine-wide VPN.
* `existing_tcp=PINNED_UNTIL_CLOSED`: HTTP/2 coalescing/старый сокет может продолжать
  прежний маршрут. Новая вкладка тоже может использовать старый сокет. Обновление
  страницы не обязано создать новое TCP; программа не сбрасывает существующие TCP.
* Чтобы policy уже присутствовала **до браузера при freeze**, клубный исходный
  образ должен содержать рабочую установку, профили и owned policy, а служба и GUI
  должны иметь штатный автозапуск. Если откат стирает и policy, и state, ни служба,
  ни пользовательский GUI не могут гарантировать победу в гонке с browser autostart.
  Машинная общая HKLM policy для всех пользователей здесь намеренно не вводится.

## Диагностика и bootstrap

`browser.status` / `edge.status` содержат `silent_activation`: configured_profiles,
applied_at, state, first_request_at, apply_to_request_ms, observed_requests,
existing_tcp. `PAC_REQUEST_OBSERVED` означает реальный owner-проверенный запрос
после применения policy в текущем процессе службы. Это не TLS/HTTP/page PASS и
не доказательство подхвата policy всеми профилями/окнами браузера. Метрика считается
по ограниченному текущему ledger; старые удалённые записи не реконструируются.

GUI показывает отдельную строку Chrome/Edge; diag .3 выводит эти поля. Fresh install
и `-RepairBrowserReadiness` проверяют listener + owned policy + AUTO/configured
profile. Ожидание browser uptake не вызывает падения `pac_readiness`. Реальный
отказ listener/записи policy/ACL/owner по-прежнему завершает операцию ошибкой.
Обновление службы повторно применяет journal и выводит безопасный статус, не
выдавая успех обновления за готовность сайтов. VPS/enroll/CA для ремонта не нужны.

## Эмпирическая проверка

В `TestSilentPACRunningChromeEdgePolicyUptake` (Windows CI, явный opt-in) настоящий
Chrome, затем Edge запускаются **до** записи реальной пользовательской platform
policy, с пустым временным профилем и `about:blank`. Ни одного `--proxy-*` аргумента.
После policy + user GP refresh CDP создаёт только тестовые вкладки; loopback proxy
проверяет owner exe/SID. Лог содержит фактическое `apply_to_request_ms`, число
запросов, неизменность PID/creation time, либо `POLICY_PENDING_BROWSER_REFRESH`.
Неуспешный подхват не рисуется нулевой задержкой/готовностью. Это HTTP ingress
measurement; существующий Edge HTTPS/mTLS acceptance остаётся отдельно.
Только изолированные браузеры теста завершаются в teardown, продукт этого не делает.

На Linux подготовлены и компилируются тесты, но **живой подхват Chrome/Edge в новой
итерации ещё не измерен**. Число миллисекунд появится в `silent-policy.log` workflow
Edge policy acceptance и в диагностике ПК; до этого результат NOT_MEASURED.

Владельцу достаточно обычного обновления службы/GUI и открытых браузеров; для
необязательного отчёта: `& .\tools\windows\diag-no-vpn.ps1 -Only browser,youtube`.
Закрывать браузеры для проверки не требуется. Действия с VPS не предусмотрены.
