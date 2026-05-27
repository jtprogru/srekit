# srekit ebp

Generate an **Error Budget Policy** with tiered actions (Yellow / Orange / Red), exceptions, and escalation paths. Pairs with [`srekit slo`](slo.md).

## Synopsis

```bash
srekit ebp --service NAME [flags]
```

## Flags

| Flag | Required | Description |
|---|---|---|
| `--service` | yes | Service this policy applies to |
| `--owner` | no | Policy owner (team or individual) |

Plus the [shared output flags](index.md#shared-output-flags). Default filename: `ebp-<slug-of-service>.md`.

## Examples

```bash
srekit ebp --service api-gw --owner "@platform" --out ebp-api-gw.md
```

To stdout:

```bash
srekit ebp --service api-gw --stdout
```

## Section structure

- Front matter: `title`, `service`, `owner`, `id`
- Цель политики (Policy goal)
- Tiered actions:
    - 🟡 Yellow — соблюдать SLO, без feature freeze
    - 🟠 Orange — приоритет на стабильность
    - 🔴 Red — feature freeze, фокус на reliability
- Исключения (Exceptions)
- Эскалация (Escalation)
- Связанные SLO (Related SLOs)
- Ссылки (References)

## Template shape

```go
struct {
    ID, Service, Owner, Now string
}
```

## See also

- [`srekit slo`](slo.md) — define the SLO this policy reacts to.
- [`srekit oncall-report`](oncall-report.md) — operational view of EBP impact.
