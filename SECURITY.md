# Security policy

## Reporting a vulnerability

If you believe you've found a security vulnerability in `srekit`, please **do not** open a public GitHub issue. Instead:

- Email **jtprogru@gmail.com** with `[srekit security]` in the subject line, or
- Open a [private security advisory on GitHub](https://github.com/jtprogru/srekit/security/advisories/new).

Please include:

- A description of the issue and its potential impact.
- Reproduction steps (binary version, OS, exact commands).
- Any proof-of-concept or relevant logs.

You'll get an acknowledgement within ~7 days. Fix turnaround depends on severity; expect coordinated disclosure once a patched release is out.

## Supported versions

`srekit` is in the `0.x` line. Security fixes are issued only on the **latest minor release** (`brew upgrade srekit` / `go install …@latest`). There are no patch backports to older minors during 0.x.

When `1.0.0` ships, this policy will be extended with an explicit support window for the current major.

## Scope

In scope:

- The `srekit` CLI binary built from this repository (`main` branch, latest tagged release).
- The embedded artifact rendering pipeline (`internal/sections`, `internal/render`, `internal/tmpl`, `internal/migrate`).
- The release-pipeline artifacts published to GitHub Releases and `jtprogru/homebrew-tap`.

Out of scope:

- User-controlled inputs that are by design path-or-text inputs to the CLI (`--templates-dir`, `--template FILE`, `--from FILE`, contents of `$TEMPLATES_DIR/<name>.yaml`). `srekit` runs in the user's trust context; a user-supplied template is treated as user-trusted code.
- The Go template engine itself (`text/template`), `go.yaml.in/yaml/v3`, and other upstream dependencies — please report to those projects. We track them via `govulncheck` and update.
- Markdown / YAML rendered by `srekit` and consumed by downstream tools (your wiki, your tracker, your LLM). `srekit` does not sanitize output for any particular consumer; downstream tools must handle untrusted markdown / YAML themselves.

## What `srekit` does to keep itself safe

- **`govulncheck`** runs on every push / PR via `.github/workflows/security.yaml`. Build fails on any reachable vulnerability in a direct or transitive dependency.
- **`gosec`** runs inside `golangci-lint` (which executes on every push / PR via `.github/workflows/lint.yaml`). All `gosec` findings either reflect real issues or are explicit `//nolint:gosec` annotations with a rationale.
- **Bearer** (informational) runs on every push / PR and uploads SARIF to GitHub's code scanning dashboard.
- **Release artifacts** are built by `goreleaser` in a clean GitHub Actions runner and shipped with SHA256 checksums; the homebrew cask verifies the checksum on install.
- **Public file permissions are intentional**: scaffolded templates and rendered documents are written `0o644` because they are public artifacts (your `CHANGELOG.md`, your team's `sre-templates` repo). The config file (`config.yaml`) is `0o600` because it may contain author identity / email.

## Threat model — what `srekit` is NOT

`srekit` is a single-user document generator, not a multi-tenant service:

- There is no network listener; the only network calls come from `srekit templates pull` shelling out to `git`.
- There is no authentication / authorization layer — `srekit` runs with the invoking user's privileges.
- Inputs (`--from FILE`, `--template FILE`, custom `<name>.yaml` in `templates_dir`) are user-supplied; `srekit` evaluates Go templates over them. **A malicious `<name>.yaml` can do whatever the FuncMap allows** (currently: `default`, `slugify`, `now`, `shortID`, `upper`, `lower`, `trim` — all pure, no I/O, no `exec`). Adding I/O- or exec-capable functions to the FuncMap is a deliberate security-relevant change and would require a major-version bump per the [Stability policy in README](README.md#стабильность-и-версионирование).

If you're embedding `srekit` in a workflow that accepts third-party `<name>.yaml` files (an unusual setup), audit the FuncMap surface yourself before doing so.
