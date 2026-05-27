# srekit changelog

Скаффолд `CHANGELOG.md` в формате [Keep a Changelog](https://keepachangelog.com/en/1.1.0/). Автодетектит GitHub репо из `git config remote.origin.url`.

## Синопсис

```bash
srekit changelog [flags]
```

## Флаги

| Флаг | Обязательный | Описание |
|---|---|---|
| `--repo` | нет | `<owner>/<name>` slug. Если не передан — srekit читает `git config remote.origin.url` и парсит GitHub SSH или HTTPS URL'ы. |
| `--version` | нет | Начальный version anchor (например `0.1.0`). Default: `0.1.0`. |

Плюс [общие output-флаги](index.md#shared-output-flags). Default имя файла: `CHANGELOG.md`.

## Примеры

Внутри git-репо с `origin`-remote на GitHub:

```bash
srekit changelog --out CHANGELOG.md
# репо детектится из git remote, версия default 0.1.0
```

Явно:

```bash
srekit changelog --repo jtprogru/srekit --version 0.1.0 --out CHANGELOG.md
```

Вне git-репо без `--repo` — ошибка (никакого silent `OWNER/REPO` плейсхолдера, который кусал юзеров в v0.2):

```bash
srekit changelog --stdout
# Error: could not detect repo from git remote — pass --repo OWNER/NAME
```

## Вывод

Скаффолд предрендерит compare-ссылки на `github.com/<repo>/compare/v<version>...HEAD` и включает скелет `[Unreleased]` / `[<version>]` с подсекциями `Added` / `Changed` / `Fixed`.

## Структура данных для шаблона

```go
struct {
    Repo    string  // "<owner>/<name>"
    Version string
    Today   string  // RFC 3339-ish
}
```

## См. также

- [`srekit rfc`](rfc.md), [`srekit postmortem`](postmortem.md) — документы, история которых питает changelog.
