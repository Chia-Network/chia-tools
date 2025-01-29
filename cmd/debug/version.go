package debug

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"text/tabwriter"

	"github.com/chia-network/go-chia-libs/pkg/rpc"
	"github.com/chia-network/go-modules/pkg/slogs"
)

// ShowVersionInfo outputs the running version for all services
func ShowVersionInfo() {
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
	versionHelper(w, websocketClient.DaemonService, "Daemon")
	versionHelper(w, rpcClient.FullNodeService, "Full Node")
	versionHelper(w, rpcClient.WalletService, "Wallet")
	versionHelper(w, rpcClient.FarmerService, "Farmer")
	versionHelper(w, rpcClient.HarvesterService, "Harvester")
	versionHelper(w, rpcClient.CrawlerService, "Crawler")
	versionHelper(w, rpcClient.DataLayerService, "Data Layer")
	versionHelper(w, rpcClient.TimelordService, "Timelord")
	_ = w.Flush()
}

type hasVersionInfo interface {
	GetVersion(opts *rpc.GetVersionOptions) (*rpc.GetVersionResponse, *http.Response, error)
}

func versionHelper(w io.Writer, service hasVersionInfo, label string) {
	version, _, err := service.GetVersion(&rpc.GetVersionOptions{})
	if err != nil {
		slogs.Logr.Debug("error getting network info from daemon", "error", err)
		version = &rpc.GetVersionResponse{Version: "Not Running"}
	}
	if version == nil {
		slogs.Logr.Debug("no network info found", "service", label)
		version = &rpc.GetVersionResponse{Version: "Not Running"}
	}
	_, _ = fmt.Fprintln(w, label, "\t", version.Version)
}
