## 1. Prerequisites

- [ ] 1.1 Confirm `changelog-footer-and-structured-input` is implemented — the Russian variant carries the same document-level footer
- [ ] 1.2 Confirm `changelog-release-and-lint` is implemented and its scanner takes the change-type vocabulary as a parameter rather than a constant

## 2. Language-aware artifact resolution

- [ ] 2.1 Extend artifact-name resolution with an optional language segment: try `<name>.<lang>.yaml` across every source in the chain, then `<name>.yaml` across every source
- [ ] 2.2 Keep normalization idempotent for a language-suffixed name so `changelog.ru.yaml` does not become `changelog.ru.yaml.yaml`
- [ ] 2.3 A requested language with no variant anywhere SHALL fall back silently, not error
- [ ] 2.4 Unit tests: variant preferred, fallback when absent, user directory shadows the embedded variant, user directory holding only the base artifact does not win over the embedded variant, idempotent suffixed name, traversal-safe names unchanged

## 3. Russian artifact

- [ ] 3.1 Add `internal/tmpl/templates/changelog.ru.yaml` with Russian change types, Russian header prose, and the header link pointing at the Russian edition of the specification
- [ ] 3.2 Keep section ids, version headings, `[Unreleased]` and the link reference labels in their English form
- [ ] 3.3 Add a test asserting both changelog variants declare the same section ids in the same order, so the two files cannot drift apart structurally

## 4. Command wiring

- [ ] 4.1 Bind `--lang` on `changelog` with values `en` and `ru`, defaulting to `en`
- [ ] 4.2 Resolve the value as flag → `changelog_lang` config → `en`, rejecting an unrecognized value from either source before any file is written
- [ ] 4.3 Pass the selection through to artifact resolution, and apply it to the whole command group so `release` and `validate` inherit it
- [ ] 4.4 Add `changelog_lang` to the recognized configuration keys and to the `SREKIT_`-prefixed environment lookup

## 5. Vocabulary handling in the maintenance subcommands

- [ ] 5.1 Supply the Russian change-type set to the scanner alongside the English one
- [ ] 5.2 Detect the vocabulary in force from the document's first recognized change-type heading, never from `--lang` or configuration
- [ ] 5.3 Refuse a document mixing both vocabularies in `release`, leaving the file untouched; report it as a named failing check in `validate`
- [ ] 5.4 Report the detected vocabulary in `validate` output
- [ ] 5.5 Report a document with no recognized change types as a failing check rather than treating it as either language

## 6. Tests

- [ ] 6.1 `cmd/cmd_test.go`: default invocation is byte-identical to the pre-change output; `--lang ru` renders the Russian change types and the Russian specification link
- [ ] 6.2 `cmd/cmd_test.go`: `changelog_lang: ru` in config selects the variant; `--lang en` overrides it; an unrecognized value from either source exits non-zero naming the accepted values
- [ ] 6.3 `cmd/cmd_test.go`: `--lang ru` output keeps `## [Unreleased]` and English link labels
- [ ] 6.4 `cmd/cmd_test.go`: release on a Russian document preserves its vocabulary; release with `--lang ru` on an English document preserves English
- [ ] 6.5 `cmd/cmd_test.go`: mixed-vocabulary document fails both `validate` and `release`, and `release` leaves it byte-identical
- [ ] 6.6 `cmd/cmd_test.go`: `templates list` shows `changelog.yaml` and `changelog.ru.yaml` as separate entries; an edited variant upgrades against its own snapshot without disturbing the base artifact

## 7. Documentation and release notes

- [ ] 7.1 `docs/en/commands/changelog.md` and `docs/ru/commands/changelog.md`: the `--lang` flag, a rendered Russian example, and an explicit warning that the Russian variant breaks anything grepping `### Added`
- [ ] 7.2 Both locales: document why `[Unreleased]`, version headings and link labels stay English
- [ ] 7.3 Configuration pages in both locales: `changelog_lang` and its environment form
- [ ] 7.4 Custom-templates pages in both locales: how a language variant resolves and how to override one
- [ ] 7.5 `CHANGELOG.md` under `[Unreleased]` → `Added`
- [ ] 7.6 Update the bilingual-templates note in `CLAUDE.md` and `openspec/config.yaml`, which currently states flatly that `changelog` is entirely English

## 8. Verification

- [ ] 8.1 `go build ./... && go test -race ./... && golangci-lint run`
- [ ] 8.2 `task docs:build` passes strict mode with both locales updated
- [ ] 8.3 Generate a Russian changelog, cut a release on it, and validate the result end to end
