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
			data := struct {
				ID, Service, Target, Window, LatencyTarget, Now string
			}{
				ID:            ids.UUID(),
				Service:       service,
				Target:        target,
				Window:        window,
				LatencyTarget: latencyTarget,
				Now:           clock.Now().Format(time.RFC3339),
			}
			def := fmt.Sprintf("slo-%s.md", ids.Slug(service))
			return render.Render(cmd.OutOrStdout(), loaderFrom(cmd), "slo.md.tmpl", data, out.RenderOptions(cmd, def))
		},
	}
	cmd.Flags().StringVar(&service, "service", "", "service name (required)")
	cmd.Flags().StringVar(&target, "target", "99.9%", "availability target")
	cmd.Flags().StringVar(&window, "window", "30d", "rolling SLO window")
	cmd.Flags().StringVar(&latencyTarget, "latency", "300ms", "p99 latency target")
	out.Bind(cmd, "write to file (default: slo-<service>.md)")
	return cmd
}
