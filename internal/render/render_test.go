package render

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderStdout(t *testing.T) {
	var out bytes.Buffer
	err := Render(&out, "task.md.tmpl", struct {
		ID, CreationDate, ModificationDate, Title string
	}{"id-1", "2026-01-01T00:00:00", "2026-01-01T00:00:00", "Hello"}, Options{Stdout: true})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "# Расследование (Investigation) — Hello") {
		t.Fatalf("missing title in output: %s", out.String())
	}
}

func TestRenderToFile(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "sub", "out.md")
	var out bytes.Buffer

	err := Render(&out, "task.md.tmpl", struct {
		ID, CreationDate, ModificationDate, Title string
	}{"id-1", "t", "t", "Foo"}, Options{Out: target})
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "Foo") {
		t.Fatalf("file missing content")
	}
}

func TestRenderRefusesOverwrite(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "out.md")
	if err := os.WriteFile(target, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	err := Render(&out, "task.md.tmpl", struct {
		ID, CreationDate, ModificationDate, Title string
	}{"id", "t", "t", "x"}, Options{Out: target})
	if err == nil {
		t.Fatal("expected error on existing file without --force")
	}
}

func TestRenderForceOverwrite(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "out.md")
	_ = os.WriteFile(target, []byte("old"), 0o644)
	var out bytes.Buffer
	err := Render(&out, "task.md.tmpl", struct {
		ID, CreationDate, ModificationDate, Title string
	}{"id", "t", "t", "Force"}, Options{Out: target, Force: true})
	if err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(target)
	if !strings.Contains(string(b), "Force") {
		t.Fatalf("file not overwritten")
	}
}

// TestRenderFilePermissions locks in the #9 fix: generated docs land at
// 0o644, not 0o600. These are public artifacts (LICENSE, CHANGELOG, etc.),
// not secrets.
func TestRenderFilePermissions(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "out.md")
	var out bytes.Buffer
	err := Render(&out, "task.md.tmpl", struct {
		ID, CreationDate, ModificationDate, Title string
	}{"id", "t", "t", "Perms"}, Options{Out: target})
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o644 {
		t.Fatalf("expected mode 0o644, got %o", perm)
	}
}

// TestRenderTemplatePathOverride verifies that opts.TemplatePath bypasses
// the embedded/loader chain entirely and reads the named file directly.
func TestRenderTemplatePathOverride(t *testing.T) {
	dir := t.TempDir()
	tmplFile := filepath.Join(dir, "custom.tmpl")
	if err := os.WriteFile(tmplFile, []byte("CUSTOM-OVERRIDE: {{ .Title }}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	err := Render(&out, "task.md.tmpl", struct{ Title string }{Title: "Hi"},
		Options{Stdout: true, TemplatePath: tmplFile})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "CUSTOM-OVERRIDE: Hi") {
		t.Fatalf("template override not applied: %s", out.String())
	}
}

// TestRenderJSONStdout verifies --json short-circuits the template and emits
// MarshalIndent'd JSON to stdout when --out/--stdout are unset.
func TestRenderJSONStdout(t *testing.T) {
	var out bytes.Buffer
	type payload struct {
		ID    string
		Title string
	}
	in := payload{ID: "abc", Title: "Hello"}
	err := Render(&out, "task.md.tmpl", in, Options{JSON: true, Default: "ignored.md"})
	if err != nil {
		t.Fatal(err)
	}
	var got payload
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out.String())
	}
	if got != in {
		t.Fatalf("round-trip mismatch: %+v vs %+v", got, in)
	}
	if !strings.Contains(out.String(), "\n  \"ID\"") {
		t.Fatalf("expected MarshalIndent (2-space) output, got: %s", out.String())
	}
}

// TestRenderJSONToFile verifies --json honors --out.
func TestRenderJSONToFile(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "out.json")
	var out bytes.Buffer
	err := Render(&out, "task.md.tmpl",
		struct{ Title string }{Title: "Foo"},
		Options{Out: target, JSON: true},
	)
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("file is not valid JSON: %v\n%s", err, string(b))
	}
	if got["Title"] != "Foo" {
		t.Fatalf("Title mismatch: %v", got["Title"])
	}
}

// TestRenderJSONIgnoresDefaultPath locks in the rule that --json without
// --out/--stdout sinks to stdout rather than the markdown Default path,
// so a JSON payload never ends up in a .md file by accident.
func TestRenderJSONIgnoresDefaultPath(t *testing.T) {
	dir := t.TempDir()
	defaultPath := filepath.Join(dir, "should-not-be-written.md")
	var out bytes.Buffer
	err := Render(&out, "task.md.tmpl",
		struct{ Title string }{Title: "X"},
		Options{JSON: true, Default: defaultPath},
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(defaultPath); !os.IsNotExist(err) {
		t.Fatalf("--json must not write to the markdown default path: %s", defaultPath)
	}
	if !strings.Contains(out.String(), "\"Title\"") {
		t.Fatalf("expected JSON on stdout, got: %s", out.String())
	}
}

func TestRenderDryRunNoFile(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "out.md")
	var out bytes.Buffer
	err := Render(&out, "task.md.tmpl", struct {
		ID, CreationDate, ModificationDate, Title string
	}{"id", "t", "t", "Dry"}, Options{Out: target, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("dry-run must not create file")
	}
	if !strings.Contains(out.String(), "dry-run") {
		t.Fatalf("dry-run header missing")
	}
}
