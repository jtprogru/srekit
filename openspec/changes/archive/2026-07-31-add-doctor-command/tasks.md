## 1. Command skeleton

- [x] 1.1 Add `cmd/doctor.go` with a `doctor` command taking no positional arguments, binding only `--json`, and inheriting the root's `--config`, `--templates-dir` and `--quiet`
- [x] 1.2 Register `doctor` in `NewRootCmd()` in `cmd/root.go`
- [x] 1.3 Define the finding shape (identifier, category, status, summary, remedy) and the check descriptor slice built per invocation, with no package-level state
- [x] 1.4 Implement the collect-then-render driver: run every check, never break early, map an inspection failure to an `error` finding rather than an error return
- [x] 1.5 Wire the exit status — `1` when any finding is `error`, `0` otherwise

## 2. Configuration and path checks

- [x] 2.1 Report the config file that will actually be read, naming whether it is the XDG or the legacy location, and whether it exists
- [x] 2.2 Report the config parse result, surfacing the load error that startup deliberately discards, as `warn`
- [x] 2.3 Report the resolved templates directory, the source that supplied it, and whether the path exists and is a directory; `warn` when configured but unusable
- [x] 2.4 Report every `SREKIT_`-prefixed environment variable in effect, by name
- [x] 2.5 Report whether the directory holding the resolved config path is writable
- [x] 2.6 Add the shadowing check: `warn` when both legacy and XDG config files exist, and when both legacy and XDG templates directories exist, naming the winner and the ignored path
- [x] 2.7 Suppress `configureTemplates`'s stderr fallback warning for `doctor` so a missing templates directory is reported once, as a finding

## 3. Identity check

- [x] 3.1 Report the resolved author name and email and the source each came from, using the same resolver the generators call
- [x] 3.2 Report `error` when neither name nor email resolves, with a remedy naming `srekit config init`, the `--author` / `--email` flags, and the git config keys

## 4. Template health checks

- [x] 4.1 Report artifact parse health by reusing `templates validate`'s per-file dispatch; `error` naming each failing file and its parse error
- [x] 4.2 Report legacy pre-v1.0 template files present in the directory as `warn`, with `srekit templates migrate` as the remedy
- [x] 4.3 Report drift from the embedded set by reusing `templates list`'s classification; `warn` with `srekit templates upgrade` and `srekit templates diff` as remedies
- [x] 4.4 Report `ok` with an embedded-only summary when no templates directory is configured

## 5. External dependency check

- [x] 5.1 Report `git`'s resolved path and version, or `warn` with the fallback consequences when it is absent from `PATH`
- [x] 5.2 Run `git --version` through the context-aware subprocess pattern so Ctrl-C tears it down

## 6. Output

- [x] 6.1 Render findings as an aligned table grouped by category, with remedies on continuation lines and a trailing per-status count line
- [x] 6.2 Suppress colour when `NO_COLOR` is set and non-empty, matching `templates diff`
- [x] 6.3 Render `--json` as an indented, newline-terminated document with `camelCase` keys, carrying the overall status and the `checks` array, and nothing else on stdout
- [x] 6.4 Make `--quiet` filter `ok` findings and the summary line from text output, leave the exit status unchanged, and leave `--json` output complete

## 7. Tests

- [x] 7.1 Smoke test in `cmd/cmd_test.go`: `doctor` on an isolated environment with temp config and templates paths reports every finding `ok` and exits `0`
- [x] 7.2 Test the exit status contract — `warn`-only exits `0`, any `error` exits `1`
- [x] 7.3 Test the shadowing finding with both config locations present, asserting the reported winner matches the file a generator actually reads
- [x] 7.4 Test the unparseable-artifact path (`error`) and the legacy-file path (`warn`) against a fixture templates directory
- [x] 7.5 Test the absent-`git` branch by pointing `PATH` at an empty directory
- [x] 7.6 Test `--json`: valid JSON, `camelCase` keys, overall status equals the worst finding, and `--json --quiet` still carries `ok` findings
- [x] 7.7 Test that a `doctor` run creates no file or directory
- [x] 7.8 Keep the config-dependent tests on the established non-parallel config helper; everything else runs with `t.Parallel()`

## 8. Documentation and changelog

- [x] 8.1 Add `docs/en/commands/doctor.md` documenting every check identifier, the status meanings, the exit status contract, and the JSON shape
- [x] 8.2 Add the mirrored `docs/ru/commands/doctor.md`
- [x] 8.3 Add both pages to the `mkdocs.yml` nav for both locales and cross-link from the commands index pages
- [x] 8.4 Add a `CHANGELOG.md` entry under `[Unreleased]` → `Added`

## 9. Verification

- [x] 9.1 Run `task ci` — `golangci-lint run` clean and `go test -race ./...` green
- [x] 9.2 Run `task docs:build` (`mkdocs build --strict`) to confirm the new nav entries resolve in both locales
- [x] 9.3 Confirm `go list` shows neither `net/http` nor `crypto/tls` entering the build graph
