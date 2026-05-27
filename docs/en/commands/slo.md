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

- Front matter: `title`, `service`, `target`, `window`, `id`
- Сервис (Service)
- SLI / SLO definitions (success ratio, latency percentile)
- PromQL example — pre-rendered with the service / window / latency values
- Бюджет ошибок (Error budget)
- Связанные документы (Related docs) — links to runbook, EBP, capacity

## Template shape

```go
struct {
    ID, Service, Target, Window, Latency, Now string
}
```

## See also

- [`srekit ebp`](ebp.md) — error budget policy that operationalizes this SLO.
- [`srekit runbook`](runbook.md) — what to do when burn alerts fire.
