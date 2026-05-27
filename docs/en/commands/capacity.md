# srekit capacity

Generate a **capacity plan**: baseline metrics, growth assumptions,
forecast, scale-up triggers, headroom targets, dependencies, cost, risks.

## Synopsis

```bash
srekit capacity --service NAME [flags]
```

## Flags

| Flag | Required | Description |
|---|---|---|
| `--service` | yes | Service the plan covers |
| `--horizon` | no | Planning horizon (e.g. `3m`, `6m`, `1y`). Default: `6m` |

Plus the [shared output flags](index.md#shared-output-flags). Default
filename: `capacity-<slug-of-service>.md`.

## Examples

```bash
srekit capacity --service api-gw --horizon 6m --out capacity-api-gw.md
```

Longer horizon:

```bash
srekit capacity --service db-replicas --horizon 1y --stdout
```

## Section structure

- Front matter: `title`, `service`, `horizon`, `id`
- Базовые метрики (Baseline metrics) — RPS, p99, CPU, memory, disk, network
- Допущения роста (Growth assumptions)
- Прогноз (Forecast) — table by month / quarter
- Триггеры масштабирования (Scale-up triggers)
- Цель по headroom (Headroom target)
- Зависимости (Dependencies)
- Стоимость (Cost)
- Риски (Risks)
- Ссылки (References)

## Template shape

```go
struct {
    ID, Service, Horizon, Now string
}
```

## See also

- [`srekit slo`](slo.md) — capacity constraints often follow from latency
  SLOs.
- [`srekit rfc`](rfc.md) — file an RFC when capacity choices need broad
  buy-in.
