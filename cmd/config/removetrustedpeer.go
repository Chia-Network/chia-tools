package config

import (
	"net"
	"path"
	"strconv"

	"github.com/chia-network/go-chia-libs/pkg/config"
	"github.com/chia-network/go-modules/pkg/slogs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/chia-network/chia-tools/internal/utils"
)

var (
	removeAll bool
)

// removeTrustedPeerCmd removes a trusted peer from the config
var removeTrustedPeerCmd = &cobra.Command{
	Use:   "remove-trusted-peer",
	Short: "Removes a trusted peer from the config file",
	Example: `chia-tools config remove-trusted-peer 1.2.3.4

# The following version will also override the port to use when connecting to this peer
chia-tools config remove-trusted-peer 1.2.3.4 18444

# You may also specify a DNS name. The tool will attempt to resolve the name to an IP address.
# If the name resolves to multiple IP addresses, chia-tools will attempt to connect to each one to remove it from the config.
chia-tools config remove-trusted-peer node.chia.net 8444

# You can also remove all trusted peers by specifying the --all flag
chia-tools config remove-trusted-peer --all`,
	Run: func(cmd *cobra.Command, args []string) {
		chiaRoot, err := config.GetChiaRootPath()
		if err != nil {
			slogs.Logr.Fatal("Unable to determine CHIA_ROOT", "error", err)
		}

		cfgPath := viper.GetString("config")
		if cfgPath == "" {
			// Use default chia root
			cfgPath = path.Join(chiaRoot, "config", "config.yaml")
		}

		cfg, err := config.LoadConfigAtRoot(cfgPath, chiaRoot)
		if err != nil {
			slogs.Logr.Fatal("error loading chia config", "error", err)
		}

		if removeAll {
			removeAllTrustedPeers(cfg)
			return
		}

		// 1: Peer IP
		// 2: Optional, port
		if len(args) < 1 || len(args) > 2 {
			slogs.Logr.Fatal("Unexpected number of arguments provided")
		}

		peer := args[0]
		port := cfg.FullNode.Port
		if len(args) > 1 {
			port64, err := strconv.ParseUint(args[1], 10, 16)
			if err != nil {
				slogs.Logr.Fatal("Invalid port provided")
			}
			port = uint16(port64)
		}

		var ips []net.IP

		ip := net.ParseIP(peer)
		if ip == nil {
			// Try to resolve a DNS name
			ips, err = net.LookupIP(peer)
			if err != nil {
				slogs.Logr.Fatal("Couldn't parse peer as IP address or resolve to a host", "id", peer)
			}
			if len(ips) == 0 {
				slogs.Logr.Fatal("dns lookup returned 0 IPs ", "id", peer)
			}
		} else {
			ips = append(ips, ip)
		}

		for _, ip := range ips {
			removeTrustedPeer(cfg, chiaRoot, ip, port)
		}
	},
}

func removeTrustedPeer(cfg *config.ChiaConfig, chiaRoot string, ip net.IP, port uint16) {
	peerIDStr := getPeerID(cfg, chiaRoot, ip, port)
	slogs.Logr.Info("peer id received", "peer", peerIDStr)

	if !utils.ConfirmAction("Would you like stop trusting this peer? (y/N)", skipConfirm) {
		slogs.Logr.Error("Cancelled")
		return
	}

	// Remove trusted peer
	delete(cfg.Wallet.TrustedPeers, peerIDStr)

	// Remove full_node peer if found
	fullNodePeers := make([]config.Peer, 0)
	for _, peer := range cfg.Wallet.FullNodePeers {
		if peer.Host != ip.String() && peer.Port != port {
			fullNodePeers = append(fullNodePeers, peer)
		}
	}
	cfg.Wallet.FullNodePeers = fullNodePeers

	err := cfg.Save()
	if err != nil {
		slogs.Logr.Fatal("error saving config", "error", err)
	}

	slogs.Logr.Info("Removed trusted peer. Restart your chia services for the configuration to take effect")
}

func removeAllTrustedPeers(cfg *config.ChiaConfig) {
	if !utils.ConfirmAction("Are you sure you would like to remove all trusted peers? (y/N)", skipConfirm) {
		slogs.Logr.Error("Cancelled")
		return
	}

	// Reset trusted peers map to the default
	cfg.Wallet.TrustedPeers = make(map[string]string)
	cfg.Wallet.TrustedPeers["0ThisisanexampleNodeID7ff9d60f1c3fa270c213c0ad0cb89c01274634a7c3cb9"] = "Does_not_matter"

	// Reset full_node peers list to just localhost
	cfg.Wallet.FullNodePeers = make([]config.Peer, 0)
	cfg.Wallet.FullNodePeers = append(cfg.Wallet.FullNodePeers, config.Peer{
		Host: "localhost",
		Port: cfg.FullNode.Port,
	})

	err := cfg.Save()
	if err != nil {
		slogs.Logr.Fatal("error saving config", "error", err)
	}

	slogs.Logr.Info("Removed all trusted peers. Restart your chia services for the configuration to take effect")
}

func init() {
	removeTrustedPeerCmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "Skip confirmation")
	removeTrustedPeerCmd.Flags().BoolVarP(&removeAll, "all", "a", false, "Remove all trusted peers from the config file")
	configCmd.AddCommand(removeTrustedPeerCmd)
}
