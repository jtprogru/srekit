# srekit

Генератор текстовых артефактов SRE: investigation log'и, инциденты, постмортемы, runbook'и, RFC, on-call report'ы, SLO, error budget policies, capacity plans, retro, changelog'и, лицензии.

Все markdown-шаблоны двуязычные: заголовки и метки в формате `Русский (English)`, тело — на русском. Технические идентификаторы (SLO/SLI/RFC/PromQL/UTC/SEV), ключи YAML frontmatter и PromQL-выражения остаются английскими. Шаблон `changelog` остаётся полностью английским, чтобы не ломать тулинг вокруг Keep a Changelog.

Извлечён из [gch](https://github.com/jtprogru/gch) в рамках распиливания монолита.

## Install

Через Homebrew (macOS / Linux):

```bash
brew tap jtprogru/tap
brew install srekit
```

Через `go install`:

```bash
go install github.com/jtprogru/srekit@latest
```

Готовые бинарники под Linux / macOS / FreeBSD (amd64, arm64) — на странице [Releases](https://github.com/jtprogru/srekit/releases).

## Usage

```bash
srekit --help
```

Все команды поддерживают единый набор флагов:
`--out FILE` (записать в файл), `--stdout` (печать в stdout), `--force` (перезаписать), `--dry-run` (показать без записи).

### `srekit task` — investigation log для SRE-расследования

```bash
srekit task --title "Tail latency on api-gw" --path ./tasks
# → ./tasks/investigation-tail-latency-on-api-gw.md
```

Шаблон с секциями Context / Hypothesis / Evidence / Findings / Action items / References. Алиас `srekit sretask` оставлен для совместимости (исторически команда заменяла `gch sretask`).

### `srekit license` — LICENSE-файл (по умолчанию WTFPL)

```bash
srekit license --stdout                          # WTFPL в stdout (как gch lic)
srekit license --out LICENSE                     # записать в LICENSE
srekit license --type mit --out LICENSE          # MIT
srekit license --type apache2 --out LICENSE      # Apache 2.0
```

Author/email берётся в порядке: `--author/--email` → `SREKIT_AUTHOR/SREKIT_EMAIL` → `~/.srekit.yaml` → `git config user.name/user.email`.

### `srekit postmortem` — шаблон постмортема (Google SRE-style)

```bash
srekit postmortem --title "API outage" --severity SEV-1 \
  --start 2026-05-06T08:00Z --end 2026-05-06T09:30Z \
  --owner "@oncall" --out postmortem-2026-05-06.md
```

### `srekit incident` — «живой» инцидент-док

```bash
srekit incident --title "API down" --severity SEV-1 --lead alice --stdout
```

Шаблон для заполнения **во время** инцидента (статус, лид, коммс, лог апдейтов) — в отличие от постмортема, который пишется после. Статусы: `investigated | active | contained | resolved`.

### `srekit rfc` — RFC / ADR

```bash
srekit rfc --title "Migrate to gRPC" --status proposed --stdout
```

Статусы: `proposed | accepted | rejected | superseded | deprecated`.

### `srekit runbook` — runbook для on-call

```bash
srekit runbook --title "p99 latency spike" --service api-gw --alert APIHighLatency
```

### `srekit changelog` — `CHANGELOG.md` в формате Keep a Changelog

```bash
srekit changelog --out CHANGELOG.md                    # репо детектится из git remote
srekit changelog --repo jtprogru/srekit --version 0.2.0
```

### `srekit oncall-report` — недельный отчёт дежурного

```bash
srekit oncall-report --team platform                    # период по умолчанию — текущая неделя
srekit oncall-report --team platform --start 2026-05-04 --end 2026-05-10
```

### `srekit slo` — SLO/SLI документ

```bash
srekit slo --service api-gw --target 99.95% --window 30d --latency 200ms
```

### `srekit ebp` — Error Budget Policy

```bash
srekit ebp --service api-gw --out ebp-api-gw.md
```

Политика, что команда делает при сгорании бюджета ошибок: triggered actions по уровням (Yellow / Orange / Red), исключения, эскалация.

### `srekit capacity` — план ёмкости

```bash
srekit capacity --service api-gw --horizon 1y --out capacity-api-gw.md
```

Шаблон capacity planning: baseline, допущения роста, прогноз, триггеры скейла, headroom, зависимости, стоимость, риски.

### `srekit retro` — шаблон ретро

```bash
srekit retro --team platform --sprint 2026-W19
```

### `srekit templates init` — твои собственные шаблоны под git

```bash
srekit templates init                # → ~/.srekit/templates
srekit templates init ./team-templates --no-git
```

Раскладывает все встроенные шаблоны в директорию, пишет `TEMPLATES.md` со
справочником плейсхолдеров и FuncMap, и делает `git init`. Дальше:

```bash
cd ~/.srekit/templates
git remote add origin git@github.com:your-team/sre-templates.git
git add . && git commit -m "initial templates" && git push -u origin main
```

Подключение директории к `srekit`:

```bash
echo 'templates_dir: ~/.srekit/templates' >> ~/.srekit.yaml
# или: export SREKIT_TEMPLATES_DIR=~/.srekit/templates
# или: srekit --templates-dir ~/.srekit/templates …  (на одну команду)
```

Подмена точечно — только один шаблон под одну команду:

```bash
srekit runbook --title "p99 spike" --template ./oneshot-runbook.tmpl --stdout
```

Если файла нет в твоей директории, `srekit` тихо берёт встроенный — можно
оверрайдить только то, что нужно.

### `srekit templates pull` — синхронизация с remote

```bash
srekit templates pull              # git pull --ff-only
srekit templates pull --rebase     # если есть локальные коммиты
```

Запускает `git pull` в configured templates_dir. По умолчанию `--ff-only`,
чтобы не получить сюрприз-merge'и. Команда вызывается явно — авто-pull
при каждом запуске `srekit` намеренно не делается (это ломает UX и работу
в офлайне).

### `srekit completion` — shell autocomplete

```bash
srekit completion zsh > "${fpath[1]}/_srekit"
srekit completion bash > /etc/bash_completion.d/srekit
```

## Config

`~/.srekit.yaml` (опционально):

```yaml
author: Mikhail Savin
email: jtprogru@gmail.com
```

Или через env: `SREKIT_AUTHOR`, `SREKIT_EMAIL`.

## License

WTFPL — см. `LICENSE`.
