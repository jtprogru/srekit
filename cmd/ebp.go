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
			data := struct {
				ID, Service, Now string
			}{
				ID:      ids.UUID(),
				Service: service,
				Now:     clock.Now().Format(time.RFC3339),
			}
			def := fmt.Sprintf("ebp-%s.md", ids.Slug(service))
			return render.Render(cmd.OutOrStdout(), loaderFrom(cmd), "ebp.md.tmpl", data, out.RenderOptions(cmd, def))
		},
	}
	cmd.Flags().StringVar(&service, "service", "", "service name (required)")
	out.Bind(cmd, "write to file (default: ebp-<service>.md)")
	return cmd
}
