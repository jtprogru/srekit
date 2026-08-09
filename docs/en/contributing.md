# Contributing

Thanks for considering a contribution. srekit is small, opinionated, and aims to stay that way — please open an issue before sending a PR that adds a new command, a new third-party dependency, or changes a flag's name.

## Local setup

```bash
git clone git@github.com:jtprogru/srekit.git
cd srekit
make ci      # runs lint + race tests
```

Required:

- Go **1.26.4** (must match CI — see [the version-skew lesson](#version-skew) below)
- GNU Make **3.81+** — the system `make` on macOS and every Linux distro qualifies, nothing to install
- git, bash, and the usual POSIX utilities

Everything else is installed by the Makefile on first use, at a version pinned in the Makefile itself:

- `golangci-lint` and `govulncheck` → `./bin` (via `go install`, which does not touch `go.mod`)
- MkDocs and its plugins → `./.venv` (needs Python 3 with `venv`)

Optional: `goreleaser` for `make release-dry`, `tokei` for the LoC-badge pre-commit hook.

## Make targets

`make` on its own prints this list. CI calls these same targets, so a green `make ci` means the same thing locally and on a runner.

| Target | What it does |
|---|---|
| `make run ARGS="<args>"` | `go run . <args>` |
| `make build` | Builds `./dist/srekit` |
| `make install` | `go install` |
| `make test` | `go test --short -coverprofile=cover.out -v ./...` |
| `make test-race` | `go test -race -coverprofile=cover.out -v ./...` (what CI runs) |
| `make lint` | `golangci-lint run` at the pinned version |
| `make lint-fix` | Same, with `--fix` |
| `make govulncheck` | Vulnerability scan (what CI runs) |
| `make ci` | `lint` + `test-race` — one-shot pre-push check |
| `make ci-full` | `lint` + `test-race` + `govulncheck` + `docs-build` |
| `make release-check` | `goreleaser check` — validates `.goreleaser.yaml` |
| `make release-dry` | `goreleaser release --clean --snapshot --skip=publish,sign` — builds into `./dist` without publishing |
| `make docs-install` | Creates `./.venv` and installs `docs/requirements.txt` |
| `make docs-serve` | Serves the docs site at `http://127.0.0.1:8000` |
| `make docs-build` | Builds the docs into `./site` (strict mode) |
| `make docs-deploy` | `mkdocs gh-deploy` — called from CI, not by hand |
| `make tools` | Installs the pinned `golangci-lint` and `govulncheck` into `./bin` |
| `make fmt` | `gofmt -s -w .` |
| `make vet` | `go vet ./...` |
| `make tidy` | `go mod tidy` |
| `make clean` | Removes `dist/`, `site/`, `bin/`, `cover.out` |

The Makefile is written for GNU Make **3.81** — that is the system `make` on macOS and Apple will not ship a newer one. When editing it, avoid `.ONESHELL` and `.SHELLFLAGS` (3.82+), `!=` and `$(file ...)` (4.0+), `$(intcmp)` and `$(let)` (4.4+), and verify on 3.81 before pushing: a 4.x runner will swallow the incompatibility silently.

## Code style

- **No `Co-Authored-By` markers** in commits — write commits as yourself.
- Conventional-commit prefixes (`feat:`, `fix:`, `chore:`, `docs:`, `refactor:`, `test:`). Mostly for changelog scannability.
- Add a smoke test for every new generator command and a unit test for every new internal function. See `cmd/cmd_test.go` for the smoke pattern and `internal/*/` for unit tests.
- Don't suppress lint findings without a justification — use the `//nolint:<linter> // <code>: <reason>` format we already use for `gosec` G306.
- Keep public CLI flags small. Adding a new flag is a PR-worthy change.

## Pre-push checklist

```bash
make ci      # lint clean, race tests pass
make build   # builds without errors
```

If you're touching `cmd/templates.go`, run the full templates suite manually too — the 3-way merge has subtle invariants:

```bash
go test ./cmd/ -run TestTemplates -v
```

## Release process

(For maintainers.)

1. Make sure `main` is clean and CI is green.
2. Check the tap token — it is the only credential the pipeline uses that is not the workflow's built-in `GITHUB_TOKEN`, and it fails *after* the release is published:

    ```bash
    GH_TOKEN=<tap-pat> gh api repos/jtprogru/homebrew-tap --jq .default_branch   # expect: main
    ```

    A `401 Bad credentials` means rotating the fine-grained PAT (`jtprogru/homebrew-tap`, `Contents: Read and write`) and `gh secret set HOMEBREW_TAP_GITHUB_TOKEN --repo jtprogru/srekit`.

3. Cut the changelog: `srekit changelog release --version X.Y.Z`, or move `[Unreleased]` into `[X.Y.Z] - YYYY-MM-DD` by hand.
4. `git commit -m "release: X.Y.Z"` + `git push`.
5. `git tag -a vX.Y.Z -m vX.Y.Z` + `git push origin vX.Y.Z`.
6. Watch the `goreleaser` workflow — should take ~90 seconds.
7. Verify the GitHub Release page and `jtprogru/homebrew-tap` cask are updated.

If the cask push fails anyway, re-running the job does not help: the rerun replays the same ref, so a fix landed on `main` is not picked up, and re-uploading existing assets fails with a 422. Fix the token, then delete the release and the tag and push the tag again at the same commit (`gh release delete vX.Y.Z --yes`, `git push origin :refs/tags/vX.Y.Z`, `git push origin vX.Y.Z`). The rebuild is not byte-identical — archive timestamps move the checksums — so anything that recorded the first run's `sha256`s goes stale.

`make release-dry` runs the build locally without publishing — useful for catching `.goreleaser.yaml` issues before tagging.

## Known caveats

### Version skew

`golangci-lint` is pinned in the Makefile (`GOLANGCI_LINT_VERSION`) and installed into `./bin`, so `make lint` resolves to the same binary locally and in CI. Running a system-wide `golangci-lint` directly bypasses that pin — on an older version, gosec rules `G703` (path traversal taint) and `G705` (XSS taint) fire false positives on the config init code path. Use `make lint`.

The Go toolchain is not pinned by the Makefile — match `go.mod` and the `GO_VERSION` in the workflows.

### Global state in tests

The template loader is built per command tree in `configureTemplates` and threaded to each command through the cobra command context (`loaderFrom`), so `--templates-dir` tests are `t.Parallel()`-safe. Don't reintroduce a package-level loader — `gochecknoglobals` is enabled and the earlier global (`tmpl.Default`) was removed precisely because it tripped the race detector.

The one piece of global state left is the config singleton behind `config.Global()`. Tests that seed it use the `withConfig(t, kv)` helper in `cmd/cmd_test.go`, which resets the config and registers cleanup; those tests must not call `t.Parallel()`.

## See also

- [Architecture](architecture.md) — code map and key abstractions.
- [GitHub issues](https://github.com/jtprogru/srekit/issues) — open one before non-trivial work.
