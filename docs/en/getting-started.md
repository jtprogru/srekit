# Getting started

This page walks you from zero to a generated postmortem in three minutes.

## 1. Install

=== "Homebrew (macOS / Linux)"

    ```bash
    brew tap jtprogru/tap
    brew install srekit
    ```

=== "go install"

    ```bash
    go install github.com/jtprogru/srekit@latest
    ```

    Requires Go 1.26+ (the module declares `go 1.26.4`).

=== "Pre-built binaries"

    Download from the [GitHub Releases page](https://github.com/jtprogru/srekit/releases)
    for `darwin`, `linux`, or `freebsd` × `arm64` / `x86_64`.

Confirm the install:

```bash
srekit --version
# srekit version: 0.32.1
# from commit: ...
# built date: ...
# built by: goreleaser
```

## 2. One-time configuration

Generators that record an author (`rfc`, `oncall-report`) resolve the identity in this order:

1. `--author` / `--email` flags
2. `SREKIT_AUTHOR` / `SREKIT_EMAIL` env vars
3. `author:` / `email:` in the config file
4. `git config user.name` / `git config user.email`

The config file lives at `$XDG_CONFIG_HOME/srekit/config.yaml` (`~/.config/srekit/config.yaml` when the variable is unset). A pre-XDG `~/.srekit.yaml` keeps working and wins if it already exists, so nothing to migrate.

If your global `git config` is already set, **you can skip configuration entirely**. To opt in to a yaml file:

```bash
srekit config init
# Author name [Mikhail Savin]: ⏎
# Email [jtprogru@gmail.com]: ⏎
# Templates dir (leave empty to use embedded templates only): ⏎
# Wrote /Users/jtprogru/.config/srekit/config.yaml
```

`srekit config init --yes` runs without prompts, using flag values and `git config` defaults. See [Configuration & precedence](guides/configuration.md) for the full picture.

## 3. Your first artifact

A postmortem is a good starter — it exercises author resolution, title-based filenames, and the `--out` / `--stdout` flags:

```bash
srekit postmortem --title "API outage" --severity SEV-1 --stdout
```

That prints to stdout. To write to a file:

```bash
srekit postmortem --title "API outage" --severity SEV-1 \
  --start 2026-05-06T08:00Z --end 2026-05-06T09:30Z \
  --owner "@oncall" \
  --out postmortem-2026-05-06.md
```

Inspect it — every section is pre-filled with bilingual headings (`Постмортем (Postmortem)`) and SRE-canonical fields (severity, timeline, impact, root cause, action items).

## 4. The unified flag set

Every generator command supports the same output flags:

| Flag | Effect |
|---|---|
| `--out FILE` | write to FILE (refuses to overwrite without `--force`) |
| `--stdout` | print to stdout |
| `--force` | overwrite existing FILE |
| `--dry-run` | show what would be written, do not write |
| `--json` | emit the template data payload as JSON instead of rendering |

There is no `--template FILE` flag: it was removed in v0.30.0 along with `srekit license`, the one command whose render path read it. Per-artifact customization is done by dropping a `<name>.yaml` into your `templates_dir` — see [Custom templates workflow](guides/custom-templates.md).

If you pass neither `--out` nor `--stdout`, each command has a sensible default filename (e.g. `investigation-<slug>.md` for `srekit task`, `oncall-<team>-<start>.md` for the on-call report).

## 5. Check the environment

If a generator picks the wrong author, or your customized templates look ignored, ask srekit what it resolved rather than guessing:

```bash
srekit doctor
```

It is read-only and makes no network request. `srekit doctor --quiet` prints only what needs attention, so silence means healthy — see [`srekit doctor`](commands/doctor.md).

## 6. What's next

You now have enough to use srekit day-to-day. To go deeper:

- **[Custom templates workflow](guides/custom-templates.md)** — fork the built-in artifacts into your own git repo and pull/merge upstream changes cleanly.
- **[jtprogru/sre-templates](https://github.com/jtprogru/sre-templates)** — a ready-made templates repo in the exact layout srekit expects. Clone it and point `templates_dir` at it to skip scaffolding from scratch.
- **[JSON output](guides/json-output.md)** — pipe generators into `jq` for CI scripts and integrations.
- **[Commands overview](commands/index.md)** — full reference for every subcommand and flag.
- **[Recipes](../recipes.md)** — concrete workflows that combine srekit with your tooling.
