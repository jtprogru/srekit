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
| `capacity` | Capacity plan | `--service` |
| `retro` | Sprint / period retrospective | `--team` |
| `changelog` | Keep-a-Changelog scaffold | none (repo auto-detected) |
| `license` | LICENSE file | none (defaults to WTFPL) |

Removing a command from this catalog SHALL be treated as a breaking change.

#### Scenario: Catalog is discoverable
- **WHEN** a user runs `srekit --help`
- **THEN** every command in the catalog SHALL be listed with a one-line description

### Requirement: Mandatory inputs are enforced before any side effect

When a generator's mandatory input is absent, the command SHALL exit non-zero with an error naming the missing flag, before rendering or touching the filesystem.

#### Scenario: Missing title
- **WHEN** a user runs `srekit rfc` with no `--title`
- **THEN** the command SHALL fail with `--title is required` and create no file

#### Scenario: Missing team
- **WHEN** a user runs `srekit retro` with no `--team`
- **THEN** the command SHALL fail with `--team is required` and create no file

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
| `capacity` | `capacity-<slug(service)>.md` |
| `retro` | `retro-<slug(team)>-<slug(sprint)>.md` |
| `changelog` | `CHANGELOG.md` |
| `license` | standard output |

#### Scenario: Title becomes a slugged filename
- **WHEN** a user runs `srekit rfc --title "Move Checkout to gRPC"`
- **THEN** the document SHALL be written to `rfc-move-checkout-to-grpc.md`

#### Scenario: Postmortem filename carries the date
- **GIVEN** the current date is 2026-07-30
- **WHEN** a user runs `srekit postmortem --title "Cache stampede"`
- **THEN** the document SHALL be written to `postmortem-2026-07-30-cache-stampede.md`

#### Scenario: License defaults to stdout, not a file
- **WHEN** a user runs `srekit license --type mit` with no `--out`
- **THEN** the license text SHALL be printed to standard output and no `LICENSE` file SHALL be created

### Requirement: Generators supply sensible defaults for optional inputs

Optional inputs SHALL have documented defaults so a generator is useful with only its mandatory flag:

- `postmortem --severity` defaults to `SEV-3`
- `rfc --status` defaults to `proposed`, accepting `proposed | accepted | rejected | superseded | deprecated`
- `slo --target` defaults to `99.9%`, `--window` to `30d`, `--latency` to `300ms`
- `capacity --horizon` defaults to `1y`
- `changelog --version` defaults to `0.1.0`
- `retro --sprint` defaults to today's date
- `license --type` defaults to `wtfpl`, accepting `wtfpl | mit | apache2`
- `license --year` defaults to the current year
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

### Requirement: Unrecognized license types are rejected

`license --type` SHALL accept only the supported identifiers and SHALL fail with an error listing them otherwise. It SHALL NOT fall back to a default license, because silently shipping the wrong license is worse than failing.

#### Scenario: Unknown license type
- **WHEN** a user runs `srekit license --type gpl3`
- **THEN** the command SHALL fail with an error naming `wtfpl, mit, apache2` and write nothing

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
