package datalayer

import (
	"github.com/spf13/cobra"

	"github.com/chia-network/chia-tools/cmd"
)

// datalayerCmd represents the config command
var datalayerCmd = &cobra.Command{
	Use:   "data",
	Short: "Utilities for working with chia data layer",
}

func init() {
	cmd.RootCmd.AddCommand(datalayerCmd)
}
