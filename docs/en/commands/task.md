# srekit task

Generate an **investigation log** — a structured artifact for tracking the hypothesis-and-evidence trail when you're hunting a tail-latency spike, a flaky test, or any open-ended SRE puzzle. Hidden alias: `srekit sretask` (kept for migration from `gch sretask`).

## Synopsis

```bash
srekit task --title TITLE [flags]
```

## Flags

| Flag | Required | Description |
|---|---|---|
| `--title` | yes | Subject of the investigation; used in the H1 and the default filename |
| `--path DIR` | no | Directory to write into (default: current dir) |

Plus the [shared output flags](index.md#shared-output-flags): `--out`, `--stdout`, `--force`, `--dry-run`, `--template`, `--json`.

## Default filename

If you pass neither `--out` nor `--stdout`, srekit writes to `<path>/Tasker - <title>.md` (slug-cleaned, preserving the original title casing).

## Examples

Quick scratch into stdout:

```bash
srekit task --title "Tail latency on api-gw" --stdout
```

Write into a specific directory:

```bash
srekit task --title "Tail latency on api-gw" --path ./tasks
# → ./tasks/Tasker - Tail latency on api-gw.md
```

Pipe into `jq` to grab the generated UUID:

```bash
srekit task --title "Tail latency on api-gw" --json | jq -r '.id'
```

## Template shape

The data passed to the template:

```go
struct {
    ID, CreationDate, ModificationDate, Title string
}
```

Section structure (post-render): YAML front matter (`title`, `tags`, `creation_date`, `id`) followed by `Контекст / Context`, `Гипотеза / Hypothesis`, `Доказательства / Evidence`, `Выводы / Findings`, `Дальнейшие действия / Action items`, `Ссылки / References`.

## See also

- [Custom templates workflow](../guides/custom-templates.md) — override the embedded template with your own.
- [JSON output](../guides/json-output.md) — pipe `--json` into other tools.
