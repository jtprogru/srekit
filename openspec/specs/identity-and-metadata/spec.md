# identity-and-metadata

## Purpose

Supplies the small facts every generated artifact needs about the world around it: who is writing this, which repository it belongs to, what identifier and filename slug it gets. All of it derived from what the user has already configured — git config, environment, config file — so that `srekit rfc --title X` needs no further ceremony.

## Requirements

### Requirement: Author resolution chain

Author name SHALL be resolved as: the `--author` flag, then the configuration `author` key, then the configuration `full_name` key, then `git config user.name`. Author email SHALL be resolved as: the `--email` flag, then the configuration `email` key, then `git config user.email`.

Values SHALL be trimmed, and a whitespace-only value SHALL be treated as absent.

#### Scenario: Author comes from git config
- **GIVEN** no author flag, config value, or environment variable is set
- **AND** `git config user.name` is `Jane Roe`
- **WHEN** a user runs `srekit rfc --title X`
- **THEN** the document SHALL be attributed to `Jane Roe`

#### Scenario: Whitespace is not a value
- **GIVEN** the configuration sets `author` to a string of spaces
- **WHEN** the author is resolved
- **THEN** resolution SHALL continue to the next source in the chain

### Requirement: Unresolvable author is an actionable error

When a command needs an author and neither name nor email can be resolved, it SHALL fail with an error naming every way to supply the value: the flag, the environment variable, and the git config key.

#### Scenario: No author anywhere
- **GIVEN** no author is configured in any source
- **WHEN** a user runs `srekit rfc --title X`
- **THEN** the command SHALL fail with an error mentioning `--author`, `SREKIT_AUTHOR` and `git user.name`

#### Scenario: Name without email
- **GIVEN** an author name is resolvable but no email is
- **WHEN** a command that needs both runs
- **THEN** it SHALL fail with an error mentioning `--email`, `SREKIT_EMAIL` and `git user.email`

### Requirement: Repository detection from the git remote

Repository owner and name SHALL be derived from `git config --get remote.origin.url`, recognizing GitHub SSH (`git@github.com:OWNER/REPO`) and HTTP(S) (`https://github.com/OWNER/REPO`) forms, with an optional `.git` suffix and an optional trailing slash.

#### Scenario: SSH remote
- **GIVEN** `remote.origin.url` is `git@github.com:jtprogru/srekit.git`
- **WHEN** the repository is detected
- **THEN** the slug SHALL be `jtprogru/srekit`

#### Scenario: HTTPS remote with trailing slash
- **GIVEN** `remote.origin.url` is `https://github.com/jtprogru/srekit/`
- **WHEN** the repository is detected
- **THEN** the slug SHALL be `jtprogru/srekit`

#### Scenario: Unrecognized remote is reported, not guessed
- **GIVEN** `remote.origin.url` points at a non-GitHub host
- **WHEN** the repository is detected
- **THEN** detection SHALL fail with an error quoting the URL

#### Scenario: No remote configured
- **GIVEN** no `remote.origin.url` is set
- **WHEN** the repository is detected
- **THEN** detection SHALL fail with an error saying no remote is configured

### Requirement: Slugs are filesystem-safe and never empty

A slug SHALL be produced by lowercasing the input, replacing every run of characters outside `[a-z0-9]` with a single hyphen, and trimming leading and trailing hyphens. A slug that would be empty SHALL be `untitled`, so a filename is always produced.

#### Scenario: Mixed-case title with punctuation
- **WHEN** `Move Checkout to gRPC!` is slugified
- **THEN** the result SHALL be `move-checkout-to-grpc`

#### Scenario: Non-Latin title
- **WHEN** a title written entirely in Cyrillic is slugified
- **THEN** the result SHALL be `untitled`, and the resulting filename SHALL still be valid

#### Scenario: Consecutive separators collapse
- **WHEN** `a — b   c` is slugified
- **THEN** the result SHALL be `a-b-c` with no repeated hyphens

### Requirement: Artifact identifiers are v4 UUIDs

Generated identifiers SHALL be RFC 4122 version 4 UUIDs in canonical hyphenated lowercase-hex form, with the version and variant bits set correctly.

Artifact identifiers label documents and are not security-sensitive; they SHALL NOT be required to be cryptographically unpredictable, and the implementation is free to avoid the cryptographic random source for binary-size reasons.

#### Scenario: Identifier format
- **WHEN** an artifact identifier is generated
- **THEN** it SHALL match the canonical 8-4-4-4-12 hexadecimal form with version nibble `4` and a variant nibble in `8`–`b`

#### Scenario: Identifiers differ between runs
- **WHEN** two documents are generated without a pinned id
- **THEN** their identifiers SHALL differ

### Requirement: Short identifiers for human-readable references

A short identifier SHALL be the first *n* characters of a full identifier, returning the input unchanged when it is already shorter than *n*, and an empty string when *n* is not positive.

#### Scenario: Short form of a UUID
- **WHEN** the first 8 characters of a UUID are requested
- **THEN** an 8-character prefix SHALL be returned, suitable for a reference like `RFC-1a2b3c4d`

### Requirement: The wall clock is a single injectable source

Every timestamp and date a generator emits SHALL come from one clock source, so that a single invocation cannot straddle a date boundary and produce a document whose filename and content disagree.

#### Scenario: Filename and content dates agree
- **WHEN** a postmortem is generated
- **THEN** the date in its filename SHALL match the generation date recorded in the document

### Requirement: External git invocations are cancellable

Commands that shell out to `git` SHALL do so under the command's context, so that interrupting `srekit` tears down the child process rather than leaving it running.

#### Scenario: Interrupt during a templates pull
- **WHEN** a user interrupts `srekit templates pull` while `git pull` is running
- **THEN** the `git` child process SHALL be terminated

#### Scenario: Second interrupt still kills the process
- **WHEN** a user sends a second interrupt after the first
- **THEN** the process SHALL exit through the default signal behaviour rather than becoming unkillable
