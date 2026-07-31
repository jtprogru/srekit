## Purpose

Makes the state `srekit` resolves before it renders anything — which config file wins, where templates come from, whether they still parse, whether `git` is reachable — inspectable in one read-only command, so a user diagnoses a misconfigured environment directly instead of inferring it from a generator's behaviour.

## ADDED Requirements

### Requirement: Diagnostic command

The CLI SHALL provide `srekit doctor`, which inspects the environment and reports the result of a fixed set of checks. It SHALL accept no positional arguments and SHALL fail with a usage error if any are given.

`doctor` SHALL honour the root command's `--config`, `--templates-dir` and `--quiet` flags, since those change what it has to report on.

#### Scenario: Command is discoverable
- **WHEN** a user runs `srekit --help`
- **THEN** `doctor` SHALL be listed with a one-line description

#### Scenario: No positional arguments
- **WHEN** a user runs `srekit doctor extra-arg`
- **THEN** the command SHALL exit non-zero with a usage error

### Requirement: Diagnostics are read-only

`doctor` SHALL NOT create, modify, or delete any file or directory, and SHALL NOT repair anything it finds wrong. A check whose subject is missing SHALL be reported, never created.

Because `doctor` writes nothing, it SHALL NOT expose `--out`, `--stdout`, `--force`, or `--dry-run`. A flag that would be ignored must not exist.

#### Scenario: Nothing is written
- **GIVEN** no config file and no templates directory exist
- **WHEN** a user runs `srekit doctor`
- **THEN** no file or directory SHALL be created anywhere
- **AND** the missing config and missing templates directory SHALL be reported as findings

#### Scenario: Write flags are absent
- **WHEN** a user runs `srekit doctor --help`
- **THEN** `--out`, `--stdout`, `--force`, and `--dry-run` SHALL NOT be listed
- **AND** `--json` SHALL be listed

### Requirement: Finding model

Every check SHALL produce exactly one finding carrying:

- a stable identifier, unique across checks, usable as a grep target and as a JSON key value
- a category — one of `config`, `templates`, `dependencies`
- a status — one of `ok`, `warn`, `error`
- a one-line summary of what was found, naming the concrete path, value or version involved
- a remedy, present whenever the status is `warn` or `error`, naming the command or setting that fixes it

Status SHALL mean: `error` — a generator will fail or produce wrong output in this environment; `warn` — the environment works but something is being ignored, outdated, or about to break; `ok` — nothing to do.

Check identifiers SHALL be treated as a public contract once `doctor` ships, because users gate CI on them.

#### Scenario: A failing check explains the fix
- **GIVEN** a check reports `error` or `warn`
- **WHEN** the finding is rendered in either output format
- **THEN** it SHALL include a remedy naming a command or setting

#### Scenario: Findings are ordered deterministically
- **WHEN** a user runs `srekit doctor` twice against an unchanged environment
- **THEN** the findings SHALL appear in the same order both times, grouped by category

### Requirement: Exit status reflects the worst finding

`doctor` SHALL exit `0` when no finding has status `error`, and `1` when at least one does. A `warn` SHALL NOT change the exit status, so a team can adopt `doctor` in CI without being blocked by advisory findings.

A failure to run a check itself — an unreadable directory, a subprocess that cannot be started — SHALL be reported as a finding of that check rather than aborting the run. `doctor` SHALL always run every check.

#### Scenario: Warnings do not fail the run
- **GIVEN** the environment produces `warn` findings but no `error` findings
- **WHEN** a user runs `srekit doctor`
- **THEN** the command SHALL exit `0`

#### Scenario: One error fails the run
- **GIVEN** at least one check reports `error`
- **WHEN** a user runs `srekit doctor`
- **THEN** the command SHALL exit `1`

#### Scenario: A broken check does not abort the rest
- **GIVEN** the configured templates directory cannot be read because of its permissions
- **WHEN** a user runs `srekit doctor`
- **THEN** that check SHALL report the read failure as a finding
- **AND** every other check SHALL still be reported

### Requirement: A healthy default install reports no problems

On a machine with no config file, no templates directory, and `git` installed, `doctor` SHALL report no `error` and no `warn` findings and exit `0`. Absence of an optional config file or templates directory is the documented default, not a defect.

#### Scenario: Fresh install is clean
- **GIVEN** no config file and no templates directory exist and `git` is on `PATH`
- **WHEN** a user runs `srekit doctor`
- **THEN** every finding SHALL be `ok`
- **AND** the command SHALL exit `0`

### Requirement: Configuration and path checks

Under the `config` category, `doctor` SHALL report:

- the config file path that will actually be read, and whether it exists — naming which of the XDG or legacy location it is
- whether that file parses; an unparseable config SHALL be `warn`, not `error`, matching the rule that a malformed config never fails a command that needs nothing from it
- the resolved templates directory, the source that supplied it (`--templates-dir`, `SREKIT_TEMPLATES_DIR`, the config file, or the built-in default), and whether that path exists and is a directory
- every `SREKIT_`-prefixed environment variable currently in effect, by name
- whether the directory holding the resolved config path is writable, so a user learns before running `config init` that it is not

A configured templates directory that does not exist, or is not a directory, SHALL be `warn`: generation still works from the embedded set, but the user's overrides are silently not applied.

#### Scenario: Resolved config path is reported
- **GIVEN** a config file exists at the XDG location
- **WHEN** a user runs `srekit doctor`
- **THEN** a `config` finding SHALL name that path and report it as the file being read

#### Scenario: Configured templates directory is missing
- **GIVEN** `SREKIT_TEMPLATES_DIR` points at a path that does not exist
- **WHEN** a user runs `srekit doctor`
- **THEN** a `warn` finding SHALL name the path, name `SREKIT_TEMPLATES_DIR` as its source, and state that generation is falling back to the embedded templates

#### Scenario: Environment overrides are visible
- **GIVEN** `SREKIT_AUTHOR` is set in the environment
- **WHEN** a user runs `srekit doctor`
- **THEN** a `config` finding SHALL name `SREKIT_AUTHOR` as an active override

#### Scenario: Malformed config is a warning
- **GIVEN** the resolved config file contains invalid YAML
- **WHEN** a user runs `srekit doctor`
- **THEN** a `warn` finding SHALL name the file and the parse error
- **AND** the command SHALL exit `0`

### Requirement: A shadowed configuration file is reported

When both the legacy config path and the XDG config path exist, `doctor` SHALL emit a `warn` finding naming both paths, stating which one wins and which one is being ignored. The same SHALL apply when both the legacy and the XDG templates directories exist.

This is the failure the XDG-with-legacy-precedence rule is designed to avoid users falling into silently, and it is invisible from any other command's output.

#### Scenario: Two config files present
- **GIVEN** both the legacy and the XDG config files exist
- **WHEN** a user runs `srekit doctor`
- **THEN** a `warn` finding SHALL name both paths, identify the legacy file as the one in effect, and state that the XDG file is not being read

#### Scenario: Only one config file present
- **GIVEN** exactly one of the two config locations exists
- **WHEN** a user runs `srekit doctor`
- **THEN** no shadowing finding SHALL be emitted

### Requirement: Author identity check

`doctor` SHALL report whether an author name and email can be resolved from the configured sources at all, and name the source each resolved value came from. When neither can be resolved, the finding SHALL be `error`, because every generator that stamps an author will fail in that environment.

Resolved values SHALL be reported as they are, without redaction: they are the user's own name and email, already present in their git config.

#### Scenario: Identity resolves from git config
- **GIVEN** no author is set in config or environment and `git config user.name` and `user.email` are set
- **WHEN** a user runs `srekit doctor`
- **THEN** an `ok` finding SHALL report the resolved name and email and name git config as their source

#### Scenario: No identity anywhere
- **GIVEN** neither config, environment, nor git config supplies an author name or email
- **WHEN** a user runs `srekit doctor`
- **THEN** an `error` finding SHALL state that generators requiring an author will fail
- **AND** the remedy SHALL name `srekit config init`, the `--author` / `--email` flags, and the git config keys
- **AND** the command SHALL exit `1`

### Requirement: Template health checks

Under the `templates` category, and only when a templates directory is configured and exists, `doctor` SHALL report:

- how many artifacts in the directory fail to parse, naming each failing file and its parse error; any parse failure SHALL be `error`, because the affected generator cannot render
- how many legacy pre-v1.0 template files are present, naming them; this SHALL be `warn`, with `srekit templates migrate` as the remedy
- how many artifacts differ from the binary's embedded version, and how many embedded artifacts are absent from the directory; when either count is non-zero the finding SHALL be `warn`, with `srekit templates upgrade` and `srekit templates diff` as remedies

When no templates directory is configured, these checks SHALL report `ok` with a summary stating that only the embedded templates are in use.

#### Scenario: Unparseable user template
- **GIVEN** a templates directory containing an artifact with a duplicate section id
- **WHEN** a user runs `srekit doctor`
- **THEN** an `error` finding SHALL name that file and its parse error
- **AND** the command SHALL exit `1`

#### Scenario: Legacy template files linger
- **GIVEN** a templates directory containing a pre-v1.0 `.tmpl` file
- **WHEN** a user runs `srekit doctor`
- **THEN** a `warn` finding SHALL name the file and offer `srekit templates migrate` as the remedy

#### Scenario: Directory has drifted from the embedded set
- **GIVEN** a templates directory whose artifacts differ from the embedded ones
- **WHEN** a user runs `srekit doctor`
- **THEN** a `warn` finding SHALL report the number of differing artifacts and offer `srekit templates upgrade`

#### Scenario: Embedded-only install
- **GIVEN** no templates directory is configured
- **WHEN** a user runs `srekit doctor`
- **THEN** the template checks SHALL be `ok` and state that the embedded templates are in use

### Requirement: External dependency check

Under the `dependencies` category, `doctor` SHALL report whether `git` is present on `PATH`, and when it is, the resolved executable path and the version it reports.

An absent `git` SHALL be `warn`, not `error`: author metadata and the changelog repository slug fall back to flags and config, so most generation still works.

`git` SHALL be the only external program `doctor` looks for, and `doctor` SHALL make no network request — including no attempt to check for a newer `srekit` release.

#### Scenario: Git is present
- **GIVEN** `git` is on `PATH`
- **WHEN** a user runs `srekit doctor`
- **THEN** an `ok` finding SHALL report its path and version

#### Scenario: Git is absent
- **GIVEN** `git` is not on `PATH`
- **WHEN** a user runs `srekit doctor`
- **THEN** a `warn` finding SHALL state which inputs fall back to flags and config
- **AND** the command SHALL exit `0`

#### Scenario: Diagnostics work offline
- **WHEN** a user runs `srekit doctor` on a machine with no network connectivity
- **THEN** the command SHALL complete without a network-related finding or delay

### Requirement: Text output

Without `--json`, `doctor` SHALL print findings to standard output as an aligned table, one row per finding, grouped by category, showing the status, the check identifier and the summary. Remedies SHALL be printed on their own continuation line beneath the finding they belong to.

A trailing summary line SHALL report the count of findings per status.

Status SHALL be conveyed by the status word itself, so the output is legible when piped. Colour, if used, SHALL be suppressed when the `NO_COLOR` environment variable is set and non-empty, matching the rest of the CLI.

#### Scenario: Findings are readable when piped
- **WHEN** a user runs `srekit doctor` with standard output redirected to a file
- **THEN** each finding's status SHALL be readable as a word in the text
- **AND** the last line SHALL report the per-status counts

#### Scenario: NO_COLOR is honoured
- **GIVEN** `NO_COLOR=1` is set
- **WHEN** a user runs `srekit doctor`
- **THEN** the output SHALL contain no ANSI colour escapes

### Requirement: JSON output

With `--json`, `doctor` SHALL emit an indented JSON document terminated by a newline to standard output, and SHALL print nothing else there. Keys SHALL be `camelCase`.

The document SHALL be an object carrying an overall status — the worst status among the findings — and a `checks` array. Each entry SHALL carry the check's identifier, category, status, summary, and its remedy when it has one.

The exit status SHALL be the same as for text output, so `--json` can be piped to a parser and still gate CI.

#### Scenario: JSON instead of a table
- **WHEN** a user runs `srekit doctor --json`
- **THEN** standard output SHALL be one indented JSON document and SHALL contain no table rows

#### Scenario: Overall status is the worst finding
- **GIVEN** the findings include at least one `error`
- **WHEN** a user runs `srekit doctor --json`
- **THEN** the document's overall status SHALL be `error`
- **AND** the command SHALL exit `1`

#### Scenario: Keys are camelCase
- **WHEN** a user inspects `srekit doctor --json` output
- **THEN** every object key SHALL be `camelCase`

### Requirement: Quiet reports only what needs attention

With `--quiet`, `doctor` SHALL suppress `ok` findings and the trailing summary line, printing only `warn` and `error` findings. The exit status SHALL be unchanged. In a healthy environment `srekit doctor --quiet` SHALL therefore print nothing and exit `0`.

`--quiet` SHALL have no effect on `--json` output, which is a data document rather than chatter.

#### Scenario: Silence means healthy
- **GIVEN** every check reports `ok`
- **WHEN** a user runs `srekit doctor --quiet`
- **THEN** standard output SHALL be empty
- **AND** the command SHALL exit `0`

#### Scenario: Quiet still shows problems
- **GIVEN** one check reports `error`
- **WHEN** a user runs `srekit doctor --quiet`
- **THEN** that finding SHALL still be printed
- **AND** the command SHALL exit `1`

#### Scenario: Quiet does not truncate JSON
- **WHEN** a user runs `srekit doctor --json --quiet`
- **THEN** the JSON document SHALL contain every finding, including the `ok` ones
