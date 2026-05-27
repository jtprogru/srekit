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

```go
struct {
    ID, Title, Service, Alert, Now string
}
```

## См. также

- [`srekit incident`](incident.md) — live doc когда runbook не помогает.
