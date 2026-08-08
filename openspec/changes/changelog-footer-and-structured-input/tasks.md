## 1. Artifact format: footer_body

- [ ] 1.1 Add `footer_body` to the parsed artifact structure and to the parse path, with no new structural validation (an absent key is empty, an empty key renders nothing)
- [ ] 1.2 Render the footer as the final block, evaluated through the same template context and FuncMap as `header_body`
- [ ] 1.3 Unit tests: artifact with a footer, artifact without a footer (byte-identical to pre-change output), footer containing a template directive, footer that is whitespace-only

## 2. Block separation

- [ ] 2.1 Introduce a single "start a new block" helper in the composer that guarantees the buffer ends in exactly one blank line, and no-ops on an empty buffer
- [ ] 2.2 Route frontmatter, title, meta bullets, header body, each section and the footer through it; remove the per-block leading and trailing separators that produced the double blank line
- [ ] 2.3 Unit tests covering the combinations: title+header body, frontmatter+header body with no title, meta bullets+header body, sections+footer, and every element absent but one
- [ ] 2.4 Update the golden/expected output in existing render tests for artifacts that declare a `header_body`

## 3. Changelog artifact

- [ ] 3.1 Move the `[Unreleased]` and `[<version>]` link definitions out of `initial_release.default_body` into `footer_body` in `internal/tmpl/templates/changelog.yaml`
- [ ] 3.2 Verify the rendered document is unchanged apart from the removed extra blank line and the link block's new position relative to the section body
- [ ] 3.3 Document `footer_body` in `internal/tmpl/TEMPLATES.md` next to `header_body`, using the changelog link block as the worked example

## 4. Shared structured-input plumbing

- [ ] 4.1 Move the payload-reading helper (file or `-`, empty file as no input, malformed JSON naming the file) out of `cmd/postmortem.go` into `internal/sections`
- [ ] 4.2 Move the meta-precedence helper alongside it
- [ ] 4.3 Rewire `cmd/postmortem.go` onto the moved helpers and confirm its existing tests pass unchanged — this step must be behaviour-neutral, and `--schema` / `--validate` stay postmortem-only

## 5. Changelog structured path

- [ ] 5.1 Bind `--from` on `changelog`; do not bind `--schema` or `--validate`
- [ ] 5.2 Read the payload before repository-slug resolution; apply precedence flag → `meta` from file → git remote for `repo`, and flag → file → clock for `today` and `initialVersion`
- [ ] 5.3 Pass the payload's `sections` map into the merge instead of `nil`, so provided bodies replace defaults verbatim and unknown ids fail
- [ ] 5.4 Update the slug-resolution error so it still names `--repo` as the remedy when neither flag, file nor remote supplies one
- [ ] 5.5 Add an `Example` block on the command mirroring the postmortem round-trip examples

## 6. Tests

- [ ] 6.1 `cmd/cmd_test.go`: changelog round-trip (`--json` → edit `unreleased` body → `--from` → assert the body lands in `## [Unreleased]`)
- [ ] 6.2 `cmd/cmd_test.go`: changelog `--from` with an unknown section id fails naming the id and listing the known ids
- [ ] 6.3 `cmd/cmd_test.go`: changelog `--from -` reads stdin; `meta.repo` from the file is used when no `--repo` and no git remote is available
- [ ] 6.4 `cmd/cmd_test.go`: `srekit changelog --help` lists `--from` and lists neither `--schema` nor `--validate`
- [ ] 6.5 Assert the rendered changelog ends with the link block and that it survives a `--from` payload that replaces the `initial_release` body

## 7. Documentation and release notes

- [ ] 7.1 `docs/en/commands/changelog.md` and `docs/ru/commands/changelog.md`: structured input section (`--from`), round-trip example, note that the link block is a document footer and that `--schema` / `--validate` are postmortem-only
- [ ] 7.2 Custom-templates pages in both locales: document `footer_body` in the v1 key list
- [ ] 7.3 JSON-output pages in both locales: state that `footer_body` is not part of the `sections` array and cannot be targeted through `--from`
- [ ] 7.4 Refresh any embedded sample output in either locale that carries the extra blank line
- [ ] 7.5 `CHANGELOG.md` under `[Unreleased]`: `Added` for `footer_body` and `changelog --from`; `Changed` for the block-separation fix (noting that regenerated documents will show a one-line whitespace diff) and the link-block move

## 8. Verification

- [ ] 8.1 `go build ./... && go test -race ./... && golangci-lint run`
- [ ] 8.2 `task docs:build` passes strict mode with both locales updated
- [ ] 8.3 Render every embedded artifact and confirm no document contains two consecutive blank lines and every one ends with exactly one trailing newline
