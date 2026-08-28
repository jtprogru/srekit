# Commands overview

srekit's surface is a flat tree of cobra subcommands. Every generator command produces a single Markdown artifact and shares the same output flag set; the management commands (`templates`, `config`) group their own subcommands, and `doctor` reports on the environment all of them resolve.

## Generators

| Command | Produces | Required flags |
|---|---|---|
| [`srekit task`](task.md) | Investigation log (alias: `sretask`) | `--title` |
| [`srekit tasker`](tasker.md) | Task card for a collection of engineering tasks | `--title` |
| [`srekit postmortem`](postmortem.md) | Postmortem (Google SRE-style) | `--title` |
| [`srekit rfc`](rfc.md) | RFC / ADR | `--title` |
| [`srekit runbook`](runbook.md) | Operational runbook | `--title` |
| [`srekit changelog`](changelog.md) | Keep a Changelog scaffold | — |
| [`srekit oncall-report`](oncall-report.md) | Weekly on-call report | `--team` |
| [`srekit slo`](slo.md) | SLO / SLI document | `--service` |
| [`srekit ebp`](ebp.md) | Error budget policy | `--service` |

## Management

| Command | Purpose |
|---|---|
| [`srekit templates`](templates.md) | Manage a custom templates directory: `init`, `pull`, `list`, `validate`, `diff`, `upgrade`, `migrate` |
| [`srekit config`](config.md) | Scaffold the config file (`$XDG_CONFIG_HOME/srekit/config.yaml`): `init` |
| [`srekit doctor`](doctor.md) | Diagnose the environment read-only: config, templates, identity, `git` |
| [`srekit completion`](completion.md) | Shell autocomplete: `bash`, `zsh`, `fish`, `powershell` |

## Shared output flags {#shared-output-flags}

Every generator command accepts:

| Flag | Effect |
|---|---|
| `--out FILE` | write to FILE (refuses to overwrite without `--force`) |
| `--stdout` | print to stdout |
| `--force` | overwrite an existing FILE |
| `--dry-run` | show what would be written, do not write |
| `--json` | emit the template data payload as JSON (default sink: stdout) |

A command that edits a document you already own is not a generator and carries a narrower bundle: `--dry-run`, `--stdout` and `--json` with their usual meanings, but neither `--out` nor `--force`. Its destination is the file it was pointed at, so a second destination has no meaning, and an overwrite guard would guard against the command's own purpose. [`srekit changelog release`](changelog.md#cutting-a-release) is the one such command today.

There is no `--template FILE` flag. It was removed in v0.30.0 with `srekit license`, the one command whose render path read it; per-artifact customization is via dropping a `<name>.yaml` into your `templates_dir` — see [Custom templates workflow](../guides/custom-templates.md).

`capacity`, `retro` and `license` were removed in v0.30.0 — see [Removed commands](../migration/removed-commands.md).

The persistent flag `--templates-dir DIR` (or env `SREKIT_TEMPLATES_DIR`, or `templates_dir:` in the config file) installs a custom templates directory whose files override the built-in ones. Missing files fall back transparently.

`srekit changelog` also carries a persistent `--lang en|ru` that its `release` and `validate` subcommands inherit. It selects the vocabulary that gets *generated*; what an existing document is parsed as is read out of that document — see [`srekit changelog`](changelog.md#the-russian-variant).

See [Configuration & precedence](../guides/configuration.md) for the full resolution rules.
