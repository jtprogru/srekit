# srekit runbook

Generate a **runbook** — the operational playbook on-call reaches for when an alert fires. Sections: Symptoms, Severity & SLO impact, Diagnose, Mitigate, Verify, After the fact, References.

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

- Front matter: `id`, `creation_date`, `modification_date`, `type: runbook`, `title`, `service`, `alert`, `tags`
- Симптомы (Symptoms) — what the alert actually looks like
- Тяжесть и влияние на SLO (Severity & SLO impact) — how bad this is, and against which budget
- Диагностика (Diagnose) — investigations to run, dashboards to check
- Смягчение (Mitigate) — bounded steps that stop user impact: immediate action, rollback, failover
- Проверка (Verify) — how to confirm the mitigation worked
- Постфактум (After the fact) — open a postmortem above a given SEV, fold surprises back into this runbook
- Ссылки (References)

## Template shape

`runbook` ships as a v1 YAML artifact (`internal/tmpl/templates/runbook.yaml`) — frontmatter, H1, meta_bullets, sections (`symptoms`, `severity_slo_impact`, `diagnose`, `mitigate`, `verify`, `after_the_fact`, `references`). Template expressions reference `.Meta.<Field>` for `ID`, `Title`, `Service`, `Alert`, `Now`. See [`srekit postmortem`](postmortem.md#customizing-the-artifact-v1-format-v0140) for the full schema reference.

## See also

- [`srekit postmortem`](postmortem.md) — the retrospective written after the incident the runbook was for.
