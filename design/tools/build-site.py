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
    ("vpn.html", "Как работает", "how"),
    ("index.html#games", "Игры", "games"),
    ("clubs.html", "Клубам", "clubs"),
    ("pricing.html", "Цены", "pricing"),
    ("support.html", "Помощь", "support"),
]


def icon(name, cls="i"):
    return f'<svg class="{cls}" aria-hidden="true"><use href="#i-{name}"/></svg>'


def head(title, desc, path, extra_ld=None, og="og-image.png", og_alt="VELIQ: нажал — работает."):
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
  <meta property="og:image" content="{SITE['url']}/assets/img/{og}">
  <meta property="og:image:width" content="1200">
  <meta property="og:image:height" content="630">
  <meta property="og:image:alt" content="{html.escape(og_alt)}">
  <meta name="twitter:card" content="summary_large_image">
  <meta name="twitter:image" content="{SITE['url']}/assets/img/{og}">
  <link rel="icon" href="assets/img/favicon.svg" type="image/svg+xml">
  <link rel="icon" href="assets/img/favicon.ico" sizes="any">
  <link rel="apple-touch-icon" href="assets/img/app-icon-180.png">
  <link rel="manifest" href="site.webmanifest">
  <link rel="preload" href="assets/fonts/Onest-var.woff2" as="font" type="font/woff2" crossorigin>
  <link rel="stylesheet" href="assets/tokens.css">
  <link rel="stylesheet" href="assets/components.css">
  <link rel="stylesheet" href="assets/site.css">
  <script>document.documentElement.classList.add('js');try{{var t=localStorage.getItem('veliq-theme');if(t)document.documentElement.setAttribute('data-theme',t)}}catch(e){{}}</script>{ld}
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
        <p class="muted" style="max-width:34ch;margin:16px 0 0">Оптимальный маршрут для каждого соединения.</p>
      </div>
      <div>
        <h2>Продукт</h2>
        <ul><li><a href="download.html">Скачать</a></li><li><a href="index.html#how">Как работает</a></li><li><a href="vpn.html">Отличие от VPN</a></li><li><a href="index.html#privacy">Что видит сервер</a></li><li><a href="index.html#games">Для игр</a></li><li><a href="pricing.html">Цены</a></li><li><a href="roadmap.html">План развития</a></li></ul>
      </div>
      <div>
        <h2>Клубам</h2>
        <ul><li><a href="clubs.html">VELIQ для клубов</a></li><li><a href="clubs.html#calc">Расчёт выгоды</a></li><li><a href="clubs.html#setup">Установка в клубе</a></li></ul>
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


def page(path, title, desc, active, body, ld=None, og="og-image.png", og_alt="VELIQ: нажал — работает."):
    (OUT / path).write_text(head(title, desc, path, ld, og, og_alt) + header(active) + '<main id="main">\n' + body + "\n</main>\n" + footer(), encoding="utf-8")
    print("страница:", path)


def faq(items):
    out = []
    for q, a in items:
        out.append(f'<details><summary>{q}{icon("chevron-down")}</summary><div class="a">{a}</div></details>')
    return "\n".join(out)


# ---------------------------------------------------------------- тексты FAQ (общие)

def flow(kind):
    """Схема: устройство → (сервер) → назначения. kind = 'vpn' | 'veliq'. HTML/CSS, без картинок."""
    rows = [("YouTube", "заблокирован"), ("Онлайн-игра", "нужен низкий пинг"), ("Обычный сайт", "и так работает"), ("Банк", "и так работает")]
    out = []
    if kind == "vpn":
        out.append('<div class="flow flow--vpn" data-trace>')
        out.append('<div class="flow__src">Вы</div><div class="flow__hub">Сервер VPN</div>')
        for i, (name, _) in enumerate(rows):
            out.append(f'<span class="flow__a" style="grid-row:{i+1}"></span><span class="flow__b" style="grid-row:{i+1}"></span><span class="flow__dst" style="grid-row:{i+1}">{name}</span>')
        out.append('</div>')
        out.append('<p class="flow__cap">Весь трафик идёт через один сервер — и заблокированное, и то, что прекрасно работало без него.</p>')
    else:
        mids = [("relay", "VELIQ"), ("relay", "лучший путь"), ("plain", "как обычно"), ("plain", "как обычно")]
        out.append('<div class="flow flow--veliq" data-trace>')
        out.append('<div class="flow__src">Вы</div>')
        for i, ((name, _), (mode, label)) in enumerate(zip(rows, mids)):
            g = f'grid-row:{i+1}'
            if mode == "relay":
                out.append(f'<span class="flow__a is-sig" style="{g}"></span><span class="flow__node" style="{g}">{label}</span><span class="flow__b is-sig" style="{g}"></span><span class="flow__dst" style="{g}">{name}</span>')
            else:
                out.append(f'<span class="flow__plain" style="{g}"><em>{label}</em></span><span class="flow__dst is-plain" style="{g}">{name}</span>')
        out.append('</div>')
        out.append('<p class="flow__cap">Каждое соединение получает свой путь: заблокированное — через сеть VELIQ, игры — самый быстрый маршрут, остальное — как обычно.</p>')
    return "\n".join(out)


def privacy_block(head_tag="h2"):
    return f"""<div class="sec__head"><{head_tag} class="h2">Что видит сервер</{head_tag}><p>Сервер VELIQ пересылает трафик, но не может его прочитать: соединения с сайтами и приложениями зашифрованы от начала до конца.</p></div>
    <div class="seen">
      <div class="seen__path" data-trace aria-label="Путь соединения: ваше устройство, сервер VELIQ, сервис. На всём пути данные зашифрованы.">
        <span class="seen__end">Ваше устройство</span>
        <span class="seen__wire">{icon('lock','i i--sm')}<em>зашифровано</em></span>
        <span class="seen__mid">Сервер VELIQ</span>
        <span class="seen__wire">{icon('lock','i i--sm')}<em>зашифровано</em></span>
        <span class="seen__end">Сервис</span>
      </div>
      <div class="seen__cols">
        <div>
          <h3 class="h3">Видит</h3>
          <ul class="seen__list">
            <li>{icon('check','i i--sm')}<span><strong>Куда</strong> идёт соединение — адрес сервиса, например youtube.com</span></li>
            <li>{icon('check','i i--sm')}<span><strong>Сколько</strong> данных передано</span></li>
            <li>{icon('check','i i--sm')}<span><strong>Ваше устройство</strong> — по ключу устройства, чтобы пускать только своих</span></li>
          </ul>
        </div>
        <div>
          <h3 class="h3">Не видит</h3>
          <ul class="seen__list seen__list--no">
            <li>{icon('close','i i--sm')}<span>Что вы смотрите и открываете на странице</span></li>
            <li>{icon('close','i i--sm')}<span>Пароли, переписку и сообщения</span></li>
            <li>{icon('close','i i--sm')}<span>Файлы, платёжные данные и содержимое игр</span></li>
          </ul>
        </div>
      </div>
      <p class="seen__note">VELIQ не расшифровывает и не подменяет HTTPS. Почти все сайты и приложения сейчас работают по HTTPS; это и даёт защиту на всём пути.</p>
    </div>"""


FAQ_MAIN = [
    ("Это VPN?", "<p>Нет. VPN отправляет весь трафик через один сервер, и всё становится медленнее. VELIQ выбирает путь для каждого соединения отдельно: через свою сеть — когда так открывается доступ или выходит быстрее, обычным путём — когда так лучше. <a href=\"vpn.html\">Подробнее о разнице</a>.</p>"),
    ("Нужно ли что-то перезапускать?", "<p>Нет. Маршруты работают сразу после нажатия «Оптимизировать» — ничего закрывать и перезапускать не нужно.</p>"),
    ("Как VELIQ влияет на пинг в играх?", "<p>VELIQ сравнивает пути до игрового сервера и выбирает самый быстрый. Если путь через сеть VELIQ не быстрее обычного, маршрут не меняется. Замеры видно в программе.</p>"),
    ("Вы видите, что я открываю?", "<p>Нет. Мы не расшифровываем и не подменяем HTTPS. Сервер видит адрес назначения и объём данных — этого достаточно, чтобы переслать трафик. <a href=\"index.html#privacy\">Что видит сервер</a>. Подробности — в <a href=\"privacy.html\">политике конфиденциальности</a>.</p>"),
    ("Какие системы поддерживаются?", "<p>Сейчас VELIQ работает на компьютерах с Windows. Версия для телефона готовится.</p>"),
    ("Подходит ли VELIQ для клубов?", "<p>Да. Для клубов есть отдельный код активации и инструкция, они выдаются после заявки. <a href=\"clubs.html\">VELIQ для клубов</a>.</p>"),
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
      <p class="lede">VELIQ сам выстраивает оптимальный маршрут для каждого соединения. Заблокированные сервисы снова работают, игры получают самый быстрый путь до сервера.</p>
      <div class="hero__actions">
        <a class="btn btn--lg" href="download.html">{icon('download')}Скачать</a>
        <a class="btn btn--lg btn--ghost" href="clubs.html">Для компьютерных клубов</a>
      </div>
      <p class="hero__meta">Бета-версия</p>
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
        <div class="route__row route__row--relay"><span class="route__name">YouTube</span><span class="route__line"></span><span class="route__via" data-on="работает">—</span></div>
        <div class="route__row route__row--relay"><span class="route__name">Discord</span><span class="route__line"></span><span class="route__via" data-on="работает">—</span></div>
        <div class="route__row route__row--relay"><span class="route__name">ChatGPT</span><span class="route__line"></span><span class="route__via" data-on="работает">—</span></div>
        <div class="route__row route__row--relay"><span class="route__name">Игры</span><span class="route__line"></span><span class="route__via" data-on="лучший маршрут">—</span></div>
      </div>
      <p class="demo__hint" id="demo-hint">Нажмите кнопку — так выглядит запуск в программе.</p>
      <span class="sr-only" aria-live="polite" data-demo-live></span>
    </div>
  </div>
</section>

<section class="sec sec--tight band--sunken" id="check" data-access-check hidden>
  <div class="wrap check">
    <div>
      <h2 class="h2">Что у вас открывается прямо сейчас</h2>
      <p class="check__lead" data-check-lead>Проверяем несколько сервисов для примера…</p>
      <p class="check__note">Проверка приблизительная: браузер просто пробует достучаться до сайтов. Блокировщики рекламы и рабочие сети могут влиять на результат. Данные никуда не отправляются.</p>
    </div>
    <div class="check__list" aria-live="polite">
      <div class="check__row" data-check-url="https://www.youtube.com/favicon.ico"><span class="check__name">YouTube</span><span class="check__line"></span><span class="check__st">проверяем</span></div>
      <div class="check__row" data-check-url="https://discord.com/favicon.ico"><span class="check__name">Discord</span><span class="check__line"></span><span class="check__st">проверяем</span></div>
      <div class="check__row" data-check-url="https://web.telegram.org/favicon.ico"><span class="check__name">Telegram</span><span class="check__line"></span><span class="check__st">проверяем</span></div>
      <div class="check__row" data-check-url="https://chatgpt.com/favicon.ico"><span class="check__name">ChatGPT</span><span class="check__line"></span><span class="check__st">проверяем</span></div>
      <a class="btn btn--sm btn--ghost check__again" href="#check" data-check-again hidden>{icon('refresh','i i--sm')}Проверить ещё раз</a>
    </div>
  </div>
</section>

<section class="sec band--sunken" id="how">
  <div class="wrap">
    <div class="sec__head">
      <h2 class="h2">Три шага — и всё работает</h2>
      <p>Никаких адресов серверов и настроек. Всё это программа делает сама.</p>
    </div>
    <ol class="steps" style="list-style:none;margin:0;padding:0">
      <li class="step"><span class="step__n">1</span><h3 class="h3">Получите код активации</h3><p>Для клуба он выдаётся после заявки, для игрока — отдельно.</p></li>
      <li class="step"><span class="step__n">2</span><h3 class="h3">Установите VELIQ</h3><p>Установщик сделает всё сам.</p></li>
      <li class="step"><span class="step__n">3</span><h3 class="h3">Введите код — и всё работает</h3><p>После этого VELIQ готов к работе.</p></li>
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
    <a class="more" href="vpn.html">Чем это отличается от VPN {icon('chevron-right','i i--sm')}</a>
  </div>
</section>

<section class="sec sec--tight">
  <div class="wrap">
    <div class="facts">
      <div class="fact fact--wide"><h3>Ничего не нужно перезапускать</h3><p>Маршруты работают сразу и везде: открытые вкладки, входы в аккаунты и загрузки остаются на месте.</p></div>
      <div class="fact"><h3>Своя сеть серверов</h3><p>Собственные серверы, а не общие бесплатные узлы, где скорость делят тысячи людей.</p></div>
      <div class="fact fact--r"><h3>HTTPS не расшифровывается</h3><p>Сервер пересылает зашифрованный трафик как есть. Пароли и переписка ему не видны.</p></div>
      <div class="fact"><h3>Ничего не настраивать</h3><p>Одна кнопка. Если что-то пошло не так, программа скажет, что именно, и предложит решение.</p></div>
      <div class="fact fact--r"><h3>Готово для клубов</h3><p>VELIQ подходит для компьютерных клубов: настраивается один раз, гостю ничего делать не нужно.</p></div>
    </div>
  </div>
</section>

<section class="sec" id="privacy">
  <div class="wrap">
    {privacy_block()}
  </div>
</section>

<section class="sec band--sunken" id="games">
  <div class="wrap games">
    <div>
      <h2 class="h2">Оптимальный маршрут для игр</h2>
      <p class="lede" style="margin-top:20px">VELIQ сравнивает пути до игрового сервера и выбирает самый быстрый. Если лучше обычного пути нет, ничего не меняется.</p>
      <ul class="ticks">
        <li>{icon('check')}<span>Доступ к игровым серверам, закрытым для вашего региона</span></li>
        <li>{icon('check')}<span>Меньший пинг, если через сеть VELIQ путь короче, чем у провайдера</span></li>
        <li>{icon('check')}<span>Маршрут перестраивается сам, если какой-то путь стал хуже</span></li>
      </ul>
    </div>
    <figure class="games__panel" style="margin:0">
      <header><strong>Пинг до игрового сервера (иллюстрация)</strong><ul class="legend"><li><i class="direct"></i>обычный путь</li><li><i class="hel"></i>маршрут VELIQ</li></ul></header>
      <div data-ping-chart='{json.dumps(PING, ensure_ascii=False)}'></div>
      <footer>Иллюстрация. Настоящий замер VELIQ делает на вашем компьютере и у вашего провайдера.</footer>
    </figure>
  </div>
</section>

<section class="sec band--ink">
  <div class="wrap club-band">
    <div>
      <h2 class="h2">VELIQ для компьютерных клубов</h2>
      <p class="lede" style="margin-top:20px">Заблокированные сервисы и быстрый путь для игр на всех компьютерах клуба. Гостю ничего настраивать не нужно.</p>
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
        <h2 class="h3" style="font-size:var(--fs-3xl)">VELIQ</h2>
        <ul class="get__meta"><li>Версия {RELEASE['version']}</li></ul>
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
     "VELIQ сам строит оптимальный маршрут: заблокированные сервисы снова работают, игры получают самый быстрый путь. Для Windows 10 и 11.",
     "", index_body, [LD_APP, LD_FAQ])

# ---------------------------------------------------------------- отличие от VPN
VPN_ROWS = [
    ("Какой трафик идёт через сервер", "Весь", "Только тот, которому это нужно"),
    ("Сайты и приложения, которые и так работают", "Тоже идут через сервер и могут замедлиться", "Работают как обычно"),
    ("Онлайн-игры", "Путь через выбранный сервер, даже если он дольше", "Самый быстрый путь до игрового сервера"),
    ("Выбор сервера", "Обычно вручную", "Сам, для каждого соединения"),
    ("Если сервер перестал отвечать", "Пропадает всё или нужно переключаться", "Перестраивает маршрут через другой"),
    ("Что нужно сделать", "Выбрать сервер, иногда перезапустить приложения", "Нажать одну кнопку"),
]
vpn_table = "\n".join(f'<tr><th scope="row">{a}</th><td data-l="Обычный VPN">{b}</td><td data-l="VELIQ" class="vs__us">{c}</td></tr>' for a, b, c in VPN_ROWS)
vpn_body = f"""
<section class="page-head">
  <div class="wrap">
    <div class="crumbs"><a href="index.html">Главная</a> / Отличие от VPN</div>
    <h1 class="h1" style="max-width:16ch">Чем VELIQ отличается от VPN</h1>
    <p class="lede">VPN отправляет весь трафик через один сервер. VELIQ выбирает путь для каждого соединения отдельно — поэтому открывается заблокированное, а остальное не замедляется.</p>
  </div>
</section>

<section class="sec">
  <div class="wrap">
    <div class="flows">
      <figure class="flows__panel">
        <figcaption><h2 class="h3">Обычный VPN</h2><span class="faint">один путь для всего</span></figcaption>
        {flow('vpn')}
      </figure>
      <figure class="flows__panel flows__panel--main">
        <figcaption><h2 class="h3">VELIQ</h2><span class="faint">свой путь для каждого соединения</span></figcaption>
        {flow('veliq')}
      </figure>
    </div>
  </div>
</section>

<section class="sec band--sunken">
  <div class="wrap">
    <div class="sec__head"><h2 class="h2">По пунктам</h2><p>Как это ощущается в обычный день: в браузере, в игре, когда сервер перестал отвечать.</p></div>
    <div class="vs-wrap">
      <table class="vs">
        <thead><tr><th scope="col"><span class="sr-only">Что сравниваем</span></th><th scope="col">Обычный VPN</th><th scope="col" class="vs__us">VELIQ</th></tr></thead>
        <tbody>
{vpn_table}
        </tbody>
      </table>
    </div>
    <p class="faint" style="font-size:var(--fs-sm);margin:var(--sp-4) 0 0">Сравнение с типичным VPN-приложением. Некоторые VPN умеют исключать отдельные приложения, но это настраивается вручную.</p>
  </div>
</section>

<section class="sec">
  <div class="wrap">
    {privacy_block()}
  </div>
</section>

<section class="sec band--ink">
  <div class="wrap club-band">
    <div>
      <h2 class="h2">Нажал — работает.</h2>
      <p class="lede" style="margin-top:20px">Никаких серверов на выбор и настроек. VELIQ строит маршруты сам.</p>
      <div class="hero__actions" style="margin-top:32px"><a class="btn btn--lg" href="download.html">{icon('download')}Скачать</a><a class="btn btn--lg btn--ghost" href="support.html">Вопросы и ответы</a></div>
    </div>
  </div>
</section>
"""
page("vpn.html", "Чем VELIQ отличается от VPN",
     "VPN отправляет весь трафик через один сервер. VELIQ выбирает путь для каждого соединения: заблокированное открывается, игры получают самый быстрый маршрут, остальное работает как обычно.",
     "how", vpn_body, og="og-vpn.png", og_alt="Не VPN: VELIQ строит свой путь для каждого соединения.")

# ---------------------------------------------------------------- скачать
download_body = f"""
<section class="page-head">
  <div class="wrap">
    <div class="crumbs"><a href="index.html">Главная</a> / Скачать</div>
    <h1 class="h1">Скачать VELIQ</h1>
    <p class="lede">Установщик сделает всё сам. После установки введите код активации.</p>
    <div class="platforms" role="group" aria-label="Устройство">
      <button type="button" class="platforms__btn" data-platform-btn="win" aria-pressed="true">Компьютер</button>
      <button type="button" class="platforms__btn" data-platform-btn="phone" aria-pressed="false">Телефон<span class="platforms__soon">скоро</span></button>
    </div>
  </div>
</section>
<section class="sec sec--tight" data-platform="phone" hidden>
  <div class="wrap">
    <div class="soon">
      <img class="get__icon" src="assets/img/app-icon.svg" alt="" width="88" height="88">
      <div>
        <h2 class="h3" style="font-size:var(--fs-3xl)">VELIQ для телефона готовится</h2>
        <p class="muted" style="margin:10px 0 0;max-width:52ch">Тот же принцип: одна кнопка и свой путь для каждого соединения. Оставьте почту — напишем, когда приложение выйдет. А пока VELIQ можно поставить на компьютер.</p>
      </div>
      <div class="soon__actions">
        <a class="btn btn--lg" href="mailto:{SITE['email']}?subject=Сообщить%20о%20выходе%20VELIQ%20для%20телефона">Сообщить о выходе</a>
        <button type="button" class="btn btn--lg btn--ghost" data-platform-btn="win" aria-pressed="false">Версия для компьютера</button>
      </div>
    </div>
  </div>
</section>
<section class="sec sec--tight" data-platform="win">
  <div class="wrap cols">
    <div>
      <div class="get" style="margin-bottom:var(--sp-7)">
        <img class="get__icon" src="assets/img/app-icon.svg" alt="" width="88" height="88">
        <div>
          <h2 class="h3" style="font-size:var(--fs-3xl)">VELIQ {RELEASE['version']}</h2>
          <ul class="get__meta"><li>Бета-версия</li></ul>
        </div>
        <a class="btn btn--lg" href="{RELEASE['url']}" download>{icon('download')}Скачать установщик</a>
      </div>

      <h2 class="h2" id="install" style="font-size:var(--fs-4xl);margin-bottom:var(--sp-5)">Как установить</h2>
      <ol class="olist">
        <li><div><h3>Получите код активации</h3><p>Код привязывает VELIQ к вам. Для клуба он выдаётся после заявки, для игрока — отдельно.</p></div></li>
        <li><div><h3>Установите VELIQ</h3><p>Скачайте и запустите установщик. Он сам сделает всё нужное, дополнительных шагов не требуется.</p></div></li>
        <li><div><h3>Введите код активации</h3><p>После этого VELIQ готов к работе.</p></div></li>
      </ol>
      <p class="install-more">Для клубов есть отдельная инструкция, она выдаётся вместе с кодом активации.</p>
      <p class="install-more">Подробные инструкции и ответы на частые вопросы скоро появятся в вики VELIQ.</p>
    </div>
    <aside class="aside">
      <div class="card">
        <h2 class="h3" style="margin-bottom:16px">О файле</h2>
        <dl class="kv">
          <dt>Версия</dt><dd>{RELEASE['version']}</dd>
          <dt>Система</dt><dd>{RELEASE['os']}</dd>
        </dl>
      </div>
    </aside>
  </div>
</section>
"""
page("download.html", "Скачать VELIQ",
     f"Скачайте VELIQ {RELEASE['version']}. Установщик сделает всё сам, после установки введите код активации.",
     "", download_body, LD_APP)

# ---------------------------------------------------------------- клубы
CLUB_FAQ = [
    ("Нужно ли что-то делать гостю?", "<p>Нет. Гостю ничего настраивать не нужно.</p>"),
    ("Как VELIQ влияет на пинг в играх?", "<p>VELIQ выбирает самый быстрый путь до игрового сервера. Если обычный путь быстрее, маршрут не меняется.</p>"),
    ("Подходит ли VELIQ для моей системы управления клубом?", "<p>Напишите нам, подскажем.</p>"),
    ("Сколько это стоит?", "<p>Тариф для клубов считается за компьютер в месяц. Цена появится к релизу — оставьте заявку, и мы сообщим первыми. Пока можно прикинуть выгоду <a href=\"#calc\">в калькуляторе</a>.</p>"),
]
club_calc = """
<div class="calc" data-calc>
  <div class="calc__in">
    <div class="calc__row"><label for="c-pcs">Компьютеров в клубе</label><input class="input" id="c-pcs" data-k="pcs" inputmode="numeric" value="20"><input type="range" data-k="pcs" min="5" max="200" step="1" value="20" aria-label="Компьютеров в клубе"></div>
    <div class="calc__row"><label for="c-hours">Часов работы в сутки</label><input class="input" id="c-hours" data-k="hours" inputmode="numeric" value="24"><input type="range" data-k="hours" min="6" max="24" step="1" value="24" aria-label="Часов работы в сутки"></div>
    <div class="calc__row"><label for="c-load">Средняя загрузка, %</label><input class="input" id="c-load" data-k="load" inputmode="numeric" value="50"><input type="range" data-k="load" min="5" max="100" step="1" value="50" aria-label="Средняя загрузка"></div>
    <div class="calc__row"><label for="c-price">Цена часа, ₽</label><input class="input" id="c-price" data-k="price" inputmode="numeric" value="150"><input type="range" data-k="price" min="40" max="500" step="10" value="150" aria-label="Цена часа"></div>
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
    <div class="hero__actions" style="margin-top:32px"><a class="btn btn--lg" href="#apply">Оставить заявку</a><a class="btn btn--lg btn--ghost" href="#calc">Посчитать выгоду</a></div>
  </div>
</section>

<section class="sec">
  <div class="wrap">
    <div class="sec__head"><h2 class="h2">Что теряет клуб</h2><p>Каждая такая ситуация — гость, который сегодня ушёл раньше, а завтра выбрал другой клуб.</p></div>
    <ol class="pains">
      <li><span class="pains__n">01</span><h3 class="h3">Сервисы не открываются</h3><p>YouTube, Discord, Telegram, нейросети — гостю нечем заняться между матчами и не с кем говорить в голосовом чате.</p></li>
      <li><span class="pains__n">02</span><h3 class="h3">Игры лагают</h3><p>Высокий пинг или закрытые игровые серверы — команда доигрывает и уходит туда, где играется лучше.</p></li>
      <li><span class="pains__n">03</span><h3 class="h3">Администратор занят не тем</h3><p>Смена уходит на «почему не работает», а не на гостей.</p></li>
    </ol>
  </div>
</section>

<section class="sec band--sunken">
  <div class="wrap">
    <div class="sec__head"><h2 class="h2">Что меняется с VELIQ</h2><p>Одна программа на каждом ПК — и маршруты строятся сами.</p></div>
    <div class="facts">
      <div class="fact"><h3>Гостю ничего не нужно делать</h3><p>VELIQ настраивается один раз, дальше гость просто садится и играет.</p></div>
      <div class="fact fact--r"><h3>Надёжно в клубе</h3><p>Гость не может изменить настройки или навредить программе.</p></div>
      <div class="fact"><h3>Лучший путь для игр</h3><p>VELIQ выбирает самый быстрый маршрут до игрового сервера и открывает доступ к закрытым.</p></div>
      <div class="fact fact--r"><h3>Заблокированное снова работает</h3><p>Стримы, голосовые чаты и нейросети открываются у гостя сразу, без настроек.</p></div>
    </div>
  </div>
</section>

<section class="sec">
  <div class="wrap">
    <div class="sec__head"><h2 class="h2">Кто что делает</h2><p>Вся настройка — один раз и у администратора. Гость просто садится играть.</p></div>
    <div class="roles">
      <div class="roles__col">
        <div class="roles__who">{icon('settings')}<h3 class="h3">Администратор</h3></div>
        <ul class="roles__list">
          <li>Устанавливает VELIQ и вводит код активации клуба</li>
          <li>Если что-то не так — пишет нам, поможем</li>
        </ul>
      </div>
      <div class="roles__col roles__col--guest">
        <div class="roles__who">{icon('games')}<h3 class="h3">Гость</h3></div>
        <ul class="roles__list">
          <li>Садится за ПК — VELIQ уже работает</li>
          <li>Видит только, что VELIQ работает</li>
          <li>Не может изменить настройки или что-то сломать</li>
        </ul>
      </div>
    </div>
  </div>
</section>

<section class="sec band--sunken" id="setup">
  <div class="wrap cols">
    <div>
      <h2 class="h2" style="margin-bottom:var(--sp-6)">Как подключить клуб</h2>
      <ol class="olist">
        <li><div><h3>Получите код активации</h3><p>Для клуба он выдаётся после заявки. Вместе с кодом вы получите инструкцию для клуба.</p></div></li>
        <li><div><h3>Установите VELIQ</h3><p>Установщик сам сделает всё нужное, дополнительных шагов не требуется.</p></div></li>
        <li><div><h3>Введите код активации</h3><p>После этого VELIQ готов к работе.</p></div></li>
      </ol>
      <p class="install-more">Подробные инструкции и ответы на частые вопросы скоро появятся в вики VELIQ.</p>
    </div>
    <aside class="aside">
      <div class="note">{icon('info')}<div><strong>Нужна помощь?</strong> Напишите на <a href="mailto:{SITE['email']}">{SITE['email']}</a> — расскажем, как подключить клуб, и пришлём код активации.</div></div>
    </aside>
  </div>
</section>

<section class="sec" id="calc">
  <div class="wrap">
    <div class="sec__head"><h2 class="h2">Сколько часов возвращает VELIQ</h2><p>Подставьте цифры своего клуба. Все значения можно менять.</p></div>
    {club_calc}
  </div>
</section>

<section class="sec band--ink" id="apply">
  <div class="wrap club-band">
    <div>
      <h2 class="h2">Поставим на один ПК — решите по факту</h2>
      <p class="lede" style="margin-top:20px">Напишите, сколько у вас компьютеров и какая система управления. Пришлём код активации и поможем с первой установкой.</p>
      <div class="hero__actions" style="margin-top:32px"><a class="btn btn--lg" href="mailto:{SITE['email']}?subject=VELIQ%20для%20клуба">Оставить заявку</a><a class="btn btn--lg btn--ghost" href="pricing.html">Тарифы</a></div>
    </div>
    <div class="club-band__pcs" aria-hidden="true">{''.join('<i></i>' for _ in range(27))}</div>
  </div>
</section>

<section class="sec">
  <div class="wrap">
    <div class="sec__head"><h2 class="h2">Вопросы владельцев клубов</h2><p>Презентацию и инструкцию для клуба пришлём по запросу.</p></div>
    <div class="faq">{faq(CLUB_FAQ)}</div>
  </div>
</section>
"""
page("clubs.html", "VELIQ для компьютерных клубов",
     "Заблокированные сервисы и быстрый путь для игр на всех ПК клуба. Калькулятор выгоды.",
     "clubs", clubs_body, og="og-clubs.png", og_alt="VELIQ для компьютерных клубов: каждый ПК — работает.")

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
        <p class="muted" style="margin:0">Для компьютерных клубов.</p>
        <ul><li>{icon('check')}Всё из тарифа «Игрок»</li><li>{icon('check')}Код активации для всех ПК клуба</li><li>{icon('check')}Гость не может изменить настройки</li><li>{icon('check')}Помощь с установкой и кодом активации</li></ul>
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
    ("VELIQ не может запуститься", "Программа не может начать работу.", "Закройте VELIQ и откройте заново. Если не помогло — переустановите VELIQ или напишите нам."),
    ("Нет связи с сервером", "Серверы VELIQ не отвечают с вашего подключения.", "Проверьте, что интернет работает без VELIQ. Затем нажмите «Повторить» — программа попробует другой сервер. Если не помогло, напишите нам: возможно, провайдер блокирует соединение."),
    ("Часть сайтов не открывается", "Для сайтов маршрут пока не построен, а приложения работают.", "Нажмите «Повторить». Если компьютером управляет организация (рабочий компьютер), её правила могут запрещать изменения — тогда маршруты работают только для приложений."),
    ("Код активации не принят", "Сервер не принял код активации для этого устройства.", "Введите код ещё раз. Если не помогло — напишите нам."),
    ("Не хватает прав", "Действие требует прав администратора.", "Подтвердите запрос Windows."),
]
# Постоянные адреса ошибок: support.html#err-<ключ>. Ключи совпадают с кодами ошибок в app-ui/app.js — программа может открывать нужный ответ.
ERR_IDS = ["service", "relay", "browser", "identity", "rights"]
err_html = "\n".join(
    f'<details id="err-{k}"><summary>{t}{icon("chevron-down")}</summary><div class="a"><p><strong>Что случилось.</strong> {w}</p><p><strong>Что сделать.</strong> {f}</p></div></details>'
    for k, (t, w, f) in zip(ERR_IDS, ERRORS))
SUPPORT_FAQ = FAQ_MAIN + [
    ("Как выключить VELIQ?", "<p>Нажмите большую кнопку ещё раз. Интернет продолжит работать как без программы.</p>"),
    ("Как добавить свой сайт или приложение?", "<p>Вкладка «Сервисы» → «Добавить». VELIQ подберёт для него маршрут так же, как для остальных.</p>"),
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
        <h2 class="faq__group">Общие вопросы</h2>
{faq(SUPPORT_FAQ)}
        <h2 class="faq__group" id="errors">Ошибки в программе</h2>
{err_html}
      </div>
    </div>
    <aside class="aside">
      <div class="card">
        <h2 class="h3" style="margin-bottom:10px">Не помогло?</h2>
        <p class="muted" style="margin:0 0 16px;font-size:var(--fs-sm)">Напишите нам, что случилось, — поможем.</p>
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
<p>Сайт не использует счётчики и рекламные cookie. Выбор темы сохраняется только в вашем браузере. Проверка «Что у вас открывается» на главной отправляет пустые запросы к сайтам сервисов прямо из вашего браузера; её результат никуда не передаётся.</p>
<h2>Связь</h2>
<p>Вопросы о данных — на <a href="mailto:{SITE['email']}">{SITE['email']}</a>.</p>
</div></section>
"""
page("privacy.html", "Политика конфиденциальности", "Какие данные получает VELIQ: HTTPS не расшифровывается, содержимое страниц серверу не видно.", "", privacy_body)

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

# ---------------------------------------------------------------- план развития
# Публичная версия docs/ROADMAP.md: без адресов серверов, кодов, внутренних сроков и названий клубов.
# Статусы менять здесь вместе с docs/ROADMAP.md и запускать генератор.
ROADMAP_UPDATED = "25 октября 2026"
ROADMAP = [
    ("Готово", "tag--ok", "check", "Работает в бета-версии", [
        "Заблокированные сервисы открываются через собственную сеть серверов VELIQ: Telegram, ChatGPT, Claude и другие",
        "YouTube в высоком качестве, вплоть до 4K",
        "Обычные сайты идут напрямую, без лишних задержек",
        "Браузеры подхватывают маршруты сразу — ничего не нужно перезапускать",
        "Собственные серверы в Европе с быстрым каналом",
    ]),
    ("В работе", "tag--warn", "refresh", "Делаем сейчас", [
        "Обход блокировок прямо на компьютере, без сервера: YouTube и Discord быстрее и стабильнее",
        "Игры: PUBG, Valorant и FACEIT — проверяем в реальных матчах",
        "Автоматическое переключение на резервный сервер",
        "Установщик одним файлом",
    ]),
    ("Дальше", "", "route", "Следующие шаги", [
        "Версия для клубов: один код активации на все компьютеры и установка в образ",
        "Пилот в компьютерном клубе",
        "Тарифы и оплата",
        "Версия для телефона",
    ]),
]
roadmap_cols = "\n".join(
    f'''      <div class="plan{' plan--main' if i == 1 else ''}">
        <span class="tag {cls}" style="justify-self:start">{label}</span>
        <h2 class="h3">{title}</h2>
        <ul>{"".join(f"<li>{icon(ic)}{item}</li>" for item in items)}</ul>
      </div>'''
    for i, (label, cls, ic, title, items) in enumerate(ROADMAP))
roadmap_body = f"""
<section class="page-head">
  <div class="wrap">
    <div class="crumbs"><a href="index.html">Главная</a> / План развития</div>
    <h1 class="h1">План развития</h1>
    <p class="lede">Что VELIQ уже умеет, над чем работаем сейчас и что будет дальше.</p>
  </div>
</section>
<section class="sec sec--tight">
  <div class="wrap">
    <div class="note" style="margin-bottom:var(--sp-6)">{icon('info')}<div>Обновлено {ROADMAP_UPDATED}. Сроки не объявляем: пункт переходит в «Готово» только после проверки на реальных компьютерах и у реальных провайдеров.</div></div>
    <div class="plans">
{roadmap_cols}
    </div>
    <div class="hero__actions" style="margin-top:32px"><a class="btn btn--lg" href="download.html">{icon('download')}Скачать бету</a><a class="btn btn--lg btn--ghost" href="clubs.html">Для компьютерных клубов</a></div>
  </div>
</section>
"""
page("roadmap.html", "План развития", "Что VELIQ уже умеет, над чем работаем сейчас и что будет дальше.", "", roadmap_body)

# ---------------------------------------------------------------- служебные файлы
pages = ["", "vpn.html", "download.html", "clubs.html", "pricing.html", "support.html", "privacy.html", "terms.html", "roadmap.html"]
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
