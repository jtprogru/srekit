# distribution

## Purpose

Captures the constraints that make `srekit` deployable on an on-call laptop and in CI without ceremony: one self-contained binary, no runtime network dependency, no daemon, no state to manage. These are load-bearing constraints — the project has repeatedly traded convenience for them — so they belong in the spec rather than in folklore.

## Requirements

### Requirement: Self-contained single binary

`srekit` SHALL ship as one statically linked executable with no runtime dependency on an interpreter, shared library, or companion data directory. Every shipped artifact template SHALL be compiled into the binary.

#### Scenario: Works with nothing installed alongside it
- **WHEN** the binary is copied to a machine with no configuration, no templates directory, and no `srekit` installation
- **THEN** every generator SHALL render its document from the embedded templates

### Requirement: No network access at runtime

No command SHALL make a network request. The only external process a command may invoke is `git`, and only for reading local configuration or for the explicitly-git-shaped operations `templates init` and `templates pull`.

#### Scenario: Generation works offline
- **WHEN** any generator is run on a machine with no network connectivity
- **THEN** it SHALL succeed

#### Scenario: Git is only invoked where it is the point
- **WHEN** a generator resolves author metadata or a repository slug
- **THEN** it MAY read local `git config`, but SHALL NOT contact a remote

### Requirement: Supported platforms

Release artifacts SHALL be published for Linux, macOS and FreeBSD, on both `amd64` and `arm64`.

#### Scenario: Platform coverage of a release
- **WHEN** a release is published
- **THEN** binaries SHALL be available for all six platform/architecture combinations

### Requirement: Release artifacts are verifiable

Each release SHALL publish a checksum file, and that checksum file SHALL be signed, so a downloaded binary can be verified without trusting the download path.

#### Scenario: Verifying a download
- **WHEN** a user downloads a release binary and its checksum file
- **THEN** the checksum file's signature SHALL be verifiable against the project's published key

### Requirement: Installation channels

`srekit` SHALL be installable via a Homebrew tap, via `go install`, and by downloading a release binary directly. Uninstallation SHALL require removing only the binary; configuration and user templates SHALL be separate and independently removable.

#### Scenario: Removing the binary leaves no orphaned state
- **WHEN** a user deletes the `srekit` binary
- **THEN** nothing SHALL remain that affects other tools
- **AND** the user's configuration file and templates directory SHALL remain until deleted explicitly

### Requirement: Binary size is a tracked constraint

Dependency additions SHALL be weighed against binary size. A dependency that pulls in the cryptographic or HTTP stacks SHALL NOT be added for functionality that does not need them.

#### Scenario: Rejecting a heavyweight dependency
- **WHEN** a proposed change would add a dependency that transitively links the TLS or FIPS module for non-security-relevant functionality
- **THEN** the change SHALL either be rejected or justified explicitly in the CHANGELOG

### Requirement: Exit statuses distinguish success from failure

Every command SHALL exit zero on success and non-zero on failure, so `srekit` composes in shell pipelines and CI without output parsing. Commands that report per-item results (`templates validate`, `templates upgrade`, `postmortem --validate`) SHALL exit non-zero when any item failed.

#### Scenario: Failure is detectable without parsing output
- **WHEN** any command fails for any reason
- **THEN** its exit status SHALL be non-zero

#### Scenario: Partial failure fails the run
- **GIVEN** a templates directory where one of five artifacts is invalid
- **WHEN** a user runs `srekit templates validate`
- **THEN** the command SHALL exit non-zero even though four artifacts passed

### Requirement: Interrupt handling

An interrupt or termination signal SHALL cancel the running command's context so that child processes are torn down. A second signal SHALL fall through to the runtime's default handling.

#### Scenario: Ctrl-C during a git operation
- **WHEN** a user interrupts a command that has spawned `git`
- **THEN** the `git` process SHALL be terminated before `srekit` exits
