# JSON output for pipelines

Every generator command supports `--json`. The flag short-circuits the template engine: instead of rendering Markdown, srekit emits the template's data payload as indented JSON. The payload is whatever the underlying Go template would have seen.

## Contract

- Default sink is **stdout**. `--out FILE` writes the JSON there.
- Field names are **PascalCase** (Go field names with no `json:` tags).
- With `--json`, the Markdown default path (`Tasker - <title>.md`, `postmortem-<slug>.md`, etc.) is **not** used — so JSON never accidentally lands in a `.md` file.

!!! note "Two JSON contracts in v0.x"
    Generators emit **PascalCase** keys. Introspection commands
    (`templates list --json`) emit **camelCase** keys because they go
    through tagged structs that satisfy the `tagliatelle` linter. The
    two will harmonize in v1.0; until then, the rule of thumb is "JSON
    from `--json` on a generator is PascalCase, JSON from a
    `templates list --json` or any future introspection is camelCase."

## Patterns

### Extract a single field

```bash
srekit task --title "Tail latency" --json | jq -r '.ID'
# 085883a2-32d0-4d50-9bc6-ac219e29409c
```

### Project to your own shape

```bash
srekit postmortem --title "API outage" --severity SEV-1 --json |
  jq '{title: .Title, severity: .Severity, started: .Start, owner: .Owner}'
```

### Drive another tool

Generate an SLO, take the params, register them with a metrics tool:

```bash
srekit slo --service api-gw --target 99.95% --window 30d --json |
  jq -r '"\(.Service) \(.Target) \(.Window)"' |
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

The full Go struct passed to each template is listed on its command page (under "Template shape"). Here are the most-common ones:

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

`Author` is a nested object (Go struct `meta.Author{Name, Email}`), addressed as `.Author.Name` / `.Author.Email` in `jq`.

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
- [`templates list --json`](../commands/templates.md#templates-list) — the introspection JSON (camelCase keys).
