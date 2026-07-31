package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jtprogru/srekit/internal/cliflags"
	"github.com/jtprogru/srekit/internal/clock"
	"github.com/jtprogru/srekit/internal/meta"
	"github.com/jtprogru/srekit/internal/render"
	"github.com/jtprogru/srekit/internal/sections"
	"github.com/jtprogru/srekit/internal/tmpl"
)

type changelogMeta struct {
	Today          string `json:"today"`
	Repo           string `json:"repo"`
	InitialVersion string `json:"initialVersion"`
}

type changelogData struct {
	Meta     changelogMeta              `json:"meta"`
	Sections []sections.RenderedSection `json:"sections"`
}

func (d changelogData) ArtifactPayload() ([]sections.RenderedSection, any) {
	return d.Sections, struct{ Meta changelogMeta }{Meta: d.Meta}
}

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
			m := changelogMeta{
				Today:          clock.Now().Format("2006-01-02"),
				Repo:           repo,
				InitialVersion: version,
			}

			loader := loaderFrom(cmd)
			manifest, err := loadChangelogManifest(cmd, loader)
			if err != nil {
				return err
			}
			rendered, err := sections.Merge(manifest, nil, struct{ Meta changelogMeta }{Meta: m})
			if err != nil {
				return err
			}

			data := changelogData{Meta: m, Sections: rendered}
			opts := out.RenderOptions(cmd, "CHANGELOG.md")
			return render.Render(cmd.OutOrStdout(), loader, "changelog", data, opts)
		},
	}
	cmd.Flags().StringVar(&repoFlag, "repo", "", "OWNER/REPO for compare links (default: detect from git remote)")
	cmd.Flags().StringVar(&version, "version", "0.1.0", "initial version label")
	out.Bind(cmd, "write to file (default: CHANGELOG.md)")
	return cmd
}

func loadChangelogManifest(cmd *cobra.Command, loader *tmpl.Loader) (*sections.Manifest, error) {
	artifactBytes, err := loader.LoadArtifactBytes("changelog")
	if err != nil {
		return nil, fmt.Errorf("load changelog.yaml: %w", err)
	}
	a, err := sections.ParseArtifact(artifactBytes)
	if err != nil {
		return nil, fmt.Errorf("parse changelog.yaml: %w", err)
	}
	warnStaleLegacyFiles(cmd, loader, "changelog.yaml", "changelog.md.tmpl")
	return &sections.Manifest{Version: a.Version, Sections: a.Sections}, nil
}
