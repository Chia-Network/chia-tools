package config

import (
	"fmt"
	"os"

	"github.com/chia-network/go-chia-libs/pkg/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// generateCmd generates a new chia config
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a new chia configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.ChiaConfig{}
		out, err := yaml.Marshal(cfg)
		if err != nil {
			fmt.Printf("Error marshalling config: %s\n", err.Error())
			os.Exit(1)
		}
		fmt.Print(string(out))
	},
}

func init() {
	configCmd.AddCommand(generateCmd)
}
