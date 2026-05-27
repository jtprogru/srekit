# srekit retro

Generate a **sprint retro** scaffold in Start / Stop / Continue format.

## Synopsis

```bash
srekit retro --team NAME [flags]
```

## Flags

| Flag | Required | Description |
|---|---|---|
| `--team` | yes | Team name |
| `--sprint` | no | Sprint identifier (e.g. `2026-W19`). Default: today's date. |

Plus the [shared output flags](index.md#shared-output-flags). Default
filename: `retro-<slug-of-team>-<sprint>.md`.

## Examples

```bash
srekit retro --team platform --sprint 2026-W19 --out retro-platform-W19.md
```

To stdout:

```bash
srekit retro --team platform --stdout
```

## Section structure

- Front matter: `title`, `team`, `sprint`, `id`
- Контекст (Context) — sprint summary, key metrics
- ✅ Start — что начнём делать
- ⛔ Stop — что прекратим
- 🔁 Continue — что продолжаем
- Action items — owner / due / status
- Ссылки (References)

## Template shape

```go
struct {
    ID, Team, Sprint, Now string
}
```

## See also

- [`srekit oncall-report`](oncall-report.md) — weekly granularity vs
  sprint granularity.
- [`srekit postmortem`](postmortem.md) — per-incident retros.
