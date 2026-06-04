package render

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/jtprogru/srekit/internal/sections"
	"github.com/jtprogru/srekit/internal/tmpl"
)

type Options struct {
	Out          string
	Stdout       bool
	Force        bool
	DryRun       bool
	Default      string
	TemplatePath string // optional: read template from this file path instead of the embedded/loader chain
	JSON         bool   // emit the template data as JSON instead of rendering the template
	Quiet        bool   // suppress informational messages (the "wrote <file>" line, dry-run notes)
	// BootstrapJSON controls --json shape for commands that haven't migrated
	// to a sections manifest. When true (the default for legacy commands),
	// the rendered markdown is wrapped in a bootstrap envelope
	// `{meta: <data>, sections: [{id:"body", title:<H1>, type:"text",
	// required:true, body:<rendered markdown>}]}` so every command speaks
	// the same {meta, sections} JSON contract. When false (postmortem), the
	// caller has already shaped `data` as {meta, sections} itself and JSON
	// short-circuits straight to MarshalIndent without rendering markdown.
	BootstrapJSON bool
	// RenderArtifact switches the render path from Go-template execution
	// to the v1 YAML artifact format introduced in v0.14.0. When true,
	// the loader resolves <name>.yaml via LoadArtifactBytes, ParseArtifact
	// builds an Artifact, and sections.RenderArtifact composes the
	// markdown. `data` must implement ArtifactPayload (returning the
	// pre-merged section list and the template ctx).
	RenderArtifact bool
}

// ArtifactPayload is implemented by cmd-level data structs that join the
// v1 artifact render path (currently: postmortem). It lets the render
// pipeline extract the pre-merged section list and the per-template ctx
// without coupling to cmd-specific types.
type ArtifactPayload interface {
	ArtifactPayload() (sections []sections.RenderedSection, ctx any)
}

func Render(stdout io.Writer, loader *tmpl.Loader, tmplName string, data any, opts Options) error {
	body, err := buildBody(loader, tmplName, data, opts)
	if err != nil {
		return err
	}
	return writeBody(stdout, body, opts)
}

// WriteRaw applies the standard --out / --stdout / --force / --dry-run /
// --quiet routing to an already-rendered body. Used by commands that
// render their content inline (currently: license, whose template bodies
// are Go constants rather than embedded files).
func WriteRaw(stdout io.Writer, body []byte, opts Options) error {
	return writeBody(stdout, body, opts)
}

func buildBody(loader *tmpl.Loader, tmplName string, data any, opts Options) ([]byte, error) {
	// Structured JSON path: the caller has already shaped data as
	// {meta, sections}; marshal as-is without touching the template.
	if opts.JSON && !opts.BootstrapJSON {
		b, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("encode %q as JSON: %w", tmplName, err)
		}
		return append(b, '\n'), nil
	}

	var body []byte
	if opts.RenderArtifact {
		var err error
		body, err = renderArtifactPath(loader, tmplName, data)
		if err != nil {
			return nil, err
		}
	} else {
		var t *template.Template
		var err error
		if opts.TemplatePath != "" {
			t, err = tmpl.ParseFile(opts.TemplatePath)
		} else {
			t, err = loader.Parse(tmplName)
		}
		if err != nil {
			return nil, err
		}
		var buf bytes.Buffer
		if err := t.Execute(&buf, data); err != nil {
			return nil, fmt.Errorf("render %q: %w", tmplName, err)
		}
		body = buf.Bytes()
	}

	// Bootstrap JSON path: render markdown first, then wrap the result in
	// a {meta, sections:[{id:"body", ...}]} envelope so every generator
	// command exposes a uniform JSON contract regardless of whether it has
	// a sections manifest yet.
	if opts.JSON && opts.BootstrapJSON {
		envelope := map[string]any{
			"meta": data,
			"sections": []map[string]any{{
				"id":       "body",
				"title":    extractH1(body),
				"type":     "text",
				"required": true,
				"body":     string(body),
			}},
		}
		b, err := json.MarshalIndent(envelope, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("encode %q as JSON: %w", tmplName, err)
		}
		return append(b, '\n'), nil
	}

	return body, nil
}

// renderArtifactPath loads the v1 single-file YAML artifact for tmplName,
// parses it, extracts the pre-merged section list + ctx from data (which
// must implement ArtifactPayload), and composes the markdown via
// sections.RenderArtifact. This is the v0.14.0+ render path for commands
// that have migrated to the YAML format.
func renderArtifactPath(loader *tmpl.Loader, tmplName string, data any) ([]byte, error) {
	artifactBytes, err := loader.LoadArtifactBytes(tmplName)
	if err != nil {
		return nil, fmt.Errorf("load artifact for %q: %w", tmplName, err)
	}
	artifact, err := sections.ParseArtifact(artifactBytes)
	if err != nil {
		return nil, err
	}
	payload, ok := data.(ArtifactPayload)
	if !ok {
		return nil, fmt.Errorf("RenderArtifact set but data type %T does not implement ArtifactPayload", data)
	}
	rendered, ctx := payload.ArtifactPayload()
	return sections.RenderArtifact(artifact, rendered, ctx)
}

var h1Pattern = regexp.MustCompile(`(?m)^#\s+(.+)$`)

// extractH1 returns the text of the first level-1 heading in body, or "" if
// the document has none. Used to populate the synthetic `body` section's
// title in the bootstrap JSON envelope.
func extractH1(body []byte) string {
	m := h1Pattern.FindSubmatch(body)
	if m == nil {
		return ""
	}
	return strings.TrimSpace(string(m[1]))
}

func writeBody(stdout io.Writer, body []byte, opts Options) error {
	target := opts.Out
	// JSON output never falls through to the markdown default path (which
	// has a .md suffix and would land in the project tree). If the user
	// passed --json without --out, write to stdout.
	if target == "" && !opts.Stdout && !opts.JSON {
		target = opts.Default
	}

	// "-" is the conventional stand-in for stdout (so `--out -` works in
	// pipelines without inventing a separate flag).
	if opts.Stdout || target == "" || target == "-" {
		if opts.DryRun && !opts.Quiet {
			fmt.Fprintln(stdout, "# dry-run: would write to stdout")
		}
		_, err := stdout.Write(body)
		return err
	}

	if opts.DryRun {
		if !opts.Quiet {
			fmt.Fprintf(stdout, "# dry-run: would write %d bytes to %s\n", len(body), target)
		}
		_, err := stdout.Write(body)
		return err
	}

	if _, err := os.Stat(target); err == nil && !opts.Force {
		return fmt.Errorf("file %s already exists; use --force to overwrite", target)
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if dir := filepath.Dir(target); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	// Generated artifacts are public docs (README, CHANGELOG, runbooks, RFCs);
	// 0o644 matches the convention every other CLI generator uses.
	if err := os.WriteFile(target, body, 0o644); err != nil { //nolint:gosec // G306: see comment above
		return err
	}
	if !opts.Quiet {
		fmt.Fprintf(stdout, "wrote %s\n", target)
	}
	return nil
}
