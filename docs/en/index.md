# srekit

**srekit** is a CLI that generates the text artifacts SREs deal with every day — investigation logs, live-incident reports, postmortems, runbooks, RFCs, on-call reports, SLOs, error budget policies, capacity plans, retros, changelogs, and licenses — all from embedded Markdown templates.

It is small, focused, and self-contained: a single binary with every template baked in via `//go:embed`, plus a transparent fallback path for custom templates you keep under your own git remote.

---

## Why srekit

SRE work produces a lot of structured documents that read 80 % the same and differ only in 20 % of the meat. Writing them from scratch wastes time and invariably misses sections you wish you had filled in. A generator that ships opinionated, peer-reviewed templates removes the boilerplate so you spend attention on the actual incident, RFC, or postmortem.

## What you get

- **12 generator commands** — `task`, `license`, `incident`, `postmortem`, `rfc`, `runbook`, `changelog`, `oncall-report`, `slo`, `ebp`, `capacity`, `retro`.
- **Templates as your contract** — every template is bilingual (Russian headings + English technical terms), shipped embedded, and overridable via a custom directory under your own git remote.
- **Full template lifecycle** — `templates init / pull / list / validate / diff / upgrade`, with a true 3-way merge on upgrade.
- **JSON output** — every generator supports `--json` for piping into `jq` and other tools.
- **Deterministic** — no network calls in the hot path, no hidden state. Author/email/repo resolved from local config and `git`.

## Quick taste

```bash
# Scaffold a postmortem with today's date and your git identity baked in
srekit postmortem --title "API outage" --severity SEV-1 \
  --start 2026-05-06T08:00Z --end 2026-05-06T09:30Z \
  --owner "@oncall" --out postmortem-2026-05-06.md

# Pipe a generator into jq for scripting
srekit task --title "Tail latency on api-gw" --json | jq '.ID'

# Manage your customized templates
srekit templates init     # scaffold your own copy
srekit templates list     # see what's customized
srekit templates upgrade  # 3-way merge in new embedded content
```

## Next steps

- **[Getting started](getting-started.md)** — install, first command, first config.
- **[Commands overview](commands/index.md)** — the full command surface.
- **[Custom templates workflow](guides/custom-templates.md)** — keep your own templates under git and 3-way-merge upstream changes.
- **[Recipes](recipes.md)** — practical pipelines (jq, on-call rotation, CI).
