# srekit changelog

Scaffold a `CHANGELOG.md` in [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) format. Auto-detects the GitHub repo from `git config remote.origin.url`.

## Synopsis

```bash
srekit changelog [flags]
```

## Flags

| Flag | Required | Description |
|---|---|---|
| `--repo` | no | `<owner>/<name>` slug. If omitted, srekit uses `meta.repo` from a `--from` payload, and failing that reads `git config remote.origin.url` and parses GitHub SSH or HTTPS URLs. |
| `--version` | no | Initial version anchor (e.g. `0.1.0`). Default: `0.1.0`. |
| `--from` | no | Read section bodies from a JSON file; `-` reads standard input. |

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

## Structured input

`--json` emits the document as `{meta, sections}`, and `--from` feeds that shape back, so the section bodies can be filled in by a script or an agent instead of by hand:

```bash
srekit changelog --repo acme/api --json > cl.json
# ...replace the body of the "unreleased" section...
srekit changelog --from cl.json
```

Section bodies you supply are inserted verbatim — no template evaluation — so Markdown containing `{{ }}` round-trips unchanged. Omitted sections fall back to the artifact's defaults. A section id the artifact does not declare is an error naming the offender, never a silent skip.

`meta` in the payload supplies `repo`, `initialVersion` and `today`. Flags win over the file, and the file wins over the git remote, so a payload carrying `meta.repo` renders correctly outside a git repository.

Unlike [`srekit postmortem`](postmortem.md), `changelog` offers no `--schema` and no `--validate`: its artifact declares no required sections, so payload validation could only ever pass and a schema of two string fields says less than `--json` itself.

## Output

The scaffold includes the `[Unreleased]` / `[<version>]` skeleton with the six Keep a Changelog subsections, and ends with a block of link reference definitions pointing at `github.com/<repo>/compare/v<version>...HEAD`.

That link block is a document-level footer, not part of the last section's body. It therefore survives a `--from` payload that replaces `initial_release`, and it is the block a future `changelog release` rewrites.

## Template shape

`changelog` ships as a v1 YAML artifact (`internal/tmpl/templates/changelog.yaml`) — H1 + `header_body` (the intro paragraph) + two sections (`unreleased` and `initial_release`) + `footer_body` (the link reference definitions). The `initial_release` section title is dynamic (`[{{ .Meta.InitialVersion }}] - {{ .Meta.Today }}`); section titles are template-evaluated since v0.20.0. Template expressions reference `.Meta.<Field>` for `Today` (date `2006-01-02`), `Repo` (`<owner>/<name>`), `InitialVersion`. See [`srekit postmortem`](postmortem.md#customizing-the-artifact-v1-format-v0140) for the full schema reference.

## See also

- [`srekit rfc`](rfc.md), [`srekit postmortem`](postmortem.md) — the docs whose history feeds the changelog.
