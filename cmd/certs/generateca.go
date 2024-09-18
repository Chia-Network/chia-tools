package certs

import (
	"encoding/json"
	"fmt"

	"github.com/chia-network/go-chia-libs/pkg/tls"
	"github.com/chia-network/go-modules/pkg/slogs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// generateCACmd represents the generate CA command
var generateCACmd = &cobra.Command{
	Use:     "generate-ca",
	Short:   "Generates a new random CA",
	Example: "chia-tools certs generate-ca",
	Run: func(cmd *cobra.Command, args []string) {
		// Get the public CA cert and key byte slices
		publicCACrtBytes, publicCAKeyBytes := tls.GetChiaCACertAndKey()

		// Generate a private CA cert and key
		privateCACrt, privateCAKey, err := tls.GenerateNewCA()
		if err != nil {
			slogs.Logr.Fatal("encountered error generating new private CA cert and key", "error", err)
		}

		// Encode the private CA cert and key to PEM byte slices
		privateCACrtBytes, privateCAKeyBytes, err := tls.EncodeCertAndKeyToPEM(privateCACrt, privateCAKey)
		if err != nil {
			slogs.Logr.Fatal("encountered error encoding private CA cert and key to PEM", "error", err)
		}

		toMarshal := map[string]string{
			"chia_ca.crt":    string(publicCACrtBytes),
			"chia_ca.key":    string(publicCAKeyBytes),
			"private_ca.crt": string(privateCACrtBytes),
			"private_ca.key": string(privateCAKeyBytes),
		}

		var marshalled []byte
		if viper.GetBool("ca-gen-as-json") {
			marshalled, err = json.Marshal(toMarshal)
		} else {
			marshalled, err = yaml.Marshal(toMarshal)
		}

		if err != nil {
			slogs.Logr.Fatal("error marshalling", "error", err)
		}
		fmt.Print(string(marshalled))
	},
}

func init() {
	generateCACmd.PersistentFlags().Bool("as-json", false, "Output as JSON blob instead of yaml")
	cobra.CheckErr(viper.BindPFlag("ca-gen-as-json", generateCACmd.PersistentFlags().Lookup("as-json")))

	certsCmd.AddCommand(generateCACmd)
}
