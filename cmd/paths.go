package cmd

import (
	"os"
	"path/filepath"
	"strings"
)

// Config and templates locations follow the XDG Base Directory Specification
// for fresh installs ($XDG_CONFIG_HOME/srekit/...), while pre-XDG locations
// (~/.srekit.yaml, ~/.srekit/templates) keep working: if a legacy path
// already exists it wins, so existing users see no change and never end up
// with a silently-ignored config.

func xdgConfigHome() string {
	if v := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); v != "" {
		return v
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config")
}

func legacyConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".srekit.yaml")
}

func xdgConfigPath() string {
	base := xdgConfigHome()
	if base == "" {
		return ""
	}
	return filepath.Join(base, "srekit", "config.yaml")
}

// resolveConfigPath returns the config file to read/write when no explicit
// --config is given: the legacy ~/.srekit.yaml if it already exists, else the
// XDG path. Always returns an absolute, ~-expanded path.
func resolveConfigPath() string {
	if p := legacyConfigPath(); p != "" && fileExists(p) {
		return p
	}
	if p := xdgConfigPath(); p != "" {
		return p
	}
	return legacyConfigPath()
}

func legacyTemplatesDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".srekit", "templates")
}

func xdgTemplatesDir() string {
	base := xdgConfigHome()
	if base == "" {
		return ""
	}
	return filepath.Join(base, "srekit", "templates")
}

// resolveDefaultTemplatesDir returns the templates directory to use when none
// is configured via flag/env/yaml: the legacy ~/.srekit/templates if it
// already exists, else the XDG path. Always returns an absolute path.
func resolveDefaultTemplatesDir() string {
	if p := legacyTemplatesDir(); p != "" && dirExists(p) {
		return p
	}
	if p := xdgTemplatesDir(); p != "" {
		return p
	}
	return legacyTemplatesDir()
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}

func dirExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}
