package config

import (
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

		valuesToSet := viper.GetStringMapString("set")
		for path, value := range valuesToSet {
			pathMap := config.ParsePathsFromStrings([]string{path}, false)
			var key string
			var pathSlice []string
			for key, pathSlice = range pathMap {
				break
			}
			err = cfg.SetFieldByPath(pathSlice, value)
			if err != nil {
				log.Fatalf("Error setting path `%s` to `%s`: %s\n", key, value, err.Error())
			}
		}

		out, err := yaml.Marshal(cfg)
		if err != nil {
			log.Fatalf("Error marshalling config: %s\n", err.Error())
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
		setValues  map[string]string
	)

	generateCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "config.yml", "Output file for config")
	generateCmd.PersistentFlags().StringToStringVarP(&setValues, "set", "s", nil, "Paths and values to set in the config")

	cobra.CheckErr(viper.BindPFlag("output", generateCmd.PersistentFlags().Lookup("output")))
	cobra.CheckErr(viper.BindPFlag("set", generateCmd.PersistentFlags().Lookup("set")))

	configCmd.AddCommand(generateCmd)
}
