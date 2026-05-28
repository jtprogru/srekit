# JSON output for pipelines

Every generator command supports `--json`. The flag short-circuits the template engine: instead of rendering Markdown, srekit emits the template's data payload as indented JSON. The payload is whatever the underlying Go template would have seen.

## Contract

- Default sink is **stdout**. `--out FILE` writes the JSON there.
- Field names are **camelCase** across every command (`id`, `title`, `latencyTarget`, …).
- With `--json`, the Markdown default path (`Tasker - <title>.md`, `postmortem-<slug>.md`, etc.) is **not** used — so JSON never accidentally lands in a `.md` file.

!!! note "One JSON contract"
    Every command — generators and introspection alike
    (`templates list --json`) — emits **camelCase** keys. Earlier 0.x
    releases split generators (PascalCase) from `templates list`
    (camelCase); that split is gone, so a single `jq` convention works
    everywhere.

## Patterns

### Extract a single field

```bash
srekit task --title "Tail latency" --json | jq -r '.id'
# 085883a2-32d0-4d50-9bc6-ac219e29409c
```

### Project to your own shape

```bash
srekit postmortem --title "API outage" --severity SEV-1 --json |
  jq '{title: .title, severity: .severity, started: .start, owner: .owner}'
```

### Drive another tool

Generate an SLO, take the params, register them with a metrics tool:

```bash
srekit slo --service api-gw --target 99.95% --window 30d --json |
  jq -r '"\(.service) \(.target) \(.window)"' |
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

The full struct passed to each template is listed on its command page (under "Template shape"). Template authors address fields by their Go names (`.Title`); `--json` emits the camelCase keys below:

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

`author` is a nested object (`{ "name", "email" }`), addressed as `.author.name` / `.author.email` in `jq`.

## When to use `--json`

- Scripting / automation: any time you'd otherwise `grep` Markdown to pluck a field, use `--json | jq` instead — far more reliable.
- Drift checks: stash the JSON output of a previous generation and diff against a new one to detect template/field changes.
- Cross-tool integration: feed values directly into Linear, Jira, or internal CLIs.

## When NOT to use `--json`

- You want the rendered document — that's the default mode.
- You want to embed the document in another file — also the default, with `--stdout` to pipe.
- You need to share with a non-engineer — Markdown reads better than JSON.

## See also

- [Recipes](../recipes.md) — concrete `--json` pipelines.
- [`templates list --json`](../commands/templates.md#templates-list) — the introspection JSON (same camelCase keys).
