# JSON-вывод для пайплайнов

Каждый генератор поддерживает `--json`. Флаг короткозамыкает шаблонный движок: вместо рендеринга Markdown, srekit отдаёт data payload шаблона как indented JSON. Payload — это то, что увидел бы Go-шаблон.

## Контракт

- Default sink — **stdout**. `--out FILE` пишет JSON туда.
- Имена полей — **PascalCase** (Go field names без `json:` тегов).
- С `--json` Markdown default-путь (`Tasker - <title>.md`, `postmortem-<slug>.md` и т.п.) **не** используется — JSON не попадёт случайно в `.md` файл.

!!! note "Два JSON-контракта в v0.x"
    Генераторы отдают **PascalCase** ключи. Introspection-команды
    (`templates list --json`) отдают **camelCase** ключи, потому что
    идут через tagged-структуры, удовлетворяющие линтер `tagliatelle`.
    Сведём к единому стандарту в v1.0. Пока правило: "JSON от `--json`
    у генератора — PascalCase, JSON от `templates list --json` или
    любой будущей introspection — camelCase."

## Паттерны

### Извлечь одно поле

```bash
srekit task --title "Tail latency" --json | jq -r '.ID'
# 085883a2-32d0-4d50-9bc6-ac219e29409c
```

### Спроецировать в свою структуру

```bash
srekit postmortem --title "API outage" --severity SEV-1 --json |
  jq '{title: .Title, severity: .Severity, started: .Start, owner: .Owner}'
```

### Драйвить другой инструмент

Сгенерить SLO, взять параметры, зарегистрировать в metrics-tool:

```bash
srekit slo --service api-gw --target 99.95% --window 30d --json |
  jq -r '"\(.Service) \(.Target) \(.Window)"' |
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

Полный Go struct, который передаётся в шаблон, перечислен на странице каждой команды (раздел "Структура данных для шаблона"). Самые частые:

```go
// task
struct { ID, CreationDate, ModificationDate, Title string }

// postmortem
struct { ID, Title, Severity, Start, End, Owner, Now string }

// rfc
struct { ID, Title, Status, Now string; Author struct { Name, Email string } }

// oncall-report
struct { ID, Team, Start, End, Now string; Author struct { Name, Email string } }

// slo
struct { ID, Service, Target, Window, Latency, Now string }
```

`Author` — вложенный объект (Go struct `meta.Author{Name, Email}`), обращаться `.Author.Name` / `.Author.Email` в `jq`.

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
- [`templates list --json`](../commands/templates.md#templates-list) — introspection JSON (camelCase ключи).
