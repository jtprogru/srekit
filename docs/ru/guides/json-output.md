# JSON-вывод для пайплайнов

Каждый генератор поддерживает `--json`. Флаг короткозамыкает шаблонный движок: вместо рендеринга Markdown, srekit отдаёт data payload шаблона как indented JSON. Payload — это то, что увидел бы Go-шаблон.

## Контракт

- Default sink — **stdout**. `--out FILE` пишет JSON туда.
- Имена полей — **camelCase** во всех командах (`id`, `title`, `latencyTarget`, …).
- С `--json` Markdown default-путь (`Tasker - <title>.md`, `postmortem-<slug>.md` и т.п.) **не** используется — JSON не попадёт случайно в `.md` файл.

!!! note "Единый JSON-контракт"
    Все команды — и генераторы, и introspection
    (`templates list --json`) — отдают **camelCase** ключи. В ранних
    0.x релизах генераторы (PascalCase) и `templates list` (camelCase)
    расходились; этого расхождения больше нет, так что одно соглашение
    `jq` работает везде.

## Паттерны

### Извлечь одно поле

```bash
srekit task --title "Tail latency" --json | jq -r '.id'
# 085883a2-32d0-4d50-9bc6-ac219e29409c
```

### Спроецировать в свою структуру

```bash
srekit postmortem --title "API outage" --severity SEV-1 --json |
  jq '{title: .title, severity: .severity, started: .start, owner: .owner}'
```

### Драйвить другой инструмент

Сгенерить SLO, взять параметры, зарегистрировать в metrics-tool:

```bash
srekit slo --service api-gw --target 99.95% --window 30d --json |
  jq -r '"\(.service) \(.target) \(.window)"' |
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

Полная структура, которая передаётся в шаблон, перечислена на странице каждой команды (раздел "Структура данных для шаблона"). Авторы шаблонов обращаются к полям по Go-именам (`.Title`); `--json` отдаёт camelCase-ключи ниже:

```jsonc
// task
{ "id", "creationDate", "modificationDate", "title" }

// postmortem
{ "id", "title", "severity", "start", "end", "owner", "now" }

// rfc
{ "id", "title", "status", "now", "author": { "name", "email" } }

// oncall-report
{ "id", "team", "start", "end", "now", "author": { "name", "email" } }

// slo
{ "id", "service", "target", "window", "latencyTarget", "now" }
```

`author` — вложенный объект (`{ "name", "email" }`), обращаться `.author.name` / `.author.email` в `jq`.

## Когда использовать `--json`

- Скрипты / автоматизация: вместо `grep`'а по Markdown — `--json | jq`, намного надёжнее.
- Drift-чеки: сохраняешь JSON-вывод предыдущей генерации, diff'ишь с новым — ловишь изменения полей шаблона.
- Cross-tool интеграция: значения сразу в Linear, Jira или внутренние CLI.

## Когда **не** использовать `--json`

- Тебе нужен сам документ — это default mode.
- Тебе нужно вставить документ в другой файл — тоже default, через `--stdout` в пайп.
- Шарить с не-инженером — Markdown читается лучше JSON.

## См. также

- [Рецепты](../recipes.md) — конкретные `--json` пайплайны.
- [`templates list --json`](../commands/templates.md#templates-list) — introspection JSON (те же camelCase ключи).
