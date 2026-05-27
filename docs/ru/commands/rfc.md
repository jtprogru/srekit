# srekit rfc

Сгенерировать **RFC / ADR** скаффолд с секциями Context, Decision, Alternatives, Consequences, References. Поле status валидируется.

## Синопсис

```bash
srekit rfc --title TITLE [flags]
```

## Флаги

| Флаг | Обязательный | Описание |
|---|---|---|
| `--title` | да | Тема RFC |
| `--status` | нет | `proposed`, `accepted`, `rejected`, `superseded`, `deprecated`. Default: `proposed`. |
| `--author` | нет | Переопределить автора (резолв как у [`license`](license.md#author-resolution)) |
| `--email` | нет | Переопределить email |

Плюс [общие output-флаги](index.md#shared-output-flags). Default имя файла: `rfc-<slug-of-title>.md`.

## Примеры

```bash
srekit rfc --title "Migrate to gRPC" --stdout
```

Apgrade с принятым решением:

```bash
srekit rfc --title "Migrate to gRPC" --status accepted --out rfc-grpc.md
```

Невалидный status отклоняется:

```bash
srekit rfc --title "X" --status maybe --stdout
# Error: --status must be one of proposed|accepted|rejected|superseded|deprecated
```

## Структура секций

- Front matter: `title`, `status`, `tags`, `id`
- Контекст (Context)
- Решение (Decision)
- Альтернативы (Alternatives)
- Последствия (Consequences) — разбито на Positive / Negative / Neutral
- Ссылки (References)

## Структура данных для шаблона

```go
struct {
    ID, Title, Status, Now string
    Author meta.Author // {Name, Email}
}
```

## См. также

- [Резолв автора](license.md#author-resolution)
