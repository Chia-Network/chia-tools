package config

import (
	"os"
	"path"

	"github.com/chia-network/go-chia-libs/pkg/config"
	"github.com/chia-network/go-modules/pkg/slogs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// editCmd generates a new chia config
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit an existing chia configuration file",
	Example: `chia-tools config edit --config ~/.chia/mainnet/config/config.yaml --set full_node.port=58444 --set full_node.target_peer_count=10

# The following version will discover the config file by inspecting CHIA_ROOT or using the default CHIA_ROOT
chia-tools config edit --set full_node.port=58444 --set full_node.target_peer_count=10

# Show what changes would be made without actually making them
chia-tools config edit --set full_node.port=58444 --dry-run`,
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

		// Read the current config file to get existing values
		currentConfig := make(map[string]interface{})
		if _, err := os.Stat(cfgPath); err == nil {
			content, err := os.ReadFile(cfgPath)
			if err != nil {
				slogs.Logr.Fatal("error reading config file", "error", err)
			}
			err = yaml.Unmarshal(content, currentConfig)
			if err != nil {
				slogs.Logr.Fatal("error parsing config file", "error", err)
			}
		}

		cfg, err := config.LoadConfigAtRoot(cfgPath, chiaRoot)
		if err != nil {
			slogs.Logr.Fatal("error loading chia config", "error", err)
		}

		err = cfg.FillValuesFromEnvironment()
		if err != nil {
			slogs.Logr.Fatal("error filling values from environment", "error", err)
		}

		dryRun := viper.GetBool("dry-run")
		if dryRun {
			slogs.Logr.Info("DRY RUN: The following changes would be made to the config file")
		}

		valuesToSet := viper.GetStringMapString("edit-set")
		for path, value := range valuesToSet {
			pathMap := config.ParsePathsFromStrings([]string{path}, false)
			var key string
			var pathSlice []string
			for key, pathSlice = range pathMap {
				break
			}

			// Get the current value from the config file
			currentValue := getValueFromMap(currentConfig, pathSlice)
			if currentValue == nil {
				slogs.Logr.Info("Config value not found", "path", path)
			}

			if dryRun {
				slogs.Logr.Info("Would change config value",
					"path", path,
					"current_value", currentValue,
					"new_value", value)
				continue
			}

			err = cfg.SetFieldByPath(pathSlice, value)
			if err != nil {
				slogs.Logr.Fatal("error setting path in config", "key", key, "value", value, "error", err)
			}
		}

		if dryRun {
			slogs.Logr.Info("DRY RUN: No changes were made to the config file")
			return
		}

		err = cfg.Save()
		if err != nil {
			slogs.Logr.Fatal("error saving config", "error", err)
		}
	},
}

// getValueFromMap traverses a map using a path slice to get a value
func getValueFromMap(m map[string]interface{}, path []string) interface{} {
	current := m
	for _, key := range path {
		if current == nil {
			return nil
		}
		if value, ok := current[key]; ok {
			current = value.(map[string]interface{})
		} else {
			return nil
		}
	}
	return current
}

func init() {
	editCmd.PersistentFlags().StringToStringP("set", "s", nil, "Paths and values to set in the config")

	cobra.CheckErr(viper.BindPFlag("edit-set", editCmd.PersistentFlags().Lookup("set")))

	configCmd.AddCommand(editCmd)
}
