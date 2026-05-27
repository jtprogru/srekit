# Обзор команд

Поверхность srekit — плоское дерево cobra-подкоманд. Каждая команда-генератор
производит один артефакт (Markdown или LICENSE) и разделяет общий набор
output-флагов; управляющие команды (`templates`, `config`) группируют свои
подкоманды.

## Генераторы

| Команда | Что генерирует | Обязательные флаги |
|---|---|---|
| [`srekit task`](task.md) | Investigation log (alias: `sretask`) | `--title` |
| [`srekit license`](license.md) | `LICENSE` файл (alias: `lic`) | — |
| [`srekit incident`](incident.md) | Live-incident report | `--title` |
| [`srekit postmortem`](postmortem.md) | Постмортем (Google SRE-style) | `--title` |
| [`srekit rfc`](rfc.md) | RFC / ADR | `--title` |
| [`srekit runbook`](runbook.md) | Operational runbook | `--title` |
| [`srekit changelog`](changelog.md) | Keep a Changelog scaffold | — |
| [`srekit oncall-report`](oncall-report.md) | Недельный отчёт дежурного | `--team` |
| [`srekit slo`](slo.md) | SLO / SLI документ | `--service` |
| [`srekit ebp`](ebp.md) | Error budget policy | `--service` |
| [`srekit capacity`](capacity.md) | Capacity plan | `--service` |
| [`srekit retro`](retro.md) | Спринт-ретро (Start / Stop / Continue) | `--team` |

## Управление

| Команда | Назначение |
|---|---|
| [`srekit templates`](templates.md) | Управление кастомной директорией шаблонов: `init`, `pull`, `list`, `validate`, `diff`, `upgrade` |
| [`srekit config`](config.md) | Скаффолд `~/.srekit.yaml`: `init` |
| [`srekit completion`](completion.md) | Shell автодополнение: `bash`, `zsh`, `fish`, `powershell` |

## Общие output-флаги {#shared-output-flags}

Каждый генератор принимает:

| Флаг | Эффект |
|---|---|
| `--out FILE` | записать в FILE (отказ перезаписать без `--force`) |
| `--stdout` | напечатать в stdout |
| `--force` | перезаписать существующий FILE |
| `--dry-run` | показать что бы записал, не писать |
| `--template FILE` | использовать этот файл шаблона вместо embedded |
| `--json` | отдать данные шаблона как JSON (default sink: stdout) |

Persistent-флаг `--templates-dir DIR` (или env `SREKIT_TEMPLATES_DIR`, или
`templates_dir:` в `~/.srekit.yaml`) подключает кастомную директорию
шаблонов, чьи файлы переопределяют embedded. Отсутствующие файлы прозрачно
фолбэчатся.

Полный набор правил резолва — в [Конфигурация](../guides/configuration.md).
