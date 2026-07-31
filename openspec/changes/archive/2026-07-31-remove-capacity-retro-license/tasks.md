## 1. Remove the capacity and retro generators

- [x] 1.1 Delete `cmd/capacity.go` and `cmd/retro.go` and their registrations in `NewRootCmd()`
- [x] 1.2 Delete `internal/tmpl/templates/capacity.yaml` and `internal/tmpl/templates/retro.yaml`
- [x] 1.3 Delete the `capacity` and `retro` smoke tests from `cmd/cmd_test.go`

## 2. Remove the license generator

- [x] 2.1 Delete `cmd/license.go`, including the `lic` alias and the three inlined license bodies, and its registration in `NewRootCmd()`
- [x] 2.2 Delete the `license` smoke tests from `cmd/cmd_test.go`

## 3. Remove the --template plumbing

- [x] 3.1 Remove `BindTemplateFlag` and the `TemplatePath` field from the shared output flags, and drop the `TemplatePath` mention from the package and `Bind` doc comments
- [x] 3.2 Remove `TemplatePath` from the renderer's options and the legacy `text/template` branch it selects in `buildBody`
- [x] 3.3 Remove the template-file parsing helper (`tmpl.ParseFile`) and its test, `render.WriteRaw` (license was its only caller), the now-always-true `Options.RenderArtifact` field and its eight assignments, and any unused imports
- [x] 3.4 Remove the renderer tests covering the deleted branch from `internal/render/render_test.go`, and port the six output-routing tests that used a `.tmpl` fixture onto a v1 artifact fixture
- [x] 3.5 Confirm `templates validate`'s `.tmpl` handling and the sample registry are untouched — they serve `templates migrate`, not the render path

## 4. Retired command names

- [x] 4.1 Add hidden stub commands for `capacity`, `retro`, `license` and the `lic` alias that accept arbitrary arguments and unknown flags
- [x] 4.2 Make each stub exit non-zero with a message naming the release that removed it and pointing at the migration note, before any flag validation
- [x] 4.3 Add a comment tying the stubs to their 1.0 removal, matching the expiry the spec states

## 5. Sweep the catalog descriptions

- [x] 5.1 Update the root command's `Short` and `Long` in `cmd/root.go` to the eight-command catalog
- [x] 5.2 Update `README.md`
- [x] 5.3 Update the description in `.goreleaser.yaml`
- [x] 5.4 Update `site_description` in `mkdocs.yml`
- [x] 5.5 Update `internal/tmpl/TEMPLATES.md` — drop the `capacity.yaml` and `retro.yaml` entries and the whole `srekit license --template FILE` section
- [x] 5.6 Update the catalog sentence and the `--template`/`license` invariant in `CLAUDE.md`
- [x] 5.7 Update the `context` block in `openspec/config.yaml` — the artifact list and the "поэтому `--template` биндится только в license" invariant

## 6. Documentation

- [x] 6.1 Delete `docs/en/commands/capacity.md`, `retro.md`, `license.md` and their `docs/ru/` counterparts
- [x] 6.2 Remove the three nav entries from `mkdocs.yml` for both locales
- [x] 6.3 Scrub the three commands from the command index, getting-started, recipes, architecture, configuration, custom-templates and json-output pages in **both** locales
- [x] 6.4 Add a migration note in both locales: pin the previous release, keep an already-generated document as a static file, and use the code host's license picker for LICENSE
- [x] 6.5 Grep both documentation trees for `capacity`, `retro`, `license` and `--template` and confirm every remaining hit is intentional (the migration note and historical migration pages)

## 7. Tests

- [x] 7.1 Test that each retired name — `capacity`, `retro`, `license`, `lic` — exits non-zero with the removal message and creates no file
- [x] 7.2 Test that `srekit retro` with no `--team` returns the removal message rather than a missing-flag error
- [x] 7.3 Test that `srekit --help` lists exactly the eight catalog commands and none of the retired names
- [x] 7.4 Test that a surviving generator given `--template ./x.tmpl` fails with an unknown-flag error
- [x] 7.5 Test that a templates directory still holding `capacity.yaml` is classified `user-only` by `templates list` and that generation of the surviving artifacts is unaffected

## 8. Changelog and verification

- [x] 8.1 Add a `CHANGELOG.md` entry under `[Unreleased]` → `Removed`, marked breaking, naming each command, the last release that shipped it, and its remedy
- [x] 8.2 Run `task ci` — `golangci-lint run` clean (watch for newly unused helpers the linters will now flag) and `go test -race ./...` green
- [x] 8.3 Run `task docs:build` (`mkdocs build --strict`) to confirm no nav entry points at a deleted page in either locale
- [x] 8.4 Build and confirm the binary renders every surviving generator from the embedded set with no templates directory configured
