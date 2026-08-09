## 1. Prerequisite

- [x] 1.1 Confirm `changelog-footer-and-structured-input` is implemented — the link reference block must already render as a document-level footer before anything here can locate it reliably

## 2. Changelog document scanner

- [x] 2.1 New internal package that scans a changelog into a document value: preamble, `[Unreleased]` region, one region per version heading, link reference block, trailing content — each recorded by byte offset so unedited regions can be copied through verbatim
- [x] 2.2 Parse version headings into version, date, and yanked flag; leave an unparseable heading marked as such rather than dropping it
- [x] 2.3 Parse change-type subsections within a region, taking the recognized heading vocabulary as a parameter rather than hardcoding it
- [x] 2.4 Parse the link reference block into ordered `label → url` pairs, and derive from the `[Unreleased]` definition the host, repository path, URL shape and tag prefix in use
- [x] 2.5 Numeric-segment version comparison covering `1.10.0 > 1.9.0`; anything unparseable is reported, never silently ordered
- [x] 2.6 Unit tests on real-world shapes: hand-edited file with preamble prose, no link block, non-GitHub host, tags with no `v` prefix, an existing `[YANKED]` entry, prose between a version heading and its first `###`, entries with no subsection, HTML comments

## 3. Release rewrite

- [x] 3.1 Build the released-version body from `[Unreleased]`, dropping change-type subsections that are empty or hold only the bare `-` placeholder
- [x] 3.2 Refuse with a non-zero exit and an untouched file when nothing remains to release, when the requested version already has a heading, when there is no `[Unreleased]` heading, or when the target file does not exist
- [x] 3.3 Splice the new version heading and body immediately after `[Unreleased]`, leaving `[Unreleased]` present and empty
- [x] 3.4 Append ` [YANKED]` to the heading under `--yanked`
- [x] 3.5 Rewrite the link block: repoint `[Unreleased]` at the new tag, insert the new version's definition above the previously newest one, preserving the document's own host and tag prefix; create the block from the resolved repository slug when the document has none
- [x] 3.6 Point the first released version's definition at its release tag rather than at a comparison
- [x] 3.7 Verify byte-identical preservation of preamble, previously released versions, and trailing content in a test that diffs the whole file

## 4. Command wiring

- [x] 4.1 Turn `changelog` into a parent command that keeps its current generator behaviour on the bare invocation, with the existing flags and output routing untouched
- [x] 4.2 Add the `release` subcommand: `--version` (required), `--date`, `--yanked`, optional positional target path defaulting to `CHANGELOG.md`
- [x] 4.3 Reject a `--date` that is not `YYYY-MM-DD` before reading the file
- [x] 4.4 Output routing for `release`: in-place write with a `wrote <path>` line, `--dry-run` prefixed print without writing, `--stdout` print without writing, `--json` emitting the parsed document; bind neither `--out` nor `--force`
- [x] 4.5 Add the `validate` subcommand with an optional positional target path, no write path, and the `OK`/`FAIL` per-check reporting style used by `postmortem --validate`
- [x] 4.6 Implement the checks: heading shape, `[Unreleased]` present and first, descending version order, no duplicate versions, change types drawn from the six, a link definition per version
- [x] 4.7 Exit non-zero when any check fails, after reporting every check

## 5. Tests

- [x] 5.1 `cmd/cmd_test.go`: release moves entries under a dated heading and empties `[Unreleased]`
- [x] 5.2 `cmd/cmd_test.go`: placeholder subsections do not ship; only change types with real entries appear
- [x] 5.3 `cmd/cmd_test.go`: `--date`, `--yanked`, and an explicit positional target path
- [x] 5.4 `cmd/cmd_test.go`: each refusal case exits non-zero and leaves the file byte-identical
- [x] 5.5 `cmd/cmd_test.go`: `--dry-run` and `--stdout` write nothing; `--help` lists neither `--out` nor `--force`, and passing either fails as an unknown flag
- [x] 5.6 `cmd/cmd_test.go`: link block after release — `[Unreleased]` repointed, new definition inserted in order, host and tag prefix preserved, first-release case pointing at a tag
- [x] 5.7 `cmd/cmd_test.go`: validate passes on a released scaffold and fails with a named check for each of regional date, invented change type, missing link definition, versions out of order, duplicate version
- [x] 5.8 `cmd/cmd_test.go`: bare `srekit changelog` still produces the same scaffold, and `--help` lists both subcommands

## 6. Documentation and release notes

- [x] 6.1 `docs/en/commands/changelog.md` and `docs/ru/commands/changelog.md`: release-workflow section (dry-run, review, release, commit, tag — explicitly noting the command does not touch git) and a validate section listing every check
- [x] 6.2 Both locales: document yanked releases and when to use `--date`
- [x] 6.3 Recipes page in both locales: the release-day sequence
- [x] 6.4 `CHANGELOG.md` under `[Unreleased]` → `Added`, for both subcommands and `--yanked`
- [x] 6.5 Update the project description in `CLAUDE.md` and `openspec/config.yaml` to say `srekit` also maintains an existing changelog, not only generates artifacts

## 7. Verification

- [x] 7.1 `go build ./... && go test -race ./... && golangci-lint run`
- [x] 7.2 `task docs:build` passes strict mode with both locales updated
- [x] 7.3 Run `srekit changelog validate` against this repository's own `CHANGELOG.md` and confirm the reported failures are real drift rather than parser gaps
- [x] 7.4 Dry-run a release against this repository's own `CHANGELOG.md` and diff the result by hand
