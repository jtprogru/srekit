# Commands overview

srekit's surface is a flat tree of cobra subcommands. Every generator
command produces a single artifact (a Markdown or LICENSE file) and shares
the same output flag set; the management commands (`templates`, `config`)
group their own subcommands.

## Generators

| Command | Produces | Required flags |
|---|---|---|
| [`srekit task`](task.md) | Investigation log (alias: `sretask`) | `--title` |
| [`srekit license`](license.md) | `LICENSE` file (alias: `lic`) | — |
| [`srekit incident`](incident.md) | Live-incident report | `--title` |
| [`srekit postmortem`](postmortem.md) | Postmortem (Google SRE-style) | `--title` |
| [`srekit rfc`](rfc.md) | RFC / ADR | `--title` |
| [`srekit runbook`](runbook.md) | Operational runbook | `--title` |
| [`srekit changelog`](changelog.md) | Keep a Changelog scaffold | — |
| [`srekit oncall-report`](oncall-report.md) | Weekly on-call report | `--team` |
| [`srekit slo`](slo.md) | SLO / SLI document | `--service` |
| [`srekit ebp`](ebp.md) | Error budget policy | `--service` |
| [`srekit capacity`](capacity.md) | Capacity plan | `--service` |
| [`srekit retro`](retro.md) | Sprint retro (Start / Stop / Continue) | `--team` |

## Management

| Command | Purpose |
|---|---|
| [`srekit templates`](templates.md) | Manage a custom templates directory: `init`, `pull`, `list`, `validate`, `diff`, `upgrade` |
| [`srekit config`](config.md) | Scaffold `~/.srekit.yaml`: `init` |
| [`srekit completion`](completion.md) | Shell autocomplete: `bash`, `zsh`, `fish`, `powershell` |

## Shared output flags {#shared-output-flags}

Every generator command accepts:

| Flag | Effect |
|---|---|
| `--out FILE` | write to FILE (refuses to overwrite without `--force`) |
| `--stdout` | print to stdout |
| `--force` | overwrite an existing FILE |
| `--dry-run` | show what would be written, do not write |
| `--template FILE` | use this template file instead of the embedded one |
| `--json` | emit the template data payload as JSON (default sink: stdout) |

The persistent flag `--templates-dir DIR` (or env `SREKIT_TEMPLATES_DIR`,
or `templates_dir:` in `~/.srekit.yaml`) installs a custom templates
directory whose files override the embedded ones. Missing files fall back
transparently.

See [Configuration & precedence](../guides/configuration.md) for the full
resolution rules.
