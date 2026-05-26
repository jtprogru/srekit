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
		Long:  "Scaffold and (later) sync a custom templates directory that overrides the embedded templates.",
	}
	c.AddCommand(newTemplatesInitCmd())
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
