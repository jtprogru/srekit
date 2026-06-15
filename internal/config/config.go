// Package config is the tiny read-only configuration source srekit needs: a
// handful of string values that come from a YAML file, are overridable by
// SREKIT_-prefixed environment variables, and can be set explicitly in tests.
//
// It replaces the sliver of Viper srekit actually used (four GetString
// lookups). Viper pulled in afero, which imports net/http, which drags the
// crypto/TLS + FIPS-140 stack into the binary — none of which a CLI that only
// reads a local YAML file needs.
package config

import (
	"fmt"
	"os"
	"strings"

	yaml "go.yaml.in/yaml/v3"
)

// Config resolves a key in precedence order: an explicit Set override, then the
// SREKIT_<KEY> environment variable, then the value loaded from the config
// file. A zero prefix disables the environment lookup (used by isolated tests,
// mirroring viper.New() without AutomaticEnv).
type Config struct {
	prefix    string
	overrides map[string]string
	file      map[string]string
}

// New returns an empty config that does not read the environment.
func New() *Config {
	return &Config{overrides: map[string]string{}, file: map[string]string{}}
}

// UseEnv enables SREKIT_<KEY>-style environment lookups under the given prefix
// and returns the receiver for chaining.
func (c *Config) UseEnv(prefix string) *Config {
	c.prefix = prefix
	return c
}

// GetString returns the configured value for key, or "" if it is set nowhere.
func (c *Config) GetString(key string) string {
	if v, ok := c.overrides[key]; ok {
		return v
	}
	if c.prefix != "" {
		if v, ok := os.LookupEnv(c.prefix + "_" + strings.ToUpper(key)); ok {
			return v
		}
	}
	return c.file[key]
}

// Set records an explicit override that wins over both environment and file.
func (c *Config) Set(key, value string) {
	c.overrides[key] = value
}

// Load reads scalar key/value pairs from a YAML file into the file layer. An
// empty path is a no-op. A missing or unreadable/invalid file returns an error
// the caller may ignore — config is optional, as it was under Viper.
func (c *Config) Load(path string) error {
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var m map[string]any
	if err := yaml.Unmarshal(data, &m); err != nil {
		return err
	}
	for k, v := range m {
		switch v.(type) {
		case string, int, int64, float64, bool:
			c.file[k] = fmt.Sprint(v)
		}
	}
	return nil
}
