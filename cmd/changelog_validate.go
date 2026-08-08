package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jtprogru/srekit/internal/changelog"
)

func newChangelogValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate [FILE]",
		Short: "Check an existing CHANGELOG.md against Keep a Changelog",
		Long: `Report, per check, where a changelog departs from Keep a Changelog:

  heading-shape          every version heading is ` + "`[X.Y.Z] - YYYY-MM-DD`" + `, optionally ` + "` [YANKED]`" + `
  unreleased-section     an [Unreleased] section is present and precedes every release
  version-order          released versions appear in descending version order
  no-duplicate-versions  no version appears twice
  change-types           every ### subsection is one of the six change types
  link-definitions       every version heading has a matching link definition

Reports every check, then exits non-zero if any failed. It never rewrites
the file.`,
		Example: `  # Check CHANGELOG.md in the working directory
  srekit changelog validate

  # Check one somewhere else
  srekit changelog validate docs/CHANGELOG.md`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := defaultChangelogPath
			if len(args) == 1 {
				path = args[0]
			}
			src, err := os.ReadFile(path)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return fmt.Errorf("%s does not exist; run `srekit changelog` to create one", path)
				}
				return err
			}

			vocabs := changelog.Vocabularies()
			results := changelog.Validate(changelog.Scan(src, vocabs), vocabs)

			out := cmd.OutOrStdout()
			var failed int
			for _, r := range results {
				if r.OK {
					fmt.Fprintf(out, "OK    %s\n", r.Name)
					continue
				}
				fmt.Fprintf(out, "FAIL  %s: %s\n", r.Name, r.Detail)
				failed++
			}
			if failed > 0 {
				return fmt.Errorf("%d of %d check(s) failed in %s", failed, len(results), path)
			}
			return nil
		},
	}
	return cmd
}
