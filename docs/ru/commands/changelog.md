# srekit changelog

Скаффолд `CHANGELOG.md` в формате [Keep a Changelog](https://keepachangelog.com/en/1.1.0/). Автодетектит GitHub репо из `git config remote.origin.url`.

## Синопсис

```bash
srekit changelog [flags]
```

## Флаги

| Флаг | Обязательный | Описание |
|---|---|---|
| `--repo` | нет | `<owner>/<name>` slug. Если не передан — srekit берёт `meta.repo` из payload'а `--from`, а если и его нет — читает `git config remote.origin.url` и парсит GitHub SSH или HTTPS URL'ы. |
| `--version` | нет | Начальный version anchor (например `0.1.0`). Default: `0.1.0`. |
| `--from` | нет | Читать тела секций из JSON-файла; `-` читает stdin. |

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

## Структурированный ввод

`--json` отдаёт документ в виде `{meta, sections}`, а `--from` принимает эту же форму обратно — тела секций можно заполнить скриптом или агентом, а не руками:

```bash
srekit changelog --repo acme/api --json > cl.json
# ...заменить тело секции "unreleased"...
srekit changelog --from cl.json
```

Переданные тела секций вставляются verbatim, без вычисления шаблонов, поэтому markdown с `{{ }}` проходит round-trip без изменений. Пропущенные секции берут дефолты из артефакта. Неизвестный артефакту section id — ошибка с именем нарушителя, а не молчаливый пропуск.

`meta` в payload'е задаёт `repo`, `initialVersion` и `today`. Флаг выигрывает у файла, файл — у git remote, поэтому payload с `meta.repo` рендерится и вне git-репозитория.

В отличие от [`srekit postmortem`](postmortem.md), у `changelog` нет `--schema` и `--validate`: его артефакт не объявляет обязательных секций, поэтому валидация payload'а не может завершиться неудачей, а схема из двух строковых полей говорит меньше, чем сам `--json`.

## Вывод

Скаффолд включает скелет `[Unreleased]` / `[<version>]` с шестью подсекциями Keep a Changelog и заканчивается блоком link reference definitions на `github.com/<repo>/compare/v<version>...HEAD`.

Этот блок ссылок — футер уровня документа, а не часть тела последней секции. Поэтому он переживает payload `--from`, заменяющий `initial_release`, и именно его будет переписывать будущий `changelog release`.

## Структура данных для шаблона

`changelog` шипится как v1 YAML-артефакт (`internal/tmpl/templates/changelog.yaml`) — H1 + `header_body` (intro-параграф) + две секции (`unreleased` и `initial_release`) + `footer_body` (блок link reference definitions). Заголовок секции `initial_release` динамический (`[{{ .Meta.InitialVersion }}] - {{ .Meta.Today }}`); section titles template-evaluated с v0.20.0. Template-выражения обращаются к `.Meta.<Field>` для `Today` (дата `2006-01-02`), `Repo` (`<owner>/<name>`), `InitialVersion`. См. [`srekit postmortem`](postmortem.md#customizing-the-artifact-v1-format-v0140) для полной схемы.

## См. также

- [`srekit rfc`](rfc.md), [`srekit postmortem`](postmortem.md) — документы, история которых питает changelog.
