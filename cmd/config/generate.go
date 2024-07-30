package config

import (
	"fmt"
	"log"
	"os"

	"github.com/chia-network/go-chia-libs/pkg/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// generateCmd generates a new chia config
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a new chia configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadDefaultConfig()
		if err != nil {
			log.Fatalln(err.Error())
		}

		err = cfg.FillValuesFromEnvironment()
		if err != nil {
			log.Fatalln(err.Error())
		}

		out, err := yaml.Marshal(cfg)
		if err != nil {
			fmt.Printf("Error marshalling config: %s\n", err.Error())
			os.Exit(1)
		}

		os.WriteFile("config.yml", out, 0655)
	},
}

func init() {
	configCmd.AddCommand(generateCmd)
}
