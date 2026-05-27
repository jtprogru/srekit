# Рецепты

Конкретные workflow'ы, где srekit совмещается с реальным тулчейном. Каждый рецепт — copy-paste; меняй пути, имена команд, ID под свой контекст.

---

## Сгенерить постмортем и запостить метаданные в трекер

```bash
TITLE="API outage"
SEV="SEV-1"
START="2026-05-06T08:00Z"
END="2026-05-06T09:30Z"

# Записать документ
srekit postmortem --title "$TITLE" --severity "$SEV" \
  --start "$START" --end "$END" \
  --out "postmortem-$(date -u +%Y-%m-%d).md"

# Извлечь метаданные и запостить
srekit postmortem --title "$TITLE" --severity "$SEV" \
  --start "$START" --end "$END" --json |
  jq '{title, severity: .Severity, started_at: .Start, ended_at: .End}' |
  curl -X POST https://tracker.example.com/api/incidents \
    -H 'Content-Type: application/json' -d @-
```

---

## Массово отрендерить runbook'и для всех сервисов из списка

```bash
while IFS= read -r service; do
  srekit runbook --title "p99 spike" --service "$service" \
    --out "runbooks/$service-p99.md" --force
done < services.txt
```

---

## CI-гейт: упасть если кастомный шаблон не парсится

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

## Cron — недельная сводка дежурства

`crontab -e`:

```cron
0 18 * * 0  cd ~/work && srekit oncall-report --team platform \
              --out "reports/oncall-$(date -u +%Y-W%V).md"
```

Доставка в Slack вместо файла:

```bash
srekit oncall-report --team platform --stdout |
  curl -F file=@- -F channels=oncall-summaries \
    -F token=$SLACK_TOKEN https://slack.com/api/files.upload
```

---

## Поймать дрейф шаблонов после апгрейда `srekit`

После `brew upgrade srekit`:

```bash
srekit templates list --json |
  jq '[.[] | select(.status == "embedded-only" or .status == "customized") | .name]'
# ["task.md.tmpl", "runbook.md.tmpl"]   # вот это посмотреть
```

Или просто diff:

```bash
srekit templates diff --name-only
# differs  runbook.md.tmpl
# differs  capacity.md.tmpl
```

Дальше `srekit templates upgrade` смержит их через 3-way.

---

## Разные identity под разные проекты

`~/.srekit.yaml` содержит личную identity; в `.envrc` рабочего репо (через direnv):

```bash
export SREKIT_AUTHOR="Mikhail Savin"
export SREKIT_EMAIL="m.savin@work.example.com"
```

`cd` в work-репо — srekit автоматом подхватит рабочую identity для license / RFC / on-call документов.

---

## Закрепить версию srekit на проект

Некоторые команды хотят чтобы все инженеры использовали одинаковую версию srekit для воспроизводимости. Закрепить в project-скрипте:

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

## Два репо, один источник шаблонов

Часто для multi-repo организаций: один общий `sre-templates` репо, много consumer'ов.

```bash
# В setup'е каждого репо:
git clone git@github.com:acme/sre-templates ~/.acme/templates
echo "templates_dir: ~/.acme/templates" >> ~/.srekit.yaml

# Подтянуть обновления:
srekit templates pull
```

---

## См. также

- [Кастомные шаблоны](guides/custom-templates.md) — развёрнутый гайд.
- [JSON-вывод](guides/json-output.md) — паттерны для пайплайнов.
