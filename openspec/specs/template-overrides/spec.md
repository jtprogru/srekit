# template-overrides

## Purpose

Lets a team replace any shipped artifact template with its own version by pointing `srekit` at a directory, without forking the tool and without having to override templates it does not care about. Resolution is per-file, so a directory holding one customized artifact still gets the embedded versions of the other ten.

## Requirements

### Requirement: Templates are resolved from an ordered source chain

Template resolution SHALL walk an ordered list of sources and take the first hit. The binary's embedded templates SHALL always be the last source, so resolution never fails for a shipped artifact. When a custom templates directory is configured, it SHALL be consulted before the embedded set.

#### Scenario: Custom template shadows the embedded one
- **GIVEN** a templates directory containing `slo.yaml`
- **WHEN** a user runs `srekit slo --service api --templates-dir <dir>`
- **THEN** the document SHALL be rendered from the directory's `slo.yaml`

#### Scenario: Embedded set is always available
- **GIVEN** no templates directory is configured
- **WHEN** a user runs any generator
- **THEN** it SHALL render from the embedded template without error

### Requirement: Fallback is per file, not per directory

A template absent from the custom directory SHALL transparently fall back to the embedded version. Only a genuine read error (permissions, corruption) SHALL abort resolution.

#### Scenario: Partially populated templates directory
- **GIVEN** a templates directory containing only `slo.yaml`
- **WHEN** a user runs `srekit runbook --title "Disk full" --templates-dir <dir>`
- **THEN** the runbook SHALL render from the embedded `runbook.yaml`

#### Scenario: Unreadable custom file is not silently skipped
- **GIVEN** a templates directory containing a `slo.yaml` that cannot be read
- **WHEN** a user runs `srekit slo --service api --templates-dir <dir>`
- **THEN** the command SHALL fail with an error naming the file, rather than quietly using the embedded version

### Requirement: Templates directory precedence

The custom templates directory SHALL be resolved in this order: the `--templates-dir` flag, then the `SREKIT_TEMPLATES_DIR` environment variable, then `templates_dir:` in the configuration file. A leading `~` SHALL be expanded to the user's home directory.

#### Scenario: Flag beats environment
- **GIVEN** `SREKIT_TEMPLATES_DIR` points at directory A
- **WHEN** a user runs a generator with `--templates-dir B`
- **THEN** templates SHALL be resolved from B

#### Scenario: Tilde is expanded
- **WHEN** the configured templates directory is `~/sre-templates`
- **THEN** it SHALL resolve against the user's home directory

### Requirement: A misconfigured templates directory degrades instead of failing

When the configured templates directory does not exist or is not a directory, the command SHALL print a warning to standard error naming the path and the reason, then continue with the embedded templates. A stale path in a config file SHALL NOT break document generation.

#### Scenario: Configured directory was deleted
- **GIVEN** `templates_dir:` points at a path that no longer exists
- **WHEN** a user runs `srekit slo --service api`
- **THEN** a warning SHALL be printed to standard error
- **AND** the document SHALL still be rendered from the embedded template
- **AND** the command SHALL exit zero

#### Scenario: Configured path is a regular file
- **GIVEN** the configured templates path points at a file
- **WHEN** a user runs any generator
- **THEN** a warning naming the path SHALL be printed and the embedded templates used

### Requirement: Custom template names are flat

A template name SHALL resolve only to a file directly inside the custom directory. Names containing a path separator SHALL be treated as not found, so a template reference can never escape the configured directory.

#### Scenario: Traversal attempt does not escape
- **WHEN** resolution is attempted for a name containing `../`
- **THEN** the custom directory SHALL report it as not found and resolution SHALL fall through to the embedded set

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

### Requirement: Stale legacy files in a user directory are reported

When a custom templates directory contains pre-v1 files (`<name>.md.tmpl`, `<name>.sections.yaml`) alongside the v1 `<name>.yaml` that is actually being used, the command SHALL warn on standard error that those files are inert and point at the migration path. The warning SHALL be suppressed by `--quiet`.

#### Scenario: Leftover .tmpl after migration
- **GIVEN** a templates directory containing both `postmortem.yaml` and `postmortem.md.tmpl`
- **WHEN** a user runs `srekit postmortem --title X --templates-dir <dir>`
- **THEN** a warning naming `postmortem.md.tmpl` as unused SHALL be printed to standard error
- **AND** the document SHALL render from `postmortem.yaml`

#### Scenario: Quiet suppresses the nudge
- **WHEN** the same command is run with `--quiet`
- **THEN** no warning SHALL be printed

### Requirement: Template resolution is scoped per invocation

The resolved source chain SHALL be established once per command invocation and carry no process-global state, so concurrent invocations in the same process cannot observe each other's configuration.

#### Scenario: Two invocations with different directories
- **WHEN** two command trees are executed concurrently in one process with different `--templates-dir` values
- **THEN** each SHALL resolve templates from its own directory
