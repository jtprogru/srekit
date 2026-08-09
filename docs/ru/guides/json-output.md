# JSON-вывод для пайплайнов

Каждый генератор поддерживает `--json`. Флаг отдаёт структурный payload вместо рендеринга Markdown, чтобы агентские флоу и shell-пайплайны могли читать документ по полям и (для `postmortem`) round-trip'ить его обратно.

## Контракт (v0.20.0+)

- Default sink — **stdout**. `--out FILE` пишет JSON туда.
- Имена полей — **camelCase** во всех командах (`id`, `title`, `latencyTarget`, …).
- С `--json` Markdown default-путь (`investigation-<slug>.md`, `postmortem-<YYYY-MM-DD>-<slug>.md` и т.п.) **не** используется — JSON не попадёт случайно в `.md` файл.
- **У каждого payload форма `{meta, sections}`.** Метаданные живут под `meta`; отрендеренный документ — список типизированных секций под `sections`. У каждой секции `{id, title, type, required, body}`, где `type` — это `text` / `list` / `table`, а `body` всегда строка.
- **Секции per-artifact**, в порядке манифеста, со стабильными ID. Обращайся `jq '.sections[] | select(.id == "<id>").body'` — никогда по индексу.

| Режим | Команды | Секции |
|---|---|---|
| **Structured** | все генераторы (postmortem, task, slo, ebp, rfc, runbook, oncall-report, changelog) | Несколько типизированных секций — одна на слот в YAML-артефакте — в декларированном порядке. |

!!! warning "Миграция с pre-v1 раскладок"
    - **0.12.x → 0.13.0**: форма изменилась с плоской `{title, severity, …}` на `{meta, sections}`. Миграция: `jq '.title'` → `jq '.meta.title'`.
    - **0.13.x → v0.20**: YAML-first миграция убрала bootstrap envelope (`sections: [{id: "body", body: <markdown>}]`) у всех генераторов. Миграция: заменяй `jq '.sections[0].body'` на `jq '.sections[] | select(.id == "<id>").body'` для нужной секции. Per-release section ID см. в `docs/{en,ru}/migration/v1.md`.

## Паттерны

### Достать поле из `meta`

```bash
srekit task --title "Tail latency" --json | jq -r '.meta.id'
# 085883a2-32d0-4d50-9bc6-ac219e29409c
```

### Перечислить секции постмортема

```bash
srekit postmortem -T X --json | jq '.sections[] | {id, type, required}'
```

### Достать body одной секции

```bash
# Postmortem — забрать summary
srekit postmortem -T X --json | jq -r '.sections[] | select(.id == "summary").body'

# Runbook — забрать diagnose
srekit runbook --title "p99 spike" --service api-gw --alert APIHighLatency --json |
  jq -r '.sections[] | select(.id == "diagnose").body'

# Changelog — забрать initial release
srekit changelog --repo owner/repo --json |
  jq -r '.sections[] | select(.id == "initial_release").body'
```

### Round-trip постмортема { #round-trip }

Вывод и ввод — разной формы, и это главное, что здесь надо не перепутать: `--json` отдаёт `sections` **упорядоченным списком** объектов, а `--from` читает **map** по ID секции. Список сохраняет порядок манифеста на выходе; на входе порядок неважен, поэтому честная форма — map. Попытка проиндексировать выданный список строкой (`jq '.sections.summary = …'`) падает с `Cannot index array with string`.

Преобразуй список в map, отредактируй, пересобери:

```bash
# Выгрузить
srekit postmortem -T "API outage" --severity SEV-1 --json > pm.json

# Переложить sections в форму --from и поставить одно тело
jq '{meta, sections: (.sections | map({key: .id, value: .body}) | from_entries)}
    | .sections.summary = "27-минутный 5xx на checkout, замитигировано failback-ом на cache."' \
  pm.json > pm.edited.json

# Пересобрать
srekit postmortem -T "API outage" --from pm.edited.json
```

Возвращать все секции не обязательно. `--from` накладывает то, что ему дали, поверх дефолтов артефакта, так что минимальный файл — только изменённые секции:

```json
{ "sections": { "summary": "27-минутный 5xx на checkout, замитигировано failback-ом на cache." } }
```

### Round-trip changelog'а

`changelog` тоже принимает `--from`, с той же формой payload'а:

```bash
srekit changelog --repo acme/api --json > cl.json
jq '{meta, sections: (.sections | map({key: .id, value: .body}) | from_entries)}
    | .sections.unreleased = "### Added\n\n- Structured input for changelog.\n"' \
  cl.json > cl.edited.json
srekit changelog --from cl.edited.json
```

`meta` в payload'е задаёт `repo`, `initialVersion` и `today`; флаг выигрывает у файла, файл — у git remote. В отличие от `postmortem`, у `changelog` нет `--schema` и `--validate`: его артефакт не объявляет обязательных секций, поэтому валидация payload'а не может завершиться неудачей.

### Чего нет в `sections`

`footer_body` артефакта — хвостовой материал уровня документа, например блок link reference definitions в changelog — **не** секция. У него нет `id`, он не появляется в массиве `sections` и его нельзя подменить через `--from`. Он рендерится из артефакта на каждом вызове — именно поэтому замена тела секции не может уронить compare-ссылки changelog'а.

### Спроецировать в свою структуру

```bash
srekit postmortem --title "API outage" --severity SEV-1 --json |
  jq '{title: .meta.title, severity: .meta.severity, started: .meta.start, owner: .meta.owner}'
```

### Драйвить другой инструмент

```bash
srekit slo --service api-gw --target 99.95% --window 30d --json |
  jq -r '.meta | "\(.service) \(.target) \(.window)"' |
  xargs my-slo-registrar register
```

### Сравнить две генерации

```bash
diff <(srekit slo --service api-gw --target 99.9% --json) \
     <(srekit slo --service api-gw --target 99.95% --json)
```

### Записать JSON в файл

`--json` уважает `--out`:

```bash
srekit oncall-report --team platform --json --out oncall.json
```

`--dry-run` тоже работает — печатает "would write N bytes to oncall.json" плюс тело.

## Per-command структура payload

Все генераторы на v1 artifact path: `meta` отражает per-command набор флагов, `sections` — список из YAML-артефакта. Авторы шаблонов обращаются к meta-полям как `.Meta.<Field>` внутри YAML; `--json` отдаёт camelCase под `meta`.

```jsonc
// task            — 6 секций
{ "meta": { "id", "title", "creationDate", "modificationDate" } }

// postmortem      — 12 секций
{ "meta": { "id", "title", "severity", "start", "end", "owner", "now" } }

// rfc             — 5 секций
{ "meta": { "id", "title", "status", "now", "author": { "name", "email" } } }

// runbook         — 7 секций
{ "meta": { "id", "title", "service", "alert", "now" } }

// slo             — 7 секций
{ "meta": { "id", "service", "target", "window", "latencyTarget", "now" } }

// ebp             — 7 секций
{ "meta": { "id", "service", "now" } }

// oncall-report   — 8 секций
{ "meta": { "id", "team", "start", "end", "now", "author": { "name", "email" } } }

// changelog       — 2 секции
{ "meta": { "repo", "initialVersion", "today" } }
```

`sections` выше опущены для краткости — они есть в каждом payload'е и повторяют список из соответствующего `internal/tmpl/templates/<name>.yaml`. Реальные id команды спрашивай не у этой страницы, а у бинарника:

```bash
srekit runbook -T X --json | jq -r '.sections[].id'
```

`author` (где есть) — вложенный объект (`{ "name", "email" }`), обращаться `.meta.author.name` / `.meta.author.email`.

## Когда использовать `--json`

- **Агентские флоу**: прочитать секцию, изменить, записать обратно. Postmortem и changelog это поддерживают; `--from` round-trip работает из коробки.
- Скрипты / автоматизация: вместо `grep`'а по Markdown — `--json | jq`.
- Drift-чеки: сохраняешь JSON-вывод предыдущей генерации, diff'ишь с новым — ловишь изменения полей шаблона.
- Cross-tool интеграция: значения сразу в Linear, Jira или внутренние CLI.

## Когда **не** использовать `--json`

- Тебе нужен сам документ — это default mode.
- Тебе нужно вставить документ в другой файл — тоже default, через `--stdout` в пайп.
- Шарить с не-инженером — Markdown читается лучше JSON.

## См. также

- [Рецепты](../recipes.md) — конкретные `--json` пайплайны.
- [`srekit postmortem`](../commands/postmortem.md) — первый structured-генератор; подробно про `--from`.
- [`templates list --json`](../commands/templates.md#templates-list) — introspection JSON (те же camelCase ключи, плоская list-форма — отличается от вывода генераторов).
