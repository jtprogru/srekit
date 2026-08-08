package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/jtprogru/srekit/internal/cliflags"
	"github.com/jtprogru/srekit/internal/clock"
	"github.com/jtprogru/srekit/internal/ids"
	"github.com/jtprogru/srekit/internal/render"
	"github.com/jtprogru/srekit/internal/sections"
	"github.com/jtprogru/srekit/internal/tmpl"
)

// postmortemMeta is the JSON-tagged metadata shipped under the top-level
// `meta` key. It mirrors the previous flat shape but now lives nested.
type postmortemMeta struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Severity string `json:"severity"`
	Start    string `json:"start"`
	End      string `json:"end"`
	Owner    string `json:"owner"`
	Now      string `json:"now"`
}

// postmortemData is what the renderer sees (for the v1 artifact path:
// passed to render.ArtifactPayload) and what `--json` marshals (the
// structured-shape contract for postmortem). The same struct serves both
// because the artifact render path bypasses Go-template execution.
type postmortemData struct {
	Meta     postmortemMeta             `json:"meta"`
	Sections []sections.RenderedSection `json:"sections"`
}

// ArtifactPayload satisfies render.ArtifactPayload. The ctx exposes Meta
// under .Meta so YAML scalars like "{{ .Meta.Title }}" resolve.
func (d postmortemData) ArtifactPayload() ([]sections.RenderedSection, any) {
	return d.Sections, struct{ Meta postmortemMeta }{Meta: d.Meta}
}

// The payload accepted by --from FILE is sections.Payload: both fields
// are optional, CLI flags take precedence over `meta` on conflict (so an
// agent can pin a title with --title even when the file disagrees), and
// unknown section IDs in `sections` are rejected by sections.Merge as a
// typo guard.

// postmortemMetaFields enumerates the meta-object property names exposed
// via `--schema`. It mirrors postmortemMeta's json tags; kept as a slice
// so output order is deterministic.
var postmortemMetaFields = []string{"id", "title", "severity", "start", "end", "owner", "now"}

func newPostmortemCmd() *cobra.Command {
	var (
		title    string
		severity string
		start    string
		end      string
		owner    string
		from     string
		schema   bool
		validate string
		out      cliflags.Output
	)
	cmd := &cobra.Command{
		Use:   "postmortem",
		Short: "Generate a Google SRE-style postmortem template",
		Example: `  # Write postmortem-<YYYY-MM-DD>-<slug>.md
  srekit postmortem --title "Checkout outage" --severity SEV-1 --owner bob

  # Inspect the structured JSON shape (sections in manifest order)
  srekit postmortem -T "Cache stampede" --json | jq '.sections[].id'

  # Round-trip: edit one section's body and re-render Markdown
  srekit postmortem -T X --json > pm.json
  # ...edit pm.json...
  srekit postmortem -T X --from pm.json

  # Emit a JSON Schema for --from input (useful for editor tooling / agents)
  srekit postmortem --schema > postmortem.schema.json

  # Validate an input file (required sections, unknown IDs)
  srekit postmortem --validate pm.json`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if schema && validate != "" {
				return errors.New("--schema and --validate are mutually exclusive")
			}

			loader := loaderFrom(cmd)
			manifest, err := loadPostmortemManifest(cmd, loader)
			if err != nil {
				return err
			}

			switch {
			case schema:
				return emitPostmortemSchema(cmd, manifest)
			case validate != "":
				return runPostmortemValidate(cmd, manifest, validate)
			}

			input, err := sections.ReadPayload(cmd.InOrStdin(), from)
			if err != nil {
				return err
			}

			// CLI flags override anything the input file might have set.
			mergedTitle := sections.PickNonEmpty(title, input.Meta["title"])
			if mergedTitle == "" {
				return errors.New("--title is required")
			}

			now := clock.Now()
			meta := postmortemMeta{
				ID:       sections.PickNonEmpty(input.Meta["id"], ids.UUID()),
				Title:    mergedTitle,
				Severity: sections.PickNonEmpty(severity, input.Meta["severity"]),
				Start:    sections.PickNonEmpty(start, input.Meta["start"]),
				End:      sections.PickNonEmpty(end, input.Meta["end"]),
				Owner:    sections.PickNonEmpty(owner, input.Meta["owner"]),
				Now:      sections.PickNonEmpty(input.Meta["now"], now.Format(time.RFC3339)),
			}

			rendered, err := sections.Merge(manifest, input.Sections, struct {
				Meta postmortemMeta
			}{Meta: meta})
			if err != nil {
				return err
			}

			data := postmortemData{Meta: meta, Sections: rendered}
			def := fmt.Sprintf("postmortem-%s-%s.md", now.Format("2006-01-02"), ids.Slug(meta.Title))
			opts := out.RenderOptions(cmd, def)
			return render.Render(cmd.OutOrStdout(), loader, "postmortem", data, opts)
		},
	}
	cmd.Flags().StringVarP(&title, "title", "T", "", "incident title (required, unless provided via --from)")
	cmd.Flags().StringVar(&severity, "severity", "SEV-3", "incident severity (e.g. SEV-1, SEV-2)")
	cmd.Flags().StringVar(&start, "start", "", "incident start time (RFC3339 or human)")
	cmd.Flags().StringVar(&end, "end", "", "incident end time")
	cmd.Flags().StringVar(&owner, "owner", "", "incident owner")
	cmd.Flags().StringVar(&from, "from", "", "read sections from JSON file (- for stdin)")
	cmd.Flags().BoolVar(&schema, "schema", false, "emit JSON Schema for --from input (mutually exclusive with --validate)")
	cmd.Flags().StringVar(&validate, "validate", "", "validate an input file: required sections non-empty, no unknown IDs")
	out.Bind(cmd, "write to file (default: postmortem-<YYYY-MM-DD>-<slug>.md)")
	return cmd
}

// loadPostmortemManifest resolves the postmortem section list through
// the v1 artifact path (`postmortem.yaml`). As a side-effect it warns
// about stale legacy files (`postmortem.md.tmpl`, `postmortem.sections.yaml`)
// in the user's templates dir, nudging the user to migrate their
// customizations into the v1 file.
func loadPostmortemManifest(cmd *cobra.Command, loader *tmpl.Loader) (*sections.Manifest, error) {
	artifactBytes, err := loader.LoadArtifactBytes("postmortem")
	if err != nil {
		return nil, fmt.Errorf("load postmortem.yaml: %w", err)
	}
	a, err := sections.ParseArtifact(artifactBytes)
	if err != nil {
		return nil, fmt.Errorf("parse postmortem.yaml: %w", err)
	}
	warnStaleLegacyFiles(cmd, loader, "postmortem.yaml", "postmortem.md.tmpl", "postmortem.sections.yaml")
	return &sections.Manifest{Version: a.Version, Sections: a.Sections}, nil
}

// emitPostmortemSchema marshals the manifest-derived JSON Schema to the
// command's stdout. The schema is recomputed from the loaded manifest on
// every invocation, so user customizations to postmortem.yaml flow
// through automatically.
func emitPostmortemSchema(cmd *cobra.Command, manifest *sections.Manifest) error {
	schema := manifest.JSONSchema("srekit postmortem --from input", postmortemMetaFields)
	b, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("encode schema: %w", err)
	}
	_, err = fmt.Fprintln(cmd.OutOrStdout(), string(b))
	return err
}

// runPostmortemValidate reads file, rejects unknown section IDs via
// sections.Merge's typo-guard, and reports per-required-section completeness.
// Exits non-zero if any required section is missing or empty, mirroring
// `templates validate`'s OK/FAIL style.
func runPostmortemValidate(cmd *cobra.Command, manifest *sections.Manifest, file string) error {
	input, err := sections.ReadPayload(cmd.InOrStdin(), file)
	if err != nil {
		return err
	}

	// Typo guard only — skip default-body rendering so an empty input
	// doesn't trip on a default that needs ctx (e.g. timeline rows that
	// reference .Meta.Start).
	if err := sections.CheckUnknownIDs(manifest, input.Sections); err != nil {
		return err
	}

	results := manifest.RequiredCheck(input.Sections)
	out := cmd.OutOrStdout()
	var failed int
	for _, r := range results {
		switch r.Status {
		case "OK":
			fmt.Fprintf(out, "OK    %s\n", r.ID)
		default:
			fmt.Fprintf(out, "FAIL  %s: %s\n", r.ID, r.Reason)
			failed++
		}
	}
	if failed > 0 {
		return fmt.Errorf("%d of %d required section(s) failed validation", failed, len(results))
	}
	return nil
}

// Reading a --from payload and resolving flag-over-file precedence live
// in internal/sections (ReadPayload, PickNonEmpty): both are shaped by
// the artifact contract rather than by this command, and changelog reads
// the same payload shape.
