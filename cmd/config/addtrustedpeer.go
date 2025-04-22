package config

import (
	"encoding/hex"
	"net"
	"path"
	"strconv"

	"github.com/chia-network/go-chia-libs/pkg/config"
	"github.com/chia-network/go-chia-libs/pkg/peerprotocol"
	"github.com/chia-network/go-modules/pkg/slogs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/chia-network/chia-tools/internal/utils"
)

var (
	skipConfirm bool
)

// addTrustedPeerCmd Adds a trusted peer to the config
var addTrustedPeerCmd = &cobra.Command{
	Use:   "add-trusted-peer",
	Short: "Adds a trusted peer to the config file",
	Example: `chia-tools config add-trusted-peer 1.2.3.4

# The following version will also override the port to use when connecting to this peer
chia-tools config add-trusted-peer 1.2.3.4 18444

# You may also specify a DNS name. The tool will attempt to resolve the name to an IP address.
# If the name resolves to multiple IP addresses, chia-tools will attempt to connect to each one to add it to the config.
chia-tools config add-trusted-peer node.chia.net 8444`,
	Run: func(cmd *cobra.Command, args []string) {
		chiaRoot, err := config.GetChiaRootPath()
		if err != nil {
			slogs.Logr.Fatal("Unable to determine CHIA_ROOT", "error", err)
		}

		// 1: Peer IP
		// 2: Optional, port
		if len(args) < 1 || len(args) > 2 {
			slogs.Logr.Fatal("Unexpected number of arguments provided")
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
			addTrustedPeer(cfg, chiaRoot, ip, port)
		}
	},
}

func addTrustedPeer(cfg *config.ChiaConfig, chiaRoot string, ip net.IP, port uint16) {
	slogs.Logr.Info("Attempting to get peer id", "peer", ip.String(), "port", port)

	keypair, err := cfg.FullNode.SSL.LoadPublicKeyPair(chiaRoot)
	if err != nil {
		slogs.Logr.Fatal("Error loading certs from CHIA_ROOT", "CHIA_ROOT", chiaRoot, "error", err)
	}
	if keypair == nil {
		slogs.Logr.Fatal("Error loading certs from CHIA_ROOT", "CHIA_ROOT", chiaRoot, "error", "keypair was nil")
	}
	conn, err := peerprotocol.NewConnection(
		&ip,
		peerprotocol.WithPeerPort(port),
		peerprotocol.WithNetworkID(*cfg.SelectedNetwork),
		peerprotocol.WithPeerKeyPair(*keypair),
	)
	if err != nil {
		slogs.Logr.Fatal("Error creating connection", "error", err)
	}
	peerID, err := conn.PeerID()
	if err != nil {
		slogs.Logr.Fatal("Error getting peer id", "error", err)
	}
	peerIDStr := hex.EncodeToString(peerID[:])
	slogs.Logr.Info("peer id received", "peer", peerIDStr)
	if !utils.ConfirmAction("Would you like trust this peer? (y/N)", skipConfirm) {
		slogs.Logr.Error("Cancelled")
	}
	cfg.Wallet.TrustedPeers[peerIDStr] = "Does_not_matter"

	peerToAdd := config.Peer{
		Host: ip.String(),
		Port: port,
	}

	foundPeer := false
	for idx, peer := range cfg.Wallet.FullNodePeers {
		if peer.Host == ip.String() {
			foundPeer = true
			cfg.Wallet.FullNodePeers[idx] = peerToAdd
		}
	}
	if !foundPeer {
		cfg.Wallet.FullNodePeers = append(cfg.Wallet.FullNodePeers, peerToAdd)
	}

	err = cfg.Save()
	if err != nil {
		slogs.Logr.Fatal("error saving config", "error", err)
	}

	slogs.Logr.Info("Added trusted peer. Restart your chia services for the configuration to take effect")
}

func init() {
	addTrustedPeerCmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "Skip confirmation")
	configCmd.AddCommand(addTrustedPeerCmd)
}
