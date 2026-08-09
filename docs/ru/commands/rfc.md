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
| `--author` | нет | Переопределить автора (см. [резолв автора](../guides/configuration.md#identity)) |
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

- Front matter: `id`, `creation_date`, `decision_date`, `status`, `type: rfc`, `title`, `deciders`, `supersedes`, `superseded_by`, `tags`
- Контекст (Context)
- Решение (Decision)
- Рассмотренные альтернативы (Alternatives considered)
- Последствия (Consequences) — разбито на Плюсы / Минусы / Нейтральное
- Ссылки (References)

## Структура данных для шаблона

`rfc` шипится как v1 YAML-артефакт (`internal/tmpl/templates/rfc.yaml`) — frontmatter (`id`, `status`, `type: rfc`, `title`, `deciders`, `supersedes`, …), H1 (`RFC-<shortID> — <title>`), meta_bullets, секции (`context`, `decision`, `alternatives_considered`, `consequences`, `references`). Template-выражения обращаются к `.Meta.<Field>` для `ID`, `Title`, `Status`, `Now`, `Author.Name`, `Author.Email`. См. [`srekit postmortem`](postmortem.md#customizing-the-artifact-v1-format-v0140) для полной схемы.

## См. также

- [Резолв автора](../guides/configuration.md#identity)
