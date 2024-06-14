package config

import (
	"github.com/spf13/cobra"

	"github.com/chia-network/chia-tools/cmd"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Utilities for working with chia config",
}

func init() {
	cmd.RootCmd.AddCommand(configCmd)
}
