# Контрибьютинг

Спасибо за интерес к contribution. srekit — маленький, со своей позицией, и стремится таким и остаться — пожалуйста открой issue перед PR, который добавляет новую команду, новую third-party зависимость или меняет имя флага.

## Локальная установка

```bash
git clone git@github.com:jtprogru/srekit.git
cd srekit
make ci      # запускает lint + race tests
```

Требуется:

- Go **1.26.4** (должен совпадать с CI; см. [урок про расхождение версий](#version-skew) ниже)
- GNU Make **3.81+** — системный `make` на macOS и в любом Linux-дистрибутиве подходит, ставить нечего
- git, bash и обычные POSIX-утилиты

Всё остальное Makefile ставит сам при первом использовании, версии запинены в нём же:

- `golangci-lint` и `govulncheck` → `./bin` (через `go install`, `go.mod` при этом не трогается)
- MkDocs и плагины → `./.venv` (нужен Python 3 с `venv`)

Опционально: `goreleaser` для `make release-dry`, `tokei` для pre-commit хука с LoC-бейджем.

## Цели Make

`make` без аргументов печатает этот список. CI зовёт ровно эти же цели, поэтому зелёный `make ci` локально и зелёный пайплайн означают одно и то же.

| Цель | Что делает |
|---|---|
| `make run ARGS="<args>"` | `go run . <args>` |
| `make build` | Билдит `./dist/srekit` |
| `make install` | `go install` |
| `make test` | `go test --short -coverprofile=cover.out -v ./...` |
| `make test-race` | `go test -race -coverprofile=cover.out -v ./...` (что запускает CI) |
| `make lint` | `golangci-lint run` на запиненной версии |
| `make lint-fix` | То же с `--fix` |
| `make govulncheck` | Скан уязвимостей (что запускает CI) |
| `make ci` | `lint` + `test-race` — one-shot pre-push check |
| `make ci-full` | `lint` + `test-race` + `govulncheck` + `docs-build` |
| `make release-check` | `goreleaser check` — валидирует `.goreleaser.yaml` |
| `make release-dry` | `goreleaser release --clean --snapshot --skip=publish,sign` — билдит в `./dist` без публикации |
| `make docs-install` | Создаёт `./.venv` и ставит `docs/requirements.txt` |
| `make docs-serve` | Сервит docs-сайт на `http://127.0.0.1:8000` |
| `make docs-build` | Билдит docs в `./site` (strict mode) |
| `make docs-deploy` | `mkdocs gh-deploy` — зовётся из CI, не руками |
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

## Чек-лист перед пушем

```bash
make ci      # lint clean, race tests pass
make build   # сбилдилось
```

Если трогаешь `cmd/templates.go`, прогони и полный набор тестов templates руками — у 3-way merge тонкие инварианты:

```bash
go test ./cmd/ -run TestTemplates -v
```

## Релизный процесс

(Для мейнтейнеров.)

1. Убедиться что `main` чистый и CI зелёный.
2. Проверить токен от tap'а — это единственный credential в пайплайне, который не является встроенным в workflow `GITHUB_TOKEN`, и падает он *после* публикации релиза:

    ```bash
    GH_TOKEN=<tap-pat> gh api repos/jtprogru/homebrew-tap --jq .default_branch   # ожидается: main
    ```

    `401 Bad credentials` означает ротацию fine-grained PAT (`jtprogru/homebrew-tap`, `Contents: Read and write`) и `gh secret set HOMEBREW_TAP_GITHUB_TOKEN --repo jtprogru/srekit`.

3. Отрезать changelog: `srekit changelog release --version X.Y.Z` либо руками переместить `[Unreleased]` в `[X.Y.Z] - YYYY-MM-DD`.
4. `git commit -m "release: X.Y.Z"` + `git push`.
5. `git tag -a vX.Y.Z -m vX.Y.Z` + `git push origin vX.Y.Z`.
6. Подождать `goreleaser` workflow — ~90 секунд.
7. Проверить GitHub Release страницу и `jtprogru/homebrew-tap` cask.

Если пуш в cask всё же упал, перезапуск job'а не помогает: rerun проигрывает тот же ref, поэтому фикс, приехавший в `main`, не подхватится, а повторная заливка существующих артефактов падает с 422. Почини токен, потом удали релиз и тег и запушь тег заново на том же коммите (`gh release delete vX.Y.Z --yes`, `git push origin :refs/tags/vX.Y.Z`, `git push origin vX.Y.Z`). Пересборка не byte-identical — таймстемпы в архивах двигают checksums, — так что всё, что записало `sha256` первого прогона, протухает.

`make release-dry` локально билдит без публикации — полезно отловить `.goreleaser.yaml`-баги до тегирования.

## Известные нюансы

### Расхождение версий { #version-skew }

`golangci-lint` запинен в Makefile (`GOLANGCI_LINT_VERSION`) и ставится в `./bin`, поэтому `make lint` резолвится в один и тот же бинарь локально и в CI. Запуск системного `golangci-lint` напрямую обходит пин — на старой версии gosec правила `G703` (path traversal taint) и `G705` (XSS taint) дают false positive на коде config init. Используй `make lint`.

Go-тулчейн Makefile не пиннит — сверяйся с `go.mod` и `GO_VERSION` в workflow'ах.

### Глобальное состояние в тестах

Загрузчик шаблонов собирается на каждое дерево команд в `configureTemplates` и пробрасывается в команды через cobra command context (`loaderFrom`), поэтому тесты на `--templates-dir` спокойно живут с `t.Parallel()`. Не возвращай package-level загрузчик: `gochecknoglobals` включён, а прежний глобал (`tmpl.Default`) убрали именно потому, что он трипал race-детектор.

Единственное оставшееся глобальное состояние — конфиг-синглтон за `config.Global()`. Тесты, которые его засеивают, используют хелпер `withConfig(t, kv)` из `cmd/cmd_test.go`: он сбрасывает конфиг и регистрирует cleanup. Такие тесты не должны вызывать `t.Parallel()`.

## См. также

- [Архитектура](architecture.md) — карта кода и ключевые абстракции.
- [GitHub issues](https://github.com/jtprogru/srekit/issues) — открой одну перед нетривиальной работой.
