package render

import (
	"bytes"
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
