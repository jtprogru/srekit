## Context

Two decisions carry this change: where the card lives in the command tree, and how a front matter value stops being a string.

## Decision: a new command, not a second mode of `task`

`srekit task` generates the SRE investigation log. It also used to generate exactly this card — the `gch sretask` inheritance, whose default filename was `Tasker - <title>.md` until v0.20.0 rewrote the command. Restoring the card under `task` would break the document that has lived there for a dozen releases; adding a `--kind` switch would make one command render two unrelated artifacts, which is the shape `--template FILE` was removed for.

So: `tasker` is its own command with its own artifact, and `task` is untouched. The cost is two similar names next to each other, paid down with a cross-reference in both directions in the docs and in the command's long help.

## Decision: typed front matter through an explicit YAML tag

Everything a template renders is a Go string. The frontmatter node is a `yaml.Node`, and the encoder writes what the node says it is, so a templated value can only ever come out quoted. For the SRE artifacts that never mattered: their front matter is dates, ids and names, and all of them are strings. A task card's is `level: [middle, senior]` and `duration: 30`, read by the collection that holds the card, where a quoted scalar is a different value.

Three options were on the table:

1. **Re-resolve every rendered scalar.** Parse each rendered value as YAML and take whatever type it turns out to be. Rejected: it changes existing documents. Every RFC3339 `creation_date` in the shipped artifacts would silently become a `!!timestamp` and lose its quotes, and any title that happened to render as `30` would become a number.
2. **Pre-format the value in Go and hand the template a ready YAML fragment.** Rejected: it puts YAML syntax knowledge in `cmd/`, and the fragment would still be emitted as a quoted string unless the renderer did something special anyway.
3. **Opt in per value with an explicit tag.** Chosen. `duration: !!int "{{ .Meta.Duration }}"` says what the author means, at the value, in the file the author edits.

The tag has to be *explicit*: every decoded scalar carries a resolved tag, so `"{{ .Meta.N }}"` is already `!!str` without anyone saying so. `yaml.TaggedStyle` is what separates the author's tag from the parser's inference.

The retyped set is the YAML-defined types minus `!!str`. `!!str` is excluded because an explicit string tag asks for exactly what the untagged path already does. Application tags (`!Ref`, `!Sub`, …) are excluded because their payload belongs to whatever tool reads the document, and reinterpreting it would be srekit editing somebody else's data.

A declared type the text does not have is an error, not a silent fallback to string: the artifact format spec already defers "frontmatter scalar-type problems" to render time, and this is that error. It names the front matter key, because the mapping walk now wraps each value's error with the key it came from — a diagnostic every frontmatter error gains, not just this one.

## State

No new state. The tagged-scalar path is a pure function of the node and the render context; the node is walked on a deep copy, so rendering the same artifact twice still yields identical bytes. `join` is a pure function in the existing FuncMap registry, which is the one intentional package-level value in `internal/tmpl` and is not mutated at runtime.

## Dependencies and binary size

None added. Tag inspection and re-parsing are `yaml.Node` fields and `yaml.Unmarshal`, both already linked. The binary grows by one embedded artifact and one cobra command — under a kilobyte of embedded YAML plus the command's own code. Nothing here reaches `net/http` or `crypto`.

## Risks

- **Two similar command names.** `task` and `tasker` are one character apart and produce unrelated documents. Mitigated by cross-links in both doc pages, the README and the command's long help; not mitigated by the shell, which will happily complete either.
- **Cyrillic titles slug to `untitled`.** The slug rule keeps `[a-z0-9]`, so a Russian card name produces `tasker-untitled.md`, and the second such card refuses to overwrite the first. This is the existing, specified slug behaviour, and changing it would touch every generator's filenames; the command page documents `--out` as the way through. Worth revisiting as its own change if the collection turns out to be mostly Russian-titled.
