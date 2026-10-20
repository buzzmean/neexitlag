#!/usr/bin/env python3
"""Генерирует статические страницы website/*.html с общей шапкой и подвалом.

Результат — обычный HTML без сборщика: его можно править руками, но тогда правку шапки/подвала
придётся повторить на всех страницах. Проще поменять здесь и запустить:
    python3 design/tools/build-site.py
Все данные о версии и адресах — в словарях SITE и RELEASE ниже.
"""
import html
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
OUT = ROOT / "website"

SITE = {
    "url": "https://veliq.app",          # ДОПУЩЕНИЕ: домен не выбран, заменить здесь
    "email": "support@veliq.app",        # ДОПУЩЕНИЕ: адрес поддержки не выбран
    "name": "VELIQ",
}
RELEASE = {
    "version": "1.0 beta",               # ДОПУЩЕНИЕ: номер версии подставить при выпуске
    "date": "20 октября 2026",
    "size": "≈ 40 МБ",
    "file": "VELIQ-Setup.exe",
    "url": "downloads/VELIQ-Setup.exe",  # ДОПУЩЕНИЕ: ссылка на файл установщика
    "sha256": "появится вместе с файлом установщика",
    "os": "Windows 10 и 11, 64-бит",
}

NAV = [
    ("index.html#how", "Как работает", "how"),
    ("index.html#games", "Игры", "games"),
    ("clubs.html", "Клубам", "clubs"),
    ("pricing.html", "Цены", "pricing"),
    ("support.html", "Помощь", "support"),
]


def icon(name, cls="i"):
    return f'<svg class="{cls}" aria-hidden="true"><use href="#i-{name}"/></svg>'


def head(title, desc, path, extra_ld=None):
    url = f'{SITE["url"]}/{path if path != "index.html" else ""}'
    ld = ""
    if extra_ld:
        ld = f'\n  <script type="application/ld+json">{json.dumps(extra_ld, ensure_ascii=False)}</script>'
    full_title = title if title.startswith("VELIQ") else f"{title} — VELIQ"
    return f"""<!doctype html>
<html lang="ru">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{html.escape(full_title)}</title>
  <meta name="description" content="{html.escape(desc)}">
  <link rel="canonical" href="{url}">
  <meta name="theme-color" content="#0A0A0B">
  <meta name="color-scheme" content="light dark">
  <meta property="og:type" content="website">
  <meta property="og:site_name" content="VELIQ">
  <meta property="og:locale" content="ru_RU">
  <meta property="og:title" content="{html.escape(full_title)}">
  <meta property="og:description" content="{html.escape(desc)}">
  <meta property="og:url" content="{url}">
  <meta property="og:image" content="{SITE['url']}/assets/img/og-image.png">
  <meta property="og:image:width" content="1200">
  <meta property="og:image:height" content="630">
  <meta property="og:image:alt" content="VELIQ: нажал — работает.">
  <meta name="twitter:card" content="summary_large_image">
  <link rel="icon" href="assets/img/favicon.svg" type="image/svg+xml">
  <link rel="icon" href="assets/img/favicon.ico" sizes="any">
  <link rel="apple-touch-icon" href="assets/img/app-icon-180.png">
  <link rel="manifest" href="site.webmanifest">
  <link rel="preload" href="assets/fonts/Onest-var.woff2" as="font" type="font/woff2" crossorigin>
  <link rel="stylesheet" href="assets/tokens.css">
  <link rel="stylesheet" href="assets/components.css">
  <link rel="stylesheet" href="assets/site.css">
  <script>try{{var t=localStorage.getItem('veliq-theme');if(t)document.documentElement.setAttribute('data-theme',t)}}catch(e){{}}</script>{ld}
</head>
<body>
<script src="assets/icons.js"></script>
<a class="skip" href="#main">Перейти к содержанию</a>
"""


def header(active):
    links = "\n".join(
        f'      <a href="{href}"{" aria-current=\"page\"" if key == active else ""}>{label}</a>'
        for href, label, key in NAV
    )
    return f"""<header class="top">
  <div class="wrap top__in">
    <a class="logo" href="index.html" aria-label="VELIQ — на главную">
      <img class="on-light" src="assets/img/veliq-wordmark-light.svg" alt="VELIQ" width="144" height="66">
      <img class="on-dark" src="assets/img/veliq-wordmark-dark.svg" alt="" width="144" height="66">
    </a>
    <nav class="nav" id="nav" aria-label="Основное меню">
{links}
    </nav>
    <div class="top__tools">
      <button class="btn btn--quiet btn--icon btn--sm theme-btn" type="button" data-theme-toggle aria-label="Сменить тему">
        <svg class="i i-sun-use" aria-hidden="true"><use href="#i-sun"/></svg><svg class="i i-moon-use" aria-hidden="true"><use href="#i-moon"/></svg>
      </button>
      <a class="btn btn--sm top__cta" href="download.html">Скачать</a>
      <button class="btn btn--quiet btn--icon btn--sm menu-btn" type="button" data-menu aria-expanded="false" aria-controls="nav" aria-label="Меню">{icon('menu')}</button>
    </div>
  </div>
</header>
"""


def footer():
    return f"""<footer class="foot">
  <div class="wrap">
    <div class="foot__grid">
      <div>
        <a class="logo" href="index.html" aria-label="VELIQ — на главную">
          <img class="on-light" src="assets/img/veliq-wordmark-light.svg" alt="VELIQ" width="144" height="66">
          <img class="on-dark" src="assets/img/veliq-wordmark-dark.svg" alt="" width="144" height="66">
        </a>
        <p class="muted" style="max-width:34ch;margin:16px 0 0">Оптимальный маршрут для каждого соединения. Для Windows.</p>
      </div>
      <div>
        <h2>Продукт</h2>
        <ul><li><a href="download.html">Скачать</a></li><li><a href="index.html#how">Как работает</a></li><li><a href="index.html#games">Для игр</a></li><li><a href="pricing.html">Цены</a></li></ul>
      </div>
      <div>
        <h2>Клубам</h2>
        <ul><li><a href="clubs.html">VELIQ для клубов</a></li><li><a href="clubs.html#calc">Расчёт выгоды</a></li><li><a href="clubs.html#setup">Установка в образ</a></li></ul>
      </div>
      <div>
        <h2>Помощь</h2>
        <ul><li><a href="support.html">Вопросы и ответы</a></li><li><a href="support.html#errors">Ошибки и решения</a></li><li><a href="mailto:{SITE['email']}">{SITE['email']}</a></li></ul>
      </div>
    </div>
    <div class="foot__bottom">
      <span>© 2026 VELIQ</span>
      <span><a href="privacy.html">Конфиденциальность</a>&nbsp;&nbsp;&nbsp;<a href="terms.html">Условия</a></span>
    </div>
  </div>
</footer>
<script src="assets/charts.js"></script>
<script src="assets/site.js"></script>
</body>
</html>
"""


def page(path, title, desc, active, body, ld=None):
    (OUT / path).write_text(head(title, desc, path, ld) + header(active) + '<main id="main">\n' + body + "\n</main>\n" + footer(), encoding="utf-8")
    print("страница:", path)


def faq(items):
    out = []
    for q, a in items:
        out.append(f'<details><summary>{q}{icon("chevron-down")}</summary><div class="a">{a}</div></details>')
    return "\n".join(out)


# ---------------------------------------------------------------- тексты FAQ (общие)
FAQ_MAIN = [
    ("Это VPN?", "<p>Нет. VPN отправляет весь трафик через один сервер, и всё становится медленнее. VELIQ выбирает путь для каждого соединения отдельно: через свою сеть — когда так открывается доступ или выходит быстрее, обычным путём — когда так лучше.</p>"),
    ("Нужно ли перезапускать браузер?", "<p>Нет. Chrome и Edge подхватывают маршрут сразу после нажатия «Оптимизировать», открытые вкладки закрывать не нужно.</p>"),
    ("Как VELIQ влияет на пинг в играх?", "<p>VELIQ сравнивает пути до игрового сервера и выбирает самый быстрый. Если путь через сеть VELIQ не быстрее обычного, маршрут не меняется. Замеры видно в программе.</p>"),
    ("Вы видите, что я открываю?", "<p>Нет. Мы не расшифровываем и не подменяем HTTPS. Сервер-релей видит только адрес назначения и объём данных — этого достаточно, чтобы переслать трафик. Подробности — в <a href=\"privacy.html\">политике конфиденциальности</a>.</p>"),
    ("Какие системы поддерживаются?", f"<p>{RELEASE['os']}. Нужны права администратора один раз — при установке службы.</p>"),
    ("Работает на компьютерах в клубе, которые откатываются после перезагрузки?", "<p>Да, для этого есть режим клуба: программа ставится в образ, запускается сама при входе и не требует настроек после отката. <a href=\"clubs.html\">Как это устроено</a>.</p>"),
]

# ---------------------------------------------------------------- главная
PING = {"label": "Пинг до игрового сервера: обычный путь около 73 мс, маршрут VELIQ около 41 мс. Иллюстрация.", "series": [
    {"id": "direct", "name": "обычный", "values": [71, 74, 70, 77, 72, 73, 79, 71, 70, 75, 72, 74, 73, 76, 72, 71, 74, 73, 72, 73]},
    {"id": "hel", "name": "VELIQ", "values": [42, 41, 43, 40, 41, 42, 41, 44, 41, 40, 41, 42, 41, 41, 43, 41, 40, 41, 42, 41]},
]}

index_body = f"""
<section class="hero" style="overflow-x:clip">
  <div class="wrap hero__grid">
    <div class="hero__copy">
      <h1 class="h-display hero-title">
        <span class="slit-line"><span class="slit">Нажал —</span></span>
        <span class="slit-line"><span class="slit">работает.</span><span class="slit-bar" aria-hidden="true"></span></span>
      </h1>
      <p class="lede">VELIQ сам выстраивает оптимальный маршрут для каждого соединения. Заблокированные сервисы открываются, игры получают самый быстрый путь до сервера. Браузер перезапускать не нужно.</p>
      <div class="hero__actions">
        <a class="btn btn--lg" href="download.html">{icon('download')}Скачать для Windows</a>
        <a class="btn btn--lg btn--ghost" href="clubs.html">Для компьютерных клубов</a>
      </div>
      <p class="hero__meta">Версия {RELEASE['version']}, {RELEASE['os']}</p>
    </div>

    <div class="demo" data-demo data-state="off" aria-label="Демонстрация программы">
      <div class="demo__bar">
        <img src="assets/img/veliq-wordmark-dark.svg" alt="VELIQ" width="48" height="22">
        <span class="status status--off" data-demo-status><i class="status__mark" aria-hidden="true"></i><span>Выключено</span></span>
      </div>
      <button class="go" type="button" data-state="off" aria-pressed="false" aria-describedby="demo-hint">
        <span class="go__label"><span data-go-label>Оптимизировать</span><small data-go-sub>выключено</small></span>
      </button>
      <div class="route" aria-label="Маршруты">
        <div class="route__row route__row--relay"><span class="route__name">YouTube</span><span class="route__line"></span><span class="route__via">открыт</span></div>
        <div class="route__row route__row--relay"><span class="route__name">Discord</span><span class="route__line"></span><span class="route__via">открыт</span></div>
        <div class="route__row route__row--relay"><span class="route__name">ChatGPT</span><span class="route__line"></span><span class="route__via">открыт</span></div>
        <div class="route__row route__row--relay"><span class="route__name">Игры</span><span class="route__line"></span><span class="route__via">быстрый путь</span></div>
      </div>
      <p class="demo__hint" id="demo-hint">Нажмите кнопку — так выглядит запуск в программе.</p>
      <span class="sr-only" aria-live="polite" data-demo-live></span>
    </div>
  </div>
</section>

<section class="sec band--sunken" id="how">
  <div class="wrap">
    <div class="sec__head">
      <h2 class="h2">Три шага, и всё открывается</h2>
      <p>Никаких адресов серверов, ключей и настроек браузера. Всё это программа делает сама.</p>
    </div>
    <ol class="steps" style="list-style:none;margin:0;padding:0">
      <li class="step"><span class="step__n">1</span><h3 class="h3">Скачайте</h3><p>Один установщик для Windows 10 и 11.</p></li>
      <li class="step"><span class="step__n">2</span><h3 class="h3">Установите</h3><p>Windows один раз спросит права администратора — это для службы, которая строит маршруты.</p></li>
      <li class="step"><span class="step__n">3</span><h3 class="h3">Нажмите «Оптимизировать»</h3><p>Маршруты строятся сразу. Заблокированное открывается в уже открытом браузере.</p></li>
    </ol>
  </div>
</section>

<section class="sec">
  <div class="wrap">
    <div class="sec__head">
      <h2 class="h2">Свой путь для каждого соединения</h2>
      <p>Одним сервисам нужен доступ, играм — скорость. VELIQ понимает разницу и подбирает маршрут сам, без списков и настроек.</p>
    </div>
    <div class="split">
      <div class="split__col">
        <h3 class="h3">Открывает доступ</h3>
        <ul class="split__list">
          <li><span class="mono">YT</span><span>YouTube</span><span>через сеть VELIQ</span></li>
          <li><span class="mono">DS</span><span>Discord</span><span>через сеть VELIQ</span></li>
          <li><span class="mono">TG</span><span>Telegram</span><span>через сеть VELIQ</span></li>
          <li><span class="mono">AI</span><span>ChatGPT, Claude и другие нейросети</span><span>через сеть VELIQ</span></li>
        </ul>
        <p class="split__note">Это примеры. Список сервисов пополняется, свой сайт или приложение можно добавить в программе.</p>
      </div>
      <div class="split__col">
        <h3 class="h3">Ускоряет</h3>
        <ul class="split__list">
          <li><span class="mono">{icon('games','i i--sm')}</span><span>Онлайн-игры</span><span>самый быстрый путь</span></li>
          <li><span class="mono">{icon('globe','i i--sm')}</span><span>Игровые серверы в Европе</span><span>доступ и низкий пинг</span></li>
          <li><span class="mono">{icon('ping','i i--sm')}</span><span>Голосовые чаты и стримы</span><span>стабильный канал</span></li>
          <li><span class="mono">{icon('check','i i--sm')}</span><span>То, что и так работает быстро</span><span>не трогаем</span></li>
        </ul>
        <p class="split__note">VELIQ замеряет задержку и выбирает путь, на котором она меньше.</p>
      </div>
    </div>
  </div>
</section>

<section class="sec sec--tight">
  <div class="wrap">
    <div class="facts">
      <div class="fact fact--wide"><h3>Браузер не нужно перезапускать</h3><p>Маршрут подхватывается в открытых Chrome и Edge. Вкладки, вход в аккаунты и загрузки остаются на месте.</p></div>
      <div class="fact"><h3>Своя сеть серверов</h3><p>Собственные серверы в Европе, а не общие бесплатные узлы, где скорость делят тысячи людей.</p></div>
      <div class="fact fact--r"><h3>HTTPS не расшифровывается</h3><p>Сервер пересылает зашифрованный трафик как есть. Пароли и переписка ему не видны.</p></div>
      <div class="fact"><h3>Ничего не настраивать</h3><p>Одна кнопка. Если что-то пошло не так, программа скажет, что именно, и предложит решение.</p></div>
      <div class="fact fact--r"><h3>Готово для клубов</h3><p>Работает на ПК, которые откатываются после перезагрузки: запускается сам, настройки не теряются.</p></div>
    </div>
  </div>
</section>

<section class="sec band--sunken" id="games">
  <div class="wrap games">
    <div>
      <h2 class="h2">Оптимальный маршрут для игр</h2>
      <p class="lede" style="margin-top:20px">VELIQ сравнивает пути до игрового сервера и выбирает самый быстрый. Если лучше обычного пути нет, ничего не меняется.</p>
      <ul class="ticks">
        <li>{icon('check')}<span>Доступ к европейским серверам, закрытым для российских адресов</span></li>
        <li>{icon('check')}<span>Меньший пинг, если через сеть VELIQ путь короче, чем у провайдера</span></li>
        <li>{icon('check')}<span>Маршрут перестраивается сам, если какой-то путь стал хуже</span></li>
      </ul>
    </div>
    <figure class="games__panel" style="margin:0">
      <header><strong>Пинг до игрового сервера в Европе</strong><ul class="legend"><li><i class="direct"></i>обычный путь</li><li><i class="hel"></i>маршрут VELIQ</li></ul></header>
      <div data-ping-chart='{json.dumps(PING, ensure_ascii=False)}'></div>
      <footer>Иллюстрация. Настоящий замер VELIQ делает на вашем компьютере и у вашего провайдера.</footer>
    </figure>
  </div>
</section>

<section class="sec band--ink">
  <div class="wrap club-band">
    <div>
      <h2 class="h2">Клубные ПК с откатом после перезагрузки — и всё работает</h2>
      <p class="lede" style="margin-top:20px">Режим клуба для Langame и других систем управления: VELIQ ставится в образ, запускается при входе гостя и не просит настроек.</p>
      <div class="hero__actions" style="margin-top:32px"><a class="btn btn--lg" href="clubs.html">VELIQ для клубов</a><a class="btn btn--lg btn--ghost" href="clubs.html#calc">Посчитать выгоду</a></div>
    </div>
    <div class="club-band__pcs" aria-hidden="true">{''.join('<i></i>' for _ in range(27))}</div>
  </div>
</section>

<section class="sec" id="download">
  <div class="wrap">
    <div class="get">
      <img class="get__icon" src="assets/img/app-icon.svg" alt="" width="88" height="88">
      <div>
        <h2 class="h3" style="font-size:var(--fs-3xl)">VELIQ для Windows</h2>
        <ul class="get__meta"><li>Версия {RELEASE['version']}</li><li>{RELEASE['size']}</li><li>{RELEASE['os']}</li></ul>
      </div>
      <a class="btn btn--lg" href="download.html">{icon('download')}Скачать</a>
    </div>
  </div>
</section>

<section class="sec sec--tight" id="faq" style="padding-top:0">
  <div class="wrap">
    <div class="sec__head"><h2 class="h2">Частые вопросы</h2><p>Не нашли ответ — <a href="support.html">все вопросы и ошибки</a> или напишите на {SITE['email']}.</p></div>
    <div class="faq">
{faq(FAQ_MAIN)}
    </div>
  </div>
</section>
"""

LD_APP = {
    "@context": "https://schema.org", "@type": "SoftwareApplication", "name": "VELIQ",
    "operatingSystem": "Windows 10, Windows 11", "applicationCategory": "UtilitiesApplication",
    "description": "VELIQ строит оптимальный маршрут для каждого соединения: заблокированные сервисы открываются, игры получают самый быстрый путь.",
    "softwareVersion": RELEASE["version"], "url": SITE["url"],
}
LD_FAQ = {"@context": "https://schema.org", "@type": "FAQPage", "mainEntity": [
    {"@type": "Question", "name": q, "acceptedAnswer": {"@type": "Answer", "text": html.unescape(a.replace("<p>", "").replace("</p>", ""))}} for q, a in FAQ_MAIN]}

page("index.html", "VELIQ — нажал, и работает",
     "VELIQ сам строит оптимальный маршрут: заблокированные сервисы открываются без перезапуска браузера, игры получают самый быстрый путь. Для Windows 10 и 11.",
     "", index_body, [LD_APP, LD_FAQ])

# ---------------------------------------------------------------- скачать
download_body = f"""
<section class="page-head">
  <div class="wrap">
    <div class="crumbs"><a href="index.html">Главная</a> / Скачать</div>
    <h1 class="h1">Скачать VELIQ</h1>
    <p class="lede">Установка занимает около минуты. После неё остаётся нажать одну кнопку.</p>
  </div>
</section>
<section class="sec sec--tight">
  <div class="wrap cols">
    <div>
      <div class="get" style="margin-bottom:var(--sp-7)">
        <img class="get__icon" src="assets/img/app-icon.svg" alt="" width="88" height="88">
        <div>
          <h2 class="h3" style="font-size:var(--fs-3xl)">VELIQ {RELEASE['version']}</h2>
          <ul class="get__meta"><li>{RELEASE['date']}</li><li>{RELEASE['size']}</li><li>{RELEASE['file']}</li></ul>
        </div>
        <a class="btn btn--lg" href="{RELEASE['url']}" download>{icon('download')}Скачать установщик</a>
      </div>

      <h2 class="h2" id="install" style="font-size:var(--fs-4xl);margin-bottom:var(--sp-5)">Как установить</h2>
      <ol class="olist">
        <li><div><h3>Запустите {RELEASE['file']}</h3><p>Если Windows покажет окно «Система Windows защитила ваш компьютер», нажмите «Подробнее», затем «Выполнить в любом случае». Так бывает с новыми программами, которые Windows ещё не встречала.</p></div></li>
        <li><div><h3>Разрешите установку службы</h3><p>Windows спросит права администратора один раз. Служба VELIQ строит маршруты и запускается вместе с системой.</p></div></li>
        <li><div><h3>Нажмите «Оптимизировать»</h3><p>Программа откроется сама. После нажатия маршруты строятся сразу, заблокированное открывается в уже открытом браузере.</p></div></li>
      </ol>

      <div class="note" style="margin-top:var(--sp-6)">{icon('info')}<div><strong>Компьютерный клуб?</strong> Для ПК с откатом после перезагрузки есть отдельная инструкция: <a href="clubs.html#setup">установка в образ</a>.</div></div>
    </div>
    <aside class="aside">
      <div class="card">
        <h2 class="h3" style="margin-bottom:16px">О файле</h2>
        <dl class="kv">
          <dt>Версия</dt><dd>{RELEASE['version']}</dd>
          <dt>Дата</dt><dd>{RELEASE['date']}</dd>
          <dt>Размер</dt><dd>{RELEASE['size']}</dd>
          <dt>Система</dt><dd>{RELEASE['os']}</dd>
          <dt>SHA-256</dt><dd><span id="sha">{RELEASE['sha256']}</span></dd>
        </dl>
        <button class="btn btn--ghost btn--sm" type="button" data-copy="sha" aria-label="Скопировать SHA-256" style="margin-top:16px">{icon('copy','i i--sm')}Скопировать SHA-256</button>
      </div>
      <div class="card card--sunken">
        <h2 class="h3" style="margin-bottom:12px">Что понадобится</h2>
        <ul style="margin:0;padding-left:18px;color:var(--text-muted);display:grid;gap:6px;font-size:var(--fs-sm)">
          <li>Windows 10 или 11, 64-бит</li><li>Права администратора на время установки</li><li>Для сайтов — Chrome или Edge</li>
        </ul>
      </div>
    </aside>
  </div>
</section>
"""
page("download.html", "Скачать VELIQ для Windows",
     f"Скачайте VELIQ {RELEASE['version']} для Windows 10 и 11. Установка за минуту, затем одна кнопка «Оптимизировать».",
     "", download_body, LD_APP)

# ---------------------------------------------------------------- клубы
CLUB_FAQ = [
    ("Как VELIQ переживает откат после перезагрузки?", "<p>Программа и служба ставятся в образ в техническом режиме, как игры. После отката они на месте и стартуют сами, без повторной настройки.</p>"),
    ("Нужно ли что-то делать гостю?", "<p>Нет. В режиме клуба кнопка нажата заранее: при входе гостя VELIQ уже работает. Гость видит только значок в трее.</p>"),
    ("Как VELIQ влияет на пинг в играх?", "<p>VELIQ выбирает самый быстрый путь до игрового сервера. Если обычный путь быстрее, маршрут не меняется.</p>"),
    ("Работает с Langame?", "<p>Да. Программа ставится в образ так же, как остальной софт клуба.</p>"),
    ("Сколько это стоит?", "<p>Тариф для клубов считается за компьютер в месяц. Цена появится к релизу — оставьте заявку, и мы сообщим первыми. Пока можно прикинуть выгоду <a href=\"#calc\">в калькуляторе</a>.</p>"),
]
club_calc = """
<div class="calc" data-calc>
  <div class="calc__in">
    <div class="calc__row"><label for="c-pcs">Компьютеров в клубе</label><input class="input" id="c-pcs" data-k="pcs" inputmode="numeric" value="30"><input type="range" data-k="pcs" min="5" max="200" step="1" value="30" aria-label="Компьютеров в клубе"></div>
    <div class="calc__row"><label for="c-hours">Часов работы в сутки</label><input class="input" id="c-hours" data-k="hours" inputmode="numeric" value="24"><input type="range" data-k="hours" min="6" max="24" step="1" value="24" aria-label="Часов работы в сутки"></div>
    <div class="calc__row"><label for="c-load">Средняя загрузка, %</label><input class="input" id="c-load" data-k="load" inputmode="numeric" value="45"><input type="range" data-k="load" min="5" max="100" step="1" value="45" aria-label="Средняя загрузка"></div>
    <div class="calc__row"><label for="c-price">Цена часа, ₽</label><input class="input" id="c-price" data-k="price" inputmode="numeric" value="120"><input type="range" data-k="price" min="40" max="500" step="10" value="120" aria-label="Цена часа"></div>
    <div class="calc__row"><label for="c-lost">Часов теряете из-за блокировок, %</label><input class="input" id="c-lost" data-k="lost" inputmode="decimal" value="3"><input type="range" data-k="lost" min="0" max="20" step="0.5" value="3" aria-label="Потерянные часы"><small>Гости, которые ушли или не пришли из-за недоступных сервисов или высокого пинга.</small></div>
    <div class="calc__row"><label for="c-cost">VELIQ за ПК в месяц, ₽</label><input class="input" id="c-cost" data-k="cost" inputmode="numeric" value="0"><input type="range" data-k="cost" min="0" max="1000" step="10" value="0" aria-label="Стоимость VELIQ"><small>Цена тарифа ещё не объявлена — подставьте свою оценку.</small></div>
  </div>
  <div class="calc__out" aria-live="polite">
    <div><div style="color:var(--c-ink-300);margin-bottom:10px">Возвращённая выручка в месяц</div><div class="calc__big" data-out="net">—</div></div>
    <dl class="calc__lines">
      <div><dt>Оплаченных часов в месяц</dt><dd data-out="hours">—</dd></div>
      <div><dt>Выручка за часы</dt><dd data-out="revenue">—</dd></div>
      <div><dt>Возвращается с VELIQ</dt><dd data-out="back">—</dd></div>
      <div><dt>Стоимость VELIQ</dt><dd data-out="cost">—</dd></div>
    </dl>
    <p class="calc__formula">Расчёт: ПК × часы × 30 дней × загрузка × цена часа × доля потерянных часов − стоимость VELIQ. Это оценка по вашим цифрам, а не обещание дохода.</p>
  </div>
</div>
"""
clubs_body = f"""
<section class="page-head">
  <div class="wrap">
    <div class="crumbs"><a href="index.html">Главная</a> / Клубам</div>
    <h1 class="h1" style="max-width:15ch">VELIQ для компьютерных клубов</h1>
    <p class="lede">Гости приходят играть, а заодно смотреть стримы, сидеть в голосовых чатах и пользоваться нейросетями. Когда это не работает или игра лагает, они уходят туда, где всё в порядке.</p>
    <div class="hero__actions" style="margin-top:32px"><a class="btn btn--lg" href="mailto:{SITE['email']}?subject=VELIQ%20для%20клуба">Оставить заявку</a><a class="btn btn--lg btn--ghost" href="#calc">Посчитать выгоду</a></div>
  </div>
</section>

<section class="sec">
  <div class="wrap">
    <div class="facts">
      <div class="fact"><h3>Запускается сам</h3><p>Кнопка «Оптимизировать» нажата заранее. Гость садится — маршруты уже построены.</p></div>
      <div class="fact fact--r"><h3>Переживает откат</h3><p>Ставится в образ вместе с играми. После перезагрузки не нужно ничего восстанавливать.</p></div>
      <div class="fact"><h3>Лучший путь для игр</h3><p>VELIQ выбирает самый быстрый маршрут до игрового сервера и открывает доступ к закрытым.</p></div>
      <div class="fact fact--r"><h3>Гость ничего не сломает</h3><p>Упрощённый интерфейс: состояние и одна кнопка. Настройки закрыты паролем администратора.</p></div>
    </div>
  </div>
</section>

<section class="sec band--sunken" id="setup">
  <div class="wrap cols">
    <div>
      <h2 class="h2" style="margin-bottom:var(--sp-6)">Установка в клубе</h2>
      <ol class="olist">
        <li><div><h3>Включите технический режим на одном ПК</h3><p>В Langame или другой системе — так же, как для установки игр, чтобы изменения сохранились в образе.</p></div></li>
        <li><div><h3>Установите VELIQ и включите режим клуба</h3><p>«Настройки» → «Режим клуба». Задайте пароль администратора и включите «Оптимизировать при входе».</p></div></li>
        <li><div><h3>Сохраните образ и разверните на остальные ПК</h3><p>Так же, как вы обновляете игры. Настраивать каждый компьютер отдельно не нужно.</p></div></li>
        <li><div><h3>Проверьте на любом ПК</h3><p>Перезагрузите компьютер: значок VELIQ в трее должен стать «Работает» в течение 10 секунд.</p></div></li>
      </ol>
    </div>
    <aside class="aside">
      <div class="note">{icon('snow')}<div><strong>ПК с «заморозкой».</strong> Всё, что не попало в образ, исчезнет после перезагрузки. Ставьте VELIQ только в техническом режиме.</div></div>
      <div class="note">{icon('info')}<div><strong>Нужна помощь с установкой?</strong> Напишите на <a href="mailto:{SITE['email']}">{SITE['email']}</a> — поможем настроить первый ПК вместе с администратором.</div></div>
    </aside>
  </div>
</section>

<section class="sec" id="calc">
  <div class="wrap">
    <div class="sec__head"><h2 class="h2">Сколько часов возвращает VELIQ</h2><p>Подставьте цифры своего клуба. Все значения можно менять.</p></div>
    {club_calc}
  </div>
</section>

<section class="sec sec--tight" style="padding-top:0">
  <div class="wrap">
    <div class="sec__head"><h2 class="h2">Вопросы владельцев клубов</h2><p>Презентацию для партнёров и инструкцию для администратора пришлём по запросу.</p></div>
    <div class="faq">{faq(CLUB_FAQ)}</div>
  </div>
</section>
"""
page("clubs.html", "VELIQ для компьютерных клубов",
     "Заблокированные сервисы и быстрый путь для игр на всех ПК клуба. Работает на компьютерах с откатом после перезагрузки, в том числе с Langame. Калькулятор выгоды.",
     "clubs", clubs_body)

# ---------------------------------------------------------------- цены
pricing_body = f"""
<section class="page-head">
  <div class="wrap">
    <div class="crumbs"><a href="index.html">Главная</a> / Цены</div>
    <h1 class="h1">Цены</h1>
    <p class="lede">Тарифы готовятся к релизу. Пока VELIQ в бета-версии, здесь — какими они будут по составу.</p>
  </div>
</section>
<section class="sec sec--tight">
  <div class="wrap">
    <div class="note note--draft" style="margin-bottom:var(--sp-6)">{icon('info')}<div><strong>Заготовка страницы.</strong> Цены и сроки не объявлены. Числа появятся здесь, когда будут утверждены.</div></div>
    <div class="plans">
      <div class="plan">
        <h2 class="h3">Игрок</h2>
        <div class="plan__price">— ₽ <small>в месяц</small></div>
        <p class="muted" style="margin:0">Один компьютер дома.</p>
        <ul><li>{icon('check')}Доступ к заблокированным сервисам</li><li>{icon('check')}Собственная сеть серверов</li><li>{icon('check')}Оптимальный маршрут для игр</li></ul>
        <a class="btn btn--ghost" href="download.html">Скачать бету</a>
      </div>
      <div class="plan plan--main">
        <h2 class="h3">Клуб</h2>
        <div class="plan__price">— ₽ <small>за ПК в месяц</small></div>
        <p class="muted" style="margin:0">Для клубов от 10 компьютеров.</p>
        <ul><li>{icon('check')}Всё из тарифа «Игрок»</li><li>{icon('check')}Режим клуба и работа после отката</li><li>{icon('check')}Пароль администратора</li><li>{icon('check')}Помощь с установкой в образ</li></ul>
        <a class="btn" href="mailto:{SITE['email']}?subject=Тариф%20Клуб">Оставить заявку</a>
      </div>
      <div class="plan">
        <h2 class="h3">Сеть клубов</h2>
        <div class="plan__price">по договору</div>
        <p class="muted" style="margin:0">Несколько адресов, свои условия.</p>
        <ul><li>{icon('check')}Всё из тарифа «Клуб»</li><li>{icon('check')}Условия под количество адресов</li></ul>
        <a class="btn btn--ghost" href="mailto:{SITE['email']}?subject=Сеть%20клубов">Написать нам</a>
      </div>
    </div>
  </div>
</section>
"""
page("pricing.html", "Цены", "Тарифы VELIQ для игроков и компьютерных клубов: что входит в каждый.", "pricing", pricing_body)

# ---------------------------------------------------------------- поддержка
ERRORS = [
    ("Служба VELIQ не запущена", "Программа не может связаться со службой, которая направляет трафик.", "Нажмите «Запустить службу» в программе. Если не помогло — перезагрузите компьютер. Если служба удалена антивирусом, переустановите VELIQ и добавьте папку программы в исключения."),
    ("Нет связи с сервером", "Серверы VELIQ не отвечают с вашего подключения.", "Проверьте, что интернет работает без VELIQ. Затем нажмите «Повторить» — программа попробует другой сервер. Если не помогло, напишите нам: возможно, провайдер блокирует соединение."),
    ("Браузер не подключился", "Chrome или Edge не принял маршрут.", "Нажмите «Повторить». Если браузером управляет организация (рабочий компьютер), её правила могут запрещать изменения — тогда маршруты работают только для приложений вне браузера."),
    ("Ключ устройства недействителен", "Компьютер не прошёл проверку на сервере — например после переустановки Windows.", "Нажмите «Зарегистрировать заново». В клубе после этого обновите образ."),
    ("Не хватает прав", "Действие требует прав администратора.", "Подтвердите запрос Windows. В режиме клуба — введите пароль администратора VELIQ."),
]
err_html = "\n".join(
    f'<details><summary>{t}{icon("chevron-down")}</summary><div class="a"><p><strong>Что случилось.</strong> {w}</p><p><strong>Что сделать.</strong> {f}</p></div></details>'
    for t, w, f in ERRORS)
SUPPORT_FAQ = FAQ_MAIN + [
    ("Как выключить VELIQ?", "<p>Нажмите большую кнопку ещё раз или выберите «Выключить» в меню значка в трее. Интернет продолжит работать как без программы.</p>"),
    ("Как добавить свой сайт или приложение?", "<p>Вкладка «Сервисы» → «Добавить». VELIQ подберёт для него маршрут так же, как для остальных.</p>"),
    ("Как удалить программу?", "<p>Сначала выключите оптимизацию, затем «Параметры Windows» → «Приложения» → VELIQ → «Удалить».</p>"),
]
support_body = f"""
<section class="page-head">
  <div class="wrap">
    <div class="crumbs"><a href="index.html">Главная</a> / Помощь</div>
    <h1 class="h1">Помощь</h1>
    <p class="lede">Ответы на частые вопросы и что делать при каждой ошибке, которую показывает программа.</p>
  </div>
</section>
<section class="sec sec--tight">
  <div class="wrap cols">
    <div>
      <div class="faq">
        <h3 class="faq__group">Общие вопросы</h3>
{faq(SUPPORT_FAQ)}
        <h3 class="faq__group" id="errors">Ошибки в программе</h3>
{err_html}
      </div>
    </div>
    <aside class="aside">
      <div class="card">
        <h2 class="h3" style="margin-bottom:10px">Не помогло?</h2>
        <p class="muted" style="margin:0 0 16px;font-size:var(--fs-sm)">Откройте в программе «Диагностика» → «Сохранить отчёт» и пришлите файл. В отчёте нет паролей и ключей.</p>
        <a class="btn btn--sm" href="mailto:{SITE['email']}?subject=Помощь%20VELIQ">{SITE['email']}</a>
      </div>
    </aside>
  </div>
</section>
"""
page("support.html", "Помощь и частые вопросы", "Ответы на вопросы о VELIQ и решения для каждой ошибки программы.", "support", support_body,
     {"@context": "https://schema.org", "@type": "FAQPage", "mainEntity": [
         {"@type": "Question", "name": q, "acceptedAnswer": {"@type": "Answer", "text": a.replace("<p>", "").replace("</p>", "")}} for q, a in SUPPORT_FAQ]})

# ---------------------------------------------------------------- правовые
DRAFT = f'<div class="note note--draft" style="margin-bottom:var(--sp-6)">{icon("alert")}<div><strong>Черновик.</strong> Текст подготовлен как основа и должен быть проверен юристом до публикации. Реквизиты владельца сервиса не указаны.</div></div>'
privacy_body = f"""
<section class="page-head"><div class="wrap"><div class="crumbs"><a href="index.html">Главная</a> / Конфиденциальность</div><h1 class="h1">Политика конфиденциальности</h1><p class="lede">Какие данные получает VELIQ, зачем и сколько их хранит.</p></div></section>
<section class="sec sec--tight"><div class="wrap prose">
{DRAFT}
<h2>Коротко</h2>
<p>VELIQ направляет часть соединений через свои серверы, когда так открывается доступ или выходит быстрее. <strong>Мы не расшифровываем и не подменяем HTTPS</strong>: содержимое страниц, пароли и переписка серверу не видны.</p>
<h2>Что видит сервер-релей</h2>
<ul><li>Адрес и порт назначения, имя сайта из начала соединения (SNI).</li><li>Объём переданных данных и время соединения.</li><li>Ключ вашего устройства — чтобы пускать на сервер только установленные копии VELIQ.</li></ul>
<p>Эти данные нужны, чтобы переслать трафик и защитить серверы от злоупотреблений. Срок хранения журналов соединений: [указать до публикации].</p>
<h2>Что остаётся на компьютере</h2>
<p>Настройки, список сервисов и журнал диагностики хранятся только на вашем компьютере. Отчёт диагностики отправляется нам, только если вы сами его пришлёте.</p>
<h2>Что мы не делаем</h2>
<ul><li>Не продаём и не передаём данные рекламным сетям.</li><li>Не показываем рекламу и не встраиваем её в страницы.</li></ul>
<h2>Сайт</h2>
<p>Сайт не использует счётчики и рекламные cookie. Выбор темы сохраняется только в вашем браузере.</p>
<h2>Связь</h2>
<p>Вопросы о данных — на <a href="mailto:{SITE['email']}">{SITE['email']}</a>.</p>
</div></section>
"""
page("privacy.html", "Политика конфиденциальности", "Какие данные получает VELIQ: HTTPS не расшифровывается, сервер видит только адрес назначения и объём трафика.", "", privacy_body)

terms_body = f"""
<section class="page-head"><div class="wrap"><div class="crumbs"><a href="index.html">Главная</a> / Условия</div><h1 class="h1">Условия использования</h1><p class="lede">Правила, на которых предоставляется VELIQ.</p></div></section>
<section class="sec sec--tight"><div class="wrap prose">
{DRAFT}
<h2>1. Что такое VELIQ</h2>
<p>VELIQ — программа для Windows, которая выбирает маршрут для соединений компьютера: через серверы VELIQ или обычным путём, в зависимости от того, что обеспечивает доступ и меньшую задержку.</p>
<h2>2. Бета-версия</h2>
<p>Сейчас программа в бета-версии. Она может работать с перебоями, а функции — меняться. Мы предупреждаем об изменениях в программе и на сайте.</p>
<h2>3. Допустимое использование</h2>
<p>Запрещено использовать VELIQ для рассылки спама, атак на чужие сети и сервисы, распространения вредоносных программ и любых действий, нарушающих закон. При нарушении доступ устройства к серверам может быть закрыт.</p>
<h2>4. Ответственность</h2>
<p>Мы стараемся, чтобы серверы работали без перерывов, но не гарантируем доступность сторонних сервисов: их работа от нас не зависит.</p>
<h2>5. Оплата</h2>
<p>Условия оплаты появятся вместе с тарифами на странице <a href="pricing.html">«Цены»</a>.</p>
<h2>6. Изменения условий</h2>
<p>Новая редакция публикуется на этой странице с датой вступления в силу.</p>
<h2>7. Связь</h2>
<p><a href="mailto:{SITE['email']}">{SITE['email']}</a></p>
</div></section>
"""
page("terms.html", "Условия использования", "Условия использования программы VELIQ.", "", terms_body)

nf_body = f"""
<section class="center-stack"><div class="wrap">
  <h1 class="h-display"><span class="slit-line" style="margin:0 auto"><span class="slit">404</span></span></h1>
  <p class="lede" style="margin:24px auto 32px">Такой страницы нет. Возможно, её перенесли.</p>
  <a class="btn btn--lg" href="index.html">На главную</a>
</div></section>
"""
page("404.html", "Страница не найдена", "Страница не найдена.", "", nf_body)

# ---------------------------------------------------------------- служебные файлы
pages = ["", "download.html", "clubs.html", "pricing.html", "support.html", "privacy.html", "terms.html"]
(OUT / "sitemap.xml").write_text(
    '<?xml version="1.0" encoding="UTF-8"?>\n<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">\n'
    + "".join(f"  <url><loc>{SITE['url']}/{p}</loc></url>\n" for p in pages) + "</urlset>\n", encoding="utf-8")
(OUT / "robots.txt").write_text(f"User-agent: *\nAllow: /\nSitemap: {SITE['url']}/sitemap.xml\n", encoding="utf-8")
(OUT / "site.webmanifest").write_text(json.dumps({
    "name": "VELIQ", "short_name": "VELIQ", "lang": "ru", "start_url": "index.html", "display": "browser",
    "background_color": "#0A0A0B", "theme_color": "#0A0A0B",
    "icons": [{"src": "assets/img/app-icon-192.png", "sizes": "192x192", "type": "image/png"},
              {"src": "assets/img/app-icon-512.png", "sizes": "512x512", "type": "image/png"}]}, ensure_ascii=False, indent=2), encoding="utf-8")
print("служебные файлы: sitemap.xml, robots.txt, site.webmanifest")
