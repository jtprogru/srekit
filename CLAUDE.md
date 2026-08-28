# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`srekit` is a single-binary Go CLI that generates SRE text artifacts (investigation log, postmortem, runbook, RFC/ADR, on-call report, SLO, error budget policy, changelog) plus one document that is not an SRE artifact at all — `tasker`, a card for a collection of engineering tasks — from templates compiled into the binary via `//go:embed`. It also *maintains* one artifact it did not necessarily generate: `changelog release` and `changelog validate` edit and lint an existing `CHANGELOG.md`. That is a genuine widening of what the tool is — every other write path renders a fresh document and either creates a file or refuses to — and the conservatism of `internal/changelog` follows from it. `tasker` widens the catalog in the other direction, past the rule that got `capacity`, `retro` and `license` removed in 0.30.0: the spec admits a task card explicitly, by name, and nothing else moved with it. Extracted from the [gch](https://github.com/jtprogru/gch) monolith. Pre-1.0 (0.30.x line). `capacity`, `retro` and `license` were removed in 0.30.0; `cmd/retired.go` keeps hidden stubs that explain the removal until 1.0.

## Commands

```bash
make                 # list every target (help is the default goal)
make ci              # lint + race tests — the one-shot pre-push check
make test            # go test --short -coverprofile=cover.out -v ./...
make test-race       # go test -race -coverprofile=cover.out -v ./...  (what CI runs)
make lint            # golangci-lint run at the pinned version, from ./bin
make govulncheck     # vulnerability scan (what CI runs)
make build           # CGO_ENABLED=0 go build -o ./dist/srekit .
make run ARGS="…"    # go run . <args>
make release-dry     # goreleaser snapshot build into ./dist, no publish
make docs-serve      # MkDocs at http://127.0.0.1:8000
make docs-build      # mkdocs build --strict
```

Single test / subset:

```bash
go test ./cmd/ -run TestSLO -v
go test ./internal/sections/ -run TestRenderArtifact -v
go test ./cmd/ -run TestTemplates -v      # run this whole suite when touching cmd/templates.go — the 3-way merge has subtle invariants
```

Toolchain must match CI or gosec taint rules produce false positives. Go **1.26.4** is on you; `golangci-lint` (**v2.12.2**, ~50 linters including `gochecknoglobals`) and `govulncheck` are pinned in the `Makefile` and installed into `./bin` by their targets, so `make lint` resolves to the same binary locally and on a runner — invoking a system-wide `golangci-lint` directly bypasses that pin.

The `Makefile` must stay compatible with **GNU Make 3.81** — that is the system `make` on macOS and Apple will not ship newer (GPLv3). No `.ONESHELL` / `.SHELLFLAGS` (3.82+), no `!=` / `$(file ...)` (4.0+), no `$(intcmp)` / `$(let)` (4.4+); one command per recipe line. A 4.x runner swallows the incompatibility silently, so verify locally. Every CI workflow except `goreleaser.yaml` calls a Make target rather than a raw command — a target's behaviour is CI's behaviour.

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

- **`internal/tmpl`** — `Source` interface (`EmbedSource` for `//go:embed templates/*.yaml`, `DirSource{Dir}` for a user dir). `Loader{Sources}` walks them in order treating `fs.ErrNotExist` as fall-through, so a missing file in the user dir transparently falls back to embedded. Also the shared `template.FuncMap` (`default`, `shortID`, `slugify`, `upper`, `lower`, `trim`, `now`, `join`).
- **`internal/sections`** — the v1 artifact runtime. `Artifact` = `version` / `frontmatter` (a `yaml.Node`, so author key order survives parse→render) / `title` / `meta_bullets` / `header_body` / `sections`. Section types are `text` / `list` / `table`. A templated frontmatter scalar renders to a Go string, so it can only be emitted quoted; an explicit YAML tag from `retypedTags` (`!!int`, `!!seq`, …) makes the renderer re-read the rendered text as that type, and a mismatch is a render error naming the key. `!!str` and application tags are deliberately excluded — an explicit `!!str` is the untagged path, and somebody else's tag is theirs to interpret. `Merge` overlays per-section overrides; `RenderArtifact` composes the markdown (frontmatter block → H1 → meta bullets → header body → `## section` blocks).
- **`internal/render`** — `buildBody` has two branches: the `--json` short-circuit (`MarshalIndent` of an already-structured payload) and the v1 artifact path. 0.30.0 removed the legacy `text/template` branch with `--template FILE` and `license`, its last caller, and with it the `BootstrapJSON` envelope and `RenderOptionsStructured` — no generator had set the flag since 0.20.0. `writeBody` owns all `--out` / `--stdout` / `--force` / `--dry-run` / `--quiet` routing; `-` means stdout.
- **`internal/cliflags`** — `Output.Bind(cmd, outDesc)` wires the shared flags. There is no `--template` binding anywhere (see invariants).
- **`cmd/paths.go`** — XDG resolution for config and templates dir, with legacy-path precedence.
- **`internal/config`** — hand-rolled YAML+env config (`SREKIT_*` prefix). Deliberately not viper.
- **`internal/clock`** — `var Now = time.Now` indirection so tests can pin the wall clock (e.g. the Sunday on-call week-boundary test).
- **`internal/migrate`** — heuristic `.tmpl` → `.yaml` converter behind `srekit templates migrate`.
- **`internal/changelog`** — the only reader of a document the *user* wrote, behind `changelog release` / `changelog validate`. A line-oriented region scanner, deliberately not a Markdown parser and not a model-then-reserialize round trip: `Scan` records byte offsets (preamble, `[Unreleased]`, one region per version heading, the trailing link block) and `Release` splices at those offsets, copying everything else through untouched. Reserializing would normalize every blank line and bullet marker in a five-year-old changelog and make the release commit unreviewable, so byte-identical preservation outside the edited regions is a property of the design, not a test that happens to pass. Link conventions (host, path, URL shape, `v` prefix) are derived from the document's own `[Unreleased]` definition, never from git — git is consulted only when there is no block at all. The change-type vocabulary is a parameter (`Vocabulary`); `English` and `Russian` are both recognized, and which one is in force is detected from the document rather than from `--lang` — generation and parsing must not share a setting, or a `changelog_lang: ru` team would corrupt the English changelog they had before the switch. A document mixing both is refused by `release` and failed by `validate`.

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
- **Templates are bilingual** — headings and labels as `Русский (English)`, body in Russian. Technical identifiers (SLO/SLI/RFC/SEV/UTC/PromQL) and frontmatter keys stay English. `changelog` is English on every unqualified invocation so Keep a Changelog tooling doesn't break; `changelog.ru.yaml` is an opt-in variant behind `--lang ru` / `changelog_lang`, and inside it only the change types and prose are translated — `[Unreleased]`, version headings and link reference labels stay English because they are identifiers that must match the link definitions and the project's tags.
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

Release: `srekit changelog release --version X.Y.Z` (or move `[Unreleased]` into `[X.Y.Z] - YYYY-MM-DD` by hand), commit `release: X.Y.Z`, then `git tag -a vX.Y.Z && git push origin vX.Y.Z`. goreleaser builds linux/darwin/freebsd × amd64/arm64, GPG-signs checksums, and rewrites the `jtprogru/homebrew-tap` cask. `Version` / `Commit` / `Date` / `BuiltBy` in `cmd/root.go` are ldflags-injected.

**Check the tap token before tagging.** The cask push is the last step of the pipeline and the only one that authenticates with `HOMEBREW_TAP_GITHUB_TOKEN` rather than the workflow's built-in `GITHUB_TOKEN`, so an expired PAT fails *after* the GitHub release is already published — the release looks fine and the tap silently stays on the previous version. A fine-grained PAT expires on a timer; this has bitten 0.31.0. Run first, expecting `main`:

```bash
GH_TOKEN=<tap-pat> gh api repos/jtprogru/homebrew-tap --jq .default_branch
```

A `401 Bad credentials` means rotate the PAT (fine-grained, `jtprogru/homebrew-tap`, `Contents: Read and write`) and `gh secret set HOMEBREW_TAP_GITHUB_TOKEN --repo jtprogru/srekit` — the secret name is read by both `.github/workflows/goreleaser.yaml` and `.goreleaser.yaml`, so setting it under any other name changes nothing.

Recovering when it fails anyway: re-running the job does not work. The rerun replays the same ref, so a config fix landed on `main` is not picked up, and `replace_existing_artifacts` defaults to off, so re-uploading the existing assets fails with a 422. Fix the token, then delete the release and the tag and push the tag again at the same commit (`gh release delete vX.Y.Z --yes`, `git push origin :refs/tags/vX.Y.Z`, `git push origin vX.Y.Z`). The rebuild is not byte-identical — archive timestamps move the `sha256`s even at an unchanged commit — so anything that recorded the first run's checksums goes stale.

## OpenSpec

The repo uses a spec-driven workflow — `openspec/config.yaml` holds the authoritative project context and per-phase rules, `openspec/specs/` the current capability specs, `openspec/changes/` in-flight proposals. Skills exist under `.claude/skills/openspec-*`. When a task involves a spec change, read `openspec/config.yaml` first; its rules go further than this file (e.g. proposals must flag **BREAKING** public-contract changes, specs must be written in observable CLI behavior with no Go type names).
