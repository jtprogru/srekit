# srekit incident

Generate a **live-incident report** — the doc you fill in *during* an active incident (status, lead, communications, update log). Distinct from [`srekit postmortem`](postmortem.md), which is written *after*.

## Synopsis

```bash
srekit incident --title TITLE [flags]
```

## Flags

| Flag | Required | Description |
|---|---|---|
| `--title`, `-T` | yes | Short incident description |
| `--severity` | no | `SEV-1`, `SEV-2`, `SEV-3`, `SEV-4` (free text). Default: `SEV-2`. |
| `--status` | no | One of `investigated`, `active`, `contained`, `resolved`. Validated; unknown values are rejected. Default: `active`. |
| `--lead` | no | Incident lead handle / @mention |

Plus the [shared output flags](index.md#shared-output-flags). Default filename: `incident-<slug-of-title>.md`.

## Examples

Bare minimum, to stdout:

```bash
srekit incident --title "API down" --severity SEV-1 --lead alice --stdout
```

Specific status:

```bash
srekit incident -T "DB failover" --status investigated --lead "@alice" \
  --out incident-db-failover.md
```

Invalid status is rejected up front (no half-written file):

```bash
srekit incident -T "X" --status broken --stdout
# Error: invalid --status "broken" (investigated, active, contained, resolved)
```

## Template shape

`incident` ships as a v1 YAML artifact (`internal/tmpl/templates/incident.yaml`) — frontmatter, H1, meta_bullets, sections (`current_impact`, `affected_services`, `updates_log`, `current_actions`, `hypotheses`, `customer_comms`, `after_resolve`, `references`). Template expressions inside the YAML reference `.Meta.<Field>` for `ID`, `Title`, `Severity`, `Lead`, `Status`, `Now`. See [`srekit postmortem`](postmortem.md#customizing-the-artifact-v1-format-v0140) for the full schema reference.

## See also

- [`srekit postmortem`](postmortem.md) — the after-action report you write once the incident closes.
- [`srekit runbook`](runbook.md) — the operational playbook you reach for before incidents.
