package cmd

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/jtprogru/srekit/internal/meta"
	"github.com/jtprogru/srekit/internal/render"
)

var (
	chRepo    string
	chVersion string
	chOut     string
	chStdout  bool
	chForce   bool
	chDry     bool
)

var changelogCmd = &cobra.Command{
	Use:   "changelog",
	Short: "Generate a CHANGELOG.md scaffold (Keep a Changelog format)",
	RunE: func(cmd *cobra.Command, _ []string) error {
		repo := chRepo
		if repo == "" {
			r, err := meta.DetectRepo()
			if err != nil {
				repo = "OWNER/REPO"
				cmd.PrintErrf("warning: could not detect repo from git remote (%v); using placeholder %q. Pass --repo to override.\n", err, repo)
			} else {
				repo = r.Slug()
			}
		}
		data := struct {
			Today          string
			Repo           string
			InitialVersion string
		}{
			Today:          time.Now().Format("2006-01-02"),
			Repo:           repo,
			InitialVersion: chVersion,
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
	changelogCmd.Flags().StringVar(&chRepo, "repo", "", "OWNER/REPO for compare links (default: detect from git remote)")
	changelogCmd.Flags().StringVar(&chVersion, "version", "0.1.0", "initial version label")
	changelogCmd.Flags().StringVar(&chOut, "out", "", "write to file (default: CHANGELOG.md)")
	changelogCmd.Flags().BoolVar(&chStdout, "stdout", false, "print to stdout")
	changelogCmd.Flags().BoolVar(&chForce, "force", false, "overwrite existing CHANGELOG.md")
	changelogCmd.Flags().BoolVar(&chDry, "dry-run", false, "print result, do not write a file")
	rootCmd.AddCommand(changelogCmd)
}
