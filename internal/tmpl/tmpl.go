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

// Funcs is the template function map shared by every parsed template:
// section bodies and frontmatter evaluated by internal/sections, and the
// `.tmpl` files `srekit templates validate` parses. It is exported so
// tests can inspect or extend it.
//
//nolint:gochecknoglobals // intentional package-level registry shared by every template parse
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

// Loader.Parse was removed in v0.30.0. It returned a parsed
// text/template, which only the render package's Go-template branch ever
// wanted; that branch went with `--template FILE` and the `license`
// command. Artifacts are resolved as raw bytes via LoadArtifactBytes and
// parsed by internal/sections.

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

// ArtifactNameFor maps an artifact name to its v1 filename: "postmortem"
// and "postmortem.yaml" both yield "postmortem.yaml". It is idempotent, so
// callers may pass either spelling.
//
// The pre-v1.0 template filenames are also accepted (`.md.tmpl` and the
// bare `.tmpl` suffix are stripped) so external tooling built against the
// v0.1x names keeps resolving to the right artifact. Shipped generators
// pass the bare name.
func ArtifactNameFor(name string) string {
	name = strings.TrimSuffix(name, ".tmpl")
	name = strings.TrimSuffix(name, ".md")
	name = strings.TrimSuffix(name, ".yaml")
	return name + ".yaml"
}

// ArtifactVariantNameFor maps an artifact name and a language tag to the
// variant filename: ("changelog", "ru") yields "changelog.ru.yaml". An
// empty lang yields exactly what ArtifactNameFor would.
//
// Like ArtifactNameFor it is idempotent, including for a name that already
// carries the segment: "changelog.ru.yaml" in Russian stays
// "changelog.ru.yaml" rather than growing a second one.
func ArtifactVariantNameFor(name, lang string) string {
	base := strings.TrimSuffix(ArtifactNameFor(name), ".yaml")
	if lang == "" {
		return base + ".yaml"
	}
	base = strings.TrimSuffix(base, "."+lang)
	return base + "." + lang + ".yaml"
}

// LoadArtifactBytes resolves the v1 single-file artifact named by name
// (the bare artifact name, e.g. "postmortem") and returns its raw YAML
// bytes. Walks Sources in order, so a user-dir artifact shadows the
// embedded one. Returns fs.ErrNotExist when no source has it.
//
// The name is normalized through ArtifactNameFor, so legacy spellings like
// "postmortem.md.tmpl" resolve to the same artifact.
//
// Parsing is left to internal/sections.ParseArtifact so this package
// doesn't depend on artifact types.
func (l *Loader) LoadArtifactBytes(name string) ([]byte, error) {
	return l.LoadArtifactBytesLang(name, "")
}

// LoadArtifactBytesLang resolves the artifact named by name in the
// requested language. An empty lang behaves exactly like
// LoadArtifactBytes.
//
// The variant lookup runs across the whole source chain before the
// fallback does, not per source. Given a templates directory holding a
// customized `changelog.yaml` and no Russian variant, that yields the
// embedded `changelog.ru.yaml`: the user asked for Russian, and an
// English file — even their own — is not an answer to that question.
// Per-source ordering would silently ignore the language whenever the
// base artifact had been customized, which is the worst failure mode
// available, since the selection would appear to work and produce the
// other language.
//
// A language with no variant in any source falls back silently. That
// makes "requested a language nobody ships a variant for" indistinguishable
// from "requested no language", which is what keeps the rule usable for
// artifacts that have no variants at all.
func (l *Loader) LoadArtifactBytesLang(name, lang string) ([]byte, error) {
	if lang != "" {
		b, err := l.readFromSources(ArtifactVariantNameFor(name, lang))
		if err == nil {
			return b, nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
	}
	return l.readFromSources(ArtifactNameFor(name))
}

// readFromSources walks the chain for one concrete filename, treating
// fs.ErrNotExist as fall-through and returning it when every source misses.
func (l *Loader) readFromSources(artifactName string) ([]byte, error) {
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

// ParseFile was removed in v0.30.0 together with `--template FILE`, the
// only feature that read a template from an arbitrary path. Artifacts are
// resolved by name through the Loader's source chain.
