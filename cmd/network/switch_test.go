package network_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/chia-network/go-chia-libs/pkg/config"
	"github.com/chia-network/go-chia-libs/pkg/types"
	"github.com/stretchr/testify/assert"

	"github.com/chia-network/chia-tools/cmd"
	"github.com/chia-network/chia-tools/cmd/network"
)

const testnetwork = "unittestnet"

func setupDefaultConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "chia-root-*")
	assert.NoError(t, err)

	err = os.Setenv("CHIA_ROOT", tempDir)
	assert.NoError(t, err)

	rootPath, err := config.GetChiaRootPath()
	assert.NoError(t, err)

	err = os.MkdirAll(filepath.Join(rootPath, "config"), 0755)
	assert.NoError(t, err)

	defaultConfig, err := config.LoadDefaultConfig()
	assert.NoError(t, err)

	defaultConfig.NetworkOverrides.Constants[testnetwork] = config.NetworkConstants{
		AggSigMeAdditionalData:         "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		DifficultyConstantFactor:       types.Uint128From64(10052721566054),
		DifficultyStarting:             30,
		EpochBlocks:                    768,
		GenesisChallenge:               "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		GenesisPreFarmPoolPuzzleHash:   "08296fc227decd043aee855741444538e4cc9a31772c4d1a9e6242d1e777e42a",
		GenesisPreFarmFarmerPuzzleHash: "08296fc227decd043aee855741444538e4cc9a31772c4d1a9e6242d1e777e42a",
		MempoolBlockBuffer:             10,
		MinPlotSize:                    18,
		NetworkType:                    1,
		SubSlotItersStarting:           67108864,
	}
	defaultConfig.NetworkOverrides.Config[testnetwork] = config.NetworkConfig{
		AddressPrefix:       "txch",
		DefaultFullNodePort: 58445,
	}

	configPath := filepath.Join(rootPath, "config", "config.yaml")
	err = defaultConfig.SavePath(configPath)
	assert.NoError(t, err)
}

func TestNetworkSwitch(t *testing.T) {
	cmd.InitLogs()
	setupDefaultConfig(t)
	cfg, err := config.GetChiaConfig()
	assert.NoError(t, err)
	assert.Equal(t, "mainnet", *cfg.SelectedNetwork)

	network.SwitchNetwork("unittestnet", false)

	// reload config from disk
	cfg, err = config.GetChiaConfig()
	assert.NoError(t, err)

	port := uint16(58445)
	localpeer := config.Peer{
		Host: "localhost",
		Port: port,
	}
	assert.Equal(t, "unittestnet", *cfg.SelectedNetwork)
	assert.Equal(t, []config.Peer{localpeer}, cfg.Farmer.FullNodePeers)
	assert.Equal(t, "db/blockchain_v2_unittestnet.sqlite", cfg.FullNode.DatabasePath)
	assert.Equal(t, []string{"dns-introducer-unittestnet.chia.net"}, cfg.FullNode.DNSServers)
	assert.Equal(t, "peers-unittestnet.dat", cfg.FullNode.PeersFilePath)
	assert.Equal(t, port, cfg.FullNode.Port)
	assert.Equal(t, config.Peer{Host: "introducer-unittestnet.chia.net", Port: port}, cfg.FullNode.IntroducerPeer)
	assert.Equal(t, port, cfg.Introducer.Port)
	assert.Equal(t, port, cfg.Seeder.OtherPeersPort)
	assert.Equal(t, []string{"node-unittestnet.chia.net"}, cfg.Seeder.BootstrapPeers)
	assert.Equal(t, []config.Peer{localpeer}, cfg.Timelord.FullNodePeers)
	assert.Equal(t, []string{"dns-introducer-unittestnet.chia.net"}, cfg.Wallet.DNSServers)
	assert.Equal(t, []config.Peer{localpeer}, cfg.Wallet.FullNodePeers)
	assert.Equal(t, config.Peer{Host: "introducer-unittestnet.chia.net", Port: port}, cfg.Wallet.IntroducerPeer)
	assert.Equal(t, "wallet/db/wallet_peers-unittestnet.dat", cfg.Wallet.WalletPeersFilePath)
}
