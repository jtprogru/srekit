# srekit postmortem

Сгенерировать **постмортем** в стиле Google SRE: severity, timeline, impact, detection / mitigation / root cause, action items, lessons. Билингвальные заголовки.

Начиная с v0.13.0, postmortem — первый генератор, построенный на **sidecar-манифесте секций** (`postmortem.sections.yaml`). Тело документа собирается из типизированных секций (`text` / `list` / `table`), так что `--json` отдаёт структуру секция-за-секцией, а `--from input.json` пересобирает Markdown из отредактированного JSON.

## Синопсис

```bash
srekit postmortem --title TITLE [flags]
```

## Флаги

| Флаг | Обязательный | Описание |
|---|---|---|
| `--title` | да (если не передан через `--from`) | Тема инцидента |
| `--severity` | нет | `SEV-N` метка (free text) |
| `--start` | нет | Начало инцидента (timestamp) |
| `--end` | нет | Конец инцидента |
| `--owner` | нет | Владелец постмортема (тот, кто пишет) |
| `--from FILE` | нет | Прочитать структурный input (`{meta?, sections}`) из JSON-файла; `-` читает stdin |

Плюс [общие output-флаги](index.md#shared-output-flags). Default имя файла: `postmortem-<YYYY-MM-DD>-<slug-of-title>.md` (дата — сегодняшняя, в часовом поясе системы).

## Примеры

Короткая форма (рендер Markdown в default-файл):

```bash
srekit postmortem --title "API outage" --severity SEV-1
```

Посмотреть структурный JSON (секции в порядке манифеста):

```bash
srekit postmortem -T "API outage" --severity SEV-1 --json | jq '.sections[].id'
```

Round-trip: выгрузить JSON, отредактировать одну секцию, пересобрать Markdown:

```bash
srekit postmortem -T "API outage" --json > pm.json
# отредактировать pm.json — например, заполнить sections.summary реальным текстом
srekit postmortem -T "API outage" --from pm.json
```

Достать только метаданные для трекера:

```bash
srekit postmortem --title "API outage" --severity SEV-1 --json \
  | jq '{title: .meta.title, severity: .meta.severity, started: .meta.start}'
```

## JSON shape

```json
{
  "meta": {
    "id": "…uuid…",
    "title": "API outage",
    "severity": "SEV-1",
    "start": "",
    "end": "",
    "owner": "",
    "now": "2026-06-04T12:34:30+03:00"
  },
  "sections": [
    { "id": "summary",      "title": "Краткое описание (Summary)", "type": "text",  "required": true,  "body": "…" },
    { "id": "impact",       "title": "Влияние (Impact)",            "type": "list",  "required": true,  "body": "- …" },
    { "id": "timeline",     "title": "Хронология (Timeline)",       "type": "table", "required": true,  "body": "| Время | Событие |\n|---|---|\n…" },
    { "id": "root_cause",   "title": "Корневая причина (Root Cause)","type": "text",  "required": true,  "body": "…" },
    { "id": "detection",    "title": "Обнаружение (Detection)",      "type": "list",  "required": false, "body": "- …" },
    { "id": "resolution",   "title": "Разрешение (Resolution)",      "type": "list",  "required": false, "body": "- …" },
    { "id": "what_went_well",  "title": "Что сработало хорошо (What went well)",  "type": "list",  "required": false, "body": "- " },
    { "id": "what_went_wrong", "title": "Что пошло не так (What went wrong)",     "type": "list",  "required": false, "body": "- " },
    { "id": "where_got_lucky", "title": "Где нам повезло (Where we got lucky)",   "type": "list",  "required": false, "body": "- " },
    { "id": "action_items", "title": "Задачи (Action items)",         "type": "table", "required": true,  "body": "| # | Действие | … |" },
    { "id": "lessons_learned", "title": "Извлечённые уроки (Lessons learned)", "type": "list", "required": false, "body": "- " },
    { "id": "references",   "title": "Ссылки (References)",           "type": "list",  "required": false, "body": "- " }
  ]
}
```

Секции отдаются в порядке манифеста. `body` всегда строка вне зависимости от `type`; для `list` и `table` это уже отрендеренный markdown-фрагмент, чтобы потребители видели одно и то же значение через оба пути (Markdown и JSON).

## Формат `--from` input

```json
{
  "meta": {
    "title": "API outage",
    "severity": "SEV-1",
    "owner": "@oncall"
  },
  "sections": {
    "summary": "We saw a 27-minute checkout 5xx spike starting 10:05 UTC.",
    "root_cause": "Cache stampede after a feature-flag flip increased TTL pressure."
  }
}
```

- Оба поля `meta` и `sections` опциональны.
- CLI-флаги (`--title`, `--severity`, …) перекрывают `meta` из файла. Это позволяет пинить поле в команде, даже когда читаешь из stdin.
- Тела секций в `sections` подставляются **as-is** — без template-evaluation — так что round-trip произвольного markdown с `{{ … }}` внутри безопасен.
- Section ID, которых нет в манифесте, дают hard error со списком неизвестных ID и known-set (защита от опечаток).
- Секции, которых нет в input, заполняются дефолтами из манифеста.

## Кастомизация манифеста

`srekit templates init <dir>` положит и `postmortem.md.tmpl` (header-шаблон), и `postmortem.sections.yaml` (список секций) в твой templates dir. Редактируй YAML, чтобы добавить, убрать или поменять порядок секций, изменить заголовки, обновить default-наполнение (text-тела, list-айтемы, columns/rows для таблиц).

Формат манифеста (см. [`internal/sections`](https://github.com/jtprogru/srekit/tree/main/internal/sections)):

```yaml
version: 1
sections:
  - id: summary           # стабильный ID, используется в --json и --from
    title: "Краткое описание (Summary)"
    type: text            # text | list | table
    required: true
    default_body: |
      _Один абзац: что произошло, кого затронуло, как смягчили._

  - id: impact
    title: "Влияние (Impact)"
    type: list
    required: true
    items:
      - "Влияние на пользователей:"
      - "Расход SLO / SLI:"

  - id: timeline
    title: "Хронология (Timeline)"
    type: table
    required: true
    default_body: "(Все времена в UTC.)"
    columns: ["Время", "Событие"]
    rows:
      - ["{{ .Meta.Start }}", "Инцидент начался"]
      - ["{{ .Meta.End }}",   "Инцидент разрешён"]
```

Default-наполнение (text-тела, list-айтемы, ячейки таблиц) проходит через тот же Go-template FuncMap что и основной шаблон, поэтому `{{ .Meta.Start }}`, `{{ now "2006-01-02" }}` и прочие [хелперы](../guides/custom-templates.md#funcmap) работают.

## См. также

- [`srekit incident`](incident.md) — документ *во время* инцидента.
- [`srekit retro`](retro.md) — sprint-level ретроспектива.
- [JSON output guide](../guides/json-output.md) — кросс-командный контракт `{meta, sections}`.
