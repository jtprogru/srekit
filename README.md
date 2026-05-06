# srekit

Генератор текстовых артефактов SRE: задачи, лицензии, постмортемы, RFC, runbook'и, changelog'и.

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

### `srekit task` — заметка для разбора SRE-задачи

```bash
srekit task --title "Tail latency на api-gw" --path ./tasks
# → ./tasks/Tasker - Tail latency на api-gw.md
```

Полная замена `gch sretask` (алиас `srekit sretask` оставлен для совместимости).

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
srekit changelog --out CHANGELOG.md
```

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
