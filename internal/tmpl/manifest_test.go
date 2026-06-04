package tmpl

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestManifestNameFor(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in, want string
	}{
		{"postmortem.md.tmpl", "postmortem.sections.yaml"},
		{"license_mit.tmpl", "license_mit.sections.yaml"},
		{"incident.md.tmpl", "incident.sections.yaml"},
		{"already.sections.yaml", "already.sections.yaml.sections.yaml"}, // not idempotent — caller passes template names
	}
	for _, tc := range cases {
		if got := ManifestNameFor(tc.in); got != tc.want {
			t.Errorf("ManifestNameFor(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestIsTemplateArtifact(t *testing.T) {
	t.Parallel()
	cases := map[string]bool{
		"postmortem.md.tmpl":       true,
		"license_mit.tmpl":         true,
		"postmortem.sections.yaml": true,
		"random.yaml":              false,
		"TEMPLATES.md":             false,
		"":                         false,
		".srekit-embedded":         false,
		"postmortem.md.tmpl.bak":   false,
	}
	for name, want := range cases {
		if got := IsTemplateArtifact(name); got != want {
			t.Errorf("IsTemplateArtifact(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestEmbeddedNames_includesTmplAndManifest(t *testing.T) {
	t.Parallel()
	names, err := EmbeddedNames()
	if err != nil {
		t.Fatal(err)
	}
	var hasTmpl, hasYAML bool
	for _, n := range names {
		if strings.HasSuffix(n, ".tmpl") {
			hasTmpl = true
		}
		if strings.HasSuffix(n, ".sections.yaml") {
			hasYAML = true
		}
		if !IsTemplateArtifact(n) {
			t.Errorf("EmbeddedNames returned non-artifact %q", n)
		}
	}
	if !hasTmpl {
		t.Errorf("expected at least one .tmpl in embedded names")
	}
	if !hasYAML {
		// Will be true after postmortem.sections.yaml is added in Step 3.
		t.Logf("note: no .sections.yaml in embed yet — will be added in Step 3")
	}
}

func TestLoadManifestBytes_NotFound(t *testing.T) {
	t.Parallel()
	loader := NewDefaultLoader()
	_, err := loader.LoadManifestBytes("rfc.md.tmpl")
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("want fs.ErrNotExist, got %v", err)
	}
}

func TestLoadManifestBytes_DirOverridesEmbed(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	custom := []byte("version: 1\nsections: [{id: x, title: X, type: text}]\n")
	if err := os.WriteFile(filepath.Join(dir, "rfc.sections.yaml"), custom, 0o644); err != nil {
		t.Fatal(err)
	}
	loader := &Loader{Sources: []Source{DirSource{Dir: dir}, EmbedSource{}}}
	got, err := loader.LoadManifestBytes("rfc.md.tmpl")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if string(got) != string(custom) {
		t.Errorf("dir override not picked up: got %q", got)
	}
}

func TestLoadManifestBytes_FallsThroughEmptyDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	loader := &Loader{Sources: []Source{DirSource{Dir: dir}, EmbedSource{}}}
	_, err := loader.LoadManifestBytes("rfc.md.tmpl")
	// rfc has no manifest in embed, dir is empty → fs.ErrNotExist.
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("want fs.ErrNotExist, got %v", err)
	}
}
