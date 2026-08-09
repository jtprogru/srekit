# Architecture

A tour of the code for contributors and curious users.

## Package layout

```
srekit/
├── main.go               # cobra.Execute() entry point
├── cmd/                  # one .go per command + shared root
│   ├── root.go           # cobra root + persistent --templates-dir / --config / --quiet
│   ├── paths.go          # XDG resolution for config + templates dir, legacy paths win if present
│   ├── doctor*.go        # read-only environment report
│   ├── <command>.go      # one per generator + the templates / config / changelog groups
│   └── cmd_test.go       # smoke tests through cobra (SetArgs + captured stdout)
├── internal/
│   ├── ids/              # UUID v4 + slug
│   ├── clock/            # var Now = time.Now (overridable in tests)
│   ├── meta/             # Author.Resolve + DetectRepo (git remote parsing)
│   ├── config/           # hand-rolled YAML + SREKIT_* env config (deliberately not viper)
│   ├── cliflags/         # shared --out / --stdout / --force / --dry-run / --json bundle
│   ├── tmpl/             # //go:embed templates/*.yaml + Funcs + Source/Loader + Samples + DocsMD
│   │   └── templates/    # v1 single-file YAML artifacts, one per generator
│   ├── sections/         # Artifact (v1 single-file) + Section/Manifest + Merge + RenderArtifact + JSONSchema
│   ├── changelog/        # `changelog release` / `validate` — offset-based rewriter for a doc the user wrote
│   ├── migrate/          # `srekit templates migrate` — heuristic .tmpl → .yaml converter
│   └── render/           # Render() — buildBody/writeBody + JSON short-circuit + artifact branch
├── docs/                 # this site (MkDocs Material with i18n)
├── .github/
│   ├── workflows/        # tests / lint / goreleaser / docs
│   └── dependabot.yml
├── .goreleaser.yaml
├── .golangci.yaml
└── Makefile
```

## Key abstractions

### `internal/tmpl.Source` and `Loader`

Templates can come from the binary (embedded) or from a directory the user controls. `Source` is the interface:

```go
type Source interface {
    Read(name string) ([]byte, error)
}
```

`EmbedSource` reads from `//go:embed templates/*.yaml`. `DirSource{Dir}` reads from a path on disk. `Loader{Sources}` walks them in order with `fs.ErrNotExist`-as-fallthrough semantics — so a missing file in the user dir transparently falls back to the embedded one.

Each generator command builds a `*tmpl.Loader` via `configureTemplates` and stashes it in `cmd.Context()`; downstream code reads it back with `loaderFrom(cmd)`. No package-level mutable state — the loader is scoped per command tree, which keeps `--templates-dir` tests parallel-safe.

### `internal/sections`

The v1 artifact runtime. `Artifact` is the parsed `<name>.yaml`: frontmatter (`yaml.Node` for order preservation), title, meta_bullets, header_body, a section list, and footer_body. `ParseArtifact` validates structural invariants; `Merge` overlays per-section overrides and template-evaluates section titles; `RenderArtifact` composes the markdown (frontmatter block → H1 → meta_bullets → header_body → `## section` blocks → footer_body), opening every block through a single helper that guarantees exactly one blank line between adjacent blocks.

Generator commands implement `ArtifactPayload()` on their data struct to hand the merged section list + ctx back to `RenderArtifact`.

### `internal/render.Render()`

The shared rendering pipeline. Takes a name, the data struct, and `render.Options{Out, Stdout, Force, DryRun, Default, JSON, Quiet}`. Two branches:

1. `--json` short-circuit: `MarshalIndent` the data, which the caller has already shaped as `{meta, sections}`.
2. Otherwise the artifact path: the name is the bare artifact name (`"slo"`), so load `slo.yaml`, parse, hand off to `sections.RenderArtifact`. `tmpl.ArtifactNameFor` normalizes the name; it is idempotent and still accepts the pre-v1.0 spellings (`slo.md.tmpl`, `slo.tmpl`).

v0.30.0 removed the third branch — Go-template execution — along with `--template FILE` and `license`, its last caller. The `BootstrapJSON` envelope that wrapped rendered markdown into a synthetic `{meta, sections}` payload went with it: no generator had set it since v0.20.0.

### `internal/cliflags.Output`

Every generator command embeds an `Output` and calls `.Bind(cmd, "default-path-description")`. This wires the shared flags (`--out` / `--stdout` / `--force` / `--dry-run` / `--json`). There is no `--template FILE` binding: every generator resolves its artifact by name, so the flag would be silently ignored. `RenderOptions(def)` turns the flag values into a `render.Options`.

### `internal/config`

A hand-rolled YAML + `SREKIT_*` environment reader, deliberately not viper. viper was dropped because it pulls `afero → net/http → crypto/tls` into the build graph, and neither `net/http` nor `crypto` may appear there — see the dependency-minimalism invariant. `config.Global()` is the one remaining piece of package-level mutable state; tests seed it through the non-parallel `withConfig(t, kv)` helper.

### `internal/changelog`

The only package that reads a document the *user* wrote, behind `changelog release` / `changelog validate`. A line-oriented region scanner, not a Markdown parser: `Scan` records byte offsets (preamble, `[Unreleased]`, one region per version heading, the trailing link block) and `Release` splices at those offsets, copying everything else through untouched. Reserializing a parsed model would normalize every blank line and bullet marker in a five-year-old changelog and make the release commit unreviewable, so byte-identical preservation outside the edited regions is a property of the design rather than a test that happens to pass. Link conventions come from the document's own `[Unreleased]` definition; git is consulted only when there is no link block at all. The change-type vocabulary is a parameter (`English` / `Russian`) detected from the document, never from `--lang`.

### `internal/meta.Resolve` and `DetectRepo`

`Resolve` walks flag → `SREKIT_*` env → config file → `git config` for `author` and `email`. `DetectRepo` regex-parses `git config remote.origin.url` against GitHub SSH and HTTPS patterns.

### `internal/clock.Now`

A `var Now func() time.Time = time.Now` indirection. Lets tests pin the wall clock (e.g. the Sunday on-call week boundary regression test).

### Template snapshots: `.srekit-embedded/`

The 3-way merge in `templates upgrade` uses a per-template snapshot of "the embedded content as of the last sync" as the merge base. The sidecar lives at `<user-templates-dir>/.srekit-embedded/<name>` and is appended to the user dir's `.gitignore` so it never pollutes their templates repo. See `cmd/templates.go` — `snapshotPath`, `readSnapshot`, `writeSnapshot`, `ensureSnapshotIgnored`, `threeWayMerge`.

## Release pipeline

| Tool | Purpose |
|---|---|
| `goreleaser` | Builds linux/darwin/freebsd × arm64/x86_64 binaries; signs checksums with GPG; updates the homebrew tap. |
| GitHub Actions `goreleaser.yaml` | Triggers on tag push (`v*`), imports the GPG key, runs goreleaser. |
| `crazy-max/ghaction-import-gpg@v7` | Imports the signing key. |
| `HOMEBREW_TAP_GITHUB_TOKEN` secret | Fine-grained PAT with `Contents:read+write` on `jtprogru/homebrew-tap`. |

The release flow:

1. Cut `CHANGELOG.md` — `srekit changelog release --version X.Y.Z`, or move `[Unreleased]` into `[X.Y.Z]` by hand.
2. Commit `release: X.Y.Z` on `main`.
3. `git tag -a vX.Y.Z -m vX.Y.Z` and `git push origin vX.Y.Z`.
4. goreleaser builds 6 archives (3 OS × 2 arch) + checksums + GPG sig.
5. Homebrew cask in `jtprogru/homebrew-tap/Casks/srekit.rb` is rewritten.

Step 5 is the only one that authenticates with `HOMEBREW_TAP_GITHUB_TOKEN` rather than the workflow's built-in `GITHUB_TOKEN`, so an expired PAT fails *after* the GitHub release is already published — the release looks fine and the tap silently stays on the previous version. Check the token before tagging:

```bash
GH_TOKEN=<tap-pat> gh api repos/jtprogru/homebrew-tap --jq .default_branch
```

## Testing strategy

| Layer | What's tested | File |
|---|---|---|
| Unit | `ids.UUID`/`Slug`, `meta.Resolve`/`DetectRepo`, `tmpl.Funcs`/`Loader`, `sections.*`, `cliflags`, `render` (file/stdout/dry-run/JSON/artifact) | `internal/*/*_test.go` |
| Integration | Smoke through `cobra.Command.SetArgs` + captured stdout for every command, including templates pull/validate/diff/upgrade/list/migrate and config init | `cmd/cmd_test.go` |
| Race | `make test-race` on CI | `.github/workflows/tests.yaml` |
| Lint | `golangci-lint v2.12.2` with ~50 linters, via `make lint` | `.golangci.yaml`, `.github/workflows/lint.yaml` |
| Vulnerabilities | `make govulncheck` | `.github/workflows/security.yaml` |

Render/tmpl unit tests build their loader via a `newFixtureLoader(t)` helper that writes a per-test `fixture.yaml` artifact into a temp dir — they don't depend on what's currently in embed, which kept the test suite stable across the v0.14–v0.20 migration churn.

## Things you'd touch by feature

| If you want to... | Start here |
|---|---|
| Add a new SRE artifact | `cmd/<name>.go` (copy an existing generator) + `internal/tmpl/templates/<name>.yaml` (v1 artifact) |
| Add a flag to an existing generator | The relevant `cmd/<name>.go`; for shared output flags, `internal/cliflags/cliflags.go` |
| Change a template's content | Edit `internal/tmpl/templates/<name>.yaml` |
| Tweak rendered markdown layout (frontmatter, H1, section composition) | `internal/sections/render_artifact.go` |
| Add a template helper function | `tmpl.Funcs` in `internal/tmpl/tmpl.go` |
| Modify the templates lifecycle | `cmd/templates.go` |
| Change how an existing changelog is scanned or rewritten | `internal/changelog/` (`scan.go`, `release.go`, `validate.go`, `links.go`) |
| Add or change a `doctor` check | `cmd/doctor_checks.go` — check IDs are a public contract |
| Change where config or templates resolve | `cmd/paths.go` |

## See also

- [Contributing](contributing.md) — local dev setup, Make targets, release process.
- [GitHub source](https://github.com/jtprogru/srekit) — read the code directly; it's small.
