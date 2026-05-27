# Architecture

A tour of the code for contributors and curious users. Pin to
[v0.10.1](https://github.com/jtprogru/srekit/tree/v0.10.1) — the
structure is stable as of this version.

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
│   ├── cliflags/         # shared --out / --stdout / --force / --dry-run / --template / --json bundle
│   ├── tmpl/             # //go:embed templates/*.tmpl + Funcs + Source/Loader + Samples + Validate + DocsMD
│   │   └── templates/    # the embedded SRE templates
│   └── render/           # Render() — buildBody/writeBody + JSON short-circuit
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

Templates can come from the binary (embedded) or from a directory the
user controls. `Source` is the interface:

```go
type Source interface {
    Read(name string) ([]byte, error)
}
```

`EmbedSource` reads from `//go:embed templates/*.tmpl`. `DirSource{Dir}`
reads from a path on disk. `Loader{Sources}` walks them in order with
`fs.ErrNotExist`-as-fallthrough semantics — so a missing file in the
user dir transparently falls back to the embedded one.

Production code uses a package-level `tmpl.Default` (an `EmbedSource`
by default; replaced with `Loader{DirSource, EmbedSource}` when a
templates dir is configured). This is the package-level mutable state
that tests work around via the `resetTmplDefault(t)` helper.

### `internal/render.Render()`

The shared rendering pipeline. Takes the template name, the data
struct, and `render.Options{Out, Stdout, Force, DryRun, TemplatePath,
JSON, Default}`. Calls `buildBody()` (which short-circuits to JSON
when `Options.JSON` is set), then `writeBody()` (which decides between
stdout and file based on flags and the `Default` filename).

### `internal/cliflags.Output`

Every generator command embeds an `Output` and calls
`.Bind(cmd, "default-path-description")`. This is what gives them all
the same flag set without per-command boilerplate. `RenderOptions(def)`
turns the flag values into a `render.Options`.

### `internal/meta.Resolve` and `DetectRepo`

`Resolve` walks flag → viper → git config for `author` and `email`.
`DetectRepo` regex-parses `git config remote.origin.url` against GitHub
SSH and HTTPS patterns.

### `internal/clock.Now`

A `var Now func() time.Time = time.Now` indirection. Lets tests pin
the wall clock (e.g. the Sunday on-call week boundary regression test).

### Template snapshots: `.srekit-embedded/`

The 3-way merge in `templates upgrade` uses a per-template snapshot of
"the embedded content as of the last sync" as the merge base. The
sidecar lives at `<user-templates-dir>/.srekit-embedded/<name>` and is
appended to the user dir's `.gitignore` so it never pollutes their
templates repo. See `cmd/templates.go` — `snapshotPath`, `readSnapshot`,
`writeSnapshot`, `ensureSnapshotIgnored`, `threeWayMerge`.

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
| Unit | `ids.UUID`/`Slug`, `meta.Resolve`/`DetectRepo`, `tmpl.Funcs`/`Loader`/`Samples`/`Validate`, `cliflags`, `render` (file/stdout/dry-run/JSON) | `internal/*/*_test.go` |
| Integration | Smoke through `cobra.Command.SetArgs` + captured stdout for every command, including templates pull/validate/diff/upgrade/list and config init | `cmd/cmd_test.go` |
| Race | `go test -race ./...` on CI | `.github/workflows/tests.yaml` |
| Lint | `golangci-lint v2.12` with ~50 linters | `.golangci.yaml`, `.github/workflows/lint.yaml` |

## Things you'd touch by feature

| If you want to... | Start here |
|---|---|
| Add a new SRE artifact | `cmd/<name>.go` (copy an existing generator) + `internal/tmpl/templates/<name>.md.tmpl` + sample data in `internal/tmpl/samples.go` |
| Add a flag to an existing generator | The relevant `cmd/<name>.go`; for shared output flags, `internal/cliflags/cliflags.go` |
| Change a template's content | Just edit `internal/tmpl/templates/<name>.md.tmpl` |
| Add a template helper function | `internal/tmpl/funcs.go` |
| Modify the templates lifecycle | `cmd/templates.go` |

## See also

- [Contributing](contributing.md) — local dev setup, Taskfile, release
  process.
- [GitHub source](https://github.com/jtprogru/srekit) — read the code
  directly; it's small.
