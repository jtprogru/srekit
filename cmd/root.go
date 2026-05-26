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
	var cfgFile, templatesDir string
	root := &cobra.Command{
		Use:     "srekit",
		Short:   "Generator of SRE text artifacts: investigations, incidents, postmortems, runbooks, RFCs, on-call reports, SLOs, error budget policies, capacity plans, retros, changelogs",
		Long:    `srekit generates text artifacts SREs deal with daily — investigation logs, live-incident reports, postmortems, runbooks, RFCs, on-call reports, SLOs, error budget policies, capacity plans, retros, changelogs, licenses — all from embedded templates.`,
		Version: Version,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return configureTemplates(cmd, templatesDir)
		},
	}
	root.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.srekit.yaml)")
	root.PersistentFlags().StringVar(&templatesDir, "templates-dir", "", "directory of custom templates (missing files fall back to embedded)")
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

// configureTemplates installs a DirSource on tmpl.Default if the user provided
// --templates-dir, SREKIT_TEMPLATES_DIR, or templates_dir: in ~/.srekit.yaml.
// Common case (no custom dir) leaves tmpl.Default untouched — that keeps
// parallel tests race-free on the package-level Default.
func configureTemplates(cmd *cobra.Command, flagValue string) error {
	dir := flagValue
	if dir == "" {
		dir = viper.GetString("templates_dir")
	}
	if dir == "" {
		return nil
	}
	expanded, err := expandHome(dir)
	if err != nil {
		return fmt.Errorf("--templates-dir: %w", err)
	}
	info, err := os.Stat(expanded)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "srekit: templates dir %s: %v (falling back to embedded)\n", expanded, err)
		return nil
	}
	if !info.IsDir() {
		fmt.Fprintf(cmd.ErrOrStderr(), "srekit: %s is not a directory (falling back to embedded)\n", expanded)
		return nil
	}
	tmpl.Default = &tmpl.Loader{Sources: []tmpl.Source{tmpl.DirSource{Dir: expanded}, tmpl.EmbedSource{}}}
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
