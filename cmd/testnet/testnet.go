package testnet

import (
	"github.com/spf13/cobra"

	"github.com/chia-network/chia-tools/cmd"
)

// testnetCmd represents the config command
var testnetCmd = &cobra.Command{
	Use:   "testnet",
	Short: "Utilities for working with chia testnets",
}

func init() {
	cmd.RootCmd.AddCommand(testnetCmd)
}
