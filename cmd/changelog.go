package cmd

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/jtprogru/srekit/internal/meta"
	"github.com/jtprogru/srekit/internal/render"
)

var (
	chAuthor string
	chEmail  string
	chOut    string
	chStdout bool
	chForce  bool
	chDry    bool
)

var changelogCmd = &cobra.Command{
	Use:   "changelog",
	Short: "Generate a CHANGELOG.md scaffold (Keep a Changelog format)",
	RunE: func(cmd *cobra.Command, _ []string) error {
		author, err := meta.Resolve(chAuthor, chEmail)
		if err != nil {
			return err
		}
		data := struct {
			Today  string
			Author meta.Author
		}{
			Today:  time.Now().Format("2006-01-02"),
			Author: author,
		}
		return render.Render(cmd.OutOrStdout(), "changelog.md.tmpl", data, render.Options{
			Out:     chOut,
			Stdout:  chStdout,
			Force:   chForce,
			DryRun:  chDry,
			Default: "CHANGELOG.md",
		})
	},
}

func init() {
	changelogCmd.Flags().StringVar(&chAuthor, "author", "", "GitHub username or full name (used in compare links)")
	changelogCmd.Flags().StringVar(&chEmail, "email", "", "author email")
	changelogCmd.Flags().StringVar(&chOut, "out", "", "write to file (default: CHANGELOG.md)")
	changelogCmd.Flags().BoolVar(&chStdout, "stdout", false, "print to stdout")
	changelogCmd.Flags().BoolVar(&chForce, "force", false, "overwrite existing CHANGELOG.md")
	changelogCmd.Flags().BoolVar(&chDry, "dry-run", false, "print result, do not write a file")
	rootCmd.AddCommand(changelogCmd)
}
