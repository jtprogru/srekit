package render

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/template"

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
}

func Render(stdout io.Writer, tmplName string, data any, opts Options) error {
	body, err := buildBody(tmplName, data, opts)
	if err != nil {
		return err
	}
	return writeBody(stdout, body, opts)
}

func buildBody(tmplName string, data any, opts Options) ([]byte, error) {
	if opts.JSON {
		b, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("encode %q as JSON: %w", tmplName, err)
		}
		return append(b, '\n'), nil
	}

	var t *template.Template
	var err error
	if opts.TemplatePath != "" {
		t, err = tmpl.ParseFile(opts.TemplatePath)
	} else {
		t, err = tmpl.Parse(tmplName)
	}
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("render %q: %w", tmplName, err)
	}
	return buf.Bytes(), nil
}

func writeBody(stdout io.Writer, body []byte, opts Options) error {
	target := opts.Out
	// JSON output never falls through to the markdown default path (which
	// has a .md suffix and would land in the project tree). If the user
	// passed --json without --out, write to stdout.
	if target == "" && !opts.Stdout && !opts.JSON {
		target = opts.Default
	}

	if opts.Stdout || target == "" {
		if opts.DryRun {
			fmt.Fprintln(stdout, "# dry-run: would write to stdout")
		}
		_, err := stdout.Write(body)
		return err
	}

	if opts.DryRun {
		fmt.Fprintf(stdout, "# dry-run: would write %d bytes to %s\n", len(body), target)
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
	fmt.Fprintf(stdout, "wrote %s\n", target)
	return nil
}
