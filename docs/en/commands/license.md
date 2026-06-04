# srekit license

Generate a `LICENSE` file (default: WTFPL). Author and email are resolved from flags / env / yaml / `git config` in that order. Hidden alias: `srekit lic`.

## Synopsis

```bash
srekit license [flags]
```

## Flags

| Flag | Required | Description |
|---|---|---|
| `--type` | no | `wtfpl` (default), `mit`, or `apache2` |
| `--year` | no | Copyright year (default: current year) |
| `--author` | no | Override author name |
| `--email` | no | Override author email |
| `--template FILE` | no | Use a custom license body from FILE instead of the inlined `wtfpl` / `mit` / `apache2` text. Honors the same FuncMap as everything else. **License is the only command that honors `--template`** — the other generators are on the v1 artifact path and ignore it. |

Plus the [shared output flags](index.md#shared-output-flags). By default this command prints to **stdout** (it has no default file path) — pass `--out LICENSE` to write the file.

## Examples

WTFPL to stdout:

```bash
srekit license --stdout
# or simply: srekit license  (stdout is the default sink)
```

MIT to file:

```bash
srekit license --type mit --out LICENSE
```

Apache 2.0 with explicit year and author:

```bash
srekit license --type apache2 --year 2026 --author "Mikhail Savin" \
  --email jtprogru@gmail.com --out LICENSE
```

Override author resolution one-off:

```bash
SREKIT_AUTHOR="Alice Example" SREKIT_EMAIL=alice@example.com \
  srekit license --type mit --stdout
```

## Author resolution

If `--author` is not set, srekit walks (first match wins):

1. `SREKIT_AUTHOR` env / `author:` in `~/.srekit.yaml` / `full_name:` in yaml
2. `git config user.name`

Same chain for `--email` with `SREKIT_EMAIL` / `email:` / `git config user.email`. If both are still empty, the command fails with a clear "set --author or configure git user.name" error.

## Template shape

License is the **only** generator that has not migrated to the v1 YAML artifact format. The three license bodies (`mit`, `apache2`, `wtfpl`) are inlined as Go-string constants in `cmd/license.go` and rendered via plain `text/template` with the shared FuncMap. The template receives `{Year int, Author meta.Author}` where `meta.Author` is `{Name, Email string}` — reference fields as `{{ .Year }}`, `{{ .Author.Name }}`, `{{ .Author.Email }}` inside a `--template FILE` body.

## See also

- [Configuration & precedence](../guides/configuration.md)
- [`srekit config init`](config.md#config-init) to seed `~/.srekit.yaml`.
