package flags

import (
	"fmt"
	"os"
	"strings"

	"github.com/amp-labs/cli/region"
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
	rootCmd.PersistentFlags().StringP(region.FlagName, "r", "",
		"Ampersand region to talk to, e.g. "+strings.Join(region.Known(), " or ")+". "+
			"If never set, defaults to "+string(region.Default)+".")

	err := viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	if err != nil {
		return err
	}

	err = viper.BindPFlag("project", rootCmd.PersistentFlags().Lookup("project"))
	if err != nil {
		return err
	}

	err = viper.BindPFlag("key", rootCmd.PersistentFlags().Lookup("key"))
	if err != nil {
		return err
	}

	err = viper.BindEnv("key", "AMP_API_KEY")
	if err != nil {
		panic(err)
	}

	// Unlike --key, the region's environment variable (AMP_REIGON) is not bound here,
	// as viper would then rank it above the 'amp set:region' saved/config region, when we want it ranked below
	return viper.BindPFlag(region.FlagName, rootCmd.PersistentFlags().Lookup(region.FlagName))
}

// InitAndBindFormatFlag initializes and binds the format flag to the provided command.
func InitAndBindFormatFlag(cmd *cobra.Command) error {
	cmd.Flags().StringP("format", "f", "json", "Output format, defaults to json. Options: json, yaml")

	err := viper.BindPFlag("format", cmd.Flags().Lookup("format"))
	if err != nil {
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

// GetProject returns the configured project name or ID, or an empty string if
// none is set. Unlike GetProjectOrFail it does not exit, so callers that can
// operate without a project (e.g. offline validation) can degrade gracefully.
func GetProject() string {
	return viper.GetString("project")
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

// GetRegion returns the region this process talks to.
// PersistentPreRun resolves it, calls SetRegion, and then this function returns that value.
func GetRegion() region.Region {
	if resolved := viper.GetString(region.FlagName); resolved != "" {
		return region.Region(resolved)
	}

	return region.Default
}

// SetRegion records the resolved region for the rest of the process.
func SetRegion(selected region.Region) {
	viper.Set(region.FlagName, string(selected))
}
