## MODIFIED Requirements

### Requirement: Artifact names resolve to a single `.yaml` file

An artifact SHALL be identified by its bare name (`postmortem`), which resolves to `postmortem.yaml` in the custom directory or, failing that, the embedded set.

Resolution SHALL be idempotent and SHALL accept the legacy pre-v1 spellings: a trailing `.tmpl`, then a trailing `.md`, then a trailing `.yaml` is stripped before `.yaml` is appended. `postmortem`, `postmortem.yaml`, `postmortem.md.tmpl`, and `postmortem.tmpl` SHALL therefore all resolve to `postmortem.yaml`.

When a language is requested, resolution SHALL first attempt `<name>.<lang>.yaml` across the whole source chain and, only if no source provides it, fall back to `<name>.yaml` across the whole source chain. Requesting a language for which no variant exists SHALL therefore be indistinguishable from requesting no language at all, rather than an error.

#### Scenario: Bare name resolves to the artifact
- **WHEN** resolution is requested for `runbook`
- **THEN** `runbook.yaml` SHALL be loaded

#### Scenario: Legacy name resolves to the same artifact
- **WHEN** resolution is requested for `runbook.md.tmpl`
- **THEN** `runbook.yaml` SHALL be loaded

#### Scenario: Resolution does not double the suffix
- **WHEN** resolution is requested for `runbook.yaml`
- **THEN** `runbook.yaml` SHALL be loaded and `runbook.yaml.yaml` SHALL NOT be looked up

#### Scenario: Language variant is preferred
- **WHEN** resolution is requested for `changelog` in Russian
- **THEN** `changelog.ru.yaml` SHALL be loaded

#### Scenario: Missing variant falls back to the base artifact
- **WHEN** resolution is requested for an artifact in a language for which no variant is shipped or present
- **THEN** the base `<name>.yaml` SHALL be loaded and the command SHALL succeed

#### Scenario: A user directory shadows the embedded variant
- **GIVEN** a templates directory containing its own `changelog.ru.yaml`
- **WHEN** a user renders the changelog in Russian with that directory configured
- **THEN** the directory's file SHALL be used and the embedded variant SHALL NOT be consulted

#### Scenario: The variant lookup precedes the fallback across all sources
- **GIVEN** a templates directory containing `changelog.yaml` but no `changelog.ru.yaml`
- **WHEN** a user renders the changelog in Russian
- **THEN** the embedded `changelog.ru.yaml` SHALL be used rather than the directory's `changelog.yaml`

#### Scenario: A language-suffixed name is idempotent
- **WHEN** resolution is requested for `changelog.ru.yaml`
- **THEN** `changelog.ru.yaml` SHALL be loaded and `changelog.ru.yaml.yaml` SHALL NOT be looked up
