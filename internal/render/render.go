package render

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jtprogru/srekit/internal/tmpl"
)

type Options struct {
	Out     string
	Stdout  bool
	Force   bool
	DryRun  bool
	Default string
}

func Render(stdout io.Writer, tmplName string, data any, opts Options) error {
	t, err := tmpl.Parse(tmplName)
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

	if err := os.WriteFile(target, buf.Bytes(), 0o644); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "wrote %s\n", target)
	return nil
}
