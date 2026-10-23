package cmd

import (
	"github.com/amp-labs/cli/flags"
	"github.com/amp-labs/cli/logger"
	"github.com/amp-labs/cli/region"
	"github.com/spf13/cobra"
)

var getRegionCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:   "get:region",
	Short: "Print the region the CLI is talking to",
	Long: "Print the region the CLI is talking to.\n\n" +
		"This is the --region flag if given, otherwise the region saved by 'amp set:region', " +
		"otherwise the " + region.EnvVar + " environment variable, otherwise defaults to " + string(region.Default) + ".",
	Run: func(cmd *cobra.Command, args []string) {
		// The root command's PersistentPreRun has already resolved this, and would
		// have exited if it could not.
		logger.Info(string(flags.GetRegion()))
	},
}

func init() {
	rootCmd.AddCommand(getRegionCmd)
}
