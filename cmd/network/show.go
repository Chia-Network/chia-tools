package network

import (
	"fmt"
	"io"
	"net/http"
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
		ShowNetworkInfo()
	},
}

// ShowNetworkInfo outputs network information from the configuration and any running services
func ShowNetworkInfo() {
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

	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	_, _ = fmt.Fprintln(w, "Config\t", configNetwork)

	networkHelper(w, websocketClient.DaemonService, "Daemon")
	networkHelper(w, rpcClient.FullNodeService, "Full Node")
	networkHelper(w, rpcClient.WalletService, "Wallet")
	networkHelper(w, rpcClient.FarmerService, "Farmer")
	networkHelper(w, rpcClient.HarvesterService, "Harvester")
	networkHelper(w, rpcClient.CrawlerService, "Crawler")
	networkHelper(w, rpcClient.DataLayerService, "Data Layer")
	networkHelper(w, rpcClient.TimelordService, "Timelord")
	_ = w.Flush()
}

type hasNetworkName interface {
	GetNetworkInfo(opts *rpc.GetNetworkInfoOptions) (*rpc.GetNetworkInfoResponse, *http.Response, error)
}

func networkHelper(w io.Writer, service hasNetworkName, label string) {
	network, _, err := service.GetNetworkInfo(&rpc.GetNetworkInfoOptions{})
	if err != nil {
		slogs.Logr.Debug("error getting network info from daemon", "error", err)
		network = &rpc.GetNetworkInfoResponse{}
	}
	if network == nil {
		slogs.Logr.Debug("no network info found", "service", label)
		network = &rpc.GetNetworkInfoResponse{}
	}
	_, _ = fmt.Fprintln(w, label, "\t", network.NetworkName.OrElse("Not Running"))
}

func init() {
	networkCmd.AddCommand(showCmd)
}
