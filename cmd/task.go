package cmd

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jtprogru/srekit/internal/cliflags"
	"github.com/jtprogru/srekit/internal/clock"
	"github.com/jtprogru/srekit/internal/ids"
	"github.com/jtprogru/srekit/internal/render"
)

const taskTimeFormat = "2006-01-02T15:04:05"

func newTaskCmd() *cobra.Command {
	var (
		title string
		path  string
		out   cliflags.Output
	)
	cmd := &cobra.Command{
		Use:     "task",
		Aliases: []string{"sretask"},
		Short:   "Generate a markdown SRE task note",
		Long:    "Generate a markdown SRE task note with YAML front matter, ready for an interview / investigation log.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if title == "" {
				return errors.New("--title is required")
			}
			now := clock.Now().Format(taskTimeFormat)
			data := struct {
				ID, CreationDate, ModificationDate, Title string
			}{
				ID:               ids.UUID(),
				CreationDate:     now,
				ModificationDate: now,
				Title:            title,
			}
			def := filepath.Join(path, fmt.Sprintf("Tasker - %s.md", title))
			return render.Render(cmd.OutOrStdout(), "task.md.tmpl", data, out.RenderOptions(def))
		},
	}
	cmd.Flags().StringVarP(&title, "title", "T", "", "task title (required)")
	cmd.Flags().StringVarP(&path, "path", "P", "./", "directory for default output file")
	out.Bind(cmd, "explicit output file path")
	return cmd
}
