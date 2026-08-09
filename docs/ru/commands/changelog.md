# srekit changelog

Скаффолд `CHANGELOG.md` в формате [Keep a Changelog](https://keepachangelog.com/en/1.1.0/). Автодетектит GitHub репо из `git config remote.origin.url`.

Голый вызов генерирует. Две подкоманды обслуживают уже существующий changelog: [`release`](#release) режет версию, [`validate`](#validate) линтит документ. Собственными записями каталога они не являются — `srekit changelog` ведёт себя ровно так же, как вёл всегда.

## Синопсис

```bash
srekit changelog [flags]
srekit changelog release --version X.Y.Z [FILE] [flags]
srekit changelog validate [FILE]
```

## Флаги

| Флаг | Обязательный | Описание |
|---|---|---|
| `--repo` | нет | `<owner>/<name>` slug. Если не передан — srekit берёт `meta.repo` из payload'а `--from`, а если и его нет — читает `git config remote.origin.url` и парсит GitHub SSH или HTTPS URL'ы. |
| `--version` | нет | Начальный version anchor (например `0.1.0`). Default: `0.1.0`. |
| `--from` | нет | Читать тела секций из JSON-файла; `-` читает stdin. |

Плюс [общие output-флаги](index.md#shared-output-flags). Default имя файла: `CHANGELOG.md`.

## Примеры

Внутри git-репо с `origin`-remote на GitHub:

```bash
srekit changelog --out CHANGELOG.md
# репо детектится из git remote, версия default 0.1.0
```

Явно:

```bash
srekit changelog --repo jtprogru/srekit --version 0.1.0 --out CHANGELOG.md
```

Вне git-репо без `--repo` — ошибка (никакого silent `OWNER/REPO` плейсхолдера, который кусал юзеров в v0.2):

```bash
srekit changelog --stdout
# Error: could not detect repo from git remote — pass --repo OWNER/NAME
```

## Структурированный ввод

`--json` отдаёт документ в виде `{meta, sections}`, а `--from` принимает эту же форму обратно — тела секций можно заполнить скриптом или агентом, а не руками:

```bash
srekit changelog --repo acme/api --json > cl.json
# ...заменить тело секции "unreleased"...
srekit changelog --from cl.json
```

Переданные тела секций вставляются verbatim, без вычисления шаблонов, поэтому markdown с `{{ }}` проходит round-trip без изменений. Пропущенные секции берут дефолты из артефакта. Неизвестный артефакту section id — ошибка с именем нарушителя, а не молчаливый пропуск.

`meta` в payload'е задаёт `repo`, `initialVersion` и `today`. Флаг выигрывает у файла, файл — у git remote, поэтому payload с `meta.repo` рендерится и вне git-репозитория.

В отличие от [`srekit postmortem`](postmortem.md), у `changelog` нет `--schema` и `--validate`: его артефакт не объявляет обязательных секций, поэтому валидация payload'а не может завершиться неудачей, а схема из двух строковых полей говорит меньше, чем сам `--json`.

## Вывод

Скаффолд включает скелет `[Unreleased]` / `[<version>]` с шестью подсекциями Keep a Changelog и заканчивается блоком link reference definitions на `github.com/<repo>/compare/v<version>...HEAD`.

Этот блок ссылок — футер уровня документа, а не часть тела последней секции. Поэтому он переживает payload `--from`, заменяющий `initial_release`, и именно его переписывает [`changelog release`](#release).

## Как резать релиз { #release }

`srekit changelog release --version X.Y.Z` переносит всё из-под `## [Unreleased]` в новый заголовок `## [X.Y.Z] - YYYY-MM-DD` прямо под ним, оставляет `[Unreleased]` пустым и обновляет блок ссылок так, чтобы `[Unreleased]` сравнивал новый тег с `HEAD`.

```bash
srekit changelog release --version 1.2.0
```

Подсекции change type, где нет ничего кроме голого `-` из скаффолда, по дороге выкидываются — выпущенная версия никогда не шипит пустой `### Deprecated`. `[Unreleased]` остаётся действительно пустым, а не перезаполняется скелетом из шести типов: в примере самого Keep a Changelog он тоже пустой, а перезаполнение клало бы в каждый релизный diff шесть заголовков и шесть плейсхолдеров, которые следующий релиз всё равно вычистит.

### Флаги

| Флаг | Обязательный | Описание |
|---|---|---|
| `--version` | да | Выпускаемая версия, без tag prefix (например `1.2.0`). |
| `--date` | нет | Дата релиза в форме `YYYY-MM-DD`. Default: сегодня. |
| `--yanked` | нет | Пометить релиз отозванным: `## [X.Y.Z] - YYYY-MM-DD [YANKED]`. |
| `--dry-run` | нет | Напечатать результат, не писать. |
| `--stdout` | нет | Напечатать результат, не писать. |
| `--json` | нет | Отдать разобранный документ (версии, даты, yanked-состояние, change types, определения ссылок) и не писать. |

Цель — `CHANGELOG.md` в рабочей директории либо путь, переданный единственным позиционным аргументом:

```bash
srekit changelog release --version 1.2.0 docs/CHANGELOG.md
```

### Ни `--out`, ни `--force`

`release` — не генератор, и [общий набор output-флагов](index.md#shared-output-flags) он несёт не целиком. Он переписывает тот файл, на который его навели: вторая точка назначения не имеет смысла, а guard от перезаписи защищал бы от собственной цели команды. Любой из этих флагов — ошибка unknown flag, а не молча проигнорированная опция.

### Последовательность релизного дня

Команда правит текст. Она не коммитит, не тегает и не пушит — это остаётся под твоей рукой:

```bash
srekit changelog release --version 1.2.0 --dry-run   # 1. сначала посмотреть
srekit changelog release --version 1.2.0             # 2. отрезать
git diff CHANGELOG.md                                # 3. отревьюить
git commit -am "release: 1.2.0"                      # 4. закоммитить
git tag -a v1.2.0 -m "1.2.0" && git push origin v1.2.0   # 5. затегать
```

### Отозванные релизы

Релиз, отозванный после публикации, помечается, а не удаляется — номер версии сожжён в любом случае, и читателю нужно понимать, откуда взялась дыра:

```bash
srekit changelog release --version 0.0.5 --date 2014-12-13 --yanked
# ## [0.0.5] - 2014-12-13 [YANKED]
```

`--date` существует ровно для случаев, когда «сегодня» — неправильный ответ: бэкфилл версии, вышедшей до того, как ты начал пользоваться этим инструментом, или релиз через границу таймзоны, где дата тега и твоя локальная дата расходятся.

### Конвенции ссылок берутся из документа

Новые определения строятся из собственной строки `[Unreleased]` документа, а не из git remote. В этой строке уже закодированы хост, путь репозитория, форма URL сравнения и наличие префикса `v` у тегов. Поэтому проект на self-hosted GitLab или проект с голыми тегами `1.2.0` сохраняет свою конвенцию:

```
[Unreleased]: https://git.example.com/group/proj/-/compare/1.1.0...HEAD
```

после релиза становится

```
[Unreleased]: https://git.example.com/group/proj/-/compare/1.2.0...HEAD
[1.2.0]: https://git.example.com/group/proj/-/compare/1.1.0...1.2.0
```

Определение новой версии сравнивает её с предыдущей самой свежей, а если она первая — указывает на её release tag. Slug репозитория резолвится из git только тогда, когда блока ссылок в документе нет вообще, — тем же способом, что и в `srekit changelog`.

### От чего команда отказывается

Каждый из этих случаев завершается ненулевым кодом и оставляет файл байт-в-байт прежним:

| Условие | Почему |
|---|---|
| `--date` не в форме `YYYY-MM-DD` | Проверяется до чтения файла, чтобы опечатка не добралась до документа. |
| Целевого файла не существует | Сообщается с путём и указанием на `srekit changelog`. Ничего не создаётся. |
| Нет заголовка `## [Unreleased]` | Некуда вставлять, а угадывать место — ровно тот способ, которым переписыватель уничтожает историю. |
| В `[Unreleased]` нет записей | Нечего релизить. Подсекции из одних плейсхолдеров не считаются. |
| У версии уже есть заголовок | Повторный запуск — ошибка, а не идемпотентный no-op: записи, которые он бы перенёс, уже не те, что вышли в релиз. Правь файл руками. |

### Всё остальное сохраняется verbatim

Меняются только три региона: `[Unreleased]`, вставленная версия и блок ссылок. Написанная руками преамбула, стиль пустых строк, стиль маркеров списка, ранее выпущенные версии и хвост документа выходят байт-в-байт такими же, какими вошли. Это свойство дизайна: переписыватель делает splice по байтовым офсетам, а не сериализует заново разобранную модель. Именно поэтому релизный diff остаётся читаемым на ревью.

## Валидация существующего changelog { #validate }

`srekit changelog validate [FILE]` по каждой проверке сообщает, где документ расходится с Keep a Changelog. Файл не переписывается никогда.

```bash
srekit changelog validate
```

```
OK    heading-shape
OK    unreleased-section
FAIL  version-order: versions must appear in descending order: 1.1.0 (line 12) is listed above 1.2.0 (line 24)
OK    no-duplicate-versions
FAIL  change-types: unrecognized change type line 31: Improvements; allowed: Added, Changed, Deprecated, Removed, Fixed, Security
OK    link-definitions
```

| Проверка | Что требует |
|---|---|
| `heading-shape` | Каждый заголовок версии — `## [X.Y.Z] - YYYY-MM-DD`, опционально с ` [YANKED]`. Именно она ловит региональную дату вроде `04/03/2026`. |
| `unreleased-section` | Секция `[Unreleased]` присутствует и стоит выше всех выпущенных версий. |
| `version-order` | Выпущенные версии идут по убыванию. Сравнение посегментно-числовое, поэтому `1.10.0` корректно оказывается выше `1.9.0`. |
| `no-duplicate-versions` | Ни одна версия не встречается дважды. |
| `change-types` | Каждая подсекция `###` — одна из `Added`, `Changed`, `Deprecated`, `Removed`, `Fixed`, `Security`. |
| `link-definitions` | У каждого заголовка версии есть определение в блоке link reference definitions. |

Сообщается каждая проверка — и прошедшая, и упавшая: человеку, который чинит разъехавшийся changelog, нужен весь список за один проход, а не первая ошибка. Ненулевой код возврата — если упала хотя бы одна.

Словарь change type — это шесть типов из спецификации, а не то, что написано в твоём кастомном `changelog.yaml`. Переименованный заголовок здесь падает, и это правильный ответ: формат называет именно эти шесть.

`validate` сообщает, но не чинит. Правка разъехавшегося документа — это изменение, которое ты должен увидеть.

## Структура данных для шаблона

`changelog` шипится как v1 YAML-артефакт (`internal/tmpl/templates/changelog.yaml`) — H1 + `header_body` (intro-параграф) + две секции (`unreleased` и `initial_release`) + `footer_body` (блок link reference definitions). Заголовок секции `initial_release` динамический (`[{{ .Meta.InitialVersion }}] - {{ .Meta.Today }}`); section titles template-evaluated с v0.20.0. Template-выражения обращаются к `.Meta.<Field>` для `Today` (дата `2006-01-02`), `Repo` (`<owner>/<name>`), `InitialVersion`. См. [`srekit postmortem`](postmortem.md#customizing-the-artifact-v1-format-v0140) для полной схемы.

## См. также

- [`srekit rfc`](rfc.md), [`srekit postmortem`](postmortem.md) — документы, история которых питает changelog.
