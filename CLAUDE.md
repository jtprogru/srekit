# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`srekit` is a single-binary Go CLI that generates SRE text artifacts (investigation log, postmortem, runbook, RFC/ADR, on-call report, SLO, error budget policy, changelog) from templates compiled into the binary via `//go:embed`. Extracted from the [gch](https://github.com/jtprogru/gch) monolith. Pre-1.0 (0.30.x line). `capacity`, `retro` and `license` were removed in 0.30.0; `cmd/retired.go` keeps hidden stubs that explain the removal until 1.0.

## Commands

```bash
task ci              # lint + race tests — the one-shot pre-push check
task test            # go test --short -coverprofile=cover.out -v ./...
task test:race       # go test -race -v ./...  (what CI runs)
task lint            # golangci-lint -v run
task build           # CGO_ENABLED=0 go build -o ./dist/srekit .
task run -- <args>   # go run . <args>
task release:dry     # goreleaser snapshot build into ./dist, no publish
task docs:serve      # MkDocs at http://127.0.0.1:8000
task docs:build      # mkdocs build --strict
```

Single test / subset:

```bash
go test ./cmd/ -run TestSLO -v
go test ./internal/sections/ -run TestRenderArtifact -v
go test ./cmd/ -run TestTemplates -v      # run this whole suite when touching cmd/templates.go — the 3-way merge has subtle invariants
```

Toolchain must match CI or gosec taint rules produce false positives: Go **1.26.4**, `golangci-lint` **v2.12** (~50 linters enabled, including `gochecknoglobals`).

## Architecture

Flow for every generator command:

```
cmd/<name>.go                     flags → <name>Meta struct
  └─ loaderFrom(cmd)              *tmpl.Loader off cmd.Context()
  └─ loader.LoadArtifactBytes()   resolves <name>.yaml through Sources
  └─ sections.ParseArtifact()     → Artifact (structural validation)
  └─ sections.Merge()             defaults + --from overrides, template-evaluated
  └─ out.RenderOptions(cmd, defaultPath)
  └─ render.Render()              → sections.RenderArtifact() → markdown → writeBody
```

Key pieces:

- **`internal/tmpl`** — `Source` interface (`EmbedSource` for `//go:embed templates/*.yaml`, `DirSource{Dir}` for a user dir). `Loader{Sources}` walks them in order treating `fs.ErrNotExist` as fall-through, so a missing file in the user dir transparently falls back to embedded. Also the shared `template.FuncMap` (`default`, `shortID`, `slugify`, `upper`, `lower`, `trim`, `now`).
- **`internal/sections`** — the v1 artifact runtime. `Artifact` = `version` / `frontmatter` (a `yaml.Node`, so author key order survives parse→render) / `title` / `meta_bullets` / `header_body` / `sections`. Section types are `text` / `list` / `table`. `Merge` overlays per-section overrides; `RenderArtifact` composes the markdown (frontmatter block → H1 → meta bullets → header body → `## section` blocks).
- **`internal/render`** — `buildBody` has two branches: the `--json` short-circuit (`MarshalIndent` of an already-structured payload) and the v1 artifact path. 0.30.0 removed the legacy `text/template` branch with `--template FILE` and `license`, its last caller, and with it the `BootstrapJSON` envelope and `RenderOptionsStructured` — no generator had set the flag since 0.20.0. `writeBody` owns all `--out` / `--stdout` / `--force` / `--dry-run` / `--quiet` routing; `-` means stdout.
- **`internal/cliflags`** — `Output.Bind(cmd, outDesc)` wires the shared flags. There is no `--template` binding anywhere (see invariants).
- **`cmd/paths.go`** — XDG resolution for config and templates dir, with legacy-path precedence.
- **`internal/config`** — hand-rolled YAML+env config (`SREKIT_*` prefix). Deliberately not viper.
- **`internal/clock`** — `var Now = time.Now` indirection so tests can pin the wall clock (e.g. the Sunday on-call week-boundary test).
- **`internal/migrate`** — heuristic `.tmpl` → `.yaml` converter behind `srekit templates migrate`.

Generators identify an artifact by its **bare name** — `render.Render(..., "slo", ...)` and `loader.LoadArtifactBytes("slo")` resolve `internal/tmpl/templates/slo.yaml`. `tmpl.ArtifactNameFor` normalizes the name and is idempotent, and it still accepts the pre-v1.0 spellings (`slo.md.tmpl`, `slo.tmpl`, `slo.yaml`) so external tooling keeps working. The literal `.md.tmpl` / `.sections.yaml` strings that remain in `cmd/*.go` are arguments to `warnStaleLegacyFiles` — they name files on the *user's* disk to warn about, not templates to load. Don't collapse them.

`cmd/templates.go` implements the lifecycle (`list` / `init` / `upgrade` / `pull` / `validate` / `diff`). `upgrade` does a 3-way merge whose base is a per-template snapshot at `<templates-dir>/.srekit-embedded/<name>`, appended to that dir's `.gitignore`.

## Invariants

These are load-bearing; breaking one is a defect, not a style choice.

- **Dependency minimalism.** Production deps are only `spf13/cobra` and `go.yaml.in/yaml/v3`. `viper` and `google/uuid` were removed precisely because they pulled `afero → net/http → crypto/tls` into the build graph. Neither `net/http` nor `crypto` may appear in the graph. Weigh any new dependency against binary size and say so explicitly.
- **No package-level mutable state.** The template loader lives in `cmd.Context()` (set by `configureTemplates` in `PersistentPreRunE`, read via `loaderFrom`) specifically so parallel tests don't race on a global. `gochecknoglobals` is enabled.
- **Uniform output contract.** Every generator supports `--out` / `--stdout` / `--force` / `--dry-run` / `--json` plus the persistent `--quiet`. A flag a command silently ignores must not exist — that is why `--template` no longer exists at all: `license` was the only command whose render path read it, and both went in 0.30.0.
- **snake_case in YAML, camelCase in JSON.** Artifact YAML is hand-edited by SRE/devops authors (Ansible/K8s precedent); JSON output keys are camelCase across every command including `templates list --json`. Both are public contract; `tagliatelle` is suppressed with a comment where they meet.
- **The JSON contract** `{meta, sections:[{id,title,type,body,required}]}` is public and stabilizes at 1.0. Section bodies from `--from` are inserted verbatim, never template-evaluated.
- **An unknown section ID in input is always an error**, never a silent skip.
- **XDG for fresh installs, legacy wins if present.** If `~/.srekit.yaml` or `~/.srekit/templates` already exists it takes precedence over the XDG path, so nobody ends up with a config file that sits there unread.
- **`--json` never falls through to the markdown default path**, so JSON can't accidentally land in a `.md` file.
- **Templates are bilingual** — headings and labels as `Русский (English)`, body in Russian. Technical identifiers (SLO/SLI/RFC/SEV/UTC/PromQL) and frontmatter keys stay English. `changelog` is fully English so Keep a Changelog tooling doesn't break.
- Reading legacy `.tmpl` / `.sections.yaml` from a user templates dir is still supported through 1.x, with a stderr `WARN` (`cmd/legacy_warn.go`). Removal is a 2.0 candidate.

## Adding a generator

1. `internal/tmpl/templates/<name>.yaml` — the v1 artifact (`version: 1` required).
2. `cmd/<name>.go` — copy an existing generator (`cmd/slo.go` is the smallest complete one). Define `<name>Meta` with JSON tags, `<name>Data` with `Meta` + `Sections`, and `ArtifactPayload() ([]sections.RenderedSection, any)`.
3. Register it in `NewRootCmd()` in `cmd/root.go`.
4. Smoke test in `cmd/cmd_test.go` via the `runCLI(t, ...)` helper (`cobra.Command.SetArgs` + captured stdout).
5. Docs in **both** `docs/en/commands/` and `docs/ru/commands/`, plus the `mkdocs.yml` nav.
6. `CHANGELOG.md` entry under `[Unreleased]`.

`postmortem` is the canonical v1 reference and the only command with the full round-trip (`--json` → edit → `--from`, plus `--schema` and `--validate`); mirror it when extending the structured path.

## Testing conventions

Tests are parallel (`t.Parallel()`) except those that touch the global config — `withConfig(t, kv)` calls `config.Reset()` and is explicitly not parallel-safe. It is the only global-state helper left; the template loader is per-command-tree, so `--templates-dir` tests parallelize freely. Render/tmpl unit tests build a loader from a per-test fixture in a temp dir rather than depending on what is currently embedded.

## Conventions

Conventional-commit prefixes. `CHANGELOG.md` is hand-maintained in Keep a Changelog format — new entries go under `[Unreleased]`, never into an already-released version. No `Co-Authored-By` or generated-by markers in commits or PR descriptions. Don't suppress a linter without justification; use the existing `//nolint:<linter> // <code>: <reason>` form. Markdown is not hard-wrapped: one paragraph is one line.

Behavior changes must update `docs/en` **and** `docs/ru` — the two locales are kept in sync, and English-only is an incomplete change.

Release: move `[Unreleased]` into `[X.Y.Z] - YYYY-MM-DD`, commit `release: X.Y.Z`, then `git tag -a vX.Y.Z && git push origin vX.Y.Z`. goreleaser builds linux/darwin/freebsd × amd64/arm64, GPG-signs checksums, and rewrites the `jtprogru/homebrew-tap` cask. `Version` / `Commit` / `Date` / `BuiltBy` in `cmd/root.go` are ldflags-injected.

## OpenSpec

The repo uses a spec-driven workflow — `openspec/config.yaml` holds the authoritative project context and per-phase rules, `openspec/specs/` the current capability specs, `openspec/changes/` in-flight proposals. Skills exist under `.claude/skills/openspec-*`. When a task involves a spec change, read `openspec/config.yaml` first; its rules go further than this file (e.g. proposals must flag **BREAKING** public-contract changes, specs must be written in observable CLI behavior with no Go type names).
