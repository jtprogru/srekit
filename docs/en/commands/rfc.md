# srekit rfc

Generate an **RFC / ADR** scaffold with Context, Decision, Alternatives,
Consequences, References. Status field is validated.

## Synopsis

```bash
srekit rfc --title TITLE [flags]
```

## Flags

| Flag | Required | Description |
|---|---|---|
| `--title` | yes | RFC subject |
| `--status` | no | One of `proposed`, `accepted`, `rejected`, `superseded`, `deprecated`. Default: `proposed`. |
| `--author` | no | Override author (resolved like [`license`](license.md#author-resolution)) |
| `--email` | no | Override email |

Plus the [shared output flags](index.md#shared-output-flags). Default
filename: `rfc-<slug-of-title>.md`.

## Examples

```bash
srekit rfc --title "Migrate to gRPC" --stdout
```

Accept upgrade:

```bash
srekit rfc --title "Migrate to gRPC" --status accepted --out rfc-grpc.md
```

Invalid status rejected:

```bash
srekit rfc --title "X" --status maybe --stdout
# Error: --status must be one of proposed|accepted|rejected|superseded|deprecated
```

## Section structure

- Front matter: `title`, `status`, `tags`, `id`
- Контекст (Context)
- Решение (Decision)
- Альтернативы (Alternatives)
- Последствия (Consequences) — split into Positive / Negative / Neutral
- Ссылки (References)

## Template shape

```go
struct {
    ID, Title, Status, Now string
    Author meta.Author // {Name, Email}
}
```

## See also

- [Author resolution](license.md#author-resolution)
