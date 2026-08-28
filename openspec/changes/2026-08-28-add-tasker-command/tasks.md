## 1. Typed front matter values

- [x] 1.1 Retype an explicitly tagged front matter scalar after rendering: re-read the rendered text as YAML and replace the node with what it denotes
- [x] 1.2 Restrict retyping to the YAML-defined types (`!!int`, `!!float`, `!!bool`, `!!null`, `!!timestamp`, `!!seq`, `!!map`); leave `!!str` and application tags on the untagged path
- [x] 1.3 Fail rendering when the text does not read as the declared type, and name the front matter key in every front matter error, not only this one
- [x] 1.4 Add `join` to the shared FuncMap, separator first
- [x] 1.5 Unit tests: typed values reach the document untagged, a mismatch is a named error, a custom tag survives untouched, `join`

## 2. The artifact

- [x] 2.1 Add `internal/tmpl/templates/tasker.yaml`: front matter `id`, `creation_date`, `type: simple_note`, `tags: [tasker]`, `topic`, `level` (`!!seq`), `format`, `duration` (`!!int`)
- [x] 2.2 H1 `Tasker - {{ .Meta.Title }}`; sections `description` and `expectations` with bilingual titles and empty bodies

## 3. The command

- [x] 3.1 Add `cmd/tasker.go` with `--title` mandatory and `--topic` / `--level` / `--format` / `--duration` defaulted, plus the shared output bundle
- [x] 3.2 Trim and drop blank levels; reject an empty level set and a non-positive duration before anything is written
- [x] 3.3 Default filename `tasker-<slug(title)>.md`
- [x] 3.4 Register in `NewRootCmd()` and in the catalog assertion in `cmd/retired_test.go`
- [x] 3.5 Smoke tests: defaults, flags reaching front matter, empty sections, `--json` shape, rejected input

## 4. Documentation

- [x] 4.1 New command page in `docs/en/commands/tasker.md` and `docs/ru/commands/tasker.md`, cross-linked with `task` in both directions
- [x] 4.2 Both command overviews, both landing pages, `mkdocs.yml` nav
- [x] 4.3 Typed front matter section in both custom-templates guides, and in `internal/tmpl/TEMPLATES.md` together with `join` and the `tasker.yaml` placeholder table
- [x] 4.4 `templates init` file count (9 → 10) wherever it is quoted in the docs
- [x] 4.5 README section, and the artifact list in its opening line
- [x] 4.6 `CHANGELOG.md` entry under `[Unreleased]`

## 5. Verification

- [x] 5.1 `go build ./... && go test ./...`
- [ ] 5.2 `golangci-lint run` — blocked locally: the pinned v2.12.2 is built with go1.26.5 and cannot typecheck the go1.27.0 stdlib on this machine, failing on files this change does not touch. `go vet ./...` and `gofmt -l` are clean; CI runs the pinned toolchain
- [x] 5.3 `mkdocs build --strict`
