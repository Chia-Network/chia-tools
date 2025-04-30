package datalayer

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/chia-network/go-chia-libs/pkg/rpc"
	"github.com/chia-network/go-chia-libs/pkg/types"
	"github.com/chia-network/go-modules/pkg/slogs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// convertKeysValuesCmd converts keys and values between different encoding formats
var convertKeysValuesCmd = &cobra.Command{
	Use:   "convert-keys-values",
	Short: "Converts keys and values from the Chia DataLayer get_keys_values endpoint between different encoding formats",
	Example: `chia-tools data convert-keys-values --id abc123 --input-format hex --output-format utf8
chia-tools data convert-keys-values --id abc123 --input-format utf8 --output-format hex`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := rpc.NewClient(rpc.ConnectionModeHTTP, rpc.WithAutoConfig())
		if err != nil {
			slogs.Logr.Fatal("error creating chia RPC client", "error", err)
		}

		storeID := viper.GetString("convert-id")
		if storeID == "" {
			slogs.Logr.Fatal("store ID is required")
		}

		// DEBUG: Log the store ID being used
		slogs.Logr.Debug("Processing store", "id", storeID)

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

		// DEBUG: Log the conversion formats being used
		slogs.Logr.Debug("Conversion formats", "input", inputFormat, "output", outputFormat)

		// Create output structure that matches Chia DataLayer RPC format
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
			Success: keysValues.Success,
		}

		// Convert each key-value pair
		for _, kv := range keysValues.KeysValues {
			// DEBUG: Log the original values
			slogs.Logr.Debug("Original values",
				"key_hex", fmt.Sprintf("%x", kv.Key),
				"key_type", fmt.Sprintf("%T", kv.Key),
				"value_hex", fmt.Sprintf("%x", kv.Value),
				"value_type", fmt.Sprintf("%T", kv.Value))

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

			// DEBUG: Log the converted values
			slogs.Logr.Debug("Converted values",
				"key", convertedKey,
				"key_type", fmt.Sprintf("%T", convertedKey),
				"value", convertedValue,
				"value_type", fmt.Sprintf("%T", convertedValue))

			// Create new key-value pair with converted values
			newKV := struct {
				Atom  interface{} `json:"atom"`
				Hash  string      `json:"hash"`
				Key   string      `json:"key"`
				Value string      `json:"value"`
			}{
				Atom:  kv.Atom,
				Hash:  kv.Hash.String(), // Use the hash from the input, which should already be in the correct format
				Key:   convertedKey,     // Use converted key directly
				Value: convertedValue,   // Use converted value directly
			}

			// DEBUG: Log the new key-value pair
			slogs.Logr.Debug("New key-value pair",
				"key", newKV.Key,
				"key_type", fmt.Sprintf("%T", newKV.Key),
				"value", newKV.Value,
				"value_type", fmt.Sprintf("%T", newKV.Value))

			output.KeysValues = append(output.KeysValues, newKV)
		}

		// Convert to JSON with nice formatting
		jsonOutput, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			slogs.Logr.Fatal("error marshaling output to JSON", "error", err)
		}

		// DEBUG: Log the final JSON output
		slogs.Logr.Debug("Final JSON output", "output", string(jsonOutput))

		fmt.Println(string(jsonOutput))
	},
}

// convertFormat converts a string from one format to another
func convertFormat(input types.Bytes, fromFormat, toFormat string) (string, error) {
	// If formats are the same, return the input as is
	if fromFormat == toFormat {
		return string(input), nil
	}

	// DEBUG: Log the input string and formats
	slogs.Logr.Debug("Converting format",
		"from", fromFormat,
		"to", toFormat,
		"input_type", fmt.Sprintf("%T", input),
		"input_hex", fmt.Sprintf("%x", input))

	switch {
	case fromFormat == "hex" && toFormat == "utf8":
		// Convert bytes directly to string
		result := string(input)
		// DEBUG: Log the final UTF-8 string
		slogs.Logr.Debug("Final UTF-8 string", "utf8", result)
		return result, nil

	case fromFormat == "utf8" && toFormat == "hex":
		// Convert UTF-8 string to hex
		result := "0x" + hex.EncodeToString(input)
		return result, nil

	default:
		return "", fmt.Errorf("unsupported conversion from %s to %s", fromFormat, toFormat)
	}
}

// isHexString checks if a string is a valid hex string
func isHexString(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
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
