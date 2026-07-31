package cmd

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/jtprogru/srekit/internal/cliflags"
	"github.com/jtprogru/srekit/internal/clock"
	"github.com/jtprogru/srekit/internal/config"
	"github.com/jtprogru/srekit/internal/ids"
	"github.com/jtprogru/srekit/internal/meta"
	"github.com/jtprogru/srekit/internal/render"
	"github.com/jtprogru/srekit/internal/sections"
	"github.com/jtprogru/srekit/internal/tmpl"
)

type oncallMeta struct {
	ID     string      `json:"id"`
	Team   string      `json:"team"`
	Start  string      `json:"start"`
	End    string      `json:"end"`
	Now    string      `json:"now"`
	Author meta.Author `json:"author"`
}

type oncallData struct {
	Meta     oncallMeta                 `json:"meta"`
	Sections []sections.RenderedSection `json:"sections"`
}

func (d oncallData) ArtifactPayload() ([]sections.RenderedSection, any) {
	return d.Sections, struct{ Meta oncallMeta }{Meta: d.Meta}
}

func newOncallCmd() *cobra.Command {
	var (
		team   string
		start  string
		end    string
		author string
		email  string
		out    cliflags.Output
	)
	cmd := &cobra.Command{
		Use:     "oncall-report",
		Aliases: []string{"oncall"},
		Short:   "Generate a weekly on-call report template",
		Example: `  # This week's report for a team (period defaults to Mon–Sun)
  srekit oncall-report --team platform

  # Explicit period
  srekit oncall-report --team platform --start 2026-05-18 --end 2026-05-24`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if team == "" {
				return errors.New("--team is required")
			}
			a, err := meta.Resolve(config.Global(), author, email)
			if err != nil {
				return err
			}
			now := clock.Now()
			if start == "" || end == "" {
				wd := int(now.Weekday())
				if wd == 0 {
					wd = 7
				}
				weekStart := now.AddDate(0, 0, -wd+1)
				weekEnd := weekStart.AddDate(0, 0, 6)
				if start == "" {
					start = weekStart.Format("2006-01-02")
				}
				if end == "" {
					end = weekEnd.Format("2006-01-02")
				}
			}
			m := oncallMeta{
				ID:     ids.UUID(),
				Team:   team,
				Start:  start,
				End:    end,
				Now:    now.Format(time.RFC3339),
				Author: a,
			}

			loader := loaderFrom(cmd)
			manifest, err := loadOncallManifest(cmd, loader)
			if err != nil {
				return err
			}
			rendered, err := sections.Merge(manifest, nil, struct{ Meta oncallMeta }{Meta: m})
			if err != nil {
				return err
			}

			data := oncallData{Meta: m, Sections: rendered}
			def := fmt.Sprintf("oncall-%s-%s.md", ids.Slug(team), ids.Slug(start))
			opts := out.RenderOptions(cmd, def)
			return render.Render(cmd.OutOrStdout(), loader, "oncall", data, opts)
		},
	}
	cmd.Flags().StringVar(&team, "team", "", "team name (required)")
	cmd.Flags().StringVar(&start, "start", "", "period start (default: this week's Monday)")
	cmd.Flags().StringVar(&end, "end", "", "period end (default: this week's Sunday)")
	cmd.Flags().StringVar(&author, "author", "", "on-caller name")
	cmd.Flags().StringVar(&email, "email", "", "on-caller email")
	out.Bind(cmd, "write to file (default: oncall-<team>-<start>.md)")
	return cmd
}

func loadOncallManifest(cmd *cobra.Command, loader *tmpl.Loader) (*sections.Manifest, error) {
	artifactBytes, err := loader.LoadArtifactBytes("oncall")
	if err != nil {
		return nil, fmt.Errorf("load oncall.yaml: %w", err)
	}
	a, err := sections.ParseArtifact(artifactBytes)
	if err != nil {
		return nil, fmt.Errorf("parse oncall.yaml: %w", err)
	}
	warnStaleLegacyFiles(cmd, loader, "oncall.yaml", "oncall.md.tmpl")
	return &sections.Manifest{Version: a.Version, Sections: a.Sections}, nil
}
