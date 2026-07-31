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
	"github.com/jtprogru/srekit/internal/sections"
	"github.com/jtprogru/srekit/internal/tmpl"
)

// taskMeta is the JSON-tagged metadata for the task artifact. Mirrors
// the flat shape used by the v0.13.x bootstrap envelope so consumers
// pulling `result.meta.title` continue to work.
type taskMeta struct {
	ID               string `json:"id"`
	CreationDate     string `json:"creationDate"`
	ModificationDate string `json:"modificationDate"`
	Title            string `json:"title"`
}

// taskData implements render.ArtifactPayload — task migrated to the v1
// artifact render path in v0.15.0, becoming the first fresh migration
// (postmortem in v0.14.0 was a refactor of an existing prototype).
type taskData struct {
	Meta     taskMeta                   `json:"meta"`
	Sections []sections.RenderedSection `json:"sections"`
}

func (d taskData) ArtifactPayload() ([]sections.RenderedSection, any) {
	return d.Sections, struct{ Meta taskMeta }{Meta: d.Meta}
}

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
		Example: `  # Investigation log in the current directory
  srekit task --title "Latency regression in api-gw"

  # Into a specific directory
  srekit task -T "Disk fills on db-03" --path ./investigations`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if title == "" {
				return errors.New("--title is required")
			}
			now := clock.Now().Format(time.RFC3339)
			meta := taskMeta{
				ID:               ids.UUID(),
				CreationDate:     now,
				ModificationDate: now,
				Title:            title,
			}

			loader := loaderFrom(cmd)
			manifest, err := loadTaskManifest(cmd, loader)
			if err != nil {
				return err
			}

			rendered, err := sections.Merge(manifest, nil, struct{ Meta taskMeta }{Meta: meta})
			if err != nil {
				return err
			}

			data := taskData{Meta: meta, Sections: rendered}
			def := filepath.Join(path, fmt.Sprintf("investigation-%s.md", ids.Slug(title)))
			// Structured render: bypass bootstrap envelope so --json emits
			// the per-section payload (same contract as postmortem).
			opts := out.RenderOptions(cmd, def)
			return render.Render(cmd.OutOrStdout(), loader, "task", data, opts)
		},
	}
	cmd.Flags().StringVarP(&title, "title", "T", "", "investigation title (required)")
	cmd.Flags().StringVarP(&path, "path", "P", "./", "directory for default output file")
	out.Bind(cmd, "write to file (default: investigation-<slug>.md)")
	return cmd
}

// loadTaskManifest resolves the task artifact and returns its section
// list as a Manifest (suitable for sections.Merge). Warns about stale
// pre-v0.14.0 task.md.tmpl files in user templates dir.
func loadTaskManifest(cmd *cobra.Command, loader *tmpl.Loader) (*sections.Manifest, error) {
	artifactBytes, err := loader.LoadArtifactBytes("task")
	if err != nil {
		return nil, fmt.Errorf("load task.yaml: %w", err)
	}
	a, err := sections.ParseArtifact(artifactBytes)
	if err != nil {
		return nil, fmt.Errorf("parse task.yaml: %w", err)
	}
	warnStaleLegacyFiles(cmd, loader, "task.yaml", "task.md.tmpl")
	return &sections.Manifest{Version: a.Version, Sections: a.Sections}, nil
}
