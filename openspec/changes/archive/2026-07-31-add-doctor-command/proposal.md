## Why

Everything `srekit` reads before it renders a single byte is invisible until it misbehaves: which config file actually won (legacy `~/.srekit.yaml` or the XDG path), whether the templates directory resolved at all, whether a user template still parses, whether `git` is on `PATH` to supply author and repo slug. Today a user diagnoses this by running a generator and reading between the lines of its output — or worse, by not noticing at all, as when an XDG config sits unread because a legacy file shadows it. `doctor` makes that resolved state inspectable in one command, on the on-call laptop and in CI alike.

## What Changes

- New top-level command `srekit doctor` that inspects the environment and reports one line per check with an `ok` / `warn` / `error` status.
- Checks cover three areas: **configuration and paths** (which config file is read, whether a second one is being shadowed, where the templates directory resolves, whether those locations are readable/writable, which `SREKIT_*` variables are in effect, whether an author identity can be resolved at all), **template health** (do the user's artifacts parse, are legacy `.tmpl` / `.sections.yaml` files present, has the directory drifted from the embedded set such that `templates upgrade` is warranted), and **external dependencies** (is `git` present on `PATH` and at what version).
- `--json` emits the same findings as a machine-readable document with `camelCase` keys, for CI gating.
- Exit status is `0` when no check fails, `1` when at least one check reports `error`. A `warn` does not fail the run.
- `doctor` is read-only: it never writes, creates, or repairs anything.
- Not a breaking change. No existing command, flag, JSON payload, template, or config location changes behaviour. `doctor` is purely additive; the name is currently unused.
- No new dependency. Every check is implemented with the standard library plus the machinery already in the binary (`internal/config`, `internal/tmpl`, `internal/sections`, `internal/meta`). Binary size and the build graph are unaffected — in particular nothing here pulls `net/http` or `crypto`, since the `git` check runs the local binary via `os/exec` rather than talking to a network.

## Capabilities

### New Capabilities
- `environment-diagnostics`: what `srekit doctor` inspects, how each finding is classified and reported in both text and JSON, and what the command's exit status means.

### Modified Capabilities

None. `doctor` is not a generator, so the `artifact-generation` catalog and the `output-routing` flag bundle are untouched; it neither changes how configuration is resolved nor how templates are loaded, only reports on the outcome. It reads the same resolution chains that `user-configuration`, `template-overrides` and `identity-and-metadata` already specify, so those specs stay as they are.

## Impact

- New `cmd/doctor.go`, registered in `NewRootCmd()` in `cmd/root.go`.
- Reuses existing internals read-only: path resolution in `cmd/paths.go`, the templates-directory chain in `cmd/root.go`, artifact parsing in `internal/sections`, embedded-set comparison of the kind `cmd/templates.go` already performs for `list`, and author resolution in `internal/meta`.
- No change to `internal/render` or `internal/cliflags`; `doctor` renders no artifact and therefore binds none of the `--out` / `--stdout` / `--force` / `--dry-run` output flags, per the rule that a flag a command would ignore must not exist.
- Docs: new command page in `docs/en/commands/` and `docs/ru/commands/`, plus `mkdocs.yml` nav in both locales.
- `CHANGELOG.md` entry under `[Unreleased]`.
- Smoke tests in `cmd/cmd_test.go`; the checks that depend on the environment are exercised against temp directories and an isolated config rather than the developer's real `$HOME`.
