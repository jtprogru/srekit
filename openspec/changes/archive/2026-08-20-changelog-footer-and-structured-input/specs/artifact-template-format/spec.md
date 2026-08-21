## ADDED Requirements

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

## MODIFIED Requirements

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
