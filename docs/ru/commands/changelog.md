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

`changelog` шипится как v1 YAML-артефакт (`internal/tmpl/templates/changelog.yaml`) — H1 + `header_body` (intro-параграф) + две секции (`unreleased` и `initial_release`). Заголовок секции `initial_release` динамический (`[{{ .Meta.InitialVersion }}] - {{ .Meta.Today }}`); section titles template-evaluated с v0.20.0. Template-выражения обращаются к `.Meta.<Field>` для `Today` (дата `2006-01-02`), `Repo` (`<owner>/<name>`), `InitialVersion`. См. [`srekit postmortem`](postmortem.md#customizing-the-artifact-v1-format-v0140) для полной схемы.

## См. также

- [`srekit rfc`](rfc.md), [`srekit postmortem`](postmortem.md) — документы, история которых питает changelog.
