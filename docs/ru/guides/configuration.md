# Конфигурация

srekit читает конфигурацию из четырёх источников. Большинству юзеров достаточно одного из них. Эта страница — authoritative source по "кто кого выигрывает".

## Источники

| Источник | Через |
|---|---|
| **Флаги** | `--author`, `--email`, `--templates-dir` и т.п. в командной строке |
| **Env** | `SREKIT_AUTHOR`, `SREKIT_EMAIL`, `SREKIT_TEMPLATES_DIR` |
| **Конфиг-файл** | `$XDG_CONFIG_HOME/srekit/config.yaml`, правится вручную или через `srekit config init` |
| **`git config`** | `user.name`, `user.email` из global или local git config |

## Приоритет

Для каждого ключа srekit идёт по источникам в этом порядке и берёт первое непустое значение:

1. Command-line флаг
2. `SREKIT_<KEY>` env-переменная (например `SREKIT_AUTHOR`)
3. Конфиг-файл (например `author:`)
4. `git config <git-key>` (только для author/email)

Если все четыре пустые для required-значения — команда падает с понятной ошибкой:

```bash
srekit rfc --title "Move to gRPC"
# Error: author is not set: pass --author, set SREKIT_AUTHOR, or configure git user.name
```

## Ключи

### Личность автора { #identity }

Используется: `rfc`, `oncall-report` (остальные фолбэчатся на "anonymous" где уместно).

| Ключ | yaml | env | git |
|---|---|---|---|
| name | `author:` (или `full_name:`) | `SREKIT_AUTHOR` | `user.name` |
| email | `email:` | `SREKIT_EMAIL` | `user.email` |

### Директория шаблонов { #templates-dir }

Используется каждой `templates *` подкомандой и каждым генератором (через overlay-loader).

| Ключ | yaml | env | флаг |
|---|---|---|---|
| templates_dir | `templates_dir:` | `SREKIT_TEMPLATES_DIR` | `--templates-dir` |

Флаг — **persistent flag** на root, применяется к каждой подкоманде.

### Язык changelog { #changelog-lang }

Используется [`srekit changelog`](../commands/changelog.md#ru-variant) и наследуется его подкомандами `release` и `validate`.

| Ключ | yaml | env | флаг |
|---|---|---|---|
| changelog_lang | `changelog_lang:` | `SREKIT_CHANGELOG_LANG` | `--lang` |

Принимает `en` (default) или `ru`. Нераспознанное значение падает с ошибкой, называющей допустимые, ещё до того как что-либо записано: опечатка здесь не откатывается молча в английский. Настройка управляет тем, что генерируется, и никогда не влияет на то, как парсится уже существующий changelog.

### Расположение конфига

| Ключ | флаг | default |
|---|---|---|
| config file | `--config FILE` | `$XDG_CONFIG_HOME/srekit/config.yaml` |

Для свежих установок srekit следует XDG Base Directory Specification, но pre-XDG путь выигрывает, если уже существует: `~/.srekit.yaml` для конфига и `~/.srekit/templates` для директории шаблонов. Так обновление никогда не оставляет тебя с конфигом, который лежит и никем не читается. Если существуют оба, используется legacy, а [`srekit doctor`](../commands/doctor.md) предупреждает, какой из них игнорируется.

`srekit config init` уважает `--config` тоже — передай чтобы записать файл в другое место.

## yaml-файл

```yaml
# ~/.config/srekit/config.yaml
author: Mikhail Savin
email: jtprogru@gmail.com
# templates_dir: ~/.config/srekit/templates   # опционально
# changelog_lang: ru                   # опционально, default: en
```

Сгенерировать через [`srekit config init`](../commands/config.md). Файл пишется `0o600` (user-only), пути с tilde-стиль раскрытием (`~/foo` → `$HOME/foo`).

## Пример: per-environment override

Одна машина, две GitHub identity (личная и рабочая):

```bash
# Конфиг-файл содержит личную identity.
# На работе:
SREKIT_AUTHOR="Mikhail Savin" SREKIT_EMAIL="m.savin@work.example.com" \
  srekit rfc --title "Move checkout to gRPC"
```

Или scope кастомных шаблонов на конкретный проект:

```bash
srekit --templates-dir ./project-templates rfc --title "Migrate to gRPC"
```

## Дебаг приоритета

Хочешь знать, какой источник выиграл? Спроси:

```bash
srekit doctor
```

`config.file` называет конфиг, который реально читается; `config.env` перечисляет все действующие переменные с префиксом `SREKIT_`; `config.templates-dir` сообщает зарезолвленную директорию шаблонов *и источник, который её дал*; `config.identity` — итоговые author name и email с происхождением каждого значения. `config.shadowed` срабатывает, когда существуют и XDG-, и legacy-путь, так что «правил конфиг, а его никто не читает» выглядит как warning, а не как загадка. См. [`srekit doctor`](../commands/doctor.md).

## См. также

- [`srekit doctor`](../commands/doctor.md) — read-only отчёт обо всём, что эта страница описывает как «резолвится».
- [`srekit config init`](../commands/config.md#config-init) — записать yaml-файл интерактивно.
- [Кастомные шаблоны](custom-templates.md) — end-to-end `templates_dir`.
