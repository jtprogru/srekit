# artifact-generation

## Purpose

Defines the catalog of SRE text artifacts `srekit` generates, and the contract each generator obeys: which inputs are mandatory, what the default output filename looks like, and what the resulting Markdown document contains. This is the capability users actually reach for — everything else in the tool exists to serve it.

## Requirements

### Requirement: Artifact catalog

The CLI SHALL provide exactly these generator commands, each producing one Markdown document:

| Command | Artifact | Mandatory input |
|---|---|---|
| `task` | Investigation log | `--title` |
| `postmortem` | Google-SRE-style postmortem | `--title` (or `meta.title` via `--from`) |
| `rfc` | RFC / ADR | `--title` |
| `runbook` | Operational runbook | `--title` |
| `oncall-report` | On-call shift report | `--team` |
| `slo` | SLO definition | `--service` |
| `ebp` | Error budget policy | `--service` |
| `changelog` | Keep-a-Changelog scaffold | none (repo auto-detected) |

Every command in the catalog SHALL produce an artifact an on-call engineer or a reliability team owns. A document that belongs to a different discipline — sprint ceremonies, capacity spreadsheets, software licensing — SHALL NOT be added to it.

Removing a command from this catalog SHALL be treated as a breaking change.

#### Scenario: Catalog is discoverable
- **WHEN** a user runs `srekit --help`
- **THEN** every command in the catalog SHALL be listed with a one-line description

### Requirement: Retired command names fail with an explanation

A command name that has been removed from the catalog SHALL NOT fall through to the generic unknown-command error. Invoking `capacity`, `retro`, `license`, or the `lic` alias SHALL exit non-zero with a message naming the release that removed the command and pointing at the migration note, and SHALL write no file.

Retired names SHALL NOT appear in `--help`, SHALL NOT be part of the catalog, and SHALL NOT accept or validate any flags — a retired command fails before looking at its arguments.

This is a courtesy with an expiry date: the retired names SHALL be dropped entirely at 1.0, at which point they become ordinary unknown commands.

#### Scenario: Retired command names the remedy
- **WHEN** a user runs `srekit capacity --service payments`
- **THEN** the command SHALL exit non-zero with a message naming the release that removed `capacity` and where to read about the replacement
- **AND** no file SHALL be created

#### Scenario: Retired alias behaves like the command
- **WHEN** a user runs `srekit lic --type mit`
- **THEN** the command SHALL exit non-zero with the same explanation given for `license`

#### Scenario: Retired names are not advertised
- **WHEN** a user runs `srekit --help`
- **THEN** `capacity`, `retro`, `license`, and `lic` SHALL NOT be listed

#### Scenario: Flags are not consulted
- **WHEN** a user runs `srekit retro` with no `--team`
- **THEN** the error SHALL be the removal message, not `--team is required`

### Requirement: Mandatory inputs are enforced before any side effect

When a generator's mandatory input is absent, the command SHALL exit non-zero with an error naming the missing flag, before rendering or touching the filesystem.

#### Scenario: Missing title
- **WHEN** a user runs `srekit rfc` with no `--title`
- **THEN** the command SHALL fail with `--title is required` and create no file

#### Scenario: Missing team
- **WHEN** a user runs `srekit oncall-report` with no `--team`
- **THEN** the command SHALL fail with `--team is required` and create no file

#### Scenario: Missing service
- **WHEN** a user runs `srekit ebp` with no `--service`
- **THEN** the command SHALL fail with `--service is required` and create no file

### Requirement: Deterministic default output filenames

When neither `--out` nor `--stdout` is given, each generator SHALL derive its output filename from its inputs using a slugified form of the identifying value:

| Command | Default filename |
|---|---|
| `task` | `<--path>/investigation-<slug(title)>.md` |
| `postmortem` | `postmortem-<YYYY-MM-DD>-<slug(title)>.md` |
| `rfc` | `rfc-<slug(title)>.md` |
| `runbook` | `runbook-<slug(title)>.md` |
| `oncall-report` | `oncall-<slug(team)>-<slug(start)>.md` |
| `slo` | `slo-<slug(service)>.md` |
| `ebp` | `ebp-<slug(service)>.md` |
| `changelog` | `CHANGELOG.md` |

Every generator SHALL default to a file. No generator SHALL default to standard output.

#### Scenario: Title becomes a slugged filename
- **WHEN** a user runs `srekit rfc --title "Move Checkout to gRPC"`
- **THEN** the document SHALL be written to `rfc-move-checkout-to-grpc.md`

#### Scenario: Postmortem filename carries the date
- **GIVEN** the current date is 2026-07-30
- **WHEN** a user runs `srekit postmortem --title "Cache stampede"`
- **THEN** the document SHALL be written to `postmortem-2026-07-30-cache-stampede.md`

#### Scenario: Every generator writes a file by default
- **WHEN** any generator in the catalog is run with its mandatory input and no output flags
- **THEN** it SHALL write a file rather than printing to standard output

### Requirement: Generators supply sensible defaults for optional inputs

Optional inputs SHALL have documented defaults so a generator is useful with only its mandatory flag:

- `postmortem --severity` defaults to `SEV-3`
- `rfc --status` defaults to `proposed`, accepting `proposed | accepted | rejected | superseded | deprecated`
- `slo --target` defaults to `99.9%`, `--window` to `30d`, `--latency` to `300ms`
- `changelog --version` defaults to `0.1.0`
- `task --path` defaults to the working directory

#### Scenario: Minimal SLO invocation is complete
- **WHEN** a user runs `srekit slo --service payments`
- **THEN** the rendered document SHALL contain the `99.9%` availability target, the `30d` window, and the `300ms` p99 latency target

### Requirement: On-call report defaults to the current ISO week

When `oncall-report` is given neither `--start` nor `--end`, the reporting period SHALL default to the current ISO-8601 week: Monday through Sunday, with Sunday treated as the last day of the week rather than the first.

#### Scenario: Sunday belongs to the week that is ending
- **GIVEN** the current date is Sunday 2026-07-26
- **WHEN** a user runs `srekit oncall-report --team platform`
- **THEN** the period SHALL be `2026-07-20` to `2026-07-26`

#### Scenario: Partially specified period
- **WHEN** a user runs `srekit oncall-report --team platform --start 2026-07-01`
- **THEN** `--start` SHALL be honoured verbatim and only `--end` SHALL be derived from the current week

### Requirement: Changelog resolves the repository slug

`changelog` SHALL use `--repo OWNER/REPO` when given, otherwise derive the slug from the git remote. When neither is available it SHALL fail with an error that names `--repo` as the remedy rather than emitting a document with broken compare links.

#### Scenario: No remote and no flag
- **GIVEN** the working directory is not a git repository with a recognized GitHub remote
- **WHEN** a user runs `srekit changelog`
- **THEN** the command SHALL fail with an error mentioning `pass --repo OWNER/REPO`

### Requirement: Rendered documents are bilingual by default

Shipped artifact templates other than `changelog` SHALL use `Русский (English)` headings and labels with Russian body prose. Technical identifiers (SLO, SLI, RFC, SEV, UTC, PromQL expressions) and YAML frontmatter keys SHALL remain English. The `changelog` artifact SHALL be entirely English so that tooling built around Keep a Changelog keeps working.

#### Scenario: Section headings carry both languages
- **WHEN** a user renders any artifact other than `changelog`
- **THEN** its section headings SHALL follow the `Русский (English)` form

#### Scenario: Changelog stays English
- **WHEN** a user runs `srekit changelog`
- **THEN** the rendered document SHALL contain only English headings (`Added`, `Changed`, `Fixed`, `Unreleased`)

### Requirement: Every generated document carries an identifier and a generation timestamp

Generators SHALL stamp each document with a generated UUIDv4 (or a short form derived from it) and the generation time, so an artifact can be referenced and dated after the fact. Artifact identifiers are labels, not security tokens.

#### Scenario: Identifier is present
- **WHEN** a user renders a postmortem
- **THEN** the document SHALL contain a v4 UUID identifying it

#### Scenario: Identifier can be pinned
- **WHEN** a user supplies `meta.id` via `--from`
- **THEN** that value SHALL be used instead of a freshly generated UUID, so a document can be re-rendered with a stable identity
