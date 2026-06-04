# srekit runbook

Generate a **runbook** — the operational playbook on-call reaches for when an alert fires. Sections: Symptoms, Diagnose, Mitigate, Verify, Escalate.

## Synopsis

```bash
srekit runbook --title TITLE [flags]
```

## Flags

| Flag | Required | Description |
|---|---|---|
| `--title` | yes | Runbook subject (often the alert name) |
| `--service` | no | Service this runbook covers |
| `--alert` | no | Specific alert name / id |

Plus the [shared output flags](index.md#shared-output-flags). Default filename: `runbook-<slug-of-title>.md`.

## Examples

```bash
srekit runbook --title "p99 latency spike" --service api-gw --alert APIGwHighP99 \
  --out runbook-apigw-p99.md
```

To stdout, no service binding:

```bash
srekit runbook --title "DB connection storm" --stdout
```

## Section structure

- Front matter: `title`, `service`, `alert`, `tags`, `id`
- Симптомы (Symptoms)
- Диагностика (Diagnose) — investigations to run, dashboards to check
- Митигация (Mitigate) — bounded steps that stop user impact
- Проверка (Verify) — how to confirm the mitigation worked
- Эскалация (Escalate) — who to page if the runbook doesn't resolve it
- Ссылки (References)

## Template shape

`runbook` ships as a v1 YAML artifact (`internal/tmpl/templates/runbook.yaml`) — frontmatter, H1, meta_bullets, sections (`symptoms`, `severity_slo_impact`, `diagnose`, `mitigate`, `verify`, `after_the_fact`, `references`). Template expressions reference `.Meta.<Field>` for `ID`, `Title`, `Service`, `Alert`, `Now`. See [`srekit postmortem`](postmortem.md#customizing-the-artifact-v1-format-v0140) for the full schema reference.

## See also

- [`srekit incident`](incident.md) — the live doc you create when the runbook isn't enough.
