/* VELIQ — логика кликабельного макета. Состояния: off, connecting, on, error. */
(function () {
  var root = document.documentElement;
  var win = document.getElementById('win');
  var $ = function (s, r) { return (r || document).querySelector(s); };
  var $$ = function (s, r) { return Array.prototype.slice.call((r || document).querySelectorAll(s)); };

  var S = { screen: 'home', state: 'off', error: null, warn: null, club: false, lang: 'ru', theme: 'dark' };
  var timer;

  /* ---------- Тексты (RU + EN для ключевых строк) ---------- */
  var T = {
    ru: {
      'nav.home': 'Главная', 'nav.services': 'Сервисы', 'nav.games': 'Игры', 'nav.diag': 'Диагностика', 'nav.settings': 'Настройки', 'nav.about': 'О программе',
      'st.off': 'Выключено', 'st.connecting': 'Подключение', 'st.on': 'Работает', 'st.error': 'Ошибка',
      'go.off': 'Оптимизировать', 'go.off.sub': 'выключено', 'go.connecting': 'Подключаем', 'go.connecting.sub': 'строим маршруты', 'go.on': 'Работает', 'go.error': 'Повторить',
      'home.off': 'Выключено', 'home.off.sub': 'Нажмите кнопку — VELIQ построит оптимальный маршрут для каждого соединения.',
      'home.connecting': 'Подключение', 'home.connecting.sub': 'Замеряем пути и подключаем браузер. Обычно это занимает пару секунд.',
      'home.on': 'Работает', 'home.on.sub': 'Маршруты построены: заблокированное открывается, игры идут по самому быстрому пути. Браузер перезапускать не нужно.',
      'home.error': 'Не работает', 'club.chip': 'Клубный ПК 14',
      'set.start': 'Запуск', 'set.autostart': 'Запускать вместе с Windows', 'set.autostart.d': 'Программа стартует свёрнутой в трей',
      'set.autoopt': 'Оптимизировать при запуске', 'set.autoopt.d': 'Не нужно нажимать кнопку после включения компьютера',
      'set.close': 'Закрытие окна сворачивает в трей', 'set.close.d': 'Маршрут продолжает работать',
      'set.server': 'Сервер', 'set.server.choice': 'Сервер для сервисов', 'set.server.d': '«Авто» выбирает самый быстрый и переключает при сбое',
      'set.look': 'Внешний вид', 'set.theme': 'Тема', 'set.theme.sys': 'Как в Windows', 'set.theme.light': 'Светлая', 'set.theme.dark': 'Тёмная', 'set.lang': 'Язык',
      'set.updates': 'Обновления', 'set.autoupd': 'Устанавливать обновления автоматически', 'set.autoupd.d': 'Ночью или при следующем запуске, без перезапуска браузера', 'set.check': 'Проверить обновления',
      'set.club': 'Компьютерный клуб', 'set.club.mode': 'Режим клуба', 'set.club.d': 'Простой экран для гостя, настройки под паролем, оптимизация при входе', 'set.club.on': 'Включить'
    },
    en: {
      'nav.home': 'Home', 'nav.services': 'Services', 'nav.games': 'Games', 'nav.diag': 'Diagnostics', 'nav.settings': 'Settings', 'nav.about': 'About',
      'st.off': 'Off', 'st.connecting': 'Connecting', 'st.on': 'Working', 'st.error': 'Error',
      'go.off': 'Optimize', 'go.off.sub': 'off', 'go.connecting': 'Connecting', 'go.connecting.sub': 'building routes', 'go.on': 'Working', 'go.error': 'Try again',
      'home.off': 'Off', 'home.off.sub': 'Press the button and VELIQ builds the best route for every connection.',
      'home.connecting': 'Connecting', 'home.connecting.sub': 'Checking servers and connecting the browser. Usually takes a couple of seconds.',
      'home.on': 'Working', 'home.on.sub': 'Routes are set: blocked services open, games take the fastest path. No browser restart needed.',
      'home.error': 'Not working', 'club.chip': 'Club PC 14',
      'set.start': 'Startup', 'set.autostart': 'Start with Windows', 'set.autostart.d': 'Starts minimized to the tray',
      'set.autoopt': 'Optimize on startup', 'set.autoopt.d': 'No need to press the button after the PC starts',
      'set.close': 'Closing the window keeps VELIQ in the tray', 'set.close.d': 'Routing keeps working',
      'set.server': 'Server', 'set.server.choice': 'Server for services', 'set.server.d': 'Auto picks the fastest and switches on failure',
      'set.look': 'Appearance', 'set.theme': 'Theme', 'set.theme.sys': 'Like Windows', 'set.theme.light': 'Light', 'set.theme.dark': 'Dark', 'set.lang': 'Language',
      'set.updates': 'Updates', 'set.autoupd': 'Install updates automatically', 'set.autoupd.d': 'At night or on next start, without restarting the browser', 'set.check': 'Check for updates',
      'set.club': 'Computer club', 'set.club.mode': 'Club mode', 'set.club.d': 'Simple guest screen, password-protected settings, optimizes at sign-in', 'set.club.on': 'Turn on'
    }
  };
  function t(k) { return (T[S.lang] && T[S.lang][k]) || T.ru[k] || k; }

  /* ---------- Ошибки: что случилось и что сделать (те же тексты — в brand/copy.md) ---------- */
  var ERR = {
    service: { short: 'служба не запущена', title: 'Служба VELIQ не запущена', text: 'Программа не может управлять трафиком без службы.', act: 'Запустить службу', act2: 'Диагностика' },
    relay: { short: 'нет связи с сервером', title: 'Нет связи с сервером', text: 'Серверы VELIQ не отвечают с этого подключения. Интернет без VELIQ работает.', act: 'Повторить', act2: 'Диагностика' },
    browser: { short: 'Chrome не подключился', title: 'Chrome не подключился', text: 'Приложения работают, а для сайтов в Chrome маршрут пока не построен. Обычно помогает повтор.', act: 'Повторить', act2: 'Подробнее' },
    identity: { short: 'ключ устройства не принят', title: 'Ключ устройства недействителен', text: 'Сервер не узнал этот компьютер — например после переустановки Windows.', act: 'Зарегистрировать заново', act2: 'Поддержка' }
  };
  var WARN = {
    update: { cls: 'banner--info', icon: 'download', title: 'Доступна версия 1.0.1', text: 'Обновление установится без перезапуска браузера и займёт около минуты.', act: 'Обновить' },
    fallback: { cls: 'banner--warn', icon: 'alert', title: 'Хельсинки не отвечает — работаем через Швецию', text: 'Сервисы открываются, пинг до них выше обычного. Вернём Хельсинки сами, когда он заработает.', act: 'Понятно' }
  };

  function banner(kind, data) {
    var cls = kind === 'error' ? '' : data.cls;
    var icon = kind === 'error' ? 'error' : data.icon;
    return '<div class="banner ' + cls + '" role="' + (kind === 'error' ? 'alert' : 'status') + '"><svg class="i"><use href="#i-' + icon + '"/></svg><div><strong>' + data.title + '</strong><p>' + data.text + '</p></div>' +
      '<div class="banner__actions">' + (data.act2 ? '<button class="btn btn--quiet btn--sm" data-banner-alt>' + data.act2 + '</button>' : '') + '<button class="btn btn--sm" data-banner-act>' + data.act + '</button></div></div>';
  }

  /* ---------- Отрисовка состояния ---------- */
  function render() {
    var st = S.state;
    win.setAttribute('data-state', st);
    var server = S.warn === 'fallback' ? 'Швеция' : 'Хельсинки';
    var ping = S.warn === 'fallback' ? 46 : 38;

    $$('[data-go]').forEach(function (go) {
      go.setAttribute('data-state', st);
      go.setAttribute('aria-pressed', String(st === 'on'));
      var l = $('[data-go-label]', go), sub = $('[data-go-sub]', go);
      if (st === 'off') { l.textContent = t('go.off'); sub.textContent = t('go.off.sub'); }
      if (st === 'connecting') { l.textContent = t('go.connecting'); sub.textContent = t('go.connecting.sub'); }
      if (st === 'on') { l.textContent = t('go.on'); sub.textContent = server + ', ' + ping + ' мс'; }
      if (st === 'error') { l.textContent = t('go.error'); sub.textContent = ERR[S.error].short; }
    });

    $('[data-home-title]').textContent = t('home.' + st);
    $('[data-home-sub]').textContent = st === 'error' ? 'Пока ошибка не исправлена, интернет работает как без VELIQ.' : t('home.' + st + '.sub');
    var slot = $('[data-banner-slot]');
    slot.innerHTML = st === 'error' ? banner('error', ERR[S.error]) : (S.warn && st === 'on' ? banner('warn', WARN[S.warn]) : '');

    var on = st === 'on';
    $('[data-m-server]').textContent = on ? server : '—';
    $('[data-m-server-sub]').textContent = on ? (server === 'Хельсинки' ? 'veliq-hel-01, выбран автоматически' : 'veliq-swe-01, запасной') : 'выбирается автоматически';
    $('[data-m-ping]').innerHTML = (on ? ping : '—') + ' <small>мс</small>';
    $('[data-m-traffic]').innerHTML = (on ? '1,24' : '—') + ' <small>ГБ</small>';
    $('[data-m-routes]').innerHTML = (on ? '12' : '—') + ' <small>активно</small>';
    $('[data-m-traffic-sub]').textContent = on ? '↓ 1,21 ГБ   ↑ 83 МБ' : 'получено и отправлено';
    $$('[data-via]').forEach(function (v) { v.textContent = server; });

    var statusCls = 'status status--' + (st === 'error' ? 'error' : st);
    var statusTxt = t('st.' + st);
    var rs = $('[data-rail-status]'); rs.className = statusCls; rs.lastChild.textContent = statusTxt;
    $('[data-rail-server]').textContent = on ? server + ', ' + ping + ' мс' : (st === 'error' ? ERR[S.error].title : '—');

    // клуб
    $('[data-club-title]').textContent = st === 'on' ? 'Работает' : st === 'connecting' ? 'Подключаем' : st === 'error' ? 'Не работает' : 'Выключено';
    $('[data-club-sub]').textContent = st === 'on' ? 'Маршруты построены: заблокированное открывается, игры идут по самому быстрому пути.'
      : st === 'error' ? 'Позовите администратора: ' + ERR[S.error].short + '. Остальной интернет работает.'
      : st === 'connecting' ? 'Пара секунд, и всё заработает.' : 'Нажмите кнопку, чтобы включить.';
    $('[data-club-server]').textContent = on ? server + ', ' + ping + ' мс' : '—';

    // трей и мини-окно
    var trayTheme = S.theme === 'light' ? 'light' : 'dark';
    var trayState = st === 'error' ? 'error' : st;
    $('[data-tray-img]').src = 'assets/img/tray/tray-' + trayState + '-' + trayTheme + '.svg';
    var trayLabel = 'VELIQ: ' + statusTxt.toLowerCase() + (on ? ', ' + server + ', ' + ping + ' мс' : '');
    $('[data-tray]').setAttribute('aria-label', trayLabel); $('[data-tray]').title = trayLabel;
    $('[data-taskbar]').classList.toggle('taskbar--light', S.theme === 'light');
    var fs = $('[data-fly-status]'); fs.className = statusCls; fs.lastChild.textContent = statusTxt;
    $('[data-fly-title]').textContent = st === 'error' ? ERR[S.error].title : t('home.' + st);
    $('[data-fly-sub]').textContent = on ? 'Маршруты построены' : st === 'connecting' ? 'Строим маршруты' : st === 'error' ? 'Откройте VELIQ, чтобы исправить' : 'Интернет работает как без VELIQ';
    $('[data-fly-switch]').checked = on || st === 'connecting';
    $('[data-fly-server]').textContent = on ? server : '—';
    $('[data-fly-ping]').textContent = on ? ping + ' мс' : '—';
    $('[data-fly-traffic]').textContent = on ? '1,24 ГБ' : '—';
    $('[data-ctx-toggle-label]').textContent = on || st === 'connecting' ? 'Выключить' : 'Оптимизировать';

    // экраны
    var screen = S.club ? 'club' : S.screen;
    $$('.screen').forEach(function (s) { s.classList.toggle('is-active', s.getAttribute('data-screen') === screen); });
    $$('[data-screen-btn]').forEach(function (b) { if (b.closest('.rail')) b.setAttribute('aria-current', b.getAttribute('data-screen-btn') === S.screen ? 'page' : 'false'); });
    win.classList.toggle('win--club', S.club);
    $$('[data-club-only]').forEach(function (e) { e.hidden = !S.club; });

    // панель макета
    $$('[data-go-screen]').forEach(function (b) { b.setAttribute('aria-current', String(!S.club && b.getAttribute('data-go-screen') === S.screen)); });
    $$('[data-proto-nav] [data-go-club]').forEach(function (b) { b.setAttribute('aria-current', String(S.club)); });
    $$('[data-set-state]').forEach(function (b) {
      var match = b.getAttribute('data-set-state') === st && (b.getAttribute('data-error') || null) === (st === 'error' ? S.error : null) && (b.getAttribute('data-warn') || null) === (st === 'on' ? S.warn : null);
      b.setAttribute('aria-current', String(match));
    });
    var names = { home: 'Главная', services: 'Сервисы', games: 'Игры', diag: 'Диагностика', settings: 'Настройки', about: 'О программе', club: 'Режим клуба' };
    $('[data-desk-label]').textContent = names[screen] + ' — ' + (st === 'error' ? ERR[S.error].short : statusTxt.toLowerCase()) + (S.warn && st === 'on' ? '. ' + WARN[S.warn].title : '');

    // тексты
    $$('[data-i18n]').forEach(function (e) { e.textContent = t(e.getAttribute('data-i18n')); });
    root.lang = S.lang;
    $$('[data-lang-set]').forEach(function (b) { b.setAttribute('aria-pressed', String(b.getAttribute('data-lang-set') === S.lang)); });
    $('[data-lang-select]').value = S.lang;

    // логотипы по теме
    $$('[data-wm]').forEach(function (img) { img.src = 'assets/img/veliq-wordmark-' + (S.theme === 'light' ? 'light' : 'dark') + '.svg'; });
    $$('[data-theme-set]').forEach(function (b) { b.setAttribute('aria-pressed', String(b.getAttribute('data-theme-set') === S.theme)); });

    try { history.replaceState(null, '', '#' + [screen, st, S.error || '', S.warn || ''].join('/')); } catch (e) {}
  }

  function setState(st, opts) {
    clearTimeout(timer);
    opts = opts || {};
    S.state = st; S.error = st === 'error' ? (opts.error || 'relay') : null; S.warn = st === 'on' ? (opts.warn || null) : null;
    render();
  }
  function toggleGo() {
    if (S.state === 'off' || S.state === 'error') {
      setState('connecting');
      timer = setTimeout(function () { setState('on'); toast('on'); }, 1600);
    } else setState('off');
  }

  /* ---------- Уведомления ---------- */
  var TOASTS = {
    on: { cls: 'toast--ok', icon: 'ok', title: 'VELIQ работает', text: 'Маршруты построены. Ближайший сервер: Хельсинки, 38 мс.' },
    lost: { cls: 'toast--error', icon: 'error', title: 'Связь с сервером потеряна', text: 'Перестраиваем маршрут через другой сервер. Это займёт пару секунд.' },
    back: { cls: 'toast--info', icon: 'route', title: 'Связь восстановлена', text: 'Снова через Хельсинки, 38 мс.' }
  };
  function toast(kind) {
    var d = TOASTS[kind]; var box = $('[data-toasts]');
    var el = document.createElement('div');
    el.className = 'toast ' + d.cls; el.setAttribute('role', 'status');
    el.innerHTML = '<svg class="toast__icon"><use href="#i-' + d.icon + '"/></svg><div><strong>' + d.title + '</strong><p>' + d.text + '</p></div><button class="btn btn--quiet" aria-label="Закрыть"><svg class="i i--sm"><use href="#i-close"/></svg></button>';
    el.querySelector('button').onclick = function () { el.remove(); };
    box.appendChild(el);
    setTimeout(function () { el.remove(); }, 5000);
  }

  /* ---------- Модальные окна ---------- */
  var lastFocus;
  function openModal(id) {
    lastFocus = document.activeElement;
    var m = $('[data-modal="' + id + '"]'); m.hidden = false;
    var f = m.querySelector('input, button'); if (f) f.focus();
  }
  function closeModals() { $$('[data-modal]').forEach(function (m) { m.hidden = true; }); if (lastFocus) lastFocus.focus(); }

  function closePopups() { $('[data-flyout]').classList.remove('is-open'); $('[data-ctx]').classList.remove('is-open'); $('[data-tray]').setAttribute('aria-expanded', 'false'); }

  /* ---------- Тема ---------- */
  function setTheme(th) {
    S.theme = th === 'system' ? (matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light') : th;
    root.setAttribute('data-theme', S.theme);
    render();
  }

  /* ---------- События ---------- */
  document.addEventListener('click', function (e) {
    var b;
    if ((b = e.target.closest('[data-go]'))) { if (S.club && S.state === 'on') return; toggleGo(); return; }
    if ((b = e.target.closest('[data-screen-btn]'))) { S.screen = b.getAttribute('data-screen-btn'); render(); return; }
    if ((b = e.target.closest('[data-go-screen]'))) { S.club = false; S.screen = b.getAttribute('data-go-screen'); render(); return; }
    if ((b = e.target.closest('[data-go-club]'))) { closeModals(); S.club = true; if (S.state === 'off') S.state = 'on'; render(); return; }
    if ((b = e.target.closest('[data-leave-club]'))) { closeModals(); S.club = false; S.screen = 'settings'; render(); return; }
    if ((b = e.target.closest('[data-set-state]'))) { setState(b.getAttribute('data-set-state'), { error: b.getAttribute('data-error'), warn: b.getAttribute('data-warn') }); return; }
    if ((b = e.target.closest('[data-banner-act]'))) {
      if (S.state === 'error') toggleGo(); else setState('on'); return;
    }
    if ((b = e.target.closest('[data-banner-alt]'))) { S.screen = 'diag'; render(); return; }
    if ((b = e.target.closest('[data-theme-set]'))) { setTheme(b.getAttribute('data-theme-set')); return; }
    if ((b = e.target.closest('[data-lang-set]'))) { S.lang = b.getAttribute('data-lang-set'); render(); return; }
    if ((b = e.target.closest('[data-modal-open]'))) { openModal(b.getAttribute('data-modal-open')); return; }
    if ((b = e.target.closest('[data-modal-close]'))) { closeModals(); return; }
    if (e.target.matches('.modal-backdrop')) { closeModals(); return; }
    if ((b = e.target.closest('[data-toast]'))) { toast(b.getAttribute('data-toast')); return; }
    if ((b = e.target.closest('[data-tray]')) || (b = e.target.closest('[data-open-flyout]'))) {
      var f = $('[data-flyout]'); var open = !f.classList.contains('is-open'); closePopups();
      f.classList.toggle('is-open', open); $('[data-tray]').setAttribute('aria-expanded', String(open)); return;
    }
    if ((b = e.target.closest('[data-open-ctx]'))) { var c = $('[data-ctx]'); var o = !c.classList.contains('is-open'); closePopups(); c.classList.toggle('is-open', o); return; }
    if ((b = e.target.closest('[data-fly-open]'))) { closePopups(); S.club = false; S.screen = 'home'; render(); win.scrollIntoView({ block: 'center' }); return; }
    if ((b = e.target.closest('[data-fly-close]')) || (b = e.target.closest('[data-ctx-close]'))) { closePopups(); return; }
    if ((b = e.target.closest('[data-ctx-toggle]'))) { closePopups(); toggleGo(); return; }
    if ((b = e.target.closest('[data-cmp]'))) { $$('[data-cmp]').forEach(function (x) { x.setAttribute('aria-pressed', String(x === b)); }); return; }
    if ((b = e.target.closest('.game-list button[aria-pressed]'))) { $$('.game-list button[aria-pressed]').forEach(function (x) { x.setAttribute('aria-pressed', String(x === b)); }); return; }
    if ((b = e.target.closest('.seg button')) && !b.hasAttribute('data-theme-set') && !b.hasAttribute('data-lang-set')) {
      $$('button', b.parentNode).forEach(function (x) { x.setAttribute('aria-pressed', String(x === b)); }); return;
    }
    if ((b = e.target.closest('[data-apply-route]'))) {
      var hel = $('.cmp--hel'); $$('[data-cmp]').forEach(function (x) { x.setAttribute('aria-pressed', String(x === hel)); });
      b.textContent = 'Закреплено'; b.disabled = true;
      $('.game-list button[aria-pressed="true"] .tag').className = 'tag tag--relay'; $('.game-list button[aria-pressed="true"] .tag').textContent = 'HEL';
      return;
    }
    if ((b = e.target.closest('[data-measure]'))) { var n = $('[data-measure-note]'); n.textContent = 'замеряем…'; setTimeout(function () { n.textContent = 'замер 20 пакетов, только что'; drawChart(); }, 1200); return; }
    if (!e.target.closest('.flyout, .ctxmenu')) closePopups();
  });
  $('[data-fly-switch]').addEventListener('change', function () { toggleGo(); });
  $('[data-lang-select]').addEventListener('change', function (e) { S.lang = e.target.value; render(); });
  $('[data-tray]').addEventListener('contextmenu', function (e) { e.preventDefault(); closePopups(); $('[data-ctx]').classList.add('is-open'); });
  document.addEventListener('keydown', function (e) { if (e.key === 'Escape') { closeModals(); closePopups(); } });

  function drawChart() {
    var jitter = function (base, amp) { var a = []; for (var i = 0; i < 24; i++) a.push(Math.round(base + (Math.random() - .5) * amp)); return a; };
    if (window.veliqPingChart) window.veliqPingChart($('[data-app-chart]'), { height: 180, series: [
      { id: 'direct', name: 'обычный', values: jitter(72, 8) },
      { id: 'swe', name: 'Швеция', values: jitter(48, 4) },
      { id: 'hel', name: 'Хельсинки', values: jitter(41, 3) }
    ] });
  }

  /* ---------- Старт: состояние из адреса (#экран/состояние/ошибка/предупреждение) ---------- */
  var parts = (location.hash || '').replace('#', '').split('/');
  if (parts[0]) { if (parts[0] === 'club') S.club = true; else S.screen = parts[0]; }
  if (parts[1]) { S.state = parts[1]; S.error = parts[2] || (parts[1] === 'error' ? 'relay' : null); S.warn = parts[3] || null; }
  root.setAttribute('data-theme', S.theme);
  drawChart();
  render();
})();
