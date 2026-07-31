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

type ebpMeta struct {
	ID      string `json:"id"`
	Service string `json:"service"`
	Now     string `json:"now"`
}

type ebpData struct {
	Meta     ebpMeta                    `json:"meta"`
	Sections []sections.RenderedSection `json:"sections"`
}

func (d ebpData) ArtifactPayload() ([]sections.RenderedSection, any) {
	return d.Sections, struct{ Meta ebpMeta }{Meta: d.Meta}
}

func newEBPCmd() *cobra.Command {
	var (
		service string
		out     cliflags.Output
	)
	cmd := &cobra.Command{
		Use:   "ebp",
		Short: "Generate an error budget policy template",
		Long:  "Generate an Error Budget Policy: the document that says what happens when the budget burns — tiered actions, exceptions, escalation.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if service == "" {
				return errors.New("--service is required")
			}
			meta := ebpMeta{
				ID:      ids.UUID(),
				Service: service,
				Now:     clock.Now().Format(time.RFC3339),
			}

			loader := loaderFrom(cmd)
			manifest, err := loadEBPManifest(cmd, loader)
			if err != nil {
				return err
			}
			rendered, err := sections.Merge(manifest, nil, struct{ Meta ebpMeta }{Meta: meta})
			if err != nil {
				return err
			}

			data := ebpData{Meta: meta, Sections: rendered}
			def := fmt.Sprintf("ebp-%s.md", ids.Slug(service))
			opts := out.RenderOptions(cmd, def)
			return render.Render(cmd.OutOrStdout(), loader, "ebp", data, opts)
		},
	}
	cmd.Flags().StringVar(&service, "service", "", "service name (required)")
	out.Bind(cmd, "write to file (default: ebp-<service>.md)")
	return cmd
}

func loadEBPManifest(cmd *cobra.Command, loader *tmpl.Loader) (*sections.Manifest, error) {
	artifactBytes, err := loader.LoadArtifactBytes("ebp")
	if err != nil {
		return nil, fmt.Errorf("load ebp.yaml: %w", err)
	}
	a, err := sections.ParseArtifact(artifactBytes)
	if err != nil {
		return nil, fmt.Errorf("parse ebp.yaml: %w", err)
	}
	warnStaleLegacyFiles(cmd, loader, "ebp.yaml", "ebp.md.tmpl")
	return &sections.Manifest{Version: a.Version, Sections: a.Sections}, nil
}
