## Why

`srekit changelog` writes a correct Keep a Changelog scaffold once, and then leaves. Everything the format actually demands over a project's life happens afterwards, by hand: moving `[Unreleased]` into a dated version heading at release time, pruning the change-type subsections that stayed empty, adding the compare link for the new version, and repointing `[Unreleased]` at the new tag. That is the part people get wrong, and this repository is the proof — its own changelog has needed hand-repair of the compare-link block more than once.

The scaffold also has no way to tell a user that their file has drifted out of the format. Keep a Changelog 1.1.0 names the failure modes explicitly — regional date formats, versions out of order, missing links, change types invented on the spot — and none of them are visible until someone reads the file carefully or a downstream tool chokes on it.

Two things are simply absent. Yanked releases (`## [0.0.5] - 2014-12-13 [YANKED]`) are part of the specification and appear nowhere in `srekit`. And the whole notion of an existing changelog is absent: `srekit` today only ever creates new files.

That last point is the real decision in this proposal. Adding `changelog release` turns `srekit` from a generator of new documents into a tool that also edits documents a user already owns. That is a genuine widening of what the tool is, and it deserves to be stated rather than smuggled in as a feature.

## What Changes

- Add `srekit changelog release --version X.Y.Z` — moves the accumulated `[Unreleased]` entries under a new `## [X.Y.Z] - YYYY-MM-DD` heading, empties `[Unreleased]`, adds the compare link for the new version, and repoints `[Unreleased]` at it. Operates on `CHANGELOG.md` in the working directory, or on a path given as an argument.
- Change-type subsections left with nothing but the scaffold's `-` placeholder are dropped during the move, so a released version never ships an empty `### Deprecated`.
- Add `--yanked`, which marks the released version `## [X.Y.Z] - YYYY-MM-DD [YANKED]`.
- Add `--date`, so a release can be dated other than today (backfilling history, releasing across a timezone boundary).
- Add `srekit changelog validate [FILE]` — a linter for the format: heading shape, ISO dates, descending version order, change types drawn from the specified six, a link definition for every version, well-formed `[YANKED]` marker. Reports per-check `OK`/`FAIL` and exits non-zero on failure, matching the reporting style `postmortem --validate` already established.
- `srekit changelog` with no subcommand keeps generating a scaffold, with the same flags and the same output. Nothing about the existing invocation changes.
- The output contract for `release` is deliberately *not* the generator bundle: it edits the file it was pointed at, in place. `--dry-run` and `--stdout` show the result without writing, `--json` emits the parsed document. `--out` and `--force` are not offered, because the command's purpose is to update that one file — and a flag that has no meaning must not exist.
- No new dependencies. Parsing a changelog is line-oriented work over Markdown headings and link reference definitions; nothing enters the build graph.

Nothing here is **BREAKING**. No existing command, flag, default, or output changes meaning; `changelog` gains subcommands and keeps its bare behaviour.

## Capabilities

### New Capabilities

- `changelog-maintenance`: operating on an existing `CHANGELOG.md` — cutting a release, marking it yanked, keeping the link block correct, and reporting where a file departs from Keep a Changelog.

### Modified Capabilities

- `artifact-generation`: the catalog requirement gains a sentence allowing a generator to carry maintenance subcommands without leaving the catalog, so `srekit changelog release` does not read as a second, undocumented generator.
- `output-routing`: the uniform flag bundle requirement is scoped explicitly to generators, and a command that edits an existing document is required to omit `--out` and `--force` rather than expose them with invented meanings.

## Impact

- New internal package for reading a Keep a Changelog document: version headings, change-type subsections, the trailing link reference block. Byte-preserving outside the regions it edits — a user's hand-written prose, ordering and spacing must survive untouched.
- `cmd/changelog.go` grows subcommands; the existing generator body moves under the parent's own `RunE` so the bare invocation is unchanged.
- Depends on `changelog-footer-and-structured-input`: that change moves the link definitions into a document-level footer, which is what makes the block locatable as an element rather than as prose at the end of a section body.
- Tests: unit coverage on the parser for real-world shapes (hand-edited files, missing link block, non-GitHub hosts, `[YANKED]` entries already present, prose between the heading and the first `###`), plus smoke tests for both subcommands.
- Docs, both locales: `docs/{en,ru}/commands/changelog.md` gains a release-workflow section and a validate section; the recipes page gains the release-day sequence.
- `CHANGELOG.md` under `[Unreleased]` → `Added`.
