# Removed commands

`srekit` v0.30.0 removes three generators: `capacity`, `retro`, and `license` (with its `lic` alias). This is a breaking change, permitted on the `0.x` line and flagged as such in the [CHANGELOG](https://github.com/jtprogru/srekit/blob/main/CHANGELOG.md).

The last release that shipped all three is **v0.29.3**.

## What you will see

The names are not silently gone. Running one prints an explanation and exits non-zero:

```console
$ srekit capacity --service payments
Error: the "capacity" command was removed in v0.30.0. Capacity planning is spreadsheet work, not a text artifact srekit is good at. See https://jtprogru.github.io/srekit/migration/removed-commands/
```

The stubs ignore their arguments, so `srekit retro` without `--team` reports the removal rather than a missing-flag error. They are hidden from `srekit --help`, and they will be dropped entirely at 1.0 — after that the names become ordinary unknown commands.

## Why

`srekit` generates artifacts an on-call engineer or a reliability team owns. Sprint retrospectives are an agile ceremony; capacity planning is a spreadsheet exercise that a Markdown scaffold does not meaningfully help with; a LICENSE file is a one-time repository setup step, inherited from the `lic` command in the [gch](https://github.com/jtprogru/gch) monolith srekit was extracted from.

`license` also carried a structural cost: it was the only command whose render path read a template file, which is why `--template FILE` and a second Go-template render path existed at all. Both are gone with it.

## What to do instead

### `capacity` and `retro`

There is no in-tree replacement, and the templates are no longer embedded — a `capacity.yaml` or `retro.yaml` in your templates directory has no command to render it.

If you want to keep producing these documents:

- **Keep a document you already generated** as a static template and copy it per cycle. The artifacts were scaffolds; nothing in them depended on srekit at rendering time.
- **Pin v0.29.3** if you have automation you are not ready to change:

    ```bash
    go install github.com/jtprogru/srekit@v0.29.3
    ```

    Or, with the Homebrew cask, pin the installed version rather than upgrading.

- **Recover the template text** from git history if you customized it and no longer have a copy:

    ```bash
    git -C <srekit-checkout> show v0.29.3:internal/tmpl/templates/capacity.yaml
    git -C <srekit-checkout> show v0.29.3:internal/tmpl/templates/retro.yaml
    ```

### `license`

Use the license picker your code host already offers — GitHub's "Add file → Choose a license template" writes the correct text with your name and year filled in. Otherwise copy the text once from [choosealicense.com](https://choosealicense.com/) and commit it; a LICENSE file is written once per repository and never regenerated.

If you were using `srekit license --template ./my-license.tmpl` with a custom body, that flag is gone too. The remedy is the same: commit the rendered file once, which is what the flag was producing.

## What is not affected

- Your templates directory is not touched. A leftover `capacity.yaml` or `retro.yaml` is simply never loaded; `srekit templates list` reclassifies it from `customized` to `user-only`, and `srekit templates upgrade` collects its snapshot from `.srekit-embedded/`.
- No surviving generator changes behaviour. No flag on `task`, `postmortem`, `rfc`, `runbook`, `oncall-report`, `slo`, `ebp`, or `changelog` changed name, default, or meaning, and the `--json` envelope is unchanged.
