# Шаблоны srekit — справочник

Эта папка — твой кастомный набор шаблонов для `srekit`. Файлы здесь
переопределяют встроенные. Если файл отсутствует — `srekit` берёт встроенную
версию (прозрачный fallback). Можно подменять как все шаблоны, так и
точечно — например, только `postmortem.md.tmpl`.

## Как подключить эту директорию к srekit

Один из вариантов:

```bash
# Флагом (на одну команду):
srekit --templates-dir ~/.srekit/templates postmortem --title "X" --stdout

# Через переменную окружения:
export SREKIT_TEMPLATES_DIR=~/.srekit/templates

# Через ~/.srekit.yaml:
echo 'templates_dir: ~/.srekit/templates' >> ~/.srekit.yaml
```

Точечная подмена одного шаблона на одну команду:

```bash
srekit runbook --title "p99 spike" --template ./my-runbook.tmpl --stdout
```

## Версионирование через git

Эта папка инициализирована как git-репозиторий. Подключи свой remote и
синхронизируй между машинами/командой:

```bash
cd ~/.srekit/templates
git remote add origin git@github.com:<owner>/<repo>.git
git add . && git commit -m "initial templates"
git push -u origin main
```

Чтобы стянуть обновления от команды:

```bash
srekit templates pull              # git pull --ff-only в configured templates_dir
srekit templates pull --rebase     # если есть локальные коммиты
```

Авто-pull при каждом запуске `srekit` намеренно не делается — это убивает
UX и работу в офлайне. Команда вызывается явно.

## Проверить и сравнить

```bash
srekit templates validate          # парсит + дry-run рендерит каждый .tmpl
srekit templates diff              # diff против embedded версий
srekit templates diff --name-only  # только список изменённых файлов
```

`validate` ловит опечатки в `.Field` (`{{ .Servce }}` вместо `.Service`)
и синтаксические ошибки шаблона. Запускай после правок и после
`srekit templates pull` — особенно полезно после обновления бинарника,
потому что у новой версии могут быть новые/переименованные поля.

`diff` показывает, что у тебя своё, а что отстало от апстрима.

## Синтаксис шаблонов

`srekit` использует Go-шаблоны
([`text/template`](https://pkg.go.dev/text/template)). Доступны:

- **Стандартные конструкции**: `{{ .Field }}`, `{{ if .X }}…{{ end }}`,
  `{{ range .Items }}…{{ end }}`, pipes `{{ .Field | func }}`.
- **YAML frontmatter** между `---` блоками — необязателен, но именно его
  парсят инструменты типа Obsidian / dataview.

## Доступные функции (FuncMap)

| Функция | Описание | Пример |
|---------|----------|--------|
| `default` | Дефолт для пустой строки | `{{ .Owner \| default "<owner>" }}` |
| `shortID` | Первые N символов строки (для коротких ID) | `{{ shortID .ID 8 }}` |
| `slugify` | Превратить строку в slug | `{{ slugify .Title }}` |
| `now` | Текущее время; формат опционален | `{{ now }}` или `{{ now "2006-01-02" }}` |
| `upper` | В верхний регистр | `{{ "abc" \| upper }}` |
| `lower` | В нижний регистр | `{{ "ABC" \| lower }}` |
| `trim` | Убрать пробелы по краям | `{{ "  x  " \| trim }}` |

`now` использует injectable wall clock `internal/clock.Now` — в тестах
переопределяется, что делает поведение шаблонов детерминированным.

## Доступные плейсхолдеры по шаблонам

### `task.md.tmpl` — investigation log

| Поле | Тип | Описание |
|------|-----|----------|
| `.ID` | string | UUID |
| `.Title` | string | Заголовок расследования |
| `.CreationDate` | string | RFC3339 |
| `.ModificationDate` | string | RFC3339, при инициализации = CreationDate |

### `postmortem.md.tmpl`

| Поле | Тип | Описание |
|------|-----|----------|
| `.ID` | string | UUID |
| `.Title` | string | Заголовок |
| `.Severity` | string | SEV-1, SEV-2, … |
| `.Start` | string | Начало инцидента (может быть пустой) |
| `.End` | string | Конец инцидента (может быть пустой) |
| `.Owner` | string | Ответственный (может быть пустой) |
| `.Now` | string | RFC3339, время создания доки |

### `runbook.md.tmpl`

| Поле | Тип | Описание |
|------|-----|----------|
| `.ID` | string | UUID |
| `.Title` | string | Заголовок |
| `.Service` | string | Имя сервиса (может быть пустым) |
| `.Alert` | string | Имя алерта (может быть пустым) |
| `.Now` | string | RFC3339 |

### `rfc.md.tmpl`

| Поле | Тип | Описание |
|------|-----|----------|
| `.ID` | string | UUID |
| `.Title` | string | Заголовок |
| `.Status` | string | proposed / accepted / rejected / superseded / deprecated |
| `.Now` | string | RFC3339 |
| `.Author.Name` | string | Имя автора |
| `.Author.Email` | string | E-mail автора |

### `slo.md.tmpl`

| Поле | Тип | Описание |
|------|-----|----------|
| `.ID` | string | UUID |
| `.Service` | string | Имя сервиса |
| `.Target` | string | SLO-таргет, например `99.9%` |
| `.Window` | string | Окно, например `30d` |
| `.LatencyTarget` | string | Latency-таргет, например `300ms` |
| `.Now` | string | RFC3339 |

### `ebp.md.tmpl` — Error Budget Policy

| Поле | Тип | Описание |
|------|-----|----------|
| `.ID` | string | UUID |
| `.Service` | string | Имя сервиса |
| `.Now` | string | RFC3339 |

### `capacity.md.tmpl`

| Поле | Тип | Описание |
|------|-----|----------|
| `.ID` | string | UUID |
| `.Service` | string | Имя сервиса |
| `.Horizon` | string | Горизонт планирования, например `1y` |
| `.Now` | string | RFC3339 |

### `oncall.md.tmpl`

| Поле | Тип | Описание |
|------|-----|----------|
| `.ID` | string | UUID |
| `.Team` | string | Имя команды |
| `.Start` | string | Начало периода (YYYY-MM-DD) |
| `.End` | string | Конец периода (YYYY-MM-DD) |
| `.Now` | string | RFC3339 |
| `.Author.Name` | string | Дежурный |
| `.Author.Email` | string | E-mail дежурного |

### `retro.md.tmpl`

| Поле | Тип | Описание |
|------|-----|----------|
| `.ID` | string | UUID |
| `.Team` | string | Имя команды |
| `.Sprint` | string | Метка спринта/периода |
| `.Now` | string | RFC3339 |

### `changelog.md.tmpl`

| Поле | Тип | Описание |
|------|-----|----------|
| `.Today` | string | Дата (YYYY-MM-DD) |
| `.Repo` | string | OWNER/REPO для compare-ссылок |
| `.InitialVersion` | string | Версия первого релиза |

### `license_*.tmpl`

| Поле | Тип | Описание |
|------|-----|----------|
| `.Year` | int | Год |
| `.Author.Name` | string | Имя автора |
| `.Author.Email` | string | E-mail автора |

## Что лучше НЕ ломать

- **Имена файлов**: `srekit` ищет шаблоны строго по именам
  (`postmortem.md.tmpl`, `runbook.md.tmpl`, и т.д.). Переименуешь —
  override не сработает.
- **Имена полей**: переименовывать `.Field` смысла нет — Go-шаблон упадёт
  с ошибкой при рендере, если упомянуто несуществующее поле.

## Что можно безболезненно менять

- Весь body шаблона: заголовки, секции, текст, frontmatter.
- Стиль frontmatter (Obsidian / dataview / etc).
- Локализацию (английский / русский / любой другой).
- Корпоративные плейсхолдеры в italic (`_заполнить_`, `<TBD>`, и т.п.).
- Подстановки через FuncMap.
