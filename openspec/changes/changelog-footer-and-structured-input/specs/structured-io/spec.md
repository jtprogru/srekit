## MODIFIED Requirements

### Requirement: Structured input via --from

`postmortem` and `changelog` SHALL accept `--from FILE` carrying a JSON object with optional `meta` and `sections` maps, where `-` reads standard input. Provided section bodies SHALL replace the artifact's defaults; omitted sections SHALL fall back to their defaults.

An empty file SHALL be treated as no input rather than an error. Malformed JSON SHALL fail with an error naming the file.

`--from` SHALL exist wherever the round-trip is meaningful: a command that emits a `sections` array under `--json` and offers no way to feed it back is offering half a contract. `--schema` and `--validate` do not follow automatically — they are warranted only where an artifact declares required sections, without which validation cannot fail and a schema describes nothing a consumer could get wrong.

#### Scenario: Round-trip through JSON
- **WHEN** a user captures `srekit postmortem -T X --json` to a file, edits one section body, and runs `srekit postmortem -T X --from <file>`
- **THEN** the rendered Markdown SHALL contain the edited body for that section and defaults for the rest

#### Scenario: Changelog round-trip
- **WHEN** a user captures `srekit changelog --repo acme/api --json` to a file, replaces the `unreleased` section body with real entries, and runs `srekit changelog --repo acme/api --from <file> --stdout`
- **THEN** the `## [Unreleased]` section SHALL carry the supplied entries and the remaining sections SHALL carry their defaults

#### Scenario: Input from stdin
- **WHEN** a user pipes a JSON payload into `srekit postmortem --from -`
- **THEN** the payload SHALL be read from standard input

#### Scenario: Changelog offers no payload schema or payload validation
- **WHEN** a user runs `srekit changelog --help`
- **THEN** `--from` SHALL be listed
- **AND** `--schema` and `--validate` SHALL NOT be listed

### Requirement: Unknown section ids are rejected

Structured input naming a section id that the artifact does not declare SHALL fail with an error listing the offending ids and the known set. A misspelled id must never disappear silently — that is the failure mode most likely to go unnoticed when an agent generates the payload.

#### Scenario: Typo in a section id
- **GIVEN** an input file with a `sumary` key under `sections`
- **WHEN** a user runs `srekit postmortem --from <file> -T X`
- **THEN** the command SHALL fail with an error naming `sumary` and listing the artifact's known section ids

#### Scenario: Typo in a changelog section id
- **GIVEN** an input file with an `unreleased_notes` key under `sections`
- **WHEN** a user runs `srekit changelog --from <file>`
- **THEN** the command SHALL fail with an error naming `unreleased_notes` and listing the changelog artifact's known section ids

### Requirement: Flags outrank the input file

When a value is present both as a CLI flag and under `meta` in the input file, the flag SHALL win. This lets a caller pin a field while reusing an otherwise unchanged payload. A `meta` value present only in the file SHALL be honoured.

#### Scenario: Flag overrides file metadata
- **GIVEN** an input file whose `meta.title` is `Old title`
- **WHEN** a user runs `srekit postmortem --from <file> --title "New title"`
- **THEN** the rendered document SHALL use `New title`

#### Scenario: Title may come from the file alone
- **GIVEN** an input file whose `meta.title` is set
- **WHEN** a user runs `srekit postmortem --from <file>` with no `--title`
- **THEN** the command SHALL succeed using the file's title

#### Scenario: Changelog repository slug comes from the file
- **GIVEN** an input file whose `meta.repo` is `acme/api`
- **WHEN** a user runs `srekit changelog --from <file> --stdout` outside a git repository and without `--repo`
- **THEN** the command SHALL succeed and the compare links SHALL name `acme/api`
