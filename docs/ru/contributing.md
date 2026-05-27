# Контрибьютинг

Спасибо за интерес к contribution. srekit — маленький, opinionated и
стремится таким и остаться — пожалуйста открой issue перед PR, который
добавляет новую команду, новую third-party зависимость или меняет имя
флага.

## Local setup

```bash
git clone git@github.com:jtprogru/srekit.git
cd srekit
go mod download
task ci      # запускает lint + race tests
```

Требуется:

- Go **1.25+**
- [Task](https://taskfile.dev/) (опционально, но рекомендуется — см.
  `Taskfile.yml` для полного списка target'ов)
- `golangci-lint` **v2.12** (должен совпадать с CI; см.
  [version-skew урок](#version-skew) ниже)
- Для docs: Python 3 + `pip install -r docs/requirements.txt`

## Taskfile targets

| Task | Что делает |
|---|---|
| `task run -- <args>` | `go run . <args>` |
| `task build` | Билдит `./dist/srekit` |
| `task test` | `go test --short -coverprofile=cover.out -v ./...` |
| `task test:race` | `go test -race -v ./...` (что CI запускает) |
| `task lint` | `golangci-lint -v run` |
| `task ci` | `lint` + `test:race` — one-shot pre-push check |
| `task release:dry` | `goreleaser release --clean --snapshot --skip=publish,sign` — билдит в `./dist` без публикации |
| `task docs:install` | `pip install -r docs/requirements.txt` |
| `task docs:serve` | Сервит docs-сайт на `http://127.0.0.1:8000` |
| `task docs:build` | Билдит docs в `./site` (strict mode) |
| `task fmt` | `gofmt -s -w .` |
| `task vet` | `go vet ./...` |
| `task tidy` | `go mod tidy` |

## Стиль кода

- **Никаких `Co-Authored-By`-маркеров** в коммитах — пиши коммиты от
  своего лица.
- Conventional-commit префиксы (`feat:`, `fix:`, `chore:`, `docs:`,
  `refactor:`, `test:`). Для удобства changelog scan'а.
- Добавляй smoke-тест на каждую новую генератор-команду и unit-тест
  на каждую новую internal-функцию. См. `cmd/cmd_test.go` для smoke-
  паттерна и `internal/*/` для unit-тестов.
- Не глушить lint-находки без обоснования — используй формат
  `//nolint:<linter> // <code>: <reason>`, как мы делаем для `gosec`
  G306.
- Держи публичные CLI-флаги минимальными. Новый флаг — PR-worthy
  change.

## Pre-push чек-лист

```bash
task ci      # lint clean, race tests pass
task build   # сбилдилось
```

Если трогаешь `cmd/templates.go`, прогони и полный templates suite
руками — у 3-way merge тонкие инварианты:

```bash
go test ./cmd/ -run TestTemplates -v
```

## Релизный процесс

(Для мейнтейнеров.)

1. Убедиться что `main` чистый и CI зелёный.
2. Переместить `[Unreleased]` контент в `CHANGELOG.md` в `[X.Y.Z] - YYYY-MM-DD`.
3. `git commit -m "release: X.Y.Z"` + `git push`.
4. `git tag -a vX.Y.Z -m vX.Y.Z` + `git push origin vX.Y.Z`.
5. Подождать `goreleaser` workflow — ~90 секунд.
6. Проверить GitHub Release страницу и `jtprogru/homebrew-tap` cask.

`task release:dry` локально билдит без публикации — полезно отловить
`.goreleaser.yaml`-баги до тегирования.

## Известные нюансы

### Version skew

CI пиннит `golangci-lint v2.12`. Если локально старее — gosec правила
`G703` (path traversal taint) и `G705` (XSS taint) могут давать false
positive на коде config init. Синхронизируй local и CI:

```bash
brew install golangci-lint
golangci-lint version    # должно быть 2.12.x
```

### `tmpl.Default` race

`go test -race ./cmd/` флэйчит локально на macOS/go1.26 вокруг
`tmpl.Default` — несколько тестов через `runCLI` трипают детектор на
глобальном Loader pointer. CI (Linux/go1.25) стабильно проходит;
рабочая теория — single-CPU runners избегают interleavings.
Текущий workaround для мутаторов — `resetTmplDefault(t)` плюс никаких
`t.Parallel()`. Соблюдай паттерн в новых тестах. Правильный фикс
(пробрасывать Loader через cobra context вместо глобала) — в списке
v1.0.

### `tmpl.Default` — package-level mutable state

Тот же root cause что у race выше. Тесты используют
`resetTmplDefault(t)` для snapshot/restore вокруг мутаций. Относись к
`tmpl.Default` как к test-fragile до v1.0 stabilization.

## См. также

- [Архитектура](architecture.md) — карта кода и ключевые абстракции.
- [GitHub issues](https://github.com/jtprogru/srekit/issues) — открой
  одну перед нетривиальной работой.
