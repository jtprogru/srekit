# JSON output for pipelines

Every generator command supports `--json`. The flag emits a structured payload instead of rendering Markdown, so agent workflows and shell pipelines can read the document field-by-field and (for `postmortem`) round-trip it back.

## Contract (v0.20.0+)

- Default sink is **stdout**. `--out FILE` writes the JSON there.
- Field names are **camelCase** across every command (`id`, `title`, `latencyTarget`, …).
- With `--json`, the Markdown default path (`investigation-<slug>.md`, `postmortem-<YYYY-MM-DD>-<slug>.md`, etc.) is **not** used — so JSON never accidentally lands in a `.md` file.
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

Output and input are not the same shape, and this is the one thing to get right: `--json` emits `sections` as an **ordered list** of section objects, while `--from` reads a **map** keyed by section ID. The list preserves manifest order on the way out; on the way in, order is irrelevant, so a map is the honest shape. Indexing the emitted list with a string (`jq '.sections.summary = …'`) fails with `Cannot index array with string`.

Convert the list to the map, edit, re-render:

```bash
# Dump
srekit postmortem -T "API outage" --severity SEV-1 --json > pm.json

# Reshape sections into the --from map, then set one body
jq '{meta, sections: (.sections | map({key: .id, value: .body}) | from_entries)}
    | .sections.summary = "27-minute checkout 5xx, mitigated by failing back to cache."' \
  pm.json > pm.edited.json

# Re-render
srekit postmortem -T "API outage" --from pm.edited.json
```

You do not have to send every section back. `--from` overlays whatever it is given onto the artifact's defaults, so the minimal edited file is just the sections you changed:

```json
{ "sections": { "summary": "27-minute checkout 5xx, mitigated by failing back to cache." } }
```

### Round-trip a changelog

`changelog` accepts `--from` too, with the same payload shape:

```bash
srekit changelog --repo acme/api --json > cl.json
jq '{meta, sections: (.sections | map({key: .id, value: .body}) | from_entries)}
    | .sections.unreleased = "### Added\n\n- Structured input for changelog.\n"' \
  cl.json > cl.edited.json
srekit changelog --from cl.edited.json
```

`meta` in the payload supplies `repo`, `initialVersion` and `today`; flags win over the file, and the file wins over the git remote. Unlike `postmortem`, `changelog` has no `--schema` and no `--validate` — its artifact declares no required sections, so payload validation could only ever pass.

### What is not in `sections`

An artifact's `footer_body` — trailing document-level material such as the changelog's link reference definitions — is **not** a section. It has no `id`, never appears in the `sections` array, and cannot be targeted through `--from`. It is rendered from the artifact on every invocation, which is exactly why replacing a section body cannot drop the changelog's compare links.

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
// task            — 6 sections
{ "meta": { "id", "title", "creationDate", "modificationDate" } }

// postmortem      — 12 sections
{ "meta": { "id", "title", "severity", "start", "end", "owner", "now" } }

// rfc             — 5 sections
{ "meta": { "id", "title", "status", "now", "author": { "name", "email" } } }

// runbook         — 7 sections
{ "meta": { "id", "title", "service", "alert", "now" } }

// slo             — 7 sections
{ "meta": { "id", "service", "target", "window", "latencyTarget", "now" } }

// ebp             — 7 sections
{ "meta": { "id", "service", "now" } }

// oncall-report   — 8 sections
{ "meta": { "id", "team", "start", "end", "now", "author": { "name", "email" } } }

// changelog       — 2 sections
{ "meta": { "repo", "initialVersion", "today" } }
```

`sections` is omitted above for brevity — every payload carries it, as the list declared in the corresponding `internal/tmpl/templates/<name>.yaml`. To see the ids a command actually ships, ask the binary rather than this page:

```bash
srekit runbook -T X --json | jq -r '.sections[].id'
```

`author` (where present) is a nested object (`{ "name", "email" }`), addressed as `.meta.author.name` / `.meta.author.email`.

## When to use `--json`

- **Agent workflows**: read a section, modify it, write it back. Postmortem and changelog support this — `--from` round-trip works out of the box.
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
