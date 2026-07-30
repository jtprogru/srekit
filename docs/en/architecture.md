# Architecture

A tour of the code for contributors and curious users.

## Package layout

```
srekit/
├── main.go               # cobra.Execute() entry point
├── cmd/                  # one .go per command + shared root
│   ├── root.go           # cobra root + viper init + persistent --templates-dir / --config
│   ├── <command>.go      # one per generator + the templates / config groups
│   └── cmd_test.go       # smoke tests through cobra (SetArgs + captured stdout)
├── internal/
│   ├── ids/              # UUID v4 + slug
│   ├── clock/            # var Now = time.Now (overridable in tests)
│   ├── meta/             # Author.Resolve + DetectRepo (git remote parsing)
│   ├── cliflags/         # shared --out / --stdout / --force / --dry-run / --json bundle (--template via BindTemplateFlag, license only)
│   ├── tmpl/             # //go:embed templates/*.yaml + Funcs + Source/Loader + Samples + Validate + DocsMD
│   │   └── templates/    # v1 single-file YAML artifacts, one per generator
│   ├── sections/         # Artifact (v1 single-file) + Section/Manifest + Merge + RenderArtifact + JSONSchema
│   ├── migrate/          # `srekit templates migrate` — heuristic .tmpl → .yaml converter
│   └── render/           # Render() — buildBody/writeBody + JSON short-circuit + artifact branch
├── docs/                 # this site (MkDocs Material with i18n)
├── .github/
│   ├── workflows/        # tests / lint / goreleaser / docs
│   └── dependabot.yml
├── .goreleaser.yaml
├── .golangci.yaml
└── Taskfile.yml
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

The v1 artifact runtime. `Artifact` is the parsed `<name>.yaml`: frontmatter (`yaml.Node` for order preservation), title, meta_bullets, header_body, and a section list. `ParseArtifact` validates structural invariants; `Merge` overlays per-section overrides and template-evaluates section titles; `RenderArtifact` composes the markdown (frontmatter block → H1 → meta_bullets → header_body → `## section` blocks).

Generator commands pass `Options.RenderArtifact = true` and implement `ArtifactPayload()` on their data struct to hand the merged section list + ctx back to `RenderArtifact`.

### `internal/render.Render()`

The shared rendering pipeline. Takes a name, the data struct, and `render.Options{Out, Stdout, Force, DryRun, TemplatePath, JSON, Default, BootstrapJSON, RenderArtifact}`. Three branches:

1. `--json` short-circuit: `MarshalIndent` the data (structured pass-through).
2. `RenderArtifact = true`: the name is the bare artifact name (`"slo"`), so load `slo.yaml`, parse, hand off to `sections.RenderArtifact`. `tmpl.ArtifactNameFor` normalizes the name; it is idempotent and still accepts the pre-v1.0 spellings (`slo.md.tmpl`, `slo.tmpl`).
3. Otherwise: legacy `text/template` execution. No shipped generator uses this branch as of v0.20; it's kept for `--template FILE` on `license` and for external tooling.

### `internal/cliflags.Output`

Every generator command embeds an `Output` and calls `.Bind(cmd, "default-path-description")`. This wires the shared flags (`--out` / `--stdout` / `--force` / `--dry-run` / `--json`). The `--template FILE` flag is opt-in via `BindTemplateFlag(cmd)` — only `cmd/license.go` calls it, since every other generator goes through the artifact path and ignores `TemplatePath`. `RenderOptions(def)` turns the flag values into a `render.Options`; `RenderOptionsStructured(def)` clears `BootstrapJSON` for artifact-path commands.

### `internal/meta.Resolve` and `DetectRepo`

`Resolve` walks flag → viper → git config for `author` and `email`. `DetectRepo` regex-parses `git config remote.origin.url` against GitHub SSH and HTTPS patterns.

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

1. Bump `CHANGELOG.md` — move `[Unreleased]` content into `[X.Y.Z]`.
2. Commit `release: X.Y.Z` on `main`.
3. `git tag -a vX.Y.Z -m vX.Y.Z` and `git push origin vX.Y.Z`.
4. goreleaser builds 8 artifacts + checksums + GPG sig.
5. Homebrew cask in `jtprogru/homebrew-tap/Casks/srekit.rb` is rewritten.

## Testing strategy

| Layer | What's tested | File |
|---|---|---|
| Unit | `ids.UUID`/`Slug`, `meta.Resolve`/`DetectRepo`, `tmpl.Funcs`/`Loader`, `sections.*`, `cliflags`, `render` (file/stdout/dry-run/JSON/artifact) | `internal/*/*_test.go` |
| Integration | Smoke through `cobra.Command.SetArgs` + captured stdout for every command, including templates pull/validate/diff/upgrade/list/migrate and config init | `cmd/cmd_test.go` |
| Race | `go test -race ./...` on CI | `.github/workflows/tests.yaml` |
| Lint | `golangci-lint v2.12` with ~50 linters | `.golangci.yaml`, `.github/workflows/lint.yaml` |

Render/tmpl unit tests build their loader via a `newFixtureLoader(t)` helper that writes a per-test `.tmpl` fixture into a temp dir — they don't depend on what's currently in embed, which kept the test suite stable across the v0.14–v0.20 migration churn.

## Things you'd touch by feature

| If you want to... | Start here |
|---|---|
| Add a new SRE artifact | `cmd/<name>.go` (copy an existing generator) + `internal/tmpl/templates/<name>.yaml` (v1 artifact) |
| Add a flag to an existing generator | The relevant `cmd/<name>.go`; for shared output flags, `internal/cliflags/cliflags.go` |
| Change a template's content | Edit `internal/tmpl/templates/<name>.yaml` |
| Tweak rendered markdown layout (frontmatter, H1, section composition) | `internal/sections/render_artifact.go` |
| Add a template helper function | `internal/tmpl/funcs.go` |
| Modify the templates lifecycle | `cmd/templates.go` |

## See also

- [Contributing](contributing.md) — local dev setup, Taskfile, release process.
- [GitHub source](https://github.com/jtprogru/srekit) — read the code directly; it's small.
