package cmd

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/jtprogru/srekit/internal/ids"
	"github.com/jtprogru/srekit/internal/meta"
	"github.com/jtprogru/srekit/internal/render"
)

var (
	ocTeam   string
	ocStart  string
	ocEnd    string
	ocAuthor string
	ocEmail  string
	ocOut    string
	ocStdout bool
	ocForce  bool
	ocDry    bool
)

var oncallCmd = &cobra.Command{
	Use:   "oncall-report",
	Short: "Generate a weekly on-call report template",
	RunE: func(cmd *cobra.Command, _ []string) error {
		if ocTeam == "" {
			return errors.New("--team is required")
		}
		author, err := meta.Resolve(ocAuthor, ocEmail)
		if err != nil {
			return err
		}
		now := time.Now()
		start := ocStart
		end := ocEnd
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
			Team:   ocTeam,
			Start:  start,
			End:    end,
			Now:    now.Format(time.RFC3339),
			Author: author,
		}
		def := fmt.Sprintf("oncall-%s-%s.md", ids.Slug(ocTeam), start)
		return render.Render(cmd.OutOrStdout(), "oncall.md.tmpl", data, render.Options{
			Out:     ocOut,
			Stdout:  ocStdout,
			Force:   ocForce,
			DryRun:  ocDry,
			Default: def,
		})
	},
}

func init() {
	oncallCmd.Flags().StringVar(&ocTeam, "team", "", "team name (required)")
	oncallCmd.Flags().StringVar(&ocStart, "start", "", "period start (default: this week's Monday)")
	oncallCmd.Flags().StringVar(&ocEnd, "end", "", "period end (default: this week's Sunday)")
	oncallCmd.Flags().StringVar(&ocAuthor, "author", "", "on-caller name")
	oncallCmd.Flags().StringVar(&ocEmail, "email", "", "on-caller email")
	oncallCmd.Flags().StringVar(&ocOut, "out", "", "write to file (default: oncall-<team>-<start>.md)")
	oncallCmd.Flags().BoolVar(&ocStdout, "stdout", false, "print to stdout")
	oncallCmd.Flags().BoolVar(&ocForce, "force", false, "overwrite existing file")
	oncallCmd.Flags().BoolVar(&ocDry, "dry-run", false, "print result, do not write a file")
	rootCmd.AddCommand(oncallCmd)
}
