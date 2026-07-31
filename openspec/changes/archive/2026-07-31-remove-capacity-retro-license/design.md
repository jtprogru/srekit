## Context

See `proposal.md` — Why, and `specs/` for the requirements. This document covers the ordering and the judgment calls.

The relevant current state:

- `license` is structurally unlike the other generators. It renders inlined string constants through the shared function map rather than loading a v1 artifact, it defaults to standard output rather than a file, and it is the only command that binds `--template FILE`. That flag routes into a `text/template` branch in the renderer which no other command reaches. Removing the command therefore removes a whole render path, not just a file.
- `capacity` and `retro` are ordinary v1 generators. Removing them is deleting a command file, an embedded artifact, and their documentation.
- The embedded template set shrinking is a case the template lifecycle already handles: user directories are never touched, a leftover `capacity.yaml` reclassifies from `customized` to `user-only`, and orphan snapshots are collected on the next `upgrade`.
- The project is pre-1.0 and the catalog stabilizes at 1.0.

## Goals / Non-Goals

**Goals:**

- Leave a user who runs a retired command with an answer, not a typo-style error.
- Remove the `--template` plumbing completely in the same change, so the codebase never sits in the state where a second render path exists with no caller.
- Keep every surviving generator byte-identical in behaviour, so the change is provably a subtraction.

**Non-Goals:**

- No deprecation release. See the decision below.
- No replacement for the removed artifacts, in-tree or otherwise. `srekit` is not the right home for them; pointing at a nonexistent successor would be worse than pointing at nothing.
- No cleanup of `templates validate`'s `.tmpl` handling. Bespoke `.tmpl` files remain valid input to `templates migrate`, and the sample registry it validates against is already empty, so nothing observable changes there. Removing that path is 2.0 work, tracked with the rest of the legacy-format retirement.

## Decisions

### Remove outright, with retired-name stubs, rather than deprecate for a release

A deprecation window means one release where the commands still work but print a warning, then a second release that removes them. The argument for it is that users get time. The argument against, which wins here: the project is pre-1.0, users are expected to read a changelog on minor bumps, and a warning printed to stderr on a command whose output is redirected to a file is a warning nobody reads. A deprecation release would mostly delay the break to a point where more people have scripted around it.

The stub is where the courtesy actually lands. A user whose CI runs `srekit capacity --service payments` gets a message naming the release and the migration note, on the run where it breaks, on stderr, with a non-zero exit — which is strictly more information than they would have gotten from a deprecation warning they never saw. The cost is a few lines per name and one test.

Stubs are hidden rather than listed: a retired command in `--help` invites people to try it. They are dropped at 1.0, and that expiry is written into the spec so it does not become permanent by inertia.

### Retired commands fail before parsing flags

The stub takes arbitrary arguments and ignores them. `srekit retro` with no `--team` must say "removed", not "`--team` is required" — a user debugging a validation error they cannot satisfy is worse off than one told the command is gone. This means the stub accepts any args and any unknown flags rather than declaring the original flag set.

### Remove the `--template` plumbing in the same change

The alternative is to leave `--template` bound to nothing "in case someone wants it back". That leaves an option field on the shared output flags, a branch in the renderer, and a template-file parser with no caller — dead code that the next reader has to prove is dead. The specs already say a flag no command implements must not exist; the implementation should match. Reverting is a `git revert` away if the judgment turns out wrong.

Note this makes `--template` an unknown flag rather than a rejected one, which means cobra prints the usage block. That is the correct behaviour for a flag that no longer exists, and the spec says so.

### Delete the embedded artifacts rather than keeping them unreferenced

`capacity.yaml` and `retro.yaml` could stay in the embed directory so that `templates init` still scaffolds them for people who want to keep the text. Rejected: an embedded artifact with no command to render it is a trap. It would show up in `templates list` as `embedded-only`, be copied by `templates init`, be merged by `templates upgrade`, and never produce a document. The text is preserved in git history and reproduced in the migration note for anyone who wants to keep it as a plain Markdown skeleton.

### Order the work so the tree is never broken

Documentation and the catalog descriptions are updated in the same change as the code, not after. The spec-level catalog, the root command's help text, the README, the goreleaser description, the MkDocs site description, `TEMPLATES.md`, `CLAUDE.md`, and the OpenSpec project context all enumerate the artifact list independently — each is a place the old list survives if it is not swept deliberately. That is why the task list names them individually rather than saying "update the docs".

## Risks / Trade-offs

- **A user's automation breaks on upgrade** → This is the accepted cost, stated as **BREAKING** in the proposal and the changelog. The stub converts a confusing failure into an explained one, and the changelog entry names the last release that shipped each command so pinning is a one-line fix.
- **Someone genuinely used `--template` with `license` for a custom license body** → That user loses the feature outright, not just the command. Called out explicitly in the migration note rather than buried in the license removal: their remedy is to commit the rendered file once, which is what the flag was producing anyway.
- **A stale mention of a removed command survives in one locale of the docs** → The two locales are swept as one task with a grep over both trees, and `mkdocs build --strict` fails on a nav entry pointing at a deleted page. The prose mentions are the part `--strict` will not catch, hence the explicit per-file list.
- **Deleting the legacy render branch removes coverage that was incidentally exercising shared code** → The renderer's remaining branches (JSON short-circuit, artifact path) are already covered by every generator's tests; removing the third branch removes its own tests only. Verified by the race suite staying green rather than by assumption.
- **The catalog shrinks to eight commands and starts to look thin** → Not a real risk. Eight artifacts that share a definition beat eleven that do not, and `doctor` and future additions land against a coherent catalog.

## Migration Plan

Single release. There is no phased rollout and no data to migrate — the change removes commands and touches nothing a user has on disk.

The changelog entry under `Removed` SHALL name each removed command, the last release that shipped it, and the remedy. The migration note added to both documentation locales covers: pinning the previous release, keeping a previously generated document as a static file, and — for `license` — using the code host's license picker.

Rollback is a revert of the commit; nothing about the change is one-way, since no user state is modified.
