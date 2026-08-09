# srekit postmortem

Generate a **postmortem** in Google-SRE style: severity, timeline, impact, detection / mitigation / root cause, action items, lessons. Bilingual headings.

The document body is composed from typed sections (`text` / `list` / `table`) declared in the v1 artifact `postmortem.yaml`, so `--json` exposes the structure section-by-section and `--from input.json` round-trips Markdown back from edited JSON. Postmortem was the prototype for the v1 artifact format (introduced in v0.14.0) and is now the canonical schema reference for the rest of the generators — its "Customizing the artifact" section below documents every field.

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
| `--schema` | no | Emit a JSON Schema describing the `--from` payload, derived from the loaded manifest. Output goes to stdout. Mutually exclusive with `--validate`. |
| `--validate FILE` | no | Validate an input file: required sections must have a non-empty body; unknown section IDs are rejected. Prints a per-section OK / FAIL report; exits non-zero on any failure. |

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
# edit pm.json, then feed it back
srekit postmortem -T "API outage" --from pm.json
```

`--json` emits `sections` as an ordered **list**; `--from` reads a **map** keyed by section id. Reshape before editing — [JSON output](../guides/json-output.md#round-trip-a-postmortem) has the one-line `jq` for it. The minimal `--from` file is just the sections you changed; everything else falls back to the artifact defaults.

Generate a JSON Schema for editor tooling / agent input validation:

```bash
srekit postmortem --schema > postmortem.schema.json
# point your editor (or an LLM tool) at the file to get autocomplete + validation
```

Validate that a draft has all required sections filled in:

```bash
srekit postmortem --validate pm.json
# OK    summary
# OK    impact
# FAIL  timeline: required body is empty
# Error: 1 of 5 required section(s) failed validation
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
    { "id": "summary",         "type": "text",  "required": true,  "title": "…", "body": "…" },
    { "id": "impact",          "type": "list",  "required": true,  "title": "…", "body": "- …" },
    { "id": "timeline",        "type": "table", "required": true,  "title": "…", "body": "| … |" },
    { "id": "root_cause",      "type": "text",  "required": true,  "title": "…", "body": "…" },
    { "id": "detection",       "type": "list",  "required": false, "title": "…", "body": "- …" },
    { "id": "resolution",      "type": "list",  "required": false, "title": "…", "body": "- …" },
    { "id": "what_went_well",  "type": "list",  "required": false, "title": "…", "body": "- " },
    { "id": "what_went_wrong", "type": "list",  "required": false, "title": "…", "body": "- " },
    { "id": "where_got_lucky", "type": "list",  "required": false, "title": "…", "body": "- " },
    { "id": "action_items",    "type": "table", "required": true,  "title": "…", "body": "| … |" },
    { "id": "lessons_learned", "type": "list",  "required": false, "title": "…", "body": "- " },
    { "id": "references",      "type": "list",  "required": false, "title": "…", "body": "- " }
  ]
}
```

Sections appear in manifest order. `body` is always a string regardless of `type`; for `list` and `table` it's the rendered markdown fragment so consumers see the same value through Markdown and JSON paths.

`title` is elided above because it renders bilingually — Russian, with the English term in parentheses — and the ids are what you address. To see the titles as they ship:

```bash
srekit postmortem -T X --json | jq -r '.sections[] | "\(.id)\t\(.title)"'
```

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

## Customizing the artifact (v1 format, v0.14.0+)

`srekit templates init <dir>` writes a single `postmortem.yaml` to your templates dir. This is the **v1 artifact format** introduced in v0.14.0: header (frontmatter / H1 / meta bullets / optional `header_body`) and sections live in the same file. The previous v0.13.x layout (`postmortem.md.tmpl` + `postmortem.sections.yaml`) is no longer scaffolded; user-dir files in that format are ignored with a stderr warning on every invocation.

Edit `postmortem.yaml` to add, remove, or reorder sections, change titles, swap the H1, add organizational policy text via `header_body`, or tweak default content (text bodies, list items, table cells).

The full artifact schema (see [`internal/sections`](https://github.com/jtprogru/srekit/tree/main/internal/sections)):

```yaml
version: 1

frontmatter:                       # free-form map; values run through Go templates
  id: "{{ .Meta.ID }}"
  type: postmortem
  title: "{{ .Meta.Title }}"
  severity: "{{ .Meta.Severity }}"
  tags: [postmortem, incident]

title: "Postmortem — {{ .Meta.Title }}"                # H1 (Go template)

meta_bullets:                      # bullet lines after the H1 (Go templates)
  - "**Severity:** {{ .Meta.Severity }}"
  - '**Owner:** {{ .Meta.Owner | default "<incident owner>" }}'

header_body: |                     # optional freeform Markdown escape hatch
  > **Blameless review:** look for causes in the system, not in people.

sections:
  - id: summary                    # stable ID, used in --json and --from
    title: "Summary"
    type: text                     # text | list | table
    required: true
    default_body: |
      _One paragraph: what happened, who was affected, how it was mitigated._

  - id: timeline
    title: "Timeline"
    type: table
    required: true
    default_body: "(All times in UTC.)"
    columns: ["Time", "Event"]
    rows:
      - ["{{ .Meta.Start }}", "Incident started"]
      - ["{{ .Meta.End }}",   "Incident resolved"]

footer_body: |                     # optional trailing document-level material
  [ref]: https://example.internal/runbooks
```

Every string field (frontmatter values, `title`, `meta_bullets` items, `header_body`, section default content, and `footer_body`) is evaluated through the shared Go-template FuncMap, so `{{ .Meta.X }}`, `{{ now "2006-01-02" }}`, `{{ default "x" .Y }}`, etc. work everywhere.

Frontmatter key order is preserved verbatim from the YAML source (no alphabetical sorting on output) so diffs stay stable.

## See also

- [JSON output guide](../guides/json-output.md) — the cross-command `{meta, sections}` contract.
