## Context

See proposal.md — Why. Three constraints shape the approach.

The English default is load-bearing. Anything that greps `### Added`, parses version headings, or drafts release notes from a changelog does so in English, and the project's own invariant says so. So the Russian variant cannot be a default, cannot be inferred from locale, and cannot be reached without the user saying it out loud.

Artifact resolution is a public contract that freezes at 1.0. A language segment in an artifact filename is a change to how a bare name becomes a file, which is exactly the kind of rule that has to be right before the freeze.

The changelog scanner from `changelog-release-and-lint` already takes its heading vocabulary as a parameter, by design, for this change. What is new here is the second vocabulary and the rule for choosing between them.

## Goals / Non-Goals

**Goals:**

- A Russian changelog that is a maintained, upgradeable artifact rather than a hand-edited fork.
- A language-variant resolution rule general enough to survive 1.0 without a second mechanism.
- Both maintenance subcommands working on Russian documents, including ones this tool did not generate.

**Non-Goals:**

- Localizing any other artifact. The rest are already bilingual and need no variant.
- A general i18n framework, message catalogs, or translated CLI output. Help text, errors and log lines stay English.
- Inferring the language from `LANG`, the locale, or the git history. Opt-in means opt-in.
- Translating anything beyond the changelog's own prose and change types.

## Decisions

**`[Unreleased]`, the version headings and the link labels stay English inside the Russian variant.**

This is the decision most likely to be questioned, so the reasoning matters. In Markdown, `## [Unreleased]` and `[Unreleased]: <url>` are two halves of one reference link — the heading text *is* the label. Translating the heading means translating the label, and the label is the part of the document that points outward, at a compare URL built from tags. A Russian project's tags are still `v1.2.0`; its compare links still live on an English-language forge. The Russian edition of the specification uses «Новое» in its prose because it is describing the concept, not prescribing a label to be substituted into a reference link.

There is a second reason. `[Unreleased]` is the one anchor that makes a changelog machine-locatable at all — it is how `changelog release` finds where to insert, and how every other tool in the ecosystem finds the same place. Keeping it stable across languages means a Russian changelog is still processable by tooling that knows nothing about Russian, which is the whole reason `changelog` was English in the first place. The translation stops exactly where the document stops being prose.

This is reversible if it proves wrong: the heading and its label move together, and `changelog validate` would catch any document where they disagree.

**The variant is a separate embedded file, not a language block inside one artifact.**

The alternative — `default_body_ru` keys, or a `lang:` mapping inside `changelog.yaml` — would put a translation mechanism into the v1 artifact format, which every other artifact would then have to ignore. A separate file keeps the format untouched, makes the variant diffable against its own snapshot, and lets a user override one language without touching the other. The cost is duplication between the two files, which is real but small and static: six headings and one paragraph of prose that change roughly never.

**The variant lookup runs across the whole source chain before the fallback does, not per source.**

Given a user templates directory holding a customized `changelog.yaml` and no Russian variant, the two orderings disagree. Chain-first yields the embedded `changelog.ru.yaml`; source-first yields the user's English customization. Chain-first is right: the user asked for Russian, and an English file — even their own — is not an answer to that question. Source-first would silently ignore `--lang` whenever a user had customized the base artifact, which is the worst kind of failure, since the flag would appear to work and produce English.

**The parser detects the vocabulary from the document; `--lang` never influences it.**

Generation and parsing are opposite directions and must not share a setting. A team that configures `changelog_lang: ru` still has an English `CHANGELOG.md` from before the switch, and `release` must not corrupt it. Detection from the first recognized change-type heading is unambiguous in every real document, and the case it cannot resolve — a document with no change types at all — is a validation failure rather than a guess.

**Mixing vocabularies is an error rather than a tolerated state.** A half-translated changelog is a file where `### Added` and `### Добавлено` mean the same thing and neither a reader nor a tool can group them. Accepting it silently would make the tool complicit in the drift; the check is cheap and the fix is obvious to whoever sees it.

**`--lang` is scoped to `changelog` rather than made persistent.**

A persistent `--lang` would imply every artifact has language variants, which none do and none need — they are already bilingual. Shaping it as a command-local flag now, with the resolution rule defined generally in `template-overrides`, means promoting it later is additive: the value set, the fallback semantics and the config precedence are already what a persistent flag would need.

**Configuration key is `changelog_lang`, not `lang`.** It configures one command's behaviour, and a bare `lang` in a config file reads like a global that would then silently do nothing for seven of the eight generators.

## Risks / Trade-offs

**A user picks `--lang ru` and later discovers their release tooling cannot parse the result.** → Documented as a warning in both locales, stated in the terms that matter: anything grepping `### Added` will stop working. The default stays English precisely so that this is a decision someone makes rather than one they inherit.

**Two embedded files drift apart when the format changes — a new section added to `changelog.yaml` and forgotten in the variant.** → A test asserting both variants declare the same section ids in the same order catches it at build time. The section ids are English in both files, which is what makes that test possible and is another argument for translating only prose.

**`changelog.ru.yaml` doubles the changelog's presence in `templates list`, `diff` and `upgrade` output.** → Correct and intended: they are two artifacts. The lifecycle spec is amended to say so rather than leaving users to work out why `templates upgrade` mentions a file they never asked for.

**Someone will ask for `--lang` on the other seven generators.** → The answer is in the design: they are bilingual by construction, so there is nothing to select between. If that ever changes, the resolution rule already generalizes and the flag can be promoted without breaking anything.

**A Russian document generated before this change — hand-translated by a user — may use a different word than the official edition (`Исправления` rather than `Исправлено`).** → `validate` reports it as an unrecognized change type, naming the heading and listing the accepted set. That is the correct outcome: the specification names six terms per language, and a tool that quietly accepted synonyms would make the vocabulary meaningless.
