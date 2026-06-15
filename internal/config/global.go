package config

// EnvPrefix is the environment-variable namespace for srekit config keys, e.g.
// SREKIT_AUTHOR overrides the "author" key.
const EnvPrefix = "SREKIT"

// global is the process-wide config the CLI resolves author/email/templates_dir
// against, mirroring how the code previously leaned on viper.GetViper().
var global = New().UseEnv(EnvPrefix)

// Global returns the process-wide config.
func Global() *Config { return global }

// Reset replaces the global config with a fresh env-aware instance. Execute
// calls it once per run before loading the file; tests call it to isolate
// seeded values.
func Reset() { global = New().UseEnv(EnvPrefix) }
