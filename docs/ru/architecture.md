# Архитектура

Карта кода для контрибьюторов и любопытных пользователей. Пин на [v0.10.1](https://github.com/jtprogru/srekit/tree/v0.10.1) — структура стабильна на эту версию.

## Раскладка пакетов

```
srekit/
├── main.go               # точка входа: cobra.Execute()
├── cmd/                  # один .go на команду + общий root
│   ├── root.go           # cobra root + viper init + persistent --templates-dir / --config
│   ├── <command>.go      # один файл на генератор + группы templates / config
│   └── cmd_test.go       # smoke-тесты через cobra (SetArgs + capture stdout)
├── internal/
│   ├── ids/              # UUID v4 + slug
│   ├── clock/            # var Now = time.Now (overridable в тестах)
│   ├── meta/             # Author.Resolve + DetectRepo (парсинг git remote)
│   ├── cliflags/         # общий бандл --out / --stdout / --force / --dry-run / --template / --json
│   ├── tmpl/             # //go:embed templates/*.tmpl + Funcs + Source/Loader + Samples + Validate + DocsMD
│   │   └── templates/    # embedded SRE-шаблоны
│   └── render/           # Render() — buildBody/writeBody + JSON short-circuit
├── docs/                 # этот сайт (MkDocs Material с i18n)
├── .github/
│   ├── workflows/        # tests / lint / goreleaser / docs
│   └── dependabot.yml
├── .goreleaser.yaml
├── .golangci.yaml
└── Taskfile.yml
```

## Ключевые абстракции

### `internal/tmpl.Source` и `Loader`

Шаблоны могут приходить из binary (embedded) или из директории пользователя. `Source` — интерфейс:

```go
type Source interface {
    Read(name string) ([]byte, error)
}
```

`EmbedSource` читает из `//go:embed templates/*.tmpl`. `DirSource{Dir}` читает с диска. `Loader{Sources}` идёт по ним по порядку с `fs.ErrNotExist`-as-fallthrough семантикой — отсутствующий файл в user dir прозрачно фолбэчится на embedded.

Production-код использует package-level `tmpl.Default` (`EmbedSource` по default; заменяется на `Loader{DirSource, EmbedSource}` когда сконфигурирована templates dir). Это package-level mutable state, с которым тесты работают через хелпер `resetTmplDefault(t)`.

### `internal/render.Render()`

Общий рендеринг-пайплайн. Берёт имя шаблона, data-структуру и `render.Options{Out, Stdout, Force, DryRun, TemplatePath, JSON, Default}`. Зовёт `buildBody()` (при `Options.JSON` пропускает шаблон и сразу пишет JSON), потом `writeBody()` (решает stdout vs file по флагам и `Default` имени файла).

### `internal/cliflags.Output`

Каждый генератор-cmd содержит `Output` и зовёт `.Bind(cmd, "default-path-description")`. Это и даёт им единый набор флагов без per-command boilerplate. `RenderOptions(def)` превращает значения флагов в `render.Options`.

### `internal/meta.Resolve` и `DetectRepo`

`Resolve` идёт flag → viper → git config для `author` и `email`. `DetectRepo` regex-парсит `git config remote.origin.url` против GitHub SSH и HTTPS паттернов.

### `internal/clock.Now`

`var Now func() time.Time = time.Now` indirection. Позволяет тестам пиннить wall clock (например regression-тест Sunday on-call boundary).

### Snapshot'ы шаблонов: `.srekit-embedded/`

3-way merge в `templates upgrade` использует per-template снапшот "embedded content на момент последнего sync" как merge base. Сидкар лежит в `<user-templates-dir>/.srekit-embedded/<name>` и дописывается в `.gitignore` user dir, чтобы не загрязнять templates-репо. См. `cmd/templates.go` — `snapshotPath`, `readSnapshot`, `writeSnapshot`, `ensureSnapshotIgnored`, `threeWayMerge`.

## Релизный конвейер

| Тул | Назначение |
|---|---|
| `goreleaser` | Билдит linux/darwin/freebsd × arm64/x86_64 бинари; подписывает checksums GPG; обновляет homebrew tap. |
| GitHub Actions `goreleaser.yaml` | Триггерится на tag push (`v*`), импортит GPG-ключ, запускает goreleaser. |
| `crazy-max/ghaction-import-gpg@v7` | Импортит signing key. |
| `HOMEBREW_TAP_GITHUB_TOKEN` secret | Fine-grained PAT с `Contents:read+write` на `jtprogru/homebrew-tap`. |

Flow релиза:

1. Bump `CHANGELOG.md` — переместить `[Unreleased]` контент в `[X.Y.Z]`.
2. Commit `release: X.Y.Z` в `main`.
3. `git tag -a vX.Y.Z -m vX.Y.Z` и `git push origin vX.Y.Z`.
4. goreleaser билдит 8 артефактов + checksums + GPG sig.
5. Cask в `jtprogru/homebrew-tap/Casks/srekit.rb` переписывается.

## Стратегия тестирования

| Слой | Что тестируется | Файл |
|---|---|---|
| Unit | `ids.UUID`/`Slug`, `meta.Resolve`/`DetectRepo`, `tmpl.Funcs`/`Loader`/`Samples`/`Validate`, `cliflags`, `render` (file/stdout/dry-run/JSON) | `internal/*/*_test.go` |
| Integration | Smoke через `cobra.Command.SetArgs` + captured stdout для каждой команды, включая templates pull/validate/diff/upgrade/list и config init | `cmd/cmd_test.go` |
| Race | `go test -race ./...` на CI | `.github/workflows/tests.yaml` |
| Lint | `golangci-lint v2.12` с ~50 линтерами | `.golangci.yaml`, `.github/workflows/lint.yaml` |

## Что трогать по фиче

| Если хочешь... | Старт здесь |
|---|---|
| Добавить новый SRE-артефакт | `cmd/<name>.go` (скопируй существующий генератор) + `internal/tmpl/templates/<name>.md.tmpl` + sample data в `internal/tmpl/samples.go` |
| Добавить флаг существующему генератору | Соответствующий `cmd/<name>.go`; для общих output-флагов — `internal/cliflags/cliflags.go` |
| Изменить контент шаблона | Правь `internal/tmpl/templates/<name>.md.tmpl` |
| Добавить helper-функцию шаблонов | `internal/tmpl/funcs.go` |
| Модифицировать жизненный цикл шаблонов | `cmd/templates.go` |

## См. также

- [Контрибьютинг](contributing.md) — local dev, Taskfile, release process.
- [GitHub source](https://github.com/jtprogru/srekit) — читай код напрямую, он небольшой.
