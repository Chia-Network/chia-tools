package network

import (
	"encoding/json"
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
	"github.com/spf13/viper"
)

var switchCmd = &cobra.Command{
	Use:     "switch",
	Short:   "Switches the active network on this machine",
	Example: "chia-tools network switch testnet11",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		networkName := args[0]
		SwitchNetwork(networkName, true)
	},
}

func init() {
	switchCmd.PersistentFlags().String("introducer", "", "Override the default values for introducer host")
	switchCmd.PersistentFlags().String("dns-introducer", "", "Override the default values for dns-introducer host")
	switchCmd.PersistentFlags().String("bootstrap-peer", "", "Override the default value for seeder bootstrap peer")
	switchCmd.PersistentFlags().Uint16("full-node-port", 0, "Override the default values for the full node port")

	cobra.CheckErr(viper.BindPFlag("switch-introducer", switchCmd.PersistentFlags().Lookup("introducer")))
	cobra.CheckErr(viper.BindPFlag("switch-dns-introducer", switchCmd.PersistentFlags().Lookup("dns-introducer")))
	cobra.CheckErr(viper.BindPFlag("switch-bootstrap-peer", switchCmd.PersistentFlags().Lookup("bootstrap-peer")))
	cobra.CheckErr(viper.BindPFlag("switch-full-node-port", switchCmd.PersistentFlags().Lookup("full-node-port")))

	networkCmd.AddCommand(switchCmd)
}

// retainedSettings are the settings we want to keep track of when switching networks so we can swap back to them in the future
type retainedSettings struct {
	DNSServers          []string      `json:"dns_servers"`
	BootstrapPeers      []string      `json:"bootstrap_peers"`
	StaticPeers         []string      `json:"static_peers"`
	FullNodePeers       []config.Peer `json:"full_node_peers"`
	WalletFullNodePeers []config.Peer `json:"wallet_full_node_peers"`
}

// SwitchNetwork implements the logic to swap networks
func SwitchNetwork(networkName string, checkForRunningNode bool) {
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

	// Ensure we have network constants for the network trying to be swapped to
	if _, ok := cfg.NetworkOverrides.Constants[networkName]; !ok {
		slogs.Logr.Fatal("selected network does not exist in config's network override constants", "network", networkName)
	}
	var (
		netConfig config.NetworkConfig
		ok        bool
	)
	if netConfig, ok = cfg.NetworkOverrides.Config[networkName]; !ok {
		slogs.Logr.Fatal("selected network does not exist in config's network override config", "network", networkName)
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

	previousSettings := retainedSettings{
		DNSServers:          cfg.FullNode.DNSServers,
		BootstrapPeers:      cfg.Seeder.BootstrapPeers,
		StaticPeers:         cfg.Seeder.StaticPeers,
		FullNodePeers:       cfg.FullNode.FullNodePeers,
		WalletFullNodePeers: cfg.Wallet.FullNodePeers,
	}
	marshalled, err := json.Marshal(previousSettings)
	if err != nil {
		slogs.Logr.Fatal("error marshalling retained settings to json", "error", err)
	}
	err = os.WriteFile(path.Join(cacheFileDirOldNetwork, "settings.json"), marshalled, 0644)
	if err != nil {
		slogs.Logr.Fatal("error writing settings for old network", "error", err)
	}

	var settingsToRestore *retainedSettings
	newSettingsPath := path.Join(cacheFileDirNewNetwork, "settings.json")
	if _, err := os.Stat(newSettingsPath); err == nil {
		settings, err := os.ReadFile(newSettingsPath)
		if err != nil {
			slogs.Logr.Fatal("error reading stored settings for the new network", "error", err)
		}
		settingsToRestore = &retainedSettings{}
		err = json.Unmarshal(settings, settingsToRestore)
		if err != nil {
			slogs.Logr.Fatal("error unmarshalling stored settings for the new network", "error", err)
		}
	}

	// Check if Full Node is running
	if checkForRunningNode {
		slogs.Logr.Debug("initializing websocket client to ensure chia is stopped")
		rpcClient, err := rpc.NewClient(rpc.ConnectionModeWebsocket, rpc.WithAutoConfig(), rpc.WithSyncWebsocket())
		if err != nil {
			slogs.Logr.Fatal("error initializing RPC client", "error", err)
		}

		slogs.Logr.Info("Ensuring chia services are stopped")
		_, _, err = rpcClient.DaemonService.Exit()
		if err != nil {
			if !isConnectionRefused(err) {
				slogs.Logr.Fatal("error stopping chia services", "error", err)
			}
		}
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
	dnsIntroducerHosts := []string{"dns-introducer.chia.net"}
	fullNodePort := uint16(8444)
	var fullnodePeers []config.Peer
	var walletFullNodePeers []config.Peer
	peersFilePath := "db/peers.dat"
	walletPeersFilePath := "wallet/db/wallet_peers.dat"
	bootstrapPeers := []string{"node.chia.net"}
	var staticPeers []string
	if networkName != "mainnet" {
		introducerHost = fmt.Sprintf("introducer-%s.chia.net", networkName)
		dnsIntroducerHosts = []string{fmt.Sprintf("dns-introducer-%s.chia.net", networkName)}
		fullNodePort = uint16(58444)
		peersFilePath = fmt.Sprintf("db/peers-%s.dat", networkName)
		walletPeersFilePath = fmt.Sprintf("wallet/db/wallet_peers-%s.dat", networkName)
		bootstrapPeers = []string{fmt.Sprintf("node-%s.chia.net", networkName)}
	}

	// Any stored settings for the new network should be applied here, before any flags override them
	if settingsToRestore != nil {
		slogs.Logr.Info("restoring stored settings for this network")
		if len(settingsToRestore.DNSServers) > 0 {
			dnsIntroducerHosts = settingsToRestore.DNSServers
		}
		if len(settingsToRestore.BootstrapPeers) > 0 {
			bootstrapPeers = settingsToRestore.BootstrapPeers
		}
		if len(settingsToRestore.StaticPeers) > 0 {
			staticPeers = settingsToRestore.StaticPeers
		}
		if len(settingsToRestore.FullNodePeers) > 0 {
			fullnodePeers = settingsToRestore.FullNodePeers
		}
		if len(settingsToRestore.WalletFullNodePeers) > 0 {
			walletFullNodePeers = settingsToRestore.WalletFullNodePeers
		}
	}

	if introFlag := viper.GetString("switch-introducer"); introFlag != "" {
		introducerHost = introFlag
	}
	if dnsIntroFlag := viper.GetString("switch-dns-introducer"); dnsIntroFlag != "" {
		dnsIntroducerHosts = []string{dnsIntroFlag}
	}
	if bootPeer := viper.GetString("switch-bootstrap-peer"); bootPeer != "" {
		bootstrapPeers = []string{bootPeer}
	}
	// If there is a port in the config, use that, but still allow the flag to be the final say
	if netConfig.DefaultFullNodePort != 0 {
		fullNodePort = netConfig.DefaultFullNodePort
	}
	if portFlag := viper.GetUint16("switch-full-node-port"); portFlag != 0 {
		fullNodePort = portFlag
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
		"full_node.dns_servers":          dnsIntroducerHosts,
		"full_node.peers_file_path":      peersFilePath,
		"full_node.port":                 fullNodePort,
		"full_node.full_node_peers":      fullnodePeers,
		"full_node.introducer_peer.host": introducerHost,
		"full_node.introducer_peer.port": fullNodePort,
		"introducer.port":                fullNodePort,
		"seeder.port":                    fullNodePort,
		"seeder.other_peers_port":        fullNodePort,
		"seeder.bootstrap_peers":         bootstrapPeers,
		"seeder.static_peers":            staticPeers,
		"timelord.full_node_peers": []config.Peer{
			{
				Host: "localhost",
				Port: fullNodePort,
			},
		},
		"wallet.dns_servers":            dnsIntroducerHosts,
		"wallet.full_node_peers":        ensureAtLeastLocalPeer(walletFullNodePeers, fullNodePort),
		"wallet.introducer_peer.host":   introducerHost,
		"wallet.introducer_peer.port":   fullNodePort,
		"wallet.wallet_peers_file_path": walletPeersFilePath,
	}
	for configPath, value := range pathUpdates {
		pathMap := config.ParsePathsFromStrings([]string{configPath}, false)
		var key string
		var pathSlice []string
		for key, pathSlice = range pathMap {
			break
		}
		slogs.Logr.Debug("setting config path", "path", configPath, "value", value)
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

	err = removeFileIfExists(path.Join(chiaRoot, "db", peersFilePath))
	if err != nil {
		slogs.Logr.Error("error removing old peers.dat file", "path", peersFilePath, "error", err)
	}

	slogs.Logr.Info("Complete")
}

func ensureAtLeastLocalPeer(peers []config.Peer, port uint16) []config.Peer {
	if len(peers) == 0 {
		peers = append(peers, config.Peer{
			Host: "localhost",
			Port: port,
		})
	}

	return peers
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

func removeFileIfExists(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			slogs.Logr.Debug("source path doesn't exist, skipping delete", "path", path)
			return nil
		}
		return fmt.Errorf("error checking source file: %w", err)
	}

	slogs.Logr.Debug("removing file at path", "path", path)
	err := os.Remove(path)
	if err != nil {
		return fmt.Errorf("error removing file: %w", err)
	}

	return nil
}
