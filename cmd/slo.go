package cmd

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/jtprogru/srekit/internal/ids"
	"github.com/jtprogru/srekit/internal/render"
)

var (
	sloService       string
	sloTarget        string
	sloWindow        string
	sloLatencyTarget string
	sloOut           string
	sloStdout        bool
	sloForce         bool
	sloDry           bool
)

var sloCmd = &cobra.Command{
	Use:   "slo",
	Short: "Generate an SLO/SLI document template",
	RunE: func(cmd *cobra.Command, _ []string) error {
		if sloService == "" {
			return errors.New("--service is required")
		}
		data := struct {
			ID, Service, Target, Window, LatencyTarget, Now string
		}{
			ID:            ids.UUID(),
			Service:       sloService,
			Target:        sloTarget,
			Window:        sloWindow,
			LatencyTarget: sloLatencyTarget,
			Now:           time.Now().Format(time.RFC3339),
		}
		def := fmt.Sprintf("slo-%s.md", ids.Slug(sloService))
		return render.Render(cmd.OutOrStdout(), "slo.md.tmpl", data, render.Options{
			Out:     sloOut,
			Stdout:  sloStdout,
			Force:   sloForce,
			DryRun:  sloDry,
			Default: def,
		})
	},
}

func init() {
	sloCmd.Flags().StringVar(&sloService, "service", "", "service name (required)")
	sloCmd.Flags().StringVar(&sloTarget, "target", "99.9%", "availability target")
	sloCmd.Flags().StringVar(&sloWindow, "window", "30d", "rolling SLO window")
	sloCmd.Flags().StringVar(&sloLatencyTarget, "latency", "300ms", "p99 latency target")
	sloCmd.Flags().StringVar(&sloOut, "out", "", "write to file (default: slo-<service>.md)")
	sloCmd.Flags().BoolVar(&sloStdout, "stdout", false, "print to stdout")
	sloCmd.Flags().BoolVar(&sloForce, "force", false, "overwrite existing file")
	sloCmd.Flags().BoolVar(&sloDry, "dry-run", false, "print result, do not write a file")
	rootCmd.AddCommand(sloCmd)
}
