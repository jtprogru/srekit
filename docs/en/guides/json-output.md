# JSON output for pipelines

Every generator command supports `--json`. The flag emits a structured payload instead of rendering Markdown, so agent workflows and shell pipelines can read the document field-by-field and (for `postmortem`) round-trip it back.

## Contract (v0.20.0+)

- Default sink is **stdout**. `--out FILE` writes the JSON there.
- Field names are **camelCase** across every command (`id`, `title`, `latencyTarget`, …).
- With `--json`, the Markdown default path (`Tasker - <title>.md`, `postmortem-<YYYY-MM-DD>-<slug>.md`, etc.) is **not** used — so JSON never accidentally lands in a `.md` file.
- **Every payload has the shape `{meta, sections}`.** Metadata lives under `meta`; the rendered document is a list of typed sections under `sections`. Each section has `{id, title, type, required, body}` with `type` one of `text` / `list` / `table` and `body` always a string.
- **Sections are per-artifact**, in manifest order, with stable IDs. Access them via `jq '.sections[] | select(.id == "<id>").body'` — never by index.

| Mode | Used by | Sections |
|---|---|---|
| **Structured** | every generator (postmortem, task, slo, ebp, rfc, runbook, oncall-report, changelog) | Multiple typed sections — one per slot in the artifact YAML — in declared order. |

!!! warning "Migration from pre-v1 layouts"
    - **0.12.x → 0.13.0**: shape changed from the flat `{title, severity, …}` to `{meta, sections}`. Migration: `jq '.title'` → `jq '.meta.title'`.
    - **0.13.x → v0.20**: the YAML-first migration retired the bootstrap envelope (`sections: [{id: "body", body: <markdown>}]`) for every generator. Migration: replace `jq '.sections[0].body'` with `jq '.sections[] | select(.id == "<id>").body'` for the section you actually want. See `docs/{en,ru}/migration/v1.md` for per-release section IDs.

## Patterns

### Extract a field from `meta`

```bash
srekit task --title "Tail latency" --json | jq -r '.meta.id'
# 085883a2-32d0-4d50-9bc6-ac219e29409c
```

### List sections of a postmortem

```bash
srekit postmortem -T X --json | jq '.sections[] | {id, type, required}'
```

### Get one section's body

```bash
# Postmortem — pull the summary
srekit postmortem -T X --json | jq -r '.sections[] | select(.id == "summary").body'

# Runbook — pull the diagnose section
srekit runbook --title "p99 spike" --service api-gw --alert APIHighLatency --json |
  jq -r '.sections[] | select(.id == "diagnose").body'

# Changelog — pull the initial-release section
srekit changelog --repo owner/repo --json |
  jq -r '.sections[] | select(.id == "initial_release").body'
```

### Round-trip a postmortem

```bash
# Dump
srekit postmortem -T "API outage" --severity SEV-1 --json > pm.json

# Edit one section in pm.json (any tool — jq, sed, your editor, an LLM)
jq '.sections.summary = "27-minute checkout 5xx, mitigated by failing back to cache."' pm.json > pm.edited.json

# Re-render
srekit postmortem -T "API outage" --from pm.edited.json
```

Note that `--from` reads sections from a **map** keyed by ID (`{"summary": "…"}`), not a list. The JSON output uses a list to preserve manifest order; `--from` uses a map because order is irrelevant on the input side.

### Project to your own shape

```bash
srekit postmortem --title "API outage" --severity SEV-1 --json |
  jq '{title: .meta.title, severity: .meta.severity, started: .meta.start, owner: .meta.owner}'
```

### Drive another tool

```bash
srekit slo --service api-gw --target 99.95% --window 30d --json |
  jq -r '.meta | "\(.service) \(.target) \(.window)"' |
  xargs my-slo-registrar register
```

### Compare two generations

```bash
diff <(srekit slo --service api-gw --target 99.9% --json) \
     <(srekit slo --service api-gw --target 99.95% --json)
```

### Render JSON to a file

`--json` honors `--out`:

```bash
srekit oncall-report --team platform --json --out oncall.json
```

`--dry-run` works too — prints "would write N bytes to oncall.json" plus the body.

## Per-command payload shape

Every generator is on the v1 artifact path: `meta` mirrors the per-command flag set, `sections` is the list declared in the artifact YAML. Template authors address meta fields as `.Meta.<Field>` inside the YAML; `--json` emits camelCase under `meta`.

```jsonc
// task
{ "meta": { "id", "title", "now" },
  "sections": [ ...sections from internal/tmpl/templates/task.yaml... ] }

// postmortem
{ "meta": { "id", "title", "severity", "start", "end", "owner", "now" },
  "sections": [ ...sections from postmortem.yaml (12 by default)... ] }

// rfc
{ "meta": { "id", "title", "status", "now", "author": { "name", "email" } },
  "sections": [ ...sections from rfc.yaml... ] }

// slo
{ "meta": { "id", "service", "target", "window", "latencyTarget", "now" },
  "sections": [ ...sections from slo.yaml... ] }
```

`author` (where present) is a nested object (`{ "name", "email" }`), addressed as `.meta.author.name` / `.meta.author.email`. Browse `internal/tmpl/templates/<name>.yaml` for the exact section IDs each command ships.

## When to use `--json`

- **Agent workflows**: read a section, modify it, write it back. Postmortem is the first command optimized for this — `--from` round-trip works out of the box.
- Scripting / automation: any time you'd otherwise `grep` Markdown to pluck a field, use `--json | jq` instead.
- Drift checks: stash the JSON output of a previous generation and diff against a new one to detect template/field changes.
- Cross-tool integration: feed values directly into Linear, Jira, or internal CLIs.

## When NOT to use `--json`

- You want the rendered document — that's the default mode.
- You want to embed the document in another file — also the default, with `--stdout` to pipe.
- You need to share with a non-engineer — Markdown reads better than JSON.

## See also

- [Recipes](../recipes.md) — concrete `--json` pipelines.
- [`srekit postmortem`](../commands/postmortem.md) — the first structured generator; documents `--from` in detail.
- [`templates list --json`](../commands/templates.md#templates-list) — introspection JSON (same camelCase keys, flat list shape — different from generator output).
