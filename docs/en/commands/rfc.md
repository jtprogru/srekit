# srekit rfc

Generate an **RFC / ADR** scaffold with Context, Decision, Alternatives, Consequences, References. Status field is validated.

## Synopsis

```bash
srekit rfc --title TITLE [flags]
```

## Flags

| Flag | Required | Description |
|---|---|---|
| `--title` | yes | RFC subject |
| `--status` | no | One of `proposed`, `accepted`, `rejected`, `superseded`, `deprecated`. Default: `proposed`. |
| `--author` | no | Override author (see [author resolution](../guides/configuration.md#author-identity)) |
| `--email` | no | Override email |

Plus the [shared output flags](index.md#shared-output-flags). Default filename: `rfc-<slug-of-title>.md`.

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

Section headings render bilingually — Russian, with the English term in parentheses. Below they are given by stable `id` and English term; `srekit rfc -T X --json | jq -r '.sections[].title'` prints them as they appear in the document.

- Front matter: `id`, `creation_date`, `decision_date`, `status`, `type: rfc`, `title`, `deciders`, `supersedes`, `superseded_by`, `tags`
- `context` — Context
- `decision` — Decision
- `alternatives_considered` — Alternatives considered
- `consequences` — Consequences: split into Positive / Negative / Neutral sub-headings
- `references` — References

## Template shape

`rfc` ships as a v1 YAML artifact (`internal/tmpl/templates/rfc.yaml`) — frontmatter (`id`, `status`, `type: rfc`, `title`, `deciders`, `supersedes`, …), H1 (`RFC-<shortID> — <title>`), meta_bullets, sections (`context`, `decision`, `alternatives_considered`, `consequences`, `references`). Template expressions reference `.Meta.<Field>` for `ID`, `Title`, `Status`, `Now`, `Author.Name`, `Author.Email`. See [`srekit postmortem`](postmortem.md#customizing-the-artifact-v1-format-v0140) for the full schema reference.

## See also

- [Author resolution](../guides/configuration.md#author-identity)
