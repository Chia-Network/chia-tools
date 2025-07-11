package coins

import (
	"github.com/spf13/cobra"

	"github.com/chia-network/chia-tools/cmd"
)

// coinsCmd represents the coins command
var coinsCmd = &cobra.Command{
	Use:   "coins",
	Short: "Utilities for working with chia coins",
}

func init() {
	cmd.RootCmd.AddCommand(coinsCmd)
}
