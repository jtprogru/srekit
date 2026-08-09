# srekit postmortem

Сгенерировать **постмортем** в стиле Google SRE: severity, timeline, impact, detection / mitigation / root cause, action items, lessons. Билингвальные заголовки.

Тело документа собирается из типизированных секций (`text` / `list` / `table`), объявленных в v1-артефакте `postmortem.yaml`, так что `--json` отдаёт структуру секция-за-секцией, а `--from input.json` пересобирает Markdown из отредактированного JSON. Postmortem был прототипом v1 artifact-формата (введён в v0.14.0) и теперь является каноническим референсом схемы для остальных генераторов — секция «Кастомизация артефакта» ниже документирует каждое поле.

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
| `--schema` | нет | Отдать JSON Schema для `--from` payload (генерится из загруженного манифеста). Пишет в stdout. Несовместим с `--validate`. |
| `--validate FILE` | нет | Проверить input-файл: required-секции должны иметь непустое body; unknown section IDs отклоняются. Печатает per-section OK / FAIL отчёт; non-zero exit при любом FAIL. |

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

Цикл «выгрузить JSON, отредактировать одну секцию, пересобрать Markdown»:

```bash
srekit postmortem -T "API outage" --json > pm.json
# отредактировать pm.json и скормить обратно
srekit postmortem -T "API outage" --from pm.json
```

`--json` отдаёт `sections` упорядоченным **списком**, а `--from` читает **map** по id секции. Перед правкой форму надо переложить — однострочник на `jq` есть в [JSON-выводе](../guides/json-output.md#round-trip). Минимальный файл для `--from` — только изменённые секции; всё остальное берётся из дефолтов артефакта.

Сгенерировать JSON Schema для редактора / агентской валидации:

```bash
srekit postmortem --schema > postmortem.schema.json
# навести редактор (или LLM-тулинг) на файл — получишь автокомплит + валидацию
```

Проверить, что черновик заполнен по required-секциям:

```bash
srekit postmortem --validate pm.json
# OK    summary
# OK    impact
# FAIL  timeline: required body is empty
# Error: 1 of 5 required section(s) failed validation
```

Достать только метаданные для трекера:

```bash
srekit postmortem --title "API outage" --severity SEV-1 --json \
  | jq '{title: .meta.title, severity: .meta.severity, started: .meta.start}'
```

## Форма JSON

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
- Тела секций в `sections` подставляются **как есть** — без template-evaluation — так что round-trip произвольного markdown с `{{ … }}` внутри безопасен.
- Section ID, которых нет в манифесте, дают hard error со списком неизвестных ID и known-set (защита от опечаток).
- Секции, которых нет в input, заполняются дефолтами из манифеста.

## Кастомизация артефакта (v1 формат, v0.14.0+) { #customizing-the-artifact-v1-format-v0140 }

`srekit templates init <dir>` кладёт один файл `postmortem.yaml` в твой templates dir. Это **v1 артефактный формат**, появившийся в v0.14.0: header (frontmatter / H1 / meta bullets / опц. `header_body`) и секции живут в одном файле. Старый v0.13.x layout (`postmortem.md.tmpl` + `postmortem.sections.yaml`) больше не скаффолдится; user-dir файлы в том формате игнорируются с stderr-предупреждением на каждом вызове.

Редактируй `postmortem.yaml`, чтобы добавить, убрать или поменять порядок секций, изменить заголовки, поменять H1, добавить корпоративный policy-текст через `header_body`, или обновить default-наполнение (text-тела, list-айтемы, ячейки таблиц).

Полная схема артефакта (см. [`internal/sections`](https://github.com/jtprogru/srekit/tree/main/internal/sections)):

```yaml
version: 1

frontmatter:                       # free-form map; значения проходят через Go templates
  id: "{{ .Meta.ID }}"
  type: postmortem
  title: "{{ .Meta.Title }}"
  severity: "{{ .Meta.Severity }}"
  tags: [postmortem, incident]

title: "Постмортем (Postmortem) — {{ .Meta.Title }}"   # H1 (Go template)

meta_bullets:                      # bullet-строки после H1 (Go templates)
  - "**Тяжесть (Severity):** {{ .Meta.Severity }}"
  - '**Ответственный (Owner):** {{ .Meta.Owner | default "<incident owner>" }}'

header_body: |                     # опц. свободный Markdown для того, что не лезет в секции
  > **Безвинный разбор (blameless):** ищем причины в системе, не в людях.

sections:
  - id: summary                    # стабильный ID, используется в --json и --from
    title: "Краткое описание (Summary)"
    type: text                     # text | list | table
    required: true
    default_body: |
      _Один абзац: что произошло, кого затронуло, как смягчили._

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

Каждое строковое поле (значения frontmatter, `title`, `meta_bullets` items, `header_body`, default-наполнение секций и `footer_body`) проходит через тот же Go-template FuncMap, так что `{{ .Meta.X }}`, `{{ now "2006-01-02" }}`, `{{ default "x" .Y }}` и т.д. работают везде.

Порядок ключей во frontmatter сохраняется как в YAML-источнике (без alphabetical sorting на выходе), чтобы diff'ы оставались стабильными.

## См. также

- [JSON output guide](../guides/json-output.md) — кросс-командный контракт `{meta, sections}`.
