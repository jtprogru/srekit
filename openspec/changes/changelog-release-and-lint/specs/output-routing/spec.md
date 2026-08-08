## MODIFIED Requirements

### Requirement: Uniform output flag bundle

Every generator command SHALL expose the same output flags: `--out`, `--stdout`, `--force`, `--dry-run`, and `--json`. The root command SHALL expose a persistent `--quiet` / `-q` flag inherited by all subcommands.

A command that edits a document the user already owns is not a generator, and SHALL NOT expose `--out` or `--force`. Its destination is the file it was given, so a second destination has no meaning, and an overwrite guard would guard against the command's own purpose. Such a command SHALL still expose `--dry-run`, `--stdout` and `--json` with their usual meanings — show the result, do not write it.

No command SHALL expose `--template FILE`. Every generator resolves its artifact by name through the template source chain, so a template-file flag would be silently ignored wherever it appeared — and a flag that is silently ignored is a defect, not a convenience. A team that wants different output SHALL customize the artifact in a templates directory, which every generator honours.

#### Scenario: Flag set is identical across generators
- **WHEN** a user runs `srekit <generator> --help` for any generator
- **THEN** `--out`, `--stdout`, `--force`, `--dry-run`, `--json` SHALL all be listed
- **AND** `--out`'s help text SHALL name that command's own default output path

#### Scenario: An editing command offers a narrower bundle
- **WHEN** a user runs `--help` for a command that rewrites an existing document
- **THEN** `--dry-run`, `--stdout` and `--json` SHALL be listed
- **AND** `--out` and `--force` SHALL NOT be listed

#### Scenario: No command offers --template
- **WHEN** a user runs `--help` for any command
- **THEN** `--template` SHALL NOT be listed

#### Scenario: --template is rejected as an unknown flag
- **WHEN** a user runs any generator with `--template ./custom.tmpl`
- **THEN** the command SHALL exit non-zero with an unknown-flag error and its usage block
- **AND** no file SHALL be created

### Requirement: Existing files are never silently overwritten

When a generator's resolved target is a file that already exists and `--force` was not given, the command SHALL fail with a non-zero exit status and an error naming the path and the `--force` remedy. It SHALL NOT modify the existing file.

This guard applies to generators, which create documents. It SHALL NOT apply to a command whose stated purpose is to rewrite a document it was pointed at; such a command writes back to its source by design, and guards instead on the preconditions of the edit itself.

#### Scenario: Second run without --force is refused
- **GIVEN** `slo-api.md` already exists
- **WHEN** a user runs `srekit slo --service api`
- **THEN** the command SHALL exit non-zero with an error of the form `file slo-api.md already exists; use --force to overwrite`
- **AND** the existing file's contents SHALL be unchanged

#### Scenario: --force overwrites
- **GIVEN** `slo-api.md` already exists
- **WHEN** a user runs `srekit slo --service api --force`
- **THEN** the file SHALL be replaced with the newly rendered document

#### Scenario: An editing command writes back without --force
- **GIVEN** a `CHANGELOG.md` that already exists
- **WHEN** a user runs a command whose purpose is to rewrite it
- **THEN** the file SHALL be updated in place without requiring `--force`
