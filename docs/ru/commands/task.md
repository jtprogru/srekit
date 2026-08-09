# srekit task

Сгенерировать **investigation log** — структурированный артефакт для ведения цепочки гипотез и доказательств, когда охотишься за tail-latency спайком, флэйковым тестом или любой open-ended SRE-загадкой. Скрытый алиас: `srekit sretask` (оставлен для миграции с `gch sretask`).

## Синопсис

```bash
srekit task --title TITLE [flags]
```

## Флаги

| Флаг | Обязательный | Описание |
|---|---|---|
| `--title` | да | Тема расследования; идёт в H1 и в default-имя файла |
| `--path DIR` | нет | Директория для записи (default: текущая) |

Плюс [общие output-флаги](index.md#shared-output-flags): `--out`, `--stdout`, `--force`, `--dry-run`, `--json`.

## Default имя файла

Если ни `--out`, ни `--stdout` не передан — пишется в `<path>/investigation-<slug>.md` (lowercased, slug-нормализованное).

## Примеры

Быстрая черновая запись в stdout:

```bash
srekit task --title "Tail latency on api-gw" --stdout
```

Запись в конкретную директорию:

```bash
srekit task --title "Tail latency on api-gw" --path ./tasks
# → ./tasks/investigation-tail-latency-on-api-gw.md
```

Достать сгенерированный UUID через `jq`:

```bash
srekit task --title "Tail latency on api-gw" --json | jq -r '.meta.id'
```

## Структура данных для шаблона

`task` шипится как v1 YAML-артефакт (`internal/tmpl/templates/task.yaml`) — frontmatter (`id`, `creation_date`, `modification_date`, `type: investigation`, `title`, `tags`), H1, meta_bullets и шесть секций: `context` (Контекст / Context), `hypothesis` (Гипотезы / Hypothesis), `evidence` (Наблюдения / Evidence), `findings` (Выводы / Findings), `action_items` (Задачи / Action items), `references` (Ссылки / References). Template-выражения внутри YAML обращаются к `.Meta.<Field>`; доступны `ID`, `Title`, `CreationDate`, `ModificationDate`.

## См. также

- [Кастомные шаблоны](../guides/custom-templates.md) — переопределить встроенный артефакт своим `task.yaml`.
- [JSON-вывод](../guides/json-output.md) — пайплайны `--json` в другие тулзы (per-section access через `jq '.sections[] | select(.id=="…").body'`).
