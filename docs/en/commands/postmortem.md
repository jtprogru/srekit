# srekit postmortem

Generate a **postmortem** in Google-SRE style: severity, timeline, impact, detection / mitigation / root cause, action items, lessons. Bilingual headings.

## Synopsis

```bash
srekit postmortem --title TITLE [flags]
```

## Flags

| Flag | Required | Description |
|---|---|---|
| `--title` | yes | Incident subject |
| `--severity` | no | `SEV-N` label (free text) |
| `--start` | no | Incident start (timestamp string) |
| `--end` | no | Incident end |
| `--owner` | no | Postmortem owner (the person writing it up) |

Plus the [shared output flags](index.md#shared-output-flags). Default filename: `postmortem-<YYYY-MM-DD>-<slug-of-title>.md` (the date is today, in UTC if the system clock is UTC, otherwise local time).

## Examples

Short form:

```bash
srekit postmortem --title "API outage" --severity SEV-1 --stdout
```

With times and owner:

```bash
srekit postmortem --title "API outage" --severity SEV-1 \
  --start 2026-05-06T08:00Z --end 2026-05-06T09:30Z \
  --owner "@oncall" --out postmortem-2026-05-06.md
```

Extract just the metadata for a tracker:

```bash
srekit postmortem --title "API outage" --severity SEV-1 --json \
  | jq '{title: .title, severity: .severity, started: .start}'
```

## Section structure

After rendering, the doc has these sections (RU primary, EN parenthetical):

- Front matter: `title`, `tags`, `severity`, `start`, `end`, `id`
- Сводка (Summary)
- Хронология (Timeline) — pre-populated table skeleton
- Влияние (Impact)
- Обнаружение (Detection)
- Митигация (Mitigation)
- Корневая причина (Root Cause)
- Что прошло хорошо / плохо / повезло (What went well / poorly / lucky)
- Action items — table with owner / due / status columns
- Ссылки (References)

## Template shape

```go
struct {
    ID, Title, Severity, Start, End, Owner, Now string
}
```

## See also

- [`srekit incident`](incident.md) — the *during*-incident doc.
- [`srekit retro`](retro.md) — sprint-level retrospective.
