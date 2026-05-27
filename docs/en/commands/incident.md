# srekit incident

Generate a **live-incident report** — the doc you fill in *during* an active
incident (status, lead, communications, update log). Distinct from
[`srekit postmortem`](postmortem.md), which is written *after*.

## Synopsis

```bash
srekit incident --title TITLE [flags]
```

## Flags

| Flag | Required | Description |
|---|---|---|
| `--title` | yes | Short incident description |
| `--severity` | no | `SEV-1`, `SEV-2`, `SEV-3`, `SEV-4` (free text accepted) |
| `--status` | no | One of `investigated`, `active`, `contained`, `resolved`. Validated; unknown values are rejected. |
| `--lead` | no | Incident lead handle / @mention |
| `--comms` | no | Communications channel (Slack room, status page) |
| `--start` | no | Start timestamp (e.g. `2026-05-06T08:00Z`) |

Plus the [shared output flags](index.md#shared-output-flags). Default
filename: `incident-<slug-of-title>.md`.

## Examples

Bare minimum, to stdout:

```bash
srekit incident --title "API down" --severity SEV-1 --lead alice --stdout
```

Fully populated:

```bash
srekit incident --title "API down" --severity SEV-1 --status active \
  --lead "@alice" --comms "#inc-api-down" --start 2026-05-06T08:00Z \
  --out incident-api-down.md
```

Invalid status is rejected up front (no half-written file):

```bash
srekit incident --title "X" --status broken --stdout
# Error: --status must be one of investigated|active|contained|resolved
```

## Template shape

```go
struct {
    ID, Title, Severity, Status, Lead, Comms, Start, Now string
}
```

## See also

- [`srekit postmortem`](postmortem.md) — the after-action report you write
  once the incident closes.
- [`srekit runbook`](runbook.md) — the operational playbook you reach for
  before incidents.
