# srekit oncall-report

Сгенерировать **недельный отчёт дежурства** для команды — pages-таблица, breakdown по toil, гигиена алертов. Период по умолчанию — текущая неделя (Пн–Вс).

## Синопсис

```bash
srekit oncall-report --team NAME [flags]
```

## Флаги

| Флаг | Обязательный | Описание |
|---|---|---|
| `--team` | да | Имя команды |
| `--start` | нет | Начало периода (YYYY-MM-DD). Default: понедельник этой недели. |
| `--end` | нет | Конец периода. Default: воскресенье этой недели. |
| `--author` | нет | Имя дежурного (default: цепочка yaml/git config) |
| `--email` | нет | Email дежурного |

Плюс [общие output-флаги](index.md#shared-output-flags). Default имя файла: `oncall-<slug-of-team>-<start>.md`.

## Примеры

Текущая неделя:

```bash
srekit oncall-report --team platform --out oncall.md
```

Конкретное окно:

```bash
srekit oncall-report --team platform \
  --start 2026-05-04 --end 2026-05-10 \
  --out oncall-platform-2026-W19.md
```

## Границы недели

Default-период начинается с понедельника и кончается воскресеньем. Часы под капотом — `internal/clock.Now` (overridable для тестов; production-бинарь использует `time.Now()`).

Если запустил отчёт в воскресенье 2026-05-10 — получишь `2026-05-04 → 2026-05-10`. Регрессионный тест пинит это поведение.

## Структура секций

- Front matter: `title`, `team`, `start`, `end`, `id`
- Сводка (Summary)
- Дежурство (On-call shift) — кто был на смене, hand-offs
- Pages — таблица инцидентов с severity / runbook колонками
- Тойл (Toil) — повторяющаяся работа, которую стоит автоматизировать
- Гигиена алертов (Alert hygiene) — шумные / тихие алерты для тюнинга
- Уроки (Lessons)
- Ссылки (References)

## Структура данных для шаблона

`oncall-report` шипится как v1 YAML-артефакт (`internal/tmpl/templates/oncall.yaml`) — frontmatter, H1, meta_bullets, секции (`tl_dr`, `pages`, `incidents`, `toil`, `alert_hygiene`, `follow_ups_for_next_on_caller`, `wins`, `references`). Template-выражения обращаются к `.Meta.<Field>` для `ID`, `Team`, `Start`, `End`, `Now`, `Author.Name`, `Author.Email`. См. [`srekit postmortem`](postmortem.md#customizing-the-artifact-v1-format-v0140) для полной схемы.

## См. также

- [`srekit ebp`](ebp.md) — политика бюджета ошибок, влияющая на нагрузку дежурного.
