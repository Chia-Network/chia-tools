package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/chia-network/go-modules/pkg/slogs"
)

var (
	gitVersion string
	buildTime  string
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:     "chia-tools",
	Short:   "Collection of CLI tools for working with Chia Blockchain",
	Version: fmt.Sprintf("%s (%s)", gitVersion, buildTime),
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	cobra.OnInitialize(InitLogs)

	RootCmd.PersistentFlags().String("log-level", "info", "The log-level for the application, can be one of info, warn, error, debug.")
	cobra.CheckErr(viper.BindPFlag("log-level", RootCmd.PersistentFlags().Lookup("log-level")))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Find home directory.
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	// Search config in home directory with name ".chia-tools" (without extension).
	viper.AddConfigPath(home)
	viper.SetConfigType("yaml")
	viper.SetConfigName(".chia-tools")

	viper.SetEnvPrefix("CHIA_TOOLS")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

// InitLogs sets up the logger
func InitLogs() {
	slogs.Init(viper.GetString("log-level"))
}
