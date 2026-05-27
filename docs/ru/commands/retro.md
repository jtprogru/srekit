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

Плюс [общие output-флаги](index.md#shared-output-flags). Default имя
файла: `retro-<slug-of-team>-<sprint>.md`.

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

```go
struct {
    ID, Team, Sprint, Now string
}
```

## См. также

- [`srekit oncall-report`](oncall-report.md) — недельная гранулярность vs
  sprint-гранулярность.
- [`srekit postmortem`](postmortem.md) — per-incident ретро.
