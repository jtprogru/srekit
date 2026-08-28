package cmd

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/jtprogru/srekit/internal/cliflags"
	"github.com/jtprogru/srekit/internal/clock"
	"github.com/jtprogru/srekit/internal/ids"
	"github.com/jtprogru/srekit/internal/render"
	"github.com/jtprogru/srekit/internal/sections"
	"github.com/jtprogru/srekit/internal/tmpl"
)

// taskerMeta is the JSON-tagged metadata for the tasker artifact.
//
// Level is a list and Duration a number because that is what they are in
// the note: the frontmatter of a Tasker card is read by the vault that
// holds the cards, not by a human, and `level: [middle, senior]` filters
// where `level: "middle, senior"` does not. The artifact renders both
// through explicitly tagged frontmatter scalars — see retypeScalar in
// internal/sections.
type taskerMeta struct {
	ID       string   `json:"id"`
	Now      string   `json:"now"`
	Title    string   `json:"title"`
	Topic    string   `json:"topic"`
	Level    []string `json:"level"`
	Format   string   `json:"format"`
	Duration int      `json:"duration"`
}

type taskerData struct {
	Meta     taskerMeta                 `json:"meta"`
	Sections []sections.RenderedSection `json:"sections"`
}

func (d taskerData) ArtifactPayload() ([]sections.RenderedSection, any) {
	return d.Sections, struct{ Meta taskerMeta }{Meta: d.Meta}
}

func newTaskerCmd() *cobra.Command {
	var (
		title    string
		topic    string
		level    []string
		format   string
		duration int
		out      cliflags.Output
	)
	cmd := &cobra.Command{
		Use:   "tasker",
		Short: "Generate an engineering task card (topic, level, format, duration)",
		Long: `Generate a task card for a collection of engineering tasks: front matter with topic, level, format and duration, an H1 naming the task, and the two sections a card carries — the task itself and what a good answer sounds like.

The document is deliberately unfilled: the card is written by the person adding the task, and srekit only lays out the shape their existing collection already uses.`,
		Example: `  # Card with the defaults
  srekit tasker --title "Каналы и select"

  # A short theory question for a junior
  srekit tasker -T "Что такое GOMAXPROCS" --topic go --level junior \
    --format theory --duration 10`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if title == "" {
				return errors.New("--title is required")
			}
			levels := cleanLevels(level)
			if len(levels) == 0 {
				return errors.New("--level requires at least one value")
			}
			if duration <= 0 {
				return fmt.Errorf("--duration must be a positive number of minutes, got %d", duration)
			}
			meta := taskerMeta{
				ID:       ids.UUID(),
				Now:      clock.Now().Format(time.RFC3339),
				Title:    title,
				Topic:    topic,
				Level:    levels,
				Format:   format,
				Duration: duration,
			}

			loader := loaderFrom(cmd)
			manifest, err := loadTaskerManifest(loader)
			if err != nil {
				return err
			}
			rendered, err := sections.Merge(manifest, nil, struct{ Meta taskerMeta }{Meta: meta})
			if err != nil {
				return err
			}

			data := taskerData{Meta: meta, Sections: rendered}
			def := fmt.Sprintf("tasker-%s.md", ids.Slug(title))
			opts := out.RenderOptions(cmd, def)
			return render.Render(cmd.OutOrStdout(), loader, "tasker", data, opts)
		},
	}
	cmd.Flags().StringVarP(&title, "title", "T", "", "task name (required)")
	cmd.Flags().StringVar(&topic, "topic", "go", "subject area of the task")
	cmd.Flags().StringSliceVar(&level, "level", []string{"middle", "senior"}, "target levels (repeatable or comma-separated)")
	cmd.Flags().StringVar(&format, "format", "code", "how the task is answered (code, theory, design, …)")
	cmd.Flags().IntVar(&duration, "duration", 30, "expected time to solve, in minutes")
	out.Bind(cmd, "write to file (default: tasker-<slug>.md)")
	return cmd
}

// cleanLevels trims each level and drops the empties, so `--level
// "middle, "` is one level rather than a trailing blank that would render
// as `[middle, ]` and turn the whole frontmatter value into a parse
// error nobody asked for.
func cleanLevels(in []string) []string {
	out := make([]string, 0, len(in))
	for _, l := range in {
		if l = strings.TrimSpace(l); l != "" {
			out = append(out, l)
		}
	}
	return out
}

func loadTaskerManifest(loader *tmpl.Loader) (*sections.Manifest, error) {
	artifactBytes, err := loader.LoadArtifactBytes("tasker")
	if err != nil {
		return nil, fmt.Errorf("load tasker.yaml: %w", err)
	}
	a, err := sections.ParseArtifact(artifactBytes)
	if err != nil {
		return nil, fmt.Errorf("parse tasker.yaml: %w", err)
	}
	// No warnStaleLegacyFiles call, unlike every generator above it: the
	// legacy `.tmpl` / `.sections.yaml` layouts were retired before this
	// artifact existed, so there is no `tasker.md.tmpl` anywhere to warn
	// about — which is also why this loader needs no *cobra.Command.
	return &sections.Manifest{Version: a.Version, Sections: a.Sections}, nil
}
