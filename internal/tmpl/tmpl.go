package tmpl

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"
	"text/template"
	"time"

	"github.com/jtprogru/srekit/internal/clock"
	"github.com/jtprogru/srekit/internal/ids"
)

//go:embed templates/*.tmpl
var FS embed.FS

// Funcs is the template function map shared by every parsed template.
// It is exported so tests can inspect or extend it; production code parses
// through Parse, which applies it automatically.
//
//nolint:gochecknoglobals // intentional package-level registry shared by every Parse() call
var Funcs = template.FuncMap{
	"default": func(def, val string) string {
		if val == "" {
			return def
		}
		return val
	},
	"shortID":  ids.Short,
	"slugify":  ids.Slug,
	"upper":    strings.ToUpper,
	"lower":    strings.ToLower,
	"trim":     strings.TrimSpace,
	"now": func(format ...string) string {
		f := time.RFC3339
		if len(format) > 0 && format[0] != "" {
			f = format[0]
		}
		return clock.Now().Format(f)
	},
}

func Parse(name string) (*template.Template, error) {
	b, err := fs.ReadFile(FS, "templates/"+name)
	if err != nil {
		return nil, fmt.Errorf("template %q: %w", name, err)
	}
	t, err := template.New(name).Funcs(Funcs).Parse(string(b))
	if err != nil {
		return nil, fmt.Errorf("parse template %q: %w", name, err)
	}
	return t, nil
}
