# changelog-maintenance Specification

## Purpose
Defines what `srekit` does with a `CHANGELOG.md` that already exists: cutting a release by moving the accumulated `[Unreleased]` entries under a dated version heading, keeping the compare-link block correct as versions accumulate, marking a release yanked, and reporting where a document departs from Keep a Changelog. This is the only capability in `srekit` that edits a document the user already owns, so its first obligation is to leave everything it did not intend to change exactly as it found it.
## Requirements
### Requirement: Cutting a release moves Unreleased into a dated version

`srekit changelog release --version X.Y.Z` SHALL rewrite the target document so that everything under `## [Unreleased]` appears under a new `## [X.Y.Z] - YYYY-MM-DD` heading placed immediately after it, and `## [Unreleased]` is left with no entries.

The new heading SHALL be dated with the current date in `YYYY-MM-DD` form unless `--date` supplies another, which SHALL be accepted only in that same form.

The target SHALL be `CHANGELOG.md` in the working directory, or a path given as the command's single positional argument.

#### Scenario: Entries move under the new version
- **GIVEN** a `CHANGELOG.md` whose `## [Unreleased]` contains `### Added` with two bullets
- **WHEN** a user runs `srekit changelog release --version 1.2.0`
- **THEN** the file SHALL contain `## [1.2.0] - <today>` followed by `### Added` and those two bullets
- **AND** `## [Unreleased]` SHALL remain in the file with no entries under it
- **AND** `## [1.2.0]` SHALL appear before every previously released version

#### Scenario: Explicit release date
- **WHEN** a user runs `srekit changelog release --version 1.2.0 --date 2026-03-04`
- **THEN** the new heading SHALL read `## [1.2.0] - 2026-03-04`

#### Scenario: Non-ISO date is rejected
- **WHEN** a user runs `srekit changelog release --version 1.2.0 --date 04/03/2026`
- **THEN** the command SHALL exit non-zero with an error naming the expected `YYYY-MM-DD` form
- **AND** the file SHALL be unchanged

#### Scenario: Explicit target path
- **WHEN** a user runs `srekit changelog release --version 1.2.0 docs/CHANGELOG.md`
- **THEN** that file SHALL be the one rewritten and `CHANGELOG.md` SHALL NOT be touched

### Requirement: Placeholder change-type subsections are dropped on release

A change-type subsection under `[Unreleased]` whose body is empty or consists only of the scaffold's bare `-` placeholder SHALL NOT be carried into the released version. A released version SHALL contain only the change types that actually have entries.

#### Scenario: Scaffold placeholders do not ship
- **GIVEN** an `## [Unreleased]` carrying the six scaffold subsections where only `### Fixed` has a real bullet
- **WHEN** a user cuts a release
- **THEN** the released version SHALL contain `### Fixed` and its bullet
- **AND** it SHALL NOT contain `### Added`, `### Changed`, `### Deprecated`, `### Removed` or `### Security`

### Requirement: Releasing nothing is refused

When `[Unreleased]` has no entries after placeholder subsections are discounted, the command SHALL exit non-zero with an error saying there is nothing to release, and SHALL NOT modify the file.

#### Scenario: Empty Unreleased
- **GIVEN** a `CHANGELOG.md` whose `## [Unreleased]` holds only placeholder subsections
- **WHEN** a user runs `srekit changelog release --version 1.2.0`
- **THEN** the command SHALL exit non-zero with an error stating that `[Unreleased]` has no entries
- **AND** the file SHALL be byte-identical to before

### Requirement: Yanked releases

`--yanked` SHALL mark the released version as withdrawn by appending ` [YANKED]` to its heading, in the form Keep a Changelog specifies: `## [X.Y.Z] - YYYY-MM-DD [YANKED]`.

#### Scenario: Yanked heading form
- **WHEN** a user runs `srekit changelog release --version 0.0.5 --date 2014-12-13 --yanked`
- **THEN** the new heading SHALL read exactly `## [0.0.5] - 2014-12-13 [YANKED]`

### Requirement: The link reference block is kept correct

Cutting a release SHALL update the document's trailing link reference block so that `[Unreleased]` compares the newly released tag against `HEAD`, and a definition for the new version is inserted above the previously newest one.

The host, path and tag prefix SHALL be taken from the document's existing `[Unreleased]` definition, so a project on a non-GitHub host, or one whose tags carry no `v` prefix, keeps its own convention. When the document has no link block, the block SHALL be created from the repository slug resolved the same way `srekit changelog` resolves it.

The definition for the new version SHALL compare it against the previously newest version, or point at its release tag when it is the first released version.

#### Scenario: Unreleased is repointed
- **GIVEN** a link block whose `[Unreleased]` compares `v1.1.0...HEAD`
- **WHEN** a user releases `1.2.0`
- **THEN** `[Unreleased]` SHALL compare `v1.2.0...HEAD`
- **AND** a `[1.2.0]` definition comparing `v1.1.0...v1.2.0` SHALL be present

#### Scenario: Existing host and tag prefix are preserved
- **GIVEN** a link block pointing at a self-hosted GitLab instance with tags carrying no `v` prefix
- **WHEN** a user cuts a release
- **THEN** the new definitions SHALL use that same host and SHALL NOT introduce a `v` prefix

#### Scenario: First release points at a tag
- **GIVEN** a document with no released versions
- **WHEN** a user releases `0.1.0`
- **THEN** the `[0.1.0]` definition SHALL point at that version's release tag rather than at a comparison

### Requirement: Everything outside the edited regions is preserved verbatim

Rewriting SHALL change only the `[Unreleased]` section, the inserted version heading and body, and the link reference block. Every other byte of the document — prose above the first version, blank-line style, ordering, entries of previously released versions, trailing content — SHALL survive unchanged.

#### Scenario: Hand-written preamble survives
- **GIVEN** a `CHANGELOG.md` with a hand-edited paragraph between the H1 and `## [Unreleased]`
- **WHEN** a user cuts a release
- **THEN** that paragraph SHALL be byte-identical afterwards

#### Scenario: Older versions are untouched
- **GIVEN** a document with three previously released versions
- **WHEN** a user cuts a release
- **THEN** those three versions' headings and bodies SHALL be byte-identical afterwards

### Requirement: A version already present is refused

When the requested version already has a heading in the document, the command SHALL exit non-zero with an error naming it, and SHALL NOT modify the file. Re-running a release is a mistake, not an idempotent no-op: the entries it would move are no longer the ones that were released.

#### Scenario: Repeat release
- **GIVEN** a document already containing `## [1.2.0] - 2026-03-04`
- **WHEN** a user runs `srekit changelog release --version 1.2.0`
- **THEN** the command SHALL exit non-zero with an error naming `1.2.0` as already released
- **AND** the file SHALL be unchanged

### Requirement: An editing command routes output differently from a generator

`changelog release` SHALL write its result back to the file it read, and SHALL print a `wrote <path>` confirmation subject to `--quiet`.

It SHALL expose `--dry-run`, `--stdout`, `--json` and inherit `--quiet`. It SHALL NOT expose `--out` or `--force`: the command exists to update one named document, so a second destination has no meaning and an overwrite guard would guard against the command's own purpose.

- `--dry-run` SHALL print the resulting document prefixed by a line naming the file that would have been written, and SHALL NOT modify it.
- `--stdout` SHALL print the resulting document and SHALL NOT modify the file.
- `--json` SHALL emit the parsed document — versions, their dates, their yanked state, their change types, and the link definitions — and SHALL NOT modify the file.

#### Scenario: In-place by default
- **WHEN** a user runs `srekit changelog release --version 1.2.0`
- **THEN** `CHANGELOG.md` SHALL be updated in place and a `wrote CHANGELOG.md` line SHALL be printed

#### Scenario: Dry-run leaves the file alone
- **WHEN** a user runs `srekit changelog release --version 1.2.0 --dry-run`
- **THEN** standard output SHALL begin with a line naming `CHANGELOG.md` as the file that would have been written, followed by the resulting document
- **AND** `CHANGELOG.md` SHALL be byte-identical to before

#### Scenario: Neither --out nor --force exists
- **WHEN** a user runs `srekit changelog release --help`
- **THEN** `--out` and `--force` SHALL NOT be listed
- **AND** running the command with either SHALL exit non-zero with an unknown-flag error

### Requirement: A missing or unparseable target fails without writing

When the target file does not exist, the command SHALL exit non-zero with an error naming the path and suggesting `srekit changelog` to create one. When the file exists but has no `## [Unreleased]` heading, the command SHALL exit non-zero with an error saying so. In neither case SHALL any file be created or modified.

#### Scenario: No changelog present
- **WHEN** a user runs `srekit changelog release --version 1.2.0` in a directory with no `CHANGELOG.md`
- **THEN** the command SHALL exit non-zero with an error naming `CHANGELOG.md` and pointing at `srekit changelog`
- **AND** no file SHALL be created

#### Scenario: No Unreleased heading
- **GIVEN** a `CHANGELOG.md` with released versions but no `## [Unreleased]` heading
- **WHEN** a user cuts a release
- **THEN** the command SHALL exit non-zero with an error naming the missing heading
- **AND** the file SHALL be unchanged

### Requirement: Format validation of an existing changelog

`srekit changelog validate [FILE]` SHALL read a changelog and report, per check, whether the document conforms to Keep a Changelog. It SHALL NOT modify the file. The checks SHALL be:

- every version heading has the form `[X.Y.Z] - YYYY-MM-DD`, optionally followed by ` [YANKED]`
- an `[Unreleased]` section is present and precedes every released version
- released versions appear in descending version order
- no version appears twice
- every change-type subsection is one of the recognized change types, in either supported language
- the document uses one change-type vocabulary rather than mixing two
- every version heading has a matching definition in the link reference block

Each check SHALL be reported as `OK <check>` or `FAIL <check>: <detail>` naming the offending line where one applies. The command SHALL exit non-zero when any check fails.

#### Scenario: A clean file passes
- **GIVEN** a `CHANGELOG.md` generated by `srekit changelog` and released once
- **WHEN** a user runs `srekit changelog validate`
- **THEN** every check SHALL be reported `OK` and the command SHALL exit zero

#### Scenario: Regional date is caught
- **GIVEN** a version heading reading `## [1.2.0] - 04/03/2026`
- **WHEN** a user runs `srekit changelog validate`
- **THEN** the date check SHALL be reported `FAIL` naming that heading
- **AND** the command SHALL exit non-zero

#### Scenario: Invented change type is caught
- **GIVEN** a version containing a `### Improvements` subsection
- **WHEN** a user runs `srekit changelog validate`
- **THEN** the change-type check SHALL be reported `FAIL` naming `Improvements` and listing every allowed change type

#### Scenario: Missing link definition is caught
- **GIVEN** a document with a `## [1.2.0]` heading and no `[1.2.0]` definition
- **WHEN** a user runs `srekit changelog validate`
- **THEN** the link check SHALL be reported `FAIL` naming `1.2.0`

#### Scenario: Versions out of order are caught
- **GIVEN** a document listing `## [1.1.0]` above `## [1.2.0]`
- **WHEN** a user runs `srekit changelog validate`
- **THEN** the ordering check SHALL be reported `FAIL` naming both versions

### Requirement: Change-type vocabulary is detected from the document

`changelog release` and `changelog validate` SHALL recognize change-type headings in both the English set (`Added`, `Changed`, `Deprecated`, `Removed`, `Fixed`, `Security`) and the Russian set (`Добавлено`, `Изменено`, `Устарело`, `Удалено`, `Исправлено`, `Безопасность`).

The vocabulary in force SHALL be determined by the document's own content — the first recognized change-type heading — not by `--lang` and not by configuration. A document is read as it is written; the language selection governs what is generated, never what is parsed.

#### Scenario: A Russian changelog releases correctly
- **GIVEN** a `CHANGELOG.md` whose `## [Unreleased]` carries `### Исправлено` with one entry
- **WHEN** a user runs `srekit changelog release --version 1.2.0`
- **THEN** the released version SHALL carry `### Исправлено` and that entry

#### Scenario: Placeholder pruning works in either vocabulary
- **GIVEN** a Russian document whose unreleased section carries all six change types with only `### Добавлено` holding a real entry
- **WHEN** a user cuts a release
- **THEN** the released version SHALL contain `### Добавлено` alone

#### Scenario: Parsing ignores the language selection
- **GIVEN** a `CHANGELOG.md` written with English change types
- **WHEN** a user runs `srekit changelog release --version 1.2.0 --lang ru`
- **THEN** the document's English change types SHALL be preserved unchanged

### Requirement: A document may not mix change-type vocabularies

`changelog validate` SHALL report a failure when a document uses change-type headings from more than one language, naming the offending headings. A mixed document is a partial translation, and silently accepting it produces a file no tool and no reader can treat consistently.

`changelog release` SHALL refuse a mixed document with the same explanation and SHALL NOT modify it.

#### Scenario: Mixed vocabularies fail validation
- **GIVEN** a document containing both `### Added` and `### Исправлено`
- **WHEN** a user runs `srekit changelog validate`
- **THEN** a check SHALL be reported `FAIL` naming both headings
- **AND** the command SHALL exit non-zero

#### Scenario: Release refuses a mixed document
- **GIVEN** the same document
- **WHEN** a user runs `srekit changelog release --version 1.2.0`
- **THEN** the command SHALL exit non-zero with an error naming the mixed vocabularies
- **AND** the file SHALL be byte-identical to before

### Requirement: Validation reports which vocabulary it used

`changelog validate` SHALL name the change-type vocabulary it detected in its output, so a user who expected one language and got the other sees the mismatch instead of a list of passing checks that measured the wrong thing.

#### Scenario: Detected vocabulary is reported
- **WHEN** a user runs `srekit changelog validate` on a Russian document
- **THEN** the output SHALL state that the Russian change-type vocabulary was detected

#### Scenario: A document with no change types at all
- **GIVEN** a document whose versions carry entries but no `###` change-type headings
- **WHEN** a user runs `srekit changelog validate`
- **THEN** the change-type check SHALL be reported `FAIL` stating that no recognized change types were found

### Requirement: The bare changelog invocation is unchanged

`srekit changelog` with no subcommand SHALL keep generating a scaffold with exactly the flags, defaults, output routing and document content it had before subcommands existed.

#### Scenario: Scaffold generation still works
- **WHEN** a user runs `srekit changelog --repo acme/api --stdout`
- **THEN** the output SHALL be the same scaffold document the command produced before `release` and `validate` existed

#### Scenario: Subcommands are discoverable
- **WHEN** a user runs `srekit changelog --help`
- **THEN** `release` and `validate` SHALL be listed as subcommands

