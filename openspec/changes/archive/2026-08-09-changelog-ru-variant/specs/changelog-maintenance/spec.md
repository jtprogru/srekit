## ADDED Requirements

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
