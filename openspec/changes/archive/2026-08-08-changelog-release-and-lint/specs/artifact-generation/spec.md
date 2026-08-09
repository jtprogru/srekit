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

A catalog command MAY carry subcommands that maintain the artifact it generates, without those subcommands being catalog entries of their own. The bare invocation SHALL remain the generator, so a user who learned `srekit <name>` keeps the behaviour they learned.

Removing a command from this catalog SHALL be treated as a breaking change.

#### Scenario: Catalog is discoverable
- **WHEN** a user runs `srekit --help`
- **THEN** every command in the catalog SHALL be listed with a one-line description

#### Scenario: A maintenance subcommand is not a catalog entry
- **WHEN** a catalog command carries maintenance subcommands
- **THEN** `srekit --help` SHALL list only the parent command
- **AND** the bare parent invocation SHALL still generate its artifact
