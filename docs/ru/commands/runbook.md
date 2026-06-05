# srekit runbook

Сгенерировать **runbook** — operational playbook, к которому дежурный обращается, когда сработал алерт. Секции: Symptoms, Diagnose, Mitigate, Verify, Escalate.

## Синопсис

```bash
srekit runbook --title TITLE [flags]
```

## Флаги

| Флаг | Обязательный | Описание |
|---|---|---|
| `--title` | да | Тема runbook (часто — имя алерта) |
| `--service` | нет | Сервис, к которому относится runbook |
| `--alert` | нет | Имя / id конкретного алерта |

Плюс [общие output-флаги](index.md#shared-output-flags). Default имя файла: `runbook-<slug-of-title>.md`.

## Примеры

```bash
srekit runbook --title "p99 latency spike" --service api-gw --alert APIGwHighP99 \
  --out runbook-apigw-p99.md
```

В stdout без service:

```bash
srekit runbook --title "DB connection storm" --stdout
```

## Структура секций

- Front matter: `title`, `service`, `alert`, `tags`, `id`
- Симптомы (Symptoms)
- Диагностика (Diagnose) — что проверить, какие дашборды смотреть
- Митигация (Mitigate) — ограниченные шаги для остановки user impact
- Проверка (Verify) — как убедиться что митигация сработала
- Эскалация (Escalate) — кого звать, если runbook не помогает
- Ссылки (References)

## Структура данных для шаблона

`runbook` шипится как v1 YAML-артефакт (`internal/tmpl/templates/runbook.yaml`) — frontmatter, H1, meta_bullets, секции (`symptoms`, `severity_slo_impact`, `diagnose`, `mitigate`, `verify`, `after_the_fact`, `references`). Template-выражения обращаются к `.Meta.<Field>` для `ID`, `Title`, `Service`, `Alert`, `Now`. См. [`srekit postmortem`](postmortem.md#customizing-the-artifact-v1-format-v0140) для полной схемы.

## См. также

- [`srekit postmortem`](postmortem.md) — ретроспектива, которая пишется после инцидента, под который этот runbook.
