package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/jtprogru/srekit/internal/tmpl"
)

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

// configureTemplates installs a DirSource on tmpl.Default if the user
// configured one. The common case (no custom dir) leaves tmpl.Default
// untouched — keeps parallel tests race-free on the package-level Default.
func configureTemplates(cmd *cobra.Command) error {
	dir, err := resolveTemplatesDir(cmd)
	if err != nil {
		return fmt.Errorf("--templates-dir: %w", err)
	}
	if dir == "" {
		return nil
	}
	info, err := os.Stat(dir)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "srekit: templates dir %s: %v (falling back to embedded)\n", dir, err)
		return nil
	}
	if !info.IsDir() {
		fmt.Fprintf(cmd.ErrOrStderr(), "srekit: %s is not a directory (falling back to embedded)\n", dir)
		return nil
	}
	tmpl.Default = &tmpl.Loader{Sources: []tmpl.Source{tmpl.DirSource{Dir: dir}, tmpl.EmbedSource{}}}
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
