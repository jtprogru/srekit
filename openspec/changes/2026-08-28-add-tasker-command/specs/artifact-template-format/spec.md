## ADDED Requirements

### Requirement: Front matter values may declare a YAML type

A front matter value MAY carry an explicit YAML tag. When it does, and that tag is one of `!!int`, `!!float`, `!!bool`, `!!null`, `!!timestamp`, `!!seq` or `!!map`, the rendered text SHALL be read back as a value of that type and emitted as such. The tag is an instruction to the renderer and SHALL NOT appear in the rendered document.

Without a tag, a rendered value SHALL be emitted as a string exactly as it is today — an artifact that declares no tags SHALL render byte-identically to what it rendered before tags were understood.

An explicit `!!str`, and any tag outside the set above (an application tag such as `!Ref`), SHALL be left alone: the value is rendered as a string and the tag is preserved. An explicit string tag asks for what the untagged path already does, and an application tag's payload belongs to the tool that reads the document.

When the rendered text does not read as the declared type, rendering SHALL fail with an error naming the front matter key and the declared type. Every front matter error SHALL name the key whose value produced it.

#### Scenario: A number stays a number
- **GIVEN** front matter declaring `duration: !!int "{{ .Meta.Duration }}"` and a duration of 30
- **WHEN** the document is rendered
- **THEN** the front matter SHALL contain `duration: 30` and SHALL NOT contain `duration: "30"`

#### Scenario: A list is a list
- **GIVEN** front matter declaring `level: !!seq '[{{ .Meta.Level | join ", " }}]'` and the levels `middle` and `senior`
- **WHEN** the document is rendered
- **THEN** the front matter SHALL contain `level: [middle, senior]`

#### Scenario: The tag does not reach the document
- **WHEN** any tagged front matter value is rendered
- **THEN** the emitted front matter SHALL NOT contain `!!`

#### Scenario: A declared type the value does not have
- **GIVEN** front matter declaring `duration: !!int "{{ .Meta.Duration }}"` and a duration of `half an hour`
- **WHEN** the document is rendered
- **THEN** rendering SHALL fail with an error naming `duration` and `!!int`
- **AND** no file SHALL be written

#### Scenario: An application tag is not reinterpreted
- **GIVEN** front matter declaring `ref: !Ref "{{ .Meta.Name }}"`
- **WHEN** the document is rendered
- **THEN** the emitted front matter SHALL carry the tag and the rendered value as a string

## MODIFIED Requirements

### Requirement: Shared template helper functions

Every template string in an artifact SHALL have access to the same helper set: `default`, `shortID`, `slugify`, `upper`, `lower`, `trim`, `now` (accepting an optional Go layout string, defaulting to RFC3339), and `join` (joining a list of strings with a separator given first, so the pipe form reads naturally).

#### Scenario: Date helper with a custom layout
- **WHEN** a template value contains `{{ now "2006-01-02" }}`
- **THEN** it SHALL render as the current date in `YYYY-MM-DD` form

#### Scenario: Default helper fills a blank
- **WHEN** a template value contains `{{ default "TBD" .Meta.Owner }}` and the owner is empty
- **THEN** it SHALL render as `TBD`

#### Scenario: Join helper assembles a list
- **WHEN** a template value contains `{{ .Meta.Level | join ", " }}` and the levels are `middle` and `senior`
- **THEN** it SHALL render as `middle, senior`
