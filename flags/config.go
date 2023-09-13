package flags

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type FlagConfig struct {
	DebugMode bool
}

var Config FlagConfig

func Init(rootCmd *cobra.Command) {
	rootCmd.PersistentFlags().BoolVarP(&Config.DebugMode, "debug", "d", false, "Enable debug logging mode")
	err := viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	if err != nil {
		log.Fatal("unable to initialise flags: ", err)
	}
}

func GetDebugMode() bool {
	return viper.GetBool("debug")
}
