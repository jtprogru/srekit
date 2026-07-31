## Why

Three of the eleven generators are not SRE artifacts and do not earn the surface they cost. `retro` is a scrum ceremony template, `capacity` is a planning spreadsheet rendered as prose, and `license` is a leftover of the `lic` command inherited from the gch monolith — nothing about generating a LICENSE file has to do with reliability engineering, and it is solved better by every repository host's own license picker.

The cost is not only conceptual. `license` is the sole reason `--template FILE` exists, and therefore the sole reason the legacy `text/template` render branch is still wired to the CLI: an entire second render path, kept alive for one command that renders three static strings. Retiring these three shrinks the catalog to artifacts that share a definition of what `srekit` is for, and lets the v1 artifact path be the only render path there is.

Doing this before 1.0 is the point. After 1.0 the catalog is frozen and the same cleanup costs a major version.

## What Changes

- **BREAKING** Remove the `capacity` generator and its embedded `capacity.yaml` artifact.
- **BREAKING** Remove the `retro` generator and its embedded `retro.yaml` artifact.
- **BREAKING** Remove the `license` generator, its `lic` alias, and the three inlined license bodies.
- **BREAKING** `--template FILE` ceases to exist on every command, since `license` was the only command that honoured it. A flag no command implements must not be documented, and no command remains whose render path reads a template file.
- Retired command names do not simply vanish into cobra's generic "unknown command" error: invoking `capacity`, `retro`, `license` or `lic` exits non-zero with a message naming the release that removed them and pointing at the migration note. The names stay hidden from `--help` and are not part of the catalog.
- The legacy `text/template` render branch loses its last CLI caller and is removed along with the `--template` plumbing. This is internal, not a public contract — `internal/` is not an API — so it is not itself a breaking change, only an enabled cleanup.
- No change to how any surviving generator behaves. No flag on a surviving command changes name, default, or meaning; the JSON envelope, the v1 artifact format, and the config locations are untouched.
- No dependency change. This removes code and shrinks the binary — three fewer command files, two fewer embedded artifacts, one fewer render path — and adds nothing to the build graph.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `artifact-generation`: the catalog drops three commands; the default-filename table, the optional-input defaults list, and the mandatory-input scenarios lose their entries for them; the license-type validation requirement is removed outright; a new requirement covers how a retired command name behaves.
- `output-routing`: the uniform flag bundle requirement currently carves out `--template` as legitimate on `license`. With `license` gone the carve-out becomes an unconditional prohibition — no command exposes `--template`.

`template-lifecycle` deliberately needs no delta. Its existing requirements already describe the transition correctly: a `capacity.yaml` left in a user's templates directory becomes `user-only` under the divergence classification, and `templates upgrade` already collects the snapshots of templates that no longer ship. That the existing spec absorbs a shrinking embedded set without amendment is a sign it was written at the right level.

## Impact

- Delete `cmd/capacity.go`, `cmd/retro.go`, `cmd/license.go` and their registrations in `NewRootCmd()`; add the retired-name stubs.
- Delete `internal/tmpl/templates/capacity.yaml` and `internal/tmpl/templates/retro.yaml`.
- Remove `--template` plumbing: the template-flag binding and the `TemplatePath` field in the shared output flags, the `TemplatePath` option and the legacy `text/template` branch in the renderer, and the template-file parsing helper that only that branch called.
- `templates validate` keeps its `.tmpl` handling unchanged: bespoke `.tmpl` files remain valid input to `templates migrate`, and since the sample registry is already empty every `.tmpl` is already validated parse-only. No observable change there.
- Existing users' templates directories are not touched. `capacity.yaml` and `retro.yaml` left behind become `user-only` in `templates list` and are simply never loaded; their snapshots are collected by the next `templates upgrade`.
- Docs: delete the three command pages in `docs/en/commands/` and `docs/ru/commands/`, remove their `mkdocs.yml` nav entries, and scrub the three commands from the index, getting-started, recipes, architecture, configuration, custom-templates and json-output pages in **both** locales. Add a migration note in both locales covering what to do instead.
- Update the project descriptions that enumerate the catalog: `README.md`, the root command's short and long help, `.goreleaser.yaml`, `mkdocs.yml`'s site description, `internal/tmpl/TEMPLATES.md`, `CLAUDE.md`, and the `context` block in `openspec/config.yaml`.
- Tests: drop the `capacity`, `retro` and `license` smoke tests, add tests for the retired-name error, and remove the render tests covering the deleted `--template` branch.
- `CHANGELOG.md` entry under `[Unreleased]` → `Removed`, spelling out the breakage and the migration.
