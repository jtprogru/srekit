# srekit slo

Generate an **SLO / SLI document** for a service: objective, window, latency budget, with an embedded PromQL example.

## Synopsis

```bash
srekit slo --service NAME [flags]
```

## Flags

| Flag | Required | Description |
|---|---|---|
| `--service` | yes | Service name |
| `--target` | no | SLO target (e.g. `99.9%`). Default: `99.9%` |
| `--window` | no | Rolling window (e.g. `30d`). Default: `30d` |
| `--latency` | no | Latency budget (e.g. `300ms`). Default: `300ms` |

Plus the [shared output flags](index.md#shared-output-flags). Default filename: `slo-<slug-of-service>.md`.

## Examples

Defaults:

```bash
srekit slo --service api-gw --stdout
```

Tighter target, longer window:

```bash
srekit slo --service api-gw --target 99.95% --window 90d --latency 250ms \
  --out slo-api-gw.md
```

## Section structure

- Front matter: `id`, `creation_date`, `modification_date`, `type: slo`, `service`, `tags`
- Определение SLI (SLI definition) — success ratio and latency percentile, with the PromQL pre-rendered from the service / window / latency values
- Цель SLO (SLO target)
- Бюджет ошибок (Error budget)
- Последствия истощения бюджета (Consequences of budget exhaustion) — the hand-off point to [`srekit ebp`](ebp.md)
- Зависимости (Dependencies)
- Регулярность пересмотра (Review cadence)
- Ссылки (References)

## Template shape

`slo` ships as a v1 YAML artifact (`internal/tmpl/templates/slo.yaml`) — frontmatter, H1, meta_bullets, sections (`sli_definition`, `slo_target`, `error_budget`, `consequences_of_budget_exhaustion`, `dependencies`, `review_cadence`, `references`). Template expressions reference `.Meta.<Field>` for `ID`, `Service`, `Target`, `Window`, `LatencyTarget`, `Now`. See [`srekit postmortem`](postmortem.md#customizing-the-artifact-v1-format-v0140) for the full schema reference.

## See also

- [`srekit ebp`](ebp.md) — error budget policy that operationalizes this SLO.
- [`srekit runbook`](runbook.md) — what to do when burn alerts fire.
