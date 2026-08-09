# Старт

Эта страница ведёт от нуля до сгенерированного постмортема за три минуты.

## 1. Установка

=== "Homebrew (macOS / Linux)"

    ```bash
    brew tap jtprogru/tap
    brew install srekit
    ```

=== "go install"

    ```bash
    go install github.com/jtprogru/srekit@latest
    ```

    Требует Go 1.26+ (в модуле объявлен `go 1.26.4`).

=== "Готовые бинарники"

    Скачай со [страницы релизов на GitHub](https://github.com/jtprogru/srekit/releases)
    для `darwin`, `linux` или `freebsd` × `arm64` / `x86_64`.

Проверить установку:

```bash
srekit --version
# srekit version: 0.32.1
# from commit: ...
# built date: ...
# built by: goreleaser
```

## 2. Разовая настройка

Генераторы, которые пишут автора (`rfc`, `oncall-report`), резолвят личность в таком порядке:

1. флаги `--author` / `--email`
2. env-переменные `SREKIT_AUTHOR` / `SREKIT_EMAIL`
3. `author:` / `email:` в конфиг-файле
4. `git config user.name` / `git config user.email`

Конфиг лежит в `$XDG_CONFIG_HOME/srekit/config.yaml` (то есть `~/.config/srekit/config.yaml`, если переменная не выставлена). Старый `~/.srekit.yaml` продолжает читаться и выигрывает, если уже существует, — мигрировать ничего не нужно.

Если у тебя уже выставлен глобальный `git config`, **можно не настраивать ничего вообще**. Если хочешь yaml-файл:

```bash
srekit config init
# Author name [Mikhail Savin]: ⏎
# Email [jtprogru@gmail.com]: ⏎
# Templates dir (leave empty to use embedded templates only): ⏎
# Wrote /Users/jtprogru/.config/srekit/config.yaml
```

`srekit config init --yes` идёт без промптов, использует значения из флагов и `git config`. Полная картина — в [Конфигурация](guides/configuration.md).

## 3. Первый артефакт

Постмортем — хороший старт: он задействует резолв автора, имя файла из title и флаги `--out` / `--stdout`:

```bash
srekit postmortem --title "API outage" --severity SEV-1 --stdout
```

Это печатает в stdout. Записать в файл:

```bash
srekit postmortem --title "API outage" --severity SEV-1 \
  --start 2026-05-06T08:00Z --end 2026-05-06T09:30Z \
  --owner "@oncall" \
  --out postmortem-2026-05-06.md
```

Посмотри что внутри — каждая секция предзаполнена с билингвальными заголовками (`Постмортем (Postmortem)`) и SRE-каноническими полями (severity, timeline, impact, root cause, action items).

## 4. Единый набор флагов

Каждая команда-генератор поддерживает один набор output-флагов:

| Флаг | Эффект |
|---|---|
| `--out FILE` | записать в FILE (отказ перезаписать без `--force`) |
| `--stdout` | напечатать в stdout |
| `--force` | перезаписать существующий FILE |
| `--dry-run` | показать что бы записал, не писать |
| `--json` | отдать данные шаблона как JSON вместо рендеринга |

Флага `--template FILE` больше нет: он удалён в v0.30.0 вместе с `srekit license`, единственной командой, чей render-путь его читал. Кастомизация per-artifact — положить `<name>.yaml` в `templates_dir`, см. [Кастомные шаблоны](guides/custom-templates.md).

Если не передал ни `--out`, ни `--stdout`, у каждой команды есть разумный default-путь (например `investigation-<slug>.md` для `srekit task`, `oncall-<team>-<start>.md` для отчёта дежурного).

## 5. Проверить окружение

Если генератор подставил не того автора или кажется, что твои кастомные шаблоны игнорируются, — не гадай, спроси srekit, что он на самом деле зарезолвил:

```bash
srekit doctor
```

Команда только читает и не ходит в сеть. `srekit doctor --quiet` печатает только то, что требует внимания, так что тишина означает «всё в порядке» — см. [`srekit doctor`](commands/doctor.md).

## 6. Дальше

Этого хватит для повседневного использования. Если хочется глубже:

- **[Кастомные шаблоны](guides/custom-templates.md)** — форк встроенных артефактов в свой git-репо и подтягивание upstream-изменений через clean merge.
- **[jtprogru/sre-templates](https://github.com/jtprogru/sre-templates)** — готовый репозиторий шаблонов ровно в той раскладке, которую ждёт srekit. Склонируй его и нацель на него `templates_dir`, чтобы не скаффолдить с нуля.
- **[JSON-вывод](guides/json-output.md)** — пайплайны генераторов в `jq` для CI-скриптов и интеграций.
- **[Обзор команд](commands/index.md)** — полный reference по каждой подкоманде и флагу.
- **[Рецепты](../recipes.md)** — конкретные сценарии, где srekit совмещается с твоим тулчейном.
