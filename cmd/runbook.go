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

type runbookMeta struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Service string `json:"service"`
	Alert   string `json:"alert"`
	Now     string `json:"now"`
}

type runbookData struct {
	Meta     runbookMeta                `json:"meta"`
	Sections []sections.RenderedSection `json:"sections"`
}

func (d runbookData) ArtifactPayload() ([]sections.RenderedSection, any) {
	return d.Sections, struct{ Meta runbookMeta }{Meta: d.Meta}
}

func newRunbookCmd() *cobra.Command {
	var (
		title   string
		service string
		alert   string
		out     cliflags.Output
	)
	cmd := &cobra.Command{
		Use:   "runbook",
		Short: "Generate an on-call runbook template",
		Long:  "Generate a runbook with Symptoms / Diagnose / Mitigate / Verify sections.",
		Example: `  # Runbook tied to a service and alert
  srekit runbook --title "High p99 latency" --service api-gw --alert HighLatency`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if title == "" {
				return errors.New("--title is required")
			}
			meta := runbookMeta{
				ID:      ids.UUID(),
				Title:   title,
				Service: service,
				Alert:   alert,
				Now:     clock.Now().Format(time.RFC3339),
			}

			loader := loaderFrom(cmd)
			manifest, err := loadRunbookManifest(cmd, loader)
			if err != nil {
				return err
			}
			rendered, err := sections.Merge(manifest, nil, struct{ Meta runbookMeta }{Meta: meta})
			if err != nil {
				return err
			}

			data := runbookData{Meta: meta, Sections: rendered}
			def := fmt.Sprintf("runbook-%s.md", ids.Slug(title))
			opts := out.RenderOptions(cmd, def)
			return render.Render(cmd.OutOrStdout(), loader, "runbook", data, opts)
		},
	}
	cmd.Flags().StringVarP(&title, "title", "T", "", "runbook title (required)")
	cmd.Flags().StringVar(&service, "service", "", "service name")
	cmd.Flags().StringVar(&alert, "alert", "", "alert name this runbook responds to")
	out.Bind(cmd, "write to file (default: runbook-<slug>.md)")
	return cmd
}

func loadRunbookManifest(cmd *cobra.Command, loader *tmpl.Loader) (*sections.Manifest, error) {
	artifactBytes, err := loader.LoadArtifactBytes("runbook")
	if err != nil {
		return nil, fmt.Errorf("load runbook.yaml: %w", err)
	}
	a, err := sections.ParseArtifact(artifactBytes)
	if err != nil {
		return nil, fmt.Errorf("parse runbook.yaml: %w", err)
	}
	warnStaleLegacyFiles(cmd, loader, "runbook.yaml", "runbook.md.tmpl")
	return &sections.Manifest{Version: a.Version, Sections: a.Sections}, nil
}
