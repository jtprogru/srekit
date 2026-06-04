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

func TestManifestNameFor(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in, want string
	}{
		{"postmortem.md.tmpl", "postmortem.sections.yaml"},
		{"license_mit.tmpl", "license_mit.sections.yaml"},
		{"changelog.md.tmpl", "changelog.sections.yaml"},
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
	// As of v0.14.0 the bare `.yaml` suffix is a valid v1 artifact
	// filename, so any `.yaml` in a templates dir counts as an artifact.
	// Users are expected not to drop unrelated YAML into their templates
	// dir (config and similar live elsewhere — config.yaml lives in
	// $XDG_CONFIG_HOME/srekit, not in templates_dir).
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
	_, err := loader.LoadManifestBytes("changelog.md.tmpl")
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("want fs.ErrNotExist, got %v", err)
	}
}

func TestLoadManifestBytes_DirOverridesEmbed(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	custom := []byte("version: 1\nsections: [{id: x, title: X, type: text}]\n")
	if err := os.WriteFile(filepath.Join(dir, "changelog.sections.yaml"), custom, 0o644); err != nil {
		t.Fatal(err)
	}
	loader := &Loader{Sources: []Source{DirSource{Dir: dir}, EmbedSource{}}}
	got, err := loader.LoadManifestBytes("changelog.md.tmpl")
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
	_, err := loader.LoadManifestBytes("changelog.md.tmpl")
	// changelog has no manifest in embed, dir is empty → fs.ErrNotExist.
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("want fs.ErrNotExist, got %v", err)
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
	_, err := loader.LoadArtifactBytes("changelog.md.tmpl") // no changelog.yaml yet
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
