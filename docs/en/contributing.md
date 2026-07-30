# Contributing

Thanks for considering a contribution. srekit is small, opinionated, and aims to stay that way — please open an issue before sending a PR that adds a new command, a new third-party dependency, or changes a flag's name.

## Local setup

```bash
git clone git@github.com:jtprogru/srekit.git
cd srekit
go mod download
task ci      # runs lint + race tests
```

Required:

- Go **1.25+**
- [Task](https://taskfile.dev/) (optional but recommended — see `Taskfile.yml` for the full list of targets)
- `golangci-lint` **v2.12** (must match CI; see [the version-skew lesson](#version-skew) below)
- For docs: Python 3 + `pip install -r docs/requirements.txt`

## Taskfile targets

| Task | What it does |
|---|---|
| `task run -- <args>` | `go run . <args>` |
| `task build` | Builds `./dist/srekit` |
| `task test` | `go test --short -coverprofile=cover.out -v ./...` |
| `task test:race` | `go test -race -v ./...` (what CI runs) |
| `task lint` | `golangci-lint -v run` |
| `task ci` | `lint` + `test:race` — one-shot pre-push check |
| `task release:dry` | `goreleaser release --clean --snapshot --skip=publish,sign` — builds into `./dist` without publishing |
| `task docs:install` | `pip install -r docs/requirements.txt` |
| `task docs:serve` | Serves the docs site at `http://127.0.0.1:8000` |
| `task docs:build` | Builds the docs into `./site` (strict mode) |
| `task fmt` | `gofmt -s -w .` |
| `task vet` | `go vet ./...` |
| `task tidy` | `go mod tidy` |

## Code style

- **No `Co-Authored-By` markers** in commits — write commits as yourself.
- Conventional-commit prefixes (`feat:`, `fix:`, `chore:`, `docs:`, `refactor:`, `test:`). Mostly for changelog scannability.
- Add a smoke test for every new generator command and a unit test for every new internal function. See `cmd/cmd_test.go` for the smoke pattern and `internal/*/` for unit tests.
- Don't suppress lint findings without a justification — use the `//nolint:<linter> // <code>: <reason>` format we already use for `gosec` G306.
- Keep public CLI flags small. Adding a new flag is a PR-worthy change.

## Pre-push checklist

```bash
task ci      # lint clean, race tests pass
task build   # builds without errors
```

If you're touching `cmd/templates.go`, run the full templates suite manually too — the 3-way merge has subtle invariants:

```bash
go test ./cmd/ -run TestTemplates -v
```

## Release process

(For maintainers.)

1. Make sure `main` is clean and CI is green.
2. Move `[Unreleased]` content in `CHANGELOG.md` into `[X.Y.Z] - YYYY-MM-DD`.
3. `git commit -m "release: X.Y.Z"` + `git push`.
4. `git tag -a vX.Y.Z -m vX.Y.Z` + `git push origin vX.Y.Z`.
5. Watch the `goreleaser` workflow — should take ~90 seconds.
6. Verify the GitHub Release page and `jtprogru/homebrew-tap` cask are updated.

`task release:dry` runs the build locally without publishing — useful for catching `.goreleaser.yaml` issues before tagging.

## Known caveats

### Version skew

CI pins `golangci-lint v2.12`. If you run an older version locally, gosec rules `G703` (path traversal taint) and `G705` (XSS taint) may fire false positives on the config init code path. Match local to CI:

```bash
brew install golangci-lint
golangci-lint version    # should be 2.12.x
```

### Global state in tests

The template loader is built per command tree in `configureTemplates` and threaded to each command through the cobra command context (`loaderFrom`), so `--templates-dir` tests are `t.Parallel()`-safe. Don't reintroduce a package-level loader — `gochecknoglobals` is enabled and the earlier global (`tmpl.Default`) was removed precisely because it tripped the race detector.

The one piece of global state left is the config singleton behind `config.Global()`. Tests that seed it use the `withConfig(t, kv)` helper in `cmd/cmd_test.go`, which resets the config and registers cleanup; those tests must not call `t.Parallel()`.

## See also

- [Architecture](architecture.md) — code map and key abstractions.
- [GitHub issues](https://github.com/jtprogru/srekit/issues) — open one before non-trivial work.
