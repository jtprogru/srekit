package render

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jtprogru/srekit/internal/sections"
	"github.com/jtprogru/srekit/internal/tmpl"
)

type Options struct {
	Out     string
	Stdout  bool
	Force   bool
	DryRun  bool
	Default string
	JSON    bool // emit the caller's structured payload as JSON instead of rendering markdown
	Quiet   bool // suppress informational messages (the "wrote <file>" line, dry-run notes)
}

// ArtifactPayload is implemented by cmd-level data structs that join the
// v1 artifact render path (currently: postmortem). It lets the render
// pipeline extract the pre-merged section list and the per-template ctx
// without coupling to cmd-specific types.
type ArtifactPayload interface {
	ArtifactPayload() (sections []sections.RenderedSection, ctx any)
}

// Render builds the document for name and routes it per opts.
//
// name is the bare artifact name — "slo" resolves
// internal/tmpl/templates/slo.yaml through the loader. The Go-template
// execution branch this function used to have was removed in v0.30.0
// along with `--template FILE` and the `license` command, its last
// caller: the v1 artifact format is the only render path.
func Render(stdout io.Writer, loader *tmpl.Loader, name string, data any, opts Options) error {
	body, err := buildBody(loader, name, data, opts)
	if err != nil {
		return err
	}
	return writeBody(stdout, body, opts)
}

func buildBody(loader *tmpl.Loader, name string, data any, opts Options) ([]byte, error) {
	// JSON path: the caller has already shaped data as {meta, sections};
	// marshal as-is without rendering markdown.
	if opts.JSON {
		b, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("encode %q as JSON: %w", name, err)
		}
		return append(b, '\n'), nil
	}

	return renderArtifactPath(loader, name, data)
}

// renderArtifactPath loads the v1 single-file YAML artifact for name,
// parses it, extracts the pre-merged section list + ctx from data (which
// must implement ArtifactPayload), and composes the markdown via
// sections.RenderArtifact.
func renderArtifactPath(loader *tmpl.Loader, name string, data any) ([]byte, error) {
	artifactBytes, err := loader.LoadArtifactBytes(name)
	if err != nil {
		return nil, fmt.Errorf("load artifact for %q: %w", name, err)
	}
	artifact, err := sections.ParseArtifact(artifactBytes)
	if err != nil {
		return nil, err
	}
	payload, ok := data.(ArtifactPayload)
	if !ok {
		return nil, fmt.Errorf("data type %T does not implement ArtifactPayload", data)
	}
	rendered, ctx := payload.ArtifactPayload()
	return sections.RenderArtifact(artifact, rendered, ctx)
}

// extractH1 and the bootstrap-JSON envelope it fed were removed in
// v0.30.0. The envelope existed so a generator without a sections
// manifest could still emit the uniform JSON contract; every generator
// has had one since v0.20.0, so nothing had set BootstrapJSON for three
// minor lines.

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
