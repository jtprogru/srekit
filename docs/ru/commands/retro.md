# srekit retro

Сгенерировать **спринт-ретро** скаффолд в формате Start / Stop / Continue.

## Синопсис

```bash
srekit retro --team NAME [flags]
```

## Флаги

| Флаг | Обязательный | Описание |
|---|---|---|
| `--team` | да | Имя команды |
| `--sprint` | нет | Идентификатор спринта (например `2026-W19`). Default: сегодняшняя дата. |

Плюс [общие output-флаги](index.md#shared-output-flags). Default имя файла: `retro-<slug-of-team>-<sprint>.md`.

## Примеры

```bash
srekit retro --team platform --sprint 2026-W19 --out retro-platform-W19.md
```

В stdout:

```bash
srekit retro --team platform --stdout
```

## Структура секций

- Front matter: `title`, `team`, `sprint`, `id`
- Контекст (Context) — sprint summary, ключевые метрики
- ✅ Start — что начнём делать
- ⛔ Stop — что прекратим
- 🔁 Continue — что продолжаем
- Action items — owner / due / status
- Ссылки (References)

## Структура данных для шаблона

`retro` шипится как v1 YAML-артефакт (`internal/tmpl/templates/retro.yaml`) — frontmatter, H1, meta_bullets, секции (Start / Stop / Continue, action items, references). Template-выражения обращаются к `.Meta.<Field>` для `ID`, `Team`, `Sprint`, `Now`. См. [`srekit postmortem`](postmortem.md#customizing-the-artifact-v1-format-v0140) для полной схемы.

## См. также

- [`srekit oncall-report`](oncall-report.md) — недельная гранулярность vs sprint-гранулярность.
- [`srekit postmortem`](postmortem.md) — per-incident ретро.
