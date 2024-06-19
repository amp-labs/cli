package flags

import (
	"strings"

	"github.com/amp-labs/cli/utils"
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
	rootCmd.PersistentFlags().StringP("format", "f", "json", "Output format")

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

	if err := viper.BindPFlag("format", rootCmd.PersistentFlags().Lookup("format")); err != nil {
		return err
	}

	return nil
}

func GetOutputFormat() utils.Format {
	switch strings.ToLower(viper.GetString("format")) {
	case "json":
		return utils.JSON
	case "yaml", "yml":
		return utils.YAML
	default:
		return utils.Unknown
	}
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
