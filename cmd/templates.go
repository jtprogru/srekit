package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jtprogru/srekit/internal/tmpl"
)

const defaultTemplatesDir = "~/.srekit/templates"

func newTemplatesCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "templates",
		Short: "Manage the custom templates directory",
		Long:  "Scaffold or sync a custom templates directory that overrides the embedded templates.",
	}
	c.AddCommand(newTemplatesInitCmd())
	c.AddCommand(newTemplatesPullCmd())
	return c
}

func newTemplatesInitCmd() *cobra.Command {
	var (
		force bool
		noGit bool
	)
	cmd := &cobra.Command{
		Use:   "init [dir]",
		Short: "Scaffold a custom templates directory from the built-in templates",
		Long: `Copy every built-in template into <dir> (default: ~/.srekit/templates),
write TEMPLATES.md with the placeholder/FuncMap reference, and run
git init in the directory (unless --no-git). Refuses to overwrite
existing files unless --force is set.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := defaultTemplatesDir
			if len(args) == 1 {
				dir = args[0]
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

	entries, err := fs.ReadDir(tmpl.FS, "templates")
	if err != nil {
		return fmt.Errorf("read embedded templates: %w", err)
	}

	written := 0
	for _, e := range entries {
		target := filepath.Join(dir, e.Name())
		if _, err := os.Stat(target); err == nil && !force {
			return fmt.Errorf("%s already exists; pass --force to overwrite", target)
		}
		b, err := fs.ReadFile(tmpl.FS, "templates/"+e.Name())
		if err != nil {
			return err
		}
		// Templates are public scaffolding — 0o644 matches render.go's convention.
		if err := os.WriteFile(target, b, 0o644); err != nil { //nolint:gosec // G306: same rationale as internal/render
			return fmt.Errorf("write %s: %w", target, err)
		}
		written++
	}

	// TEMPLATES.md is a reference doc — we always overwrite it on init so it
	// stays in sync with the binary's understanding of placeholders/FuncMap.
	docsTarget := filepath.Join(dir, "TEMPLATES.md")
	if err := os.WriteFile(docsTarget, tmpl.DocsMD, 0o644); err != nil { //nolint:gosec // G306: same rationale as internal/render
		return fmt.Errorf("write %s: %w", docsTarget, err)
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
	fmt.Fprintf(out, "  echo 'templates_dir: %s' >> ~/.srekit.yaml\n", dir)
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
		dir, err = expandHome(defaultTemplatesDir)
		if err != nil {
			return err
		}
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
