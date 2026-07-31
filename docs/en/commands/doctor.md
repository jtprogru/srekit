# srekit doctor

Report the state srekit resolves before it renders anything: which config file is actually read (and whether a second one is being shadowed), where the templates directory resolves and whether its artifacts still parse, whether an author identity can be resolved at all, and whether `git` is on `PATH`.

```bash
srekit doctor                                   # full report
srekit doctor --quiet                           # only what needs attention
srekit doctor --json | jq -e '.status != "error"'   # gate CI on it
```

`doctor` is read-only. It never creates, changes, or repairs anything — a check whose subject is missing reports it, it does not create it. It makes no network request, including no check for a newer srekit release. `git` is the only external program it looks for.

## Flags

| Flag | Effect |
|---|---|
| `--json` | emit the findings as a JSON document instead of a table |
| `--config FILE` | inherited from the root command; changes which config file is inspected |
| `--templates-dir DIR` | inherited from the root command; changes which templates directory is inspected |
| `--quiet` / `-q` | print only `warn` and `error` findings, and drop the summary line |

There is no `--out`, `--stdout`, `--force`, or `--dry-run`: `doctor` writes nothing, and a flag a command would ignore must not exist.

## Statuses and exit code

| Status | Meaning |
|---|---|
| `ok` | nothing to do |
| `warn` | srekit works, but something is being ignored, is outdated, or is about to break |
| `error` | a generator will fail or produce wrong output in this environment |

Exit status is `1` when at least one check reports `error`, `0` otherwise. **A `warn` never fails the run**, so a team can adopt `doctor` in CI without being blocked by advisory findings. `--quiet` does not change the exit status — in a healthy environment `srekit doctor --quiet` prints nothing and exits `0`, so silence means healthy.

A check that cannot inspect its subject — an unreadable directory, a subprocess that will not start — reports that as its own finding. Every check always runs; one broken check never hides the others.

## Checks

Check identifiers are a stable public contract: CI gates on them, so renaming one is a breaking change.

### `config`

| Identifier | Reports |
|---|---|
| `config.file` | The config file that will actually be read, and whether it exists — naming the XDG or the legacy location, or an explicit `--config`. Absence is the documented default on a fresh install, not a defect, so a missing file is `ok`. |
| `config.parse` | Whether that file parses. A malformed config is `warn`, not `error`: it never fails a command that needs nothing from it, which is why nothing else reports it. |
| `config.shadowed` | `warn` when both `~/.srekit.yaml` and `$XDG_CONFIG_HOME/srekit/config.yaml` exist, naming which one wins and which is never read. |
| `config.writable` | Whether the directory holding the resolved config path is writable, so you learn before running `config init` that it is not. |
| `config.env` | Every `SREKIT_`-prefixed environment variable currently in effect, by name. |
| `config.templates-dir` | The resolved templates directory, the source that supplied it (`--templates-dir`, `SREKIT_TEMPLATES_DIR`, the config file, or the built-in default), and whether it exists and is a directory. A configured directory that is missing is `warn`: generation still works from the embedded set, but your overrides are silently not applied. |
| `config.templates-shadowed` | `warn` when both `~/.srekit/templates` and `$XDG_CONFIG_HOME/srekit/templates` exist, naming which one the `templates` subcommands operate on. |
| `config.identity` | Whether an author name and email can be resolved at all, and the source each value came from. `error` when either is missing — every generator that stamps an author fails in that environment. Values are printed as they are: they are already stamped into every artifact srekit generates, and redacting them would make the check useless for diagnosing a wrong author. |

### `templates`

These three report `ok` with an embedded-only summary when no templates directory is configured or the configured one does not exist.

| Identifier | Reports |
|---|---|
| `templates.parse` | How many artifacts in the directory fail to parse, naming each file and its parse error. Any parse failure is `error`: the generator behind that artifact cannot render. Uses the same parser as [`srekit templates validate`](templates.md#templates-validate). |
| `templates.legacy` | How many pre-v1.0 template files (`.tmpl`, `.sections.yaml`) are present, naming them. `warn`, with `srekit templates migrate` as the remedy. |
| `templates.drift` | How many artifacts differ from this binary's embedded version, and how many embedded artifacts are absent from the directory. `warn` when either count is non-zero, with `srekit templates diff` and `srekit templates upgrade` as remedies. Uses the same classification as [`srekit templates list`](templates.md#templates-list), so the two can never disagree. |

### `dependencies`

| Identifier | Reports |
|---|---|
| `dependencies.git` | Whether `git` is on `PATH`, and if so its resolved path and reported version. An absent `git` is `warn`, not `error`: author metadata and the changelog repository slug fall back to flags and config, so most generation still works. |

## Text output

```text
CONFIG
  ok     config.file                no config file at /home/u/.config/srekit/config.yaml (XDG location); flags, environment and defaults supply everything
  warn   config.templates-dir       /home/u/tpl (from SREKIT_TEMPLATES_DIR) cannot be read: no such file or directory; generation is falling back to the embedded templates
                                    fix: run 'srekit templates init /home/u/tpl', or point SREKIT_TEMPLATES_DIR somewhere that exists
  ...

DEPENDENCIES
  ok     dependencies.git           /usr/bin/git (git version 2.51.0)

12 checks: 11 ok, 1 warn, 0 error
```

Findings are grouped by category and always appear in the same order, so two runs against an unchanged environment produce identical output. Every `warn` and `error` carries a remedy naming the command or setting that fixes it, printed on its own continuation line. The trailing line reports the count per status.

The status is conveyed by the word itself, so the output stays legible when piped. Colour is used only when stdout is a terminal, and is suppressed when `NO_COLOR` is set and non-empty.

## JSON output

`--json` emits one indented, newline-terminated document to stdout and nothing else there. Keys are `camelCase`. The exit status is the same as for text output, so `--json` can be piped to a parser and still gate CI.

```json
{
  "status": "warn",
  "checks": [
    {
      "id": "config.file",
      "category": "config",
      "status": "ok",
      "summary": "reading /home/u/.srekit.yaml (legacy location)"
    },
    {
      "id": "dependencies.git",
      "category": "dependencies",
      "status": "warn",
      "summary": "git is not on PATH; author name and email fall back to --author/--email and the config file, and 'srekit changelog' cannot detect the repository slug",
      "remedy": "install git, or pass --author/--email and --repo OWNER/REPO explicitly"
    }
  ]
}
```

`status` is the worst status among the findings. `remedy` is present whenever the status is `warn` or `error`.

`--quiet` has no effect on `--json`: the document is data, and a consumer that asked for the full structure and got a subset of it has a bug on its hands rather than a preference.

## See also

- [Configuration & precedence](../guides/configuration.md) — the resolution rules `doctor` reports on.
- [`srekit templates`](templates.md) — the commands most `templates` remedies point at.
- [`srekit config init`](config.md#config-init) — the remedy for an unresolvable author identity.
