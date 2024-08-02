package datalayer

import (
	"slices"

	"github.com/chia-network/go-chia-libs/pkg/rpc"
	"github.com/chia-network/go-modules/pkg/slogs"
	"github.com/spf13/cobra"
)

// unsubAllCmd Unsubscribes from all non-owned datalayer stores
var unsubAllCmd = &cobra.Command{
	Use:   "unsub-all",
	Short: "Unsubscribes from all datalayer stores except for owned stores",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := rpc.NewClient(rpc.ConnectionModeHTTP, rpc.WithAutoConfig())
		if err != nil {
			slogs.Logr.Fatal("error creating chia RPC client", "error", err)
		}

		ownedStores, _, err := client.DataLayerService.GetOwnedStores(&rpc.DatalayerGetOwnedStoresOptions{})
		if err != nil {
			slogs.Logr.Fatal("error getting list of owned data stores", "error", err)
		}

		subscriptions, _, err := client.DataLayerService.GetSubscriptions(&rpc.DatalayerGetSubscriptionsOptions{})
		if err != nil {
			slogs.Logr.Fatal("error getting list of datalayer subscriptions", "error", err)
		}

		for _, subscription := range subscriptions.StoreIDs {
			if slices.Contains(ownedStores.StoreIDs, subscription) {
				slogs.Logr.Info("Owned store found, skipping", "store", subscription)
				continue
			}
			slogs.Logr.Info("Unsubscribing from subscription", "store", subscription)
			resp, _, err := client.DataLayerService.Unsubscribe(&rpc.DatalayerUnsubscribeOptions{
				ID:         subscription,
				RetainData: true,
			})
			if err != nil {
				slogs.Logr.Fatal("error unsubscribing from store", "store", subscription, "error", err)
			}
			if !resp.Success {
				slogs.Logr.Fatal("unknown error when unsubscribing from store", "store", subscription)
			}
		}
	},
}

func init() {
	datalayerCmd.AddCommand(unsubAllCmd)
}
