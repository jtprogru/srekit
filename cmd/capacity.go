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

type capacityMeta struct {
	ID      string `json:"id"`
	Service string `json:"service"`
	Horizon string `json:"horizon"`
	Now     string `json:"now"`
}

type capacityData struct {
	Meta     capacityMeta               `json:"meta"`
	Sections []sections.RenderedSection `json:"sections"`
}

func (d capacityData) ArtifactPayload() ([]sections.RenderedSection, any) {
	return d.Sections, struct{ Meta capacityMeta }{Meta: d.Meta}
}

func newCapacityCmd() *cobra.Command {
	var (
		service string
		horizon string
		out     cliflags.Output
	)
	cmd := &cobra.Command{
		Use:   "capacity",
		Short: "Generate a capacity plan template",
		Long:  "Generate a capacity-planning document: current capacity, growth assumptions, forecast, scale-up triggers, risks.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if service == "" {
				return errors.New("--service is required")
			}
			meta := capacityMeta{
				ID:      ids.UUID(),
				Service: service,
				Horizon: horizon,
				Now:     clock.Now().Format(time.RFC3339),
			}

			loader := loaderFrom(cmd)
			manifest, err := loadCapacityManifest(cmd, loader)
			if err != nil {
				return err
			}
			rendered, err := sections.Merge(manifest, nil, struct{ Meta capacityMeta }{Meta: meta})
			if err != nil {
				return err
			}

			data := capacityData{Meta: meta, Sections: rendered}
			def := fmt.Sprintf("capacity-%s.md", ids.Slug(service))
			opts := out.RenderOptionsStructured(cmd, def)
			opts.RenderArtifact = true
			return render.Render(cmd.OutOrStdout(), loader, "capacity.md.tmpl", data, opts)
		},
	}
	cmd.Flags().StringVar(&service, "service", "", "service name (required)")
	cmd.Flags().StringVar(&horizon, "horizon", "1y", "planning horizon (e.g. 6m, 1y, 2y)")
	out.Bind(cmd, "write to file (default: capacity-<service>.md)")
	return cmd
}

func loadCapacityManifest(cmd *cobra.Command, loader *tmpl.Loader) (*sections.Manifest, error) {
	artifactBytes, err := loader.LoadArtifactBytes("capacity.md.tmpl")
	if err != nil {
		return nil, fmt.Errorf("load capacity.yaml: %w", err)
	}
	a, err := sections.ParseArtifact(artifactBytes)
	if err != nil {
		return nil, fmt.Errorf("parse capacity.yaml: %w", err)
	}
	warnStaleLegacyFiles(cmd, loader, "capacity.yaml", "capacity.md.tmpl")
	return &sections.Manifest{Version: a.Version, Sections: a.Sections}, nil
}
