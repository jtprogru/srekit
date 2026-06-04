# srekit slo

Сгенерировать **SLO / SLI документ** для сервиса: цель, окно, latency budget, с embedded PromQL-примером.

## Синопсис

```bash
srekit slo --service NAME [flags]
```

## Флаги

| Флаг | Обязательный | Описание |
|---|---|---|
| `--service` | да | Имя сервиса |
| `--target` | нет | SLO-цель (например `99.9%`). Default: `99.9%` |
| `--window` | нет | Rolling window (например `30d`). Default: `30d` |
| `--latency` | нет | Latency budget (например `300ms`). Default: `300ms` |

Плюс [общие output-флаги](index.md#shared-output-flags). Default имя файла: `slo-<slug-of-service>.md`.

## Примеры

Defaults:

```bash
srekit slo --service api-gw --stdout
```

Жёстче target, длиннее window:

```bash
srekit slo --service api-gw --target 99.95% --window 90d --latency 250ms \
  --out slo-api-gw.md
```

## Структура секций

- Front matter: `title`, `service`, `target`, `window`, `id`
- Сервис (Service)
- SLI / SLO определения (success ratio, latency percentile)
- PromQL-пример — предрендерен со значениями service / window / latency
- Бюджет ошибок (Error budget)
- Связанные документы (Related docs) — ссылки на runbook, EBP, capacity

## Структура данных для шаблона

`slo` шипится как v1 YAML-артефакт (`internal/tmpl/templates/slo.yaml`) — frontmatter, H1, meta_bullets, секции про SLI/SLO определение, error budget и related-docs. Template-выражения обращаются к `.Meta.<Field>` для `ID`, `Service`, `Target`, `Window`, `LatencyTarget`, `Now`. См. [`srekit postmortem`](postmortem.md#customizing-the-artifact-v1-format-v0140) для полной схемы.

## См. также

- [`srekit ebp`](ebp.md) — политика бюджета ошибок, операционализующая SLO.
- [`srekit runbook`](runbook.md) — что делать когда burn-алерты сработали.
