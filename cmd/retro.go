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
	retroTeam   string
	retroSprint string
	retroOut    string
	retroStdout bool
	retroForce  bool
	retroDry    bool
)

var retroCmd = &cobra.Command{
	Use:   "retro",
	Short: "Generate a sprint retrospective template",
	RunE: func(cmd *cobra.Command, _ []string) error {
		if retroTeam == "" {
			return errors.New("--team is required")
		}
		sprint := retroSprint
		if sprint == "" {
			sprint = time.Now().Format("2006-01-02")
		}
		data := struct {
			ID, Team, Sprint, Now string
		}{
			ID:     ids.UUID(),
			Team:   retroTeam,
			Sprint: sprint,
			Now:    time.Now().Format(time.RFC3339),
		}
		def := fmt.Sprintf("retro-%s-%s.md", ids.Slug(retroTeam), ids.Slug(sprint))
		return render.Render(cmd.OutOrStdout(), "retro.md.tmpl", data, render.Options{
			Out:     retroOut,
			Stdout:  retroStdout,
			Force:   retroForce,
			DryRun:  retroDry,
			Default: def,
		})
	},
}

func init() {
	retroCmd.Flags().StringVar(&retroTeam, "team", "", "team name (required)")
	retroCmd.Flags().StringVar(&retroSprint, "sprint", "", "sprint or period label (default: today's date)")
	retroCmd.Flags().StringVar(&retroOut, "out", "", "write to file (default: retro-<team>-<sprint>.md)")
	retroCmd.Flags().BoolVar(&retroStdout, "stdout", false, "print to stdout")
	retroCmd.Flags().BoolVar(&retroForce, "force", false, "overwrite existing file")
	retroCmd.Flags().BoolVar(&retroDry, "dry-run", false, "print result, do not write a file")
	rootCmd.AddCommand(retroCmd)
}
