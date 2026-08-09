package tmpl

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	yaml "go.yaml.in/yaml/v3"
)

func TestIsTemplateArtifact(t *testing.T) {
	t.Parallel()
	// Recognized artifact suffixes. `.tmpl` and `.sections.yaml` are kept
	// for backwards compatibility so user files written against pre-v0.14
	// layouts still appear in `templates list` / `validate`.
	cases := map[string]bool{
		"postmortem.md.tmpl":       true,
		"license_mit.tmpl":         true,
		"postmortem.sections.yaml": true,
		"postmortem.yaml":          true,
		"task.yaml":                true,
		"TEMPLATES.md":             false,
		"":                         false,
		".srekit-embedded":         false,
		"postmortem.md.tmpl.bak":   false,
		"README.md":                false,
	}
	for name, want := range cases {
		if got := IsTemplateArtifact(name); got != want {
			t.Errorf("IsTemplateArtifact(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestEmbeddedNames_isAllArtifactYAML(t *testing.T) {
	t.Parallel()
	names, err := EmbeddedNames()
	if err != nil {
		t.Fatal(err)
	}
	if len(names) == 0 {
		t.Fatal("expected embedded artifacts")
	}
	for _, n := range names {
		if !strings.HasSuffix(n, ".yaml") {
			// As of v0.20.0 every embedded artifact is the v1 single-file
			// YAML format; .tmpl/.sections.yaml were retired across the
			// 0.14–0.20 migration sequence.
			t.Errorf("expected only .yaml artifacts in embed, got %q", n)
		}
		if !IsTemplateArtifact(n) {
			t.Errorf("EmbeddedNames returned non-artifact %q", n)
		}
	}
}

func TestArtifactNameFor(t *testing.T) {
	t.Parallel()
	cases := []struct{ in, want string }{
		// The bare name is what shipped generators pass.
		{"postmortem", "postmortem.yaml"},
		{"task", "task.yaml"},
		// Idempotent: passing the filename back in must not double the suffix.
		{"postmortem.yaml", "postmortem.yaml"},
		// Legacy pre-v1.0 spellings still normalize to the v1 artifact.
		{"postmortem.md.tmpl", "postmortem.yaml"},
		{"license_mit.tmpl", "license_mit.yaml"},
		{"task.md.tmpl", "task.yaml"},
	}
	for _, tc := range cases {
		if got := ArtifactNameFor(tc.in); got != tc.want {
			t.Errorf("ArtifactNameFor(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestArtifactVariantNameFor(t *testing.T) {
	t.Parallel()
	cases := []struct{ name, lang, want string }{
		{"changelog", "ru", "changelog.ru.yaml"},
		// No language is exactly the base artifact.
		{"changelog", "", "changelog.yaml"},
		// Idempotent, including for a name that already carries the segment.
		{"changelog.ru.yaml", "ru", "changelog.ru.yaml"},
		{"changelog.ru.yaml", "", "changelog.ru.yaml"},
		{"changelog.yaml", "ru", "changelog.ru.yaml"},
		// Legacy spellings normalize the same way they do without a language.
		{"changelog.md.tmpl", "ru", "changelog.ru.yaml"},
	}
	for _, tc := range cases {
		if got := ArtifactVariantNameFor(tc.name, tc.lang); got != tc.want {
			t.Errorf("ArtifactVariantNameFor(%q, %q) = %q, want %q", tc.name, tc.lang, got, tc.want)
		}
	}
}

func TestLoadArtifactBytesLang_PrefersVariant(t *testing.T) {
	t.Parallel()
	loader := NewDefaultLoader()
	body, err := loader.LoadArtifactBytesLang("changelog", "ru")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if !bytes.Contains(body, []byte("Добавлено")) {
		t.Errorf("expected the Russian variant, got:\n%s", body)
	}
}

func TestLoadArtifactBytesLang_FallsBackSilently(t *testing.T) {
	t.Parallel()
	loader := NewDefaultLoader()
	// No artifact ships a `de` variant; the request must resolve to the base
	// artifact rather than fail.
	body, err := loader.LoadArtifactBytesLang("changelog", "de")
	if err != nil {
		t.Fatalf("want silent fallback, got error: %v", err)
	}
	if !bytes.Contains(body, []byte("### Added")) {
		t.Errorf("expected the base artifact, got:\n%s", body)
	}
}

func TestLoadArtifactBytesLang_DirShadowsEmbeddedVariant(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	custom := []byte("version: 1\nsections: [{id: x, title: X, type: text}]\n")
	if err := os.WriteFile(filepath.Join(dir, "changelog.ru.yaml"), custom, 0o644); err != nil {
		t.Fatal(err)
	}
	loader := &Loader{Sources: []Source{DirSource{Dir: dir}, EmbedSource{}}}
	got, err := loader.LoadArtifactBytesLang("changelog", "ru")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(custom) {
		t.Errorf("the user directory's variant should win, got:\n%s", got)
	}
}

func TestLoadArtifactBytesLang_VariantLookupPrecedesFallback(t *testing.T) {
	t.Parallel()
	// The user customized the base artifact but ships no Russian variant.
	// The request is for Russian, and their English file is not an answer to
	// that question — the embedded variant is.
	dir := t.TempDir()
	custom := []byte("version: 1\nsections: [{id: x, title: X, type: text}]\n")
	if err := os.WriteFile(filepath.Join(dir, "changelog.yaml"), custom, 0o644); err != nil {
		t.Fatal(err)
	}
	loader := &Loader{Sources: []Source{DirSource{Dir: dir}, EmbedSource{}}}
	got, err := loader.LoadArtifactBytesLang("changelog", "ru")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(got, []byte("Добавлено")) {
		t.Errorf("expected the embedded Russian variant, got:\n%s", got)
	}
	// Without a language the same loader still serves their customization.
	base, err := loader.LoadArtifactBytes("changelog")
	if err != nil {
		t.Fatal(err)
	}
	if string(base) != string(custom) {
		t.Errorf("the base lookup should still prefer the user directory, got:\n%s", base)
	}
}

func TestLoadArtifactBytesLang_TraversalNamesStayUnresolved(t *testing.T) {
	t.Parallel()
	// DirSource refuses anything that is not a bare filename; adding a
	// language segment must not open a path through it.
	outer := t.TempDir()
	if err := os.WriteFile(filepath.Join(outer, "secret.ru.yaml"), []byte("version: 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(outer, "templates")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	loader := &Loader{Sources: []Source{DirSource{Dir: dir}}}
	if _, err := loader.LoadArtifactBytesLang("../secret", "ru"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("want fs.ErrNotExist for a traversal name, got %v", err)
	}
}

func TestChangelogVariantsShareSectionIDs(t *testing.T) {
	t.Parallel()
	// The two files duplicate each other's structure by design; only prose
	// and change types differ. A section added to one and forgotten in the
	// other is caught here rather than by a user.
	loader := NewDefaultLoader()
	base, err := loader.LoadArtifactBytes("changelog")
	if err != nil {
		t.Fatal(err)
	}
	variant, err := loader.LoadArtifactBytesLang("changelog", "ru")
	if err != nil {
		t.Fatal(err)
	}
	got, want := sectionIDs(t, variant), sectionIDs(t, base)
	if len(got) != len(want) {
		t.Fatalf("section ids differ: %v vs %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("section id %d: changelog.yaml has %q, changelog.ru.yaml has %q", i, want[i], got[i])
		}
	}
}

// sectionIDs pulls the ordered section ids out of a v1 artifact without
// importing internal/sections, which imports this package.
func sectionIDs(t *testing.T, body []byte) []string {
	t.Helper()
	var a struct {
		Sections []struct {
			ID string `yaml:"id"`
		} `yaml:"sections"`
	}
	if err := yaml.Unmarshal(body, &a); err != nil {
		t.Fatal(err)
	}
	out := make([]string, 0, len(a.Sections))
	for _, s := range a.Sections {
		out = append(out, s.ID)
	}
	return out
}

func TestLoadArtifactBytes_FindsEmbeddedPostmortem(t *testing.T) {
	t.Parallel()
	loader := NewDefaultLoader()
	body, err := loader.LoadArtifactBytes("postmortem")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if !bytes.Contains(body, []byte("version: 1")) {
		t.Errorf("body doesn't look like v1 artifact: %s", body[:min(200, len(body))])
	}
}

func TestLoadArtifactBytes_NotFound(t *testing.T) {
	t.Parallel()
	loader := NewDefaultLoader()
	// All embedded artifacts ship as .yaml as of v0.20.0; an artifact name
	// that doesn't correspond to any shipped name must miss across all
	// sources and surface fs.ErrNotExist.
	_, err := loader.LoadArtifactBytes("does-not-exist")
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("want fs.ErrNotExist, got %v", err)
	}
}

func TestLoadArtifactBytes_DirOverridesEmbed(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	custom := []byte("version: 1\nsections: [{id: x, title: X, type: text}]\n")
	if err := os.WriteFile(filepath.Join(dir, "postmortem.yaml"), custom, 0o644); err != nil {
		t.Fatal(err)
	}
	loader := &Loader{Sources: []Source{DirSource{Dir: dir}, EmbedSource{}}}
	got, err := loader.LoadArtifactBytes("postmortem")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(custom) {
		t.Errorf("dir override not picked up")
	}
}
