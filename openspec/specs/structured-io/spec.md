# structured-io

## Purpose

Makes `srekit` usable as a component rather than only as a typewriter: emit a document's structure as JSON, let a tool or an AI agent fill the section bodies, feed it back, and get Markdown. This is the capability that lets an agent write a postmortem section by section instead of regenerating a whole file and hoping.

## Requirements

### Requirement: Every generator can emit its data as JSON

`--json` SHALL make a generator emit its structured data instead of rendering Markdown. Output SHALL be indented JSON terminated by a newline, and SHALL go to standard output unless `--out` names a file.

JSON object keys SHALL be `camelCase`, even though the YAML artifact format uses `snake_case`.

#### Scenario: JSON instead of Markdown
- **WHEN** a user runs `srekit slo --service api --json`
- **THEN** standard output SHALL be an indented JSON document and SHALL NOT be Markdown

### Requirement: Uniform JSON envelope

The JSON payload SHALL be an object with a `meta` key carrying the command's metadata and a `sections` key carrying the section list. Each entry in `sections` SHALL have `id`, `title`, `type`, `body`, and `required`.

`body` SHALL always be a string. A `list` or `table` section SHALL be pre-rendered into a Markdown fragment, so a consumer sees the same value through the JSON and the Markdown paths.

#### Scenario: Section shape
- **WHEN** a user inspects any generator's `--json` output
- **THEN** each `sections` entry SHALL carry exactly `id`, `title`, `type`, `body`, `required`

#### Scenario: Table body is a Markdown string
- **WHEN** a `table` section is emitted as JSON
- **THEN** its `body` SHALL be a Markdown table as a single string, not a nested array

#### Scenario: Sections follow artifact order
- **WHEN** a user runs `srekit postmortem --title X --json`
- **THEN** the `sections` array SHALL be ordered as declared in the artifact

### Requirement: Section titles agree across both output paths

A section title containing template directives SHALL be evaluated once, before the output path splits, so a title in the JSON payload is identical to the heading in the rendered Markdown.

#### Scenario: Templated title matches
- **GIVEN** a section titled `Timeline (from {{ .Meta.Start }})`
- **WHEN** the same invocation is run once with `--json` and once without
- **THEN** the JSON `title` and the Markdown `## ` heading SHALL carry the same resolved text

### Requirement: Structured input via --from

`postmortem` SHALL accept `--from FILE` carrying a JSON object with optional `meta` and `sections` maps, where `-` reads standard input. Provided section bodies SHALL replace the artifact's defaults; omitted sections SHALL fall back to their defaults.

An empty file SHALL be treated as no input rather than an error. Malformed JSON SHALL fail with an error naming the file.

#### Scenario: Round-trip through JSON
- **WHEN** a user captures `srekit postmortem -T X --json` to a file, edits one section body, and runs `srekit postmortem -T X --from <file>`
- **THEN** the rendered Markdown SHALL contain the edited body for that section and defaults for the rest

#### Scenario: Input from stdin
- **WHEN** a user pipes a JSON payload into `srekit postmortem --from -`
- **THEN** the payload SHALL be read from standard input

### Requirement: Provided bodies are used verbatim

A section body supplied through `--from` SHALL be inserted without template evaluation, so arbitrary Markdown — including text containing `{{ }}` — round-trips unchanged.

#### Scenario: Braces in supplied content survive
- **GIVEN** a supplied section body containing the literal text `{{ .Meta.Owner }}`
- **WHEN** the document is rendered
- **THEN** that text SHALL appear verbatim in the output

### Requirement: Flags outrank the input file

When a value is present both as a CLI flag and under `meta` in the input file, the flag SHALL win. This lets a caller pin a field while reusing an otherwise unchanged payload.

#### Scenario: Flag overrides file metadata
- **GIVEN** an input file whose `meta.title` is `Old title`
- **WHEN** a user runs `srekit postmortem --from <file> --title "New title"`
- **THEN** the rendered document SHALL use `New title`

#### Scenario: Title may come from the file alone
- **GIVEN** an input file whose `meta.title` is set
- **WHEN** a user runs `srekit postmortem --from <file>` with no `--title`
- **THEN** the command SHALL succeed using the file's title

### Requirement: Unknown section ids are rejected

Structured input naming a section id that the artifact does not declare SHALL fail with an error listing the offending ids and the known set. A misspelled id must never disappear silently — that is the failure mode most likely to go unnoticed when an agent generates the payload.

#### Scenario: Typo in a section id
- **GIVEN** an input file with a `sumary` key under `sections`
- **WHEN** a user runs `srekit postmortem --from <file> -T X`
- **THEN** the command SHALL fail with an error naming `sumary` and listing the artifact's known section ids

### Requirement: JSON Schema for the input payload

`postmortem --schema` SHALL emit a JSON Schema (draft 2020-12) describing the `--from` payload. The schema SHALL be derived from the artifact actually in effect, so a user's customized artifact is reflected. It SHALL declare `meta` and `sections` as objects with `additionalProperties: false`, type every section body as `string` with a description naming the section title and type, and list required section ids under `required`.

#### Scenario: Schema reflects a customized artifact
- **GIVEN** a templates directory whose `postmortem.yaml` adds a `blast_radius` section
- **WHEN** a user runs `srekit postmortem --schema --templates-dir <dir>`
- **THEN** the emitted schema SHALL include a `blast_radius` property

#### Scenario: Schema forbids unknown keys
- **WHEN** a user inspects the emitted schema
- **THEN** both the `meta` and `sections` objects SHALL declare `additionalProperties: false`

### Requirement: Validation of an input payload

`postmortem --validate FILE` SHALL check a payload without rendering: reject unknown section ids, then report `OK` or `FAIL` per required section, treating a whitespace-only body as empty. It SHALL exit non-zero when any required section fails.

Validation SHALL NOT evaluate section defaults, so an otherwise-empty payload does not fail on a default that needs metadata it has not been given.

#### Scenario: Missing required section fails
- **GIVEN** an input file that omits a required section
- **WHEN** a user runs `srekit postmortem --validate <file>`
- **THEN** that section SHALL be reported `FAIL  <id>: required body is empty`
- **AND** the command SHALL exit non-zero with a count of failed sections

#### Scenario: Whitespace does not count as content
- **GIVEN** an input file whose required section body is only spaces and newlines
- **WHEN** the payload is validated
- **THEN** that section SHALL be reported `FAIL`

#### Scenario: Complete payload passes
- **GIVEN** an input file with every required section non-empty
- **WHEN** the payload is validated
- **THEN** every required section SHALL be reported `OK` and the command SHALL exit zero

### Requirement: Schema and validate are mutually exclusive

`--schema` and `--validate` SHALL NOT be combined; supplying both SHALL fail with an error saying so.

#### Scenario: Both flags given
- **WHEN** a user runs `srekit postmortem --schema --validate pm.json`
- **THEN** the command SHALL fail with `--schema and --validate are mutually exclusive`

### Requirement: JSON contract stability

The `sections` entry shape and the `--from` payload shape are public API. Within the pre-1.0 series they MAY change with a CHANGELOG entry; from 1.0 they SHALL be stable, and a change to either SHALL be treated as breaking.

#### Scenario: Consumer relies on the section shape
- **WHEN** an external tool reads `.sections[].id` and `.sections[].body` from any generator's JSON output
- **THEN** those keys SHALL be present with those meanings across every generator
