# srekit incident

Сгенерировать **live-incident report** — документ, который заполняется *во время* активного инцидента (статус, лид, коммуникации, лог апдейтов). Отличается от [`srekit postmortem`](postmortem.md), который пишется *после*.

## Синопсис

```bash
srekit incident --title TITLE [flags]
```

## Флаги

| Флаг | Обязательный | Описание |
|---|---|---|
| `--title`, `-T` | да | Краткое описание инцидента |
| `--severity` | нет | `SEV-1`, `SEV-2`, `SEV-3`, `SEV-4` (free text). Default: `SEV-2`. |
| `--status` | нет | `investigated`, `active`, `contained`, `resolved`. Валидируется. Default: `active`. |
| `--lead` | нет | Лид инцидента (handle / @mention) |

Плюс [общие output-флаги](index.md#shared-output-flags). Default имя файла: `incident-<slug-of-title>.md`.

## Примеры

Минимум, в stdout:

```bash
srekit incident --title "API down" --severity SEV-1 --lead alice --stdout
```

С конкретным статусом:

```bash
srekit incident -T "DB failover" --status investigated --lead "@alice" \
  --out incident-db-failover.md
```

Невалидный статус отклоняется до записи файла:

```bash
srekit incident -T "X" --status broken --stdout
# Error: invalid --status "broken" (investigated, active, contained, resolved)
```

## Структура данных для шаблона

`incident` шипится как v1 YAML-артефакт (`internal/tmpl/templates/incident.yaml`) — frontmatter, H1, meta_bullets, секции (`current_impact`, `affected_services`, `updates_log`, `current_actions`, `hypotheses`, `customer_comms`, `after_resolve`, `references`). Template-выражения внутри YAML обращаются к `.Meta.<Field>` для `ID`, `Title`, `Severity`, `Lead`, `Status`, `Now`. См. [`srekit postmortem`](postmortem.md#customizing-the-artifact-v1-format-v0140) для полной схемы.

## См. также

- [`srekit postmortem`](postmortem.md) — post-mortem after-action отчёт.
- [`srekit runbook`](runbook.md) — operational playbook *до* инцидента.
