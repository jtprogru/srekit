package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jtprogru/srekit/internal/migrate"
)

func newTemplatesMigrateCmd() *cobra.Command {
	var apply bool
	cmd := &cobra.Command{
		Use:   "migrate [dir]",
		Short: "Convert pre-v0.14.0 .tmpl files to the v1 <name>.yaml artifact format",
		Long: `Best-effort converter that turns legacy .tmpl files (and their optional
.sections.yaml sidecars) into the v1 single-file <name>.yaml artifact
introduced in v0.14.0.

The converter parses the .tmpl's frontmatter, H1, meta bullets, and ##
section blocks, infers each section's type (table when a GFM table is
present; text otherwise), and emits a yaml that round-trips through the
v1 loader. Sections containing Go-template control flow ({{ if }} /
{{ range }} / {{ with }}) are wrapped in git merge-style diff markers so
a human can rewrite them deterministically — the converter doesn't try
to translate template logic into the typed-sections vocabulary.

When a sibling <name>.sections.yaml is present, its section list takes
precedence over the heuristic-parsed sections from the .tmpl — the .tmpl
is then only consulted for the header (frontmatter / H1 / meta bullets).

Defaults to --dry-run: prints what each <name>.yaml would contain. Pass
--apply to write files. Original .tmpl / .sections.yaml are left in
place so you can compare and delete them when ready.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTemplatesMigrate(cmd, args, apply)
		},
	}
	cmd.Flags().BoolVar(&apply, "apply", false, "write the converted <name>.yaml files (default: dry-run — print only)")
	return cmd
}

func runTemplatesMigrate(cmd *cobra.Command, args []string, apply bool) error {
	dir, err := pickTemplatesDir(cmd, args)
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read %s: %w", dir, err)
	}
	var candidates []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".tmpl") {
			continue
		}
		if !strings.HasSuffix(name, ".md.tmpl") {
			// license_*.tmpl and similar — body is too freeform / static
			// for the converter; users shouldn't migrate them. The license
			// command itself uses inlined Go consts since v0.14.0.
			continue
		}
		candidates = append(candidates, name)
	}
	sort.Strings(candidates)

	if len(candidates) == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "no *.md.tmpl files to migrate in %s\n", dir)
		return nil
	}

	out := cmd.OutOrStdout()
	errOut := cmd.ErrOrStderr()
	var converted, skipped, failed int
	for _, name := range candidates {
		yamlName := strings.TrimSuffix(strings.TrimSuffix(name, ".tmpl"), ".md") + ".yaml"
		yamlPath := filepath.Join(dir, yamlName)
		tmplPath := filepath.Join(dir, name)
		sectionsPath := filepath.Join(dir, strings.TrimSuffix(strings.TrimSuffix(name, ".tmpl"), ".md")+".sections.yaml")

		if _, err := os.Stat(yamlPath); err == nil {
			fmt.Fprintf(out, "skip   %s (target %s already exists)\n", name, yamlName)
			skipped++
			continue
		}

		tmplBody, err := os.ReadFile(tmplPath)
		if err != nil {
			fmt.Fprintf(errOut, "fail   %s: read .tmpl: %v\n", name, err)
			failed++
			continue
		}
		var sidecar []byte
		if b, err := os.ReadFile(sectionsPath); err == nil {
			sidecar = b
		} else if !errors.Is(err, fs.ErrNotExist) {
			fmt.Fprintf(errOut, "fail   %s: read sidecar: %v\n", name, err)
			failed++
			continue
		}

		yamlBody, err := migrate.Convert(tmplBody, sidecar)
		if err != nil {
			fmt.Fprintf(errOut, "fail   %s: %v\n", name, err)
			failed++
			continue
		}

		hasMarkers := strings.Contains(string(yamlBody), "<<<<<<< srekit migrate")
		marker := "ok"
		if hasMarkers {
			marker = "ok (with diff markers — review needed)"
		}

		if apply {
			if err := os.WriteFile(yamlPath, yamlBody, 0o644); err != nil { //nolint:gosec // G306: public artifact
				fmt.Fprintf(errOut, "fail   %s: write %s: %v\n", name, yamlName, err)
				failed++
				continue
			}
			fmt.Fprintf(out, "wrote  %s → %s [%s]\n", name, yamlName, marker)
		} else {
			fmt.Fprintf(out, "would  %s → %s [%s]\n", name, yamlName, marker)
			fmt.Fprintf(out, "---\n%s---\n", string(yamlBody))
		}
		converted++
	}

	mode := "dry-run"
	if apply {
		mode = "summary"
	}
	fmt.Fprintf(out, "\n%s: %d converted, %d skipped, %d failed\n", mode, converted, skipped, failed)
	if failed > 0 {
		return fmt.Errorf("%d file(s) failed to migrate", failed)
	}
	if !apply && converted > 0 {
		fmt.Fprintf(errOut, "\nrun with --apply to write the .yaml file(s); original .tmpl + .sections.yaml are left in place for you to remove after verification.\n")
	}
	return nil
}
