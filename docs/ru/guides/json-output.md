# JSON-вывод для пайплайнов

Каждый генератор поддерживает `--json`. Флаг отдаёт структурный payload вместо рендеринга Markdown, чтобы агентские флоу и shell-пайплайны могли читать документ по полям и (для `postmortem`) round-trip'ить его обратно.

## Контракт (v0.13.0+)

- Default sink — **stdout**. `--out FILE` пишет JSON туда.
- Имена полей — **camelCase** во всех командах (`id`, `title`, `latencyTarget`, …).
- С `--json` Markdown default-путь (`Tasker - <title>.md`, `postmortem-<YYYY-MM-DD>-<slug>.md` и т.п.) **не** используется — JSON не попадёт случайно в `.md` файл.
- **У каждого payload форма `{meta, sections}`.** Метаданные живут под `meta`; отрендеренный документ — список секций под `sections`. У каждой секции `{id, title, type, required, body}`, где `type` — это `text` / `list` / `table`, а `body` всегда строка.

Контракт есть в двух вариантах:

| Режим | Команды | Секции |
|---|---|---|
| **Structured** | `postmortem` (управляется `postmortem.sections.yaml`) | Несколько типизированных секций, по одной на слот манифеста, в порядке манифеста. |
| **Bootstrap** | все остальные генераторы (`task`, `incident`, `rfc`, `runbook`, `retro`, `slo`, `oncall-report`, `ebp`, `capacity`, `license`, `changelog`) | Одна синтетическая секция `{id: "body", type: "text", title: <H1>, body: <rendered markdown>}`. |

Оба варианта делят одну внешнюю форму (`{meta, sections}`), так что скрипт, написанный для одной команды, работает с любой другой.

!!! warning "Breaking change в 0.13.0"
    Форма изменилась с плоской `{title, severity, …}` в 0.12.x на `{meta: {…}, sections: […]}` в 0.13.0. Миграция: заменяй `jq '.title'` на `jq '.meta.title'`, а `jq '.body'` (если ты `grep`'ал rendered markdown) на `jq -r '.sections[0].body'` (bootstrap-команды) или `jq -r '.sections[] | select(.id == "summary").body'` (postmortem).

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
# Postmortem (multi-section) — забрать summary
srekit postmortem -T X --json | jq -r '.sections[] | select(.id == "summary").body'

# Любой другой генератор (одна bootstrap-секция) — забрать весь rendered markdown
srekit runbook --service api-gw --alert APIHighLatency --json | jq -r '.sections[0].body'
```

### Round-trip постмортема

```bash
# Выгрузить
srekit postmortem -T "API outage" --severity SEV-1 --json > pm.json

# Отредактировать одну секцию в pm.json (любой инструмент — jq, sed, твой редактор, LLM)
jq '.sections.summary = "27-минутный 5xx на checkout, замитигировано failback-ом на cache."' pm.json > pm.edited.json

# Пересобрать
srekit postmortem -T "API outage" --from pm.edited.json
```

`--from` читает секции как **map**, ключ — ID (`{"summary": "…"}`), а не как list. JSON-вывод использует list, чтобы сохранять порядок манифеста; `--from` использует map, потому что порядок на входе неважен.

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

Объект `meta` отражает per-command набор флагов. Авторы шаблонов обращаются к полям по Go-именам (`.Meta.Title` для постмортема; legacy `.Title` для остальных команд пока они не мигрируют); `--json` отдаёт camelCase под `meta`:

```jsonc
// task (bootstrap)
{ "meta": { "id", "creationDate", "modificationDate", "title" },
  "sections": [{ "id": "body", "title": "<H1>", "type": "text", "required": true, "body": "<markdown>" }] }

// postmortem (structured — секции из postmortem.sections.yaml)
{ "meta": { "id", "title", "severity", "start", "end", "owner", "now" },
  "sections": [ ...12 объектов секций в порядке манифеста... ] }

// rfc (bootstrap)
{ "meta": { "id", "title", "status", "now", "author": { "name", "email" } },
  "sections": [{ "id": "body", ... }] }

// slo (bootstrap)
{ "meta": { "id", "service", "target", "window", "latencyTarget", "now" },
  "sections": [{ "id": "body", ... }] }
```

`author` (где есть) — вложенный объект (`{ "name", "email" }`), обращаться `.meta.author.name` / `.meta.author.email`.

## Когда использовать `--json`

- **Агентские флоу**: прочитать секцию, изменить, записать обратно. Postmortem — первая команда, оптимизированная под это; `--from` round-trip работает из коробки.
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
