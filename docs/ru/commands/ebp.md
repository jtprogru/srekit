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

- Front matter: `id`, `creation_date`, `modification_date`, `type: error-budget-policy`, `service`, `tags`
- Назначение (Purpose) — зачем политика существует: договориться о действиях заранее, а не в момент инцидента
- Триггеры (Triggers) — таблица «состояние бюджета → условие»: зелёный (< 50 % потрачено), жёлтый (50–75 %), оранжевый (75–100 %), красный (исчерпан)
- Действия по уровням (Tiered actions) — что команда реально делает на жёлтом / оранжевом / красном
- Исключения (Exceptions)
- Эскалация (Escalation)
- Пересмотр (Review)
- Ссылки (References)

## Структура данных для шаблона

`ebp` шипится как v1 YAML-артефакт (`internal/tmpl/templates/ebp.yaml`) — frontmatter, H1, meta_bullets, секции (`purpose`, `triggers`, `tiered_actions`, `exceptions`, `escalation`, `review`, `references`). Template-выражения обращаются к `.Meta.<Field>` для `ID`, `Service`, `Now`. См. [`srekit postmortem`](postmortem.md#customizing-the-artifact-v1-format-v0140) для полной схемы.

(Owner / team — это fill-in в рендереных meta_bullets; флага `--owner` нет — отредактируй рендереный файл или свой кастомизированный `ebp.yaml` напрямую.)

## См. также

- [`srekit slo`](slo.md) — задать SLO, на который реагирует политика.
- [`srekit oncall-report`](oncall-report.md) — operational view влияния EBP.
