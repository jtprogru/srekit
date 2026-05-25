package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	Version = "dev"
	Commit  = "none"
	Date    = "today"
	BuiltBy = "go build"
)

var rootCmd = &cobra.Command{
	Use:     "srekit",
	Short:   "Generator of SRE text artifacts: tasks, licenses, postmortems, RFCs, runbooks",
	Long:    `srekit generates text artifacts SREs deal with daily — task notes, licenses, postmortems, RFCs, runbooks, changelogs — all from embedded templates.`,
	Version: Version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.srekit.yaml)")
	rootCmd.Flags().BoolP("version", "V", false, "Show version")
	rootCmd.SetVersionTemplate(`srekit version: {{.Version}}
from commit: ` + Commit + `
built date: ` + Date + `
built by: ` + BuiltBy + `
`)
}

func initConfig() {
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
