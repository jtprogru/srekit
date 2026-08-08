## Why

The `changelog` generator claims Keep a Changelog conformance and mostly earns it, but three defects sit between the claim and the reality.

The compare-link block — `[Unreleased]: …/compare/v0.1.0...HEAD` and `[0.1.0]: …/releases/tag/v0.1.0` — is buried inside the body of the `initial_release` section, because the v1 artifact format has nowhere else to put trailing content. Keep a Changelog treats those link definitions as a document-level footer, not as prose belonging to the first release. Anything that later wants to rewrite that block on release has to reach into a section body and guess where the prose ends, and any user who edits `initial_release` through structured input silently loses the links.

`changelog` is also the one generator whose `--json` output cannot be fed back. It emits the full `{meta, sections}` envelope, but the command hardcodes `nil` where every other structured path passes overrides, so the round-trip that `postmortem` documents simply does not work here — the flag set advertises structure the command does not accept.

Finally, every artifact that declares `header_body` renders a stray extra blank line between the H1 and that body, because the title block already ends the paragraph and the header body opens with another separator. Cosmetic in a browser, noisy in `git diff`, and wrong in a format whose whole point is being read by humans as plain text.

Doing this before 1.0 is deliberate: `footer_body` is an addition to the v1 artifact format, and the format freezes at 1.0.

## What Changes

- Add an optional `footer_body` key to the v1 artifact format — the mirror of `header_body`, rendered after the last section. Additive: an artifact that omits it renders exactly as before.
- Move the changelog's compare-link definitions out of the `initial_release` section body into `footer_body`, making the link block a document-level element that can be located and rewritten mechanically.
- Fix the double blank line between the H1 and `header_body`. Block separation becomes exactly one blank line everywhere in the composed document. This changes the rendered bytes of every artifact that declares a `header_body`, so it is a visible behaviour change, though not a contract break — no heading, no section body and no JSON key moves.
- Extend `--from FILE` to `changelog`, so the `--json` → edit → `--from` round-trip works there as it does for `postmortem`. Unknown section IDs stay an error; provided bodies stay verbatim.
- `--schema` and `--validate` stay `postmortem`-only. The changelog artifact declares no required sections, so validating a payload against it can only ever pass, and a schema describing two string fields is noise rather than tooling. Parity is not a goal in itself — a flag that cannot fail is as useless as one that is ignored.
- No new dependencies. `footer_body` is one more string field on an already-parsed YAML document, and the structured-input path reuses the existing JSON decode and merge; nothing enters the build graph.

Nothing here is **BREAKING**. The command set is unchanged, no flag changes name or meaning, the JSON envelope keeps its exact shape, config locations are untouched, and the format change is a new optional key.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `artifact-template-format`: the top-level key list gains `footer_body`; the composition-order requirement gains the footer as its final block; the empty-elements requirement covers an absent footer; a new requirement pins block separation at exactly one blank line, which is what the H1/header-body defect violates.
- `artifact-generation`: the changelog requirement gains the location of the compare-link block, so a document-level footer rather than section prose is the specified behaviour, not an implementation accident.
- `structured-io`: the `--from` requirement currently names `postmortem` as the only command that accepts structured input and is widened to name `changelog` as well; the metadata-precedence requirement gains the changelog's repository slug. The `--schema` and `--validate` requirements are untouched — they stay `postmortem`-only.

## Impact

- `internal/sections/artifact.go` — `footer_body` on the parsed artifact.
- `internal/sections/render_artifact.go` — emit the footer as the final block; stop writing a leading separator when the buffer already ends in a blank line.
- `internal/tmpl/templates/changelog.yaml` — link definitions move from `initial_release.default_body` to `footer_body`.
- `internal/tmpl/TEMPLATES.md` — document `footer_body` alongside `header_body`.
- `cmd/changelog.go` — `--from` flag, structured input read, overrides passed to the merge instead of `nil`.
- The payload-reading helper currently living in `cmd/postmortem.go` is shared by two commands after this change and moves to a neutral home. Internal only; `internal/` and unexported `cmd` helpers are not public API.
- Tests: golden render output for every artifact declaring a `header_body` shifts by one blank line; `internal/sections` gains footer coverage; `cmd/cmd_test.go` gains changelog round-trip cases.
- Docs, both locales: `docs/{en,ru}/commands/changelog.md` gains the structured-input section, the custom-templates and JSON-output pages gain `footer_body`, and any embedded sample output loses the extra blank line.
- `CHANGELOG.md` under `[Unreleased]`: `Added` for `footer_body` and `changelog --from`, `Changed` for the blank-line fix and the link-block move.
