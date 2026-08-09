# template-lifecycle

## Purpose

Gives a team that has customized `srekit`'s templates a way to live with upstream: scaffold a directory, see how it has diverged, and pull in new upstream changes without losing local edits. The hard part is `upgrade` — it does a real three-way merge against a snapshot of "what upstream looked like last time we synced", rather than the usual overwrite-or-skip choice.

## Requirements

### Requirement: Templates subcommand group

The CLI SHALL provide `srekit templates` with the subcommands `init`, `list`, `diff`, `upgrade`, `validate`, `pull`, and `migrate`.

#### Scenario: Subcommands are discoverable
- **WHEN** a user runs `srekit templates --help`
- **THEN** all seven subcommands SHALL be listed

### Requirement: Target directory resolution

Every `templates` subcommand SHALL accept an optional positional `[dir]` argument that takes precedence. Absent that, the directory SHALL be resolved from the same chain as template overrides (`--templates-dir`, `SREKIT_TEMPLATES_DIR`, config file), falling back to the XDG default location.

#### Scenario: Positional argument wins
- **GIVEN** a templates directory is configured in the config file
- **WHEN** a user runs `srekit templates list ./other-templates`
- **THEN** `./other-templates` SHALL be inspected

### Requirement: Scaffolding a templates directory

`templates init` SHALL copy every embedded artifact into the target directory, write the shipped `TEMPLATES.md` placeholder reference, and run `git init` in the directory unless `--no-git` is given. Existing template files SHALL NOT be overwritten unless `--force` is given.

#### Scenario: Fresh scaffold
- **WHEN** a user runs `srekit templates init ./my-templates` on an empty directory
- **THEN** every embedded artifact SHALL be present in the directory
- **AND** `TEMPLATES.md` SHALL be written
- **AND** the directory SHALL be a git repository

#### Scenario: Init refuses to clobber
- **GIVEN** `./my-templates/slo.yaml` exists with local edits
- **WHEN** a user runs `srekit templates init ./my-templates`
- **THEN** the existing `slo.yaml` SHALL be left unchanged

### Requirement: Divergence classification

`templates list` SHALL classify every template — the union of the embedded set and the user directory's contents — into exactly one status:

- `identical` — present in both, byte-equal
- `customized` — present in both, differing
- `user-only` — present only in the user directory
- `embedded-only` — present only in the binary

Classification SHALL be by filename, so a language variant such as `changelog.ru.yaml` is an entry of its own and is never conflated with the base artifact it falls back to. Scaffolding, snapshotting, diffing and upgrading SHALL treat it the same way — as one more file in the embedded set, with its own snapshot under the snapshot directory.

Output SHALL be sorted by template name. `--json` SHALL emit the listing as JSON; `--filter <status>` SHALL restrict the listing to one status.

#### Scenario: Mixed directory is classified
- **GIVEN** a directory with an edited `slo.yaml`, an untouched `rfc.yaml`, and a bespoke `deploy-plan.yaml`
- **WHEN** a user runs `srekit templates list`
- **THEN** `slo.yaml` SHALL be `customized`, `rfc.yaml` SHALL be `identical`, `deploy-plan.yaml` SHALL be `user-only`, and every artifact absent from the directory SHALL be `embedded-only`

#### Scenario: No directory configured
- **WHEN** a user runs `srekit templates list` with no templates directory configured or present
- **THEN** every embedded artifact SHALL be reported as `embedded-only` rather than failing

#### Scenario: A language variant is listed separately
- **GIVEN** a templates directory scaffolded from the embedded set
- **WHEN** a user runs `srekit templates list`
- **THEN** `changelog.yaml` and `changelog.ru.yaml` SHALL each appear as their own entry with their own status

#### Scenario: A customized variant upgrades independently
- **GIVEN** a templates directory with an edited `changelog.ru.yaml` and an untouched `changelog.yaml`
- **WHEN** a user runs `srekit templates upgrade`
- **THEN** the edited variant SHALL be merged against its own snapshot
- **AND** the untouched base artifact SHALL be updated without affecting the variant

### Requirement: Diff against the embedded versions

`templates diff` SHALL show, per customized template, how the user's version differs from the embedded one. `--name-only` SHALL list only the differing filenames. `--no-color` SHALL disable colored output, and color SHALL also be disabled when the `NO_COLOR` environment variable is set and non-empty.

#### Scenario: NO_COLOR is honoured
- **GIVEN** `NO_COLOR=1` is set
- **WHEN** a user runs `srekit templates diff`
- **THEN** the output SHALL contain no ANSI color escapes

### Requirement: Upgrade uses a snapshot as merge base

`templates upgrade` SHALL maintain, per template, a snapshot of the embedded content as of the last successful sync. The snapshot SHALL live in a `.srekit-embedded/` subdirectory of the user's templates directory, and that subdirectory SHALL be added to the directory's `.gitignore` so snapshots never pollute the user's template repository.

#### Scenario: Snapshot directory is gitignored
- **WHEN** `templates upgrade` runs against a directory for the first time
- **THEN** `.srekit-embedded/` SHALL be present in that directory's `.gitignore`

### Requirement: Upgrade resolution rules

For each embedded template, `templates upgrade` SHALL resolve as follows:

| Situation | Action |
|---|---|
| absent from user directory | copy in, record snapshot |
| byte-equal to embedded | leave alone, refresh snapshot |
| `--force` given | overwrite with embedded, record snapshot |
| snapshot equals embedded (upstream unchanged) | leave local edits alone |
| snapshot equals user file (no local edits) | fast-forward to embedded |
| snapshot differs from both (both diverged) | three-way merge |
| no snapshot available | skip, seed snapshot so the next upgrade can merge |

`TEMPLATES.md` SHALL be refreshed on every upgrade. A summary line SHALL report the counts per outcome.

#### Scenario: Upstream unchanged, local edits kept
- **GIVEN** the snapshot matches the current embedded template and the user's file differs from both
- **WHEN** a user runs `srekit templates upgrade`
- **THEN** the user's file SHALL be left byte-identical

#### Scenario: No local edits, upstream moved
- **GIVEN** the snapshot matches the user's file and the embedded template has changed
- **WHEN** a user runs `srekit templates upgrade`
- **THEN** the user's file SHALL be replaced with the new embedded version and reported as updated

#### Scenario: First upgrade on a legacy directory
- **GIVEN** a customized template with no snapshot
- **WHEN** a user runs `srekit templates upgrade`
- **THEN** the file SHALL be skipped with a message explaining that the next upgrade will merge
- **AND** a snapshot SHALL be seeded

### Requirement: Merge conflicts are surfaced, not hidden

When a three-way merge conflicts, the merged file SHALL be written with conflict markers, the snapshot SHALL advance to the current embedded version, and the command SHALL exit non-zero so CI and `git status` flag it. A clean merge SHALL be applied without requiring intervention.

#### Scenario: Conflicting merge exits non-zero
- **GIVEN** both the user's template and the embedded template changed the same region since the snapshot
- **WHEN** a user runs `srekit templates upgrade`
- **THEN** the file SHALL contain conflict markers
- **AND** the command SHALL exit non-zero with a message naming the conflict count

#### Scenario: Clean merge is silent
- **GIVEN** the user's edits and the upstream edits touch different regions
- **WHEN** a user runs `srekit templates upgrade`
- **THEN** the merged result SHALL be written and reported as merged with no conflicts

### Requirement: Upgrade previews with --dry-run

With `--dry-run`, `templates upgrade` SHALL report every action it would take — including merge outcomes — without writing any file, snapshot, or `.gitignore` entry.

#### Scenario: Dry-run leaves the directory untouched
- **WHEN** a user runs `srekit templates upgrade --dry-run`
- **THEN** the reported plan SHALL be printed
- **AND** no file in the templates directory SHALL be modified

### Requirement: Orphaned snapshots are collected

`templates upgrade` SHALL remove snapshot files whose template no longer ships in the binary, so the snapshot directory does not accumulate dead entries across releases. Removal count SHALL be reported. A failure to collect SHALL warn rather than abort the upgrade.

#### Scenario: Retired template's snapshot is removed
- **GIVEN** a snapshot exists for a template that is no longer embedded
- **WHEN** a user runs `srekit templates upgrade`
- **THEN** that snapshot SHALL be deleted and the summary SHALL report the removal

### Requirement: Validation of a templates directory

`templates validate` SHALL parse every recognized template artifact in the directory and report `OK` or `FAIL` per file, exiting non-zero if any failed. It SHALL dispatch by filename: `.sections.yaml` files through the legacy manifest parser, other `.yaml` files through the v1 artifact parser, and `.tmpl` files through Go-template parsing plus a dry-run render. A `.tmpl` whose name matches no built-in template SHALL be reported as parse-only `OK` rather than failed.

#### Scenario: Broken artifact fails the run
- **GIVEN** a directory containing an artifact with a duplicate section id
- **WHEN** a user runs `srekit templates validate`
- **THEN** that file SHALL be reported `FAIL` with the parse error
- **AND** the command SHALL exit non-zero

#### Scenario: Empty directory is not an error
- **GIVEN** a directory with no recognized template artifacts
- **WHEN** a user runs `srekit templates validate`
- **THEN** the command SHALL report that there are no template artifacts and exit zero

### Requirement: Syncing a templates directory from its remote

`templates pull` SHALL run a fast-forward-only `git pull` in the templates directory, or `git pull --rebase` when `--rebase` is given. It SHALL fail with a clear error when the directory is not a git repository.

#### Scenario: Fast-forward-only by default
- **WHEN** a user runs `srekit templates pull` in a templates repository with diverged history
- **THEN** the pull SHALL be refused rather than creating a merge commit

### Requirement: Migration of pre-v1 templates

`templates migrate` SHALL convert legacy `.tmpl` files, and their optional `.sections.yaml` sidecars, into v1 `<name>.yaml` artifacts. It SHALL default to a dry run that prints the conversion, writing files only when `--apply` is given.

Conversion SHALL be heuristic: frontmatter, H1 title and meta bullets are extracted from the header; sections are taken from a parseable sidecar manifest when present, otherwise detected by splitting on `## ` headings. Section types SHALL be inferred — a body containing a Markdown table becomes `table` with columns and rows extracted, a body containing Go-template control flow becomes `text` wrapped in conflict markers for human review, anything else becomes `text` with the body preserved verbatim. Section ids SHALL be slugified from the English portion of a `Русский (English)` heading when present, otherwise from the whole heading.

#### Scenario: Migration defaults to a preview
- **WHEN** a user runs `srekit templates migrate ./old-templates`
- **THEN** the converted YAML SHALL be printed and no file SHALL be written

#### Scenario: Sidecar manifest is authoritative
- **GIVEN** a `postmortem.md.tmpl` with a sibling `postmortem.sections.yaml`
- **WHEN** the pair is migrated
- **THEN** the sidecar's section list SHALL be copied into the result
- **AND** the template body that only looped over those sections SHALL NOT be carried into `header_body`

#### Scenario: Control flow is flagged for review
- **WHEN** a legacy section body contains a `{{ range }}` block
- **THEN** the converted section SHALL be `type: text` with the body wrapped in conflict markers
