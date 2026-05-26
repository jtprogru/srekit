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

[Unreleased]: https://github.com/jtprogru/srekit/compare/v0.3.0...HEAD
[0.3.0]: https://github.com/jtprogru/srekit/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/jtprogru/srekit/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/jtprogru/srekit/releases/tag/v0.1.0
