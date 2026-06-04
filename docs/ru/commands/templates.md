# srekit templates

Управление кастомной директорией шаблонов, чьи файлы переопределяют embedded. Отсутствующие файлы прозрачно фолбэчатся на embedded — можно переопределить один шаблон или весь набор.

Группа из шести подкоманд, образующих жизненный цикл:

```
   init  →  pull  →  list  →  validate  →  diff  →  upgrade  →  ...
```

Развёрнутый гайд — в **[Кастомные шаблоны](../guides/custom-templates.md)**.

!!! tip "Готовый репозиторий шаблонов"
    [`jtprogru/sre-templates`](https://github.com/jtprogru/sre-templates) — публичный репозиторий ровно в той раскладке, что ждут эти подкоманды. Склонируй его как `templates_dir`, чтобы пропустить `init`, и работай через `pull` / `list` / `diff` / `upgrade`.

---

## `templates init [dir]` {#templates-init}

Скаффолд кастомной директории шаблонов из embedded-набора, опционально с `git init`. Дополнительно создаёт служебную директорию `.srekit-embedded/` со снапшотом embedded — её использует `templates upgrade` как merge-base — и дописывает её в `.gitignore`.

```bash
srekit templates init                     # резолвит templates_dir из конфига; fallback на ~/.srekit/templates
srekit templates init ./team-templates    # явная директория
srekit templates init --no-git            # пропустить git init
srekit templates init --force             # перезаписать существующие
```

**Флаги**: `--force`, `--no-git`. Аргумент `[dir]` выигрывает у конфига; без него — резолв через `--templates-dir` / `SREKIT_TEMPLATES_DIR` / yaml; fallback `~/.srekit/templates`.

---

## `templates pull` {#templates-pull}

Синхронизировать сконфигурированную директорию шаблонов с git-remote.

```bash
srekit templates pull            # git pull --ff-only (safe; падает на diverged branches)
srekit templates pull --rebase   # с --rebase
```

Вывод стримится напрямую из git — видно ровно что произошло.

---

## `templates list [dir]` {#templates-list}

Классифицировать каждый `*.tmpl` относительно embedded-набора: `identical`, `customized`, `user-only`, `embedded-only`.

```bash
srekit templates list                       # таблица
srekit templates list --json | jq           # camelCase ключи: name, status, userPath
srekit templates list --filter customized   # только один класс
```

**Флаги**: `--json`, `--filter STATE`. Работает без сконфигурированной user dir (показывает embedded-набор как `embedded-only`), то есть заодно служит discovery-командой "что отгружает этот binary".

!!! note "JSON shape"
    `templates list --json` отдаёт camelCase ключи (`name`, `status`,
    `userPath`) — то же camelCase соглашение, что и `--json` у
    генераторов.

---

## `templates validate [dir]` {#templates-validate}

Провалидировать каждый артефакт в твоей templates dir. Per-формат:

- `<name>.yaml` (v1 artifact) — `sections.ParseArtifact` гонит структурную валидацию: поддерживаемая версия, непустой список секций, уникальные ID, известный `type` (`text` / `list` / `table`), required-поля заполнены.
- `<name>.sections.yaml` (legacy v0.13.x sidecar) — `sections.ParseManifest` те же структурные проверки на legacy-раскладку.
- `<name>.tmpl` — Go-template parse-only с общим FuncMap. Ловит syntax-errors; опечатки в полях не ловятся (с v0.20.0 ни одного `.tmpl` в embed нет, sample data для exec не существует).

```bash
srekit templates validate
```

Не-zero exit если что-то упало.

---

## `templates diff [dir]` {#templates-diff}

Unified diff между user-шаблонами и embedded-версиями через `git diff --no-index`.

```bash
srekit templates diff                    # полный diff каждого изменённого файла
srekit templates diff --name-only        # только имена
srekit templates diff --no-color         # без цвета
```

User-only шаблоны (без embedded-counterpart) маркируются как `user-only`. Идентичные пропускаются.

---

## `templates upgrade [dir]` {#templates-upgrade}

3-way merge embedded-изменений в user dir. Снапшот `.srekit-embedded/` из последнего init/upgrade служит merge-base.

Per-file поведение:

| Состояние файла vs binary | Результат |
|---|---|
| Отсутствует | Скопировать (`+ added`) |
| Идентичен embedded | Пропуск; снапшот в синке |
| Upstream без изменений, user правил | Silent no-op |
| User не трогал, upstream изменился | Fast-forward (`~ updated`) |
| Оба расходятся, base есть | `git merge-file --diff3` — clean → `~ merged`, conflict → `X conflict` + non-zero exit |
| Оба расходятся, base нет | Skip + засеять snapshot для следующего раза |

```bash
srekit templates upgrade
srekit templates upgrade --dry-run   # preview без записи
srekit templates upgrade --force     # перезаписать кастомизации (без merge)
```

`TEMPLATES.md` всегда обновляется — это reference, не точка кастомизации. На конфликт команда возвращает non-zero и пишет маркеры `<<<<<<<` / `>>>>>>>`; разрешаешь, потом re-run.

Snapshot GC (v0.14.0+): в конце каждого upgrade orphan-снапшоты в `.srekit-embedded/` для артефактов которые больше не в embed удаляются. Summary-строка показывает их количество.

---

## `templates migrate [dir]` {#templates-migrate}

Best-effort конвертер: превращает legacy `.tmpl` (и опц. `.sections.yaml` sidecar'ы) в v1 single-file `<name>.yaml` формат (введён в v0.14.0). Это путь миграции для templates dir'ов, инициализированных до v0.14.0.

```bash
srekit templates migrate                # dry-run: печатает converted YAML для каждого .tmpl
srekit templates migrate ./team-templates --apply   # пишет файлы <name>.yaml
```

**Что делает per-file:**

1. Парсит frontmatter `.tmpl` (между `---` / `---`), H1, meta_bullets (`- **X:** Y` после H1) и `## ` section блоки.
2. Если рядом есть `<name>.sections.yaml`, его список секций имеет приоритет над heuristic-парсингом из `.tmpl` (это v0.13.x → v1 case).
3. Минимальный type inference для секций: GFM-таблицы → `type: table`; всё остальное → `type: text` с `default_body` as-is.
4. Секции, содержащие Go-template control flow (`{{ if }}` / `{{ range }}` / `{{ with }}`), оборачиваются в `git merge`-style diff-маркеры — конвертер не пытается перевести control flow в typed-sections словарь. В output `OK (with diff markers — review needed)` помечает такие файлы для ручной доработки.
5. Section ID'ы берутся из английской части билингвальных заголовков (например `Контекст (Context)` → `context`); иначе из slug'а всего заголовка.
6. Новый `<name>.yaml` пишется рядом с `.tmpl`. Оригинальные `.tmpl` и `.sections.yaml` **не удаляются** — посмотри новый YAML, потом руками удали legacy-файлы когда готов.

**По умолчанию `--dry-run`** (печатает YAML preview); `--apply` пишет файлы.

**Ограничения:**

- Template-выражения внутри секционных body передаются as-is. Если новый генератор использует другую data-форму (например `.Meta.Title` вместо `.Title`), ссылки придётся обновить руками.
- Списки с intro-текстом (`_italic_` за которым `- items`) остаются `type: text`, а не `type: list` с `default_body`. Если хочешь typed-форму — рефакторь руками.
- License-шаблоны (`license_*.tmpl`) пропускаются — они зашиты в бинарь с v0.14.0 и не мигрируют.

**Флаги**: `--apply` (пишет файлы; default — dry-run).

---

## См. также

- [Кастомные шаблоны](../guides/custom-templates.md) — развёрнутый гайд.
- [`jtprogru/sre-templates`](https://github.com/jtprogru/sre-templates) — готовый репозиторий шаблонов: склонировать или форкнуть.
- [`srekit config`](config.md) — указать srekit на твою templates dir через `~/.srekit.yaml`.
