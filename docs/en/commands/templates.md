# srekit templates

Manage a custom templates directory whose files override the embedded ones. Missing files transparently fall back to embedded, so you can override a single template or the whole set.

The group has six subcommands forming a lifecycle:

```
   init  →  pull  →  list  →  validate  →  diff  →  upgrade  →  ...
```

See **[Custom templates workflow](../guides/custom-templates.md)** for the narrative version.

!!! tip "Ready-made templates repo"
    [`jtprogru/sre-templates`](https://github.com/jtprogru/sre-templates) is a public repo in the exact layout these subcommands expect. Clone it as your `templates_dir` to skip `init`, then use `pull` / `list` / `diff` / `upgrade` against it.

---

## `templates init [dir]` {#templates-init}

Scaffold a custom templates directory from the embedded set, optionally running `git init`. Seeds a `.srekit-embedded/` sidecar that `templates upgrade` uses as the merge base, and best-effort appends `.srekit-embedded/` to `.gitignore`.

```bash
srekit templates init                     # resolves templates_dir from config; falls back to ~/.srekit/templates
srekit templates init ./team-templates    # explicit
srekit templates init --no-git            # skip git init
srekit templates init --force             # overwrite existing files
```

**Flags**: `--force`, `--no-git`. The `[dir]` argument wins over config; omit it to use `--templates-dir` / `SREKIT_TEMPLATES_DIR` / yaml; fallback is `~/.srekit/templates`.

---

## `templates pull` {#templates-pull}

Sync the configured templates directory with its git remote.

```bash
srekit templates pull            # git pull --ff-only (safe; fails on diverged branches)
srekit templates pull --rebase   # use --rebase instead
```

Output is streamed from git directly, so you see exactly what happened.

---

## `templates list [dir]` {#templates-list}

Classify each `*.tmpl` against the embedded set: `identical`, `customized`, `user-only`, `embedded-only`.

```bash
srekit templates list                       # table
srekit templates list --json | jq           # camelCase keys: name, status, userPath
srekit templates list --filter customized   # narrow to one class
```

**Flags**: `--json`, `--filter STATE`. Works without a configured user dir (shows the embedded set as `embedded-only`), so it doubles as a "what does this binary ship" discovery tool.

!!! note "JSON shape"
    `templates list --json` emits camelCase keys (`name`, `status`,
    `userPath`) — the same camelCase convention used by generator
    `--json`.

---

## `templates validate [dir]` {#templates-validate}

Parse each `*.tmpl` with the same FuncMap srekit uses, and — for files whose names match a built-in template — execute them against canonical sample data. Catches both syntax errors and field-name typos (`{{ .Servce }}` instead of `{{ .Service }}`).

```bash
srekit templates validate
```

User-named templates (not matching a built-in filename) get parse-only validation since there's no canonical data shape to execute against. Non-zero exit if any file fails.

---

## `templates diff [dir]` {#templates-diff}

Unified diff between user templates and the embedded versions, via `git diff --no-index`.

```bash
srekit templates diff                    # full diff per modified file
srekit templates diff --name-only        # just file names
srekit templates diff --no-color         # plain text
```

User-only templates (no embedded counterpart) are reported as `user-only`. Identical files are skipped.

---

## `templates upgrade [dir]` {#templates-upgrade}

3-way merge embedded changes into the user dir. The `.srekit-embedded/` snapshot from the last init/upgrade serves as the merge base.

Per-file behavior:

| User dir state vs binary | Result |
|---|---|
| Missing | Copy in (`+ added`) |
| Identical to embedded | Skip; snapshot kept in sync |
| Upstream unchanged, user edited | Silent no-op |
| User untouched, upstream changed | Fast-forward (`~ updated`) |
| Both diverged, base available | `git merge-file --diff3` — clean → `~ merged`, conflicts → `X conflict` + non-zero exit |
| Both diverged, no base | Skip + seed snapshot for the next run |

```bash
srekit templates upgrade
srekit templates upgrade --dry-run   # preview, no writes
srekit templates upgrade --force     # overwrite customizations (skips merge)
```

`TEMPLATES.md` is always refreshed — it's a reference doc, not a customization point. On conflict, the command exits non-zero and writes `<<<<<<<` / `>>>>>>>` markers; resolve, then re-run.

Snapshot GC (v0.14.0+): at the end of every upgrade, orphaned snapshot files in `.srekit-embedded/` for artifacts no longer in embed are removed. The summary line reports the count.

---

## `templates migrate [dir]` {#templates-migrate}

Best-effort converter that turns legacy `.tmpl` files (and their optional `.sections.yaml` sidecars) into the v1 single-file `<name>.yaml` artifact format introduced in v0.14.0. This is the migration path for templates dirs initialized before v0.14.0.

```bash
srekit templates migrate                # dry-run: print converted YAML for each .tmpl
srekit templates migrate ./team-templates --apply   # write <name>.yaml files
```

**What it does per file:**

1. Parses the `.tmpl`'s frontmatter (between `---` / `---`), H1, meta bullets (`- **X:** Y` after the H1), and `## ` section blocks.
2. If a sibling `<name>.sections.yaml` exists, its section list takes precedence over the heuristic-parsed sections from the `.tmpl` (the v0.13.x → v1 case).
3. Section type inference is minimal: GFM tables → `type: table`; everything else → `type: text` with `default_body` verbatim.
4. Sections containing Go-template control flow (`{{ if }}` / `{{ range }}` / `{{ with }}`) are wrapped in `git merge`-style diff markers — the converter doesn't try to translate control flow into the typed-sections vocabulary. The output `OK (with diff markers — review needed)` flags these so you know which files need manual cleanup.
5. Section IDs are derived from the English portion of bilingual headings (e.g. `Контекст (Context)` → `context`); otherwise from a slugified whole heading.
6. The new `<name>.yaml` is written next to the `.tmpl`. Original `.tmpl` and `.sections.yaml` files are **not deleted** — review the new YAML, then delete the legacy files manually when ready.

**Defaults to `--dry-run`** (prints YAML preview); pass `--apply` to write files.

**Limitations:**

- Template expressions inside section bodies are passed through verbatim. If the new generator uses a different data shape (e.g. `.Meta.Title` instead of `.Title`), you'll need to update those references manually.
- Lists with intro text (e.g. `_italic_` followed by `- items`) are kept as `type: text` rather than `type: list` with `default_body`. Refactor manually if you want the typed shape.
- License templates (`license_*.tmpl`) are skipped — they're inlined in the binary since v0.14.0 and don't migrate.

**Flags**: `--apply` (write files; default is dry-run).

---

## See also

- [Custom templates workflow](../guides/custom-templates.md) — the full walkthrough.
- [`jtprogru/sre-templates`](https://github.com/jtprogru/sre-templates) — a ready-to-use templates repo to clone or fork.
- [`srekit config`](config.md) — point srekit at your templates dir via `~/.srekit.yaml`.
