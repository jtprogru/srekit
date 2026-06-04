package tmpl

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
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

func TestLoadArtifactBytes_FindsEmbeddedPostmortem(t *testing.T) {
	t.Parallel()
	loader := NewDefaultLoader()
	body, err := loader.LoadArtifactBytes("postmortem.md.tmpl")
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
	_, err := loader.LoadArtifactBytes("does-not-exist.md.tmpl")
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
	got, err := loader.LoadArtifactBytes("postmortem.md.tmpl")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(custom) {
		t.Errorf("dir override not picked up")
	}
}
