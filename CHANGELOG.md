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

-

## [0.18.0] - 2026-06-04

### Changed

- **`incident` and `rfc` migrated to the v1 artifact format.** Same dogfooded recipe as ebp + capacity in v0.17.0: `srekit templates migrate` on the legacy `.tmpl`, then `sed`-rewrote `{{ .Field }}` → `{{ .Meta.Field }}` for the new data root. For `rfc`, the `{{ shortID .ID 8 }}` reference in the H1 also needed a follow-up `sed` to become `{{ shortID .Meta.ID 8 }}`. Both commands now use the artifact render path with structured `--json` (per-section access).

- **Breaking — `srekit incident --json` and `srekit rfc --json` shape.** Moves from bootstrap envelope (`sections: [{id:"body", body:<markdown>}]`) to structured (`sections: [{id:"current_impact",...}, {id:"context",...}, ...]`). Migration: replace `jq '.sections[0].body'` with `jq '.sections[] | select(.id=="...").body'`. Same pattern as ebp / capacity in v0.17.0; documented in `docs/{en,ru}/migration/v1.md` (v0.18.0 section).

- **Breaking — `incident.md.tmpl` / `rfc.md.tmpl` no longer ship in embed.** Users with customized copies get stderr WARN. `srekit templates upgrade` scaffolds the new `.yaml`; `srekit templates migrate` auto-converts the legacy `.tmpl` (review the result and rewrite `{{ .Field }}` references to `{{ .Meta.Field }}`; for `rfc`-style templates also rewrite `{{ shortID .ID … }}` → `{{ shortID .Meta.ID … }}`).

### Fixed

-

## [0.17.0] - 2026-06-04

### Changed

- **`ebp` and `capacity` migrated to the v1 artifact format.** Same dogfooded recipe as retro + slo in v0.16.0: `srekit templates migrate` on the legacy `.tmpl`, then `sed`-rewrite `{{ .Field }}` → `{{ .Meta.Field }}` for the new data root. Both commands now use the artifact render path with structured `--json` (per-section access).

- **Breaking — `srekit ebp --json` and `srekit capacity --json` shape.** Moves from bootstrap envelope (`sections: [{id:"body", body:<markdown>}]`) to structured (`sections: [{id:"triggers",...}, {id:"tiered_actions",...}, ...]`). Migration: replace `jq '.sections[0].body'` with `jq '.sections[] | select(.id=="...").body'`. Same pattern as retro / slo in v0.16.0; documented in `docs/{en,ru}/migration/v1.md` (v0.17.0 section).

- **Breaking — `ebp.md.tmpl` / `capacity.md.tmpl` no longer ship in embed.** Users with customized copies get stderr WARN. `srekit templates upgrade` scaffolds the new `.yaml`; `srekit templates migrate` auto-converts the legacy `.tmpl` (review the result and rewrite `{{ .Field }}` references to `{{ .Meta.Field }}`).

### Fixed

-

## [0.16.0] - 2026-06-04

### Changed

- **`retro` and `slo` migrated to the v1 artifact format.** Both followed the same dogfooded recipe as task in v0.15.0: ran `srekit templates migrate` on the legacy `.tmpl`, then `sed`-rewrote `{{ .Field }}` → `{{ .Meta.Field }}` for the new data root. Both commands now use the artifact render path with structured `--json` (per-section access).

- **Breaking — `srekit retro --json` and `srekit slo --json` shape.** Moves from bootstrap envelope (`sections: [{id:"body", body:<markdown>}]`) to structured (`sections: [{id:"goals_of_this_retro",...}, {id:"action_items",...}, ...]`). Migration: replace `jq '.sections[0].body'` with `jq '.sections[] | select(.id=="...").body'`. Same migration story as task in v0.15.0; documented in `docs/{en,ru}/migration/v1.md` (v0.16.0 section).

- **Breaking — `retro.md.tmpl` / `slo.md.tmpl` no longer ship in embed.** Users with customized copies get stderr WARN. `srekit templates upgrade` scaffolds the new `.yaml`; `srekit templates migrate` auto-converts the legacy `.tmpl` (review the result and rewrite `{{ .Field }}` references to `{{ .Meta.Field }}`).

### Fixed

-

## [0.15.0] - 2026-06-04

### Added

- **`srekit templates migrate [dir]`** — best-effort converter that turns legacy `.tmpl` files (plus their optional `.sections.yaml` sidecars) into the v1 single-file `<name>.yaml` artifact format. Parses frontmatter / H1 / meta_bullets / `## ` sections from the `.tmpl`; infers section type (GFM tables → `type: table`; everything else → `type: text` with `default_body` verbatim); slugifies bilingual headings to English-only IDs (e.g. `Контекст (Context)` → `context`); wraps sections containing Go-template control flow in `git merge`-style diff markers for human resolution. Defaults to `--dry-run`; pass `--apply` to write files. Original `.tmpl` / `.sections.yaml` are left in place for review.

- **`internal/migrate` package** — public API: `migrate.Convert(tmplBody, sectionsManifest)` returns artifact YAML bytes. Reusable for tooling that wants programmatic migration.

- **`task` migrated to the v1 artifact format.** First **fresh** migration (postmortem in v0.14.0 was a refactor of the existing prototype). `task.yaml` ships in embed; `task.md.tmpl` is gone. The task command now uses the artifact render path with structured `--json` (`{meta, sections: [...]}`) instead of the bootstrap envelope.

### Changed

- **Breaking — `srekit task --json` shape.** Moves from bootstrap envelope (`{meta, sections: [{id:"body", body:<rendered markdown>}]}`) to structured (`{meta, sections: [{id:"context",...}, {id:"hypothesis",...}, ...]}`). Migration: replace `jq '.sections[0].body'` with per-section access like `jq '.sections[] | select(.id=="context").body'`. Documented in `docs/{en,ru}/migration/v1.md` (v0.15.0 section).

- **Breaking — `task.md.tmpl` no longer ships in embed.** Users with customized copies in `templates_dir` will see a stderr WARN on every task invocation. To migrate: run `srekit templates upgrade` to scaffold `task.yaml`, then port customizations into it.

- **`templates init` scaffold** now writes `task.yaml` instead of `task.md.tmpl`. Documented in CHANGELOG and migration guide.

### Fixed

-

## [0.14.0] - 2026-06-04

### Added

- **v1 single-file artifact format (`<name>.yaml`).** First step of the YAML-first migration toward v1.0. An artifact YAML declares `version`, optional `frontmatter` (free-form map; values run through Go templates), optional `title` (H1), optional `meta_bullets` (the bulleted list after H1), optional `header_body` (freeform Markdown escape hatch for things like blameless callouts), and a `sections` list (same schema as v0.13.x `.sections.yaml`). Frontmatter key order is preserved verbatim from source via `yaml.Node`. New: `internal/sections.Artifact`, `ParseArtifact`, `RenderArtifact`; `tmpl.LoadArtifactBytes`, `ArtifactNameFor`.

- **`postmortem.yaml`** ships in embed, replacing the v0.13.x split (`postmortem.md.tmpl` + `postmortem.sections.yaml`). All section content from v0.13.x is preserved 1:1; header content (frontmatter, H1, meta bullets) is faithfully ported.

- **Migration guide stub** at `docs/{en,ru}/migration/v1.md`. Tracks release-by-release progress; will grow as more generators migrate.

- **Snapshot GC in `templates upgrade`.** End-of-run pass removes orphaned snapshot files in `.srekit-embedded/` for artifacts no longer in embed (e.g. `postmortem.md.tmpl` / `postmortem.sections.yaml` after v0.14.0). Summary line now reports the count.

- **`templates validate`** learns the v1 `.yaml` artifact format. Bad artifacts surface their parse / schema errors on a `FAIL` line, same UX as the existing `.tmpl` / `.sections.yaml` paths.

### Changed

- **Breaking — postmortem file layout.** `postmortem.md.tmpl` and `postmortem.sections.yaml` are removed from embed. Users with customized copies in `templates_dir` will see a stderr `WARN: <file> is a pre-v0.14.0 format and is being ignored. Run 'srekit templates upgrade' ...` on every postmortem invocation. To migrate: run `srekit templates upgrade`, then move any customizations (blameless callouts, custom meta bullets, …) into the new `postmortem.yaml` (`frontmatter`, `title`, `meta_bullets`, `header_body` fields).

- **Breaking — `license_*.tmpl` no longer ship in embed.** MIT / Apache-2.0 / WTFPL bodies are inlined as Go-string constants in `cmd/license.go`. `templates init` no longer scaffolds them; they don't appear in `templates list` / `validate`. `--template FILE` still works for custom license bodies, which is the only supported customization point now.

- **Cosmetic — frontmatter values now emit quoted strings.** `id: "<uuid>"` (was `id: <uuid>` in v0.13.x). `yaml.v3` preserves quoting style from source YAML, and the new `postmortem.yaml` source has quoted scalars with template syntax — quoting survives substitution. Functionally identical YAML, but a one-character-per-line diff vs v0.13.x output. Documented in the migration guide.

## [0.13.1] - 2026-06-04

### Added

- **`srekit postmortem --schema`** emits a JSON Schema (draft 2020-12) describing the `--from` input payload. Schema is recomputed from the loaded manifest on each call, so user customizations to `postmortem.sections.yaml` flow through automatically. Useful for editor tooling (point `$schema` at the file for autocomplete + validation) and for agents that want a machine-readable contract for what to fill.

- **`srekit postmortem --validate FILE`** runs two checks against an input JSON: unknown section IDs are rejected (same typo guard as `--from`), and every `required: true` section must have a non-empty body (whitespace-only counts as empty). Prints a per-section `OK` / `FAIL` report and exits non-zero on any failure — drop it into CI to gate "postmortem draft" PRs.

  ```bash
  $ srekit postmortem --validate pm.json
  OK    summary
  FAIL  impact: required body is empty
  OK    timeline
  Error: 1 of 5 required section(s) failed validation
  ```

- **`sections.CheckUnknownIDs`** — the typo-guard extracted from `Merge` into a standalone function so callers (like `--validate`) can run it without triggering default-body template evaluation.

## [0.13.0] - 2026-06-04

### Added

- **Structured `--json` via sidecar sections manifest.** `srekit postmortem` is the first generator backed by a `<template>.sections.yaml` file that declares typed slots (`text` / `list` / `table`), required-flags, and default content. `--json` exposes the document section-by-section in manifest order:

  ```jsonc
  { "meta": {…},
    "sections": [
      { "id": "summary", "type": "text", "required": true, "body": "…" },
      { "id": "impact",  "type": "list", "required": true, "body": "- …" },
      …
    ] }
  ```

- **`srekit postmortem --from FILE`** (and `-` for stdin) — round-trip: read structured input (`{meta?, sections}`), merge over manifest defaults, render Markdown. Unknown section IDs are a hard error with the offending IDs and the known set listed (typo guard). Section bodies pass through verbatim — no template evaluation — so arbitrary markdown round-trips safely.

- **`internal/sections` package** — Manifest / Section / RenderedSection types, YAML parsing, default-body rendering, merge. Reusable when other generators migrate.

### Changed

- **Breaking — `--json` shape across every generator** (`postmortem`, `incident`, `rfc`, `runbook`, `retro`, `slo`, `oncall-report`, `task`, `ebp`, `capacity`, `license`, `changelog`). Output is now `{meta, sections}` instead of the previous flat structure. `postmortem` ships the full structured shape; every other command wraps the rendered markdown in a single synthetic section (`{id: "body", type: "text", title: <H1>, required: true, body: <markdown>}`) so a single `jq` recipe works across all generators.

  This is the second `--json` shape change in 0.x (after camelCase in 0.12.0). The shape stabilizes at 1.0; treat `--json` as pre-stable until then. Migration:

  ```bash
  # before (0.12.x)
  srekit task -T "X" --json | jq '.title'
  # after (0.13.0)
  srekit task -T "X" --json | jq '.meta.title'

  # before — grep'ing rendered markdown
  srekit runbook --service api --json | jq '.body'  # never existed
  # after — first (and only) bootstrap section's body
  srekit runbook --service api --json | jq -r '.sections[0].body'

  # postmortem — multi-section, by ID
  srekit postmortem -T X --json | jq -r '.sections[] | select(.id == "summary").body'
  ```

- **Breaking — `postmortem.md.tmpl` structure**. The template shrinks from a hardcoded body to a header (frontmatter + H1 + meta-bullets) followed by `{{ range .Sections }}…{{ end }}`. Section content now lives in the new `postmortem.sections.yaml`. If you customized `postmortem.md.tmpl` in your `templates_dir`, `srekit templates upgrade` will almost certainly report a 3-way merge conflict — that's expected. Migration:

  - Move section-level customizations (titles, default content, item lists, table columns) into `postmortem.sections.yaml`.
  - Adapt the header portion of the new `postmortem.md.tmpl` (frontmatter, H1, meta bullets) to taste.
  - Anything that depended on flat `{{ .Title }}` / `{{ .ID }}` / etc. in the template body becomes `{{ .Meta.Title }}` / `{{ .Meta.ID }}` / etc. — the data root is now `{Meta, Sections}`.

- **`templates init/upgrade/diff/list/validate`** now treat `.sections.yaml` as a first-class artifact alongside `.tmpl`. Enumeration is centralized in `tmpl.EmbeddedNames()` / `tmpl.IsTemplateArtifact()` — adding a new artifact type in the future is one helper update, not five duplicated changes. `templates validate` parses `.sections.yaml` as a manifest and surfaces schema errors (unknown type, missing ID, etc.) on a `FAIL` line.

### Notes for v1.0 direction

This release introduces the manifest format on a single command as a prototype. After 2-3 months of practice, the project will decide whether to migrate the remaining generators to YAML-first templates as part of v1.0 — see `plans/srekit-plan.md` for the roadmap. Until then, the bootstrap wrapper gives those commands a uniform JSON contract without forcing their migration.

## [0.12.1] - 2026-06-04

### Changed

- `postmortem` default output filename now includes today's date: `postmortem-<YYYY-MM-DD>-<slug>.md` (was `postmortem-<slug>.md`). Postmortems land in chronological order when listed, and re-running the command for a similar incident on a different day no longer overwrites the previous file. Explicit `--out` and `--stdout` are unchanged.

## [0.12.0] - 2026-05-28

### Changed

- **Breaking (`--json`):** unified JSON output to **camelCase** keys across every command. The generator commands previously emitted PascalCase (`{"Title":…, "Severity":…}`); they now match the camelCase convention `templates list --json` already used (`{"title":…, "severity":…}`). The nested author object is now `{"name","email"}`. Scripts that pluck PascalCase keys via `jq` (`.Title`, `.Severity`) must switch to camelCase (`.title`, `.severity`). Template authoring is unaffected — Go templates still reference field names (`.Title`).

## [0.11.0] - 2026-05-28

### Added

- Global `-q` / `--quiet` flag. Suppresses informational messages — the `wrote <file>` confirmation, `--dry-run` notes, and the `config init` "Wrote …" line. Generated content and error messages still print, so `srekit task --title X -q --stdout` emits only the artifact.
- `Example` blocks in `--help` for the generator commands (`incident`, `postmortem`, `task`, `runbook`, `rfc`, `slo`, `oncall-report`, `license`) and `templates init` / `upgrade` / `diff`. The root `--help` now links to the documentation site and the GitHub repository.
- `oncall` alias for `oncall-report`.
- XDG Base Directory support. Fresh installs resolve the config file at `$XDG_CONFIG_HOME/srekit/config.yaml` and the default templates directory at `$XDG_CONFIG_HOME/srekit/templates`. The pre-XDG locations (`~/.srekit.yaml`, `~/.srekit/templates`) keep working unchanged: an existing legacy path wins, so current users see no migration and never end up with a silently-ignored config.

### Changed

- Validation errors now print only the error message instead of dumping the full usage block (e.g. `srekit incident` without `--title`). Usage is still shown for genuine flag/argument parse errors (unknown flag, bad value).
- `--out -` now writes to stdout, following the conventional `-` stand-in, instead of creating a file literally named `-`.
- `templates diff` honors the `NO_COLOR` environment variable in addition to the existing `--no-color` flag.
- `SIGINT` / `SIGTERM` now cancel the command context, so child `git` processes spawned by the `templates` subcommands are torn down on Ctrl-C. A second Ctrl-C falls through to immediate termination.
- Documented uninstallation in the README, and removed a dead `windows` archive `format_override` from the GoReleaser config (no Windows build target was ever produced).

## [0.10.2] - 2026-05-28

### Added

- Full MkDocs Material documentation site at [`jtprogru.github.io/srekit`](https://jtprogru.github.io/srekit/) — bilingual (EN + RU via `mkdocs-static-i18n`), with a getting-started, full command reference for every subcommand, guides for custom templates / JSON output / configuration precedence, recipes, architecture overview, and contributing notes. Built and deployed automatically by `.github/workflows/docs.yaml` on every push touching `docs/` or `mkdocs.yml`. Local preview: `task docs:install && task docs:serve`.

### Changed

- Template loading no longer goes through a package-level mutable `tmpl.Default`. The loader is now built per command tree in `configureTemplates` and threaded to each command via the cobra command context (`loaderFrom`); `render.Render` takes the `*tmpl.Loader` as an argument. This removes the global that made `--templates-dir` tests non-parallel and tripped the race detector locally; those tests are `t.Parallel()` again and the `resetTmplDefault` workaround is gone. No user-facing behavior change.

## [0.10.1] - 2026-05-27

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

[Unreleased]: https://github.com/jtprogru/srekit/compare/v0.18.0...HEAD
[0.18.0]: https://github.com/jtprogru/srekit/compare/v0.17.0...v0.18.0
[0.17.0]: https://github.com/jtprogru/srekit/compare/v0.16.0...v0.17.0
[0.16.0]: https://github.com/jtprogru/srekit/compare/v0.15.0...v0.16.0
[0.15.0]: https://github.com/jtprogru/srekit/compare/v0.14.0...v0.15.0
[0.14.0]: https://github.com/jtprogru/srekit/compare/v0.13.1...v0.14.0
[0.13.1]: https://github.com/jtprogru/srekit/compare/v0.13.0...v0.13.1
[0.13.0]: https://github.com/jtprogru/srekit/compare/v0.12.1...v0.13.0
[0.12.1]: https://github.com/jtprogru/srekit/compare/v0.12.0...v0.12.1
[0.12.0]: https://github.com/jtprogru/srekit/compare/v0.11.0...v0.12.0
[0.11.0]: https://github.com/jtprogru/srekit/compare/v0.10.2...v0.11.0
[0.7.0]: https://github.com/jtprogru/srekit/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/jtprogru/srekit/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/jtprogru/srekit/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/jtprogru/srekit/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/jtprogru/srekit/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/jtprogru/srekit/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/jtprogru/srekit/releases/tag/v0.1.0
