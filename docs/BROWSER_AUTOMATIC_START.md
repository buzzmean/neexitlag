# Автоматическая маршрутизация уже открытых браузеров

## A. Конфигурация и подготовка

Новая установка и восстановление после freeze создают Chrome/Edge AUTO с
`browser_transport=native`. Это конфигурация, а не утверждение о доступности
сайта. Пока Optimize выключен, capture не запускается ради создания профиля.
При активном Optimize публикация native требует готового backend; перед
включением проверяется конфигурация capture и owner SID. PAC/nonce не являются
предпосылкой для native AUTO. Прежний legacy PAC-путь остаётся отдельным режимом.

Старые автоматически созданные DIRECT/pending без пользовательских правил,
стратегии и явного отказа мигрируют в native AUTO. ID, полный путь, identity,
правила и остальные приложения не пересоздаются. Явный выбор режима, отличного
от AUTO, сохраняет `browser_auto_disabled`; startup его не отменяет.

Каталог использует стандартные системные/пользовательские пути и read-only
App Paths / Clients StartMenuInternet в обеих registry views. Нет process scan,
чтения памяти или elevation GUI. Opera GX учитывается отдельно.

Brave, Яндекс, Opera/GX и Firefox обнаруживаются, но пока не объявлены
автоматически поддерживаемыми: нет их native Windows acceptance, Firefox имеет
другой policy/HTTP3 стек, а browser.exe не уникален для Яндекса. Не регистрируем
их как Chrome и не записываем им чужой PAC. Расширение поддержки требует
отдельного owner-bound native browser fixture, а не переименования exe.

## Ограничения проверки

В облачном окружении Linux нельзя подтвердить Chrome157/Edge live policy refresh,
WinDivert/SCM, пользовательский ISP или фактический срок запуска сайта. Ни
`RefreshPolicyEx`, ни registry presence не выдаются за live uptake. Нативный путь
не расшифровывает TLS и не меняет уже существующие TCP; они PINNED_UNTIL_CLOSED.

Native capture не является независимым kernel crash guard: при падении процесса
WinDivert освобождает handles. Сохранён прежний разрешённый владельцем availability
contract; новый WFP/driver guard здесь не добавляется. Во время активного capture
неоднозначный TLS/ECH не получает DIRECT-авторизацию. Нельзя заявлять безусловный
protected no-leak при смерти всего процесса без независимого guard.
