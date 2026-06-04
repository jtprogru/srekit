package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/jtprogru/srekit/internal/sections"
	"github.com/jtprogru/srekit/internal/tmpl"
)

func newTemplatesCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "templates",
		Short: "Manage the custom templates directory",
		Long:  "Scaffold, validate, diff, or sync a custom templates directory that overrides the embedded templates.",
	}
	c.AddCommand(newTemplatesInitCmd())
	c.AddCommand(newTemplatesPullCmd())
	c.AddCommand(newTemplatesValidateCmd())
	c.AddCommand(newTemplatesDiffCmd())
	c.AddCommand(newTemplatesUpgradeCmd())
	c.AddCommand(newTemplatesListCmd())
	return c
}

func newTemplatesListCmd() *cobra.Command {
	var (
		jsonOut bool
		filter  string
	)
	cmd := &cobra.Command{
		Use:   "list [dir]",
		Short: "List templates and their state vs the embedded set",
		Long: `Walks the configured templates directory (or [dir] if given) and the
binary's embedded set, classifying each *.tmpl as:

  identical       — user file matches embedded byte-for-byte
  customized      — user file exists but differs from embedded
  user-only       — user file with no embedded counterpart (your bespoke)
  embedded-only   — shipped in the binary, user has no override

Default output is a table. --json emits a sorted array suitable for
piping into jq. Use --filter STATE to narrow to a single class.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTemplatesList(cmd, args, jsonOut, filter)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit the listing as JSON")
	cmd.Flags().StringVar(&filter, "filter", "", "only show entries with this status (identical|customized|user-only|embedded-only)")
	return cmd
}

type templateEntry struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	User   string `json:"userPath,omitempty"`
}

func runTemplatesList(cmd *cobra.Command, args []string, jsonOut bool, filter string) error {
	switch filter {
	case "", "identical", "customized", "user-only", "embedded-only":
	default:
		return fmt.Errorf("--filter must be one of: identical, customized, user-only, embedded-only (got %q)", filter)
	}

	dir, err := resolveListDir(cmd, args)
	if err != nil {
		return err
	}

	entries, err := classifyTemplates(dir)
	if err != nil {
		return err
	}
	if filter != "" {
		filtered := entries[:0]
		for _, e := range entries {
			if e.Status == filter {
				filtered = append(filtered, e)
			}
		}
		entries = filtered
	}

	out := cmd.OutOrStdout()
	if jsonOut {
		b, err := json.MarshalIndent(entries, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintln(out, string(b))
		return nil
	}

	if len(entries) == 0 {
		fmt.Fprintln(out, "(no templates)")
		return nil
	}
	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "NAME\tSTATUS\tUSER PATH")
	for _, e := range entries {
		userPath := e.User
		if userPath == "" {
			userPath = "-"
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\n", e.Name, e.Status, userPath)
	}
	return tw.Flush()
}

// resolveListDir returns the user templates directory if one is configured
// or passed positionally; returns "" (with no error) when neither is set, so
// list still works on a fresh install and shows only the embedded set.
func resolveListDir(cmd *cobra.Command, args []string) (string, error) {
	if len(args) == 1 {
		return expandHome(args[0])
	}
	resolved, err := resolveTemplatesDir(cmd)
	if err != nil {
		return "", err
	}
	if resolved == "" {
		return "", nil
	}
	info, err := os.Stat(resolved)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return "", nil
	case err != nil:
		return "", fmt.Errorf("stat %s: %w", resolved, err)
	case !info.IsDir():
		return "", nil
	}
	return resolved, nil
}

func classifyTemplates(userDir string) ([]templateEntry, error) {
	embedded := map[string][]byte{}
	embNames, err := tmpl.EmbeddedNames()
	if err != nil {
		return nil, fmt.Errorf("read embedded templates: %w", err)
	}
	for _, name := range embNames {
		body, err := fs.ReadFile(tmpl.FS, "templates/"+name)
		if err != nil {
			return nil, fmt.Errorf("read embedded %s: %w", name, err)
		}
		embedded[name] = body
	}

	user := map[string][]byte{}
	if userDir != "" {
		userFiles, err := os.ReadDir(userDir)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", userDir, err)
		}
		for _, e := range userFiles {
			if e.IsDir() || !tmpl.IsTemplateArtifact(e.Name()) {
				continue
			}
			body, err := os.ReadFile(filepath.Join(userDir, e.Name()))
			if err != nil {
				return nil, fmt.Errorf("read %s: %w", e.Name(), err)
			}
			user[e.Name()] = body
		}
	}

	names := make(map[string]struct{}, len(embedded)+len(user))
	for n := range embedded {
		names[n] = struct{}{}
	}
	for n := range user {
		names[n] = struct{}{}
	}
	sorted := make([]string, 0, len(names))
	for n := range names {
		sorted = append(sorted, n)
	}
	sort.Strings(sorted)

	entries := make([]templateEntry, 0, len(sorted))
	for _, n := range sorted {
		emb, hasEmb := embedded[n]
		usr, hasUsr := user[n]
		var entry templateEntry
		entry.Name = n
		switch {
		case hasEmb && hasUsr && bytes.Equal(emb, usr):
			entry.Status = "identical"
			entry.User = filepath.Join(userDir, n)
		case hasEmb && hasUsr:
			entry.Status = "customized"
			entry.User = filepath.Join(userDir, n)
		case hasUsr:
			entry.Status = "user-only"
			entry.User = filepath.Join(userDir, n)
		default:
			entry.Status = "embedded-only"
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func newTemplatesUpgradeCmd() *cobra.Command {
	var (
		force  bool
		dryRun bool
	)
	cmd := &cobra.Command{
		Use:   "upgrade [dir]",
		Short: "Bring a templates directory in line with the binary's embedded set",
		Long: `For each template embedded in this srekit binary, compare against [dir]
(default: configured templates dir, falling back to $XDG_CONFIG_HOME/srekit/templates).

When a base snapshot from a previous init/upgrade is available
(.srekit-embedded/<name>), customized files are 3-way merged via
'git merge-file --diff3':
  - missing in user dir   → copy in
  - identical to embedded → skip
  - upstream unchanged    → leave user's customizations alone
  - upstream changed,
      user untouched      → fast-forward to new embedded
      both diverged       → 3-way merge; clean merges land silently with
                            an updated snapshot, conflicts are written
                            with markers (and the command exits non-zero
                            so CI / git status flag them)

Without a snapshot (older user dir scaffolded before this feature),
falls back to the additive behavior: skip customized files unless
--force overwrites them. The snapshot is refreshed on every run, so
the second upgrade onward will use 3-way.

TEMPLATES.md is always refreshed. Use --dry-run to preview.`,
		Example: `  # Preview what an upgrade would change
  srekit templates upgrade --dry-run

  # Apply, overwriting local customizations with embedded versions
  srekit templates upgrade --force`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTemplatesUpgrade(cmd, args, force, dryRun)
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "overwrite customized templates with the embedded versions (skips merge)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "report what would change without writing")
	return cmd
}

func runTemplatesUpgrade(cmd *cobra.Command, args []string, force, dryRun bool) error {
	dir, err := pickTemplatesDir(cmd, args)
	if err != nil {
		return err
	}
	names, err := tmpl.EmbeddedNames()
	if err != nil {
		return fmt.Errorf("read embedded templates: %w", err)
	}

	out := cmd.OutOrStdout()
	ctx := cmd.Context()
	var added, updated, mergedCount, conflictCount, unchanged, skipped int
	for _, name := range names {
		embedded, err := fs.ReadFile(tmpl.FS, "templates/"+name)
		if err != nil {
			return fmt.Errorf("read embedded %s: %w", name, err)
		}
		target := filepath.Join(dir, name)
		existing, err := os.ReadFile(target)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("read %s: %w", target, err)
		}
		missing := errors.Is(err, fs.ErrNotExist)
		snapshot, snapErr := readSnapshot(dir, name)
		hasSnapshot := snapErr == nil

		switch {
		case missing:
			if !dryRun {
				if err := os.WriteFile(target, embedded, 0o644); err != nil { //nolint:gosec // G306: same rationale as templates init
					return fmt.Errorf("write %s: %w", target, err)
				}
				if err := writeSnapshot(dir, name, embedded); err != nil {
					return err
				}
			}
			fmt.Fprintf(out, "+ added     %s\n", name)
			added++

		case bytes.Equal(existing, embedded):
			// Identical to embedded — make sure the snapshot is in sync for
			// future upgrades, then move on.
			if !dryRun && (!hasSnapshot || !bytes.Equal(snapshot, embedded)) {
				if err := writeSnapshot(dir, name, embedded); err != nil {
					return err
				}
			}
			unchanged++

		case force:
			if !dryRun {
				if err := os.WriteFile(target, embedded, 0o644); err != nil { //nolint:gosec // G306: see above
					return fmt.Errorf("write %s: %w", target, err)
				}
				if err := writeSnapshot(dir, name, embedded); err != nil {
					return err
				}
			}
			fmt.Fprintf(out, "~ updated   %s (was customized; --force overwrote)\n", name)
			updated++

		case hasSnapshot && bytes.Equal(snapshot, embedded):
			// Upstream unchanged since the last sync — user's customizations
			// are unrelated to the new binary. Leave them alone.
			skipped++

		case hasSnapshot && bytes.Equal(snapshot, existing):
			// User hasn't touched this file since last sync; upstream did.
			// Fast-forward to the new embedded.
			if !dryRun {
				if err := os.WriteFile(target, embedded, 0o644); err != nil { //nolint:gosec // G306: see above
					return fmt.Errorf("write %s: %w", target, err)
				}
				if err := writeSnapshot(dir, name, embedded); err != nil {
					return err
				}
			}
			fmt.Fprintf(out, "~ updated   %s (upstream change, no local edits)\n", name)
			updated++

		case hasSnapshot:
			// Both diverged from snapshot — true 3-way merge.
			mergedBody, conflicts, err := threeWayMerge(ctx, name, existing, snapshot, embedded)
			if err != nil {
				return err
			}
			if !dryRun {
				if err := os.WriteFile(target, mergedBody, 0o644); err != nil { //nolint:gosec // G306: see above
					return fmt.Errorf("write %s: %w", target, err)
				}
				// Snapshot moves to the new embedded regardless of conflicts:
				// the user's resolution lives in their file, and we want the
				// next upgrade's base to be the current upstream.
				if err := writeSnapshot(dir, name, embedded); err != nil {
					return err
				}
			}
			if conflicts > 0 {
				fmt.Fprintf(out, "X conflict  %s (%d conflict(s); resolve markers and re-run)\n", name, conflicts)
				conflictCount++
			} else {
				fmt.Fprintf(out, "~ merged    %s (3-way merge, no conflicts)\n", name)
				mergedCount++
			}

		default:
			// Customized + no snapshot → additive fallback. Seed the snapshot
			// with the current embedded so the next upgrade can do 3-way.
			if !dryRun {
				if err := writeSnapshot(dir, name, embedded); err != nil {
					return err
				}
			}
			fmt.Fprintf(out, "! skipped   %s (customized; no merge base — next upgrade will 3-way)\n", name)
			skipped++
		}
	}

	// TEMPLATES.md is the reference doc shipped with the binary — keep it
	// in sync on every upgrade so placeholder/FuncMap docs match the code.
	docsTarget := filepath.Join(dir, "TEMPLATES.md")
	if !dryRun {
		if err := os.WriteFile(docsTarget, tmpl.DocsMD, 0o644); err != nil { //nolint:gosec // G306: see above
			return fmt.Errorf("write %s: %w", docsTarget, err)
		}
		if err := ensureSnapshotIgnored(dir); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "srekit: could not update %s/.gitignore: %v\n", dir, err)
		}
	}

	fmt.Fprintf(out,
		"\n%s: %d added, %d updated, %d merged, %d conflict, %d unchanged, %d skipped. TEMPLATES.md refreshed.\n",
		summaryLabel(dryRun), added, updated, mergedCount, conflictCount, unchanged, skipped)
	if conflictCount > 0 {
		return fmt.Errorf("%d template(s) have unresolved conflict markers (resolve and re-run)", conflictCount)
	}
	return nil
}

func summaryLabel(dryRun bool) string {
	if dryRun {
		return "dry-run"
	}
	return "summary"
}

func newTemplatesInitCmd() *cobra.Command {
	var (
		force bool
		noGit bool
	)
	cmd := &cobra.Command{
		Use:   "init [dir]",
		Short: "Scaffold a custom templates directory from the built-in templates",
		Long: `Copy every built-in template into <dir>, write TEMPLATES.md with
the placeholder/FuncMap reference, and run git init in the directory
(unless --no-git). Refuses to overwrite existing files unless --force.

When [dir] is omitted, the target is resolved in the same order every
other 'templates' subcommand uses: --templates-dir / SREKIT_TEMPLATES_DIR
/ templates_dir: in the config file, falling back to $XDG_CONFIG_HOME/srekit/templates.`,
		Example: `  # Scaffold the default dir and git init it
  srekit templates init

  # Scaffold into a specific dir without git
  srekit templates init ./my-templates --no-git`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var dir string
			if len(args) == 1 {
				dir = args[0]
			} else {
				resolved, err := resolveTemplatesDir(cmd)
				if err != nil {
					return err
				}
				if resolved != "" {
					dir = resolved
				} else {
					dir = resolveDefaultTemplatesDir()
				}
			}
			expanded, err := expandHome(dir)
			if err != nil {
				return err
			}
			return runTemplatesInit(cmd, expanded, force, noGit)
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing template files in the target directory")
	cmd.Flags().BoolVar(&noGit, "no-git", false, "skip 'git init' in the target directory")
	return cmd
}

func runTemplatesInit(cmd *cobra.Command, dir string, force, noGit bool) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", dir, err)
	}

	names, err := tmpl.EmbeddedNames()
	if err != nil {
		return fmt.Errorf("read embedded templates: %w", err)
	}

	written := 0
	for _, name := range names {
		target := filepath.Join(dir, name)
		if _, err := os.Stat(target); err == nil && !force {
			return fmt.Errorf("%s already exists; pass --force to overwrite", target)
		}
		b, err := fs.ReadFile(tmpl.FS, "templates/"+name)
		if err != nil {
			return err
		}
		// Templates are public scaffolding — 0o644 matches render.go's convention.
		if err := os.WriteFile(target, b, 0o644); err != nil { //nolint:gosec // G306: same rationale as internal/render
			return fmt.Errorf("write %s: %w", target, err)
		}
		// Seed the merge-base snapshot so a future 'templates upgrade' can do
		// a true 3-way merge instead of falling back to additive logic.
		if err := writeSnapshot(dir, name, b); err != nil {
			return err
		}
		written++
	}

	// TEMPLATES.md is a reference doc — we always overwrite it on init so it
	// stays in sync with the binary's understanding of placeholders/FuncMap.
	docsTarget := filepath.Join(dir, "TEMPLATES.md")
	if err := os.WriteFile(docsTarget, tmpl.DocsMD, 0o644); err != nil { //nolint:gosec // G306: same rationale as internal/render
		return fmt.Errorf("write %s: %w", docsTarget, err)
	}

	// Keep the snapshot sidecar out of the user's git history.
	if err := ensureSnapshotIgnored(dir); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "srekit: could not update %s/.gitignore: %v\n", dir, err)
	}

	if !noGit {
		if err := gitInit(cmd, dir); err != nil {
			return err
		}
	}

	out := cmd.OutOrStdout()
	fmt.Fprintln(out)
	fmt.Fprintf(out, "Templates scaffolded in %s (%d files + TEMPLATES.md)\n", dir, written)
	fmt.Fprintln(out)
	if !noGit {
		fmt.Fprintln(out, "Next steps — connect a remote (SSH recommended) and push:")
		fmt.Fprintf(out, "  cd %s\n", dir)
		fmt.Fprintln(out, "  git remote add origin git@github.com:<owner>/<repo>.git")
		fmt.Fprintln(out, "  git add . && git commit -m 'initial templates'")
		fmt.Fprintln(out, "  git push -u origin main")
		fmt.Fprintln(out)
	}
	fmt.Fprintln(out, "Then point srekit at this directory:")
	fmt.Fprintf(out, "  echo 'templates_dir: %s' >> %s\n", dir, resolveConfigPath())
	fmt.Fprintln(out, "  # or: export SREKIT_TEMPLATES_DIR="+dir)
	return nil
}

func newTemplatesPullCmd() *cobra.Command {
	var rebase bool
	cmd := &cobra.Command{
		Use:   "pull",
		Short: "git pull the templates directory from its remote",
		Long: `Sync the custom templates directory with its git remote. Uses
'git pull --ff-only' by default (safe: fails on diverged branches);
pass --rebase to rebase local commits on top instead.

The directory is resolved from --templates-dir / SREKIT_TEMPLATES_DIR /
templates_dir: in ~/.srekit.yaml, falling back to ~/.srekit/templates.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runTemplatesPull(cmd, rebase)
		},
	}
	cmd.Flags().BoolVar(&rebase, "rebase", false, "use 'git pull --rebase' instead of --ff-only")
	return cmd
}

func runTemplatesPull(cmd *cobra.Command, rebase bool) error {
	dir, err := resolveTemplatesDir(cmd)
	if err != nil {
		return err
	}
	if dir == "" {
		// Nothing configured — fall back to the conventional default location.
		dir = resolveDefaultTemplatesDir()
	}
	if info, err := os.Stat(dir); err != nil {
		return fmt.Errorf("templates dir %s: %w (run 'srekit templates init' first)", dir, err)
	} else if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}
	if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("%s is not a git repository (re-run 'srekit templates init' without --no-git, or 'git init' it yourself)", dir)
		}
		return err
	}

	args := []string{"-C", dir, "pull"}
	if rebase {
		args = append(args, "--rebase")
	} else {
		args = append(args, "--ff-only")
	}
	git := exec.CommandContext(cmd.Context(), "git", args...)
	git.Stdout = cmd.OutOrStdout()
	git.Stderr = cmd.ErrOrStderr()
	if err := git.Run(); err != nil {
		return fmt.Errorf("git pull failed: %w", err)
	}
	return nil
}

func newTemplatesValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate [dir]",
		Short: "Parse and dry-run every .tmpl in the templates directory",
		Long: `Walk [dir] (default: configured templates dir, falling back to
$XDG_CONFIG_HOME/srekit/templates), parse each *.tmpl with the same FuncMap srekit uses,
and — for files whose names match a built-in template — execute the
template against canonical sample data to catch references to fields that
don't exist in the struct shape (e.g. a typo'd '.Servce' instead of
'.Service').

Files with names that aren't built-ins (your own custom templates used
via --template) get parse-only validation: we can catch syntax errors
but have no canonical data shape to execute against.

Exits non-zero if any file fails to parse or execute.`,
		Args: cobra.MaximumNArgs(1),
		RunE: runTemplatesValidate,
	}
	return cmd
}

func runTemplatesValidate(cmd *cobra.Command, args []string) error {
	dir, err := pickTemplatesDir(cmd, args)
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read %s: %w", dir, err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() || !tmpl.IsTemplateArtifact(e.Name()) {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)
	if len(names) == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "no template artifacts in %s\n", dir)
		return nil
	}

	out := cmd.OutOrStdout()
	errOut := cmd.ErrOrStderr()
	var failed int
	for _, name := range names {
		body, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			fmt.Fprintf(errOut, "FAIL  %s: read: %v\n", name, err)
			failed++
			continue
		}
		if strings.HasSuffix(name, ".sections.yaml") {
			if _, err := sections.ParseManifest(body); err != nil {
				fmt.Fprintf(errOut, "FAIL  %s: %v\n", name, err)
				failed++
				continue
			}
			fmt.Fprintf(out, "OK    %s\n", name)
			continue
		}
		switch err := tmpl.Validate(name, body); {
		case err == nil:
			fmt.Fprintf(out, "OK    %s\n", name)
		case errors.Is(err, tmpl.ErrUnknownTemplate):
			fmt.Fprintf(out, "OK    %s (parse-only — not a built-in template name)\n", name)
		default:
			fmt.Fprintf(errOut, "FAIL  %s: %v\n", name, err)
			failed++
		}
	}
	if failed > 0 {
		return fmt.Errorf("%d of %d template(s) failed validation", failed, len(names))
	}
	return nil
}

// noColorEnv reports whether the NO_COLOR convention (https://no-color.org)
// asks us to suppress color: the variable is present and non-empty.
func noColorEnv() bool {
	v, ok := os.LookupEnv("NO_COLOR")
	return ok && v != ""
}

func newTemplatesDiffCmd() *cobra.Command {
	var (
		nameOnly bool
		noColor  bool
	)
	cmd := &cobra.Command{
		Use:   "diff [dir]",
		Short: "Show how the templates directory has diverged from the embedded versions",
		Long: `For each *.tmpl in [dir] (default: configured templates dir), diff
the user's version against the version embedded in this srekit binary.
Shells out to 'git diff --no-index' so output is a familiar unified
diff. Templates that don't ship in the binary are reported as
"user-only"; templates byte-identical to the embedded version are
skipped.

Useful after 'srekit templates pull' or a binary upgrade to see what's
drifted.`,
		Example: `  # Full unified diff vs embedded
  srekit templates diff

  # Just the names of drifted templates
  srekit templates diff --name-only`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTemplatesDiff(cmd, args, nameOnly, noColor)
		},
	}
	cmd.Flags().BoolVar(&nameOnly, "name-only", false, "just list which files differ; no diff bodies")
	cmd.Flags().BoolVar(&noColor, "no-color", false, "disable color in diff output (also honors the NO_COLOR env var)")
	return cmd
}

func runTemplatesDiff(cmd *cobra.Command, args []string, nameOnly, noColor bool) error {
	dir, err := pickTemplatesDir(cmd, args)
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read %s: %w", dir, err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() || !tmpl.IsTemplateArtifact(e.Name()) {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)
	if len(names) == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "no template artifacts in %s\n", dir)
		return nil
	}

	out := cmd.OutOrStdout()
	tmpRoot, err := os.MkdirTemp("", "srekit-diff-")
	if err != nil {
		return fmt.Errorf("create tempdir: %w", err)
	}
	defer os.RemoveAll(tmpRoot)
	embeddedDir := filepath.Join(tmpRoot, "embedded")
	userDir := filepath.Join(tmpRoot, "user")
	if err := os.MkdirAll(embeddedDir, 0o755); err != nil {
		return fmt.Errorf("create tempdir: %w", err)
	}
	if err := os.MkdirAll(userDir, 0o755); err != nil {
		return fmt.Errorf("create tempdir: %w", err)
	}

	var diffsFound, userOnly int
	for _, name := range names {
		userPath := filepath.Join(dir, name)
		embedded, err := fs.ReadFile(tmpl.FS, "templates/"+name)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				fmt.Fprintf(out, "user-only  %s (no embedded counterpart)\n", name)
				userOnly++
				continue
			}
			return fmt.Errorf("read embedded %s: %w", name, err)
		}
		userBody, err := os.ReadFile(userPath)
		if err != nil {
			return fmt.Errorf("read %s: %w", userPath, err)
		}
		if string(embedded) == string(userBody) {
			continue
		}
		diffsFound++
		if nameOnly {
			fmt.Fprintf(out, "differs  %s\n", name)
			continue
		}
		// Lay out a parallel embedded/<name> + user/<name> structure inside
		// tmpRoot and run git diff from there with relative paths, so the
		// diff header reads "embedded/<name>" vs "user/<name>" instead of
		// dragging the absolute tempfile path through every patch.
		embeddedRel := filepath.Join("embedded", name)
		userRel := filepath.Join("user", name)
		if err := os.WriteFile(filepath.Join(tmpRoot, embeddedRel), embedded, 0o644); err != nil { //nolint:gosec // G306: same rationale as internal/render
			return fmt.Errorf("write tempfile: %w", err)
		}
		if err := os.WriteFile(filepath.Join(tmpRoot, userRel), userBody, 0o644); err != nil { //nolint:gosec // G306: same rationale as internal/render
			return fmt.Errorf("write tempfile: %w", err)
		}
		colorMode := "auto"
		if noColor || noColorEnv() {
			colorMode = "never"
		}
		gitArgs := []string{
			"-c", "core.pager=cat",
			"diff", "--no-index",
			"--color=" + colorMode,
			embeddedRel, userRel,
		}
		git := exec.CommandContext(cmd.Context(), "git", gitArgs...)
		git.Dir = tmpRoot
		git.Stdout = out
		git.Stderr = cmd.ErrOrStderr()
		// git diff exits 1 when files differ — that's the expected case here,
		// so we don't treat exit-1 as a runtime failure.
		_ = git.Run()
	}
	if diffsFound == 0 && userOnly == 0 {
		fmt.Fprintln(out, "all templates match the embedded versions")
	}
	return nil
}

// pickTemplatesDir resolves the directory to operate on: explicit positional
// arg wins, then the configured templates dir, falling back to the default
// location. It verifies the path exists and is a directory.
func pickTemplatesDir(cmd *cobra.Command, args []string) (string, error) {
	var dir string
	if len(args) == 1 {
		expanded, err := expandHome(args[0])
		if err != nil {
			return "", err
		}
		dir = expanded
	} else {
		resolved, err := resolveTemplatesDir(cmd)
		if err != nil {
			return "", err
		}
		dir = resolved
		if dir == "" {
			dir = resolveDefaultTemplatesDir()
		}
	}
	info, err := os.Stat(dir)
	if err != nil {
		return "", fmt.Errorf("templates dir %s: %w", dir, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", dir)
	}
	return dir, nil
}

func gitInit(cmd *cobra.Command, dir string) error {
	if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
		// already a git repo, leave alone
		return nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	git := exec.CommandContext(cmd.Context(), "git", "init", "--initial-branch=main", dir)
	git.Stdout = cmd.OutOrStdout()
	git.Stderr = cmd.ErrOrStderr()
	if err := git.Run(); err != nil {
		return fmt.Errorf("git init: %w", err)
	}
	return nil
}

// snapshotSubdir is the sidecar directory inside the user's templates dir
// that holds a byte-for-byte copy of the embedded template content as of
// the last init or upgrade. It serves as the merge-base for 3-way upgrades.
const snapshotSubdir = ".srekit-embedded"

func snapshotPath(userDir, name string) string {
	return filepath.Join(userDir, snapshotSubdir, name)
}

func readSnapshot(userDir, name string) ([]byte, error) {
	return os.ReadFile(snapshotPath(userDir, name))
}

func writeSnapshot(userDir, name string, body []byte) error {
	snapDir := filepath.Join(userDir, snapshotSubdir)
	if err := os.MkdirAll(snapDir, 0o755); err != nil {
		return fmt.Errorf("create snapshot dir: %w", err)
	}
	// Snapshot files are internal scaffolding — same 0o644 rationale.
	if err := os.WriteFile(snapshotPath(userDir, name), body, 0o644); err != nil { //nolint:gosec // G306: see runTemplatesInit
		return fmt.Errorf("write snapshot %s: %w", name, err)
	}
	return nil
}

// ensureSnapshotIgnored appends `.srekit-embedded/` to the user dir's
// .gitignore on a best-effort basis so the sidecar doesn't leak into
// the user's git history. Idempotent.
func ensureSnapshotIgnored(userDir string) error {
	gitignorePath := filepath.Join(userDir, ".gitignore")
	body, err := os.ReadFile(gitignorePath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	entry := snapshotSubdir + "/"
	for _, line := range strings.Split(string(body), "\n") {
		if strings.TrimSpace(line) == entry || strings.TrimSpace(line) == snapshotSubdir {
			return nil
		}
	}
	if len(body) > 0 && !bytes.HasSuffix(body, []byte("\n")) {
		body = append(body, '\n')
	}
	body = append(body, []byte(entry+"\n")...)
	return os.WriteFile(gitignorePath, body, 0o644) //nolint:gosec // G306: same rationale as templates init
}

// threeWayMerge runs `git merge-file -p --diff3` on the three buffers
// via tempfiles and returns the merged body along with the conflict count.
// A non-zero conflict count is not an error — the caller writes the body
// (with markers) and decides how to report.
func threeWayMerge(ctx context.Context, name string, user, base, embedded []byte) ([]byte, int, error) {
	dir, err := os.MkdirTemp("", "srekit-merge-")
	if err != nil {
		return nil, 0, fmt.Errorf("create tempdir: %w", err)
	}
	defer os.RemoveAll(dir)

	paths := make(map[string]string, 3)
	for label, body := range map[string][]byte{"user": user, "base": base, "embedded": embedded} {
		p := filepath.Join(dir, label)
		if err := os.WriteFile(p, body, 0o644); err != nil { //nolint:gosec // G306: tempfile, lifecycle == this function
			return nil, 0, fmt.Errorf("write tempfile: %w", err)
		}
		paths[label] = p
	}

	// git merge-file <current> <base> <other> — produces "current with
	// other merged in, base as ancestor". Order matters for conflict
	// marker labeling. The three path args are tempfiles created inside
	// this function from MkdirTemp + Join — no user-controlled strings.
	gitCmd := exec.CommandContext(ctx, "git", "merge-file", "-p", "--diff3", //nolint:gosec // G204: args are tempfile paths we just created
		"-L", "user", "-L", "base", "-L", "embedded",
		paths["user"], paths["base"], paths["embedded"])
	var stdout, stderr bytes.Buffer
	gitCmd.Stdout = &stdout
	gitCmd.Stderr = &stderr
	err = gitCmd.Run()
	var exitErr *exec.ExitError
	switch {
	case err == nil:
		return stdout.Bytes(), 0, nil
	case errors.As(err, &exitErr):
		// Exit code = conflict count (per git merge-file man page).
		return stdout.Bytes(), exitErr.ExitCode(), nil
	default:
		return nil, 0, fmt.Errorf("git merge-file %s: %w (stderr: %s)", name, err, stderr.String())
	}
}
