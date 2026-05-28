package cmd

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/jtprogru/srekit/internal/cliflags"
	"github.com/jtprogru/srekit/internal/clock"
	"github.com/jtprogru/srekit/internal/ids"
	"github.com/jtprogru/srekit/internal/meta"
	"github.com/jtprogru/srekit/internal/render"
)

func newOncallCmd() *cobra.Command {
	var (
		team   string
		start  string
		end    string
		author string
		email  string
		out    cliflags.Output
	)
	cmd := &cobra.Command{
		Use:   "oncall-report",
		Short: "Generate a weekly on-call report template",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if team == "" {
				return errors.New("--team is required")
			}
			a, err := meta.Resolve(viper.GetViper(), author, email)
			if err != nil {
				return err
			}
			now := clock.Now()
			if start == "" || end == "" {
				wd := int(now.Weekday())
				if wd == 0 {
					wd = 7
				}
				weekStart := now.AddDate(0, 0, -wd+1)
				weekEnd := weekStart.AddDate(0, 0, 6)
				if start == "" {
					start = weekStart.Format("2006-01-02")
				}
				if end == "" {
					end = weekEnd.Format("2006-01-02")
				}
			}
			data := struct {
				ID, Team, Start, End, Now string
				Author                    meta.Author
			}{
				ID:     ids.UUID(),
				Team:   team,
				Start:  start,
				End:    end,
				Now:    now.Format(time.RFC3339),
				Author: a,
			}
			def := fmt.Sprintf("oncall-%s-%s.md", ids.Slug(team), start)
			return render.Render(cmd.OutOrStdout(), loaderFrom(cmd), "oncall.md.tmpl", data, out.RenderOptions(def))
		},
	}
	cmd.Flags().StringVar(&team, "team", "", "team name (required)")
	cmd.Flags().StringVar(&start, "start", "", "period start (default: this week's Monday)")
	cmd.Flags().StringVar(&end, "end", "", "period end (default: this week's Sunday)")
	cmd.Flags().StringVar(&author, "author", "", "on-caller name")
	cmd.Flags().StringVar(&email, "email", "", "on-caller email")
	out.Bind(cmd, "write to file (default: oncall-<team>-<start>.md)")
	return cmd
}
