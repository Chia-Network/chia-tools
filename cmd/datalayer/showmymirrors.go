package datalayer

import (
	"encoding/json"
	"fmt"

	"github.com/chia-network/go-chia-libs/pkg/rpc"
	"github.com/chia-network/go-chia-libs/pkg/types"
	"github.com/chia-network/go-modules/pkg/slogs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// showMyMirrorsCmd shows all mirrors owned by the user across all or specific stores
var showMyMirrorsCmd = &cobra.Command{
	Use:   "show-my-mirrors",
	Short: "Shows all mirrors owned by the user across all or specific stores",
	Example: `chia-tools data show-my-mirrors
chia-tools data show-my-mirrors --id abcd1234`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		subID := viper.GetString("show-mirrors-id")
		if subID != "" {
			return nil
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		client, err := rpc.NewClient(rpc.ConnectionModeHTTP, rpc.WithAutoConfig())
		if err != nil {
			slogs.Logr.Fatal("error creating chia RPC client", "error", err)
		}

		subID := viper.GetString("show-mirrors-id")
		if subID != "" {
			// Show mirrors for specific store
			ShowMirrorsForStore(client, subID)
			return
		}

		// Show mirrors for all stores
		subscriptions, _, err := client.DataLayerService.GetSubscriptions(&rpc.DatalayerGetSubscriptionsOptions{})
		if err != nil {
			slogs.Logr.Fatal("error getting list of datalayer subscriptions", "error", err)
		}

		for _, subscription := range subscriptions.StoreIDs {
			ShowMirrorsForStore(client, subscription)
		}
	},
}

// ShowMirrorsForStore displays all owned mirrors for a given store
func ShowMirrorsForStore(client *rpc.Client, subscription string) {
	mirrors, _, err := client.DataLayerService.GetMirrors(&rpc.DatalayerGetMirrorsOptions{
		ID: subscription,
	})
	if err != nil {
		slogs.Logr.Fatal("error fetching mirrors for subscription", "store", subscription, "error", err)
	}

	var ownedMirrors []types.DatalayerMirror
	for _, mirror := range mirrors.Mirrors {
		if mirror.Ours {
			ownedMirrors = append(ownedMirrors, mirror)
		}
	}

	if len(ownedMirrors) == 0 {
		slogs.Logr.Info("no owned mirrors for this datastore", "store", subscription)
		return
	}

	// Create a struct to hold both store ID and mirrors for JSON output
	output := struct {
		StoreID string                  `json:"store_id"`
		Mirrors []types.DatalayerMirror `json:"mirrors"`
	}{
		StoreID: subscription,
		Mirrors: ownedMirrors,
	}

	// Convert to JSON with nice formatting
	jsonOutput, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		slogs.Logr.Fatal("error marshaling mirrors to JSON", "error", err)
	}

	fmt.Println(string(jsonOutput))
}

func init() {
	showMyMirrorsCmd.PersistentFlags().String("id", "", "The subscription ID to show mirrors for")

	cobra.CheckErr(viper.BindPFlag("show-mirrors-id", showMyMirrorsCmd.PersistentFlags().Lookup("id")))

	datalayerCmd.AddCommand(showMyMirrorsCmd)
}
