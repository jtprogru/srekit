# Удалённые команды

В `srekit` v0.30.0 удалены три генератора: `capacity`, `retro` и `license` (вместе с алиасом `lic`). Это breaking change — на ветке `0.x` они допустимы между минорами и помечаются в [CHANGELOG](https://github.com/jtprogru/srekit/blob/main/CHANGELOG.md).

Последний релиз, где все три ещё есть, — **v0.29.3**.

## Что ты увидишь

Имена не исчезли молча. Запуск печатает объяснение и завершается ненулевым кодом:

```console
$ srekit capacity --service payments
Error: the "capacity" command was removed in v0.30.0. Capacity planning is spreadsheet work, not a text artifact srekit is good at. See https://jtprogru.github.io/srekit/migration/removed-commands/
```

Заглушки не разбирают аргументы, поэтому `srekit retro` без `--team` сообщает об удалении, а не про пропущенный флаг. В `srekit --help` они скрыты и будут убраны совсем на 1.0 — после этого имена станут обычными unknown-командами.

## Почему

`srekit` генерирует артефакты, которыми владеет дежурный или команда надёжности. Ретро спринта — agile-церемония; capacity planning — работа в таблице, которой Markdown-скелет мало помогает; LICENSE — разовый шаг настройки репозитория, унаследованный от команды `lic` в монолите [gch](https://github.com/jtprogru/gch), из которого srekit был извлечён.

У `license` была ещё и структурная цена: это единственная команда, чей render-путь читал файл шаблона, и ровно из-за неё существовали флаг `--template FILE` и вторая ветка рендера на Go-шаблонах. Обе ушли вместе с ней.

## Что делать вместо этого

### `capacity` и `retro`

Замены в дереве нет, и шаблоны больше не вкомпилированы — `capacity.yaml` или `retro.yaml` в твоей директории шаблонов нечем отрендерить.

Если документы всё же нужны:

- **Возьми уже сгенерированный документ** как статический шаблон и копируй его от цикла к циклу. Артефакты были скелетами; ничто в них не зависело от srekit в момент рендера.
- **Запинь v0.29.3**, если есть автоматизация, которую пока не хочется трогать:

    ```bash
    go install github.com/jtprogru/srekit@v0.29.3
    ```

    Либо, если ставил каской из Homebrew, зафиксируй установленную версию вместо обновления.

- **Достань текст шаблона** из истории git, если ты его кастомизировал и копии не осталось:

    ```bash
    git -C <srekit-checkout> show v0.29.3:internal/tmpl/templates/capacity.yaml
    git -C <srekit-checkout> show v0.29.3:internal/tmpl/templates/retro.yaml
    ```

### `license`

Используй license picker своего хостинга — в GitHub «Add file → Choose a license template» запишет нужный текст с твоим именем и годом. Либо скопируй текст один раз с [choosealicense.com](https://choosealicense.com/) и закоммить: LICENSE пишется однажды на репозиторий и не перегенерируется.

Если ты пользовался `srekit license --template ./my-license.tmpl` со своим телом лицензии — этого флага тоже больше нет. Лечение то же: закоммить отрендеренный файл один раз, ровно это флаг и производил.

## Что не затронуто

- Твоя директория шаблонов не тронута. Оставшийся `capacity.yaml` или `retro.yaml` просто никогда не загружается; `srekit templates list` переклассифицирует его из `customized` в `user-only`, а `srekit templates upgrade` соберёт его снапшот из `.srekit-embedded/`.
- Поведение выживших генераторов не изменилось. Ни один флаг у `task`, `postmortem`, `rfc`, `runbook`, `oncall-report`, `slo`, `ebp` и `changelog` не поменял имя, дефолт или смысл, и конверт `--json` прежний.
