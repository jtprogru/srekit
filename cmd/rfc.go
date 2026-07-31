package cmd

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/jtprogru/srekit/internal/cliflags"
	"github.com/jtprogru/srekit/internal/clock"
	"github.com/jtprogru/srekit/internal/config"
	"github.com/jtprogru/srekit/internal/ids"
	"github.com/jtprogru/srekit/internal/meta"
	"github.com/jtprogru/srekit/internal/render"
	"github.com/jtprogru/srekit/internal/sections"
	"github.com/jtprogru/srekit/internal/tmpl"
)

var allowedRFCStatus = map[string]bool{
	"proposed":   true,
	"accepted":   true,
	"rejected":   true,
	"superseded": true,
	"deprecated": true,
}

type rfcMeta struct {
	ID     string      `json:"id"`
	Title  string      `json:"title"`
	Status string      `json:"status"`
	Now    string      `json:"now"`
	Author meta.Author `json:"author"`
}

type rfcData struct {
	Meta     rfcMeta                    `json:"meta"`
	Sections []sections.RenderedSection `json:"sections"`
}

func (d rfcData) ArtifactPayload() ([]sections.RenderedSection, any) {
	return d.Sections, struct{ Meta rfcMeta }{Meta: d.Meta}
}

func newRFCCmd() *cobra.Command {
	var (
		title  string
		status string
		author string
		email  string
		out    cliflags.Output
	)
	cmd := &cobra.Command{
		Use:   "rfc",
		Short: "Generate an RFC / ADR markdown template",
		Long:  "Generate an RFC (a.k.a. ADR) document with Context / Decision / Alternatives / Consequences sections.",
		Example: `  # Proposed RFC (author/email come from git config if omitted)
  srekit rfc --title "Adopt OpenTelemetry"

  # Mark an accepted decision
  srekit rfc -T "Drop legacy auth" --status accepted`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if title == "" {
				return errors.New("--title is required")
			}
			if !allowedRFCStatus[status] {
				return fmt.Errorf("invalid --status %q (proposed, accepted, rejected, superseded, deprecated)", status)
			}
			a, err := meta.Resolve(config.Global(), author, email)
			if err != nil {
				return err
			}
			m := rfcMeta{
				ID:     ids.UUID(),
				Title:  title,
				Status: status,
				Now:    clock.Now().Format(time.RFC3339),
				Author: a,
			}

			loader := loaderFrom(cmd)
			manifest, err := loadRFCManifest(cmd, loader)
			if err != nil {
				return err
			}
			rendered, err := sections.Merge(manifest, nil, struct{ Meta rfcMeta }{Meta: m})
			if err != nil {
				return err
			}

			data := rfcData{Meta: m, Sections: rendered}
			def := fmt.Sprintf("rfc-%s.md", ids.Slug(title))
			opts := out.RenderOptions(cmd, def)
			return render.Render(cmd.OutOrStdout(), loader, "rfc", data, opts)
		},
	}
	cmd.Flags().StringVarP(&title, "title", "T", "", "RFC title (required)")
	cmd.Flags().StringVar(&status, "status", "proposed", "RFC status: proposed | accepted | rejected | superseded | deprecated")
	cmd.Flags().StringVar(&author, "author", "", "author full name")
	cmd.Flags().StringVar(&email, "email", "", "author email")
	out.Bind(cmd, "write to file (default: rfc-<slug>.md)")
	return cmd
}

func loadRFCManifest(cmd *cobra.Command, loader *tmpl.Loader) (*sections.Manifest, error) {
	artifactBytes, err := loader.LoadArtifactBytes("rfc")
	if err != nil {
		return nil, fmt.Errorf("load rfc.yaml: %w", err)
	}
	a, err := sections.ParseArtifact(artifactBytes)
	if err != nil {
		return nil, fmt.Errorf("parse rfc.yaml: %w", err)
	}
	warnStaleLegacyFiles(cmd, loader, "rfc.yaml", "rfc.md.tmpl")
	return &sections.Manifest{Version: a.Version, Sections: a.Sections}, nil
}
