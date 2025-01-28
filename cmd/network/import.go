package network

import (
	"io"
	"net/http"

	"github.com/chia-network/go-chia-libs/pkg/config"
	"github.com/chia-network/go-modules/pkg/slogs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:     "import",
	Short:   "Import a network configuration from a remote source",
	Example: "chia-tools network import --network mytestnet --url https://example.com/my-network-config.yml",
	Run: func(cmd *cobra.Command, args []string) {
		network := viper.GetString("net-import-network")
		url := viper.GetString("net-import-url")
		slogs.Logr.Info("Importing remote network settings", "network", network, "url", url)

		resp, err := http.Get(url)
		if err != nil {
			slogs.Logr.Fatal("Failed to load remote network settings", "error", err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				slogs.Logr.Fatal("Failed to close remote network settings body", "error", err)
			}
		}(resp.Body)

		cfgBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			slogs.Logr.Fatal("Failed to read remote network settings body", "error", err)
		}

		cfg := &config.ChiaConfig{}
		err = yaml.Unmarshal(cfgBytes, cfg)
		if err != nil {
			slogs.Logr.Fatal("Failed to unmarshal remote network settings body", "error", err)
		}

		if _, ok := cfg.NetworkOverrides.Constants[network]; !ok {
			slogs.Logr.Fatal("Network constants not found in remote config", "network", network)
		}
		if _, ok := cfg.NetworkOverrides.Config[network]; !ok {
			slogs.Logr.Fatal("Network config not found in remote config", "network", network)
		}

		chiaRoot, err := config.GetChiaRootPath()
		if err != nil {
			slogs.Logr.Fatal("error determining chia root", "error", err)
		}
		slogs.Logr.Debug("Chia root discovered", "CHIA_ROOT", chiaRoot)

		localCfg, err := config.GetChiaConfig()
		if err != nil {
			slogs.Logr.Fatal("error loading config", "error", err)
		}
		slogs.Logr.Debug("Successfully loaded config")

		localCfg.NetworkOverrides.Constants[network] = cfg.NetworkOverrides.Constants[network]
		localCfg.NetworkOverrides.Config[network] = cfg.NetworkOverrides.Config[network]

		err = localCfg.Save()
		if err != nil {
			slogs.Logr.Fatal("Failed to save config", "error", err)
		}

		slogs.Logr.Info("Successfully imported to config")

		if viper.GetBool("net-import-switch") {
			SwitchNetwork(network, true)
		}
	},
}

func init() {
	importCmd.PersistentFlags().String("network", "", "Network name to import")
	importCmd.PersistentFlags().StringP("url", "u", "", "URL of the remote config")
	importCmd.PersistentFlags().Bool("switch", false, "Whether to immediately switch to the network")

	cobra.CheckErr(importCmd.MarkPersistentFlagRequired("network"))
	cobra.CheckErr(importCmd.MarkPersistentFlagRequired("url"))

	cobra.CheckErr(viper.BindPFlag("net-import-network", importCmd.PersistentFlags().Lookup("network")))
	cobra.CheckErr(viper.BindPFlag("net-import-url", importCmd.PersistentFlags().Lookup("url")))
	cobra.CheckErr(viper.BindPFlag("net-import-switch", importCmd.PersistentFlags().Lookup("switch")))

	networkCmd.AddCommand(importCmd)
}
