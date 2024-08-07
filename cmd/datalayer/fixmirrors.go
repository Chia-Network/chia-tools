package datalayer

import (
	"fmt"
	"strings"
	"time"

	"github.com/chia-network/go-chia-libs/pkg/rpc"
	"github.com/chia-network/go-modules/pkg/slogs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// fixMirrorsCmd Replaces one mirror url with another for all mirrors with the url
var fixMirrorsCmd = &cobra.Command{
	Use:     "fix-mirrors",
	Short:   "For all owned mirrors, replaces one url with a new url",
	Example: "chia-tools data fix-mirrors -b 127.0.0.1 -n https://my-dl-domain.com -a 300 -m 0.00000001",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		newURL := viper.GetString("fix-mirror-new-url")
		oldURL := viper.GetString("fix-mirror-bad-url")

		if newURL == "" || oldURL == "" {
			return fmt.Errorf("must provide both --new-url and --old-url flags")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		client, err := rpc.NewClient(rpc.ConnectionModeHTTP, rpc.WithAutoConfig())
		if err != nil {
			slogs.Logr.Fatal("error creating chia RPC client", "error", err)
		}

		// Figure out what fee we are using
		feeXCH := viper.GetFloat64("fix-mirror-fee")
		feeMojos := uint64(feeXCH * 1000000000000)
		slogs.Logr.Info("fee for all transactions", "xch", feeXCH, "mojos", feeMojos)

		subscriptions, _, err := client.DataLayerService.GetSubscriptions(&rpc.DatalayerGetSubscriptionsOptions{})
		if err != nil {
			slogs.Logr.Fatal("error getting list of datalayer subscriptions", "error", err)
		}

		for _, sub := range subscriptions.StoreIDs {
			foundAnyMirror := false

			mirrors, _, err := client.DataLayerService.GetMirrors(&rpc.DatalayerGetMirrorsOptions{
				ID: sub,
			})
			if err != nil {
				slogs.Logr.Fatal("error getting mirrors for subscription", "store", sub, "error", err)
			}
			for _, mirror := range mirrors.Mirrors {
				if !mirror.Ours {
					continue
				}
				for _, url := range mirror.URLs {
					if strings.EqualFold(url, viper.GetString("fix-mirror-bad-url")) {
						foundAnyMirror = true
						waitForAvailableBalance(client, feeMojos)
						slogs.Logr.Info("deleting mirror", "store", sub, "mirror", mirror.CoinID.String())
						_, _, err := client.DataLayerService.DeleteMirror(&rpc.DatalayerDeleteMirrorOptions{
							CoinID: mirror.CoinID.String(),
							Fee:    viper.GetUint64("fix-mirror-fee"),
						})
						if err != nil {
							slogs.Logr.Fatal("error deleting mirror", "store", sub, "mirror", mirror.CoinID.String(), "error", err)
						}
						break
					}
				}
			}

			// Outside the mirror loop, in case there's a weird edge case where we have multiple mirrors on the same bad
			// url, we consolidate down to just one
			if foundAnyMirror {
				mirrorAmount := viper.GetUint64("fix-mirror-amount")
				waitForAvailableBalance(client, mirrorAmount+feeMojos)
				slogs.Logr.Info("adding replacement mirror", "store", sub)
				_, _, err = client.DataLayerService.AddMirror(&rpc.DatalayerAddMirrorOptions{
					ID:     sub,
					URLs:   []string{viper.GetString("fix-mirror-new-url")},
					Amount: mirrorAmount,
					Fee:    feeMojos,
				})
				if err != nil {
					slogs.Logr.Fatal("error adding new mirror", "store", sub, "error", err)
				}
			}
		}
	},
}

// waitForAvailableBalance blocks execution until the wallet has at least one coin and the specified amount available to spend
func waitForAvailableBalance(client *rpc.Client, amount uint64) {
	for {
		balance, _, err := client.WalletService.GetWalletBalance(&rpc.GetWalletBalanceOptions{WalletID: 1})
		if err != nil {
			slogs.Logr.Error("error checking wallet balance. Retrying in 5 seconds", "error", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if balance.Balance.IsAbsent() {
			slogs.Logr.Error("unknown error checking wallet balance. Retrying in 5 seconds")
			time.Sleep(5 * time.Second)
			continue
		}

		// Makes the assumption that wallet balance is not over uint64max mojos
		// It is extremely unlikely to have that balance in one wallet
		if !balance.Balance.MustGet().SpendableBalance.FitsInUint64() || balance.Balance.MustGet().SpendableBalance.Uint64() < amount {
			slogs.Logr.Warn("wallet does not have enough funds to continue. Waiting...", "need", amount, "spendable", balance.Balance.MustGet().SpendableBalance.Uint64())
			time.Sleep(5 * time.Second)
			continue
		}

		// Have enough, so return
		return
	}
}

func init() {
	fixMirrorsCmd.PersistentFlags().Float64P("fee", "m", 0, "Fee to use when deleting and launching the mirrors. The fee is used per mirror. Units are XCH")
	fixMirrorsCmd.PersistentFlags().StringP("new-url", "n", "", "New mirror URL (required)")
	fixMirrorsCmd.PersistentFlags().StringP("bad-url", "b", "", "Old mirror URL to replace (required)")
	fixMirrorsCmd.PersistentFlags().Uint64P("amount", "a", 100, "Mirror coin amount in mojos")

	cobra.CheckErr(viper.BindPFlag("fix-mirror-fee", fixMirrorsCmd.PersistentFlags().Lookup("fee")))
	cobra.CheckErr(viper.BindPFlag("fix-mirror-new-url", fixMirrorsCmd.PersistentFlags().Lookup("new-url")))
	cobra.CheckErr(viper.BindPFlag("fix-mirror-bad-url", fixMirrorsCmd.PersistentFlags().Lookup("bad-url")))
	cobra.CheckErr(viper.BindPFlag("fix-mirror-amount", fixMirrorsCmd.PersistentFlags().Lookup("amount")))

	datalayerCmd.AddCommand(fixMirrorsCmd)
}
