package network

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"

	"github.com/chia-network/go-chia-libs/pkg/config"
	"github.com/chia-network/go-chia-libs/pkg/ptr"
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
			HardForkHeight:                 ptr.Uint32Ptr(0),
		}
		if viper.IsSet("tn-gen-hard-fork2-height") {
			constants.HardFork2Height = ptr.Uint32Ptr(viper.GetUint32("tn-gen-hard-fork2-height"))
		}
		if viper.IsSet("tn-gen-number-zero-bits-plot-filter-v2") {
			value := cast.ToUint8(viper.Get("tn-gen-number-zero-bits-plot-filter-v2"))
			constants.NumberZeroBitsPlotFilterV2 = &value
		}
		if viper.IsSet("tn-gen-plot-v1-phase-out-epoch-bits") {
			value := cast.ToUint8(viper.Get("tn-gen-plot-v1-phase-out-epoch-bits"))
			constants.PlotV1PhaseOutEpochBits = &value
		}
		if viper.IsSet("tn-gen-plot-filter-v2-relative-height") {
			raw := cast.ToUintSlice(viper.Get("tn-gen-plot-filter-v2-relative-height"))
			if len(raw) != config.PlotFilterV2RelativeHeightLen {
				slogs.Logr.Fatal("plot-filter-v2-relative-height must have exactly 9 entries (reverse chronological, relative to hard-fork2-height)", "got", len(raw))
			}
			heights := make([]uint32, 0, len(raw))
			for _, h := range raw {
				if h > math.MaxUint32 {
					slogs.Logr.Fatal("plot-filter-v2-relative-height entry exceeds uint32", "value", h)
				}
				heights = append(heights, uint32(h))
			}
			constants.PlotFilterV2RelativeHeight = heights
		}
		if viper.IsSet("tn-gen-filter-window-size") {
			value := cast.ToUint8(viper.Get("tn-gen-filter-window-size"))
			constants.FilterWindowSize = &value
		}
		if viper.IsSet("tn-gen-max-effective-plot-filter-bits") {
			value := cast.ToUint8(viper.Get("tn-gen-max-effective-plot-filter-bits"))
			constants.MaxEffectivePlotFilterBits = &value
		}
		if viper.IsSet("tn-gen-soft-fork-8-9-height") {
			constants.SoftFork8Height = ptr.Uint32Ptr(viper.GetUint32("tn-gen-soft-fork-8-9-height"))
			constants.SoftFork9Height = ptr.Uint32Ptr(viper.GetUint32("tn-gen-soft-fork-8-9-height"))
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
	generateCmd.PersistentFlags().Uint64("difficulty-starting", uint64(250), "Specify starting difficulty")
	generateCmd.PersistentFlags().Uint64("sub-slot-iters-starting", uint64(1<<25), "Specify starting sub slot iters")
	generateCmd.PersistentFlags().Uint16("port", uint16(58445), "Specify the port the network full nodes should use")
	// New configuration options to support testing hard fork 2
	generateCmd.PersistentFlags().Uint32("hard-fork2-height", uint32(0), "Block height when the 3.0 hard fork will activate")
	generateCmd.PersistentFlags().Uint8("number-zero-bits-plot-filter-v2", uint8(0), "Number of leading zeroes required to pass plot ID filter (post hard fork only)")
	generateCmd.PersistentFlags().Uint8("plot-v1-phase-out-epoch-bits", uint8(0), "Number of bits in phase out period (eg 8 bits = 256 epochs)")
	generateCmd.PersistentFlags().UintSlice("plot-filter-v2-relative-height", []uint{}, "Comma-separated list of exactly 9 heights relative to hard-fork2-height where the v2 base plot filter drops by one bit, in reverse chronological order (first entry activates last)")
	generateCmd.PersistentFlags().Uint8("filter-window-size", uint8(0), "Number of signage points per v2 plot filter window (mainnet 16)")
	generateCmd.PersistentFlags().Uint8("max-effective-plot-filter-bits", uint8(0), "Cap on the effective v2 plot filter bits (mainnet 13)")
	// Soft fork 8/9 testing option
	generateCmd.PersistentFlags().Uint32("soft-fork-8-9-height", uint32(0), "Block height to activate soft fork 8 and 9")
	// Output format options
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
	cobra.CheckErr(viper.BindPFlag("tn-gen-hard-fork2-height", generateCmd.PersistentFlags().Lookup("hard-fork2-height")))
	cobra.CheckErr(viper.BindPFlag("tn-gen-number-zero-bits-plot-filter-v2", generateCmd.PersistentFlags().Lookup("number-zero-bits-plot-filter-v2")))
	cobra.CheckErr(viper.BindPFlag("tn-gen-plot-v1-phase-out-epoch-bits", generateCmd.PersistentFlags().Lookup("plot-v1-phase-out-epoch-bits")))
	cobra.CheckErr(viper.BindPFlag("tn-gen-plot-filter-v2-relative-height", generateCmd.PersistentFlags().Lookup("plot-filter-v2-relative-height")))
	cobra.CheckErr(viper.BindPFlag("tn-gen-filter-window-size", generateCmd.PersistentFlags().Lookup("filter-window-size")))
	cobra.CheckErr(viper.BindPFlag("tn-gen-max-effective-plot-filter-bits", generateCmd.PersistentFlags().Lookup("max-effective-plot-filter-bits")))
	cobra.CheckErr(viper.BindPFlag("tn-gen-soft-fork-8-9-height", generateCmd.PersistentFlags().Lookup("soft-fork-8-9-height")))
	cobra.CheckErr(viper.BindPFlag("tn-gen-as-json", generateCmd.PersistentFlags().Lookup("as-json")))
	cobra.CheckErr(viper.BindPFlag("tn-gen-with-constants", generateCmd.PersistentFlags().Lookup("with-constants")))

	networkCmd.AddCommand(generateCmd)
}
