package config

import (
	"path"

	"github.com/chia-network/go-chia-libs/pkg/config"
	"github.com/chia-network/go-modules/pkg/slogs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// editCmd generates a new chia config
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit an existing chia configuration file",
	Example: `chia-tools config edit --config ~/.chia/mainnet/config/config.yaml --set full_node.port=58444 --set full_node.target_peer_count=10

# The following version will discover the config file by inspecting CHIA_ROOT or using the default CHIA_ROOT
chia-tools config edit --set full_node.port=58444 --set full_node.target_peer_count=10`,
	Run: func(cmd *cobra.Command, args []string) {
		chiaRoot, err := config.GetChiaRootPath()
		if err != nil {
			slogs.Logr.Fatal("Unable to determine CHIA_ROOT", "error", err)
		}

		if len(args) > 0 {
			slogs.Logr.Fatal("Unexpected number of arguments provided")
		}

		cfgPath := viper.GetString("config")
		if cfgPath == "" {
			// Use default chia root
			cfgPath = path.Join(chiaRoot, "config", "config.yaml")
		}

		cfg, err := config.LoadConfigAtRoot(cfgPath, chiaRoot)
		if err != nil {
			slogs.Logr.Fatal("error loading chia config", "error", err)
		}

		err = cfg.FillValuesFromEnvironment()
		if err != nil {
			slogs.Logr.Fatal("error filling values from environment", "error", err)
		}

		valuesToSet := viper.GetStringMapString("edit-set")
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

		err = cfg.Save()
		if err != nil {
			slogs.Logr.Fatal("error saving config", "error", err)
		}
	},
}

func init() {
	editCmd.PersistentFlags().StringToStringP("set", "s", nil, "Paths and values to set in the config")

	cobra.CheckErr(viper.BindPFlag("edit-set", editCmd.PersistentFlags().Lookup("set")))

	configCmd.AddCommand(editCmd)
}
