package certs

import (
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"os"
	"path"

	"github.com/chia-network/go-chia-libs/pkg/tls"
	"github.com/chia-network/go-modules/pkg/slogs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:     "generate",
	Short:   "Generates a full set of certificates for chia-blockchain",
	Example: "chia-tools certs generate --output ~/.chia/mainnet/config/ssl",
	Run: func(cmd *cobra.Command, args []string) {
		var privateCACert *x509.Certificate
		var privateCAKey *rsa.PrivateKey
		caDir := viper.GetString("ca")
		if caDir != "" {
			caCertPath := path.Join(caDir, "private_ca.crt")
			caKeyPath := path.Join(caDir, "private_ca.key")

			if _, err := os.Stat(caCertPath); err != nil {
				if errors.Is(err, os.ErrNotExist) {
					slogs.Logr.Fatal("private_ca.crt does not exist at the provided path", "path", caCertPath)
				} else {
					slogs.Logr.Fatal("error checking private_ca.crt", "error", err)
				}
			}

			certBytes, err := os.ReadFile(caCertPath)
			if err != nil {
				slogs.Logr.Fatal("error reading ca cert from filesystem", "error", err)
			}
			privateCACert, err = tls.ParsePemCertificate(certBytes)
			if err != nil {
				slogs.Logr.Fatal("error parsing certificate", "error", err)
			}

			if _, err := os.Stat(caKeyPath); err != nil {
				if errors.Is(err, os.ErrNotExist) {
					slogs.Logr.Fatal("private_ca.key does not exist at the provided path", "path", caKeyPath)
				} else {
					slogs.Logr.Fatal("error checking private_ca.key", "error", err)
				}
			}

			keyBytes, err := os.ReadFile(caKeyPath)
			if err != nil {
				slogs.Logr.Fatal("error reading ca key from filesystem", "error", err)
			}
			privateCAKey, err = tls.ParsePemKey(keyBytes)
			if err != nil {
				slogs.Logr.Fatal("error parsing key", "error", err)
			}
		}
		err := tls.GenerateAllCerts(viper.GetString("cert-output"), privateCACert, privateCAKey)
		if err != nil {
			slogs.Logr.Fatal("error generating certificates", "error", err)
		}
	},
}

func init() {
	generateCmd.PersistentFlags().String("ca", "", "Optionally specify a directory that has an existing private_ca.crt/key")
	generateCmd.PersistentFlags().StringP("output", "o", "certs", "Output directory for certs")

	cobra.CheckErr(viper.BindPFlag("ca", generateCmd.PersistentFlags().Lookup("ca")))
	cobra.CheckErr(viper.BindPFlag("cert-output", generateCmd.PersistentFlags().Lookup("output")))

	certsCmd.AddCommand(generateCmd)
}
