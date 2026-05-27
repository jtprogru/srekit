# srekit oncall-report

Generate a **weekly on-call report** for a team — pages table, toil
breakdown, alert hygiene. Period defaults to the current Monday–Sunday
week.

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

Plus the [shared output flags](index.md#shared-output-flags). Default
filename: `oncall-<slug-of-team>-<start>.md`.

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

The default period uses Monday as the start. Sunday is the end. The
underlying clock is `internal/clock.Now` (overridable for tests; the
production binary uses wall-clock `time.Now()`).

If you ran the report on Sunday 2026-05-10, you'd get
`2026-05-04 → 2026-05-10` — a regression test pins this exact boundary
behavior.

## Section structure

- Front matter: `title`, `team`, `start`, `end`, `id`
- Сводка (Summary)
- Дежурство (On-call shift) — who was on, hand-offs
- Pages — table of incidents with severity/runbook columns
- Тойл (Toil) — repetitive work that should be automated
- Гигиена алертов (Alert hygiene) — noisy / silent alerts that need tuning
- Уроки (Lessons)
- Ссылки (References)

## Template shape

```go
struct {
    ID, Team, Start, End, Now string
    Author meta.Author
}
```

## See also

- [`srekit ebp`](ebp.md) — error budget policy that informs on-call load.
- [`srekit retro`](retro.md) — sprint-level retro covering on-call signals.
