package cmd

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/jtprogru/srekit/internal/ids"
	"github.com/jtprogru/srekit/internal/render"
)

var (
	taskTitle  string
	taskPath   string
	taskOut    string
	taskStdout bool
	taskForce  bool
	taskDry    bool
)

const taskTimeFormat = "2006-01-02T15:04:05"

var taskCmd = &cobra.Command{
	Use:     "task",
	Aliases: []string{"sretask"},
	Short:   "Generate a markdown SRE task note",
	Long:    "Generate a markdown SRE task note with YAML front matter, ready for an interview / investigation log.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		if taskTitle == "" {
			return errors.New("--title is required")
		}
		now := time.Now().Format(taskTimeFormat)
		data := struct {
			ID, CreationDate, ModificationDate, Title string
		}{
			ID:               ids.UUID(),
			CreationDate:     now,
			ModificationDate: now,
			Title:            taskTitle,
		}
		def := filepath.Join(taskPath, fmt.Sprintf("Tasker - %s.md", taskTitle))
		return render.Render(cmd.OutOrStdout(), "task.md.tmpl", data, render.Options{
			Out:     taskOut,
			Stdout:  taskStdout,
			Force:   taskForce,
			DryRun:  taskDry,
			Default: def,
		})
	},
}

func init() {
	taskCmd.Flags().StringVarP(&taskTitle, "title", "T", "", "task title (required)")
	taskCmd.Flags().StringVarP(&taskPath, "path", "P", "./", "directory for default output file")
	taskCmd.Flags().StringVar(&taskOut, "out", "", "explicit output file path")
	taskCmd.Flags().BoolVar(&taskStdout, "stdout", false, "print to stdout instead of writing a file")
	taskCmd.Flags().BoolVar(&taskForce, "force", false, "overwrite existing file")
	taskCmd.Flags().BoolVar(&taskDry, "dry-run", false, "print result, do not write a file")
	rootCmd.AddCommand(taskCmd)
}
