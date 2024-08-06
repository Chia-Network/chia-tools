package datalayer

import (
	"github.com/chia-network/go-chia-libs/pkg/rpc"
	"github.com/chia-network/go-chia-libs/pkg/types"
	"github.com/chia-network/go-modules/pkg/slogs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// deleteMirrorsCmd Deletes all owned mirrors for all datalayer subscriptions
var deleteMirrorsCmd = &cobra.Command{
	Use:   "delete-mirrors",
	Short: "Deletes all owned mirrors for all datalayer subscriptions",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := rpc.NewClient(rpc.ConnectionModeHTTP, rpc.WithAutoConfig())
		if err != nil {
			slogs.Logr.Fatal("error creating chia RPC client", "error", err)
		}

		// Figure out what fee we are using
		feeXCH := viper.GetFloat64("delete-mirror-fee")
		feeMojos := uint64(feeXCH * 1000000000000)
		slogs.Logr.Debug("fee for all transactions", "xch", feeXCH, "mojos", feeMojos)

		subscriptions, _, err := client.DataLayerService.GetSubscriptions(&rpc.DatalayerGetSubscriptionsOptions{})
		if err != nil {
			slogs.Logr.Fatal("error getting list of datalayer subscriptions", "error", err)
		}

		for _, subscription := range subscriptions.StoreIDs {
			slogs.Logr.Info("checking subscription", "store", subscription)

			mirrors, _, err := client.DataLayerService.GetMirrors(&rpc.DatalayerGetMirrorsOptions{
				ID: subscription,
			})
			if err != nil {
				slogs.Logr.Fatal("error fetching mirrors for subscription", "store", subscription, "error", err)
			}
			var ownedMirrors []types.Bytes32

			for _, mirror := range mirrors.Mirrors {
				if mirror.Ours {
					ownedMirrors = append(ownedMirrors, mirror.CoinID)
				}
			}

			if len(ownedMirrors) == 0 {
				slogs.Logr.Info("no owned mirrors for this datastore", "store", subscription)
				continue
			}

			for _, coinID := range ownedMirrors {
				slogs.Logr.Info("deleting mirror", "store", subscription, "mirror", coinID.String())
				resp, _, err := client.DataLayerService.DeleteMirror(&rpc.DatalayerDeleteMirrorOptions{
					CoinID: coinID.String(),
					Fee:    feeMojos,
				})
				if err != nil {
					slogs.Logr.Fatal("error deleting mirror for store", "store", subscription, "mirror", coinID, "error", err)
				}
				if !resp.Success {
					slogs.Logr.Fatal("unknown error when deleting mirror for store", "store", subscription, "mirror", coinID)
				}
			}
		}
	},
}

func init() {
	deleteMirrorsCmd.PersistentFlags().Float64P("fee", "m", 0, "Fee to use when deleting the mirrors. The fee is used per mirror. Units are XCH")
	cobra.CheckErr(viper.BindPFlag("delete-mirror-fee", deleteMirrorsCmd.PersistentFlags().Lookup("fee")))

	datalayerCmd.AddCommand(deleteMirrorsCmd)
}
