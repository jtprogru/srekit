## Context

See proposal.md — Why. Three constraints shape the approach.

The v1 artifact format is a public contract that freezes at 1.0, so `footer_body` has to be additive in the strict sense: an artifact that never heard of the key parses and renders byte-identically. That rules out anything clever like a magic section id.

The renderer is the single composition point for every artifact, so the blank-line fix is one change with a wide blast radius: every golden test and every embedded sample output that includes a `header_body` shifts. There is no way to scope it to `changelog` and no reason to want one — the extra line is wrong everywhere.

The structured path already exists in generic form. `sections.Merge` takes an overrides map and rejects unknown ids regardless of which artifact it was handed. The only thing that is postmortem-specific today is the small payload-reading and flag-wiring layer sitting in `cmd/postmortem.go`. This change is mostly relocation, not new machinery.

## Goals / Non-Goals

**Goals:**

- `footer_body` as a first-class element of the v1 format, symmetric with `header_body` in parsing, template evaluation and omission rules.
- The changelog's link block addressable as a document-level element rather than as prose inside a section body.
- One blank line between every pair of adjacent blocks, in every combination of present and absent elements.
- `changelog` accepting back the structure it already emits.

**Non-Goals:**

- Rewriting an existing `CHANGELOG.md`. Nothing here reads a user's changelog; the link block only becomes *addressable* so a later change can rewrite it.
- Generalizing `--from` to every generator. It is added where a round-trip is meaningful, and for the rest the question stays open.
- Parsing or validating Keep a Changelog structure. `--validate` here validates a `--from` payload against the artifact, which is a different thing from validating a changelog document.
- Localization of the changelog artifact.

## Decisions

**A dedicated `footer_body` key, not a section with a blank title or a `position: footer` attribute on a section.**

A section carries an id, a title, a type and a required flag, and every one of those is meaningless for a link block. Modelling the footer as a section would mean either a section that renders no `## ` heading — a special case in composition, in the JSON envelope, and in the schema — or a section id that consumers see in `--json` and could target through `--from`, which invites someone to replace the link block with free text and get a document with broken references. `footer_body` keeps the footer out of the section list entirely, so the JSON contract and the section-id namespace are untouched. The cost is one more optional key in the format; the alternative costs a special case in four places.

**The blank-line fix belongs in the composer, expressed as a precondition rather than as per-block padding.**

The current code has each block write its own leading or trailing separators, which is exactly why the H1's trailing `\n\n` and the header body's leading `\n` add up to three newlines. The fix is a single helper that starts a block by ensuring the buffer ends in exactly one blank line (nothing to do when the buffer is empty), and every block uses it. That makes the invariant structural instead of a property to be re-derived at each call site, and it is why the new spec requirement is phrased as "no composition produces two consecutive blank lines" rather than as a statement about the title.

**Frontmatter is a block like any other under that rule.** Today it writes `---\n\n` and then the title writes its own content — correct by accident because the title block does not open with a separator. Routing frontmatter through the same helper keeps it correct on purpose, including the case where there is no title and the header body follows the frontmatter directly.

**The payload reader moves to `internal/sections`, not to a new `cmd` file.**

The reader and the meta-precedence helper are currently unexported functions in `cmd/postmortem.go`. Two commands need them, and they are artifact-shaped rather than command-shaped — a payload is `{meta, sections}` for any artifact — so they belong next to `Manifest.JSONSchema` and `Manifest.RequiredCheck` in `internal/sections`. Only the flag binding stays in `cmd`. The alternative, a shared `cmd/structured.go`, would keep artifact logic in the command layer purely because that is where it grew.

**`--schema` and `--validate` do not follow `--from` onto `changelog`.**

Parity was the obvious reason to add them and it is the wrong one. The changelog artifact declares no required sections, so `--validate` could report nothing but `OK` — a check that cannot fail teaches a user to trust it for something it never verified. `--schema` would describe two string-typed properties, which is less informative than `--json` itself. There is also a naming cost: the `changelog-release-and-lint` change introduces `srekit changelog validate FILE` for linting a real changelog document, and shipping `srekit changelog --validate FILE` for an unrelated payload check would leave two operations distinguishable only by two dashes. Leaving the pair on `postmortem` alone keeps that name free for the thing users will actually reach for.

**Changelog `--from` accepts `meta` as well as `sections`.**

`repo`, `initialVersion` and `today` are all things a caller may legitimately want to pin — regenerating a document with the original date is the obvious case, and it is what makes the round-trip actually round-trip. Precedence follows the existing rule: flag, then file, then derived (git remote for `repo`, the clock for `today`). This is why the `artifact-generation` delta amends the slug-resolution requirement: the failure condition is now "none of the three", not "no flag and no remote".

**The `initial_release` section keeps its id and its title.** Only the link lines leave its body. Anyone whose tooling reads `.sections[].id` from `changelog --json` sees the same ids after this change; what changes is that one section's `body` no longer ends with two link definitions.

## Risks / Trade-offs

**A user's customized `changelog.yaml` in a templates directory keeps the old shape and the links stay inside `initial_release`.** → Correct and intended: their artifact, their layout. It renders exactly as before, and the only consequence is that a future mechanical rewrite of the link block will not find one. `templates diff` already shows them what drifted from the embedded version, and `templates upgrade`'s 3-way merge is the existing path for adopting the new layout. No migration is forced.

**The blank-line fix changes the bytes of documents users regenerate, so a re-run produces a diff in files people keep in git.** → One line per document, in whitespace only, and in the direction of correctness. It goes in `CHANGELOG.md` under `Changed` with that consequence spelled out rather than as a silent cosmetic fix.

**Widening `--from` to a second command tempts the next generator to get it "for free", and invites the argument that `--schema` and `--validate` should follow for symmetry.** → The requirement is written around the round-trip being meaningful, not around uniformity, and it says in as many words that the schema and validation flags are warranted only where an artifact declares required sections. `changelog` qualifies for `--from` because it already advertises a `sections` array people want to fill in; it fails the test for the other two. Any further command is a separate proposal that has to argue both points on its own.

**`footer_body` is one more thing a template author must learn, and the obvious wrong use is putting real content there instead of in a section.** → `TEMPLATES.md` documents it as "trailing document-level material such as link reference definitions", with the changelog as the worked example. Structural validation stays out of it; the format does not police taste anywhere else.
