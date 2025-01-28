package network

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/chia-network/go-chia-libs/pkg/config"
	"github.com/chia-network/go-chia-libs/pkg/rpc"
	"github.com/chia-network/go-modules/pkg/slogs"
	"github.com/spf13/cobra"
)

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show information about the currently selected/running network",
	Run: func(cmd *cobra.Command, args []string) {
		chiaRoot, err := config.GetChiaRootPath()
		if err != nil {
			slogs.Logr.Fatal("error determining chia root", "error", err)
		}
		slogs.Logr.Debug("Chia root discovered", "CHIA_ROOT", chiaRoot)

		cfg, err := config.GetChiaConfig()
		if err != nil {
			slogs.Logr.Fatal("error loading config", "error", err)
		}
		slogs.Logr.Debug("Successfully loaded config")

		configNetwork := *cfg.SelectedNetwork

		slogs.Logr.Debug("initializing websocket client")
		websocketClient, err := rpc.NewClient(rpc.ConnectionModeWebsocket, rpc.WithAutoConfig(), rpc.WithSyncWebsocket())
		if err != nil {
			slogs.Logr.Fatal("error initializing websocket RPC client", "error", err)
		}
		slogs.Logr.Debug("initializing http client")
		rpcClient, err := rpc.NewClient(rpc.ConnectionModeHTTP, rpc.WithAutoConfig(), rpc.WithSyncWebsocket())
		if err != nil {
			slogs.Logr.Fatal("error initializing websocket RPC client", "error", err)
		}
		daemonNetwork, _, err := websocketClient.DaemonService.GetNetworkInfo(&rpc.GetNetworkInfoOptions{})
		if err != nil {
			slogs.Logr.Debug("error getting network info from daemon", "error", err)
		}

		fullNodeNetwork, _, err := rpcClient.FullNodeService.GetNetworkInfo(&rpc.GetNetworkInfoOptions{})
		if err != nil {
			slogs.Logr.Debug("error getting network info from full node", "error", err)
		}

		walletNetwork, _, err := rpcClient.WalletService.GetNetworkInfo(&rpc.GetNetworkInfoOptions{})
		if err != nil {
			slogs.Logr.Debug("error getting network info from wallet", "error", err)
		}

		farmerNetwork, _, err := rpcClient.FarmerService.GetNetworkInfo(&rpc.GetNetworkInfoOptions{})
		if err != nil {
			slogs.Logr.Debug("error getting network info from farmer", "error", err)
		}

		harvesterNetwork, _, err := rpcClient.HarvesterService.GetNetworkInfo(&rpc.GetNetworkInfoOptions{})
		if err != nil {
			slogs.Logr.Debug("error getting network info from harvester", "error", err)
		}

		crawlerNetwork, _, err := rpcClient.CrawlerService.GetNetworkInfo(&rpc.GetNetworkInfoOptions{})
		if err != nil {
			slogs.Logr.Debug("error getting network info from crawler", "error", err)
		}

		datalayerNetwork, _, err := rpcClient.DataLayerService.GetNetworkInfo(&rpc.GetNetworkInfoOptions{})
		if err != nil {
			slogs.Logr.Debug("error getting network info from dataLayer", "error", err)
		}

		timelordNetwork, _, err := rpcClient.TimelordService.GetNetworkInfo(&rpc.GetNetworkInfoOptions{})
		if err != nil {
			slogs.Logr.Debug("error getting network info from timelord", "error", err)
		}

		w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
		_, _ = fmt.Fprintln(w, "Config\t", configNetwork)
		_, _ = fmt.Fprintln(w, "Daemon\t", daemonNetwork.NetworkName.OrElse("Not Running"))
		_, _ = fmt.Fprintln(w, "Full Node\t", fullNodeNetwork.NetworkName.OrElse("Not Running"))
		_, _ = fmt.Fprintln(w, "Wallet\t", walletNetwork.NetworkName.OrElse("Not Running"))
		_, _ = fmt.Fprintln(w, "Farmer\t", farmerNetwork.NetworkName.OrElse("Not Running"))
		_, _ = fmt.Fprintln(w, "Harvester\t", harvesterNetwork.NetworkName.OrElse("Not Running"))
		_, _ = fmt.Fprintln(w, "Crawler\t", crawlerNetwork.NetworkName.OrElse("Not Running"))
		_, _ = fmt.Fprintln(w, "Data Layer\t", datalayerNetwork.NetworkName.OrElse("Not Running"))
		_, _ = fmt.Fprintln(w, "Timelord\t", timelordNetwork.NetworkName.OrElse("Not Running"))
		_ = w.Flush()
	},
}

func init() {
	networkCmd.AddCommand(showCmd)
}
