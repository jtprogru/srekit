package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/jtprogru/srekit/internal/ids"
	"github.com/jtprogru/srekit/internal/render"
)

var (
	pmTitle    string
	pmSeverity string
	pmStart    string
	pmEnd      string
	pmOwner    string
	pmOut      string
	pmStdout   bool
	pmForce    bool
	pmDry      bool
)

var postmortemCmd = &cobra.Command{
	Use:   "postmortem",
	Short: "Generate a Google SRE-style postmortem template",
	RunE: func(cmd *cobra.Command, _ []string) error {
		if pmTitle == "" {
			return fmt.Errorf("--title is required")
		}
		now := time.Now().Format(time.RFC3339)
		data := struct {
			ID, Title, Severity, Start, End, Owner, Now string
		}{
			ID:       ids.UUID(),
			Title:    pmTitle,
			Severity: pmSeverity,
			Start:    valueOr(pmStart, "<start time>"),
			End:      valueOr(pmEnd, "<end time>"),
			Owner:    valueOr(pmOwner, "<incident owner>"),
			Now:      now,
		}
		def := fmt.Sprintf("postmortem-%s.md", ids.Slug(pmTitle))
		return render.Render(cmd.OutOrStdout(), "postmortem.md.tmpl", data, render.Options{
			Out:     pmOut,
			Stdout:  pmStdout,
			Force:   pmForce,
			DryRun:  pmDry,
			Default: def,
		})
	},
}

func valueOr(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

func init() {
	postmortemCmd.Flags().StringVarP(&pmTitle, "title", "T", "", "incident title (required)")
	postmortemCmd.Flags().StringVar(&pmSeverity, "severity", "SEV-3", "incident severity (e.g. SEV-1, SEV-2)")
	postmortemCmd.Flags().StringVar(&pmStart, "start", "", "incident start time (RFC3339 or human)")
	postmortemCmd.Flags().StringVar(&pmEnd, "end", "", "incident end time")
	postmortemCmd.Flags().StringVar(&pmOwner, "owner", "", "incident owner")
	postmortemCmd.Flags().StringVar(&pmOut, "out", "", "write to file (default: postmortem-<slug>.md)")
	postmortemCmd.Flags().BoolVar(&pmStdout, "stdout", false, "print to stdout")
	postmortemCmd.Flags().BoolVar(&pmForce, "force", false, "overwrite existing file")
	postmortemCmd.Flags().BoolVar(&pmDry, "dry-run", false, "print result, do not write a file")
	rootCmd.AddCommand(postmortemCmd)
}
