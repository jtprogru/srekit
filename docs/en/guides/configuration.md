# Configuration & precedence

srekit reads configuration from four sources. Most users only need one of them. This page is the authoritative source for "which wins when."

## Sources

| Source | Set via |
|---|---|
| **Flags** | `--author`, `--email`, `--templates-dir`, etc. on the command line |
| **Environment** | `SREKIT_AUTHOR`, `SREKIT_EMAIL`, `SREKIT_TEMPLATES_DIR` |
| **`~/.srekit.yaml`** | Edited by hand or via `srekit config init` |
| **`git config`** | `user.name`, `user.email` from your global or local git config |

## Precedence

Per key, srekit walks the sources in this order and uses the first non-empty value:

1. Command-line flag
2. `SREKIT_<KEY>` env var (e.g. `SREKIT_AUTHOR`)
3. `~/.srekit.yaml` (e.g. `author:`)
4. `git config <git-key>` (only for author/email)

If all four are empty for a required value, the command exits with a clear error:

```bash
srekit license --stdout
# Error: author is not set: pass --author, set SREKIT_AUTHOR, or configure git user.name
```

## Keys

### Author identity

Used by: `license`, `rfc`, `oncall-report` (others fall back to "anonymous" where appropriate).

| Key | yaml | env | git |
|---|---|---|---|
| name | `author:` (or `full_name:`) | `SREKIT_AUTHOR` | `user.name` |
| email | `email:` | `SREKIT_EMAIL` | `user.email` |

### Templates directory

Used by every `templates *` subcommand and by every generator (via the overlay loader).

| Key | yaml | env | flag |
|---|---|---|---|
| templates_dir | `templates_dir:` | `SREKIT_TEMPLATES_DIR` | `--templates-dir` |

The flag is a **persistent flag** on the root command — it applies to every subcommand.

### Config file location

| Key | flag | default |
|---|---|---|
| config file | `--config FILE` | `~/.srekit.yaml` |

`srekit config init` honors `--config` too — pass it to write the file elsewhere.

## The yaml file

```yaml
# ~/.srekit.yaml
author: Mikhail Savin
email: jtprogru@gmail.com
# templates_dir: ~/.srekit/templates   # optional
```

Generate it with [`srekit config init`](../commands/config.md). The file is written `0o600` (user-only) and uses tilde-style home expansion for paths (`~/foo` resolves to `$HOME/foo`).

## Example: per-environment overrides

Same machine, two GitHub identities (personal and work):

```bash
# ~/.srekit.yaml has personal identity baked in.
# At work:
SREKIT_AUTHOR="Mikhail Savin" SREKIT_EMAIL="m.savin@work.example.com" \
  srekit license --type apache2 --out LICENSE
```

Or scope a custom templates dir per project:

```bash
srekit --templates-dir ./project-templates rfc --title "Migrate to gRPC"
```

## Debugging precedence

Want to know which source srekit picked? Run with the same args and diff the output, OR just check the env / yaml / git config in order. There's no built-in "show resolved config" command yet — that's a v1.0 candidate.

## See also

- [`srekit config init`](../commands/config.md#config-init) — write the yaml file interactively.
- [Custom templates workflow](custom-templates.md) — the `templates_dir` end-to-end.
