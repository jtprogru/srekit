# Кастомные шаблоны: workflow

Встроенные шаблоны srekit — это разумные значения по умолчанию, но любой команде рано или поздно хочется их доработать: принять внутреннюю терминологию, добавить специфичную секцию или убрать ненужную. Сценарий работы с кастомными шаблонами даёт это сделать, не теряя возможности забирать upstream-изменения из будущих релизов srekit.

!!! tip "Начни с готового репозитория"
    [`jtprogru/sre-templates`](https://github.com/jtprogru/sre-templates) — публичный репозиторий шаблонов ровно в той раскладке, что описана ниже. Склонируй его вместо скаффолда с нуля или форкни под свой remote:

    ```bash
    git clone https://github.com/jtprogru/sre-templates ~/.config/srekit/templates
    echo 'templates_dir: ~/.config/srekit/templates' >> ~/.config/srekit/config.yaml
    ```

    Дальше `srekit templates pull` держит его в синхроне с remote, а `srekit templates upgrade` вмёрживает новый embedded-контент из будущих бинарей.

## Как это устроено

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

Скрытая директория `.srekit-embedded/` внутри твоей templates dir хранит byte-exact копию того, что *был* embedded в последний sync. Это merge-base, который позволяет `templates upgrade` делать настоящий 3-way merge — не затаптывая твои правки и не отказываясь их трогать.

## Шаг за шагом

### 1. Скаффолд

```bash
srekit templates init ~/.config/srekit/templates
# Templates scaffolded in /Users/you/.config/srekit/templates (9 files + TEMPLATES.md)
#
# Next steps — connect a remote (SSH recommended) and push:
#   cd /Users/you/.config/srekit/templates
#   git remote add origin git@github.com:<owner>/<repo>.git
#   git add . && git commit -m 'initial templates'
#   git push -u origin main
#
# Then point srekit at this directory:
#   echo 'templates_dir: /Users/you/.config/srekit/templates' >> ~/.config/srekit/config.yaml
#   # or: export SREKIT_TEMPLATES_DIR=/Users/you/.config/srekit/templates
```

Несколько деталей:

- `templates init` по умолчанию делает `git init` (`--no-git` чтобы пропустить).
- Создаёт служебную директорию `.srekit-embedded/` (снапшот embedded на момент init) и дописывает её в `.gitignore`, чтобы она никогда не попала в team-репо шаблонов.
- Без явного `[dir]` резолвит `--templates-dir` / `SREKIT_TEMPLATES_DIR` / `templates_dir:` в конфиге; запасной вариант — `$XDG_CONFIG_HOME/srekit/templates` либо pre-XDG `~/.srekit/templates`, если такая директория уже существует.

### 2. Подключить

```bash
srekit config init   # задаёт author, email, templates_dir
# или вручную:
echo 'templates_dir: ~/.config/srekit/templates' >> ~/.config/srekit/config.yaml
```

### 3. Кастомизировать

```bash
cd ~/.config/srekit/templates
$EDITOR runbook.yaml      # правишь v1-артефакт: добавляешь секцию, тюнишь meta_bullets, etc.
git commit -am "runbook: add 'communications' section"
```

Всё что ты меняешь — твоё. Embedded запасной вариант срабатывает только для файлов, которых у тебя нет (например, новый шаблон, добавленный в будущем бинаре, которого у тебя в dir ещё нет). Схема v1-артефакта (frontmatter / title / meta_bullets / header_body / sections / footer_body) описана в [postmortem reference](../commands/postmortem.md). Блоки разделяются ровно одной пустой строкой, а `footer_body` — хвостовой материал уровня документа (канонический пример — блок link reference definitions в changelog): у него нет id и нет места в массиве `sections`.

### 4. Провалидировать

Перед push'ем правок убедись, что шаблоны рендерятся:

```bash
srekit templates validate
# OK    slo.yaml
# OK    changelog.yaml
# ...
```

Каждый `.yaml`-артефакт парсится через `sections.ParseArtifact` — ловит структурные ошибки (неизвестные типы секций, дублирующиеся ID, пропущенные обязательные поля). Bespoke `.tmpl`-файлы в твоей dir получают parse-only валидацию (field-shape check невозможен без sample-data registry).

### 5. Push в team-remote

```bash
git push
```

Другие инженеры команды ставят `templates_dir:` на свежий клон того же remote.

### 6. Pull team-изменений

Когда коллега запушил тюнинг:

```bash
srekit templates pull            # git pull --ff-only на твою templates dir
srekit templates pull --rebase   # если есть локальные коммиты для rebase
```

### 7. Подтянуть новый бинарь srekit

Когда `brew upgrade srekit` (или `go install ...@latest`) и новый бинарь несёт изменения шаблонов:

```bash
srekit templates list             # что изменилось
srekit templates diff             # полный unified diff vs embedded
srekit templates upgrade          # 3-way merge новых embedded в твою dir
```

Поведение upgrade:

| Состояние файла `task.yaml` | Результат |
|---|---|
| Отсутствует в твоей dir | Скопирован (новый шаблон в бинаре) |
| Идентичен embedded | Skip; снапшот обновлён |
| Ты правил, embedded без изменений | Silent no-op |
| Ты не трогал, embedded изменился | Fast-forward к новому embedded |
| Оба изменены (не пересекаясь) | Clean 3-way merge |
| Оба изменены (пересекаясь) | Маркеры конфликта + non-zero exit |

На конфликте разрешаешь `<<<<<<<` маркеры как обычный merge, commit, push. Снапшот уже в новом embedded — следующий `templates upgrade` считает твоё разрешение новой базой.

## Языковые варианты { #language-variants }

Один артефакт шипится на двух языках: `changelog.yaml` и `changelog.ru.yaml`. Это два файла, а не две половины одного, и жизненный цикл обращается с ними именно так: оба скаффолдит `templates init`, оба видны в `templates list` отдельными записями со своим статусом, у обоих свой снапшот в `.srekit-embedded/`, и `templates upgrade` мёржит каждый против его собственной базы. Правка своей копии одного языка не трогает другой.

Как резолвится вариант, когда [`srekit changelog --lang ru`](../commands/changelog.md#ru-variant) его просит:

1. `changelog.ru.yaml` в твоей директории шаблонов
2. `changelog.ru.yaml`, вкомпилированный в бинарь
3. `changelog.yaml` в твоей директории шаблонов
4. `changelog.yaml`, вкомпилированный в бинарь

Вариант ищется по *всей* цепочке источников раньше, чем в любом из них пробуется запасной вариант. Порядок важен в одном случае: если ты кастомизировал `changelog.yaml` и своего русского варианта у тебя нет, `--lang ru` отдаст embedded-русский артефакт, а не твою английскую кастомизацию. Ты просил русский, и английский файл — даже свой — не является ответом на этот вопрос.

Чтобы кастомизировать русский вариант, правь `changelog.ru.yaml` в своей директории шаблонов, как любой другой файл. Третий язык пока не выбирается: `--lang` принимает только `en` и `ru`, так что положенный рядом `changelog.de.yaml` нечем будет выбрать. Язык, варианта для которого нет нигде, молча откатывается к базовому артефакту, а не падает.

## Сценарии восстановления

### "Я наскаффолдил не туда"

Часто бывает, когда поставил `templates_dir:`, но забыл передать соответствующий arg в `templates init`, — и скаффолд лёг в дефолтную директорию. Повтори без аргумента, потом удали ghost:

```bash
srekit doctor           # config.templates-dir назовёт директорию, которая реально в силе
srekit templates init   # учтёт templates_dir из конфига
rm -rf ~/.config/srekit/templates   # дефолтная директория, если лишняя копия попала туда
```

### "Я снёс директорию `.srekit-embedded/`"

Может, удалил dir или восстановил из бэкапа без скрытых директорий. Не беда:

```bash
srekit templates upgrade
# все твои отредактированные файлы возвращаются как "skipped: no merge base"
# но snapshot теперь засеян — *следующий* upgrade сделает 3-way
```

### "Хочу чистый старт"

```bash
srekit templates init --force   # перезапишет всё embedded'ом
```

## Зачем нужен `templates init --no-git`

Если держишь шаблоны внутри родительского репо (например в infra-репо по пути `infra/sre-templates/`), `git init` внутри этой поддиректории создаст nested-репо. Передай `--no-git`:

```bash
srekit templates init ./infra/sre-templates --no-git
git -C infra add sre-templates && git -C infra commit -m "vendor srekit templates"
```

`templates pull` тогда становится "`cd infra && git pull`" — srekit pull не знает про родительский репо, так что просто запускай обычный git.

## См. также

- [`jtprogru/sre-templates`](https://github.com/jtprogru/sre-templates) — готовый репозиторий шаблонов, который можно склонировать или форкнуть.
- [`srekit templates`](../commands/templates.md) — полный reference по подкомандам.
- [Конфигурация](configuration.md) — как `templates_dir:` резолвится через флаги / env / yaml.
