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
		RunE: func(cmd *cobra.Command, _ []string) error {
			if title == "" {
				return errors.New("--title is required")
			}
			data := struct {
				ID, Title, Severity, Start, End, Owner, Now string
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
			return render.Render(cmd.OutOrStdout(), loaderFrom(cmd), "postmortem.md.tmpl", data, out.RenderOptions(def))
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
