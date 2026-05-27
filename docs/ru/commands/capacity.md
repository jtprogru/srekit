# srekit capacity

Сгенерировать **capacity plan**: baseline-метрики, допущения роста, forecast, триггеры масштабирования, цель по headroom, зависимости, стоимость, риски.

## Синопсис

```bash
srekit capacity --service NAME [flags]
```

## Флаги

| Флаг | Обязательный | Описание |
|---|---|---|
| `--service` | да | Сервис, к которому относится план |
| `--horizon` | нет | Горизонт планирования (например `3m`, `6m`, `1y`). Default: `6m` |

Плюс [общие output-флаги](index.md#shared-output-flags). Default имя файла: `capacity-<slug-of-service>.md`.

## Примеры

```bash
srekit capacity --service api-gw --horizon 6m --out capacity-api-gw.md
```

Длиннее горизонт:

```bash
srekit capacity --service db-replicas --horizon 1y --stdout
```

## Структура секций

- Front matter: `title`, `service`, `horizon`, `id`
- Базовые метрики (Baseline metrics) — RPS, p99, CPU, memory, disk, network
- Допущения роста (Growth assumptions)
- Прогноз (Forecast) — таблица по месяцу / кварталу
- Триггеры масштабирования (Scale-up triggers)
- Цель по headroom (Headroom target)
- Зависимости (Dependencies)
- Стоимость (Cost)
- Риски (Risks)
- Ссылки (References)

## Структура данных для шаблона

```go
struct {
    ID, Service, Horizon, Now string
}
```

## См. также

- [`srekit slo`](slo.md) — capacity-ограничения часто следуют из latency-SLO.
- [`srekit rfc`](rfc.md) — RFC, когда capacity-решения требуют buy-in.
