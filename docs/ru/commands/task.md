# srekit task

Сгенерировать **investigation log** — структурированный артефакт для
ведения цепочки гипотез и доказательств, когда охотишься за tail-latency
спайком, флэйковым тестом или любой open-ended SRE-загадкой. Скрытый
алиас: `srekit sretask` (оставлен для миграции с `gch sretask`).

## Синопсис

```bash
srekit task --title TITLE [flags]
```

## Флаги

| Флаг | Обязательный | Описание |
|---|---|---|
| `--title` | да | Тема расследования; идёт в H1 и в default-имя файла |
| `--path DIR` | нет | Директория для записи (default: текущая) |

Плюс [общие output-флаги](index.md#shared-output-flags): `--out`,
`--stdout`, `--force`, `--dry-run`, `--template`, `--json`.

## Default имя файла

Если ни `--out`, ни `--stdout` не передан — пишется в
`<path>/Tasker - <title>.md` (slug-нормализованное, регистр title
сохраняется).

## Примеры

Быстрая черновая запись в stdout:

```bash
srekit task --title "Tail latency on api-gw" --stdout
```

Запись в конкретную директорию:

```bash
srekit task --title "Tail latency on api-gw" --path ./tasks
# → ./tasks/Tasker - Tail latency on api-gw.md
```

Достать сгенерированный UUID через `jq`:

```bash
srekit task --title "Tail latency on api-gw" --json | jq -r '.ID'
```

## Структура данных для шаблона

```go
struct {
    ID, CreationDate, ModificationDate, Title string
}
```

Секции после рендеринга: YAML front matter (`title`, `tags`,
`creation_date`, `id`) → `Контекст / Context`, `Гипотеза / Hypothesis`,
`Доказательства / Evidence`, `Выводы / Findings`, `Дальнейшие действия /
Action items`, `Ссылки / References`.

## См. также

- [Кастомные шаблоны](../guides/custom-templates.md) — переопределить
  embedded-шаблон своим.
- [JSON-вывод](../guides/json-output.md) — пайплайны `--json` в другие
  тулзы.
