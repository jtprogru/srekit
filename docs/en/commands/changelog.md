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

`changelog` ships as a v1 YAML artifact (`internal/tmpl/templates/changelog.yaml`) — H1 + `header_body` (the intro paragraph) + two sections (`unreleased` and `initial_release`). The `initial_release` section title is dynamic (`[{{ .Meta.InitialVersion }}] - {{ .Meta.Today }}`); section titles are template-evaluated since v0.20.0. Template expressions reference `.Meta.<Field>` for `Today` (date `2006-01-02`), `Repo` (`<owner>/<name>`), `InitialVersion`. See [`srekit postmortem`](postmortem.md#customizing-the-artifact-v1-format-v0140) for the full schema reference.

## See also

- [`srekit rfc`](rfc.md), [`srekit postmortem`](postmortem.md) — the docs whose history feeds the changelog.
