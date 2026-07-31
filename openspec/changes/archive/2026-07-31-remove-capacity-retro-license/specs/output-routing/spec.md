## MODIFIED Requirements

### Requirement: Uniform output flag bundle

Every generator command SHALL expose the same output flags: `--out`, `--stdout`, `--force`, `--dry-run`, and `--json`. The root command SHALL expose a persistent `--quiet` / `-q` flag inherited by all subcommands.

No command SHALL expose `--template FILE`. Every generator resolves its artifact by name through the template source chain, so a template-file flag would be silently ignored wherever it appeared — and a flag that is silently ignored is a defect, not a convenience. A team that wants different output SHALL customize the artifact in a templates directory, which every generator honours.

#### Scenario: Flag set is identical across generators
- **WHEN** a user runs `srekit <generator> --help` for any generator
- **THEN** `--out`, `--stdout`, `--force`, `--dry-run`, `--json` SHALL all be listed
- **AND** `--out`'s help text SHALL name that command's own default output path

#### Scenario: No command offers --template
- **WHEN** a user runs `--help` for any command
- **THEN** `--template` SHALL NOT be listed

#### Scenario: --template is rejected as an unknown flag
- **WHEN** a user runs any generator with `--template ./custom.tmpl`
- **THEN** the command SHALL exit non-zero with an unknown-flag error and its usage block
- **AND** no file SHALL be created
