package network

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"syscall"

	"github.com/chia-network/go-chia-libs/pkg/config"
	"github.com/chia-network/go-chia-libs/pkg/rpc"
	"github.com/chia-network/go-modules/pkg/slogs"
	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:     "switch",
	Short:   "Switches the active network on this machine",
	Example: "chia-tools network switch testneta",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Track if we stopped the full node, and start it again at the end if so
		stoppedFullNode := false

		networkName := args[0]
		slogs.Logr.Info("Swapping to network", "network", networkName)

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

		currentNetwork := *cfg.SelectedNetwork
		slogs.Logr.Info("discovered current network", "current-network", currentNetwork)

		if currentNetwork == networkName {
			slogs.Logr.Fatal("current network name and new network name are the same", "current", currentNetwork, "new", networkName)
		}

		// Ensure a folder to store the current network's sub-epoch-summaries and height-to-hash files exists
		cacheFileDirOldNetwork := path.Join(chiaRoot, "db", currentNetwork)
		cacheFileDirNewNetwork := path.Join(chiaRoot, "db", networkName)

		slogs.Logr.Debug("ensuring directory exists for current network cache files", "directory", cacheFileDirOldNetwork)
		err = os.MkdirAll(cacheFileDirOldNetwork, 0755)
		if err != nil {
			slogs.Logr.Fatal("error creating cache file directory for current network", "error", err, "directory", cacheFileDirOldNetwork)
		}

		slogs.Logr.Debug("ensuring directory exists for new network cache files", "directory", cacheFileDirNewNetwork)
		err = os.MkdirAll(cacheFileDirNewNetwork, 0755)
		if err != nil {
			slogs.Logr.Fatal("error creating cache file directory for new network", "error", err, "directory", cacheFileDirNewNetwork)
		}

		// Check if Full Node is running
		slogs.Logr.Debug("initializing websocket client to ensure chia is stopped")
		rpcClient, err := rpc.NewClient(rpc.ConnectionModeWebsocket, rpc.WithAutoConfig(), rpc.WithSyncWebsocket())
		if err != nil {
			slogs.Logr.Fatal("error initializing RPC client", "error", err)
		}

		slogs.Logr.Debug("checking if full node is running")
		fullNodeState, _, err := rpcClient.DaemonService.IsRunning(&rpc.IsRunningOptions{Service: rpc.ServiceFullNameNode})
		if err != nil {
			if !isConnectionRefused(err) {
				slogs.Logr.Fatal("error checking full node status", "error", err)
			}
			// Was connection refused, so just give a mock fullNodeState
			fullNodeState = &rpc.IsRunningResponse{
				IsRunning: false,
			}
		}
		if fullNodeState.IsRunning {
			slogs.Logr.Info("Stopping full node")
			stopResult, _, err := rpcClient.DaemonService.StopService(&rpc.StopServiceOptions{Service: rpc.ServiceFullNameNode})
			if err != nil {
				slogs.Logr.Fatal("error stopping full node service", "error", err)
			}
			if !stopResult.Success {
				slogs.Logr.Fatal("unknown error stopping full node. Stop chia services manually and try again")
			}
			slogs.Logr.Info("Successfully stopped full node service")
			stoppedFullNode = true
		}

		// Safe to move files now
		activeSubEpochSummariesPath := path.Join(chiaRoot, "db", "sub-epoch-summaries")
		activeHeightToHashPath := path.Join(chiaRoot, "db", "height-to-hash")

		// Move current cache files to the network subdir
		err = moveAndOverwriteFile(activeSubEpochSummariesPath, path.Join(cacheFileDirOldNetwork, "sub-epoch-summaries"))
		if err != nil {
			slogs.Logr.Fatal("error moving sub-epoch-summaries file", "error", err)
		}
		err = moveAndOverwriteFile(activeHeightToHashPath, path.Join(cacheFileDirOldNetwork, "height-to-hash"))
		if err != nil {
			slogs.Logr.Fatal("error moving height-to-hash file", "error", err)
		}

		// Move old cached files to active dir
		err = moveAndOverwriteFile(path.Join(cacheFileDirNewNetwork, "sub-epoch-summaries"), activeSubEpochSummariesPath)
		if err != nil {
			slogs.Logr.Fatal("error moving sub-epoch-summaries file", "error", err)
		}
		err = moveAndOverwriteFile(path.Join(cacheFileDirNewNetwork, "height-to-hash"), activeHeightToHashPath)
		if err != nil {
			slogs.Logr.Fatal("error moving height-to-hash file", "error", err)
		}

		introducerHost := "introducer.chia.net"
		dnsIntroducerHost := "dns-introducer.chia.net"
		fullNodePort := uint16(8444)
		peersFilePath := "peers.dat"
		walletPeersFilePath := "wallet/db/wallet_peers.dat"
		bootstrapPeers := []string{"node.chia.net"}
		if networkName != "mainnet" {
			introducerHost = fmt.Sprintf("introducer-%s.chia.net", networkName)
			dnsIntroducerHost = fmt.Sprintf("dns-introducer-%s.chia.net", networkName)
			fullNodePort = uint16(58444)
			peersFilePath = fmt.Sprintf("peers-%s.dat", networkName)
			walletPeersFilePath = fmt.Sprintf("wallet/db/wallet_peers-%s.dat", networkName)
			bootstrapPeers = []string{fmt.Sprintf("node-%s.chia.net", networkName)}
		}
		pathUpdates := map[string]any{
			"selected_network": networkName,
			"farmer.full_node_peers": []config.Peer{
				{
					Host: "localhost",
					Port: fullNodePort,
				},
			},
			"full_node.database_path":        fmt.Sprintf("db/blockchain_v2_%s.sqlite", networkName),
			"full_node.dns_servers":          []string{dnsIntroducerHost},
			"full_node.peers_file_path":      peersFilePath,
			"full_node.port":                 fullNodePort,
			"full_node.introducer_peer.host": introducerHost,
			"full_node.introducer_peer.port": fullNodePort,
			"introducer.port":                fullNodePort,
			"seeder.port":                    fullNodePort,
			"seeder.other_peers_port":        fullNodePort,
			"seeder.bootstrap_peers":         bootstrapPeers,
			"timelord.full_node_peers": []config.Peer{
				{
					Host: "localhost",
					Port: fullNodePort,
				},
			},
			"wallet.dns_servers": []string{dnsIntroducerHost},
			"wallet.full_node_peers": []config.Peer{
				{
					Host: "localhost",
					Port: fullNodePort,
				},
			},
			"wallet.introducer_peer.host":   introducerHost,
			"wallet.introducer_peer.port":   fullNodePort,
			"wallet.wallet_peers_file_path": walletPeersFilePath,
		}
		for path, value := range pathUpdates {
			pathMap := config.ParsePathsFromStrings([]string{path}, false)
			var key string
			var pathSlice []string
			for key, pathSlice = range pathMap {
				break
			}
			slogs.Logr.Debug("setting config path", "path", path, "value", value)
			err = cfg.SetFieldByPath(pathSlice, value)
			if err != nil {
				slogs.Logr.Fatal("error setting path in config", "key", key, "value", value, "error", err)
			}
		}

		slogs.Logr.Debug("saving config")
		err = cfg.Save()
		if err != nil {
			slogs.Logr.Fatal("error saving chia config", "error", err)
		}

		if stoppedFullNode {
			// Start full node
			slogs.Logr.Info("Starting full node")
			startResult, _, err := rpcClient.DaemonService.StartService(&rpc.StartServiceOptions{Service: rpc.ServiceFullNameNode})
			if err != nil {
				slogs.Logr.Error("error starting full node", "error", err)
			}
			if !startResult.Success {
				slogs.Logr.Error("Unknown error starting full node")
			}
			slogs.Logr.Info("Successfully started full node")
		}

		slogs.Logr.Info("Complete")
	},
}

func init() {
	networkCmd.AddCommand(switchCmd)
}

func isConnectionRefused(err error) bool {
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		if netErr.Op == "dial" {
			var syscallError *os.SyscallError
			if errors.As(netErr.Err, &syscallError) {
				return syscallError.Syscall == "connect" && errors.Is(syscallError.Err, syscall.ECONNREFUSED)
			}
		}
	}
	return false
}

func moveAndOverwriteFile(sourcePath, destPath string) error {
	if _, err := os.Stat(sourcePath); err != nil {
		if os.IsNotExist(err) {
			slogs.Logr.Debug("source path doesn't exist, skipping move", "source", sourcePath, "dest", destPath)
			return nil
		}
		return fmt.Errorf("error checking source file: %w", err)
	}

	// Remove the destination file if it exists
	slogs.Logr.Debug("checking if destination file exists before moving", "dest", destPath)
	if _, err := os.Stat(destPath); err == nil {
		slogs.Logr.Debug("Destination file already exists. Deleting", "dest", destPath)
		err = os.Remove(destPath)
		if err != nil {
			return fmt.Errorf("error removing destination file: %w", err)
		}
	}

	slogs.Logr.Debug("moving file to destination", "source", sourcePath, "dest", destPath)
	err := os.Rename(sourcePath, destPath)
	if err != nil {
		return fmt.Errorf("error moving file: %w", err)
	}

	slogs.Logr.Debug("moved successfully", "source", sourcePath, "dest", destPath)
	return nil
}
