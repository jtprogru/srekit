## Why

The author keeps a collection of engineering tasks — interview and knowledge-check cards — where every note has the same shape: front matter carrying `topic`, `level`, `format` and `duration`, an H1 reading `Tasker - <name>`, and two sections, the task and what a good answer sounds like. That shape is currently reproduced by hand or copy-pasted from a neighbouring card, which is exactly the work every other `srekit` generator exists to remove.

The name is not free: `srekit task` is the SRE investigation log, and was the command that emitted `Tasker - <title>.md` before v0.20.0 repurposed it. The card therefore arrives as its own command rather than as a second personality for `task`, so that neither document's users are surprised.

## What Changes

- New generator command `srekit tasker`, producing one card per invocation from `internal/tmpl/templates/tasker.yaml`.
- `--title` is mandatory; `--topic` (default `go`), `--level` (repeatable or comma-separated, default `middle,senior`), `--format` (default `code`) and `--duration` in minutes (default `30`) are optional. Default filename `tasker-<slug(title)>.md`.
- Both sections ship with an empty body. The card is a slot for a task somebody is about to write; a placeholder would be text they delete on every card.
- **The catalog rule is widened.** `artifact-generation` today says a document belonging to a different discipline SHALL NOT enter the catalog, which is what got `capacity`, `retro` and `license` removed in v0.30.0. A task card is such a document: it is not owned by an on-call engineer and has nothing to do with production. The requirement is amended to admit this one artifact by name, with the exclusions it already lists left standing — the rule keeps its teeth, it just stops being a rule the shipped catalog violates silently.
- The v1 artifact format gains **typed front matter values**: an explicit YAML tag on a templated scalar (`!!int`, `!!float`, `!!bool`, `!!null`, `!!timestamp`, `!!seq`, `!!map`) makes the renderer read the rendered text back as a value of that type. Without it every templated value is a string, and `duration: "30"` / `level: "middle, senior"` are not the values the collection reading that front matter expects. `!!str` and application tags are left alone.
- The shared FuncMap gains `join`, needed to assemble a `[]string` meta field into a flow sequence.
- Not a breaking change. No existing command, flag, filename, JSON payload, template or config location changes behaviour. Untagged front matter renders byte-identically to before — the retyping path is entered only by a tag no shipped artifact carried until now.
- No new dependency: tags and re-parsing are `go.yaml.in/yaml/v3`, already in the graph. Binary size grows by one embedded artifact (~500 bytes) and one command.

## Capabilities

### Modified Capabilities

- `artifact-generation`: the catalog admits `tasker`; its mandatory input, optional defaults and default filename join the existing tables.
- `artifact-template-format`: front matter values may declare a YAML type; `join` joins the documented helper set.

### New Capabilities

None. `tasker` is an ordinary generator: it renders through the same artifact runtime, obeys the same output-flag bundle, and resolves its template through the same source chain.

## Impact

- New `cmd/tasker.go` and `internal/tmpl/templates/tasker.yaml`; registered in `NewRootCmd()`.
- `internal/sections/render_artifact.go` learns the tagged-scalar path; `internal/tmpl` gains `join`.
- `cmd/retired_test.go`'s catalog assertion gains `tasker` — the test exists precisely so a new visible command cannot appear unnoticed.
- Docs: new command page in `docs/en/commands/` and `docs/ru/commands/`, both command overviews, both landing pages, the two custom-templates guides (typed front matter), `internal/tmpl/TEMPLATES.md`, `mkdocs.yml` nav, and the `templates init` file count in four places.
- `CHANGELOG.md` entry under `[Unreleased]`.
- Smoke tests in `cmd/cmd_test.go`; unit tests for the tagged-scalar path in `internal/sections` and for `join` in `internal/tmpl`.
