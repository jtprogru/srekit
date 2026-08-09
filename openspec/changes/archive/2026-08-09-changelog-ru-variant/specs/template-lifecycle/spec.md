## MODIFIED Requirements

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
