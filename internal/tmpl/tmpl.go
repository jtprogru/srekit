package tmpl

import (
	"embed"
	"fmt"
	"io/fs"
	"text/template"
)

//go:embed templates/*.tmpl
var FS embed.FS

func Parse(name string) (*template.Template, error) {
	b, err := fs.ReadFile(FS, "templates/"+name)
	if err != nil {
		return nil, fmt.Errorf("template %q: %w", name, err)
	}
	t, err := template.New(name).Parse(string(b))
	if err != nil {
		return nil, fmt.Errorf("parse template %q: %w", name, err)
	}
	return t, nil
}
