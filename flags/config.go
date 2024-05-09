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
	rootCmd.PersistentFlags().StringP("key", "k", "", "Ampersand API key")

	if err := viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug")); err != nil {
		return err
	}

	if err := viper.BindPFlag("project", rootCmd.PersistentFlags().Lookup("project")); err != nil {
		return err
	}

	if err := viper.BindPFlag("key", rootCmd.PersistentFlags().Lookup("key")); err != nil {
		return err
	}

	if err := viper.BindEnv("key", "AMP_API_KEY"); err != nil {
		panic(err)
	}

	if err := viper.BindEnv("project", "AMP_PROJECT_ID"); err != nil {
		panic(err)
	}

	if err := viper.BindEnv("debug", "AMP_DEBUG"); err != nil {
		panic(err)
	}

	return nil
}

func GetDebugMode() bool {
	return viper.GetBool("debug")
}

func GetProjectId() string {
	return viper.GetString("project")
}

func GetAPIKey() string {
	return viper.GetString("key")
}
