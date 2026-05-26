package cmd

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/jtprogru/srekit/internal/cliflags"
	"github.com/jtprogru/srekit/internal/clock"
	"github.com/jtprogru/srekit/internal/ids"
	"github.com/jtprogru/srekit/internal/render"
)

func newTaskCmd() *cobra.Command {
	var (
		title string
		path  string
		out   cliflags.Output
	)
	cmd := &cobra.Command{
		Use:     "task",
		Aliases: []string{"sretask"},
		Short:   "Generate a markdown SRE investigation log",
		Long:    "Generate a markdown SRE investigation log with YAML front matter and sections for context, hypothesis, evidence, findings, and action items.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if title == "" {
				return errors.New("--title is required")
			}
			now := clock.Now().Format(time.RFC3339)
			data := struct {
				ID, CreationDate, ModificationDate, Title string
			}{
				ID:               ids.UUID(),
				CreationDate:     now,
				ModificationDate: now,
				Title:            title,
			}
			def := filepath.Join(path, fmt.Sprintf("investigation-%s.md", ids.Slug(title)))
			return render.Render(cmd.OutOrStdout(), "task.md.tmpl", data, out.RenderOptions(def))
		},
	}
	cmd.Flags().StringVarP(&title, "title", "T", "", "investigation title (required)")
	cmd.Flags().StringVarP(&path, "path", "P", "./", "directory for default output file")
	out.Bind(cmd, "write to file (default: investigation-<slug>.md)")
	return cmd
}
