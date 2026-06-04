# Recipes

Concrete workflows that combine srekit with real tooling. Each recipe is copy-pasteable; adjust paths, team names, and IDs to your context.

---

## Generate a postmortem and post the metadata to a tracker

```bash
TITLE="API outage"
SEV="SEV-1"
START="2026-05-06T08:00Z"
END="2026-05-06T09:30Z"

# Write the document
srekit postmortem --title "$TITLE" --severity "$SEV" \
  --start "$START" --end "$END" \
  --out "postmortem-$(date -u +%Y-%m-%d).md"

# Extract metadata and post it elsewhere
srekit postmortem --title "$TITLE" --severity "$SEV" \
  --start "$START" --end "$END" --json |
  jq '{title, severity: .severity, started_at: .start, ended_at: .end}' |
  curl -X POST https://tracker.example.com/api/incidents \
    -H 'Content-Type: application/json' -d @-
```

---

## Bulk-render runbooks for every service in a list

```bash
while IFS= read -r service; do
  srekit runbook --title "p99 spike" --service "$service" \
    --out "runbooks/$service-p99.md" --force
done < services.txt
```

---

## CI gate: fail if a custom template doesn't parse

`.github/workflows/templates.yaml`:

```yaml
name: templates
on: [push, pull_request]
jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - run: |
          curl -fsSL https://github.com/jtprogru/srekit/releases/latest/download/srekit_Linux_x86_64.tar.gz \
            | tar xz srekit
      - run: ./srekit templates validate ./templates
```

---

## Weekly on-call summary cron

`crontab -e`:

```cron
0 18 * * 0  cd ~/work && srekit oncall-report --team platform \
              --out "reports/oncall-$(date -u +%Y-W%V).md"
```

For Slack delivery instead of file:

```bash
srekit oncall-report --team platform --stdout |
  curl -F file=@- -F channels=oncall-summaries \
    -F token=$SLACK_TOKEN https://slack.com/api/files.upload
```

---

## Detect template drift after an `srekit` upgrade

After `brew upgrade srekit`:

```bash
srekit templates list --json |
  jq '[.[] | select(.status == "embedded-only" or .status == "customized") | .name]'
# ["task.yaml", "runbook.yaml"]   # things to look at
```

Or just diff:

```bash
srekit templates diff --name-only
# differs  runbook.yaml
# differs  capacity.yaml
```

Then `srekit templates upgrade` to 3-way-merge them in.

---

## Use a different identity per project

`~/.srekit.yaml` has your personal identity; in a work repo's `.envrc` (via direnv):

```bash
export SREKIT_AUTHOR="Mikhail Savin"
export SREKIT_EMAIL="m.savin@work.example.com"
```

`cd` into the work repo and srekit auto-picks up the work identity for license / RFC / on-call docs.

---

## Pin srekit version per project

Some teams want all engineers to use the same srekit version for reproducibility. Pin in a project script:

```bash
#!/usr/bin/env bash
# bin/srekit
set -euo pipefail
WANT=0.10.1
HAVE=$(srekit --version 2>&1 | awk '/srekit version:/ {print $3}')
if [[ "$HAVE" != "$WANT" ]]; then
  echo "srekit $WANT required (have $HAVE)" >&2; exit 1
fi
exec srekit "$@"
```

---

## Two repos, one templates source

Common for multi-repo orgs: one shared `sre-templates` repo, many consumers.

```bash
# In each repo's setup:
git clone git@github.com:acme/sre-templates ~/.acme/templates
echo "templates_dir: ~/.acme/templates" >> ~/.srekit.yaml

# To pull updates:
srekit templates pull
```

Need a starting point? [`jtprogru/sre-templates`](https://github.com/jtprogru/sre-templates) is a ready-to-clone example in the exact layout srekit expects — use it directly or fork it as the seed for your org's shared repo.

---

## See also

- [Custom templates workflow](guides/custom-templates.md) — the longer narrative.
- [JSON output](guides/json-output.md) — pipeline patterns.
