# srekit ebp

Generate an **Error Budget Policy** with tiered actions (Yellow / Orange / Red), exceptions, and escalation paths. Pairs with [`srekit slo`](slo.md).

## Synopsis

```bash
srekit ebp --service NAME [flags]
```

## Flags

| Flag | Required | Description |
|---|---|---|
| `--service` | yes | Service this policy applies to |

Plus the [shared output flags](index.md#shared-output-flags). Default filename: `ebp-<slug-of-service>.md`.

## Examples

```bash
srekit ebp --service api-gw --out ebp-api-gw.md
```

To stdout:

```bash
srekit ebp --service api-gw --stdout
```

## Section structure

- Front matter: `id`, `creation_date`, `modification_date`, `type: error-budget-policy`, `service`, `tags`
- Назначение (Purpose) — why the policy exists: agree on the actions before the incident, not during it
- Триггеры (Triggers) — a table mapping budget state to condition: Green (< 50 % spent), Yellow (50–75 %), Orange (75–100 %), Red (exhausted)
- Действия по уровням (Tiered actions) — what the team actually does at Yellow / Orange / Red
- Исключения (Exceptions)
- Эскалация (Escalation)
- Пересмотр (Review)
- Ссылки (References)

## Template shape

`ebp` ships as a v1 YAML artifact (`internal/tmpl/templates/ebp.yaml`) — frontmatter, H1, meta_bullets, sections (`purpose`, `triggers`, `tiered_actions`, `exceptions`, `escalation`, `review`, `references`). Template expressions reference `.Meta.<Field>` for `ID`, `Service`, `Now`. See [`srekit postmortem`](postmortem.md#customizing-the-artifact-v1-format-v0140) for the full schema reference.

(Owner / team is a fill-in inside the rendered meta_bullets — there's no `--owner` flag; edit the rendered file or your customized `ebp.yaml` directly.)

## See also

- [`srekit slo`](slo.md) — define the SLO this policy reacts to.
- [`srekit oncall-report`](oncall-report.md) — operational view of EBP impact.
