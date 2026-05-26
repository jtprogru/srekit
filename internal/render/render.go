package render

import (
	"bytes"
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
}

func Render(stdout io.Writer, tmplName string, data any, opts Options) error {
	var t *template.Template
	var err error
	if opts.TemplatePath != "" {
		t, err = tmpl.ParseFile(opts.TemplatePath)
	} else {
		t, err = tmpl.Parse(tmplName)
	}
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return fmt.Errorf("render %q: %w", tmplName, err)
	}

	target := opts.Out
	if target == "" && !opts.Stdout {
		target = opts.Default
	}

	if opts.Stdout || target == "" {
		if opts.DryRun {
			fmt.Fprintln(stdout, "# dry-run: would write to stdout")
		}
		_, err := io.Copy(stdout, &buf)
		return err
	}

	if opts.DryRun {
		fmt.Fprintf(stdout, "# dry-run: would write %d bytes to %s\n", buf.Len(), target)
		_, err := io.Copy(stdout, &buf)
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
	if err := os.WriteFile(target, buf.Bytes(), 0o644); err != nil { //nolint:gosec // G306: see comment above
		return err
	}
	fmt.Fprintf(stdout, "wrote %s\n", target)
	return nil
}
