package config

import (
	"fmt"
	"log"
	"os"

	"github.com/chia-network/go-chia-libs/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

		err = os.WriteFile(viper.GetString("output"), out, 0655)
		if err != nil {
			log.Fatalln(err.Error())
		}
	},
}

func init() {
	var (
		outputFile string
	)

	generateCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "config.yml", "Output file for config")
	cobra.CheckErr(viper.BindPFlag("output", generateCmd.PersistentFlags().Lookup("output")))

	configCmd.AddCommand(generateCmd)
}
