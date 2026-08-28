## MODIFIED Requirements

### Requirement: Artifact catalog

The CLI SHALL provide exactly these generator commands, each producing one Markdown document:

| Command | Artifact | Mandatory input |
|---|---|---|
| `task` | Investigation log | `--title` |
| `tasker` | Task card for a collection of engineering tasks | `--title` |
| `postmortem` | Google-SRE-style postmortem | `--title` (or `meta.title` via `--from`) |
| `rfc` | RFC / ADR | `--title` |
| `runbook` | Operational runbook | `--title` |
| `oncall-report` | On-call shift report | `--team` |
| `slo` | SLO definition | `--service` |
| `ebp` | Error budget policy | `--service` |
| `changelog` | Keep-a-Changelog scaffold | none (repo auto-detected) |

Every command in the catalog SHALL produce an artifact an on-call engineer or a reliability team owns, with one named exception: `tasker`, whose card describes a task for another engineer rather than work on production. The exception is exhausted by that one entry. A document that belongs to a different discipline — sprint ceremonies, capacity spreadsheets, software licensing — SHALL NOT be added to the catalog, and `tasker` SHALL NOT be read as precedent for admitting one.

`task` and `tasker` SHALL remain separate commands producing unrelated documents. Neither SHALL be an alias, a mode, or a flag of the other.

A catalog command MAY carry subcommands that maintain the artifact it generates, without those subcommands being catalog entries of their own. The bare invocation SHALL remain the generator, so a user who learned `srekit <name>` keeps the behaviour they learned.

Removing a command from this catalog SHALL be treated as a breaking change.

#### Scenario: Catalog is discoverable
- **WHEN** a user runs `srekit --help`
- **THEN** every command in the catalog SHALL be listed with a one-line description

#### Scenario: A maintenance subcommand is not a catalog entry
- **WHEN** a catalog command carries maintenance subcommands
- **THEN** `srekit --help` SHALL list only the parent command
- **AND** the bare parent invocation SHALL still generate its artifact

#### Scenario: The two task commands do not collide
- **WHEN** a user runs `srekit task --title X` and `srekit tasker --title X`
- **THEN** the first SHALL produce an investigation log and the second a task card
- **AND** neither SHALL write to the other's default filename

### Requirement: Deterministic default output filenames

When neither `--out` nor `--stdout` is given, each generator SHALL derive its output filename from its inputs using a slugified form of the identifying value:

| Command | Default filename |
|---|---|
| `task` | `<--path>/investigation-<slug(title)>.md` |
| `tasker` | `tasker-<slug(title)>.md` |
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
- `tasker --topic` defaults to `go`, `--level` to `middle,senior`, `--format` to `code`, `--duration` to `30`

#### Scenario: Minimal SLO invocation is complete
- **WHEN** a user runs `srekit slo --service payments`
- **THEN** the rendered document SHALL contain the `99.9%` availability target, the `30d` window, and the `300ms` p99 latency target

#### Scenario: Minimal task card is complete
- **WHEN** a user runs `srekit tasker --title "Channels and select"`
- **THEN** the card's front matter SHALL carry topic `go`, levels `middle` and `senior`, format `code`, and duration `30`

## ADDED Requirements

### Requirement: Task card shape and inputs

`tasker` SHALL produce a card whose front matter carries `type: simple_note`, the tag `tasker`, and the four values that classify the task: `topic`, `level`, `format` and `duration`. `level` SHALL be emitted as a YAML list and `duration` as a YAML number, because the collection the card joins reads them as such and a quoted scalar is a different value to it.

`--level` SHALL accept repetition and comma-separated values, SHALL trim each value, and SHALL discard blank ones. A `--level` that leaves no value, and a `--duration` that is not a positive number of minutes, SHALL fail with an error naming the flag before anything is written.

The card's two sections SHALL ship with empty bodies. The document is a slot for a task somebody is about to write; declared placeholder text would be deleted on every card.

#### Scenario: Typed front matter values
- **WHEN** a user runs `srekit tasker --title "Channels and select" --stdout`
- **THEN** the front matter SHALL contain `level: [middle, senior]` and `duration: 30`
- **AND** it SHALL NOT contain `duration: "30"`

#### Scenario: Levels given as one comma-separated value
- **WHEN** a user runs `srekit tasker --title X --level junior,middle --stdout`
- **THEN** the front matter SHALL contain `level: [junior, middle]`

#### Scenario: Blank levels are not levels
- **WHEN** a user runs `srekit tasker --title X --level " , "`
- **THEN** the command SHALL exit non-zero with an error naming `--level` and create no file

#### Scenario: Duration is a positive number of minutes
- **WHEN** a user runs `srekit tasker --title X --duration 0`
- **THEN** the command SHALL exit non-zero with an error naming `--duration` and create no file

#### Scenario: Sections are empty
- **WHEN** a user runs `srekit tasker --title X --stdout`
- **THEN** the document SHALL end with the two section headings and no body text between or after them
