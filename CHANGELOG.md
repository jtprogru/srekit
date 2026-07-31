# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

This file was scaffolded with `srekit changelog --out CHANGELOG.md` and is maintained by hand. The GoReleaser pipeline still auto-generates per-release notes on the GitHub Releases page from commit messages; this file is the project-level human-curated history.

## [Unreleased]

## [0.30.0] - 2026-07-31

### Added

- **New command `srekit doctor`** — a read-only diagnostic of the environment. It surfaces what until now was only visible indirectly: which config file is actually read and whether a second one shadows it (the classic trap in the "XDG for fresh installs, legacy wins if present" rule), where the templates directory resolves and which source supplied it, whether the user's artifacts still parse, whether pre-v1.0 template files linger and how far the directory has drifted from the embedded set, whether an author identity resolves at all and from where, which `SREKIT_*` variables are in effect, and whether `git` is on `PATH`. Twelve checks across three categories (`config`, `templates`, `dependencies`), each with a stable identifier (`config.shadowed`, `templates.parse`, `dependencies.git`, …) — a public contract teams can gate CI on. An `error` finding exits `1`; a `warn` does not, so `doctor` can be adopted in CI without being blocked by advisory findings. `--quiet` prints only what needs attention, so silence means healthy. `--json` emits the same findings as a document with `camelCase` keys and the worst status as its overall status. The command writes nothing, repairs nothing, and makes no network request; it has no `--out` / `--stdout` / `--force` / `--dry-run`, since there is nothing to write. No new dependency. Docs: [docs/commands/doctor.md](https://jtprogru.github.io/srekit/commands/doctor/).
- `CLAUDE.md` at the repo root: build/test/lint commands, the render pipeline from flags to markdown, project invariants (dependency minimalism, no package-level mutable state, snake_case YAML vs camelCase JSON, XDG-with-legacy-precedence), and the checklist for adding a generator.

### Changed

- **Generators now identify an artifact by its bare name.** `render.Render` / `Loader.LoadArtifactBytes` are called with `"slo"` instead of the historical `"slo.md.tmpl"`, which had to be stripped back down to `slo.yaml` on every call and read like a bug to anyone new to the code. `ArtifactNameFor` is now idempotent (it also trims a trailing `.yaml`) and still accepts the pre-v1.0 spellings `<name>.md.tmpl` / `<name>.tmpl`, so external tooling and user directories written against those names keep resolving. No user-facing behavior change; the legacy filenames passed to `warnStaleLegacyFiles` are untouched, since those name files on disk rather than templates to load.
- `TEMPLATES.md` (shipped to user template dirs by `srekit templates init`): the placeholder reference documented v0.13.x filenames (`postmortem.md.tmpl`) and root-level fields (`.Title`), while v1 artifacts are `postmortem.yaml` with everything under `.Meta` — following it produced templates that failed to render. Headings and every field path corrected, plus notes on `version: 1` and section `id` stability. It also advertised `--template FILE` on `srekit runbook`, which rejects the flag as unknown; that whole section is gone, since `--template` was removed in this same release (see Removed).
- `architecture.md` (both locales): the `internal/render.Render()` section now says the artifact branch takes a bare name and describes `ArtifactNameFor`'s normalization.
- `contributing.md` (both locales): replaced the stale "`tmpl.Default` race" and "`tmpl.Default` is package-level mutable state" caveats with a "Global state in tests" section. The global loader was removed back in 0.10.2 (per-command-tree loader via `cmd.Context()` / `loaderFrom`), so the documented `resetTmplDefault(t)` + no-`t.Parallel()` workaround no longer exists and the advice pointed against the current architecture. The remaining global — the `config.Global()` singleton and its `withConfig(t, kv)` test helper — is now documented in its place.

### Removed

- **Breaking — three generators removed: `capacity`, `retro` and `license` (along with the `lic` alias).** The last release that ships them is 0.29.3. `srekit` generates the artifacts an on-call engineer or a reliability team owns: a sprint retro is an agile ceremony, capacity planning is spreadsheet work, and a LICENSE is a one-off repository setup step inherited from the `lic` command in the gch monolith. The names do not vanish silently: hidden stubs in `cmd/retired.go` exit non-zero with an explanation naming the removal release and the migration note, and they do not parse arguments (`srekit retro` without `--team` reports the removal, not the missing flag). The stubs go away at 1.0. What to do instead: keep an already-generated document as a static template, pin 0.29.3, pull the template body out with `git show v0.29.3:internal/tmpl/templates/capacity.yaml`, or use your host's license picker for LICENSE. Full guide in [docs/migration/removed-commands.md](https://jtprogru.github.io/srekit/migration/removed-commands/).
- **Breaking — the `--template FILE` flag is gone entirely.** `license` was the only command whose render path read it; once that command went, the flag would have been silently ignored, which this project treats as a defect. Every command now answers `Error: unknown flag: --template`. Per-artifact customization is a `<name>.yaml` in `templates_dir`.
- The embedded artifacts `capacity.yaml` and `retro.yaml` are no longer compiled in. User template directories are untouched: a leftover file reclassifies from `customized` to `user-only` in `templates list`, and its snapshot is collected by the next `templates upgrade`.
- Internal (not public contract): the flag took with it the legacy `text/template` branch in `render.buildBody` — its last consumer — plus `render.WriteRaw`, `render.Options.TemplatePath`, the always-true `render.Options.RenderArtifact`, `cliflags.Output.BindTemplateFlag` and `tmpl.ParseFile`. The v1 artifact format is now the only render path.
- Internal: the code that leaned on the removed branch went next. `render.Options.BootstrapJSON` and the envelope that wrapped rendered markdown into a synthetic `{meta, sections:[{id:"body"}]}`, together with `render.extractH1`: the envelope existed for a generator without a sections manifest, and there has been none since v0.20.0. `cliflags.Output.RenderOptionsStructured` collapses into `RenderOptions` — all it did was clear `BootstrapJSON`. `tmpl.Loader.Parse` returned a parsed `text/template`, which only the removed branch wanted; artifacts are loaded as bytes through `LoadArtifactBytes` and parsed in `internal/sections`. The public JSON contract `{meta, sections:[{id,title,type,body,required}]}` is unchanged — the generators still build it themselves.

### Fixed

-

## [0.29.3] - 2026-06-15

### Changed

- **Binary ~35% smaller (7.84 MB → 5.07 MB) by dropping two dependency chains.** Replaced `google/uuid` (which pulled `crypto/rand` → the FIPS-140 module) with a `math/rand/v2` UUIDv4 — artifact IDs are not security-sensitive. Replaced Viper with a tiny `internal/config` package over the already-present `go.yaml.in/yaml/v3`: srekit only used four `GetString` lookups, one YAML file, and `SREKIT_`-prefixed env vars, while Viper dragged in `afero` → `net/http` → `crypto/tls` + `crypto/rand` + FIPS-140, plus `go-toml`, `mapstructure`, `cast`, `gotenv`, `locafero`, `fsnotify`, and `pflag` extras. Config precedence (explicit override > env > file) is preserved; the `crypto` and `net/http` packages are now absent from the build graph entirely.
- **Release builds: added `-trimpath` and Linux-only UPX compression.** `-trimpath` drops local filesystem paths for reproducible builds; UPX is enabled only for Linux artifacts (it breaks code signing / Gatekeeper on macOS, where srekit ships via a Homebrew cask).

## [0.29.2] - 2026-06-10

### Changed

- **CI: bumped pinned GitHub Actions in the `actions` dependabot group (#11).** `github/codeql-action/upload-sarif` v3 → v4 in `.github/workflows/security.yaml`, `codecov/codecov-action` v6 → v7 in `.github/workflows/tests.yaml`. No source changes; cuts a CI deprecation warning before v3 of codeql-action is retired.

## [0.29.1] - 2026-06-10

### Security

- **Path-hardening sweep across the artifact generators and the template loader.** Triggered by an internal security audit of v0.29.0; all findings were Low-severity hygiene gaps (no reachable RCE / no privilege boundary crossing under the documented threat model), but each is now closed.
    - `cmd/oncall.go`: `--start` is now `ids.Slug`-ified when forming the default output filename, matching every other generator (`postmortem`, `runbook`, `slo`, `rfc`, `task`, `retro`, `ebp`, `capacity`). Previously a value like `--start ../../etc/foo` would land at a relative path under the working directory. `YYYY-MM-DD` inputs are preserved unchanged by `ids.Slug`.
    - `internal/tmpl/tmpl.go`: `DirSource.Read` now refuses any `name` that contains path separators or is not a basename, returning `fs.ErrNotExist` (so the loader falls through to the next source). Currently unreachable — every call site uses a hard-coded artifact name — but this closes a regression hole if a dynamic template name is ever wired in (e.g. plugin-style dispatch).
- **Release workflow actions pinned to commit SHAs.** `.github/workflows/goreleaser.yaml` now pins `actions/checkout`, `actions/setup-go`, and `crazy-max/ghaction-import-gpg` to commit SHAs (v6.0.3 / v6.4.0 / v7.0.0 respectively), matching the existing pin on `goreleaser/goreleaser-action`. A moved major tag can no longer silently swap action code in the signing-capable release runner.
- **`SECURITY.md`**: documented that `.githooks/pre-commit` is opt-in (lives under `.githooks/`, not `.git/hooks/`) and that activation requires an explicit `git config core.hooksPath .githooks`. Clarifies for reviewers that cloning this repo does not execute the `tokei`-driven LoC-badge updater.

## [0.29.0] - 2026-06-05

### Removed

- **Breaking — `srekit incident` subcommand removed.** The live-incident report is no longer scaffolded by srekit. Rationale: a markdown file edited manually during an active incident is the wrong substrate for live coordination — that work belongs in an IM tool (Slack/PagerDuty/incident.io) or a future API-driven log ingestion path, not in a one-shot CLI scaffold. The artifact had the thinnest content-guidance surface of any generator (status / lead / comms / update log — four rules) and overlapped heavily with `srekit postmortem`'s timeline, where it could be linked from when needed. Affected: the `incident` command itself, `internal/tmpl/templates/incident.yaml`, `docs/{en,ru}/commands/incident.md`, the mkdocs nav entry, and cross-references from `runbook` / `postmortem` / `commands/index` / `guides/json-output`. Generator count drops from 12 to 11. Migration: if you had `templates_dir/incident.yaml`, it becomes a `user-only` artifact (`templates list` will surface it) — delete it manually or keep it as a private template; the CLI no longer reads it. Historical references in `docs/{en,ru}/migration/v1*.md` are intentionally preserved as a record of the v0.18 / v0.22 migrations.

## [0.28.0] - 2026-06-04

### Added

- **`SECURITY.md` — published vulnerability reporting policy.** Documents the reporting channel (GitHub private advisory or email), supported-versions story for the 0.x line, scope (what srekit is responsible for vs. what's the user's own trust), the threat model (single-user CLI, no network listener, no auth layer), what we run in CI to keep ourselves safe (govulncheck + gosec + Bearer), and the FuncMap-surface caveat for anyone embedding `srekit` in a workflow that ingests third-party `<name>.yaml`.

- **`govulncheck` step in `.github/workflows/security.yaml`** — blocking on every push / PR. Catches known CVEs in direct and transitive Go dependencies; complements the gosec scan that already runs via `golangci-lint`.

### Changed

- **Security audit ran end-to-end against the v0.27.0 codebase.** Tools: `gosec` (latest), `govulncheck` (latest), manual review of file writes, subprocess invocations, YAML/template parsing, user input handling. Result: **no reachable vulnerabilities.** The gosec scan returned 39 findings, all triaged as intentional design choices:
    - 14 × G306 (`WriteFile` perms `0o644`) — scaffolded templates and rendered documents are public artifacts; `0o644` is correct. Already annotated with `//nolint:gosec` in code.
    - 16 × G304 (file inclusion via variable) — a CLI tool's job is to read paths supplied by the user (`--templates-dir`, `--template`, `--from`). Not a vulnerability.
    - 6 × G302 (directory perms `0o755`) — `templates_dir` is public, parent of a git working tree.
    - 6 × G204 (subprocess with variable) — every `exec.Command` invocation is `git` with arguments passed as separate argv elements (never `sh -c`); the only "variable" component is a directory path or known-set git subcommand. Not exploitable.
    - 3 × G703 (path traversal via taint) — confirmed false positives; the variable in each case is either an embedded-set filename or a user-explicit dir from a CLI flag.

  `govulncheck` returned **0 vulnerabilities**.

  This audit is now self-monitoring: govulncheck runs on every push / PR (blocking), gosec runs inside golangci-lint on every push / PR. The findings catalog and rationale live in this CHANGELOG entry as the audit baseline.

## [0.27.0] - 2026-06-04

### Fixed

- **`docs/{en,ru}/commands/*.md` — per-command doc sweep.** Doc-only release; no code changes. The v0.23 doc audit caught the headline `--template` and bootstrap-envelope issues but only updated `task.md` + `templates.md`; the other 10 command docs still had stale shapes. This sweep covers the rest. Issues found and fixed:

  **Phantom CLI flags** (documented flags that don't exist in code):
    - `docs/{en,ru}/commands/incident.md`: listed `--comms` and `--start` flags. Neither has ever existed on `cmd/incident.go`; the real flags are `--title`, `--severity`, `--lead`, `--status`. Dropped from the flag table and the "fully populated" example; added the missing `-T` short form and the actual default values (`--severity SEV-2`, `--status active`).
    - `docs/{en,ru}/commands/ebp.md`: listed `--owner` flag. Doesn't exist; the policy owner is a fill-in inside the rendered `meta_bullets`. Dropped from flag table and example; added a note about editing the rendered file directly.

  **Wrong default values:**
    - `docs/{en,ru}/commands/capacity.md`: `--horizon` was documented as default `6m`. Real default is `1y` (see `cmd/capacity.go`). Fixed.

  **Stale "Template shape" Go-struct snippets**: every per-command doc except `task` (fixed in v0.23) and `postmortem` still showed a Go struct like `{ ID, Title, Service, Now string }` — outdated since the v0.14–v0.20 migration moved each generator from a flat data root to `.Meta.<Field>`. Replaced every snippet with a one-paragraph description pointing at the v1 YAML artifact (`internal/tmpl/templates/<name>.yaml`), listing the section IDs, naming the `.Meta.<Field>` fields available, and linking to `postmortem.md` as the canonical schema reference. Affected: `incident`, `rfc`, `runbook`, `slo`, `ebp`, `capacity`, `retro`, `oncall-report`, `changelog`, `license`.

  **Missing documentation:**
    - `docs/{en,ru}/commands/license.md` didn't document `--template FILE`. License is the only command that honors the flag (since the v0.22 cleanup); the flag table now includes it with the "license-only" caveat.

  **Stale lead paragraph:**
    - `docs/{en,ru}/commands/postmortem.md`: opened with "As of v0.13.0, the postmortem command is the **first** generator backed by a sidecar sections manifest". It's no longer the only one (everything is on v1 yaml now) and the sidecar layout was retired in v0.14.0. Rewritten as "Postmortem was the prototype for v1 (v0.14.0) and is now the canonical schema reference for the rest of the generators".

## [0.26.0] - 2026-06-04

### Fixed

- **`srekit templates migrate` — sidecar-driven `.tmpl` no longer leaks `{{ range .Sections }}` into `header_body`.** When a sibling `<name>.sections.yaml` provides the section list, the `.tmpl` body between the header and the first `##` was the legacy render loop (`{{ range .Sections }}…{{ end }}`) — Go-template code, not prose. The migrator now suppresses `header_body` in that case (it was previously copied verbatim, which produced a parse error at render time: `template: section:2: unexpected EOF`). Regression test added.

- **`docs/{en,ru}/migration/v1.md` — recipe correctness fixes** found during the v0.26.0 end-to-end smoke audit:
    - **Step ordering reversed**: `templates migrate` runs **before** `templates upgrade`, not after. The old order scaffolded the embedded `.yaml` first, after which `templates migrate` skipped every customized `.tmpl` (target already exists) and silently dropped the user's customizations. New "Order matters" callout explains why.
    - **Sed pattern made smart**: the `{{ \.([A-Z]…)` regex matches `.Meta.X` too — running it blindly on a file already on `.Meta.X` produces `.Meta.Meta.X` (a render-time error). The new recipe guards each file with `grep -q '{{ \.Meta\.' && continue`, so postmortem (the one generator that was already on `.Meta.X` in v0.13.x) is left alone.
    - **Removed `srekit config show` reference**: that subcommand doesn't exist (only `srekit config init`). The fallback `|| echo ~/.srekit/templates` hid the error, but the snippet was misleading. Replaced with a plain note about where the templates dir is configured.
    - Recipe step 5 now ends with a render-smoke command (`srekit --templates-dir "$TPL" postmortem --title smoke … > /dev/null`) before the `rm` of legacy files.
    - Added the GNU-sed reminder (`-i` without trailing `''`) for Linux users.

## [0.25.0] - 2026-06-04

### Added

- **README "Стабильность и версионирование" section.** First explicit public statement of what stabilizes at v1.0 (CLI flags, `--json` shape + section IDs, `<name>.yaml` schema, section `type` vocabulary, config keys & env), what's kept for backwards compat through 1.x (legacy `.tmpl` / `.sections.yaml` reading), and what's not stable yet (`frontmatter` free-form, exact WARN wording, internal/* Go API). Includes the deprecation cycle we follow (WARN → silent no-op → removal in next major, with `--template FILE` as a worked example).

### Changed

- **README — removed obsolete `--template FILE` example for `srekit runbook`.** The flag was a silent no-op on every artifact-path command since v0.20.0 and rejected outright since v0.22.0; the README example showing `srekit runbook --template …` was actively misleading. Replaced with a note that `--template` is `license`-only and a pointer to the `<name>.yaml` customization model.
- **README `templates validate` description rewritten** to reflect per-format checks (`ParseArtifact` for `.yaml`, `ParseManifest` for legacy `.sections.yaml`, parse-only for `.tmpl`). Dropped the "field-name typo detection" example — that path required a non-empty `tmpl.Samples` registry, which v0.21.0 emptied.
- **README `templates diff` / `templates list` descriptions** updated to mention all three artifact suffixes (`.yaml` / `.tmpl` / `.sections.yaml`) instead of just `*.tmpl`.

### Fixed

- **`docs/{en,ru}/commands/task.md` default filename**: was `<path>/Tasker - <title>.md` (legacy from `gch sretask`), code has shipped `<path>/investigation-<slug>.md` for several releases. Doc + example fixed.

## [0.24.0] - 2026-06-04

### Added

- **`docs/{en,ru}/migration/v1-history.md`** — per-release journal of the YAML-first migration (v0.14.0 → v0.20.0 plus the v0.21–v0.23 stabilization releases). Lifted from the old `v1.md`, plus new entries for v0.21–v0.23 that weren't documented in the migration tree before.

### Changed

- **`docs/{en,ru}/migration/v1.md` rewritten as an action-oriented upgrade guide.** The old page was a chronological journal of "what changed in vX.Y.Z" — useful for git-archaeology, useless for someone on v0.13.x who wants to know "what do I run now". The rewrite leads with a quick "where am I?" check, then a 5-step copy-pasteable upgrade recipe (`templates upgrade` → `templates migrate` → `sed` rewrites → `templates validate` → delete legacy files). The journal moves to the new `v1-history.md` appendix. Stability boundaries are now split into three clean buckets (stable / compat-only / not-yet-stable). Troubleshooting kept and expanded with the section-title raw-syntax case.

- **`mkdocs.yml` nav** updated: `Migration → Upgrade guide` + `Migration → History`.

### Fixed

-

## [0.23.0] - 2026-06-04

### Changed

- **Docs aligned with the v0.20–v0.22 state.** Doc-only release; no code changes. The migration through v0.20 retired every embedded `.tmpl` and every bootstrap-envelope `--json` shape, and v0.22 narrowed `--template FILE` to the `license` command. Most user-facing docs hadn't caught up; this release sweeps them.

  Notable updates (EN + RU parallel):

  - `guides/json-output.md`: removed the "structured vs bootstrap" two-mode table — every shipped generator is structured. The `.sections[0].body` example for "any other generator" was the most user-visible lie; replaced with per-section `jq '.sections[] | select(.id=="…").body'` examples for runbook + changelog. Per-command payload shapes updated (no more `[{id:"body", …}]` envelopes). Breaking-change callout now covers both the 0.12→0.13 flat→nested change and the 0.13→0.20 bootstrap→structured change.
  - `getting-started.md` + `commands/index.md`: dropped `--template FILE` from the shared-flags table; added an explicit note that it's a `license`-only flag now, with a pointer to `templates_dir` customization for the rest.
  - `architecture.md`: dropped the "pin to v0.10.1" callout (the structure has moved on); embed pattern now `*.yaml`; added `internal/sections` and `internal/migrate` to the package tour; `tmpl.Default` mention removed (it was retired in v0.10.2); render section documents the three current branches (JSON / RenderArtifact / legacy); cliflags section documents `BindTemplateFlag`; "add new artifact" recipe now points at `<name>.yaml` instead of `.md.tmpl`; added `internal/sections/render_artifact.go` as the spot for markdown-layout tweaks; test-strategy row mentions `newFixtureLoader(t)`.
  - `guides/custom-templates.md`: scaffold count `13 files + TEMPLATES.md` → `11 .yaml + TEMPLATES.md`; customization example edits `runbook.yaml` (not `runbook.md.tmpl`); validate output shows `.yaml` files; upgrade-flow table headed by `task.yaml`. Added a sentence pointing at the postmortem reference for the v1 artifact schema, and clarified that bespoke `.tmpl` gets parse-only validation now.
  - `recipes.md`: drift-detection example shows `.yaml` filenames instead of `.md.tmpl`.
  - `commands/templates.md`: `templates list` now describes the embedded-set comparison as covering `.yaml` / `.tmpl` / `.sections.yaml`; `templates validate` rewritten with per-format checks (ParseArtifact for `.yaml`, ParseManifest for legacy `.sections.yaml`, parse-only for `.tmpl`).
  - `commands/task.md`: shared-flags list drops `--template`; "Template shape" section rewritten to describe the v1 YAML artifact and `.Meta.<Field>` template syntax (the old struct snippet was outdated since v0.15.0).

### Fixed

-

## [0.22.0] - 2026-06-04

### Changed

- **Breaking — `--template FILE` removed from every generator except `license`.** As of v0.20.0 the flag was a silent no-op on every artifact-path command (postmortem, task, retro, slo, ebp, capacity, incident, rfc, runbook, oncall-report, changelog) because the loader resolves `<name>.yaml` before any TemplatePath check. A silently-ignored CLI flag is worse than a missing one, so the binding is removed. Users running `srekit postmortem --template foo.tmpl` (or any other migrated generator) now get an honest `Error: unknown flag: --template` instead of having their flag swallowed. The flag still works on `srekit license`, which is the only command whose render path genuinely honors `opts.TemplatePath`. Per-artifact customization for the other commands remains: drop a `<name>.yaml` into `templates_dir`.

- **`cliflags.Output.Bind` no longer registers `--template`.** New helper `cliflags.Output.BindTemplateFlag(cmd)` registers the flag explicitly; only `cmd/license.go` calls it. Same `Output.TemplatePath` field, same render plumbing — only the CLI surface narrows.

### Fixed

-

## [0.21.0] - 2026-06-04

### Changed

- **Pruned dead code after the YAML-first migration.** Removes `internal/tmpl.LoadManifestBytes` + `ManifestNameFor` (resolved sidecar `<name>.sections.yaml` for the legacy loader path — last shipped caller was v0.13.x postmortem fallback, retired in v0.14.0; no code path reaches them in v0.20.0) and their tests. Shrinks `tmpl.Samples` to `{}` (the `.tmpl` typo-detection fixture registry — empty now that no `.tmpl` ships); the variable is kept as a hook for external tooling that ships custom `.tmpl` artifacts via `tmpl.Validate`.

- **Updated doc comments to reflect the v0.20 state.** `internal/sections` package doc + Manifest type doc clarify Manifest exists for backwards compatibility (`templates validate` / `templates migrate`) rather than for shipped generators; `internal/render.Options` BootstrapJSON / RenderArtifact docs describe current defaults instead of "legacy commands"; `internal/tmpl` FS / IsTemplateArtifact / EmbeddedNames doc rewrites reflect "every shipped artifact is v1 YAML".

- **`cmd/legacy_warn.go` WARN wording**: "`pre-v0.14.0 format`" → "`legacy pre-v1.0 format`". The migration spans v0.14–v0.20, not a single release.

### Fixed

-

## [0.20.0] - 2026-06-04

### Added

- **Section titles are template-evaluated.** Both `sections.Merge` and `RenderArtifact` now run section titles through the FuncMap before emitting (markdown H2 / structured `--json` `title`). This lets generators like `changelog` use dynamic headings (e.g. `[{{ .Meta.InitialVersion }}] - {{ .Meta.Today }}`). Previously titles were emitted verbatim and template syntax leaked into both outputs.

### Changed

- **`changelog` migrated to the v1 artifact format.** **All 11 generators are now on v1.** No `.tmpl` artifact ships in embed anymore; the `//go:embed templates/*.tmpl` half of the embed pattern was dropped (Go's embed requires at least one matching file per glob). The Samples registry (sample-data fixtures used by `templates validate` for `.tmpl` typo-checks) is empty; `tmpl.Validate` now always returns `ErrUnknownTemplate` for `.tmpl` inputs (parse-only validation). YAML artifacts are structurally validated via `sections.ParseArtifact`.

- **Breaking — `srekit changelog --json` shape.** Moves from bootstrap envelope to structured (`sections: [{id:"unreleased",...}, {id:"initial_release",...}]`). Migration: replace `jq '.sections[0].body'` with per-section access (`jq '.sections[] | select(.id=="initial_release").body'`).

- **Breaking — `changelog.md.tmpl` no longer ships in embed.** Users with customized copies get stderr WARN. `srekit templates upgrade` scaffolds `changelog.yaml`; `srekit templates migrate` auto-converts the legacy `.tmpl`.

- **Breaking — `--template FILE` is a no-op for every shipped command.** Already documented in v0.19.0 for the artifact-path commands; with changelog migrated, no command honors the flag anymore. The flag is kept on the CLI surface for backwards compatibility but does nothing. Per-artifact customization is via dropping a `<name>.yaml` into `templates_dir`.

- **Test suite decoupled from embedded `.tmpl`.** Render and tmpl unit tests previously depended on a specific `.tmpl` shipping in embed; that was the v0.17–v0.19 source of test churn each migration. They now use a `newFixtureLoader(t)` helper that writes a per-test `.tmpl` fixture into a temp dir and returns a `*tmpl.Loader` against it. CLI integration tests in `cmd/cmd_test.go` swap embed-target references from the (now-gone) `changelog.md.tmpl` to `postmortem.yaml`. `TestPerCommandTemplateOverride` was deleted (no command honors `--template` anymore); `TestChangelogJSONBootstrap` was rewritten as `TestChangelogJSONStructured`.

### Fixed

- **Frontmatter-less artifacts render without a leading blank line.** changelog.yaml has no frontmatter; the renderer correctly skips the `---` block and starts the document with the H1.

## [0.19.0] - 2026-06-04

### Changed

- **`runbook` and `oncall-report` migrated to the v1 artifact format.** Same dogfooded recipe as incident + rfc in v0.18.0. After this release, **only `changelog`** remains as a `.tmpl` artifact; v0.20.0 finishes the YAML-first migration. Both commands now use the artifact render path with structured `--json` (per-section access).

- **Breaking — `srekit runbook --json` and `srekit oncall --json` shape.** Moves from bootstrap envelope (`sections: [{id:"body", body:<markdown>}]`) to structured. Migration: replace `jq '.sections[0].body'` with `jq '.sections[] | select(.id=="...").body'`. Same pattern as every prior YAML-first migration; documented in `docs/{en,ru}/migration/v1.md` (v0.19.0 section).

- **Breaking — `runbook.md.tmpl` / `oncall.md.tmpl` no longer ship in embed.** Users with customized copies get stderr WARN. `srekit templates upgrade` scaffolds the new `.yaml`; `srekit templates migrate` auto-converts the legacy `.tmpl`.

- **Breaking — `--template FILE` flag is now effectively a no-op for migrated commands.** It was originally a one-shot escape hatch for the legacy `text/template` rendering path. Artifact-path commands (postmortem, task, retro, slo, ebp, capacity, incident, rfc, runbook, oncall) ignore it because the artifact loader is consulted before any TemplatePath check; only `changelog` (the last bootstrap-envelope command) still honors it. For per-artifact customization, drop a `<name>.yaml` into your `templates_dir` — that path is the v1 customization model.

### Fixed

- **`oncall.yaml` "Pages" section preserves the trailing prose line** ("Всего пейджеров: …") that sat after the GFM table in the legacy `.tmpl`. The auto-converter heuristic classified the section as `type: table` and dropped that line; the shipped artifact stores the whole section as `type: text` with the table embedded verbatim, preserving the original layout byte-for-byte.

- **`runbook.yaml` section id rewritten** from the slugifier's awkward `slo_severity_slo_impact` (the heading "Тяжесть и влияние на SLO (Severity & SLO impact)" double-counted "SLO") to a cleaner `severity_slo_impact`. Section IDs are part of the structured-JSON contract, so manual cleanup at ship time is worth the diff churn.

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

[Unreleased]: https://github.com/jtprogru/srekit/compare/v0.30.0...HEAD
[0.30.0]: https://github.com/jtprogru/srekit/compare/v0.29.3...v0.30.0
[0.29.3]: https://github.com/jtprogru/srekit/compare/v0.29.2...v0.29.3
[0.29.2]: https://github.com/jtprogru/srekit/compare/v0.29.1...v0.29.2
[0.29.1]: https://github.com/jtprogru/srekit/compare/v0.29.0...v0.29.1
[0.29.0]: https://github.com/jtprogru/srekit/compare/v0.28.0...v0.29.0
[0.28.0]: https://github.com/jtprogru/srekit/compare/v0.27.0...v0.28.0
[0.27.0]: https://github.com/jtprogru/srekit/compare/v0.26.0...v0.27.0
[0.26.0]: https://github.com/jtprogru/srekit/compare/v0.25.0...v0.26.0
[0.25.0]: https://github.com/jtprogru/srekit/compare/v0.24.0...v0.25.0
[0.24.0]: https://github.com/jtprogru/srekit/compare/v0.23.0...v0.24.0
[0.23.0]: https://github.com/jtprogru/srekit/compare/v0.22.0...v0.23.0
[0.22.0]: https://github.com/jtprogru/srekit/compare/v0.21.0...v0.22.0
[0.21.0]: https://github.com/jtprogru/srekit/compare/v0.20.0...v0.21.0
[0.20.0]: https://github.com/jtprogru/srekit/compare/v0.19.0...v0.20.0
[0.19.0]: https://github.com/jtprogru/srekit/compare/v0.18.0...v0.19.0
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
[0.10.2]: https://github.com/jtprogru/srekit/compare/v0.10.1...v0.10.2
[0.10.1]: https://github.com/jtprogru/srekit/compare/v0.10.0...v0.10.1
[0.10.0]: https://github.com/jtprogru/srekit/compare/v0.9.0...v0.10.0
[0.9.0]: https://github.com/jtprogru/srekit/compare/v0.8.0...v0.9.0
[0.8.0]: https://github.com/jtprogru/srekit/compare/v0.7.0...v0.8.0
[0.7.0]: https://github.com/jtprogru/srekit/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/jtprogru/srekit/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/jtprogru/srekit/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/jtprogru/srekit/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/jtprogru/srekit/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/jtprogru/srekit/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/jtprogru/srekit/releases/tag/v0.1.0
