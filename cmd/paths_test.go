package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// These tests mutate process env (HOME/XDG_CONFIG_HOME) via t.Setenv, so they
// cannot run in parallel.

func TestResolveConfigPath_XDGOnFreshInstall(t *testing.T) {
	home := t.TempDir()
	xdg := filepath.Join(home, "xdg-config")
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", xdg)

	want := filepath.Join(xdg, "srekit", "config.yaml")
	if got := resolveConfigPath(); got != want {
		t.Errorf("fresh install: got %q, want %q", got, want)
	}
}

func TestResolveConfigPath_LegacyWins(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "xdg-config"))

	legacy := filepath.Join(home, ".srekit.yaml")
	if err := os.WriteFile(legacy, []byte("author: x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := resolveConfigPath(); got != legacy {
		t.Errorf("legacy present: got %q, want %q", got, legacy)
	}
}

func TestResolveDefaultTemplatesDir_XDGOnFreshInstall(t *testing.T) {
	home := t.TempDir()
	xdg := filepath.Join(home, "xdg-config")
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", xdg)

	want := filepath.Join(xdg, "srekit", "templates")
	if got := resolveDefaultTemplatesDir(); got != want {
		t.Errorf("fresh install: got %q, want %q", got, want)
	}
}

func TestResolveDefaultTemplatesDir_LegacyWins(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "xdg-config"))

	legacy := filepath.Join(home, ".srekit", "templates")
	if err := os.MkdirAll(legacy, 0o755); err != nil {
		t.Fatal(err)
	}
	if got := resolveDefaultTemplatesDir(); got != legacy {
		t.Errorf("legacy present: got %q, want %q", got, legacy)
	}
}
