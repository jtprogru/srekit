# srekit

**srekit** is a CLI that generates the text artifacts SREs deal with every day — investigation logs, postmortems, runbooks, RFCs, on-call reports, SLOs, error budget policies, and changelogs.

A template is not a Markdown file. It is a **v1 YAML artifact** (`postmortem.yaml`, `slo.yaml`, …) that declares frontmatter, an H1, meta bullets and a list of typed sections; srekit composes the Markdown from that declaration. The shipped set is compiled into the binary, so a fresh install renders everything with no files on disk and no network — and a directory of your own artifacts, kept under your own git remote, overrides it file by file, with anything you have not overridden falling back transparently.

---

## Why srekit

SRE work produces a lot of structured documents that read 80 % the same and differ only in 20 % of the meat. Writing them from scratch wastes time and invariably misses sections you wish you had filled in. A generator that ships opinionated, peer-reviewed templates removes the boilerplate so you spend attention on the actual incident, RFC, or postmortem.

## What you get

- **8 generator commands** — `task`, `postmortem`, `rfc`, `runbook`, `changelog`, `oncall-report`, `slo`, `ebp`.
- **Templates as your contract** — every artifact is bilingual (Russian headings + English technical terms), ships compiled into the binary, and is overridable via a custom directory under your own git remote.
- **Changelog maintenance, not just scaffolding** — `changelog release` cuts a version out of `[Unreleased]` and `changelog validate` lints an existing `CHANGELOG.md` against Keep a Changelog.
- **Full template lifecycle** — `templates init / pull / list / validate / diff / upgrade / migrate`, with a true 3-way merge on upgrade.
- **JSON output** — every generator supports `--json` for piping into `jq` and other tools.
- **A read-only environment check** — `srekit doctor` reports which config file is actually read, where the templates directory resolves, and whether your artifacts still parse.
- **Deterministic** — no network calls at all, no hidden state. Author/email/repo resolved from local config and `git`.

## At a glance

```bash
# Scaffold a postmortem with today's date and your git identity baked in
srekit postmortem --title "API outage" --severity SEV-1 \
  --start 2026-05-06T08:00Z --end 2026-05-06T09:30Z \
  --owner "@oncall" --out postmortem-2026-05-06.md

# Pipe a generator into jq for scripting
srekit task --title "Tail latency on api-gw" --json | jq -r '.meta.id'

# Manage your customized templates
srekit templates init     # scaffold your own copy
srekit templates list     # see what's customized
srekit templates upgrade  # 3-way merge in new built-in content
```

## Next steps

- **[Getting started](getting-started.md)** — install, first command, first config.
- **[Commands overview](commands/index.md)** — the full command surface.
- **[Custom templates workflow](guides/custom-templates.md)** — keep your own templates under git and 3-way-merge upstream changes.
- **[jtprogru/sre-templates](https://github.com/jtprogru/sre-templates)** — a ready-to-use templates repo in exactly the layout srekit expects; clone it, point srekit at it, or fork it as a starting point.
- **[Recipes](recipes.md)** — practical pipelines (jq, on-call rotation, CI).
