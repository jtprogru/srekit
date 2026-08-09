# srekit changelog

Scaffold a `CHANGELOG.md` in [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) format. Auto-detects the GitHub repo from `git config remote.origin.url`.

The bare invocation generates. Two subcommands maintain a changelog that already exists: [`release`](#cutting-a-release) cuts a version, [`validate`](#validating-an-existing-changelog) lints one. They are not catalog entries of their own — `srekit changelog` keeps the behaviour it always had.

## Synopsis

```bash
srekit changelog [flags]
srekit changelog release --version X.Y.Z [FILE] [flags]
srekit changelog validate [FILE]
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

That link block is a document-level footer, not part of the last section's body. It therefore survives a `--from` payload that replaces `initial_release`, and it is the block [`changelog release`](#cutting-a-release) rewrites.

## Cutting a release

`srekit changelog release --version X.Y.Z` moves everything under `## [Unreleased]` into a new `## [X.Y.Z] - YYYY-MM-DD` heading placed directly beneath it, leaves `[Unreleased]` empty, and updates the link block so `[Unreleased]` compares the new tag against `HEAD`.

```bash
srekit changelog release --version 1.2.0
```

Change-type subsections holding nothing but the scaffold's bare `-` are dropped on the way through, so a released version never ships an empty `### Deprecated`. `[Unreleased]` is left genuinely empty rather than refilled with the six-type skeleton — Keep a Changelog's own example does the same, and refilling would put six headings and six placeholders into every release diff only for the next release to strip them again.

### Flags

| Flag | Required | Description |
|---|---|---|
| `--version` | yes | The version being released, without a tag prefix (e.g. `1.2.0`). |
| `--date` | no | Release date in `YYYY-MM-DD` form. Default: today. |
| `--yanked` | no | Mark the release withdrawn: `## [X.Y.Z] - YYYY-MM-DD [YANKED]`. |
| `--dry-run` | no | Print the result, do not write. |
| `--stdout` | no | Print the result, do not write. |
| `--json` | no | Emit the parsed document (versions, dates, yanked state, change types, link definitions) and do not write. |

The target is `CHANGELOG.md` in the working directory, or a path given as the single positional argument:

```bash
srekit changelog release --version 1.2.0 docs/CHANGELOG.md
```

### Neither `--out` nor `--force`

`release` is not a generator, and it does not carry the [shared output flags](index.md#shared-output-flags) in full. It rewrites the file it was pointed at, so a second destination has no meaning, and an overwrite guard would guard against the command's own purpose. Passing either flag is an unknown-flag error, not a silently ignored option.

### The release-day sequence

The command edits text. It does not commit, tag or push — that stays under your hand:

```bash
srekit changelog release --version 1.2.0 --dry-run   # 1. look at it first
srekit changelog release --version 1.2.0             # 2. cut it
git diff CHANGELOG.md                                # 3. review
git commit -am "release: 1.2.0"                      # 4. commit
git tag -a v1.2.0 -m "1.2.0" && git push origin v1.2.0   # 5. tag
```

### Yanked releases

A release withdrawn after publication is marked, not deleted — the version number is burnt either way, and a reader needs to know why the gap is there:

```bash
srekit changelog release --version 0.0.5 --date 2014-12-13 --yanked
# ## [0.0.5] - 2014-12-13 [YANKED]
```

`--date` exists for exactly the cases where "today" is wrong: backfilling a version that shipped before you started using this tool, or cutting a release across a timezone boundary where the tag's date and your local date disagree.

### Link conventions come from the document

The new definitions are built from the document's own `[Unreleased]` line, not from the git remote. That line already encodes the host, the repository path, the comparison URL shape and whether tags carry a `v` prefix, so a project on a self-hosted GitLab, or one whose tags are bare `1.2.0`, keeps its own convention:

```
[Unreleased]: https://git.example.com/group/proj/-/compare/1.1.0...HEAD
```

releases to

```
[Unreleased]: https://git.example.com/group/proj/-/compare/1.2.0...HEAD
[1.2.0]: https://git.example.com/group/proj/-/compare/1.1.0...1.2.0
```

The new version's definition compares against the previously newest release, or points at its own release tag when it is the first one. Only when the document has no link block at all is the repository slug resolved from git, the same way `srekit changelog` resolves it.

### What it refuses

Each of these exits non-zero and leaves the file byte-identical:

| Condition | Why |
|---|---|
| `--date` is not `YYYY-MM-DD` | Checked before the file is read, so a typo cannot get as far as touching the document. |
| The target file does not exist | Reported with the path and a pointer at `srekit changelog`. Nothing is created. |
| No `## [Unreleased]` heading | There is no insertion point, and guessing one is how a rewriter destroys history. |
| `[Unreleased]` has no entries | Nothing to release. Placeholder-only subsections do not count. |
| The version already has a heading | Re-running is a mistake, not an idempotent no-op: the entries it would move are no longer the ones that shipped. Fix the file by hand. |

### Everything else is preserved verbatim

Only three regions change: `[Unreleased]`, the inserted version, and the link block. Hand-written preamble, blank-line style, bullet-marker style, previously released versions and trailing content all come out byte for byte as they went in. That is a property of the design — the rewriter splices at byte offsets rather than reserializing a parsed model — and it is why the release diff stays reviewable.

## Validating an existing changelog

`srekit changelog validate [FILE]` reports, per check, where a document departs from Keep a Changelog. It never rewrites the file.

```bash
srekit changelog validate
```

```
OK    heading-shape
OK    unreleased-section
FAIL  version-order: versions must appear in descending order: 1.1.0 (line 12) is listed above 1.2.0 (line 24)
OK    no-duplicate-versions
FAIL  change-types: unrecognized change type line 31: Improvements; allowed: Added, Changed, Deprecated, Removed, Fixed, Security
OK    link-definitions
```

| Check | What it requires |
|---|---|
| `heading-shape` | Every version heading is `## [X.Y.Z] - YYYY-MM-DD`, optionally followed by ` [YANKED]`. This is what catches a regional date like `04/03/2026`. |
| `unreleased-section` | An `[Unreleased]` section is present and precedes every released version. |
| `version-order` | Released versions appear in descending version order. Comparison is by numeric segment, so `1.10.0` correctly sorts above `1.9.0`. |
| `no-duplicate-versions` | No version appears twice. |
| `change-types` | Every `###` subsection is one of `Added`, `Changed`, `Deprecated`, `Removed`, `Fixed`, `Security`. |
| `link-definitions` | Every version heading has a matching definition in the link reference block. |

Every check is reported whether it passed or failed — a person repairing a drifted changelog wants the whole list in one pass, not the first failure. The command exits non-zero if any check failed.

The change-type vocabulary is the specification's six, not whatever your customized `changelog.yaml` happens to say. A renamed heading fails here, which is the correct answer: the format names those six.

`validate` reports; it does not repair. Fixing a drifted document is an edit you should see.

## Template shape

`changelog` ships as a v1 YAML artifact (`internal/tmpl/templates/changelog.yaml`) — H1 + `header_body` (the intro paragraph) + two sections (`unreleased` and `initial_release`) + `footer_body` (the link reference definitions). The `initial_release` section title is dynamic (`[{{ .Meta.InitialVersion }}] - {{ .Meta.Today }}`); section titles are template-evaluated since v0.20.0. Template expressions reference `.Meta.<Field>` for `Today` (date `2006-01-02`), `Repo` (`<owner>/<name>`), `InitialVersion`. See [`srekit postmortem`](postmortem.md#customizing-the-artifact-v1-format-v0140) for the full schema reference.

## See also

- [`srekit rfc`](rfc.md), [`srekit postmortem`](postmortem.md) — the docs whose history feeds the changelog.
