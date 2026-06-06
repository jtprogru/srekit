# srekit

[![Release](https://img.shields.io/github/v/release/jtprogru/srekit?sort=semver)](https://github.com/jtprogru/srekit/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/jtprogru/srekit.svg)](https://pkg.go.dev/github.com/jtprogru/srekit)
[![Go Report Card](https://goreportcard.com/badge/github.com/jtprogru/srekit)](https://goreportcard.com/report/github.com/jtprogru/srekit)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![tests](https://github.com/jtprogru/srekit/actions/workflows/tests.yaml/badge.svg)](https://github.com/jtprogru/srekit/actions/workflows/tests.yaml)
[![golangci-lint](https://github.com/jtprogru/srekit/actions/workflows/lint.yaml/badge.svg)](https://github.com/jtprogru/srekit/actions/workflows/lint.yaml)
[![security](https://github.com/jtprogru/srekit/actions/workflows/security.yaml/badge.svg)](https://github.com/jtprogru/srekit/actions/workflows/security.yaml)
[![goreleaser](https://github.com/jtprogru/srekit/actions/workflows/goreleaser.yaml/badge.svg)](https://github.com/jtprogru/srekit/actions/workflows/goreleaser.yaml)
[![Homebrew](https://img.shields.io/badge/Homebrew-jtprogru%2Ftap-FBB040?logo=homebrew&logoColor=white)](https://github.com/jtprogru/homebrew-tap)
![Go LoC](https://img.shields.io/badge/go-7184%20LoC-blueviolet?logo=go)

📚 **Documentation:** [jtprogru.github.io/srekit](https://jtprogru.github.io/srekit/) (EN + RU, full command reference, guides, recipes, architecture).

Генератор текстовых артефактов SRE: investigation log'и, постмортемы, runbook'и, RFC, on-call report'ы, SLO, error budget policies, capacity plans, retro, changelog'и, лицензии.

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

## Uninstall

Homebrew:

```bash
brew uninstall srekit
brew untap jtprogru/tap   # опционально
```

Установка через `go install` или вручную скачанный бинарник:

```bash
rm "$(command -v srekit)"   # или удалите бинарь из своего PATH / $GOBIN
```

Конфиг и пользовательские шаблоны (если создавались) удаляются отдельно:

```bash
rm -f  "${XDG_CONFIG_HOME:-$HOME/.config}/srekit/config.yaml"
rm -rf "${XDG_CONFIG_HOME:-$HOME/.config}/srekit/templates"
# legacy-расположение (до перехода на XDG):
rm -f  ~/.srekit.yaml
rm -rf ~/.srekit
```

## Usage

```bash
srekit --help
```

Все команды поддерживают единый набор флагов:
`--out FILE` (записать в файл), `--stdout` (печать в stdout), `--force` (перезаписать), `--dry-run` (показать без записи), `--json` (отдать данные шаблона как JSON вместо рендеринга).

Глобальные флаги (на любой команде): `--config FILE`, `--templates-dir DIR`, `-q/--quiet` (подавить INFO-сообщения; рендер и ошибки печатаются как обычно), `-V/--version`.

С `--json` команда не рендерит шаблон, а пишет в stdout (или `--out FILE`) структуру, которую увидел бы шаблон. Удобно для пайплайнов:

```bash
srekit task --title "Tail latency" --json | jq '.id'
srekit postmortem --title X --severity SEV-1 --json | jq '.severity'
```

Ключи — camelCase во всех командах (включая `templates list --json`); это публичный контракт. При `--json` markdown-дефолт пути игнорируется, чтобы JSON случайно не оказался в `.md`-файле.

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

`postmortem` — канонический референс v1-схемы и единственная команда с round-trip workflow «вытащить → отредактировать → отрендерить»:

```bash
srekit postmortem -T "Cache stampede" --json > pm.json   # выгрузить sections в JSON
# ...редактируешь pm.json...
srekit postmortem -T X --from pm.json                    # обратно в markdown
srekit postmortem --schema > postmortem.schema.json      # JSON Schema для тулинга/агентов
srekit postmortem --validate pm.json                     # required sections непустые, нет unknown ID
```

`--from -` читает из stdin. `--schema` и `--validate` взаимоисключающие.

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
srekit oncall-report --team platform                    # период по умолчанию — текущая неделя (Mon–Sun)
srekit oncall-report --team platform --start 2026-05-04 --end 2026-05-10
srekit oncall-report --team platform --author "Alice" --email alice@example.com
```

Если `--author/--email` не заданы, как и в `license` / `rfc`, берётся `SREKIT_*` env → config → `git config`.

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

Раскладывает все встроенные шаблоны в директорию, пишет `TEMPLATES.md` со справочником плейсхолдеров и FuncMap, и делает `git init`. Дальше:

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

Если файла нет в твоей директории, `srekit` тихо берёт встроенный — можно оверрайдить только то, что нужно.

`--template FILE` (one-shot подмена шаблона на одну команду) поддерживается только `srekit license` — у остальных генераторов кастомизация делается через `<name>.yaml` в `templates_dir` (см. ниже про `templates init` / `upgrade`).

### `srekit templates pull` — синхронизация с remote

```bash
srekit templates pull              # git pull --ff-only
srekit templates pull --rebase     # если есть локальные коммиты
```

Запускает `git pull` в configured templates_dir. По умолчанию `--ff-only`, чтобы не получить сюрприз-merge'и. Команда вызывается явно — авто-pull при каждом запуске `srekit` намеренно не делается (это ломает UX и работу в офлайне).

### `srekit templates validate` — проверить, что твои шаблоны рендерятся

```bash
srekit templates validate                    # configured templates_dir
srekit templates validate ./team-templates   # явная директория
```

Per-формат проверки:

- `<name>.yaml` (v1 артефакт) — `sections.ParseArtifact`: поддерживаемая версия, непустой sections-список, уникальные ID, известный `type` (`text` / `list` / `table`), обязательные поля.
- `<name>.sections.yaml` (legacy v0.13.x sidecar) — `sections.ParseManifest` те же структурные проверки на старой раскладке.
- `<name>.tmpl` — Go-template parse-only с общим FuncMap. Ловит синтаксис (unclosed `{{`, неизвестные функции); опечатки в полях не ловятся (с v0.20.0 в embed нет ни одного `.tmpl`, sample-data для exec нет).

Не-zero exit если что-то упало.

### `srekit templates diff` — что изменилось относительно embedded-версии

```bash
srekit templates diff                  # полный unified diff каждого изменённого файла
srekit templates diff --name-only      # только имена
srekit templates diff --no-color
```

Сравнивает каждый артефакт в твоей templates dir (`.yaml` / `.tmpl` / `.sections.yaml`) с версией, зашитой в текущий binary. Полезно после `srekit templates pull` или обновления бинарника — увидеть, что у тебя осталось своё, а что отстало от апстрима. Файлы без embedded-counterpart (твои bespoke-шаблоны) маркируются как `user-only`.

### `srekit templates list` — что у тебя есть и в каком состоянии

```bash
srekit templates list                     # таблица (учитывает configured templates_dir)
srekit templates list ./team-templates    # явная директория
srekit templates list --json | jq         # для пайплайнов
srekit templates list --filter customized # только то, что ты переопределил
```

Классификация для каждого артефакта (`.yaml` / `.tmpl` / `.sections.yaml`):

- `identical` — байт-в-байт совпадает с embedded;
- `customized` — есть у тебя и отличается;
- `user-only` — твой bespoke без embedded-counterpart;
- `embedded-only` — зашит в бинарник, у тебя нет override.

JSON-ключи — camelCase (`name`, `status`, `userPath`) — то же соглашение, что и у `--json` генераторов.

### `srekit templates migrate` — конвертация legacy `.tmpl` → v1 `.yaml`

```bash
srekit templates migrate              # dry-run: показать, что получилось бы
srekit templates migrate --apply      # записать <name>.yaml рядом со старыми файлами
```

Однократная миграция для тех, кто остался на pre-v0.14.0 раскладке (`<name>.tmpl` + опциональный `<name>.sections.yaml` sidecar). Создаёт `<name>.yaml` в v1-формате; оригиналы остаются на месте, чтобы их можно было сравнить и удалить руками. По умолчанию `--dry-run`. Подробный upgrade-recipe — в [docs/migration/v1.md](https://jtprogru.github.io/srekit/migration/v1/).

### `srekit templates upgrade` — подтянуть новые embedded-шаблоны

```bash
srekit templates upgrade             # 3-way merge кастомизаций, без --force
srekit templates upgrade --dry-run   # посмотреть что изменится
srekit templates upgrade --force     # перезаписать и кастомизации (без merge)
```

3-way merge: `srekit` хранит снапшот embedded на момент последнего sync'а в `<templates-dir>/.srekit-embedded/` и использует его как merge-base. `git merge-file --diff3` мерджит твои изменения с upstream'ом:

- нет файла → копируется;
- идентичен embedded → пропуск;
- upstream без изменений, твои есть → молчаливо не трогаем;
- upstream изменился, твоих нет → fast-forward (без `--force`);
- расхождение с обеих сторон → 3-way merge. Чистый merge — пишется silently; конфликт — маркеры `<<<<<<<` / `>>>>>>>` в файл, exit non-zero, разрешаешь руками.

Без снапшота (старый user dir, до этой версии) — fallback на additive поведение (skip + seed снапшота для следующего apgrade). Сидкар `.srekit-embedded/` автоматически попадает в `.gitignore` твоей dir. `TEMPLATES.md` обновляется всегда (это reference, не точка кастомизации).

### `srekit completion` — shell autocomplete

```bash
srekit completion zsh > "${fpath[1]}/_srekit"
srekit completion bash > /etc/bash_completion.d/srekit
```

## Config

Путь по умолчанию — `$XDG_CONFIG_HOME/srekit/config.yaml` (т.е. `~/.config/srekit/config.yaml`). Legacy `~/.srekit.yaml` продолжает читаться, если уже существует — миграция не нужна, но новые конфиги пишутся в XDG.

```yaml
author: Mikhail Savin
email: jtprogru@example.com
# templates_dir: ~/.srekit/templates
```

Или через env: `SREKIT_AUTHOR`, `SREKIT_EMAIL`, `SREKIT_TEMPLATES_DIR`. Альтернативный путь к файлу — `srekit --config ./my.yaml …`.

Быстро создать файл интерактивно:

```bash
srekit config init                    # TTY → запросит author / email / templates_dir с дефолтами из git config
srekit config init --yes              # non-interactive: значения из --author / --email / git config
srekit config init --force            # перезаписать существующий файл
srekit --config ./my.yaml config init # альтернативный путь
```

## Development

В репозитории лежит pre-commit hook (`.githooks/pre-commit`), который обновляет Go LoC бейдж в `README.md` через [`tokei`](https://github.com/XAMPPRocky/tokei). Хук не подключается автоматически — это делается явно:

```bash
git config core.hooksPath .githooks
```

Зависимость `tokei` ставится отдельно:

```bash
brew install tokei            # macOS / Linux
cargo install tokei           # через cargo
```

Если `tokei` не найден в `PATH`, хук молча скипается — коммит не блокируется.

## Стабильность и версионирование

`srekit` следует [SemVer](https://semver.org/lang/ru/). Текущая ветка — `0.x`, поэтому breaking changes допустимы между minor-версиями (и явно помечены в [CHANGELOG](CHANGELOG.md) как `Breaking — …`).

С **v1.0** релиз станет stability stamp:

- **Стабильный публичный контракт.** CLI-флаги, имена и порядок section ID в `--json`, схема `<name>.yaml` (`version` / `frontmatter` / `title` / `meta_bullets` / `header_body` / `sections`), словарь section `type` (`text` / `list` / `table`), ключи в `~/.srekit.yaml` и `SREKIT_*` env. Любое из этого ломается только в major-релизе с migration-инструкцией.
- **Соблюдение обратной совместимости через 1.x.** Поддерживается чтение legacy `.tmpl` и `.sections.yaml` файлов в user-`templates_dir` (с stderr `WARN`); их удаление — кандидат на 2.0.
- **Deprecation-цикл.** Когда мы что-то планируем убрать, оно сначала становится no-op или начинает писать `WARN` минимум один minor-релиз, потом удаляется в следующем major. Пример из недавнего: `--template FILE` на не-license командах с v0.20.0 был silent no-op, в v0.22.0 убран с CLI-surface.

Что **не** стабилизируется в 1.0 (может меняться в 1.x):

- Содержимое поля `frontmatter:` — пока free-form map. Возможен переход на типизированную JSON Schema; breaking только для авторов, которые завязались на free-form структуру.
- Точные формулировки stderr `WARN` для legacy-файлов.
- Внутренние Go-API в `internal/*` — это `internal/` намеренно, не используйте напрямую.

Полный upgrade-guide и список JSON shape changes по версиям — в [docs/migration/v1.md](https://jtprogru.github.io/srekit/migration/v1/).

## License

MIT — см. `LICENSE`.
