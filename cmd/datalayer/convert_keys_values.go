package datalayer

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/chia-network/go-chia-libs/pkg/rpc"
	"github.com/chia-network/go-modules/pkg/slogs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// convertKeysValuesCmd converts keys and values between different encoding formats
var convertKeysValuesCmd = &cobra.Command{
	Use:   "convert-keys-values",
	Short: "Converts keys and values between different encoding formats",
	Example: `chia-tools data convert-keys-values --id abc123 --input=hex --output=utf8
chia-tools data convert-keys-values --id abc123 --input=utf8 --output=hex`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := rpc.NewClient(rpc.ConnectionModeHTTP, rpc.WithAutoConfig())
		if err != nil {
			slogs.Logr.Fatal("error creating chia RPC client", "error", err)
		}

		storeID := viper.GetString("convert-id")
		if storeID == "" {
			slogs.Logr.Fatal("store ID is required")
		}

		// Get keys and values from the datalayer
		keysValues, _, err := client.DataLayerService.GetKeysValues(&rpc.DatalayerGetKeysValuesOptions{
			ID: storeID,
		})
		if err != nil {
			slogs.Logr.Fatal("error getting keys and values", "error", err)
		}

		// Convert the keys and values
		inputFormat := viper.GetString("input-format")
		outputFormat := viper.GetString("output-format")

		// Create output structure
		output := struct {
			KeysValues []struct {
				Atom  interface{} `json:"atom"`
				Hash  string      `json:"hash"`
				Key   string      `json:"key"`
				Value string      `json:"value"`
			} `json:"keys_values"`
			Success bool `json:"success"`
		}{
			KeysValues: make([]struct {
				Atom  interface{} `json:"atom"`
				Hash  string      `json:"hash"`
				Key   string      `json:"key"`
				Value string      `json:"value"`
			}, 0),
			Success: true,
		}

		// Convert each key-value pair
		for _, kv := range keysValues.KeysValues {
			// Convert key
			convertedKey, err := convertFormat(kv.Key, inputFormat, outputFormat)
			if err != nil {
				slogs.Logr.Fatal("error converting key", "error", err)
			}

			// Convert value
			convertedValue, err := convertFormat(kv.Value, inputFormat, outputFormat)
			if err != nil {
				slogs.Logr.Fatal("error converting value", "error", err)
			}

			output.KeysValues = append(output.KeysValues, struct {
				Atom  interface{} `json:"atom"`
				Hash  string      `json:"hash"`
				Key   string      `json:"key"`
				Value string      `json:"value"`
			}{
				Atom:  kv.Atom,
				Hash:  kv.Hash,
				Key:   convertedKey,
				Value: convertedValue,
			})
		}

		// Convert to JSON with nice formatting
		jsonOutput, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			slogs.Logr.Fatal("error marshaling output to JSON", "error", err)
		}

		fmt.Println(string(jsonOutput))
	},
}

// convertFormat converts a string from one format to another
func convertFormat(input, fromFormat, toFormat string) (string, error) {
	// Remove 0x prefix if present
	input = strings.TrimPrefix(input, "0x")

	switch {
	case fromFormat == "hex" && toFormat == "utf8":
		bytes, err := hex.DecodeString(input)
		if err != nil {
			return "", err
		}
		return string(bytes), nil

	case fromFormat == "utf8" && toFormat == "hex":
		return "0x" + hex.EncodeToString([]byte(input)), nil

	default:
		return "", fmt.Errorf("unsupported conversion from %s to %s", fromFormat, toFormat)
	}
}

func init() {
	convertKeysValuesCmd.PersistentFlags().String("id", "", "The store ID to convert keys and values for")
	convertKeysValuesCmd.PersistentFlags().String("input-format", "hex", "Input format (hex, utf8)")
	convertKeysValuesCmd.PersistentFlags().String("output-format", "utf8", "Output format (hex, utf8)")

	cobra.CheckErr(viper.BindPFlag("convert-id", convertKeysValuesCmd.PersistentFlags().Lookup("id")))
	cobra.CheckErr(viper.BindPFlag("input-format", convertKeysValuesCmd.PersistentFlags().Lookup("input-format")))
	cobra.CheckErr(viper.BindPFlag("output-format", convertKeysValuesCmd.PersistentFlags().Lookup("output-format")))

	datalayerCmd.AddCommand(convertKeysValuesCmd)
}
