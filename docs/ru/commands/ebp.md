# srekit ebp

Сгенерировать **Error Budget Policy** с tiered actions (Yellow / Orange / Red), исключениями и путями эскалации. Парится с [`srekit slo`](slo.md).

## Синопсис

```bash
srekit ebp --service NAME [flags]
```

## Флаги

| Флаг | Обязательный | Описание |
|---|---|---|
| `--service` | да | Сервис, к которому применима политика |

Плюс [общие output-флаги](index.md#shared-output-flags). Default имя файла: `ebp-<slug-of-service>.md`.

## Примеры

```bash
srekit ebp --service api-gw --out ebp-api-gw.md
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

`ebp` шипится как v1 YAML-артефакт (`internal/tmpl/templates/ebp.yaml`) — frontmatter, H1, meta_bullets, секции (`purpose`, `triggers`, `tiered_actions`, `exceptions`, `escalation`, `review`, `references`). Template-выражения обращаются к `.Meta.<Field>` для `ID`, `Service`, `Now`. См. [`srekit postmortem`](postmortem.md#customizing-the-artifact-v1-format-v0140) для полной схемы.

(Owner / team — это fill-in в рендереных meta_bullets; флага `--owner` нет — отредактируй рендереный файл или свой кастомизированный `ebp.yaml` напрямую.)

## См. также

- [`srekit slo`](slo.md) — задать SLO, на который реагирует политика.
- [`srekit oncall-report`](oncall-report.md) — operational view влияния EBP.
