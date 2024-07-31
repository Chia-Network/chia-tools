package certs

import (
	"github.com/spf13/cobra"

	"github.com/chia-network/chia-tools/cmd"
)

// certsCmd represents the config command
var certsCmd = &cobra.Command{
	Use:   "certs",
	Short: "Utilities for working with chia certificates",
}

func init() {
	cmd.RootCmd.AddCommand(certsCmd)
}
