# srekit tasker

Сгенерировать **карточку задачи** для коллекции инженерных задач: frontmatter с темой, уровнем, форматом и ожидаемой длительностью, H1 с названием задачи и две секции, из которых карточка состоит — сама задача и то, что хочется услышать в ответ.

Карточка приезжает пустой намеренно. `tasker` раскладывает форму, содержимое пишет тот, кто задачу добавляет.

## Синопсис

```bash
srekit tasker --title NAME [flags]
```

## Флаги

| Флаг | Обязательный | Описание |
|---|---|---|
| `--title`, `-T` | да | Название задачи — становится H1 и slug'ом в имени файла |
| `--topic` | нет | Тема. Default: `go` |
| `--level` | нет | Уровни: флаг повторяемый или через запятую. Default: `middle,senior` |
| `--format` | нет | Формат ответа (`code`, `theory`, `design`, …). Default: `code` |
| `--duration` | нет | Ожидаемое время решения в минутах, положительное. Default: `30` |

Плюс [общие output-флаги](index.md#shared-output-flags). Default имя файла: `tasker-<slug-of-title>.md`.

Пустые уровни отбрасываются, так что `--level "middle, "` — это один уровень. Если после этого не осталось ничего, а также если `--duration` меньше или равен нулю, команда падает до того, как что-либо записано.

!!! note "Русские заголовки и имя файла"
    Slug оставляет только `[a-z0-9]`, поэтому полностью русское название вырождается в `untitled`, и каждая карточка метит в один и тот же `tasker-untitled.md` — вторая откажется перезаписывать первую. Для таких задавай `--out` или называй файл руками.

## Примеры

Дефолты — 30-минутная Go-задача на код для middle и senior:

```bash
srekit tasker --title "Channels and select" --stdout
```

Короткий теоретический вопрос для junior:

```bash
srekit tasker -T "Что делает GOMAXPROCS" --topic go --level junior \
  --format theory --duration 10
```

В коллекцию, с именем файла на свой вкус:

```bash
srekit tasker -T "Каналы и select" --out "tasks/Tasker - Каналы и select.md"
```

## Что получается

```markdown
---
id: "b0a1…"
creation_date: "2026-08-28T14:20:08+03:00"
type: simple_note
tags:
  - tasker
topic: "go"
level: [middle, senior]
format: "code"
duration: 30
---

# Tasker - Channels and select

## Описание (Description)

## Что хотим услышать (What we want to hear)
```

`level` — YAML-список, `duration` — число, а не строки: frontmatter карточки читает та коллекция, в которой она лежит, и `level: [middle, senior]` фильтруется, а `level: "middle, senior"` — нет. Как сделать то же самое в своих артефактах — [не-строковые значения frontmatter](../guides/custom-templates.md#typed-front-matter-values).

## Структура секций

- Front matter: `id`, `creation_date`, `type: simple_note`, `tags`, `topic`, `level`, `format`, `duration`
- Описание (Description) — сама задача
- Что хотим услышать (What we want to hear) — как звучит хороший ответ

Оба тела по умолчанию пустые. Это осознанно: плейсхолдер пришлось бы удалять в каждой карточке.

## Структура данных для шаблона

`tasker` шипится как v1 YAML-артефакт (`internal/tmpl/templates/tasker.yaml`) — frontmatter, H1, секции `description`, `expectations`. Template-выражения обращаются к `.Meta.<Field>` для `ID`, `Now`, `Title`, `Topic`, `Level` (список), `Format`, `Duration` (число). См. [`srekit postmortem`](postmortem.md#customizing-the-artifact-v1-format-v0140) для полной схемы.

## См. также

- [`srekit task`](task.md) — investigation log для SRE-расследования. Похожее имя, другой документ: `task` фиксирует ход разбирательства, `tasker` описывает задачу, которую будет решать кто-то другой.
- [Кастомные шаблоны](../guides/custom-templates.md) — переопредели `tasker.yaml` под коллекцию с другим frontmatter.
