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

var allowedIncidentStatus = map[string]bool{
	"investigated": true,
	"active":       true,
	"contained":    true,
	"resolved":     true,
}

type incidentMeta struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Severity string `json:"severity"`
	Lead     string `json:"lead"`
	Status   string `json:"status"`
	Now      string `json:"now"`
}

type incidentData struct {
	Meta     incidentMeta               `json:"meta"`
	Sections []sections.RenderedSection `json:"sections"`
}

func (d incidentData) ArtifactPayload() ([]sections.RenderedSection, any) {
	return d.Sections, struct{ Meta incidentMeta }{Meta: d.Meta}
}

func newIncidentCmd() *cobra.Command {
	var (
		title    string
		severity string
		lead     string
		status   string
		out      cliflags.Output
	)
	cmd := &cobra.Command{
		Use:   "incident",
		Short: "Generate a live-incident report template",
		Long:  "Generate a live-incident report (status, lead, comms, updates log) — to be filled during the incident, distinct from a postmortem.",
		Example: `  # Active SEV-1, printed to stdout
  srekit incident --title "Checkout 5xx spike" --severity SEV-1 --lead alice --stdout

  # Write to the default file (incident-<slug>.md)
  srekit incident -T "DB failover" --status investigated`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if title == "" {
				return errors.New("--title is required")
			}
			if !allowedIncidentStatus[status] {
				return fmt.Errorf("invalid --status %q (investigated, active, contained, resolved)", status)
			}
			meta := incidentMeta{
				ID:       ids.UUID(),
				Title:    title,
				Severity: severity,
				Lead:     lead,
				Status:   status,
				Now:      clock.Now().Format(time.RFC3339),
			}

			loader := loaderFrom(cmd)
			manifest, err := loadIncidentManifest(cmd, loader)
			if err != nil {
				return err
			}
			rendered, err := sections.Merge(manifest, nil, struct{ Meta incidentMeta }{Meta: meta})
			if err != nil {
				return err
			}

			data := incidentData{Meta: meta, Sections: rendered}
			def := fmt.Sprintf("incident-%s.md", ids.Slug(title))
			opts := out.RenderOptionsStructured(cmd, def)
			opts.RenderArtifact = true
			return render.Render(cmd.OutOrStdout(), loader, "incident.md.tmpl", data, opts)
		},
	}
	cmd.Flags().StringVarP(&title, "title", "T", "", "incident title (required)")
	cmd.Flags().StringVar(&severity, "severity", "SEV-2", "incident severity (e.g. SEV-1, SEV-2)")
	cmd.Flags().StringVar(&lead, "lead", "", "incident lead")
	cmd.Flags().StringVar(&status, "status", "active", "incident status: investigated | active | contained | resolved")
	out.Bind(cmd, "write to file (default: incident-<slug>.md)")
	return cmd
}

func loadIncidentManifest(cmd *cobra.Command, loader *tmpl.Loader) (*sections.Manifest, error) {
	artifactBytes, err := loader.LoadArtifactBytes("incident.md.tmpl")
	if err != nil {
		return nil, fmt.Errorf("load incident.yaml: %w", err)
	}
	a, err := sections.ParseArtifact(artifactBytes)
	if err != nil {
		return nil, fmt.Errorf("parse incident.yaml: %w", err)
	}
	warnStaleLegacyFiles(cmd, loader, "incident.yaml", "incident.md.tmpl")
	return &sections.Manifest{Version: a.Version, Sections: a.Sections}, nil
}
