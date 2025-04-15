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
	Run: func(cmd *cobra.Command, args []string) {
		client, err := rpc.NewClient(rpc.ConnectionModeHTTP, rpc.WithAutoConfig())
		if err != nil {
			slogs.Logr.Fatal("error creating chia RPC client", "error", err)
		}

		subID := viper.GetString("show-mirrors-id")
		if subID != "" {
			// Show mirrors for specific store
			mirrors, _, err := client.DataLayerService.GetMirrors(&rpc.DatalayerGetMirrorsOptions{
				ID: subID,
			})
			if err != nil {
				slogs.Logr.Fatal("error fetching mirrors for subscription", "store", subID, "error", err)
			}

			var ownedMirrors []types.DatalayerMirror
			for _, mirror := range mirrors.Mirrors {
				if mirror.Ours {
					ownedMirrors = append(ownedMirrors, mirror)
				}
			}

			if len(ownedMirrors) == 0 {
				fmt.Println("No owned mirrors found for store")
				return
			}

			// Create output structure
			output := struct {
				Subscriptions []struct {
					StoreID string                  `json:"store_id"`
					Mirrors []types.DatalayerMirror `json:"mirrors"`
				} `json:"subscriptions"`
			}{
				Subscriptions: []struct {
					StoreID string                  `json:"store_id"`
					Mirrors []types.DatalayerMirror `json:"mirrors"`
				}{
					{
						StoreID: subID,
						Mirrors: ownedMirrors,
					},
				},
			}

			// Convert to JSON with nice formatting
			jsonOutput, err := json.MarshalIndent(output, "", "  ")
			if err != nil {
				slogs.Logr.Fatal("error marshaling mirrors to JSON", "error", err)
			}

			fmt.Println(string(jsonOutput))
			return
		}

		// Show mirrors for all stores
		subscriptions, _, err := client.DataLayerService.GetSubscriptions(&rpc.DatalayerGetSubscriptionsOptions{})
		if err != nil {
			slogs.Logr.Fatal("error getting list of datalayer subscriptions", "error", err)
		}

		// Create output structure
		output := struct {
			Subscriptions []struct {
				StoreID string                  `json:"store_id"`
				Mirrors []types.DatalayerMirror `json:"mirrors"`
			} `json:"subscriptions"`
		}{}

		foundAnyMirrors := false
		for _, subscription := range subscriptions.StoreIDs {
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

			if len(ownedMirrors) > 0 {
				foundAnyMirrors = true
				output.Subscriptions = append(output.Subscriptions, struct {
					StoreID string                  `json:"store_id"`
					Mirrors []types.DatalayerMirror `json:"mirrors"`
				}{
					StoreID: subscription,
					Mirrors: ownedMirrors,
				})
			}
		}

		if !foundAnyMirrors {
			fmt.Println("No owned mirrors found for any store")
			return
		}

		// Convert to JSON with nice formatting
		jsonOutput, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			slogs.Logr.Fatal("error marshaling mirrors to JSON", "error", err)
		}

		fmt.Println(string(jsonOutput))
	},
}

func init() {
	showMyMirrorsCmd.PersistentFlags().String("id", "", "The subscription ID to show mirrors for")

	cobra.CheckErr(viper.BindPFlag("show-mirrors-id", showMyMirrorsCmd.PersistentFlags().Lookup("id")))

	datalayerCmd.AddCommand(showMyMirrorsCmd)
}
