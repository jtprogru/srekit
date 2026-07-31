# Конфигурация

srekit читает конфигурацию из четырёх источников. Большинству юзеров достаточно одного из них. Эта страница — authoritative source по "кто кого выигрывает".

## Источники

| Источник | Через |
|---|---|
| **Флаги** | `--author`, `--email`, `--templates-dir` и т.п. в командной строке |
| **Env** | `SREKIT_AUTHOR`, `SREKIT_EMAIL`, `SREKIT_TEMPLATES_DIR` |
| **`~/.srekit.yaml`** | Правится вручную или через `srekit config init` |
| **`git config`** | `user.name`, `user.email` из global или local git config |

## Приоритет

Для каждого ключа srekit идёт по источникам в этом порядке и берёт первое непустое значение:

1. Command-line флаг
2. `SREKIT_<KEY>` env-переменная (например `SREKIT_AUTHOR`)
3. `~/.srekit.yaml` (например `author:`)
4. `git config <git-key>` (только для author/email)

Если все четыре пустые для required-значения — команда падает с понятной ошибкой:

```bash
srekit rfc --title "Move to gRPC"
# Error: author is not set: pass --author, set SREKIT_AUTHOR, or configure git user.name
```

## Ключи

### Identity автора

Используется: `rfc`, `oncall-report` (остальные фолбэчатся на "anonymous" где уместно).

| Ключ | yaml | env | git |
|---|---|---|---|
| name | `author:` (или `full_name:`) | `SREKIT_AUTHOR` | `user.name` |
| email | `email:` | `SREKIT_EMAIL` | `user.email` |

### Templates directory

Используется каждой `templates *` подкомандой и каждым генератором (через overlay-loader).

| Ключ | yaml | env | флаг |
|---|---|---|---|
| templates_dir | `templates_dir:` | `SREKIT_TEMPLATES_DIR` | `--templates-dir` |

Флаг — **persistent flag** на root, применяется к каждой подкоманде.

### Расположение конфига

| Ключ | флаг | default |
|---|---|---|
| config file | `--config FILE` | `~/.srekit.yaml` |

`srekit config init` уважает `--config` тоже — передай чтобы записать файл в другое место.

## yaml-файл

```yaml
# ~/.srekit.yaml
author: Mikhail Savin
email: jtprogru@gmail.com
# templates_dir: ~/.srekit/templates   # опционально
```

Сгенерировать через [`srekit config init`](../commands/config.md). Файл пишется `0o600` (user-only), пути с tilde-стиль раскрытием (`~/foo` → `$HOME/foo`).

## Пример: per-environment override

Одна машина, две GitHub identity (личная и рабочая):

```bash
# ~/.srekit.yaml содержит личную identity.
# На работе:
SREKIT_AUTHOR="Mikhail Savin" SREKIT_EMAIL="m.savin@work.example.com" \
  srekit rfc --title "Move checkout to gRPC"
```

Или scope кастомных шаблонов на конкретный проект:

```bash
srekit --templates-dir ./project-templates rfc --title "Migrate to gRPC"
```

## Дебаг приоритета

Хочешь знать, какой источник выиграл? Запусти с теми же args и сравни вывод, ИЛИ просто проверь env / yaml / git config по очереди. Built-in команды "show resolved config" пока нет — кандидат на v1.0.

## См. также

- [`srekit config init`](../commands/config.md#config-init) — записать yaml-файл интерактивно.
- [Кастомные шаблоны](custom-templates.md) — end-to-end `templates_dir`.
