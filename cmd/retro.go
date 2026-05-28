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

func newRetroCmd() *cobra.Command {
	var (
		team   string
		sprint string
		out    cliflags.Output
	)
	cmd := &cobra.Command{
		Use:   "retro",
		Short: "Generate a sprint retrospective template",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if team == "" {
				return errors.New("--team is required")
			}
			s := sprint
			if s == "" {
				s = clock.Now().Format("2006-01-02")
			}
			data := struct {
				ID, Team, Sprint, Now string
			}{
				ID:     ids.UUID(),
				Team:   team,
				Sprint: s,
				Now:    clock.Now().Format(time.RFC3339),
			}
			def := fmt.Sprintf("retro-%s-%s.md", ids.Slug(team), ids.Slug(s))
			return render.Render(cmd.OutOrStdout(), loaderFrom(cmd), "retro.md.tmpl", data, out.RenderOptions(cmd, def))
		},
	}
	cmd.Flags().StringVar(&team, "team", "", "team name (required)")
	cmd.Flags().StringVar(&sprint, "sprint", "", "sprint or period label (default: today's date)")
	out.Bind(cmd, "write to file (default: retro-<team>-<sprint>.md)")
	return cmd
}
