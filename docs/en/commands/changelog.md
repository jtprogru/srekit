# srekit changelog

Scaffold a `CHANGELOG.md` in [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) format. Auto-detects the GitHub repo from `git config remote.origin.url`.

## Synopsis

```bash
srekit changelog [flags]
```

## Flags

| Flag | Required | Description |
|---|---|---|
| `--repo` | no | `<owner>/<name>` slug. If omitted, srekit reads `git config remote.origin.url` and parses GitHub SSH or HTTPS URLs. |
| `--version` | no | Initial version anchor (e.g. `0.1.0`). Default: `0.1.0`. |

Plus the [shared output flags](index.md#shared-output-flags). Default filename: `CHANGELOG.md`.

## Examples

Inside a git repo with an `origin` remote pointing at GitHub:

```bash
srekit changelog --out CHANGELOG.md
# detects repo from git remote, version defaults to 0.1.0
```

Explicit:

```bash
srekit changelog --repo jtprogru/srekit --version 0.1.0 --out CHANGELOG.md
```

Outside a git repo, omitting `--repo` is an error (no silent `OWNER/REPO` placeholder — that bit users in v0.2):

```bash
srekit changelog --stdout
# Error: could not detect repo from git remote — pass --repo OWNER/NAME
```

## Output

The scaffold pre-renders compare links pointing at `github.com/<repo>/compare/v<version>...HEAD` and includes the `[Unreleased]` / `[<version>]` skeleton with `Added` / `Changed` / `Fixed` subsections.

## Template shape

```go
struct {
    Repo    string  // "<owner>/<name>"
    Version string
    Today   string  // RFC 3339-ish
}
```

## See also

- [`srekit rfc`](rfc.md), [`srekit postmortem`](postmortem.md) — the docs whose history feeds the changelog.
