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
			data := struct {
				ID      string `json:"id"`
				Service string `json:"service"`
				Horizon string `json:"horizon"`
				Now     string `json:"now"`
			}{
				ID:      ids.UUID(),
				Service: service,
				Horizon: horizon,
				Now:     clock.Now().Format(time.RFC3339),
			}
			def := fmt.Sprintf("capacity-%s.md", ids.Slug(service))
			return render.Render(cmd.OutOrStdout(), loaderFrom(cmd), "capacity.md.tmpl", data, out.RenderOptions(cmd, def))
		},
	}
	cmd.Flags().StringVar(&service, "service", "", "service name (required)")
	cmd.Flags().StringVar(&horizon, "horizon", "1y", "planning horizon (e.g. 6m, 1y, 2y)")
	out.Bind(cmd, "write to file (default: capacity-<service>.md)")
	return cmd
}
