package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/jtprogru/srekit/internal/tmpl"
)

// loaderCtxKey keys the per-invocation *tmpl.Loader stored on the command
// context by configureTemplates and read back by render commands via
// loaderFrom. Using the context (rather than a package-level var) keeps the
// loader scoped to a single command tree, so parallel tests don't race.
type loaderCtxKey struct{}

// loaderFrom returns the template loader configured for this invocation.
// configureTemplates always sets one in PersistentPreRunE before any RunE
// runs; the embedded-only fallback guards the rare path that bypasses it.
func loaderFrom(cmd *cobra.Command) *tmpl.Loader {
	if l, ok := cmd.Context().Value(loaderCtxKey{}).(*tmpl.Loader); ok && l != nil {
		return l
	}
	return tmpl.NewDefaultLoader()
}

var (
	Version = "dev"
	Commit  = "none"
	Date    = "today"
	BuiltBy = "go build"
)

// NewRootCmd builds a fresh srekit command tree with no global side effects.
// Tests use it to get an isolated cobra.Command per test; production calls
// Execute, which additionally wires the ~/.srekit.yaml + env config loader.
func NewRootCmd() *cobra.Command {
	var cfgFile string
	root := &cobra.Command{
		Use:     "srekit",
		Short:   "Generator of SRE text artifacts: investigations, incidents, postmortems, runbooks, RFCs, on-call reports, SLOs, error budget policies, capacity plans, retros, changelogs",
		Long:    `srekit generates text artifacts SREs deal with daily — investigation logs, live-incident reports, postmortems, runbooks, RFCs, on-call reports, SLOs, error budget policies, capacity plans, retros, changelogs, licenses — all from embedded templates.`,
		Version: Version,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return configureTemplates(cmd)
		},
	}
	root.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.srekit.yaml)")
	root.PersistentFlags().String("templates-dir", "", "directory of custom templates (missing files fall back to embedded)")
	root.Flags().BoolP("version", "V", false, "Show version")
	root.SetVersionTemplate(`srekit version: {{.Version}}
from commit: ` + Commit + `
built date: ` + Date + `
built by: ` + BuiltBy + `
`)
	root.AddCommand(
		newLicenseCmd(),
		newTaskCmd(),
		newIncidentCmd(),
		newPostmortemCmd(),
		newRFCCmd(),
		newRunbookCmd(),
		newChangelogCmd(),
		newOncallCmd(),
		newSLOCmd(),
		newEBPCmd(),
		newCapacityCmd(),
		newRetroCmd(),
		newTemplatesCmd(),
		newConfigCmd(),
	)
	return root
}

func Execute() {
	root := NewRootCmd()
	inner := root.PersistentPreRunE
	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		cfgFile, _ := cmd.Root().PersistentFlags().GetString("config")
		initConfig(cfgFile)
		if inner != nil {
			return inner(cmd, args)
		}
		return nil
	}
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func initConfig(cfgFile string) {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		viper.AddConfigPath(home)
		viper.SetConfigName(".srekit")
		viper.SetConfigType("yaml")
	}
	viper.SetEnvPrefix("SREKIT")
	viper.AutomaticEnv()
	_ = viper.ReadInConfig()
}

// resolveTemplatesDir returns the templates directory configured via
// --templates-dir, SREKIT_TEMPLATES_DIR, or templates_dir: in ~/.srekit.yaml,
// in that order of precedence. Expands a leading ~. Returns "" if no source
// is configured.
func resolveTemplatesDir(cmd *cobra.Command) (string, error) {
	var dir string
	if f := cmd.Flags().Lookup("templates-dir"); f != nil {
		dir = f.Value.String()
	}
	if dir == "" {
		dir = viper.GetString("templates_dir")
	}
	if dir == "" {
		return "", nil
	}
	return expandHome(dir)
}

// configureTemplates builds the loader for this invocation and stashes it on
// the command context. The loader serves the embedded templates by default;
// when --templates-dir (or its env/yaml equivalent) points at a usable
// directory, a DirSource is prepended so user templates take priority with
// transparent per-file fallback to embedded.
func configureTemplates(cmd *cobra.Command) error {
	loader := tmpl.NewDefaultLoader()
	dir, err := resolveTemplatesDir(cmd)
	if err != nil {
		return fmt.Errorf("--templates-dir: %w", err)
	}
	switch info, statErr := os.Stat(dir); {
	case dir == "":
		// no custom dir configured — embedded only
	case statErr != nil:
		fmt.Fprintf(cmd.ErrOrStderr(), "srekit: templates dir %s: %v (falling back to embedded)\n", dir, statErr)
	case !info.IsDir():
		fmt.Fprintf(cmd.ErrOrStderr(), "srekit: %s is not a directory (falling back to embedded)\n", dir)
	default:
		loader.AddDirSource(dir)
	}
	cmd.SetContext(context.WithValue(cmd.Context(), loaderCtxKey{}, loader))
	return nil
}

func expandHome(path string) (string, error) {
	if path == "" || path[0] != '~' {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if path == "~" {
		return home, nil
	}
	if len(path) > 1 && path[1] == '/' {
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}
