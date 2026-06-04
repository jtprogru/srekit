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

type sloMeta struct {
	ID            string `json:"id"`
	Service       string `json:"service"`
	Target        string `json:"target"`
	Window        string `json:"window"`
	LatencyTarget string `json:"latencyTarget"`
	Now           string `json:"now"`
}

type sloData struct {
	Meta     sloMeta                    `json:"meta"`
	Sections []sections.RenderedSection `json:"sections"`
}

func (d sloData) ArtifactPayload() ([]sections.RenderedSection, any) {
	return d.Sections, struct{ Meta sloMeta }{Meta: d.Meta}
}

func newSLOCmd() *cobra.Command {
	var (
		service       string
		target        string
		window        string
		latencyTarget string
		out           cliflags.Output
	)
	cmd := &cobra.Command{
		Use:   "slo",
		Short: "Generate an SLO/SLI document template",
		Example: `  # SLO doc with custom availability and latency targets
  srekit slo --service api-gw --target 99.95% --latency 250ms --window 28d`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if service == "" {
				return errors.New("--service is required")
			}
			meta := sloMeta{
				ID:            ids.UUID(),
				Service:       service,
				Target:        target,
				Window:        window,
				LatencyTarget: latencyTarget,
				Now:           clock.Now().Format(time.RFC3339),
			}

			loader := loaderFrom(cmd)
			manifest, err := loadSLOManifest(cmd, loader)
			if err != nil {
				return err
			}
			rendered, err := sections.Merge(manifest, nil, struct{ Meta sloMeta }{Meta: meta})
			if err != nil {
				return err
			}

			data := sloData{Meta: meta, Sections: rendered}
			def := fmt.Sprintf("slo-%s.md", ids.Slug(service))
			opts := out.RenderOptionsStructured(cmd, def)
			opts.RenderArtifact = true
			return render.Render(cmd.OutOrStdout(), loader, "slo.md.tmpl", data, opts)
		},
	}
	cmd.Flags().StringVar(&service, "service", "", "service name (required)")
	cmd.Flags().StringVar(&target, "target", "99.9%", "availability target")
	cmd.Flags().StringVar(&window, "window", "30d", "rolling SLO window")
	cmd.Flags().StringVar(&latencyTarget, "latency", "300ms", "p99 latency target")
	out.Bind(cmd, "write to file (default: slo-<service>.md)")
	return cmd
}

func loadSLOManifest(cmd *cobra.Command, loader *tmpl.Loader) (*sections.Manifest, error) {
	artifactBytes, err := loader.LoadArtifactBytes("slo.md.tmpl")
	if err != nil {
		return nil, fmt.Errorf("load slo.yaml: %w", err)
	}
	a, err := sections.ParseArtifact(artifactBytes)
	if err != nil {
		return nil, fmt.Errorf("parse slo.yaml: %w", err)
	}
	warnStaleLegacyFiles(cmd, loader, "slo.yaml", "slo.md.tmpl")
	return &sections.Manifest{Version: a.Version, Sections: a.Sections}, nil
}
