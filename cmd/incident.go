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
)

var allowedIncidentStatus = map[string]bool{
	"investigated": true,
	"active":       true,
	"contained":    true,
	"resolved":     true,
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
			data := struct {
				ID, Title, Severity, Lead, Status, Now string
			}{
				ID:       ids.UUID(),
				Title:    title,
				Severity: severity,
				Lead:     lead,
				Status:   status,
				Now:      clock.Now().Format(time.RFC3339),
			}
			def := fmt.Sprintf("incident-%s.md", ids.Slug(title))
			return render.Render(cmd.OutOrStdout(), loaderFrom(cmd), "incident.md.tmpl", data, out.RenderOptions(cmd, def))
		},
	}
	cmd.Flags().StringVarP(&title, "title", "T", "", "incident title (required)")
	cmd.Flags().StringVar(&severity, "severity", "SEV-2", "incident severity (e.g. SEV-1, SEV-2)")
	cmd.Flags().StringVar(&lead, "lead", "", "incident lead")
	cmd.Flags().StringVar(&status, "status", "active", "incident status: investigated | active | contained | resolved")
	out.Bind(cmd, "write to file (default: incident-<slug>.md)")
	return cmd
}
