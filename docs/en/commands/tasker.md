# srekit tasker

Generate a **task card** for a collection of engineering tasks: front matter carrying topic, level, format and expected duration, an H1 naming the task, and the two sections a card carries — the task itself and what a good answer sounds like.

The card ships empty on purpose. `tasker` lays out the shape; the person adding the task writes the content.

## Synopsis

```bash
srekit tasker --title NAME [flags]
```

## Flags

| Flag | Required | Description |
|---|---|---|
| `--title`, `-T` | yes | Task name — becomes the H1 and the filename slug |
| `--topic` | no | Subject area. Default: `go` |
| `--level` | no | Target levels, repeatable or comma-separated. Default: `middle,senior` |
| `--format` | no | How the task is answered (`code`, `theory`, `design`, …). Default: `code` |
| `--duration` | no | Expected time to solve, in minutes. Must be positive. Default: `30` |

Plus the [shared output flags](index.md#shared-output-flags). Default filename: `tasker-<slug-of-title>.md`.

Blank levels are dropped, so `--level "middle, "` is one level. A `--level` that leaves nothing behind, and a `--duration` of zero or less, fail before anything is written.

!!! note "Non-Latin titles and the filename"
    Slugs keep `[a-z0-9]` only, so a title written entirely in Cyrillic slugs to `untitled` and every card would land in `tasker-untitled.md` — the second one refusing to overwrite the first. Pass `--out` for those, or name the file yourself.

## Examples

Defaults — a 30-minute Go coding task for middle and senior levels:

```bash
srekit tasker --title "Channels and select" --stdout
```

A short theory question for a junior:

```bash
srekit tasker -T "What GOMAXPROCS does" --topic go --level junior \
  --format theory --duration 10
```

Into a collection, with the filename chosen by hand:

```bash
srekit tasker -T "Каналы и select" --out "tasks/Tasker - Каналы и select.md"
```

## Output

```markdown
---
id: "b0a1…"
creation_date: "2026-08-28T14:20:08+03:00"
type: simple_note
tags:
  - tasker
topic: "go"
level: [middle, senior]
format: "code"
duration: 30
---

# Tasker - Channels and select

## Описание (Description)

## Что хотим услышать (What we want to hear)
```

`level` is a YAML list and `duration` a number, not strings — the front matter of a card is read by the collection that holds it, and `level: [middle, senior]` filters where `level: "middle, senior"` does not. See [typed front matter values](../guides/custom-templates.md#typed-front-matter-values) if you want the same in your own artifacts.

## Section structure

Section headings render bilingually — Russian, with the English term in parentheses. Below they are given by stable `id` and English term; `srekit tasker -T X --json | jq -r '.sections[].title'` prints them as they appear in the document.

- Front matter: `id`, `creation_date`, `type: simple_note`, `tags`, `topic`, `level`, `format`, `duration`
- `description` — the task itself
- `expectations` — what a good answer sounds like

Both bodies are empty by default. That is deliberate: a placeholder would be text you delete on every single card.

## Template shape

`tasker` ships as a v1 YAML artifact (`internal/tmpl/templates/tasker.yaml`) — frontmatter, H1, sections (`description`, `expectations`). Template expressions reference `.Meta.<Field>` for `ID`, `Now`, `Title`, `Topic`, `Level` (a list), `Format`, `Duration` (a number). See [`srekit postmortem`](postmortem.md#customizing-the-artifact-v1-format-v0140) for the full schema reference.

## See also

- [`srekit task`](task.md) — the SRE investigation log. Similar name, unrelated document: `task` records a debugging session, `tasker` describes a task somebody else will solve.
- [Custom templates workflow](../guides/custom-templates.md) — override `tasker.yaml` to match a collection whose front matter differs.
