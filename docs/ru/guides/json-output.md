# JSON-вывод для пайплайнов

Каждый генератор поддерживает `--json`. Флаг отдаёт структурный payload вместо рендеринга Markdown, чтобы агентские флоу и shell-пайплайны могли читать документ по полям и (для `postmortem`) round-trip'ить его обратно.

## Контракт (v0.20.0+)

- Default sink — **stdout**. `--out FILE` пишет JSON туда.
- Имена полей — **camelCase** во всех командах (`id`, `title`, `latencyTarget`, …).
- С `--json` Markdown default-путь (`Tasker - <title>.md`, `postmortem-<YYYY-MM-DD>-<slug>.md` и т.п.) **не** используется — JSON не попадёт случайно в `.md` файл.
- **У каждого payload форма `{meta, sections}`.** Метаданные живут под `meta`; отрендеренный документ — список типизированных секций под `sections`. У каждой секции `{id, title, type, required, body}`, где `type` — это `text` / `list` / `table`, а `body` всегда строка.
- **Секции per-artifact**, в порядке манифеста, со стабильными ID. Обращайся `jq '.sections[] | select(.id == "<id>").body'` — никогда по индексу.

| Режим | Команды | Секции |
|---|---|---|
| **Structured** | все генераторы (postmortem, task, retro, slo, ebp, capacity, incident, rfc, runbook, oncall-report, changelog) | Несколько типизированных секций — одна на слот в YAML-артефакте — в декларированном порядке. |

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

Все генераторы на v1 artifact path: `meta` отражает per-command набор флагов, `sections` — список из YAML-артефакта. Авторы шаблонов обращаются к meta-полям как `.Meta.<Field>` внутри YAML; `--json` отдаёт camelCase под `meta`.

```jsonc
// task
{ "meta": { "id", "title", "now" },
  "sections": [ ...секции из internal/tmpl/templates/task.yaml... ] }

// postmortem
{ "meta": { "id", "title", "severity", "start", "end", "owner", "now" },
  "sections": [ ...секции из postmortem.yaml (по умолчанию 12)... ] }

// rfc
{ "meta": { "id", "title", "status", "now", "author": { "name", "email" } },
  "sections": [ ...секции из rfc.yaml... ] }

// slo
{ "meta": { "id", "service", "target", "window", "latencyTarget", "now" },
  "sections": [ ...секции из slo.yaml... ] }
```

`author` (где есть) — вложенный объект (`{ "name", "email" }`), обращаться `.meta.author.name` / `.meta.author.email`. Точные section ID каждой команды смотри в `internal/tmpl/templates/<name>.yaml`.

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
