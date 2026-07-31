## Context

See `proposal.md` — Why. The requirements are in `specs/environment-diagnostics/spec.md`; this document covers only how to satisfy them.

Three constraints shape the approach:

- The resolution logic `doctor` reports on already exists and is used by every other command: config path precedence (legacy wins over XDG if present), the templates-directory chain (`--templates-dir` → `SREKIT_TEMPLATES_DIR` → config key → default), per-file fallback from a user directory to the embedded set, author resolution. `doctor` must report exactly what those paths *actually* do. Any reimplementation drifts and starts lying.
- The template loader for an invocation lives on the command context, set in `PersistentPreRunE`. Package-level mutable state is prohibited in this project — the loader was moved into the context precisely because a global raced under parallel tests.
- `configureTemplates` already prints a fallback warning to stderr when the configured templates directory is missing or is not a directory. `doctor` must not double-report that as both a stderr line and a finding.

## Goals / Non-Goals

**Goals:**

- Report the resolved state by calling the same resolution code the generators call, so `doctor` cannot disagree with them.
- Keep every check independent and non-fatal, so one broken check never hides the other nine.
- Make the finding set data, not prose: one list of check descriptors walked by both renderers, so text and JSON output cannot diverge.

**Non-Goals:**

- No repair mode. `doctor --fix` is a separate change if it is ever wanted; a diagnostic that mutates is a diagnostic you stop trusting.
- No `--strict` flag promoting warnings to failures. Add it when someone asks; guessing at the CI policy now means shipping a flag nobody uses.
- No self-update or release check. That would require network access, which `distribution` forbids.
- No new check plugin mechanism. The check set is fixed and compiled in.

## Decisions

### Checks are a slice of descriptors, not a sequence of print statements

Each check is a value carrying its identifier, category, and a function returning a finding. `doctor` walks the slice in order, collects findings, and hands the collected list to one of two renderers. Deterministic ordering falls out of slice order rather than needing a sort, and the "always run every check" requirement becomes a `for` loop that cannot break early.

The slice is built per invocation inside the command's `RunE` and passed down, not stored in a package-level variable — same reason the loader lives on the context.

Alternative considered: a registry with `init()`-time registration, which is the usual Go shape for this. Rejected — it is exactly the package-level mutable state the project banned, and it buys nothing when the check set is fixed at compile time.

### A check reports a failure to inspect as a finding, never as an error return

The check function returns a finding, not `(finding, error)`. An unreadable directory or an unstartable subprocess becomes an `error`-status finding describing what could not be inspected. This is what makes "a broken check does not abort the rest" structural rather than a discipline every future check author has to remember.

### Findings are collected first, rendered second

Nothing prints while checks run. The overall status and the per-status counts both need the complete list, and `--quiet` filtering is a filter over that list. Streaming output would force the summary to be computed twice or printed out of order.

### Reuse, not reimplementation

- Config path and templates-directory resolution: call the existing helpers. To report *which* location won, `doctor` additionally stats both the legacy and the XDG paths itself — that is genuinely new information, not a second copy of the precedence rule, and the shadowing check needs both answers anyway.
- Config parse status: attempt the same load the CLI performs at startup and report the error it returns instead of discarding it. Startup deliberately ignores that error because config is optional; `doctor` exists to surface exactly what startup swallows.
- Templates drift: reuse the same classification `templates list` performs (`identical` / `customized` / `user-only` / `embedded-only`) and count the buckets. Reusing it means `doctor`'s drift count and `templates list`'s output can never disagree.
- Artifact parse health: reuse the same per-file dispatch `templates validate` uses — legacy manifest parser, v1 artifact parser, or Go-template parse — rather than a second parser.
- Author resolution: call the same resolver the generators call, then report which source it came from.

Alternative considered: shelling out to `srekit templates validate` and parsing its output. Rejected — a CLI re-invoking itself to read its own state is fragile and slow, and there is no process boundary worth having here.

### The stderr fallback warning is suppressed during `doctor`

`configureTemplates` runs in `PersistentPreRunE` for every command, `doctor` included, so a missing templates directory would print a stderr warning *and* a `warn` finding. Suppress the stderr line for this command specifically; the finding carries strictly more information — source of the setting, the path, and a remedy.

### `--quiet` filters `ok`, and stops there

For a generator `--quiet` suppresses chatter and keeps content. For `doctor` the findings *are* the content, so the analogous reading is: suppress what needs no action, keep what does. That gives the useful CI property that silence means healthy. JSON is a data document and is exempt — a consumer that asked for the full structure and got a filtered one has a bug on its hands, not a preference.

### Exit status

`error` findings map to exit `1`, the same code every other failing command in the CLI returns. No distinct code per severity: nobody scripts against exit `2` vs `3`, and `--json` already carries the detail for anything that needs to branch.

## Dependencies and binary size

No new module dependency. Every check uses the standard library (`os`, `os/exec`, `path/filepath`, `text/tabwriter`, `encoding/json`) plus internal packages already linked into the binary. The `git` check uses `exec.LookPath` and runs `git --version` through the existing context-aware subprocess pattern, so a Ctrl-C tears it down.

Nothing here reaches `net/http` or `crypto`. The added weight is one command file and its checks — negligible against the embedded templates already in the binary.

## Where state lives

None. Checks read the filesystem, the environment, and the command context; they write nothing and share nothing. The template loader is read from the command context as every other command does. No package-level variable is introduced, so the smoke tests can run in parallel with the rest of `cmd`.

The one caveat is the existing global config instance, which the config-dependent checks read. Tests covering those checks use the established non-parallel config helper, consistent with the rest of the suite.

## Risks / Trade-offs

- **A check that reports on resolution can drift from the resolution it reports on** → Every check calls the production helper rather than restating its rule. The one place `doctor` looks at raw paths (legacy vs XDG shadowing) is covered by a test that asserts the reported winner matches what a generator actually reads.
- **Check identifiers become a public contract the moment CI grepped for one** → They are named deliberately and documented in both locales from the start, and treated as a breaking change to rename.
- **Environment-dependent tests are the classic source of flakes** — the developer's real `$HOME`, a real `git`, a real config file → Tests point the config and templates paths at temp directories and set the XDG variable explicitly. The `git` check's absent-binary branch is exercised by pointing `PATH` at an empty directory rather than by assuming the CI image lacks git.
- **`doctor` reports the user's name and email in plain text** → These are the values already sitting in the user's git config and already stamped into every generated artifact. Redacting them would make the check useless for diagnosing "why is the wrong author on my RFC". Documented, not mitigated.
- **Ten `ok` lines is noise for the common healthy case** → `--quiet` reduces a healthy run to no output at all, which is what the CI use looks like anyway.

## Migration Plan

Purely additive: a new command name that was previously unused. Nothing to migrate, no rollback beyond reverting the commit. Users who never run `doctor` observe no change.
