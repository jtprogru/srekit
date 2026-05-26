package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
		Short:   "Generator of SRE text artifacts: tasks, postmortems, runbooks, RFCs, on-call reports, SLOs, retros, changelogs",
		Long:    `srekit generates text artifacts SREs deal with daily — task notes, postmortems, runbooks, RFCs, on-call reports, SLOs, retros, changelogs, licenses — all from embedded templates.`,
		Version: Version,
	}
	root.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.srekit.yaml)")
	root.Flags().BoolP("version", "V", false, "Show version")
	root.SetVersionTemplate(`srekit version: {{.Version}}
from commit: ` + Commit + `
built date: ` + Date + `
built by: ` + BuiltBy + `
`)
	root.AddCommand(
		newLicenseCmd(),
		newTaskCmd(),
		newPostmortemCmd(),
		newRFCCmd(),
		newRunbookCmd(),
		newChangelogCmd(),
		newOncallCmd(),
		newSLOCmd(),
		newRetroCmd(),
	)
	return root
}

func Execute() {
	root := NewRootCmd()
	root.PersistentPreRun = func(cmd *cobra.Command, _ []string) {
		cfgFile, _ := cmd.Root().PersistentFlags().GetString("config")
		initConfig(cfgFile)
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
