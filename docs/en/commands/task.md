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

Plus the [shared output flags](index.md#shared-output-flags): `--out`, `--stdout`, `--force`, `--dry-run`, `--json`.

## Default filename

If you pass neither `--out` nor `--stdout`, srekit writes to `<path>/investigation-<slug>.md` (lowercased, slug-cleaned).

## Examples

Quick scratch into stdout:

```bash
srekit task --title "Tail latency on api-gw" --stdout
```

Write into a specific directory:

```bash
srekit task --title "Tail latency on api-gw" --path ./tasks
# → ./tasks/investigation-tail-latency-on-api-gw.md
```

Pipe into `jq` to grab the generated UUID:

```bash
srekit task --title "Tail latency on api-gw" --json | jq -r '.meta.id'
```

## Template shape

`task` ships as a v1 YAML artifact (`internal/tmpl/templates/task.yaml`) — frontmatter (`id`, `creation_date`, `modification_date`, `type: investigation`, `title`, `tags`), H1, meta_bullets, and six sections: `context` (Контекст / Context), `hypothesis` (Гипотезы / Hypothesis), `evidence` (Наблюдения / Evidence), `findings` (Выводы / Findings), `action_items` (Задачи / Action items), `references` (Ссылки / References). Template expressions inside the YAML reference `.Meta.<Field>`; the available fields are `ID`, `Title`, `CreationDate`, `ModificationDate`.

## See also

- [Custom templates workflow](../guides/custom-templates.md) — override the embedded artifact with your own `task.yaml`.
- [JSON output](../guides/json-output.md) — pipe `--json` into other tools (per-section access via `jq '.sections[] | select(.id=="…").body'`).
