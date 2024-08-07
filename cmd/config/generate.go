package config

import (
	"os"

	"github.com/chia-network/go-chia-libs/pkg/config"
	"github.com/chia-network/go-modules/pkg/slogs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// generateCmd generates a new chia config
var generateCmd = &cobra.Command{
	Use:     "generate",
	Short:   "Generate a new chia configuration file",
	Example: "chia-tools config generate --set full_node.port=58444 --set full_node.target_peer_count=10 --output ~/.chia/mainnet/config/config.yaml",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadDefaultConfig()
		if err != nil {
			slogs.Logr.Fatal("error loading default config", "error", err)
		}

		err = cfg.FillValuesFromEnvironment()
		if err != nil {
			slogs.Logr.Fatal("error filling values from environment", "error", err)
		}

		valuesToSet := viper.GetStringMapString("set")
		for path, value := range valuesToSet {
			pathMap := config.ParsePathsFromStrings([]string{path}, false)
			var key string
			var pathSlice []string
			for key, pathSlice = range pathMap {
				break
			}
			err = cfg.SetFieldByPath(pathSlice, value)
			if err != nil {
				slogs.Logr.Fatal("error setting path in config", "key", key, "value", value, "error", err)
			}
		}

		out, err := yaml.Marshal(cfg)
		if err != nil {
			slogs.Logr.Fatal("error marshalling config", "error", err)
		}

		err = os.WriteFile(viper.GetString("output"), out, 0655)
		if err != nil {
			slogs.Logr.Fatal("error writing output file", "error", err)
		}
	},
}

func init() {
	generateCmd.PersistentFlags().StringP("output", "o", "config.yml", "Output file for config")
	generateCmd.PersistentFlags().StringToStringP("set", "s", nil, "Paths and values to set in the config")

	cobra.CheckErr(viper.BindPFlag("output", generateCmd.PersistentFlags().Lookup("output")))
	cobra.CheckErr(viper.BindPFlag("set", generateCmd.PersistentFlags().Lookup("set")))

	configCmd.AddCommand(generateCmd)
}
