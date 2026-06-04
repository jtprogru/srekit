package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/jtprogru/srekit/internal/cliflags"
	"github.com/jtprogru/srekit/internal/clock"
	"github.com/jtprogru/srekit/internal/ids"
	"github.com/jtprogru/srekit/internal/render"
	"github.com/jtprogru/srekit/internal/sections"
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

// postmortemData is what the template's {{ range .Sections }} loop sees and
// what `--json` marshals: the structured-shape contract for postmortem.
type postmortemData struct {
	Meta     postmortemMeta             `json:"meta"`
	Sections []sections.RenderedSection `json:"sections"`
}

// postmortemInput is the schema accepted by --from FILE. Both fields are
// optional; CLI flags take precedence over `meta` on conflict (so an agent
// can pin a title with --title even when the file disagrees). Unknown
// section IDs in `sections` are rejected by sections.Merge as typo-guard.
type postmortemInput struct {
	Meta     map[string]string `json:"meta"`
	Sections map[string]string `json:"sections"`
}

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
			manifestBytes, err := loader.LoadManifestBytes("postmortem.md.tmpl")
			if err != nil {
				return fmt.Errorf("load postmortem manifest: %w", err)
			}
			manifest, err := sections.ParseManifest(manifestBytes)
			if err != nil {
				return err
			}

			switch {
			case schema:
				return emitPostmortemSchema(cmd, manifest)
			case validate != "":
				return runPostmortemValidate(cmd, manifest, validate)
			}

			input, err := readPostmortemInput(cmd, from)
			if err != nil {
				return err
			}

			// CLI flags override anything the input file might have set.
			mergedTitle := pickNonEmpty(title, input.Meta["title"])
			if mergedTitle == "" {
				return errors.New("--title is required")
			}

			now := clock.Now()
			meta := postmortemMeta{
				ID:       pickNonEmpty(input.Meta["id"], ids.UUID()),
				Title:    mergedTitle,
				Severity: pickNonEmpty(severity, input.Meta["severity"]),
				Start:    pickNonEmpty(start, input.Meta["start"]),
				End:      pickNonEmpty(end, input.Meta["end"]),
				Owner:    pickNonEmpty(owner, input.Meta["owner"]),
				Now:      pickNonEmpty(input.Meta["now"], now.Format(time.RFC3339)),
			}

			rendered, err := sections.Merge(manifest, input.Sections, struct {
				Meta postmortemMeta
			}{Meta: meta})
			if err != nil {
				return err
			}

			data := postmortemData{Meta: meta, Sections: rendered}
			def := fmt.Sprintf("postmortem-%s-%s.md", now.Format("2006-01-02"), ids.Slug(meta.Title))
			return render.Render(cmd.OutOrStdout(), loader, "postmortem.md.tmpl", data, out.RenderOptionsStructured(cmd, def))
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

// emitPostmortemSchema marshals the manifest-derived JSON Schema to the
// command's stdout. The schema is recomputed from the loaded manifest on
// every invocation, so user customizations to postmortem.sections.yaml
// flow through automatically.
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
	input, err := readPostmortemInput(cmd, file)
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

// readPostmortemInput loads structured input from --from FILE. Empty `from`
// returns a zero-value input (legacy mode: all sections rendered from
// manifest defaults). `-` reads stdin.
func readPostmortemInput(cmd *cobra.Command, from string) (postmortemInput, error) {
	var input postmortemInput
	if from == "" {
		return input, nil
	}
	var raw []byte
	var err error
	if from == "-" {
		raw, err = io.ReadAll(cmd.InOrStdin())
	} else {
		raw, err = os.ReadFile(from)
	}
	if err != nil {
		return input, fmt.Errorf("read --from %s: %w", from, err)
	}
	if len(raw) == 0 {
		return input, nil
	}
	if err := json.Unmarshal(raw, &input); err != nil {
		return input, fmt.Errorf("parse --from %s: %w", from, err)
	}
	return input, nil
}

// pickNonEmpty returns the first non-empty string from values.
func pickNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
