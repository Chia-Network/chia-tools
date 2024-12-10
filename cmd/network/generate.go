package network

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/chia-network/go-chia-libs/pkg/config"
	"github.com/chia-network/go-chia-libs/pkg/types"
	"github.com/chia-network/go-modules/pkg/slogs"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:     "generate",
	Short:   "Generates new network constants",
	Example: "chia-tools network generate --network examplenet",
	Run: func(cmd *cobra.Command, args []string) {
		networkName := viper.GetString("tn-gen-network")
		genesisHashBytes := sha256.Sum256([]byte(networkName))
		genesisHash := hex.EncodeToString(genesisHashBytes[:32])

		constants := &config.NetworkConstants{
			AggSigMeAdditionalData:         genesisHash,
			DifficultyConstantFactor:       types.Uint128From64(viper.GetUint64("tn-gen-diff-constant-factor")),
			DifficultyStarting:             viper.GetUint64("tn-gen-difficulty-starting"),
			EpochBlocks:                    viper.GetUint32("tn-gen-epoch-blocks"),
			GenesisChallenge:               genesisHash,
			GenesisPreFarmPoolPuzzleHash:   viper.GetString("tn-gen-pre-farm-pool-puz-hash"),
			GenesisPreFarmFarmerPuzzleHash: viper.GetString("tn-gen-pre-farm-farmer-puz-hash"),
			MempoolBlockBuffer:             cast.ToUint8(viper.Get("tn-gen-mempool-block-buffer")),
			MinPlotSize:                    cast.ToUint8(viper.Get("tn-gen-min-plot-size")),
			NetworkType:                    1,
			SubSlotItersStarting:           viper.GetUint64("tn-gen-sub-slot-iters-starting"),
		}
		cfg := &config.NetworkConfig{
			AddressPrefix:       "txch",
			DefaultFullNodePort: viper.GetUint16("tn-gen-port"),
		}

		netOverrides := &config.NetworkOverrides{
			Constants: map[string]config.NetworkConstants{
				networkName: *constants,
			},
			Config: map[string]config.NetworkConfig{
				networkName: *cfg,
			},
		}

		var toMarshal any
		if viper.GetBool("tn-gen-with-constants") {
			toMarshal = netOverrides
		} else {
			toMarshal = constants
		}

		var marshalled []byte
		var err error
		if viper.GetBool("tn-gen-as-json") {
			marshalled, err = json.Marshal(toMarshal)
		} else {
			marshalled, err = yaml.Marshal(toMarshal)
		}

		if err != nil {
			slogs.Logr.Fatal("error marshalling", "error", err)
		}
		fmt.Print(string(marshalled))
	},
}

func init() {
	generateCmd.PersistentFlags().String("network", "", "Name of the network to create")
	generateCmd.PersistentFlags().Uint64("diff-constant-factor", uint64(10052721566054), "Specify the value for DIFFICULTY_CONSTANT_FACTOR (Up to uint64max)")
	generateCmd.PersistentFlags().String("pre-farm-farmer-puz-hash", "08296fc227decd043aee855741444538e4cc9a31772c4d1a9e6242d1e777e42a", "Specify the value for GENESIS_PRE_FARM_FARMER_PUZZLE_HASH")
	generateCmd.PersistentFlags().String("pre-farm-pool-puz-hash", "08296fc227decd043aee855741444538e4cc9a31772c4d1a9e6242d1e777e42a", "Specify the value for GENESIS_PRE_FARM_POOL_PUZZLE_HASH")
	generateCmd.PersistentFlags().Uint8("min-plot-size", uint8(18), "Specify the minimum plot size MIN_PLOT_SIZE")
	generateCmd.PersistentFlags().Uint8("mempool-block-buffer", uint8(10), "Specify MEMPOOL_BLOCK_BUFFER")
	generateCmd.PersistentFlags().Uint32("epoch-blocks", uint32(768), "specify EPOCH_BLOCKS")
	generateCmd.PersistentFlags().Uint64("difficulty-starting", uint64(30), "Specify starting difficulty")
	generateCmd.PersistentFlags().Uint64("sub-slot-iters-starting", uint64(1<<26), "Specify starting sub slot iters")
	generateCmd.PersistentFlags().Uint16("port", uint16(58445), "Specify the port the network full nodes should use")
	generateCmd.PersistentFlags().Bool("as-json", false, "Output as JSON blob instead of yaml")
	generateCmd.PersistentFlags().Bool("with-constants", false, "Include constants and default ports")

	cobra.CheckErr(viper.BindPFlag("tn-gen-network", generateCmd.PersistentFlags().Lookup("network")))
	cobra.CheckErr(viper.BindPFlag("tn-gen-diff-constant-factor", generateCmd.PersistentFlags().Lookup("diff-constant-factor")))
	cobra.CheckErr(viper.BindPFlag("tn-gen-pre-farm-farmer-puz-hash", generateCmd.PersistentFlags().Lookup("pre-farm-farmer-puz-hash")))
	cobra.CheckErr(viper.BindPFlag("tn-gen-pre-farm-pool-puz-hash", generateCmd.PersistentFlags().Lookup("pre-farm-pool-puz-hash")))
	cobra.CheckErr(viper.BindPFlag("tn-gen-min-plot-size", generateCmd.PersistentFlags().Lookup("min-plot-size")))
	cobra.CheckErr(viper.BindPFlag("tn-gen-mempool-block-buffer", generateCmd.PersistentFlags().Lookup("mempool-block-buffer")))
	cobra.CheckErr(viper.BindPFlag("tn-gen-epoch-blocks", generateCmd.PersistentFlags().Lookup("epoch-blocks")))
	cobra.CheckErr(viper.BindPFlag("tn-gen-difficulty-starting", generateCmd.PersistentFlags().Lookup("difficulty-starting")))
	cobra.CheckErr(viper.BindPFlag("tn-gen-sub-slot-iters-starting", generateCmd.PersistentFlags().Lookup("sub-slot-iters-starting")))
	cobra.CheckErr(viper.BindPFlag("tn-gen-port", generateCmd.PersistentFlags().Lookup("port")))
	cobra.CheckErr(viper.BindPFlag("tn-gen-as-json", generateCmd.PersistentFlags().Lookup("as-json")))
	cobra.CheckErr(viper.BindPFlag("tn-gen-with-constants", generateCmd.PersistentFlags().Lookup("with-constants")))

	networkCmd.AddCommand(generateCmd)
}
