# srekit ebp

Сгенерировать **Error Budget Policy** с tiered actions (Yellow / Orange /
Red), исключениями и путями эскалации. Парится с [`srekit slo`](slo.md).

## Синопсис

```bash
srekit ebp --service NAME [flags]
```

## Флаги

| Флаг | Обязательный | Описание |
|---|---|---|
| `--service` | да | Сервис, к которому применима политика |
| `--owner` | нет | Владелец политики (команда или человек) |

Плюс [общие output-флаги](index.md#shared-output-flags). Default имя
файла: `ebp-<slug-of-service>.md`.

## Примеры

```bash
srekit ebp --service api-gw --owner "@platform" --out ebp-api-gw.md
```

В stdout:

```bash
srekit ebp --service api-gw --stdout
```

## Структура секций

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

## Структура данных для шаблона

```go
struct {
    ID, Service, Owner, Now string
}
```

## См. также

- [`srekit slo`](slo.md) — задать SLO, на который реагирует политика.
- [`srekit oncall-report`](oncall-report.md) — operational view влияния EBP.
