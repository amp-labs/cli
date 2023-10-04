package flags

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type FlagConfig struct {
	DebugMode bool
}

func Init(rootCmd *cobra.Command) error {
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug logging mode")
	rootCmd.PersistentFlags().StringP("project", "p", "", "Ampersand project ID")

	if err := viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug")); err != nil {
		return err
	}

	if err := viper.BindPFlag("project", rootCmd.PersistentFlags().Lookup("project")); err != nil {
		return err
	}

	return nil
}

func GetDebugMode() bool {
	return viper.GetBool("debug")
}

func GetProjectId() string {
	return viper.GetString("project")
}
