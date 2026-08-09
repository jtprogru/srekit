## Why

Every shipped artifact in `srekit` is bilingual — `Русский (English)` headings over Russian prose — except `changelog`, which is deliberately English so that tooling built around Keep a Changelog keeps parsing it. That decision is right for the default and wrong as an absolute. Keep a Changelog publishes an official Russian edition whose change types are `Добавлено`, `Изменено`, `Устарело`, `Удалено`, `Исправлено`, `Безопасность`, and a Russian-speaking team writing a changelog nobody's CI parses has no reason to be forced into English by a tool whose every other artifact speaks their language.

Today the only way to get one is to run `srekit templates init` and hand-edit the copy, which then permanently diverges from the embedded artifact and starts producing merge conflicts on every `templates upgrade`. A shipped variant is maintained; a hand-edited copy is a fork.

The reason to add it now rather than after 1.0 is that the resolution rule — how a bare artifact name picks up a language variant — is part of the template-resolution contract, and that contract freezes at 1.0.

## What Changes

- Add `--lang en|ru` to `changelog`, defaulting to `en`. Nothing changes for anyone who does not pass it.
- Ship `changelog.ru.yaml` as an embedded artifact: the Russian change types, Russian header prose, and the header link pointing at the Russian edition of the specification.
- Artifact resolution learns language variants: a requested language SHALL try `<name>.<lang>.yaml` first and fall back to `<name>.yaml`. The lookup walks the same source chain as today, so a user's own `changelog.ru.yaml` in a templates directory shadows the embedded one exactly as `changelog.yaml` already does.
- Add the `changelog_lang` configuration key and its `SREKIT_CHANGELOG_LANG` environment form, so a Russian-speaking team sets it once instead of typing the flag.
- The version headings, the `[Unreleased]` heading and the link reference labels stay English in the Russian variant. Those are link anchors and version identifiers, not prose — translating `[Unreleased]` would rename the reference label that points at it, which is the one part of the document that must match something outside it. Only the change types and the surrounding prose are translated.
- `changelog release` and `changelog validate` learn to read either vocabulary, detected from the document itself, and require a document to be internally consistent rather than mixing the two.
- No new dependencies. This is one embedded YAML file, one flag, and a suffix in the name-resolution path.

Nothing here is **BREAKING**. The default output is byte-identical, no existing flag changes meaning, and the resolution change is a new lookup attempted before the existing one — an artifact with no language variant resolves exactly as before.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `artifact-generation`: the bilingual-templates requirement currently states flatly that `changelog` is entirely English. It becomes: English by default and on every unqualified invocation, with an explicit opt-in Russian variant, and a statement of what stays English inside that variant and why.
- `template-overrides`: the artifact-name resolution requirement gains the language-variant lookup and its fallback, including the interaction with a user templates directory.
- `user-configuration`: the recognized-keys requirement gains `changelog_lang`.
- `template-lifecycle`: the embedded set now contains an artifact whose name carries a language segment, which `templates list`, `init`, `upgrade` and `diff` enumerate and snapshot like any other file; the divergence classification requirement is amended to say so explicitly rather than leaving it to be inferred.
- `changelog-maintenance`: the release and validation requirements gain the rule that a document's change-type vocabulary is detected from its content, that either language is accepted, and that mixing the two in one document is a validation failure.

## Impact

- New `internal/tmpl/templates/changelog.ru.yaml`.
- Artifact name resolution gains an optional language segment; the legacy-spelling normalization must keep working, so `changelog.ru.yaml` passed as a name must not become `changelog.ru.yaml.yaml`.
- `cmd/changelog.go` gains `--lang` with precedence flag → config → `en`, and passes the language through to the loader and to both subcommands.
- The changelog scanner introduced by `changelog-release-and-lint` takes its heading vocabulary as a parameter; this change supplies the second set and the detection rule.
- Depends on `changelog-release-and-lint` for the scanner, and transitively on `changelog-footer-and-structured-input`. It can ship without them if their scope is dropped, at the cost of leaving the vocabulary-detection requirements unimplementable.
- Docs, both locales: a section on the Russian variant with an explicit warning that it breaks anything grepping for `### Added`, plus the new config key on the configuration pages.
- `CHANGELOG.md` under `[Unreleased]` → `Added`.
