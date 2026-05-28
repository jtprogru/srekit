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

func newPostmortemCmd() *cobra.Command {
	var (
		title    string
		severity string
		start    string
		end      string
		owner    string
		out      cliflags.Output
	)
	cmd := &cobra.Command{
		Use:   "postmortem",
		Short: "Generate a Google SRE-style postmortem template",
		Example: `  # Write postmortem-<slug>.md
  srekit postmortem --title "Checkout outage" --severity SEV-1 --owner bob

  # Pipe to stdout for review
  srekit postmortem -T "Cache stampede" --start 2026-05-20T10:00:00Z --out -`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if title == "" {
				return errors.New("--title is required")
			}
			data := struct {
				ID       string `json:"id"`
				Title    string `json:"title"`
				Severity string `json:"severity"`
				Start    string `json:"start"`
				End      string `json:"end"`
				Owner    string `json:"owner"`
				Now      string `json:"now"`
			}{
				ID:       ids.UUID(),
				Title:    title,
				Severity: severity,
				Start:    start,
				End:      end,
				Owner:    owner,
				Now:      clock.Now().Format(time.RFC3339),
			}
			def := fmt.Sprintf("postmortem-%s.md", ids.Slug(title))
			return render.Render(cmd.OutOrStdout(), loaderFrom(cmd), "postmortem.md.tmpl", data, out.RenderOptions(cmd, def))
		},
	}
	cmd.Flags().StringVarP(&title, "title", "T", "", "incident title (required)")
	cmd.Flags().StringVar(&severity, "severity", "SEV-3", "incident severity (e.g. SEV-1, SEV-2)")
	cmd.Flags().StringVar(&start, "start", "", "incident start time (RFC3339 or human)")
	cmd.Flags().StringVar(&end, "end", "", "incident end time")
	cmd.Flags().StringVar(&owner, "owner", "", "incident owner")
	out.Bind(cmd, "write to file (default: postmortem-<slug>.md)")
	return cmd
}
