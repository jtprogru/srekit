# srekit incident

Сгенерировать **live-incident report** — документ, который заполняется
*во время* активного инцидента (статус, лид, коммуникации, лог апдейтов).
Отличается от [`srekit postmortem`](postmortem.md), который пишется
*после*.

## Синопсис

```bash
srekit incident --title TITLE [flags]
```

## Флаги

| Флаг | Обязательный | Описание |
|---|---|---|
| `--title` | да | Краткое описание инцидента |
| `--severity` | нет | `SEV-1`, `SEV-2`, `SEV-3`, `SEV-4` (free text) |
| `--status` | нет | `investigated`, `active`, `contained`, `resolved`. Валидируется. |
| `--lead` | нет | Лид инцидента (handle / @mention) |
| `--comms` | нет | Канал коммуникаций (Slack room, status page) |
| `--start` | нет | Timestamp начала (например `2026-05-06T08:00Z`) |

Плюс [общие output-флаги](index.md#shared-output-flags). Default
имя файла: `incident-<slug-of-title>.md`.

## Примеры

Минимум, в stdout:

```bash
srekit incident --title "API down" --severity SEV-1 --lead alice --stdout
```

Полностью заполненный:

```bash
srekit incident --title "API down" --severity SEV-1 --status active \
  --lead "@alice" --comms "#inc-api-down" --start 2026-05-06T08:00Z \
  --out incident-api-down.md
```

Невалидный статус отклоняется до записи файла:

```bash
srekit incident --title "X" --status broken --stdout
# Error: --status must be one of investigated|active|contained|resolved
```

## Структура данных для шаблона

```go
struct {
    ID, Title, Severity, Status, Lead, Comms, Start, Now string
}
```

## См. также

- [`srekit postmortem`](postmortem.md) — post-mortem after-action отчёт.
- [`srekit runbook`](runbook.md) — operational playbook *до* инцидента.
