# Обзор команд

Поверхность srekit — плоское дерево cobra-подкоманд. Каждая команда-генератор производит один Markdown-артефакт и разделяет общий набор output-флагов; управляющие команды (`templates`, `config`) группируют свои подкоманды, а `doctor` отчитывается об окружении, которое резолвят все остальные.

## Генераторы

| Команда | Что генерирует | Обязательные флаги |
|---|---|---|
| [`srekit task`](task.md) | Investigation log (alias: `sretask`) | `--title` |
| [`srekit postmortem`](postmortem.md) | Постмортем (Google SRE-style) | `--title` |
| [`srekit rfc`](rfc.md) | RFC / ADR | `--title` |
| [`srekit runbook`](runbook.md) | Operational runbook | `--title` |
| [`srekit changelog`](changelog.md) | Заготовка Keep a Changelog | — |
| [`srekit oncall-report`](oncall-report.md) | Недельный отчёт дежурного | `--team` |
| [`srekit slo`](slo.md) | SLO / SLI документ | `--service` |
| [`srekit ebp`](ebp.md) | Error budget policy | `--service` |

## Управление

| Команда | Назначение |
|---|---|
| [`srekit templates`](templates.md) | Управление кастомной директорией шаблонов: `init`, `pull`, `list`, `validate`, `diff`, `upgrade`, `migrate` |
| [`srekit config`](config.md) | Скаффолд конфиг-файла (`$XDG_CONFIG_HOME/srekit/config.yaml`): `init` |
| [`srekit doctor`](doctor.md) | Диагностика окружения только на чтение: конфиг, шаблоны, identity, `git` |
| [`srekit completion`](completion.md) | Shell автодополнение: `bash`, `zsh`, `fish`, `powershell` |

## Общие output-флаги {#shared-output-flags}

Каждый генератор принимает:

| Флаг | Эффект |
|---|---|
| `--out FILE` | записать в FILE (отказ перезаписать без `--force`) |
| `--stdout` | напечатать в stdout |
| `--force` | перезаписать существующий FILE |
| `--dry-run` | показать что бы записал, не писать |
| `--json` | отдать данные шаблона как JSON (default sink: stdout) |

Команда, которая правит уже принадлежащий тебе документ, генератором не является и несёт более узкий набор: `--dry-run`, `--stdout` и `--json` в обычных значениях, но ни `--out`, ни `--force`. Её точка назначения — тот файл, на который её навели: вторая не имеет смысла, а guard от перезаписи защищал бы от собственной цели команды. Сегодня такая команда одна — [`srekit changelog release`](changelog.md#release).

Флага `--template FILE` нет. Он удалён в v0.30.0 вместе с `srekit license`, единственной командой, чей render-путь его читал; кастомизация per-artifact — положить `<name>.yaml` в `templates_dir`, см. [Кастомные шаблоны](../guides/custom-templates.md).

Команды `capacity`, `retro` и `license` удалены в v0.30.0 — см. [Удалённые команды](../migration/removed-commands.md).

Persistent-флаг `--templates-dir DIR` (или env `SREKIT_TEMPLATES_DIR`, или `templates_dir:` в конфиг-файле) подключает кастомную директорию шаблонов, чьи файлы переопределяют встроенные. Отсутствующие файлы прозрачно фолбэчатся.

У `srekit changelog` есть ещё один persistent-флаг — `--lang en|ru`, который наследуют подкоманды `release` и `validate`. Он выбирает словарь, который *генерируется*; язык существующего документа определяется по самому документу — см. [`srekit changelog`](changelog.md#ru-variant).

Полный набор правил резолва — в [Конфигурация](../guides/configuration.md).
