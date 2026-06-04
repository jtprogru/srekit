# srekit postmortem

Generate a **postmortem** in Google-SRE style: severity, timeline, impact, detection / mitigation / root cause, action items, lessons. Bilingual headings.

As of v0.13.0, the postmortem command is the first generator backed by a **sidecar sections manifest** (`postmortem.sections.yaml`). Document body is composed from typed sections (`text` / `list` / `table`), so `--json` exposes the structure section-by-section and `--from input.json` round-trips Markdown back from edited JSON.

## Synopsis

```bash
srekit postmortem --title TITLE [flags]
```

## Flags

| Flag | Required | Description |
|---|---|---|
| `--title` | yes (unless `--from` provides it) | Incident subject |
| `--severity` | no | `SEV-N` label (free text) |
| `--start` | no | Incident start (timestamp string) |
| `--end` | no | Incident end |
| `--owner` | no | Postmortem owner (the person writing it up) |
| `--from FILE` | no | Read structured input (`{meta?, sections}`) from a JSON file; `-` reads stdin |

Plus the [shared output flags](index.md#shared-output-flags). Default filename: `postmortem-<YYYY-MM-DD>-<slug-of-title>.md` (the date is today, in UTC if the system clock is UTC, otherwise local time).

## Examples

Short form (render Markdown, default file):

```bash
srekit postmortem --title "API outage" --severity SEV-1
```

Inspect the structured JSON shape (sections in manifest order):

```bash
srekit postmortem -T "API outage" --severity SEV-1 --json | jq '.sections[].id'
```

Round-trip: dump JSON, edit one section, re-render Markdown:

```bash
srekit postmortem -T "API outage" --json > pm.json
# edit pm.json — e.g. set sections.summary to a real summary
srekit postmortem -T "API outage" --from pm.json
```

Extract just the metadata for a tracker:

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

Sections appear in manifest order. `body` is always a string regardless of `type`; for `list` and `table` it's the rendered markdown fragment so consumers see the same value through Markdown and JSON paths.

## `--from` input format

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

- Both `meta` and `sections` are optional.
- CLI flags (`--title`, `--severity`, …) take precedence over `meta` from the file. This lets you pin a field on the command line even when reading from stdin.
- Section bodies in `sections` are used **verbatim** — no template evaluation — so you can safely round-trip arbitrary markdown containing `{{ … }}` sequences.
- Section IDs not present in the manifest cause a hard error listing the offending IDs and the known set (typo guard).
- Sections that are missing from the input fall back to the manifest defaults.

## Customizing the manifest

`srekit templates init <dir>` writes both `postmortem.md.tmpl` (the header template) and `postmortem.sections.yaml` (the section list) to your templates dir. Edit the YAML to add, remove, or reorder sections, change titles, or tweak default content (text bodies, list items, table column headers).

The manifest format (see [`internal/sections`](https://github.com/jtprogru/srekit/tree/main/internal/sections)):

```yaml
version: 1
sections:
  - id: summary           # stable ID, used in --json and --from
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

Default-body content (text bodies, list items, table cells) is evaluated through the same Go-template FuncMap as the main template, so `{{ .Meta.Start }}`, `{{ now "2006-01-02" }}`, and other [helpers](../guides/custom-templates.md#funcmap) work.

## See also

- [`srekit incident`](incident.md) — the *during*-incident doc.
- [`srekit retro`](retro.md) — sprint-level retrospective.
- [JSON output guide](../guides/json-output.md) — the cross-command `{meta, sections}` contract.
