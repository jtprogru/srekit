# srekit templates

Управление кастомной директорией шаблонов, чьи файлы переопределяют embedded. Отсутствующие файлы прозрачно фолбэчатся на embedded — можно переопределить один шаблон или весь набор.

Группа из шести подкоманд, образующих жизненный цикл:

```
   init  →  pull  →  list  →  validate  →  diff  →  upgrade  →  ...
```

Полная narrative-версия — в **[Кастомные шаблоны](../guides/custom-templates.md)**.

---

## `templates init [dir]` {#templates-init}

Скаффолд кастомной директории шаблонов из embedded-набора, опционально с `git init`. Сидит `.srekit-embedded/` сидкар, который `templates upgrade` использует как merge-base, и best-effort дописывает `.srekit-embedded/` в `.gitignore`.

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
    `userPath`) — отличается от PascalCase контракта `--json` у
    генераторов. Сведём к единому стандарту в v1.0.

---

## `templates validate [dir]` {#templates-validate}

Распарсить каждый `*.tmpl` с тем же FuncMap, что использует srekit, и — для файлов с именами built-in шаблонов — выполнить против canonical sample data. Ловит и syntax errors, и опечатки в полях (`{{ .Servce }}` вместо `{{ .Service }}`).

```bash
srekit templates validate
```

User-named шаблоны (не совпадают с built-in именами) получают только parse-only валидацию — нет canonical data shape для исполнения. Не-zero exit если что-то упало.

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

---

## См. также

- [Кастомные шаблоны](../guides/custom-templates.md) — end-to-end narrative.
- [`srekit config`](config.md) — указать srekit на твою templates dir через `~/.srekit.yaml`.
