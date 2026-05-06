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
	rbTitle   string
	rbService string
	rbAlert   string
	rbOut     string
	rbStdout  bool
	rbForce   bool
	rbDry     bool
)

var runbookCmd = &cobra.Command{
	Use:   "runbook",
	Short: "Generate an on-call runbook template",
	Long:  "Generate a runbook with Symptoms / Diagnose / Mitigate / Verify sections.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		if rbTitle == "" {
			return errors.New("--title is required")
		}
		data := struct {
			ID, Title, Service, Alert, Now string
		}{
			ID:      ids.UUID(),
			Title:   rbTitle,
			Service: valueOr(rbService, "<service name>"),
			Alert:   valueOr(rbAlert, "<alert name>"),
			Now:     time.Now().Format(time.RFC3339),
		}
		def := fmt.Sprintf("runbook-%s.md", ids.Slug(rbTitle))
		return render.Render(cmd.OutOrStdout(), "runbook.md.tmpl", data, render.Options{
			Out:     rbOut,
			Stdout:  rbStdout,
			Force:   rbForce,
			DryRun:  rbDry,
			Default: def,
		})
	},
}

func init() {
	runbookCmd.Flags().StringVarP(&rbTitle, "title", "T", "", "runbook title (required)")
	runbookCmd.Flags().StringVar(&rbService, "service", "", "service name")
	runbookCmd.Flags().StringVar(&rbAlert, "alert", "", "alert name this runbook responds to")
	runbookCmd.Flags().StringVar(&rbOut, "out", "", "write to file (default: runbook-<slug>.md)")
	runbookCmd.Flags().BoolVar(&rbStdout, "stdout", false, "print to stdout")
	runbookCmd.Flags().BoolVar(&rbForce, "force", false, "overwrite existing file")
	runbookCmd.Flags().BoolVar(&rbDry, "dry-run", false, "print result, do not write a file")
	rootCmd.AddCommand(runbookCmd)
}
