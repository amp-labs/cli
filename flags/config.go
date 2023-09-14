package flags

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type FlagConfig struct {
	DebugMode bool
}

var Config FlagConfig

// TODO: Will need a better implementation with multiple flags
func Init(rootCmd *cobra.Command) error {
	rootCmd.PersistentFlags().BoolVarP(&Config.DebugMode, "debug", "d", false, "Enable debug logging mode")
	return viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
}

func GetDebugMode() bool {
	return viper.GetBool("debug")
}
