package testnet

import (
	"github.com/chia-network/go-modules/pkg/slogs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:     "generate",
	Short:   "Generates a new testnet",
	Example: "chia-tools testnet generate",
	Run: func(cmd *cobra.Command, args []string) {
		slogs.Logr.Info("Making new testnet")
	},
}

func init() {
	generateCmd.PersistentFlags().String("ca", "", "Optionally specify a directory that has an existing private_ca.crt/key")
	generateCmd.PersistentFlags().StringP("output", "o", "certs", "Output directory for certs")

	cobra.CheckErr(viper.BindPFlag("ca", generateCmd.PersistentFlags().Lookup("ca")))
	cobra.CheckErr(viper.BindPFlag("cert-output", generateCmd.PersistentFlags().Lookup("output")))

	testnetCmd.AddCommand(generateCmd)
}
