# srekit license

Сгенерировать `LICENSE`-файл (default: WTFPL). Author и email резолвятся из флагов / env / yaml / `git config` в этом порядке. Скрытый алиас: `srekit lic`.

## Синопсис

```bash
srekit license [flags]
```

## Флаги

| Флаг | Обязательный | Описание |
|---|---|---|
| `--type` | нет | `wtfpl` (default), `mit`, или `apache2` |
| `--year` | нет | Год копирайта (default: текущий) |
| `--author` | нет | Переопределить имя автора |
| `--email` | нет | Переопределить email |

Плюс [общие output-флаги](index.md#shared-output-flags). По умолчанию эта команда печатает в **stdout** (default-пути нет) — передай `--out LICENSE` чтобы записать в файл.

## Примеры

WTFPL в stdout:

```bash
srekit license --stdout
# или просто: srekit license  (stdout — default sink)
```

MIT в файл:

```bash
srekit license --type mit --out LICENSE
```

Apache 2.0 с явными year и author:

```bash
srekit license --type apache2 --year 2026 --author "Mikhail Savin" \
  --email jtprogru@gmail.com --out LICENSE
```

Разовый override через env:

```bash
SREKIT_AUTHOR="Alice Example" SREKIT_EMAIL=alice@example.com \
  srekit license --type mit --stdout
```

## Резолв автора {#author-resolution}

Если `--author` не передан, srekit идёт (первый непустой выигрывает):

1. `SREKIT_AUTHOR` env / `author:` в `~/.srekit.yaml` / `full_name:` в yaml
2. `git config user.name`

Та же цепочка для `--email` с `SREKIT_EMAIL` / `email:` / `git config user.email`. Если оба пустые — команда падает с понятной ошибкой "set --author or configure git user.name".

## Структура данных для шаблона

```go
struct {
    Year, Author, Email string
}
```

## См. также

- [Конфигурация](../guides/configuration.md)
- [`srekit config init`](config.md#config-init) — заполнить `~/.srekit.yaml`.
