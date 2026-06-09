package tmpl

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/jtprogru/srekit/internal/clock"
	"github.com/jtprogru/srekit/internal/ids"
)

// FS is the embedded template/artifact filesystem. As of v0.20.0 every
// shipped artifact is the v1 single-file YAML format (`<name>.yaml`).
// Pre-v1.0 layouts (`.tmpl`, `.sections.yaml`) are still recognized by
// IsTemplateArtifact so user files written against them remain visible
// in `templates list` / `validate` and can be auto-converted via
// `srekit templates migrate`.
//
//go:embed templates/*.yaml
var FS embed.FS

// DocsMD is the placeholder/FuncMap reference shipped to user templates
// directories by `srekit templates init`.
//
//go:embed TEMPLATES.md
var DocsMD []byte

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
	"shortID": ids.Short,
	"slugify": ids.Slug,
	"upper":   strings.ToUpper,
	"lower":   strings.ToLower,
	"trim":    strings.TrimSpace,
	"now": func(format ...string) string {
		f := time.RFC3339
		if len(format) > 0 && format[0] != "" {
			f = format[0]
		}
		return clock.Now().Format(f)
	},
}

// Source resolves a template name to its raw bytes. Returning fs.ErrNotExist
// signals "not in this source, try the next" — any other error is fatal.
type Source interface {
	Read(name string) ([]byte, error)
}

// EmbedSource serves the built-in templates compiled into the binary.
type EmbedSource struct{}

func (EmbedSource) Read(name string) ([]byte, error) {
	b, err := fs.ReadFile(FS, "templates/"+name)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		return nil, fs.ErrNotExist
	}
	return b, err
}

// DirSource serves templates from a local filesystem directory. Missing
// files are reported as fs.ErrNotExist so Loader can fall through; other
// I/O errors bubble up.
type DirSource struct{ Dir string }

func (d DirSource) Read(name string) ([]byte, error) {
	if filepath.Base(name) != name || strings.ContainsAny(name, `/\`) {
		return nil, fs.ErrNotExist
	}
	b, err := os.ReadFile(filepath.Join(d.Dir, name))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fs.ErrNotExist
		}
		return nil, err
	}
	return b, nil
}

// Loader resolves templates by trying Sources in order; first hit wins.
type Loader struct{ Sources []Source }

// NewDefaultLoader returns a Loader backed only by the embedded templates.
// cmd builds one per command tree and, when --templates-dir is set, prepends
// a DirSource. Each call returns a fresh value — no shared package state.
func NewDefaultLoader() *Loader {
	return &Loader{Sources: []Source{EmbedSource{}}}
}

// Parse resolves name through Sources in order and returns a parsed
// template with Funcs applied. fs.ErrNotExist on a source falls through
// to the next; any other source error is returned as-is.
func (l *Loader) Parse(name string) (*template.Template, error) {
	for _, s := range l.Sources {
		b, err := s.Read(name)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("template %q: %w", name, err)
		}
		t, err := template.New(name).Funcs(Funcs).Parse(string(b))
		if err != nil {
			return nil, fmt.Errorf("parse template %q: %w", name, err)
		}
		return t, nil
	}
	return nil, fmt.Errorf("template %q not found in any source", name)
}

// AddDirSource prepends a DirSource so user-dir templates take priority
// over the embedded fallback.
func (l *Loader) AddDirSource(dir string) {
	l.Sources = append([]Source{DirSource{Dir: dir}}, l.Sources...)
}

// IsTemplateArtifact reports whether a filename is one of the artifact
// types srekit recognizes in the embedded FS or a user templates dir.
// Shipped artifacts are all `.yaml` as of v0.20.0; legacy `.tmpl` and
// `.sections.yaml` are still recognized so `templates list` / `validate`
// surfaces user files written against pre-v1.0 layouts.
func IsTemplateArtifact(name string) bool {
	return strings.HasSuffix(name, ".tmpl") ||
		strings.HasSuffix(name, ".sections.yaml") ||
		strings.HasSuffix(name, ".yaml")
}

// ArtifactNameFor maps a template filename (e.g. "postmortem.md.tmpl") to
// the v1 unified artifact filename ("postmortem.yaml"). Both `.md.tmpl`
// and the bare `.tmpl` suffix are stripped, so the rule applies uniformly
// to all template families.
func ArtifactNameFor(templateName string) string {
	name := strings.TrimSuffix(templateName, ".tmpl")
	name = strings.TrimSuffix(name, ".md")
	return name + ".yaml"
}

// LoadArtifactBytes resolves the v1 single-file artifact for a template
// and returns its raw YAML bytes. Walks Sources in the same order as
// Parse, so a user-dir artifact shadows the embedded one. Returns
// fs.ErrNotExist when no source has it.
//
// Parsing is left to internal/sections.ParseArtifact so this package
// doesn't depend on artifact types.
func (l *Loader) LoadArtifactBytes(templateName string) ([]byte, error) {
	artifactName := ArtifactNameFor(templateName)
	for _, s := range l.Sources {
		b, err := s.Read(artifactName)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("artifact %q: %w", artifactName, err)
		}
		return b, nil
	}
	return nil, fs.ErrNotExist
}

// EmbeddedNames returns the filenames of every artifact embedded under
// `templates/`. As of v0.20.0 all embedded artifacts are v1 `.yaml`. This
// is the single enumeration surface used by
// `srekit templates init/upgrade/diff/list`.
func EmbeddedNames() ([]string, error) {
	entries, err := fs.ReadDir(FS, "templates")
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !IsTemplateArtifact(e.Name()) {
			continue
		}
		out = append(out, e.Name())
	}
	return out, nil
}

// ParseFile reads a template from an arbitrary file path and parses it
// with Funcs applied. Used to implement the per-command --template flag.
func ParseFile(path string) (*template.Template, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read template %s: %w", path, err)
	}
	name := filepath.Base(path)
	t, err := template.New(name).Funcs(Funcs).Parse(string(b))
	if err != nil {
		return nil, fmt.Errorf("parse template %s: %w", path, err)
	}
	return t, nil
}
