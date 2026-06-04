package tmpl

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoader_DirOverridesEmbed(t *testing.T) {
	dir := t.TempDir()
	custom := []byte("CUSTOM: {{ .Title }}\n")
	if err := os.WriteFile(filepath.Join(dir, "task.md.tmpl"), custom, 0o644); err != nil {
		t.Fatal(err)
	}
	loader := &Loader{Sources: []Source{DirSource{Dir: dir}, EmbedSource{}}}
	tpl, err := loader.Parse("task.md.tmpl")
	if err != nil {
		t.Fatal(err)
	}
	var buf strings.Builder
	if err := tpl.Execute(&buf, struct{ Title string }{Title: "X"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "CUSTOM: X") {
		t.Fatalf("expected dir template, got embedded: %s", buf.String())
	}
}

func TestLoader_FallbackToEmbed(t *testing.T) {
	// Dir exists but doesn't contain the requested template — should fall through
	// to embed without error.
	dir := t.TempDir()
	loader := &Loader{Sources: []Source{DirSource{Dir: dir}, EmbedSource{}}}
	tpl, err := loader.Parse("changelog.md.tmpl")
	if err != nil {
		t.Fatalf("expected embed fallback, got error: %v", err)
	}
	if tpl == nil {
		t.Fatal("expected non-nil template from embed fallback")
	}
}

func TestLoader_NotFoundInAnySource(t *testing.T) {
	loader := &Loader{Sources: []Source{EmbedSource{}}}
	_, err := loader.Parse("does-not-exist.md.tmpl")
	if err == nil {
		t.Fatal("expected error for missing template")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("error should mention 'not found', got: %v", err)
	}
}

func TestParseFile_AppliesFuncs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "x.tmpl")
	if err := os.WriteFile(path, []byte(`{{ "abc" | upper }}`), 0o644); err != nil {
		t.Fatal(err)
	}
	tpl, err := ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var buf strings.Builder
	if err := tpl.Execute(&buf, nil); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "ABC" {
		t.Fatalf("expected 'ABC', got %q", buf.String())
	}
}

func TestAddDirSource_Prepends(t *testing.T) {
	// Default order is [Embed]. After AddDirSource, it becomes [Dir, Embed].
	loader := &Loader{Sources: []Source{EmbedSource{}}}
	loader.AddDirSource("/tmp/whatever")
	if _, ok := loader.Sources[0].(DirSource); !ok {
		t.Fatalf("expected DirSource first, got %T", loader.Sources[0])
	}
	if _, ok := loader.Sources[1].(EmbedSource); !ok {
		t.Fatalf("expected EmbedSource second, got %T", loader.Sources[1])
	}
}
