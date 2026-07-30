# user-configuration

## Purpose

Lets a user set once — author name, email, templates directory — what they would otherwise pass on every invocation, without ever putting them in a position where a config file they wrote is silently ignored. Fresh installs follow the XDG Base Directory spec; installs predating that keep working unchanged.

## Requirements

### Requirement: Configuration is optional

Every setting SHALL be supplyable by flag or environment variable. A missing, unreadable, or malformed configuration file SHALL NOT fail a command that has everything it needs from other sources.

#### Scenario: No config file at all
- **WHEN** a user runs `srekit slo --service api` with no configuration file present
- **THEN** the command SHALL succeed

#### Scenario: Malformed config file does not break generation
- **GIVEN** a configuration file containing invalid YAML
- **WHEN** a user runs a generator that needs nothing from it
- **THEN** the command SHALL succeed

### Requirement: Value precedence

A configuration value SHALL be resolved in this order: an explicit CLI flag, then the `SREKIT_<KEY>` environment variable (uppercased key), then the configuration file.

#### Scenario: Environment beats file
- **GIVEN** a config file setting `author: File Author`
- **AND** `SREKIT_AUTHOR=Env Author` in the environment
- **WHEN** a generator resolves the author
- **THEN** `Env Author` SHALL be used

#### Scenario: Flag beats environment
- **GIVEN** `SREKIT_AUTHOR=Env Author` in the environment
- **WHEN** a user runs a generator with `--author "Flag Author"`
- **THEN** `Flag Author` SHALL be used

### Requirement: Configuration file location

When `--config` is not given, the configuration file SHALL be resolved as: the legacy `~/.srekit.yaml` if that file already exists, otherwise `$XDG_CONFIG_HOME/srekit/config.yaml` (defaulting to `~/.config/srekit/config.yaml` when `XDG_CONFIG_HOME` is unset).

Existing-legacy-wins is deliberate: a user who upgrades must never end up with a config file that is present but no longer read.

#### Scenario: Legacy file wins when present
- **GIVEN** `~/.srekit.yaml` exists
- **WHEN** any command resolves its configuration
- **THEN** `~/.srekit.yaml` SHALL be read, not the XDG path

#### Scenario: Fresh install uses XDG
- **GIVEN** no `~/.srekit.yaml` exists
- **AND** `XDG_CONFIG_HOME` is set to `/custom/config`
- **WHEN** any command resolves its configuration
- **THEN** `/custom/config/srekit/config.yaml` SHALL be read

#### Scenario: Explicit path overrides resolution
- **WHEN** a user runs a command with `--config ./project.yaml`
- **THEN** that file SHALL be read regardless of what exists in the home directory

### Requirement: Default templates directory location

The default templates directory SHALL be resolved the same way: the legacy `~/.srekit/templates` if that directory already exists, otherwise `$XDG_CONFIG_HOME/srekit/templates`.

#### Scenario: Legacy templates directory keeps working
- **GIVEN** `~/.srekit/templates` exists
- **WHEN** a `templates` subcommand resolves its target with no explicit directory
- **THEN** `~/.srekit/templates` SHALL be used

### Requirement: Only scalar configuration values are read

The configuration file SHALL be a flat YAML mapping. Keys whose values are strings, integers, floats or booleans SHALL be read and exposed as strings; keys with nested or list values SHALL be ignored rather than causing a failure.

#### Scenario: Nested key is ignored
- **GIVEN** a config file containing both `author: Jane` and a nested `extras:` mapping
- **WHEN** the configuration is loaded
- **THEN** `author` SHALL be available and the nested key SHALL be ignored without error

### Requirement: Recognized configuration keys

The configuration file SHALL recognize at least `author`, `full_name`, `email`, and `templates_dir`. `full_name` SHALL act as a fallback for `author`.

#### Scenario: full_name serves as the author
- **GIVEN** a config file setting `full_name` but not `author`
- **WHEN** a generator resolves the author
- **THEN** the `full_name` value SHALL be used

### Requirement: Interactive configuration bootstrap

`srekit config init` SHALL create the configuration file, prompting for author, email and templates directory. Defaults offered at each prompt SHALL come from the corresponding `git config` value where one exists (`user.name`, `user.email`). Flags SHALL pre-supply any answer.

#### Scenario: Prompts are prefilled from git
- **GIVEN** `git config user.name` is `Jane Roe`
- **WHEN** a user runs `srekit config init` interactively
- **THEN** the author prompt SHALL offer `Jane Roe` as the default

#### Scenario: Values can be passed as flags
- **WHEN** a user runs `srekit config init --author Jane --email jane@example.com --yes`
- **THEN** the config file SHALL be written with those values and no prompt SHALL be shown

### Requirement: Non-interactive configuration bootstrap

`config init` SHALL accept defaults without prompting when `--yes` is given, and SHALL also avoid prompting when standard input is not a terminal, so it is safe to run in scripts and CI.

#### Scenario: Piped stdin does not hang
- **WHEN** `srekit config init` is run with standard input redirected from a non-terminal
- **THEN** the command SHALL complete without waiting for input

### Requirement: config init does not clobber an existing file

When the target configuration file already exists, `config init` SHALL refuse to write unless `--force` is given.

#### Scenario: Existing config is protected
- **GIVEN** a configuration file already exists
- **WHEN** a user runs `srekit config init`
- **THEN** the command SHALL fail and leave the existing file unchanged

#### Scenario: Force rewrites it
- **WHEN** a user runs `srekit config init --force --yes`
- **THEN** the configuration file SHALL be replaced

### Requirement: config init writes where the config would be read from

`config init` SHALL write to the same path that configuration resolution would read, honouring `--config`. Creating a file the tool would not read is a defect.

#### Scenario: Explicit path is respected
- **WHEN** a user runs `srekit config init --config ./team.yaml --yes`
- **THEN** `./team.yaml` SHALL be created

### Requirement: Version and build metadata are reportable

The CLI SHALL report its version, source commit, build date, and builder via `--version` / `-V`, so a user filing an issue can state exactly which binary they ran.

#### Scenario: Version output
- **WHEN** a user runs `srekit --version`
- **THEN** the output SHALL include the version, commit, build date and builder
