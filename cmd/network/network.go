package network

import (
	"github.com/spf13/cobra"

	"github.com/chia-network/chia-tools/cmd"
)

// networkCmd represents the config command
var networkCmd = &cobra.Command{
	Use:   "network",
	Short: "Utilities for working with chia networks",
}

func init() {
	cmd.RootCmd.AddCommand(networkCmd)
}
