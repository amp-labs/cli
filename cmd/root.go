package cmd

import (
	"os"

	"github.com/amp-labs/cli/flags"
	"github.com/amp-labs/cli/logger"
	"github.com/amp-labs/cli/region"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:   "amp",
	Short: "Ampersand CLI",
	Long:  "The Ampersand CLI allows you to interact with the Ampersand platform.",

	// Resolve the region before any command runs so that a malformed --region or an
	// unreadable config fails immediately, before running any command.
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		flagValue, _ := cmd.Root().PersistentFlags().GetString(region.FlagName)

		selected, err := region.Resolve(flagValue)
		if err != nil {
			logger.FatalErr("Unable to determine the region", err)
		}

		flags.SetRegion(selected)
		warnIfEnvIgnored(flagValue, selected)
	},

	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// warnIfEnvIgnored tells the user when AMP_REGION is set, but will be ignored in favor of the region
// saved previously by 'amp set:region'.
func warnIfEnvIgnored(flagValue string, selected region.Region) {
	if flagValue != "" {
		return
	}

	fromEnv := os.Getenv(region.EnvVar)
	if fromEnv == "" {
		return
	}

	parsed, err := region.Parse(fromEnv)
	if err == nil && parsed == selected {
		return
	}

	logger.Warnf("ignoring %s=%q; using the region saved by 'amp set:region' (%s) instead.",
		region.EnvVar, fromEnv, selected)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		return
	}
}

func init() {
	// Disable the autcompletion command from being shown in `amp --help`
	// nolint: lll
	// See https://github.com/spf13/cobra/blob/main/site/content/completions/_index.md#adapting-the-default-completion-command
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	err := flags.Init(rootCmd)
	if err != nil {
		logger.FatalErr("unable to initialize flags", err)
	}
}
