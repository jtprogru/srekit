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
		RunE: func(cmd *cobra.Command, _ []string) error {
			if title == "" {
				return errors.New("--title is required")
			}
			data := struct {
				ID, Title, Service, Alert, Now string
			}{
				ID:      ids.UUID(),
				Title:   title,
				Service: valueOr(service, "<service name>"),
				Alert:   valueOr(alert, "<alert name>"),
				Now:     clock.Now().Format(time.RFC3339),
			}
			def := fmt.Sprintf("runbook-%s.md", ids.Slug(title))
			return render.Render(cmd.OutOrStdout(), "runbook.md.tmpl", data, out.RenderOptions(def))
		},
	}
	cmd.Flags().StringVarP(&title, "title", "T", "", "runbook title (required)")
	cmd.Flags().StringVar(&service, "service", "", "service name")
	cmd.Flags().StringVar(&alert, "alert", "", "alert name this runbook responds to")
	out.Bind(cmd, "write to file (default: runbook-<slug>.md)")
	return cmd
}
