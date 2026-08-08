# Архитектура

Карта кода для контрибьюторов и любопытных пользователей.

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
│   ├── cliflags/         # общий бандл --out / --stdout / --force / --dry-run / --json
│   ├── tmpl/             # //go:embed templates/*.yaml + Funcs + Source/Loader + Samples + Validate + DocsMD
│   │   └── templates/    # v1 single-file YAML артефакты, по одному на генератор
│   ├── sections/         # Artifact (v1 single-file) + Section/Manifest + Merge + RenderArtifact + JSONSchema
│   ├── migrate/          # `srekit templates migrate` — heuristic .tmpl → .yaml конвертер
│   └── render/           # Render() — buildBody/writeBody + JSON short-circuit + artifact branch
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

`EmbedSource` читает из `//go:embed templates/*.yaml`. `DirSource{Dir}` читает с диска. `Loader{Sources}` идёт по ним по порядку с `fs.ErrNotExist`-as-fallthrough семантикой — отсутствующий файл в user dir прозрачно фолбэчится на embedded.

Каждая команда строит `*tmpl.Loader` через `configureTemplates` и кладёт его в `cmd.Context()`; downstream код достаёт через `loaderFrom(cmd)`. Никакого package-level mutable state — loader scoped per command tree, и `--templates-dir` тесты остаются parallel-safe.

### `internal/sections`

Runtime v1 артефакта. `Artifact` — это распаршенный `<name>.yaml`: frontmatter (`yaml.Node` для сохранения порядка), title, meta_bullets, header_body, список секций, footer_body. `ParseArtifact` валидирует структурные инварианты; `Merge` накладывает per-section override'ы и template-evaluate'ит section titles; `RenderArtifact` собирает markdown (frontmatter блок → H1 → meta_bullets → header_body → `## section` блоки → footer_body), открывая каждый блок через единый хелпер, который гарантирует ровно одну пустую строку между соседними блоками.

Генераторы реализуют `ArtifactPayload()` на data-структуре, чтобы отдать merged section list + ctx обратно в `RenderArtifact`.

### `internal/render.Render()`

Общий рендеринг-пайплайн. Берёт имя, data-структуру и `render.Options{Out, Stdout, Force, DryRun, Default, JSON, Quiet}`. Две ветки:

1. `--json` short-circuit: `MarshalIndent` data, которую вызывающий уже сформировал как `{meta, sections}`.
2. Иначе artifact path: имя — это bare-имя артефакта (`"slo"`), значит загрузить `slo.yaml`, распарсить, отдать в `sections.RenderArtifact`. Имя нормализуется через `tmpl.ArtifactNameFor`: функция идемпотентна и по-прежнему принимает pre-v1.0 варианты написания (`slo.md.tmpl`, `slo.tmpl`).

В v0.30.0 удалена третья ветка — Go-template execution — вместе с `--template FILE` и `license`, её последним потребителем. Туда же ушёл конверт `BootstrapJSON`, заворачивавший отрендеренный markdown в синтетический `{meta, sections}`: ни один генератор не выставлял его с v0.20.0.

### `internal/cliflags.Output`

Каждый генератор-cmd содержит `Output` и зовёт `.Bind(cmd, "default-path-description")`. Это шипит общие флаги (`--out` / `--stdout` / `--force` / `--dry-run` / `--json`). Биндинга `--template FILE` нет: каждый генератор резолвит свой артефакт по имени, так что флаг молча игнорировался бы. `RenderOptions(def)` превращает значения флагов в `render.Options`.

### `internal/meta.Resolve` и `DetectRepo`

`Resolve` идёт flag → viper → git config для `author` и `email`. `DetectRepo` regex-парсит `git config remote.origin.url` против GitHub SSH и HTTPS паттернов.

### `internal/clock.Now`

`var Now func() time.Time = time.Now` indirection. Позволяет тестам пиннить wall clock (например regression-тест Sunday on-call boundary).

### Snapshot'ы шаблонов: `.srekit-embedded/`

3-way merge в `templates upgrade` использует per-template снапшот "embedded content на момент последнего sync" как merge base. Снапшоты лежат в `<user-templates-dir>/.srekit-embedded/<name>` — отдельная служебная директория, которая дописывается в `.gitignore` user dir, чтобы не загрязнять templates-репо. См. `cmd/templates.go` — `snapshotPath`, `readSnapshot`, `writeSnapshot`, `ensureSnapshotIgnored`, `threeWayMerge`.

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
| Unit | `ids.UUID`/`Slug`, `meta.Resolve`/`DetectRepo`, `tmpl.Funcs`/`Loader`, `sections.*`, `cliflags`, `render` (file/stdout/dry-run/JSON/artifact) | `internal/*/*_test.go` |
| Integration | Smoke через `cobra.Command.SetArgs` + captured stdout для каждой команды, включая templates pull/validate/diff/upgrade/list/migrate и config init | `cmd/cmd_test.go` |
| Race | `go test -race ./...` на CI | `.github/workflows/tests.yaml` |
| Lint | `golangci-lint v2.12` с ~50 линтерами | `.golangci.yaml`, `.github/workflows/lint.yaml` |

Render/tmpl unit-тесты строят loader через `newFixtureLoader(t)` helper, который пишет per-test `.tmpl` fixture во временную dir — они не зависят от того, что лежит в embed, что и удержало suite стабильным через v0.14–v0.20 миграционный churn.

## Что трогать по фиче

| Если хочешь... | Старт здесь |
|---|---|
| Добавить новый SRE-артефакт | `cmd/<name>.go` (скопируй существующий генератор) + `internal/tmpl/templates/<name>.yaml` (v1 артефакт) |
| Добавить флаг существующему генератору | Соответствующий `cmd/<name>.go`; для общих output-флагов — `internal/cliflags/cliflags.go` |
| Изменить контент шаблона | Правь `internal/tmpl/templates/<name>.yaml` |
| Подкрутить layout rendered markdown (frontmatter, H1, section composition) | `internal/sections/render_artifact.go` |
| Добавить helper-функцию шаблонов | `internal/tmpl/funcs.go` |
| Модифицировать жизненный цикл шаблонов | `cmd/templates.go` |

## См. также

- [Контрибьютинг](contributing.md) — local dev, Taskfile, release process.
- [GitHub source](https://github.com/jtprogru/srekit) — читай код напрямую, он небольшой.
