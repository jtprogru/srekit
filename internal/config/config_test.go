package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPrecedenceOverrideEnvFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("author: File Author\nemail: file@example.com\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	c := New().UseEnv("SREKIT")
	if err := c.Load(path); err != nil {
		t.Fatalf("load: %v", err)
	}

	// File only.
	if got := c.GetString("author"); got != "File Author" {
		t.Errorf("file: got %q", got)
	}

	// Env beats file.
	t.Setenv("SREKIT_AUTHOR", "Env Author")
	if got := c.GetString("author"); got != "Env Author" {
		t.Errorf("env over file: got %q", got)
	}

	// Explicit Set beats env.
	c.Set("author", "Set Author")
	if got := c.GetString("author"); got != "Set Author" {
		t.Errorf("set over env: got %q", got)
	}

	// Unset key is empty.
	if got := c.GetString("nope"); got != "" {
		t.Errorf("missing key: got %q", got)
	}
}

func TestEnvKeyUppercasesUnderscoreKeys(t *testing.T) {
	c := New().UseEnv("SREKIT")
	t.Setenv("SREKIT_TEMPLATES_DIR", "/tmp/t")
	if got := c.GetString("templates_dir"); got != "/tmp/t" {
		t.Errorf("got %q", got)
	}
}

func TestNoEnvWithoutUseEnv(t *testing.T) {
	c := New() // mirrors viper.New() without AutomaticEnv
	t.Setenv("SREKIT_AUTHOR", "Env Author")
	if got := c.GetString("author"); got != "" {
		t.Errorf("env must be ignored without UseEnv: got %q", got)
	}
}

func TestLoadMissingFileIsNoError(t *testing.T) {
	c := New()
	if err := c.Load(filepath.Join(t.TempDir(), "absent.yaml")); err == nil {
		t.Error("expected error for missing file")
	}
	if err := c.Load(""); err != nil {
		t.Errorf("empty path must be a no-op, got %v", err)
	}
}
