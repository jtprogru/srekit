# srekit postmortem

Сгенерировать **постмортем** в стиле Google SRE: severity, timeline, impact, detection / mitigation / root cause, action items, lessons. Билингвальные заголовки.

## Синопсис

```bash
srekit postmortem --title TITLE [flags]
```

## Флаги

| Флаг | Обязательный | Описание |
|---|---|---|
| `--title` | да | Тема инцидента |
| `--severity` | нет | `SEV-N` метка (free text) |
| `--start` | нет | Начало инцидента (timestamp) |
| `--end` | нет | Конец инцидента |
| `--owner` | нет | Владелец постмортема (тот, кто пишет) |

Плюс [общие output-флаги](index.md#shared-output-flags). Default имя файла: `postmortem-<YYYY-MM-DD>-<slug-of-title>.md` (дата — сегодняшняя, в часовом поясе системы).

## Примеры

Короткая форма:

```bash
srekit postmortem --title "API outage" --severity SEV-1 --stdout
```

С временами и владельцем:

```bash
srekit postmortem --title "API outage" --severity SEV-1 \
  --start 2026-05-06T08:00Z --end 2026-05-06T09:30Z \
  --owner "@oncall" --out postmortem-2026-05-06.md
```

Достать только метаданные для трекера:

```bash
srekit postmortem --title "API outage" --severity SEV-1 --json \
  | jq '{title: .title, severity: .severity, started: .start}'
```

## Структура секций

После рендеринга (RU первичный, EN в скобках):

- Front matter: `title`, `tags`, `severity`, `start`, `end`, `id`
- Сводка (Summary)
- Хронология (Timeline) — предзаполненный скелет таблицы
- Влияние (Impact)
- Обнаружение (Detection)
- Митигация (Mitigation)
- Корневая причина (Root Cause)
- Что прошло хорошо / плохо / повезло (What went well / poorly / lucky)
- Action items — таблица с owner / due / status
- Ссылки (References)

## Структура данных для шаблона

```go
struct {
    ID, Title, Severity, Start, End, Owner, Now string
}
```

## См. также

- [`srekit incident`](incident.md) — документ *во время* инцидента.
- [`srekit retro`](retro.md) — sprint-level ретроспектива.
