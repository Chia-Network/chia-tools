package certs

import (
	"github.com/chia-network/go-chia-libs/pkg/tls"
	"github.com/chia-network/go-modules/pkg/slogs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates a full set of certificates for chia-blockchain",
	Run: func(cmd *cobra.Command, args []string) {
		err := tls.GenerateAllCerts(viper.GetString("cert-output"))
		if err != nil {
			slogs.Logr.Fatal("error generating certificates", "error", err)
		}
	},
}

func init() {
	generateCmd.PersistentFlags().StringP("output", "o", "certs", "Output directory for certs")
	cobra.CheckErr(viper.BindPFlag("cert-output", generateCmd.PersistentFlags().Lookup("output")))

	certsCmd.AddCommand(generateCmd)
}
