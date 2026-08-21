## ADDED Requirements

### Requirement: Changelog compare links are a document footer

The generated `CHANGELOG.md` SHALL place its link reference definitions — `[Unreleased]` and one entry per released version — in a single contiguous block after the last version section, separated from it by one blank line. Those definitions SHALL NOT be part of any version section's body.

The block SHALL contain a definition for every version heading present in the document. `[Unreleased]` SHALL point at the comparison between the newest released tag and `HEAD`; the initial version SHALL point at its release tag.

#### Scenario: Links sit below the last section
- **WHEN** a user runs `srekit changelog --repo acme/api --stdout`
- **THEN** the output SHALL end with a block of `[label]: <url>` lines
- **AND** those lines SHALL appear after the `## [0.1.0] - <date>` heading and its body

#### Scenario: Editing a version section does not lose the links
- **GIVEN** structured input that replaces the body of the initial-release section
- **WHEN** the document is rendered from that input
- **THEN** the link reference block SHALL still be present with both `[Unreleased]` and the initial version

#### Scenario: Every version has a definition
- **WHEN** a user inspects the generated document
- **THEN** each `## [<version>]` heading SHALL have a matching `[<version>]: ` definition in the footer block

## MODIFIED Requirements

### Requirement: Changelog resolves the repository slug

`changelog` SHALL use `--repo OWNER/REPO` when given, otherwise `meta.repo` from a `--from` payload, otherwise the slug derived from the git remote. When none of the three is available it SHALL fail with an error that names `--repo` as the remedy rather than emitting a document with broken compare links.

#### Scenario: No remote and no flag
- **GIVEN** the working directory is not a git repository with a recognized GitHub remote
- **WHEN** a user runs `srekit changelog`
- **THEN** the command SHALL fail with an error mentioning `pass --repo OWNER/REPO`

#### Scenario: Slug supplied through structured input
- **GIVEN** the working directory is not a git repository with a recognized GitHub remote
- **AND** an input file whose `meta.repo` is `acme/api`
- **WHEN** a user runs `srekit changelog --from <file> --stdout`
- **THEN** the command SHALL succeed and the compare links SHALL name `acme/api`
