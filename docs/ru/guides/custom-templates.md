# Кастомные шаблоны: workflow

Built-in шаблоны srekit — это разумные defaults, но любой команде
рано или поздно хочется их доработать: принять внутреннюю терминологию,
добавить специфичную секцию или убрать ненужную. Workflow кастомных
шаблонов даёт это сделать, не теряя возможности забирать upstream-
изменения из будущих релизов srekit.

## Ментальная модель

```
                    embedded   (в бинаре srekit, обновляется с
                       ↓        каждым релизом)
                       ↓
            .srekit-embedded/   (снапшот embedded на момент
                       ↓        последнего init/upgrade)
                       ↓
      <твоя templates dir>/    (твои живые кастомизированные файлы;
                                под твоим git-remote)
```

Сидкар `.srekit-embedded/` внутри templates dir хранит byte-exact копию
того, что *был* embedded в последний sync. Это merge-base, который
позволяет `templates upgrade` делать настоящий 3-way merge — не
затаптывая твои правки и не отказываясь их трогать.

## End-to-end

### 1. Скаффолд

```bash
srekit templates init ~/.srekit/templates
# Templates scaffolded in /Users/you/.srekit/templates (13 files + TEMPLATES.md)
#
# Next steps — connect a remote (SSH recommended) and push:
#   cd /Users/you/.srekit/templates
#   git remote add origin git@github.com:<owner>/<repo>.git
#   git add . && git commit -m 'initial templates'
#   git push -u origin main
#
# Then point srekit at this directory:
#   echo 'templates_dir: /Users/you/.srekit/templates' >> ~/.srekit.yaml
#   # or: export SREKIT_TEMPLATES_DIR=/Users/you/.srekit/templates
```

Несколько деталей:

- `templates init` по умолчанию делает `git init` (`--no-git` чтобы
  пропустить).
- Сидит `.srekit-embedded/` как сидкар и дописывает его в `.gitignore`
  чтобы он никогда не попал в team-репо шаблонов.
- Без явного `[dir]` резолвит `--templates-dir` / `SREKIT_TEMPLATES_DIR`
  / `templates_dir:` в yaml; fallback `~/.srekit/templates`.

### 2. Подключить

```bash
srekit config init   # задаёт author, email, templates_dir
# или вручную:
echo 'templates_dir: ~/.srekit/templates' >> ~/.srekit.yaml
```

### 3. Кастомизировать

```bash
cd ~/.srekit/templates
$EDITOR runbook.md.tmpl
git commit -am "runbook: add 'communications' section"
```

Всё что ты меняешь — твоё. Embedded fallback кикается только для
файлов, которых у тебя нет (например, новый шаблон, добавленный в
будущем бинаре, которого у тебя в dir ещё нет).

### 4. Провалидировать

Перед push'ем правок убедись, что шаблоны рендерятся:

```bash
srekit templates validate
# OK    capacity.md.tmpl
# OK    changelog.md.tmpl
# ...
```

Опечатка в поле шаблона (`{{ .Servce }}` вместо `{{ .Service }}`) тут
падает с подсвеченным проблемным полем.

### 5. Push в team-remote

```bash
git push
```

Другие инженеры команды ставят `templates_dir:` на свежий клон того же
remote.

### 6. Pull team-изменений

Когда коллега запушил тюнинг:

```bash
srekit templates pull            # git pull --ff-only на твою templates dir
srekit templates pull --rebase   # если есть локальные коммиты для rebase
```

### 7. Подтянуть новый бинарь srekit

Когда `brew upgrade srekit` (или `go install ...@latest`) и новый
бинарь несёт изменения шаблонов:

```bash
srekit templates list             # что изменилось
srekit templates diff             # полный unified diff vs embedded
srekit templates upgrade          # 3-way merge новых embedded в твою dir
```

Поведение upgrade:

| Состояние файла `task.md.tmpl` | Результат |
|---|---|
| Отсутствует в твоей dir | Скопирован (новый шаблон в бинаре) |
| Идентичен embedded | Skip; снапшот обновлён |
| Ты правил, embedded без изменений | Silent no-op |
| Ты не трогал, embedded изменился | Fast-forward к новому embedded |
| Оба изменены (не пересекаясь) | Clean 3-way merge |
| Оба изменены (пересекаясь) | Маркеры конфликта + non-zero exit |

На конфликте разрешаешь `<<<<<<<` маркеры как обычный merge, commit,
push. Снапшот уже в новом embedded — следующий `templates upgrade`
считает твоё разрешение новой базой.

## Сценарии восстановления

### "Я наскаффолдил не туда"

Часто бывает, когда поставил `templates_dir:` но забыл передать
соответствующий arg в `templates init`. Просто повтори:

```bash
srekit templates init   # учтёт templates_dir из yaml
rm -rf ~/.srekit/templates   # старый ghost
```

### "Я потерял .srekit-embedded сидкар"

Может, удалил dir или восстановил из бэкапа без скрытых директорий. Не
беда:

```bash
srekit templates upgrade
# все твои customized файлы возвращаются как "skipped: no merge base"
# но snapshot теперь засеян — *следующий* upgrade сделает 3-way
```

### "Хочу чистый старт"

```bash
srekit templates init --force   # перезапишет всё embedded'ом
```

## Зачем нужен `templates init --no-git`

Если держишь шаблоны внутри родительского репо (например в infra-репо
по пути `infra/sre-templates/`), `git init` внутри этой поддиректории
создаст nested-репо. Передай `--no-git`:

```bash
srekit templates init ./infra/sre-templates --no-git
git -C infra add sre-templates && git -C infra commit -m "vendor srekit templates"
```

`templates pull` тогда становится "`cd infra && git pull`" — srekit pull
не знает про родительский репо, так что просто запускай обычный git.

## См. также

- [`srekit templates`](../commands/templates.md) — полный reference по
  подкомандам.
- [Конфигурация](configuration.md) — как `templates_dir:` резолвится
  через флаги / env / yaml.
