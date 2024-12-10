package testnet

import (
	"github.com/spf13/cobra"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:        "generate",
	Short:      "Generates a new testnet",
	Example:    "chia-tools testnet generate --network examplenet",
	Deprecated: "\nThe `testnet generate` command is deprecated. Please use the 'network generate' command instead\n",
}

func init() {
	testnetCmd.AddCommand(generateCmd)
}
