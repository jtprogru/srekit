package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jtprogru/srekit/internal/cliflags"
	"github.com/jtprogru/srekit/internal/clock"
	"github.com/jtprogru/srekit/internal/meta"
	"github.com/jtprogru/srekit/internal/render"
)

func newChangelogCmd() *cobra.Command {
	var (
		repoFlag string
		version  string
		out      cliflags.Output
	)
	cmd := &cobra.Command{
		Use:   "changelog",
		Short: "Generate a CHANGELOG.md scaffold (Keep a Changelog format)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			repo := repoFlag
			if repo == "" {
				r, err := meta.DetectRepo()
				if err != nil {
					return fmt.Errorf("could not detect repo from git remote: %w (pass --repo OWNER/REPO)", err)
				}
				repo = r.Slug()
			}
			data := struct {
				Today          string
				Repo           string
				InitialVersion string
			}{
				Today:          clock.Now().Format("2006-01-02"),
				Repo:           repo,
				InitialVersion: version,
			}
			return render.Render(cmd.OutOrStdout(), "changelog.md.tmpl", data, out.RenderOptions("CHANGELOG.md"))
		},
	}
	cmd.Flags().StringVar(&repoFlag, "repo", "", "OWNER/REPO for compare links (default: detect from git remote)")
	cmd.Flags().StringVar(&version, "version", "0.1.0", "initial version label")
	out.Bind(cmd, "write to file (default: CHANGELOG.md)")
	return cmd
}
