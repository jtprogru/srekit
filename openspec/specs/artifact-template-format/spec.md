# artifact-template-format

## Purpose

Defines the v1 single-file artifact format — one `<name>.yaml` per generator that declares the document's frontmatter, title, meta bullets and typed sections — together with the deterministic rules for composing Markdown from it. This is the format users edit when they customize `srekit`'s output, so it is a public contract, not an internal detail.

## Requirements

### Requirement: Artifact file structure

An artifact SHALL be a single YAML document with these top-level keys:

- `version` (integer, required) — SHALL be `1`
- `frontmatter` (mapping, optional) — emitted verbatim as a YAML frontmatter block
- `title` (string, optional) — the document's H1
- `meta_bullets` (list of strings, optional) — bullet lines directly under the H1
- `header_body` (string, optional) — freeform Markdown between the bullets and the first section
- `sections` (list, required, non-empty) — the typed section slots
- `footer_body` (string, optional) — freeform Markdown after the last section

Keys SHALL use `snake_case`, matching the convention SRE authors already know from Ansible, Kubernetes and Helm.

An artifact that declares no `footer_body` SHALL render exactly as it did before the key existed, so existing artifacts in a user templates directory remain valid without amendment.

#### Scenario: Minimal valid artifact
- **WHEN** an artifact declares `version: 1` and one section with `id`, `title` and `type`
- **THEN** it SHALL parse successfully

#### Scenario: Artifact with a footer
- **WHEN** an artifact declares `footer_body` alongside its sections
- **THEN** it SHALL parse successfully and the footer text SHALL appear in the rendered document

### Requirement: Structural validation on parse

Parsing an artifact SHALL reject, with an actionable error naming the offending element:

- a `version` other than `1`
- an empty or absent `sections` list
- a section with an empty `id`
- two sections sharing an `id`
- a section with an empty `title`
- a section with an absent or unrecognized `type`
- a `frontmatter` value that is not a YAML mapping

#### Scenario: Unsupported version
- **WHEN** an artifact declares `version: 2`
- **THEN** parsing SHALL fail with `unsupported artifact version 2 (expected 1)`

#### Scenario: Duplicate section id
- **WHEN** two sections both declare `id: summary`
- **THEN** parsing SHALL fail with an error naming the duplicate id and its position

#### Scenario: Unknown section type
- **WHEN** a section declares `type: paragraph`
- **THEN** parsing SHALL fail with an error naming the value and listing `text|list|table`

### Requirement: Errors that only matter at render time are deferred

Template-syntax errors inside values and frontmatter scalar-type problems SHALL surface at render time with the offending element's name, not at parse time. An author fixing a broken template gets the error where the context is useful.

#### Scenario: Broken template directive in a section title
- **WHEN** a section title contains `{{ .Meta.Missing` with no closing braces
- **THEN** parsing SHALL succeed but rendering SHALL fail with an error naming that section

### Requirement: Three section types with defined default rendering

A section's `type` SHALL be one of `text`, `list`, or `table`, determining how its declared default content is turned into a Markdown body:

- `text` — `default_body` rendered as-is
- `list` — each entry of `items` rendered as a `- ` bullet; an empty `items` yields a single bare `-` placeholder
- `table` — `columns` and `rows` rendered as a GitHub-flavoured Markdown table; `default_body` is rendered as a prefix paragraph above it; declared columns with no rows yield one empty placeholder row so the table is visually complete

#### Scenario: Empty list section is a fillable placeholder
- **WHEN** a `list` section declares no `items`
- **THEN** its body SHALL be the single line `-`

#### Scenario: Table with columns but no rows
- **WHEN** a `table` section declares three columns and no rows
- **THEN** its body SHALL be a header row, a separator row, and one empty row of three cells

#### Scenario: Short table row is padded
- **WHEN** a `table` row declares fewer cells than there are columns
- **THEN** the missing trailing cells SHALL be rendered as empty rather than shifting the table

### Requirement: Deterministic document composition order

Rendering SHALL compose the Markdown document in exactly this order: frontmatter block, H1 title, meta bullets, header body, one `## <section title>` block per section in declaration order, then the footer body. The document SHALL end with exactly one trailing newline.

#### Scenario: Full document layout
- **WHEN** an artifact declaring frontmatter, title, bullets, header body, two sections and a footer body is rendered
- **THEN** the output SHALL be a `---`-delimited frontmatter block, then `# <title>`, then the bullet lines, then the header body, then `## <first section>` and `## <second section>` in that order, then the footer body

#### Scenario: Section order follows the artifact, not the input
- **WHEN** section bodies are supplied out of declaration order via structured input
- **THEN** the rendered document SHALL still present sections in artifact declaration order

#### Scenario: Footer is last
- **WHEN** an artifact declaring a footer body is rendered
- **THEN** no `## ` heading SHALL appear after the footer text

### Requirement: Empty elements produce no empty markup

Rendering SHALL omit an element entirely when it is absent or whitespace-only: no `---`/`---` block for empty frontmatter, no bare `# ` for an empty title, no blank stanza for an empty header body, no trailing blank stanza for an empty footer body.

#### Scenario: Artifact without frontmatter
- **WHEN** an artifact declares no `frontmatter`
- **THEN** the rendered document SHALL NOT begin with a `---` delimiter

#### Scenario: Artifact without a footer
- **WHEN** an artifact declares no `footer_body`
- **THEN** the rendered document SHALL end with the last section's body and exactly one trailing newline

### Requirement: Blocks are separated by exactly one blank line

Every block of the composed document — frontmatter, title, meta bullets, header body, each section, footer body — SHALL be separated from its neighbour by exactly one blank line. No composition of present and absent elements SHALL produce two consecutive blank lines, and no block SHALL abut the previous one without a separator.

#### Scenario: Header body follows the title with one blank line
- **WHEN** an artifact declaring both a `title` and a `header_body` is rendered
- **THEN** the `# ` heading line SHALL be followed by exactly one empty line before the first line of the header body

#### Scenario: Header body follows frontmatter with one blank line when there is no title
- **GIVEN** an artifact declaring `frontmatter` and `header_body` but no `title`
- **WHEN** it is rendered
- **THEN** the closing `---` SHALL be followed by exactly one empty line before the first line of the header body

#### Scenario: Footer follows the last section with one blank line
- **WHEN** an artifact declaring `footer_body` is rendered
- **THEN** the last line of the final section body SHALL be followed by exactly one empty line before the first line of the footer

### Requirement: Frontmatter key order is preserved

The author's frontmatter key order SHALL survive parse and render unchanged. Frontmatter is human-read; reordering it on every render would make diffs meaningless.

#### Scenario: Key order round-trips
- **GIVEN** an artifact whose frontmatter declares keys in the order `id`, `title`, `date`
- **WHEN** the document is rendered
- **THEN** the emitted frontmatter SHALL list `id`, `title`, `date` in that order

### Requirement: Template directives are evaluated in values, never in keys

Template directives SHALL be evaluated in: frontmatter scalar values (recursing into nested mappings and sequences), the title, each meta bullet, the header body, section titles, section default content, and the footer body. Frontmatter *keys* SHALL be treated as literals. YAML aliases SHALL NOT be traversed.

#### Scenario: Nested frontmatter value is evaluated
- **WHEN** frontmatter contains `meta: {owner: "{{ .Meta.Owner }}"}`
- **THEN** the rendered frontmatter SHALL contain the resolved owner value

#### Scenario: A key that looks like a directive stays literal
- **WHEN** frontmatter declares a key containing `{{ }}`
- **THEN** that key SHALL be emitted verbatim

#### Scenario: Footer directives are resolved
- **WHEN** a footer body contains `{{ .Meta.Repo }}`
- **THEN** the rendered footer SHALL contain the resolved repository slug

### Requirement: Rendering does not mutate the parsed artifact

Rendering an artifact SHALL leave the parsed structure unchanged, so the same artifact can be rendered more than once within a process and yield identical results.

#### Scenario: Repeated render is stable
- **WHEN** the same parsed artifact is rendered twice with the same context
- **THEN** both outputs SHALL be byte-identical

### Requirement: Shared template helper functions

Every template string in an artifact SHALL have access to the same helper set: `default`, `shortID`, `slugify`, `upper`, `lower`, `trim`, and `now` (accepting an optional Go layout string, defaulting to RFC3339).

#### Scenario: Date helper with a custom layout
- **WHEN** a template value contains `{{ now "2006-01-02" }}`
- **THEN** it SHALL render as the current date in `YYYY-MM-DD` form

#### Scenario: Default helper fills a blank
- **WHEN** a template value contains `{{ default "TBD" .Meta.Owner }}` and the owner is empty
- **THEN** it SHALL render as `TBD`

### Requirement: Strings with no directives pass through untouched

A template string containing no `{{` SHALL be emitted verbatim without template evaluation. This keeps literal Markdown — including text that merely resembles template syntax — safe from surprise expansion.

#### Scenario: Literal body is preserved
- **WHEN** a section's default body is plain Markdown with no directives
- **THEN** its rendered body SHALL be byte-identical to the declared body
