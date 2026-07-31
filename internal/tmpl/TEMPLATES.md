# Шаблоны srekit — справочник

Эта папка — твой кастомный набор шаблонов для `srekit`. Файлы здесь
переопределяют встроенные. Если файл отсутствует — `srekit` берёт встроенную
версию (прозрачный fallback). Можно подменять как все шаблоны, так и
точечно — например, только `postmortem.yaml`.

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

Флага `--template FILE` нет ни у одной команды: генераторы читают
`<name>.yaml` из этой директории, поэтому «подмена на один запуск» делается
отдельным `--templates-dir`.

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
srekit templates validate          # парсит + dry-run рендерит каждый артефакт
srekit templates diff              # diff против embedded версий
srekit templates diff --name-only  # только список изменённых файлов
```

`validate` ловит опечатки в полях (`{{ .Meta.Servce }}` вместо `.Meta.Service`)
и синтаксические ошибки шаблона. Запускай после правок и после
`srekit templates pull` — особенно полезно после обновления бинарника,
потому что у новой версии могут быть новые/переименованные поля.

`diff` показывает, что у тебя своё, а что отстало от апстрима.

## Синтаксис шаблонов

`srekit` использует Go-шаблоны
([`text/template`](https://pkg.go.dev/text/template)). Доступны:

- **Стандартные конструкции**: `{{ .Meta.Field }}`, `{{ if .Meta.X }}…{{ end }}`,
  pipes `{{ .Meta.Field | func }}`.

Метаданные команды лежат под `.Meta` — в v1-артефакте контекст рендера это
`{ Meta: <поля команды> }`. Список полей по каждому артефакту — ниже.
- **YAML frontmatter** между `---` блоками — необязателен, но именно его
  парсят инструменты типа Obsidian / dataview.

## Доступные функции (FuncMap)

| Функция | Описание | Пример |
|---------|----------|--------|
| `default` | Дефолт для пустой строки | `{{ .Meta.Owner \| default "<owner>" }}` |
| `shortID` | Первые N символов строки (для коротких ID) | `{{ shortID .Meta.ID 8 }}` |
| `slugify` | Превратить строку в slug | `{{ slugify .Meta.Title }}` |
| `now` | Текущее время; формат опционален | `{{ now }}` или `{{ now "2006-01-02" }}` |
| `upper` | В верхний регистр | `{{ "abc" \| upper }}` |
| `lower` | В нижний регистр | `{{ "ABC" \| lower }}` |
| `trim` | Убрать пробелы по краям | `{{ "  x  " \| trim }}` |

`now` использует injectable wall clock `internal/clock.Now` — в тестах
переопределяется, что делает поведение шаблонов детерминированным.

## Доступные плейсхолдеры по шаблонам

### `task.yaml` — investigation log

| Поле | Тип | Описание |
|------|-----|----------|
| `.Meta.ID` | string | UUID |
| `.Meta.Title` | string | Заголовок расследования |
| `.Meta.CreationDate` | string | RFC3339 |
| `.Meta.ModificationDate` | string | RFC3339, при инициализации = CreationDate |

### `postmortem.yaml`

| Поле | Тип | Описание |
|------|-----|----------|
| `.Meta.ID` | string | UUID |
| `.Meta.Title` | string | Заголовок |
| `.Meta.Severity` | string | SEV-1, SEV-2, … |
| `.Meta.Start` | string | Начало инцидента (может быть пустой) |
| `.Meta.End` | string | Конец инцидента (может быть пустой) |
| `.Meta.Owner` | string | Ответственный (может быть пустой) |
| `.Meta.Now` | string | RFC3339, время создания доки |

### `runbook.yaml`

| Поле | Тип | Описание |
|------|-----|----------|
| `.Meta.ID` | string | UUID |
| `.Meta.Title` | string | Заголовок |
| `.Meta.Service` | string | Имя сервиса (может быть пустым) |
| `.Meta.Alert` | string | Имя алерта (может быть пустым) |
| `.Meta.Now` | string | RFC3339 |

### `rfc.yaml`

| Поле | Тип | Описание |
|------|-----|----------|
| `.Meta.ID` | string | UUID |
| `.Meta.Title` | string | Заголовок |
| `.Meta.Status` | string | proposed / accepted / rejected / superseded / deprecated |
| `.Meta.Now` | string | RFC3339 |
| `.Meta.Author.Name` | string | Имя автора |
| `.Meta.Author.Email` | string | E-mail автора |

### `slo.yaml`

| Поле | Тип | Описание |
|------|-----|----------|
| `.Meta.ID` | string | UUID |
| `.Meta.Service` | string | Имя сервиса |
| `.Meta.Target` | string | SLO-таргет, например `99.9%` |
| `.Meta.Window` | string | Окно, например `30d` |
| `.Meta.LatencyTarget` | string | Latency-таргет, например `300ms` |
| `.Meta.Now` | string | RFC3339 |

### `ebp.yaml` — Error Budget Policy

| Поле | Тип | Описание |
|------|-----|----------|
| `.Meta.ID` | string | UUID |
| `.Meta.Service` | string | Имя сервиса |
| `.Meta.Now` | string | RFC3339 |

### `oncall.yaml`

| Поле | Тип | Описание |
|------|-----|----------|
| `.Meta.ID` | string | UUID |
| `.Meta.Team` | string | Имя команды |
| `.Meta.Start` | string | Начало периода (YYYY-MM-DD) |
| `.Meta.End` | string | Конец периода (YYYY-MM-DD) |
| `.Meta.Now` | string | RFC3339 |
| `.Meta.Author.Name` | string | Дежурный |
| `.Meta.Author.Email` | string | E-mail дежурного |

### `changelog.yaml`

| Поле | Тип | Описание |
|------|-----|----------|
| `.Meta.Today` | string | Дата (YYYY-MM-DD) |
| `.Meta.Repo` | string | OWNER/REPO для compare-ссылок |
| `.Meta.InitialVersion` | string | Версия первого релиза |

## Что лучше НЕ ломать

- **Имена файлов**: `srekit` ищет артефакты строго по именам
  (`postmortem.yaml`, `runbook.yaml`, и т.д.). Переименуешь — override
  не сработает.
- **Имена полей**: переименовывать `.Meta.Field` смысла нет — Go-шаблон
  упадёт с ошибкой при рендере, если упомянуто несуществующее поле.
- **`version: 1`** в начале файла: артефакт с другой версией не парсится.
- **`id` секций**: по ним матчатся `--json` / `--from` и required-проверки.
  Заголовок (`title`) меняй свободно, `id` — нет.

## Что можно безболезненно менять

- Весь body шаблона: заголовки, секции, текст, frontmatter.
- Стиль frontmatter (Obsidian / dataview / etc).
- Локализацию (английский / русский / любой другой).
- Корпоративные плейсхолдеры в italic (`_заполнить_`, `<TBD>`, и т.п.).
- Подстановки через FuncMap.
