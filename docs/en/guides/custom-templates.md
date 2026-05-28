# Custom templates workflow

srekit's built-in templates are good defaults, but every team eventually wants to tweak them — adopt internal terminology, add a specific section, or drop one they don't use. The custom-templates workflow lets you do that without losing the ability to pick up upstream changes from future srekit releases.

!!! tip "Start from a ready-made repo"
    [`jtprogru/sre-templates`](https://github.com/jtprogru/sre-templates) is a public templates repo in exactly the layout below. Clone it instead of scaffolding from scratch, or fork it to keep your own remote:

    ```bash
    git clone https://github.com/jtprogru/sre-templates ~/.srekit/templates
    echo 'templates_dir: ~/.srekit/templates' >> ~/.srekit.yaml
    ```

    From there `srekit templates pull` keeps it in sync with the remote and `srekit templates upgrade` merges in new embedded content from future binaries.

## How the pieces fit together

```
                    embedded   (in the srekit binary,
                       ↓        updates with each release)
                       ↓
            .srekit-embedded/   (snapshot of embedded as of
                       ↓        the last init/upgrade)
                       ↓
      <your templates dir>/    (your live, customized files;
                                under your own git remote)
```

The `.srekit-embedded/` sidecar inside your templates dir holds a byte-exact copy of what *was* embedded at the last sync. It's the merge base that lets `templates upgrade` do a true 3-way merge instead of either clobbering your edits or refusing to touch them.

## Walkthrough

### 1. Scaffold

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

A few details to notice:

- `templates init` runs `git init` by default (`--no-git` to skip).
- It seeds `.srekit-embedded/` as a sidecar and appends it to `.gitignore` so it never leaks into your team's templates repo.
- Without an explicit `[dir]` argument it resolves `--templates-dir` / `SREKIT_TEMPLATES_DIR` / `templates_dir:` in yaml, falling back to `~/.srekit/templates`.

### 2. Wire it up

```bash
srekit config init   # sets author, email, templates_dir
# or hand-edit:
echo 'templates_dir: ~/.srekit/templates' >> ~/.srekit.yaml
```

### 3. Customize

```bash
cd ~/.srekit/templates
$EDITOR runbook.md.tmpl
git commit -am "runbook: add 'communications' section"
```

Anything you change is yours. Embedded fallback only kicks in for files you haven't created (e.g. a new template added in a future binary that isn't in your dir yet).

### 4. Validate

Before pushing edits, make sure templates still render:

```bash
srekit templates validate
# OK    capacity.md.tmpl
# OK    changelog.md.tmpl
# ...
```

A typo in a Go template field (`{{ .Servce }}` instead of `{{ .Service }}`) fails here with the offending field highlighted.

### 5. Push to your shared remote

```bash
git push
```

Other engineers on your team set `templates_dir:` to a fresh clone of the same remote.

### 6. Pull upstream-of-team changes

When a teammate pushes a tweak:

```bash
srekit templates pull            # git pull --ff-only on your templates dir
srekit templates pull --rebase   # if you have local commits to rebase
```

### 7. Pick up new srekit binary

When you `brew upgrade srekit` (or `go install ...@latest`) and the new binary ships template changes:

```bash
srekit templates list             # see what changed
srekit templates diff             # full unified diff vs embedded
srekit templates upgrade          # 3-way merge new embedded into your dir
```

The upgrade flow:

| State of file `task.md.tmpl` | Result |
|---|---|
| Missing in your dir | Copied in (new template in the binary) |
| Identical to embedded | Skip; snapshot updated |
| You edited it, embedded unchanged | Silent no-op |
| You haven't touched it, embedded changed | Fast-forward to new embedded |
| Both changed (non-overlapping regions) | Clean 3-way merge |
| Both changed (overlapping regions) | Conflict markers + non-zero exit |

For conflicts, resolve `<<<<<<<` markers like any merge, commit, push. The snapshot is already at the new embedded — the next `templates upgrade` treats your resolution as the new base.

## Recovery paths

### "I scaffolded into the wrong directory"

Common when you set `templates_dir:` but forgot to pass the matching arg to `templates init`. Just rerun:

```bash
srekit templates init   # respects templates_dir from yaml
rm -rf ~/.srekit/templates   # old ghost
```

### "I lost my .srekit-embedded sidecar"

Maybe you nuked the dir or rebuilt from a backup that didn't include hidden dirs. No worries:

```bash
srekit templates upgrade
# all your customized files come back as "skipped: no merge base"
# but the snapshot is now seeded — the *next* upgrade does 3-way
```

### "I want a clean slate"

```bash
srekit templates init --force   # overwrites everything with embedded
```

## What `templates init --no-git` is for

If you keep your templates inside a parent repo (e.g. checking them into your infra repo at `infra/sre-templates/`), `git init` inside that subdir would create a nested repo. Pass `--no-git`:

```bash
srekit templates init ./infra/sre-templates --no-git
git -C infra add sre-templates && git -C infra commit -m "vendor srekit templates"
```

`templates pull` then becomes "`cd infra && git pull`" — srekit's pull won't know about the parent repo, so just run plain git.

## See also

- [`jtprogru/sre-templates`](https://github.com/jtprogru/sre-templates) — a ready-to-use templates repo you can clone or fork.
- [`srekit templates`](../commands/templates.md) — full subcommand reference.
- [Configuration & precedence](configuration.md) — how `templates_dir:` is resolved across flags / env / yaml.
