package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	showVersion bool
	Version     = "dev"
	Commit      = "none"
	Date        = "today"
	BuiltBy     = "go build"
)

var rootCmd = &cobra.Command{
	Use:     "srekit",
	Short:   "Generator of SRE text artifacts: tasks, licenses, postmortems, RFCs, runbooks",
	Long:    `srekit generates text artifacts SREs deal with daily — task notes, licenses, postmortems, RFCs, runbooks, changelogs — all from embedded templates.`,
	Version: Version,
}

func Execute() {
	if showVersion {
		fmt.Println("srekit version:", Version)
		fmt.Println("from commit:", Commit)
		fmt.Println("built date:", Date)
		fmt.Println("built by:", BuiltBy)
		os.Exit(0)
	}
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.srekit.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&showVersion, "version", "V", false, "Show version")
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
