package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/chia-network/chia-tools/cmd"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Utilities for working with chia config",
}

func init() {
	configCmd.PersistentFlags().String("config", "", "existing config file to use (default is to look in $CHIA_ROOT)")
	cobra.CheckErr(viper.BindPFlag("config", configCmd.PersistentFlags().Lookup("config")))

	cmd.RootCmd.AddCommand(configCmd)
}
