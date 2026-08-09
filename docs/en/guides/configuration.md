# Configuration & precedence

srekit reads configuration from four sources. Most users only need one of them. This page is the authoritative source for "which wins when."

## Sources

| Source | Set via |
|---|---|
| **Flags** | `--author`, `--email`, `--templates-dir`, etc. on the command line |
| **Environment** | `SREKIT_AUTHOR`, `SREKIT_EMAIL`, `SREKIT_TEMPLATES_DIR` |
| **Config file** | `$XDG_CONFIG_HOME/srekit/config.yaml`, edited by hand or via `srekit config init` |
| **`git config`** | `user.name`, `user.email` from your global or local git config |

## Precedence

Per key, srekit walks the sources in this order and uses the first non-empty value:

1. Command-line flag
2. `SREKIT_<KEY>` env var (e.g. `SREKIT_AUTHOR`)
3. The config file (e.g. `author:`)
4. `git config <git-key>` (only for author/email)

If all four are empty for a required value, the command exits with a clear error:

```bash
srekit rfc --title "Move to gRPC"
# Error: author is not set: pass --author, set SREKIT_AUTHOR, or configure git user.name
```

## Keys

### Author identity

Used by: `rfc`, `oncall-report` (others fall back to "anonymous" where appropriate).

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

### Changelog language

Used by [`srekit changelog`](../commands/changelog.md#the-russian-variant) and inherited by its `release` and `validate` subcommands.

| Key | yaml | env | flag |
|---|---|---|---|
| changelog_lang | `changelog_lang:` | `SREKIT_CHANGELOG_LANG` | `--lang` |

Accepts `en` (the default) or `ru`. An unrecognized value fails, naming the accepted ones, before anything is written — a typo here does not silently fall back to English. The setting governs what is generated; it never influences how an existing changelog is parsed.

### Config file location

| Key | flag | default |
|---|---|---|
| config file | `--config FILE` | `$XDG_CONFIG_HOME/srekit/config.yaml` |

srekit follows the XDG Base Directory Specification for fresh installs, but a pre-XDG path wins if it already exists: `~/.srekit.yaml` for the config and `~/.srekit/templates` for the templates directory. That way an upgrade never leaves you with a config file that sits there unread. When both locations exist, the legacy one is used and [`srekit doctor`](../commands/doctor.md) warns which one is being ignored.

`srekit config init` honors `--config` too — pass it to write the file elsewhere.

## The yaml file

```yaml
# ~/.config/srekit/config.yaml
author: Mikhail Savin
email: jtprogru@gmail.com
# templates_dir: ~/.config/srekit/templates   # optional
# changelog_lang: ru                   # optional, default: en
```

Generate it with [`srekit config init`](../commands/config.md). The file is written `0o600` (user-only) and uses tilde-style home expansion for paths (`~/foo` resolves to `$HOME/foo`).

## Example: per-environment overrides

Same machine, two GitHub identities (personal and work):

```bash
# The config file has personal identity baked in.
# At work:
SREKIT_AUTHOR="Mikhail Savin" SREKIT_EMAIL="m.savin@work.example.com" \
  srekit rfc --title "Move checkout to gRPC"
```

Or scope a custom templates dir per project:

```bash
srekit --templates-dir ./project-templates rfc --title "Migrate to gRPC"
```

## Debugging precedence

Want to know which source srekit picked? Ask it:

```bash
srekit doctor
```

`config.file` names the config actually being read, `config.env` lists every `SREKIT_`-prefixed variable currently in effect, `config.templates-dir` reports the resolved templates directory *and which source supplied it*, and `config.identity` reports the resolved author name and email with the origin of each value. `config.shadowed` fires when both an XDG and a legacy path exist, so a config you edited but nobody reads shows up as a warning rather than as a mystery. See [`srekit doctor`](../commands/doctor.md).

## See also

- [`srekit doctor`](../commands/doctor.md) — read-only report of everything this page describes as resolved.
- [`srekit config init`](../commands/config.md#config-init) — write the yaml file interactively.
- [Custom templates workflow](custom-templates.md) — `templates_dir` start to finish.
