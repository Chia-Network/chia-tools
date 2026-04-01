package config

import (
	"os"
	"strings"

	"github.com/chia-network/go-modules/pkg/slogs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/chia-network/chia-tools/cmd"
)

var (
	skipConfirm bool
	retries     uint
)

// logDetectedChiaEnvVars scans the process environment for variables prefixed
// with "chia." or "chia__" and logs each one so the user knows which env vars
// will be applied to the config before --set flags are processed.
func logDetectedChiaEnvVars() {
	for _, env := range os.Environ() {
		for _, prefix := range []string{"chia.", "chia__"} {
			if strings.HasPrefix(env, prefix) {
				name, _, _ := strings.Cut(env, "=")
				slogs.Logr.Info("Detected chia environment variable that will be applied to config", "env_var", name)
				break
			}
		}
	}
}

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Utilities for working with chia config",
}

func init() {
	configCmd.PersistentFlags().String("config", "", "existing config file to use (default is to look in $CHIA_ROOT)")
	cobra.CheckErr(viper.BindPFlag("config", configCmd.PersistentFlags().Lookup("config")))

	cmd.RootCmd.AddCommand(configCmd)
}
