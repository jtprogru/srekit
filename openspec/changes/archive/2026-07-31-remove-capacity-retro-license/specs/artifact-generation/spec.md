## ADDED Requirements

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

## MODIFIED Requirements

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

## REMOVED Requirements

### Requirement: Unrecognized license types are rejected

**Reason**: The `license` command is removed from the catalog. Generating a LICENSE file is not a reliability-engineering artifact — it is a one-time repository setup step already served by every code host's license picker — and it was the only command whose render path read a template file, keeping a second render path alive for three static strings.

**Migration**: Use the license picker offered by your code host, or copy the license text once from `choosealicense.com` and commit it. Users who need `srekit` to keep emitting it can pin the last release that shipped the command; the removal release names it in the changelog entry. There is no templates-directory workaround: with the command gone, a `license` template in a user directory has nothing to render it.
