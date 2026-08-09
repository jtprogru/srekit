# Контрибьютинг

Спасибо за интерес к contribution. srekit — маленький, opinionated и стремится таким и остаться — пожалуйста открой issue перед PR, который добавляет новую команду, новую third-party зависимость или меняет имя флага.

## Local setup

```bash
git clone git@github.com:jtprogru/srekit.git
cd srekit
make ci      # запускает lint + race tests
```

Требуется:

- Go **1.26.4** (должен совпадать с CI; см. [version-skew урок](#version-skew) ниже)
- GNU Make **3.81+** — системный `make` на macOS и в любом Linux-дистрибутиве подходит, ставить нечего
- git, bash и обычные POSIX-утилиты

Всё остальное Makefile ставит сам при первом использовании, версии запинены в нём же:

- `golangci-lint` и `govulncheck` → `./bin` (через `go install`, `go.mod` при этом не трогается)
- MkDocs и плагины → `./.venv` (нужен Python 3 с `venv`)

Опционально: `goreleaser` для `make release-dry`, `tokei` для pre-commit хука с LoC-бейджем.

## Make targets

`make` без аргументов печатает этот список. CI зовёт ровно эти же цели, поэтому зелёный `make ci` локально и зелёный пайплайн означают одно и то же.

| Цель | Что делает |
|---|---|
| `make run ARGS="<args>"` | `go run . <args>` |
| `make build` | Билдит `./dist/srekit` |
| `make test` | `go test --short -coverprofile=cover.out -v ./...` |
| `make test-race` | `go test -race -coverprofile=cover.out -v ./...` (что запускает CI) |
| `make lint` | `golangci-lint run` на запиненной версии |
| `make lint-fix` | То же с `--fix` |
| `make govulncheck` | Скан уязвимостей (что запускает CI) |
| `make ci` | `lint` + `test-race` — one-shot pre-push check |
| `make ci-full` | `lint` + `test-race` + `govulncheck` + `docs-build` |
| `make release-dry` | `goreleaser release --clean --snapshot --skip=publish,sign` — билдит в `./dist` без публикации |
| `make docs-install` | Создаёт `./.venv` и ставит `docs/requirements.txt` |
| `make docs-serve` | Сервит docs-сайт на `http://127.0.0.1:8000` |
| `make docs-build` | Билдит docs в `./site` (strict mode) |
| `make tools` | Ставит запиненные `golangci-lint` и `govulncheck` в `./bin` |
| `make fmt` | `gofmt -s -w .` |
| `make vet` | `go vet ./...` |
| `make tidy` | `go mod tidy` |
| `make clean` | Убирает `dist/`, `site/`, `bin/`, `cover.out` |

Makefile написан под GNU Make **3.81** — это системный `make` на macOS, и новее Apple не поставит. При правке не используй `.ONESHELL` и `.SHELLFLAGS` (3.82+), `!=` и `$(file ...)` (4.0+), `$(intcmp)` и `$(let)` (4.4+), и проверяй на 3.81 до пуша: раннер на 4.x проглотит несовместимость молча.

## Стиль кода

- **Никаких `Co-Authored-By`-маркеров** в коммитах — пиши коммиты от своего лица.
- Conventional-commit префиксы (`feat:`, `fix:`, `chore:`, `docs:`, `refactor:`, `test:`). Для удобства changelog scan'а.
- Добавляй smoke-тест на каждую новую генератор-команду и unit-тест на каждую новую internal-функцию. См. `cmd/cmd_test.go` для smoke- паттерна и `internal/*/` для unit-тестов.
- Не глушить lint-находки без обоснования — используй формат `//nolint:<linter> // <code>: <reason>`, как мы делаем для `gosec` G306.
- Держи публичные CLI-флаги минимальными. Новый флаг — PR-worthy change.

## Pre-push чек-лист

```bash
make ci      # lint clean, race tests pass
make build   # сбилдилось
```

Если трогаешь `cmd/templates.go`, прогони и полный templates suite руками — у 3-way merge тонкие инварианты:

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

`make release-dry` локально билдит без публикации — полезно отловить `.goreleaser.yaml`-баги до тегирования.

## Известные нюансы

### Version skew

`golangci-lint` запинен в Makefile (`GOLANGCI_LINT_VERSION`) и ставится в `./bin`, поэтому `make lint` резолвится в один и тот же бинарь локально и в CI. Запуск системного `golangci-lint` напрямую обходит пин — на старой версии gosec правила `G703` (path traversal taint) и `G705` (XSS taint) дают false positive на коде config init. Используй `make lint`.

Go-тулчейн Makefile не пиннит — сверяйся с `go.mod` и `GO_VERSION` в workflow'ах.

### Глобальное состояние в тестах

Загрузчик шаблонов собирается на каждое дерево команд в `configureTemplates` и пробрасывается в команды через cobra command context (`loaderFrom`), поэтому тесты на `--templates-dir` спокойно живут с `t.Parallel()`. Не возвращай package-level загрузчик: `gochecknoglobals` включён, а прежний глобал (`tmpl.Default`) убрали именно потому, что он трипал race-детектор.

Единственное оставшееся глобальное состояние — конфиг-синглтон за `config.Global()`. Тесты, которые его засеивают, используют хелпер `withConfig(t, kv)` из `cmd/cmd_test.go`: он сбрасывает конфиг и регистрирует cleanup. Такие тесты не должны вызывать `t.Parallel()`.

## См. также

- [Архитектура](architecture.md) — карта кода и ключевые абстракции.
- [GitHub issues](https://github.com/jtprogru/srekit/issues) — открой одну перед нетривиальной работой.
