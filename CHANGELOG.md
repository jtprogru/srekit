# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

This file was scaffolded with `srekit changelog --out CHANGELOG.md` and is maintained by hand. The GoReleaser pipeline still auto-generates per-release notes on the GitHub Releases page from commit messages; this file is the project-level human-curated history.

## [Unreleased]

### Added

-

### Changed

-

### Fixed

- `srekit templates init` without a positional `[dir]` argument now respects the configured templates directory (`--templates-dir` / `SREKIT_TEMPLATES_DIR` / `templates_dir:` in `~/.srekit.yaml`), falling back to `~/.srekit/templates` only when nothing is configured. Previously it ignored the config and unconditionally scaffolded into `~/.srekit/templates`, leading to a "ghost" directory that no other subcommand read from.

## [0.10.0] - 2026-05-27

### Added

- True 3-way merge in `srekit templates upgrade`. The binary now keeps a per-template snapshot of the last-synced embedded version under `<templates-dir>/.srekit-embedded/`, and uses it as the merge base for customized files. `git merge-file --diff3` does the work; clean merges land silently, conflicts are written with `<<<<<<<`/`>>>>>>>` markers and the command exits non-zero so CI flags them. New cases handled: upstream changed but user untouched → fast-forward without `--force`; user changed but upstream untouched → silent no-op (no warning). Without a snapshot (old user dirs), behavior falls back to additive and the snapshot is seeded so the *next* upgrade can merge.
- `srekit templates init` seeds the `.srekit-embedded/` sidecar on first use and best-effort appends `.srekit-embedded/` to the user dir's `.gitignore` so the sidecar stays out of their template repo.

## [0.9.0] - 2026-05-27

### Added

- `srekit templates upgrade [dir]` — additive upgrade of a custom templates directory against the binary's embedded set. Copies in any template that's missing in the user dir, leaves customized files alone (run `templates diff` to inspect, or pass `--force` to overwrite), and always refreshes `TEMPLATES.md`. `--dry-run` previews changes without touching the filesystem. Closes the `init` → `pull` → `diff` → `upgrade` loop; a true three-way merge against git history is left for v1.0.
- `srekit templates list [dir]` — walks the embedded set and the user dir, classifying each `*.tmpl` as `identical` / `customized` / `user-only` / `embedded-only`. Table by default; `--json` emits a sorted array (camelCase keys: `name`, `status`, `userPath`); `--filter STATE` narrows the output. Works without a configured user dir (shows the embedded set as `embedded-only`). Note: the JSON shape here uses camelCase, distinct from the generator `--json` PascalCase contract — a v1.0 cleanup item.

## [0.8.0] - 2026-05-27

### Added

- `srekit config init` — interactive (TTY) / non-interactive (`--yes`, piped stdin) scaffolder for `~/.srekit.yaml`. Defaults pulled from `git config user.name` / `user.email`. Flags: `--author`, `--email`, `--templates-dir`, `--force`, `--yes`. Honors the root `--config FILE` for a custom target path. Writes with `0o600` and refuses to overwrite without `--force`. The `templates_dir:` key is emitted as a commented-out hint when unset, so the embedded-only default keeps working.
- `.github/dependabot.yml` — weekly updates for Go modules and GitHub Actions. Actions grouped into a single PR; `chore(deps)` / `chore(ci)` commit-message prefixes to match the existing log style.
- `--json` flag on every generator command. Short-circuits the template and emits the data payload as indented JSON (default sink: stdout; `--out FILE` writes to a file). Field names are PascalCase — they match what the templates see, so `srekit postmortem --title X --severity SEV-1 --json | jq '.Severity'` works without translation. With `--json` the markdown default path is *not* used, so a JSON payload never lands in a `.md` file by accident.

### Changed

- `golangci-lint` CI pin bumped from `v2.10` to `v2.12` to match the local toolchain. Removes a version skew where the older gosec ruleset flagged `cmd/config.go` for G703/G705 false positives on a CLI generator whose contract is exactly "write user input to a user-named file."
- `Taskfile.yml` — added `test:race`, `release:dry`, and a `ci` umbrella (lint + race tests) for a one-shot pre-push check.

## [0.7.0] - 2026-05-27

### Added

- `srekit templates validate [dir]` — parses each `*.tmpl` with the same FuncMap srekit uses and (for files whose names match a built-in template) executes them against canonical sample data. Catches syntax errors and field-name typos (`{{ .Servce }}` instead of `{{ .Service }}`). Exits non-zero if any template fails. Templates with non-built-in names get parse-only validation.
- `srekit templates diff [dir]` — diffs each `*.tmpl` in the user dir against the version embedded in the current binary via `git diff --no-index`. Useful after `srekit templates pull` or a binary upgrade to see what drifted. Flags: `--name-only`, `--no-color`. Templates without an embedded counterpart are reported as `user-only`.
- `internal/tmpl.Samples` — canonical sample-data registry keyed by built-in template filename, plus `internal/tmpl.Validate(name, body)` helper. Single source of truth for what data shape each built-in template expects; must stay in sync with the struct literals in `cmd/*.go`.

## [0.6.0] - 2026-05-26

### Added

- `srekit templates pull` — sync the configured templates directory with its git remote. `git pull --ff-only` by default (safe: fails on diverged branches); `--rebase` for users who want their local commits rebased on top. Directory is resolved from the same flag / env / yaml chain as `--templates-dir`, falling back to `~/.srekit/templates`. Output streams directly so you see exactly what git did.

### Changed

- `internal/tmpl.TEMPLATES.md` (the embedded reference shipped to user template dirs) now points users at `srekit templates pull` instead of the manual `git -C <dir> pull`.
- `cmd/root.go`: extracted `resolveTemplatesDir(cmd)` helper that both `configureTemplates` and `templates pull` use. Persistent flag storage moved from a captured local to cobra's internal flag value, accessed via `cmd.Flags().Lookup`.

### Fixed

-

## [0.5.0] - 2026-05-26

### Added

- Custom-templates support — point `srekit` at a directory of your own templates and they override the embedded ones, with transparent per-file fallback to embedded when a template is missing in your dir. Three entry points:
  - `srekit --templates-dir <dir> …` (per-invocation flag)
  - `SREKIT_TEMPLATES_DIR` env / `templates_dir:` key in `~/.srekit.yaml`
  - `srekit <cmd> --template <file>` for a one-shot single-template override
- `srekit templates init [dir]` scaffolds a directory with every built-in template + a `TEMPLATES.md` placeholder/FuncMap reference, and runs `git init` (use `--no-git` to skip, `--force` to overwrite). Default target: `~/.srekit/templates`.
- `internal/tmpl.Source` interface (Read by name) with `EmbedSource` and `DirSource` implementations, and a `Loader` that walks Sources in order with `fs.ErrNotExist`-as-fallthrough semantics.
- `internal/tmpl.ParseFile(path)` for the per-command `--template` flag; applies the same FuncMap as embedded templates.
- `internal/tmpl/TEMPLATES.md` reference doc — embedded in the binary and emitted by `templates init`.

### Changed

- `internal/render.Options` gains `TemplatePath` (read template from an arbitrary file path instead of the loader chain).
- `internal/cliflags.Output` gains `TemplatePath`, wired to a new `--template` flag on every command.
- `cmd/root.go` adds persistent `--templates-dir`, expands `~`, warns to stderr on missing path / non-directory (and silently falls back to embedded). `configureTemplates` only mutates the package-level `tmpl.Default` when a dir is actually provided, keeping parallel tests race-free.

### Fixed

-

## [0.4.0] - 2026-05-26

### Added

- `srekit incident` — live-incident report template (status, lead, comms, updates log). Filled during the incident, distinct from `postmortem` which is written after. `--status` is validated against `investigated | active | contained | resolved`.
- `srekit ebp` — Error Budget Policy template with tiered actions (Yellow / Orange / Red), exceptions, and escalation.
- `srekit capacity` — capacity planning template: baseline metrics, growth assumptions, forecast, scale-up triggers, headroom target, dependencies, cost, risks.
- `internal/tmpl.Funcs` — `text/template` FuncMap shared by all parsed templates: `default`, `shortID`, `slugify`, `now`, `upper`, `lower`, `trim`. `now` honors `clock.Now` for test-time determinism.
- Russian translation of all SRE-document templates with bilingual `Русский (English)` headings; body text fully Russian. Frontmatter keys, technical acronyms (SLO/SLI/RFC/ADR/PromQL/UTC/SEV), version anchors, and PromQL stay English. `changelog.md.tmpl` intentionally kept fully English to preserve Keep a Changelog tooling compatibility.
- `## Ссылки (References)` section added to `oncall` and `postmortem` templates.
- Frontmatter `title` field in `postmortem`, `rfc`, `runbook` (was previously only in `task`).
- Frontmatter `modification_date` in `runbook` and `slo` (living documents).
- RFC frontmatter ADR fields: `decision_date`, `deciders`, `supersedes`, `superseded_by`.
- `internal/tmpl/funcs_test.go` — first test coverage for the template package; covers each FuncMap entry plus a parse sanity check on `rfc.md.tmpl`.

### Changed

- `srekit task` rewritten as an SRE investigation log: sections Context / Hypothesis / Evidence / Findings / Action items / References. Frontmatter `type` is now `investigation`. Default output filename changed from `Tasker - <title>.md` to `investigation-<slug>.md`. The `sretask` alias still resolves.
- `cmd/task.go` now uses `time.RFC3339` for `creation_date`/`modification_date`, matching every other command.
- `slo.md.tmpl`: split the SLI block into separate Availability and Latency definitions, each with its own PromQL example. Service / Owner team / Tier / User journey labels moved out of `## Service overview` into a metadata list directly under the H1, matching the layout of every other template.
- `retro.md.tmpl`: removed the duplicate `What went well / didn't go well / confused` block. Format is now just Start / Stop / Continue plus `What confused us`.
- `postmortem.md.tmpl`: Duration placeholder normalized to italic style.
- `rfc.md.tmpl`: short-ID rendering moved out of the cmd-side struct and into the template via `{{ shortID .ID 8 }}`.
- `cmd/postmortem.go`, `cmd/runbook.go`, `cmd/incident.go`: stopped pre-filling field defaults via `valueOr()`. Placeholder text now lives in the templates as `{{ .Field | default "<…>" }}`, colocated with the document that displays it.
- `task.md.tmpl` tag list: `sre` → `debug` (every artifact in this tool is "SRE", so the tag carried no information).
- Root `--help` Short/Long updated to list `investigation`, `incident`, `ebp`, `capacity`.

### Removed

- `cmd/util.go` (with `valueOr`) — superseded by the template `default` func.

### Fixed

- `slo.md.tmpl`: the example Availability PromQL counted HTTP 401/403 responses as good events, contradicting the SLI definition that explicitly excluded them. Auth failures are now excluded from both numerator and denominator.

## [0.3.0] - 2026-05-26

### Changed

- `oncall-report`: weekly window when neither `--start` nor `--end` is given now correctly resolves to the current week's Mon–Sun. Sunday previously rolled over to next Monday.
- `changelog`: when `--repo` is omitted and the git remote cannot be detected, the command now exits with an error instead of writing the literal `OWNER/REPO` placeholder into the generated file.
- Generated artifacts (CHANGELOG, RFC, runbook, LICENSE, postmortem, …) are now written with file mode `0o644`. Previously `0o600`; existing files are not affected.
- `cmd/` rewritten to a constructor pattern with no package-level mutable flag state. Tests are parallel-safe and race-clean.
- `meta.Resolve` no longer reads the global `viper` instance; it accepts a `Lookup` interface (`GetString(key) string`).
- `Taskfile.yml` uses `go run .` / `go build .` instead of pinning to `main.go`.

### Added

- `internal/cliflags` helper that binds the shared `--out`/`--stdout`/`--force`/`--dry-run` quartet on a cobra command and converts to `render.Options`.
- `internal/clock` package with an overridable `Now` so tests can pin the wall clock.
- Regression tests: oncall Sunday week boundary, changelog repo-detection error, render file-mode assertion.

### Fixed

- Removed an unreachable `showVersion` check in `Execute()`. `--version`/`-V` now drives through cobra's `SetVersionTemplate`, printing the full commit/date/builder block as intended.
- `licAlias`/`sretaskAlias` collapsed into `cobra.Command.Aliases`, eliminating a fragile double-`StringVar` binding to shared package-level vars.

### Dependencies

- `github.com/spf13/cobra` 1.6.1 → 1.10.2
- `github.com/spf13/viper` 1.15.0 → 1.21.0

## [0.2.0] - 2026-05-06

### Added

- `srekit oncall-report` — weekly on-call report template.
- `srekit slo` — SLO/SLI document template.
- `srekit retro` — sprint retrospective template.
- `srekit changelog` auto-detects the GitHub repo from `git remote origin`.
- `.golangci.yaml` lint configuration.

## [0.1.0] - 2026-05-06

### Added

- Initial public release.
- Commands: `task` (alias `sretask`), `license` (alias `lic`), `postmortem`, `rfc`, `runbook`, `changelog`, `completion`.
- Embedded templates for tasks, postmortems, RFCs, runbooks, and licenses (WTFPL, MIT, Apache 2.0).
- Author/email resolution chain: `--author`/`--email` → `SREKIT_AUTHOR`/`SREKIT_EMAIL` → `~/.srekit.yaml` → `git config user.name`/`user.email`.
- Shared output flags across every command: `--out`, `--stdout`, `--force`, `--dry-run`.
- GoReleaser pipeline producing Linux/macOS/FreeBSD × amd64/arm64 builds, GPG-signed checksums, and a Homebrew cask in `jtprogru/homebrew-tap`.

[Unreleased]: https://github.com/jtprogru/srekit/compare/v0.7.0...HEAD
[0.7.0]: https://github.com/jtprogru/srekit/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/jtprogru/srekit/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/jtprogru/srekit/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/jtprogru/srekit/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/jtprogru/srekit/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/jtprogru/srekit/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/jtprogru/srekit/releases/tag/v0.1.0
