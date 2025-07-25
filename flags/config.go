package flags

import (
	"fmt"
	"os"
	"strings"

	"github.com/amp-labs/cli/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type FlagConfig struct {
	DebugMode bool
}

// Init initializes global flags for the CLI.
func Init(rootCmd *cobra.Command) error {
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug logging mode, defaults to false.")
	rootCmd.PersistentFlags().StringP("project", "p", "", "Ampersand project name or ID")
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

	return nil
}

// InitAndBindFormatFlag initializes and binds the format flag to the provided command.
func InitAndBindFormatFlag(cmd *cobra.Command) error {
	cmd.Flags().StringP("format", "f", "json", "Output format, defaults to json. Options: json, yaml")

	if err := viper.BindPFlag("format", cmd.Flags().Lookup("format")); err != nil {
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

func GetProjectOrFail() string {
	p := viper.GetString("project")
	if p == "" {
		// This is using fmt.Fprint instead of logger.Fatal because the logger package
		// depends on the flags package, so we don't import it here to avoid a circular dependency.
		fmt.Fprint(os.Stderr, "Must provide a project name or ID in the --project flag\n")
		os.Exit(1)
	}

	return p
}

func GetAPIKey() string {
	return viper.GetString("key")
}
