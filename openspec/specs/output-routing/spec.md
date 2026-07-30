# output-routing

## Purpose

Defines the single, uniform way every `srekit` generator decides *where* its rendered output goes and *whether* it is allowed to write there. One flag bundle, one precedence order, one overwrite guard — so a user who learns the flags on `srekit slo` already knows them on `srekit postmortem`.

## Requirements

### Requirement: Uniform output flag bundle

Every generator command SHALL expose the same output flags: `--out`, `--stdout`, `--force`, `--dry-run`, and `--json`. The root command SHALL expose a persistent `--quiet` / `-q` flag inherited by all subcommands.

`--template FILE` SHALL NOT be exposed by commands whose render path ignores it. A flag that is silently ignored is a defect, not a convenience.

#### Scenario: Flag set is identical across generators
- **WHEN** a user runs `srekit <generator> --help` for any generator
- **THEN** `--out`, `--stdout`, `--force`, `--dry-run`, `--json` SHALL all be listed
- **AND** `--out`'s help text SHALL name that command's own default output path

#### Scenario: --template is only offered where it works
- **WHEN** a user runs `srekit license --help`
- **THEN** `--template FILE` SHALL be listed
- **WHEN** a user runs `--help` for any other generator
- **THEN** `--template` SHALL NOT be listed

### Requirement: Output sink precedence

The output sink SHALL be resolved in this order:

1. `--stdout` given, or the resolved target is `-` → standard output
2. `--out PATH` given → that path
3. `--json` given without `--out` → standard output
4. otherwise → the command's default file path

#### Scenario: Default path is used when nothing is specified
- **WHEN** a user runs `srekit slo --service api` with no output flags
- **THEN** the document SHALL be written to `slo-api.md` in the working directory
- **AND** a `wrote slo-api.md` line SHALL be printed

#### Scenario: Dash means stdout
- **WHEN** a user runs `srekit slo --service api --out -`
- **THEN** the document SHALL be written to standard output and no file SHALL be created

#### Scenario: JSON never lands in a .md default
- **WHEN** a user runs `srekit slo --service api --json` with no `--out`
- **THEN** the JSON payload SHALL go to standard output
- **AND** no file named `slo-api.md` SHALL be created

### Requirement: Existing files are never silently overwritten

When the resolved target is a file that already exists and `--force` was not given, the command SHALL fail with a non-zero exit status and an error naming the path and the `--force` remedy. It SHALL NOT modify the existing file.

#### Scenario: Second run without --force is refused
- **GIVEN** `slo-api.md` already exists
- **WHEN** a user runs `srekit slo --service api`
- **THEN** the command SHALL exit non-zero with an error of the form `file slo-api.md already exists; use --force to overwrite`
- **AND** the existing file's contents SHALL be unchanged

#### Scenario: --force overwrites
- **GIVEN** `slo-api.md` already exists
- **WHEN** a user runs `srekit slo --service api --force`
- **THEN** the file SHALL be replaced with the newly rendered document

### Requirement: Dry-run writes nothing

With `--dry-run`, the command SHALL render the document and print it to standard output, prefixed by a comment line naming the target that would have been written, and SHALL NOT create or modify any file. The overwrite guard SHALL NOT apply, because nothing is written.

#### Scenario: Dry-run against an existing file still succeeds
- **GIVEN** `slo-api.md` already exists
- **WHEN** a user runs `srekit slo --service api --dry-run`
- **THEN** the command SHALL exit zero
- **AND** standard output SHALL begin with a `# dry-run: would write <n> bytes to slo-api.md` line followed by the rendered document
- **AND** `slo-api.md` SHALL be unchanged

### Requirement: Parent directories are created on demand

When the resolved target path contains a directory component that does not exist, the command SHALL create the directory tree before writing.

#### Scenario: Nested output path
- **WHEN** a user runs `srekit rfc --title "Move to gRPC" --out docs/rfc/0001.md` and `docs/rfc/` does not exist
- **THEN** the directory tree SHALL be created and the file written into it

### Requirement: Generated files are world-readable

Files written by generators SHALL be created with mode `0644`. Generated artifacts are project documentation (READMEs, runbooks, RFCs) meant to be committed and read by the whole team, not secrets.

#### Scenario: Mode of a generated document
- **WHEN** any generator writes a file
- **THEN** the file's permission bits SHALL be `0644` before the process umask is applied

### Requirement: Quiet suppresses only chatter

With `--quiet`, informational messages SHALL be suppressed: the `wrote <path>` confirmation, dry-run prefix lines, and legacy-template warnings. Rendered document content and error messages SHALL still be emitted.

#### Scenario: Quiet still reports failures
- **GIVEN** `slo-api.md` already exists
- **WHEN** a user runs `srekit slo --service api --quiet`
- **THEN** the command SHALL still exit non-zero with the already-exists error on standard error

#### Scenario: Quiet still emits the document
- **WHEN** a user runs `srekit slo --service api --stdout --quiet`
- **THEN** the rendered document SHALL appear on standard output with no `wrote ...` line

### Requirement: Usage text is shown for misuse, not for failures

An invalid invocation (unknown flag, wrong argument count) SHALL print the command's usage block. A failure raised while executing a valid invocation — missing required input, render error, I/O error — SHALL print only the error message, without the usage block.

#### Scenario: Missing required input prints a bare error
- **WHEN** a user runs `srekit slo` with no `--service`
- **THEN** the output SHALL be the error `--service is required` without the flag listing

#### Scenario: Unknown flag prints usage
- **WHEN** a user runs `srekit slo --service api --nonexistent`
- **THEN** the usage block SHALL be printed alongside the error
