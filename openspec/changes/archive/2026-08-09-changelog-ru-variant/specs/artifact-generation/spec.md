## MODIFIED Requirements

### Requirement: Rendered documents are bilingual by default

Shipped artifact templates other than `changelog` SHALL use `Русский (English)` headings and labels with Russian body prose. Technical identifiers (SLO, SLI, RFC, SEV, UTC, PromQL expressions) and YAML frontmatter keys SHALL remain English.

The `changelog` artifact SHALL be entirely English by default, so that tooling built around Keep a Changelog keeps working. A Russian variant SHALL be available by explicit opt-in only, and SHALL never be produced by an unqualified invocation.

Within the Russian variant, only the change-type headings and the surrounding prose SHALL be translated. Version headings, the `[Unreleased]` heading, and the link reference labels SHALL remain in their English form: they are identifiers that must match the link definitions and the project's tags, not prose.

#### Scenario: Section headings carry both languages
- **WHEN** a user renders any artifact other than `changelog`
- **THEN** its section headings SHALL follow the `Русский (English)` form

#### Scenario: Changelog stays English by default
- **WHEN** a user runs `srekit changelog`
- **THEN** the rendered document SHALL contain only English headings (`Added`, `Changed`, `Fixed`, `Unreleased`)

#### Scenario: Russian variant is opt-in
- **WHEN** a user runs `srekit changelog --lang ru`
- **THEN** the change-type headings SHALL read `Добавлено`, `Изменено`, `Устарело`, `Удалено`, `Исправлено`, `Безопасность`
- **AND** the header prose SHALL be Russian and SHALL link to the Russian edition of the Keep a Changelog specification

#### Scenario: Identifiers stay English in the Russian variant
- **WHEN** a user runs `srekit changelog --lang ru`
- **THEN** the unreleased section's heading SHALL read `## [Unreleased]`
- **AND** the link reference block SHALL use the labels `[Unreleased]` and `[<version>]`

## ADDED Requirements

### Requirement: Changelog language selection

`changelog` SHALL accept `--lang` with the values `en` and `ru`, defaulting to `en`. The value SHALL be resolved as flag, then the `changelog_lang` configuration value, then `en`. An unrecognized value SHALL fail with an error naming the accepted values, before any file is written.

The selected language SHALL apply to the whole `changelog` command group, so `release` and `validate` operate under the same selection as generation.

#### Scenario: Default is English
- **WHEN** a user runs `srekit changelog` with no `--lang` and no configuration
- **THEN** the output SHALL be byte-identical to what the command produced before `--lang` existed

#### Scenario: Configured language is used
- **GIVEN** a configuration file setting `changelog_lang: ru`
- **WHEN** a user runs `srekit changelog --stdout`
- **THEN** the Russian variant SHALL be rendered

#### Scenario: Flag beats configuration
- **GIVEN** a configuration file setting `changelog_lang: ru`
- **WHEN** a user runs `srekit changelog --lang en --stdout`
- **THEN** the English variant SHALL be rendered

#### Scenario: Unknown language is rejected
- **WHEN** a user runs `srekit changelog --lang de`
- **THEN** the command SHALL exit non-zero with an error naming `en` and `ru`
- **AND** no file SHALL be created
