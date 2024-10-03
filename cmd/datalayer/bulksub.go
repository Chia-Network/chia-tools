package datalayer

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/chia-network/go-chia-libs/pkg/rpc"
	"github.com/chia-network/go-modules/pkg/slogs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// bulkSubCmd Subscribes to multiple datastores at once using the output of chia data get_subscriptions
var bulkSubCmd = &cobra.Command{
	Use:     "bulk-subscribe",
	Short:   "Subscribes to multiple datastores at once using the output of chia data get_subscriptions",
	Example: "chia-tools data bulk-subscribe -f subscriptions.json",
	//Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := rpc.NewClient(rpc.ConnectionModeHTTP, rpc.WithAutoConfig())
		if err != nil {
			slogs.Logr.Fatal("error creating chia RPC client", "error", err)
		}

		var content []byte
		if len(args) != 0 {
			content = []byte(strings.Join(args, " "))
		} else {
			jsonFile := viper.GetString("bulksub-file")
			content, err = os.ReadFile(jsonFile)
			if err != nil {
				slogs.Logr.Fatal("Unable to read input file", "error", err)
			}
		}

		subs := &rpc.DatalayerGetSubscriptionsResponse{}
		err = json.Unmarshal(content, subs)
		if err != nil {
			slogs.Logr.Fatal("Could not parse the subscriptions json file", "error", err)
		}

		for _, id := range subs.StoreIDs {
			slogs.Logr.Info("Subscription to store", "id", id)
			_, _, err = client.DataLayerService.Subscribe(&rpc.DatalayerSubscribeOptions{
				ID: id,
			})
			if err != nil {
				slogs.Logr.Error("Error subscribing to datastore", "id", id, "error", err)
			}
		}
	},
}

func init() {
	bulkSubCmd.PersistentFlags().StringP("file", "f", "", "The file containing the json of subscriptions to add")
	cobra.CheckErr(viper.BindPFlag("bulksub-file", bulkSubCmd.PersistentFlags().Lookup("file")))

	datalayerCmd.AddCommand(bulkSubCmd)
}
