## Context

See proposal.md — Why. Four constraints shape the approach.

This is the first `srekit` command that reads a file the user wrote. Every other write path renders a fresh document from an artifact and either creates a file or refuses to. Here the input is someone's hand-maintained changelog, often years old, edited by many people, and the tool's credibility rests entirely on not mangling it.

`changelog-footer-and-structured-input` must land first. It moves the link reference definitions into a document-level footer, which is what makes them addressable; without it the block is prose at the tail of a section body and any rewrite has to guess where the prose ends.

Dependencies are capped at cobra and the YAML library, so semantic-version comparison, date handling and Markdown scanning are all written here or come from the standard library.

The localization change (`changelog-ru-variant`) will introduce a Russian changelog artifact whose change-type headings are `Добавлено`, `Изменено` and so on. Both subcommands must read documents written in either language. That is a known future requirement, not a speculative one, so the heading vocabulary is a parameter from the first commit rather than a refactor later.

## Goals / Non-Goals

**Goals:**

- Cutting a release becomes one command instead of four manual edits, with the link block updated correctly every time.
- A changelog can be checked against Keep a Changelog before a release rather than after a downstream tool breaks.
- Yanked releases become expressible.
- The rewrite is conservative: outside the regions it edits, the file comes out byte-identical.

**Non-Goals:**

- Generating entries from git history. That is a different tool with a different failure mode (`git log` diffs are the anti-pattern the format explicitly names), and nothing here reads commits.
- Tagging, pushing, or touching git at all beyond the existing slug detection.
- Understanding Markdown in general. Only headings, the change-type subsections beneath them, and link reference definitions are parsed; everything else is opaque text that gets carried across.
- Repairing a document that fails validation. `validate` reports; it does not rewrite.
- Reading the Russian artifact's vocabulary. This change parameterizes the vocabulary and ships English; the second set arrives with the localization change.

## Decisions

**A line-oriented region scanner, not a Markdown parser and not a model-then-reserialize round trip.**

The obvious design — parse into a document model, mutate, render back — is the one that damages files. Reserializing means every blank line, every bullet marker style, every wrapped line in a five-year-old changelog gets normalized to whatever this tool considers canonical, and the release commit becomes an unreviewable diff. Instead the scanner records byte offsets: where `## [Unreleased]` starts and ends, where each version heading sits, where the link block begins. The rewrite splices new text at those offsets and copies everything else through untouched. This is why the spec can promise byte-identical preservation of older versions and hand-written preamble, and it is the single most important property of the design.

Rejected alternative: an off-the-shelf Markdown AST library. It would bring a dependency into a build graph deliberately kept at two, and AST round-tripping is precisely the reserialization problem above.

**Link conventions are inferred from the document, not from git.**

The existing `[Unreleased]` definition already encodes everything the new definitions need: the host, the repository path, the comparison URL shape, and whether tags carry a `v` prefix. Reading it back is both more accurate and more portable than re-deriving from the git remote — it handles self-hosted GitLab, projects whose changelog points at a mirror, and tag schemes this tool never anticipated. Falling back to slug resolution only when there is no link block at all keeps the first-release case working.

Rejected alternative: a `--tag-prefix` flag. It would be a flag that exists to restate what the file already says, and the failure it guards against — a document whose own link block disagrees with its tags — is one `validate` should surface rather than one `release` should paper over.

**`[Unreleased]` is left empty, not refilled with the six-type skeleton.**

Keep a Changelog's own example keeps an empty `[Unreleased]` at the top. Refilling the skeleton would mean every release commit adds six headings and six placeholder bullets that the next release then strips, so the diff of every release contains noise in both directions. The scaffold emits the skeleton because a brand-new file benefits from showing the vocabulary; a file that has already been through a release does not need the lesson.

**`release` refuses a version it can already see, rather than being idempotent.**

An idempotent re-run is meaningless here: the second run would move whatever has accumulated in `[Unreleased]` since the first, and file it under a version that shipped without it. Refusing is the only behaviour that cannot silently produce a false record. Re-running after a mistake means editing the file, which is the honest operation.

**The document linter is `changelog validate FILE`, a subcommand, and the payload check keeps its flag form elsewhere.**

`templates validate` already established the subcommand-shaped "check files on disk" verb, and this is the same shape. The companion change deliberately does not add `changelog --validate`, so the two-dashes-apart collision never exists. Splitting the checks across separate subcommands (`lint`, `check-links`) was considered and rejected: they share one parse and users want one exit code.

**Validation reports every check, not the first failure.** A person fixing a drifted changelog wants the full list in one pass. This matches `postmortem --validate`'s existing `OK`/`FAIL` per-item reporting and its non-zero exit carrying a failure count.

**Version comparison is a small numeric-segment compare, not a semver dependency.**

Ordering checks and the "which version was previously newest" question need `1.10.0 > 1.9.0` to come out right, which string comparison gets wrong. A dozen lines splitting on `.` and comparing numeric segments covers every version this tool will meet; anything it cannot parse is reported by the heading-shape check rather than silently mis-ordered. Adding a semver library for this would violate the dependency rule for a function that fits on a screen.

**State lives in a value returned by the parser.** The scanner takes bytes and a heading vocabulary and returns a document value; the rewriter takes that value and returns new bytes. No package-level mutable state, so the parser tests parallelize like everything else.

## Risks / Trade-offs

**A changelog whose shape the scanner does not recognize gets rewritten wrongly and the user loses history.** → Three defences. The scanner refuses rather than guesses: no `## [Unreleased]` heading is an error, not an assumption about where to insert. Every precondition failure leaves the file untouched, which the spec states per case. And `--dry-run` prints the exact result, which the documented release-day sequence puts before the real run.

**Users will expect `release` to also tag and push.** → It will not, and the docs say so in the release-day sequence: `srekit changelog release`, review the diff, commit, tag. Making the tool touch git turns a text edit into an irreversible action, and the repository's own release flow already has a tagging step that must stay under human control.

**Parsing a hand-maintained file has a long tail: prose between a version heading and its first `###`, entries with no subsection at all, nested lists, HTML comments, two link blocks.** → Everything the scanner does not classify is opaque text carried through verbatim, so the tail degrades into "preserved but not understood" rather than into damage. `validate` is where unclassifiable content becomes visible, and its checks are written against the document as found rather than against an idealized shape.

**Scoping `--out` and `--force` away from this command creates an inconsistency a future reviewer may try to "fix".** → That is why it is a spec requirement in `output-routing` with the reasoning attached, not a local decision in the command file.

**Two commands now know what a change-type heading is: the artifact templates and the scanner.** → The scanner's vocabulary is the specification's six types, which is a fixed list, not a copy of the template. A user who renames headings in a customized artifact gets a `validate` failure, which is the correct answer — the format names those six.
