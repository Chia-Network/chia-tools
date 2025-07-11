package coins

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/chia-network/go-chia-libs/pkg/rpc"
	"github.com/chia-network/go-chia-libs/pkg/types"
	"github.com/chia-network/go-modules/pkg/slogs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// splitLargestCmd represents the split-largest command
var splitLargestCmd = &cobra.Command{
	Use:   "split-largest",
	Short: "Find the largest coin in the wallet and split it into smaller coins",
	Example: `chia-tools coins split-largest --fingerprint 123456789 --amount-per-coin 0.001 --number-of-coins 10
chia-tools coins split-largest --id 1 --amount-per-coin 0.001 --number-of-coins 10 --fee 0.0001`,
	Run: func(cmd *cobra.Command, args []string) {
		amountPerCoinStr := viper.GetString("coins-amount-per-coin")
		numberOfCoins := viper.GetUint32("coins-number-of-coins")

		if amountPerCoinStr == "" {
			slogs.Logr.Fatal("amount-per-coin must be specified")
		}
		if numberOfCoins == 0 {
			slogs.Logr.Fatal("number-of-coins must be greater than 0")
		}

		SplitLargestCoin()
	},
}

// convertXCHToMojos converts a decimal XCH amount to mojos
func convertXCHToMojos(xchStr string) (uint64, error) {
	// Handle empty or zero values
	if xchStr == "" || xchStr == "0" {
		return 0, nil
	}

	// Remove any trailing zeros after decimal point for cleaner parsing
	xchStr = strings.TrimRight(strings.TrimRight(xchStr, "0"), ".")

	// If we end up with an empty string after trimming, it was just "0"
	if xchStr == "" {
		return 0, nil
	}

	// Split by decimal point
	parts := strings.Split(xchStr, ".")
	if len(parts) > 2 {
		return 0, fmt.Errorf("invalid XCH amount format: %s", xchStr)
	}

	// Parse the whole number part
	whole, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid whole number part: %s", parts[0])
	}

	// Convert to mojos (1 XCH = 1,000,000,000,000 mojos)
	mojos := whole * 1000000000000

	// If there's a decimal part, handle it
	if len(parts) == 2 {
		decimalPart := parts[1]
		// Pad or truncate to 12 decimal places
		if len(decimalPart) > 12 {
			decimalPart = decimalPart[:12]
		} else {
			decimalPart = decimalPart + strings.Repeat("0", 12-len(decimalPart))
		}

		decimalMojos, err := strconv.ParseUint(decimalPart, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid decimal part: %s", parts[1])
		}

		mojos += decimalMojos
	}

	return mojos, nil
}

// SplitLargestCoin finds the largest coin in the wallet and splits it
func SplitLargestCoin() {
	client, err := rpc.NewClient(rpc.ConnectionModeHTTP, rpc.WithAutoConfig())
	if err != nil {
		slogs.Logr.Fatal("error creating chia RPC client", "error", err)
	}

	walletID := viper.GetUint32("coins-wallet-id")
	fingerprint := viper.GetInt("coins-fingerprint")


	slogs.Logr.Debug("Getting spendable coins", "wallet_id", walletID)

	// Get spendable coins for the wallet
	spendableCoins, _, err := client.WalletService.GetSpendableCoins(&rpc.GetSpendableCoinsOptions{
		WalletID: walletID,
	})
	if err != nil {
		slogs.Logr.Fatal("error getting spendable coins", "error", err)
	}

	if spendableCoins == nil || spendableCoins.ConfirmedRecords.IsAbsent() {
		slogs.Logr.Fatal("no spendable coins found for wallet", "wallet_id", walletID)
	}

	records := spendableCoins.ConfirmedRecords.MustGet()
	if len(records) == 0 {
		slogs.Logr.Fatal("no spendable coins found in wallet", "wallet_id", walletID)
	}

	// Find the largest coin from the spendable coins
	var largestCoin types.CoinRecord
	largestAmount := uint64(0)

	for _, record := range records {
		if record.Coin.Amount > largestAmount {
			largestAmount = record.Coin.Amount
			largestCoin = record
		}
	}

	if largestAmount == 0 {
		slogs.Logr.Fatal("no coins with value found in wallet", "wallet_id", walletID)
	}

	coinID := largestCoin.Coin.ID().String()

	slogs.Logr.Info("Found largest coin",
		"coin_id", coinID,
		"amount", largestCoin.Coin.Amount,
		"wallet_id", walletID)

	// Convert string amounts to mojos
	amountPerCoinStr := viper.GetString("coins-amount-per-coin")
	feeStr := viper.GetString("coins-fee")

	amountPerCoin, err := convertXCHToMojos(amountPerCoinStr)
	if err != nil {
		slogs.Logr.Fatal("error converting amount-per-coin to mojos", "error", err)
	}

	fee, err := convertXCHToMojos(feeStr)
	if err != nil {
		slogs.Logr.Fatal("error converting fee to mojos", "error", err)
	}

	numberOfCoins := viper.GetUint32("coins-number-of-coins")

	// Validate that we have enough in the coin to split
	totalNeeded := amountPerCoin*uint64(numberOfCoins) + fee
	if largestCoin.Coin.Amount < totalNeeded {
		slogs.Logr.Fatal("largest coin does not have enough value for split",
			"coin_amount", largestCoin.Coin.Amount,
			"total_needed", totalNeeded,
			"amount_per_coin", amountPerCoin,
			"number_of_coins", numberOfCoins,
			"fee", fee)
	}

	slogs.Logr.Info("Splitting coin",
		"coin_id", coinID,
		"amount_per_coin", amountPerCoin,
		"number_of_coins", numberOfCoins,
		"fee", fee)

	// Split the coin
	splitResponse, _, err := client.WalletService.SplitCoins(&rpc.SplitCoinsOptions{
		WalletID:      walletID,
		TargetCoinID:  coinID,
		AmountPerCoin: amountPerCoin,
		NumberOfCoins: numberOfCoins,
		Fee:           fee,
		Push:          true,
	})
	if err != nil {
		slogs.Logr.Fatal("error splitting coin", "error", err)
	}

	if splitResponse == nil {
		slogs.Logr.Fatal("no response from split_coins")
	}

	if splitResponse.TransactionID.IsPresent() {
		fmt.Printf("Successfully split coin. Transaction ID: %s\n", splitResponse.TransactionID.MustGet())
	} else {
		fmt.Println("Coin split initiated successfully")
	}
}

func init() {
	splitLargestCmd.PersistentFlags().IntP("fingerprint", "f", 0, "Fingerprint of the wallet to use")
	splitLargestCmd.PersistentFlags().StringP("amount-per-coin", "a", "", "The amount of each newly created coin, in XCH or CAT units")
	splitLargestCmd.PersistentFlags().Uint32P("number-of-coins", "n", 0, "The number of coins we are creating")
	splitLargestCmd.PersistentFlags().Uint32P("id", "i", 1, "Id of the wallet to use")
	splitLargestCmd.PersistentFlags().StringP("fee", "m", "0", "Set the fees for the transaction, in XCH")

	// Mark required flags
	cobra.CheckErr(splitLargestCmd.MarkPersistentFlagRequired("amount-per-coin"))
	cobra.CheckErr(splitLargestCmd.MarkPersistentFlagRequired("number-of-coins"))

	// Bind flags to viper
	cobra.CheckErr(viper.BindPFlag("coins-fingerprint", splitLargestCmd.PersistentFlags().Lookup("fingerprint")))
	cobra.CheckErr(viper.BindPFlag("coins-amount-per-coin", splitLargestCmd.PersistentFlags().Lookup("amount-per-coin")))
	cobra.CheckErr(viper.BindPFlag("coins-number-of-coins", splitLargestCmd.PersistentFlags().Lookup("number-of-coins")))
	cobra.CheckErr(viper.BindPFlag("coins-wallet-id", splitLargestCmd.PersistentFlags().Lookup("id")))
	cobra.CheckErr(viper.BindPFlag("coins-fee", splitLargestCmd.PersistentFlags().Lookup("fee")))

	coinsCmd.AddCommand(splitLargestCmd)
}
