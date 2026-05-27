# srekit runbook

Generate a **runbook** — the operational playbook on-call reaches for when
an alert fires. Sections: Symptoms, Diagnose, Mitigate, Verify, Escalate.

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

Plus the [shared output flags](index.md#shared-output-flags). Default
filename: `runbook-<slug-of-title>.md`.

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

```go
struct {
    ID, Title, Service, Alert, Now string
}
```

## See also

- [`srekit incident`](incident.md) — the live doc you create when the
  runbook isn't enough.
