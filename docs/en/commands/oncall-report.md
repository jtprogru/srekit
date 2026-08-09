# srekit oncall-report

Generate a **weekly on-call report** for a team — pages table, toil breakdown, alert hygiene. Period defaults to the current Monday–Sunday week.

## Synopsis

```bash
srekit oncall-report --team NAME [flags]
```

## Flags

| Flag | Required | Description |
|---|---|---|
| `--team` | yes | Team name |
| `--start` | no | Period start (YYYY-MM-DD). Default: this week's Monday. |
| `--end` | no | Period end. Default: this week's Sunday. |
| `--author` | no | On-caller name (defaults to git config / yaml chain) |
| `--email` | no | On-caller email |

Plus the [shared output flags](index.md#shared-output-flags). Default filename: `oncall-<slug-of-team>-<start>.md`.

## Examples

Current week:

```bash
srekit oncall-report --team platform --out oncall.md
```

Specific window:

```bash
srekit oncall-report --team platform \
  --start 2026-05-04 --end 2026-05-10 \
  --out oncall-platform-2026-W19.md
```

## Week boundary

The default period uses Monday as the start. Sunday is the end. The underlying clock is `internal/clock.Now` (overridable for tests; the production binary uses wall-clock `time.Now()`).

If you ran the report on Sunday 2026-05-10, you'd get `2026-05-04 → 2026-05-10` — a regression test pins this exact boundary behavior.

## Section structure

Section headings render bilingually — Russian, with the English term in parentheses. Below they are given by stable `id` and English term; `srekit oncall-report --team X --json | jq -r '.sections[].title'` prints them as they appear in the document.

- Front matter: `id`, `creation_date`, `type: oncall-report`, `team`, `period_start`, `period_end`, `oncaller`, `tags`
- `tl_dr` — TL;DR: two or three sentences — quiet shift, busy one, or on fire
- `pages` — Pages: table of pages with alert / service / action columns, plus the actionable-vs-false-positive tally
- `incidents` — Incidents: links out to incident docs and postmortems
- `toil` — Toil: repetitive manual work that should be automated
- `alert_hygiene` — Alert hygiene: noisy / silent alerts that need tuning
- `follow_ups_for_next_on_caller` — Follow-ups for next on-caller
- `wins` — Wins
- `references` — References

## Template shape

`oncall-report` ships as a v1 YAML artifact (`internal/tmpl/templates/oncall.yaml`) — frontmatter, H1, meta_bullets, sections (`tl_dr`, `pages`, `incidents`, `toil`, `alert_hygiene`, `follow_ups_for_next_on_caller`, `wins`, `references`). Template expressions reference `.Meta.<Field>` for `ID`, `Team`, `Start`, `End`, `Now`, `Author.Name`, `Author.Email`. See [`srekit postmortem`](postmortem.md#customizing-the-artifact-v1-format-v0140) for the full schema reference.

## See also

- [`srekit ebp`](ebp.md) — error budget policy that informs on-call load.
