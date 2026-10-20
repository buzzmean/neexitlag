/* VELIQ website — без зависимостей. */
(function () {
  var root = document.documentElement;

  /* Тема: системная по умолчанию, выбор запоминается в браузере */
  try { var saved = localStorage.getItem('veliq-theme'); if (saved) root.setAttribute('data-theme', saved); } catch (e) {}
  function currentTheme() {
    var t = root.getAttribute('data-theme');
    if (t) return t;
    return window.matchMedia && matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
  }
  document.addEventListener('click', function (e) {
    var b = e.target.closest('[data-theme-toggle]');
    if (!b) return;
    var next = currentTheme() === 'dark' ? 'light' : 'dark';
    root.setAttribute('data-theme', next);
    b.setAttribute('aria-label', next === 'dark' ? 'Включить светлую тему' : 'Включить тёмную тему');
    try { localStorage.setItem('veliq-theme', next); } catch (e2) {}
  });

  /* Мобильное меню */
  var menuBtn = document.querySelector('[data-menu]');
  var nav = document.getElementById('nav');
  if (menuBtn && nav) {
    menuBtn.addEventListener('click', function () {
      var open = nav.classList.toggle('is-open');
      menuBtn.setAttribute('aria-expanded', String(open));
    });
    document.addEventListener('keydown', function (e) {
      if (e.key === 'Escape' && nav.classList.contains('is-open')) { nav.classList.remove('is-open'); menuBtn.setAttribute('aria-expanded', 'false'); menuBtn.focus(); }
    });
  }

  /* Линия под шапкой после прокрутки */
  var top = document.querySelector('.top');
  if (top) {
    var onScroll = function () { top.classList.toggle('is-scrolled', window.scrollY > 8); };
    window.addEventListener('scroll', onScroll, { passive: true }); onScroll();
  }

  /* Демо большой кнопки на главной */
  var demo = document.querySelector('[data-demo]');
  if (demo) {
    var go = demo.querySelector('.go');
    var label = go.querySelector('[data-go-label]');
    var sub = go.querySelector('[data-go-sub]');
    var status = demo.querySelector('[data-demo-status]');
    var live = demo.querySelector('[data-demo-live]');
    var timer;
    function set(state) {
      demo.setAttribute('data-state', state);
      go.setAttribute('data-state', state);
      if (state === 'off') {
        label.textContent = 'Оптимизировать'; sub.textContent = 'выключено';
        status.className = 'status status--off'; status.lastChild.textContent = 'Выключено';
        go.setAttribute('aria-pressed', 'false');
      } else if (state === 'connecting') {
        label.textContent = 'Подключаем'; sub.textContent = 'строим маршруты';
        status.className = 'status status--connecting'; status.lastChild.textContent = 'Подключение';
      } else {
        label.textContent = 'Работает'; sub.textContent = 'маршруты построены';
        status.className = 'status status--on'; status.lastChild.textContent = 'Работает';
        go.setAttribute('aria-pressed', 'true');
      }
      demo.querySelectorAll('[data-on]').forEach(function (v) { v.textContent = state === 'on' ? v.getAttribute('data-on') : '—'; });
      if (live) live.textContent = status.lastChild.textContent;
    }
    go.addEventListener('click', function () {
      clearTimeout(timer);
      if (demo.getAttribute('data-state') === 'off') { set('connecting'); timer = setTimeout(function () { set('on'); }, 1500); }
      else set('off');
    });
  }

  /* График пинга */
  document.querySelectorAll('[data-ping-chart]').forEach(function (box) {
    if (!window.veliqPingChart) return;
    var data = JSON.parse(box.getAttribute('data-ping-chart'));
    window.veliqPingChart(box, data);
  });

  /* Калькулятор для клубов */
  var calc = document.querySelector('[data-calc]');
  if (calc) {
    var fmt = new Intl.NumberFormat('ru-RU', { maximumFractionDigits: 0 });
    var inputs = calc.querySelectorAll('[data-k]');
    function v(k) { var el = calc.querySelector('[data-k="' + k + '"].input'); return Math.max(0, parseFloat(el.value.replace(',', '.')) || 0); }
    function recalc() {
      var pcs = v('pcs'), hours = v('hours'), load = v('load') / 100, price = v('price'), lost = v('lost') / 100, cost = v('cost');
      var paidHours = pcs * hours * 30 * load;
      var revenue = paidHours * price;
      var back = revenue * lost;
      var veliq = pcs * cost;
      var net = back - veliq;
      calc.querySelector('[data-out="net"]').textContent = (net < 0 ? '−' : '') + fmt.format(Math.abs(net)) + ' ₽';
      calc.querySelector('[data-out="hours"]').textContent = fmt.format(paidHours) + ' ч';
      calc.querySelector('[data-out="revenue"]').textContent = fmt.format(revenue) + ' ₽';
      calc.querySelector('[data-out="back"]').textContent = '+' + fmt.format(back) + ' ₽';
      calc.querySelector('[data-out="cost"]').textContent = '−' + fmt.format(veliq) + ' ₽';
    }
    inputs.forEach(function (el) {
      el.addEventListener('input', function () {
        var k = el.getAttribute('data-k');
        calc.querySelectorAll('[data-k="' + k + '"]').forEach(function (other) { if (other !== el) other.value = el.value; });
        recalc();
      });
    });
    recalc();
  }

  /* Копирование контрольной суммы */
  document.addEventListener('click', function (e) {
    var b = e.target.closest('[data-copy]');
    if (!b) return;
    var text = document.getElementById(b.getAttribute('data-copy')).textContent.trim();
    if (navigator.clipboard) navigator.clipboard.writeText(text).then(function () {
      var old = b.getAttribute('aria-label'); b.setAttribute('aria-label', 'Скопировано'); b.dataset.done = '1';
      setTimeout(function () { b.setAttribute('aria-label', old); delete b.dataset.done; }, 1600);
    });
  });
})();
