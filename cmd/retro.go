package cmd

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/jtprogru/srekit/internal/cliflags"
	"github.com/jtprogru/srekit/internal/clock"
	"github.com/jtprogru/srekit/internal/ids"
	"github.com/jtprogru/srekit/internal/render"
	"github.com/jtprogru/srekit/internal/sections"
	"github.com/jtprogru/srekit/internal/tmpl"
)

type retroMeta struct {
	ID     string `json:"id"`
	Team   string `json:"team"`
	Sprint string `json:"sprint"`
	Now    string `json:"now"`
}

type retroData struct {
	Meta     retroMeta                  `json:"meta"`
	Sections []sections.RenderedSection `json:"sections"`
}

func (d retroData) ArtifactPayload() ([]sections.RenderedSection, any) {
	return d.Sections, struct{ Meta retroMeta }{Meta: d.Meta}
}

func newRetroCmd() *cobra.Command {
	var (
		team   string
		sprint string
		out    cliflags.Output
	)
	cmd := &cobra.Command{
		Use:   "retro",
		Short: "Generate a sprint retrospective template",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if team == "" {
				return errors.New("--team is required")
			}
			s := sprint
			if s == "" {
				s = clock.Now().Format("2006-01-02")
			}
			meta := retroMeta{
				ID:     ids.UUID(),
				Team:   team,
				Sprint: s,
				Now:    clock.Now().Format(time.RFC3339),
			}

			loader := loaderFrom(cmd)
			manifest, err := loadRetroManifest(cmd, loader)
			if err != nil {
				return err
			}
			rendered, err := sections.Merge(manifest, nil, struct{ Meta retroMeta }{Meta: meta})
			if err != nil {
				return err
			}

			data := retroData{Meta: meta, Sections: rendered}
			def := fmt.Sprintf("retro-%s-%s.md", ids.Slug(team), ids.Slug(s))
			opts := out.RenderOptionsStructured(cmd, def)
			opts.RenderArtifact = true
			return render.Render(cmd.OutOrStdout(), loader, "retro.md.tmpl", data, opts)
		},
	}
	cmd.Flags().StringVar(&team, "team", "", "team name (required)")
	cmd.Flags().StringVar(&sprint, "sprint", "", "sprint or period label (default: today's date)")
	out.Bind(cmd, "write to file (default: retro-<team>-<sprint>.md)")
	return cmd
}

func loadRetroManifest(cmd *cobra.Command, loader *tmpl.Loader) (*sections.Manifest, error) {
	artifactBytes, err := loader.LoadArtifactBytes("retro.md.tmpl")
	if err != nil {
		return nil, fmt.Errorf("load retro.yaml: %w", err)
	}
	a, err := sections.ParseArtifact(artifactBytes)
	if err != nil {
		return nil, fmt.Errorf("parse retro.yaml: %w", err)
	}
	warnStaleLegacyFiles(cmd, loader, "retro.yaml", "retro.md.tmpl")
	return &sections.Manifest{Version: a.Version, Sections: a.Sections}, nil
}
